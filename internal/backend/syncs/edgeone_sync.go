package syncs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/edgeonezone"
	"vibe_ip/internal/ent/edgeonezoneaccelerationdomain"
	"vibe_ip/internal/ent/edgeonezonednsrecord"
	"vibe_ip/internal/ent/edgeonezoneorigingroup"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/cdn"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gotime"
	"golang.org/x/sync/errgroup"
)

type EdgeOneSync struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewEdgeOneSync 创建EdgeOne同步实例
func NewEdgeOneSync(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *EdgeOneSync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &EdgeOneSync{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *EdgeOneSync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取EdgeOne提供程序
func (s *EdgeOneSync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*cdn.TencentEOProvider, error) {

	// 创建EdgeOne提供程序选项
	var cdnProviderOpts = make([]cdn.Option, 0, 1)
	if s.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的EdgeOne配置
	cdnProviderCfg := cdn.NewDefaultTencentEOConfig()
	err := cdnProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析EdgeOne配置失败: %w", err)
	}

	// 创建EdgeOne提供程序
	cdnProvider, err := cdn.NewTencentEOProviderRaw(ctx, s.logger, cdnProviderCfg, cdnProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建EdgeOne提供程序失败: %w", err)
	}

	return cdnProvider, nil
}

// 同步 全部账号 所有站点
func (s *EdgeOneSync) SyncAllZones(ctx context.Context) (err error) {
	// 查询全部EdgeOne平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部EdgeOne平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisZones(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点
func (s *EdgeOneSync) SyncThisZones(ctx context.Context, config_id int64) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取EdgeOne提供程序失败: %w", err)
	}

	// 获取所有站点
	zones, err := cdnProvider.ZonesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, zone := range zones {
		// 查询站点记录
		firstZone, err := s.db.EdgeOneZone.Query().
			Where(edgeonezone.ConfigIDEQ(config_id)).
			Where(edgeonezone.ZoneIDEQ(utils.PtrToString(zone.ZoneId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			err = s.db.EdgeOneZone.Create().
				SetSyncAt(time.Now()).
				SetConfigID(firstPlatformCofing.ID).
				SetZoneID(utils.PtrToString(zone.ZoneId)).
				SetName(utils.PtrToString(zone.ZoneName)).
				SetArea(edgeonezone.Area(utils.PtrToString(zone.Area))).
				SetType(edgeonezone.Type(utils.PtrToString(zone.Type))).
				SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status))).
				SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus))).
				SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus))).
				SetPaused(utils.PtrToBool(zone.Paused)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			err = s.db.EdgeOneZone.UpdateOne(firstZone).
				SetSyncAt(time.Now()).
				SetName(utils.PtrToString(zone.ZoneName)).
				SetArea(edgeonezone.Area(utils.PtrToString(zone.Area))).
				SetType(edgeonezone.Type(utils.PtrToString(zone.Type))).
				SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status))).
				SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus))).
				SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus))).
				SetPaused(utils.PtrToBool(zone.Paused)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 某个账号 某个站点
func (s *EdgeOneSync) SyncThisZone(ctx context.Context, config_id int64, zone_id string) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取EdgeOne提供程序失败: %w", err)
	}

	// 获取所有站点
	zones, err := cdnProvider.ZonesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, zone := range zones {
		if utils.PtrToString(zone.ZoneId) != zone_id {
			continue
		}
		// 查询站点记录
		firstZone, err := s.db.EdgeOneZone.Query().
			Where(edgeonezone.ZoneIDEQ(utils.PtrToString(zone.ZoneId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			err = s.db.EdgeOneZone.Create().
				SetSyncAt(time.Now()).
				SetConfigID(firstPlatformCofing.ID).
				SetZoneID(utils.PtrToString(zone.ZoneId)).
				SetName(utils.PtrToString(zone.ZoneName)).
				SetArea(edgeonezone.Area(utils.PtrToString(zone.Area))).
				SetType(edgeonezone.Type(utils.PtrToString(zone.Type))).
				SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status))).
				SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus))).
				SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus))).
				SetPaused(utils.PtrToBool(zone.Paused)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			err = s.db.EdgeOneZone.UpdateOne(firstZone).
				SetSyncAt(time.Now()).
				SetName(utils.PtrToString(zone.ZoneName)).
				SetArea(edgeonezone.Area(utils.PtrToString(zone.Area))).
				SetType(edgeonezone.Type(utils.PtrToString(zone.Type))).
				SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status))).
				SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus))).
				SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus))).
				SetPaused(utils.PtrToBool(zone.Paused)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有源站组
func (s *EdgeOneSync) SyncAllZoneOriginGroups(ctx context.Context) (err error) {
	// 查询全部EdgeOne平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部EdgeOne平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisZoneOriginGroups(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点 所有源站组
func (s *EdgeOneSync) SyncThisZoneOriginGroups(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEdgeOneZoneIDs, err := s.db.EdgeOneZone.Query().
		Where(edgeonezone.ConfigIDEQ(config_id)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询所有站点失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemEdgeOneZoneID := range allEdgeOneZoneIDs {
		itemEdgeOneZoneID := itemEdgeOneZoneID
		eg.Go(func() error {
			return s.SyncThisZoneOriginGroup(egCtx, config_id, itemEdgeOneZoneID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 某个站点 所有源站组
func (s *EdgeOneSync) SyncThisZoneOriginGroup(ctx context.Context, config_id int64, zone_id int64) (err error) {
	// 查询站点记录
	firstEdgeOneZone, err := s.db.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEdgeOneZone.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取EdgeOne提供程序失败: %w", err)
	}

	// 获取所有源站组
	originGroups, err := cdnProvider.OriginGroupsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		return fmt.Errorf("获取所有源站组失败: %w", err)
	}

	// 入库所有源站组
	for _, originGroup := range originGroups {
		// 查询源站组记录
		firstOriginGroup, err := s.db.EdgeOneZoneOriginGroup.Query().
			Where(edgeonezoneorigingroup.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezoneorigingroup.GroupIDEQ(utils.PtrToString(originGroup.GroupId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询源站组记录失败: %w", err)
			}
			// 创建源站组记录
			err = s.db.EdgeOneZoneOriginGroup.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(originGroup.CreateTime)).Time).
				SetSyncAt(time.Now()).
				SetEdgeoneZoneID(firstEdgeOneZone.ID).
				SetGroupID(utils.PtrToString(originGroup.GroupId)).
				SetName(utils.PtrToString(originGroup.Name)).
				SetType(edgeonezoneorigingroup.Type(utils.PtrToString(originGroup.Type))).
				SetRecords([]byte(gojson.JsonEncodeNoError(originGroup.Records))).
				SetReferences([]byte(gojson.JsonEncodeNoError(originGroup.References))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建源站组记录失败: %w", err)
			}
		} else {
			// 更新源站组记录
			err = s.db.EdgeOneZoneOriginGroup.UpdateOne(firstOriginGroup).
				SetSyncAt(time.Now()).
				SetName(utils.PtrToString(originGroup.Name)).
				SetType(edgeonezoneorigingroup.Type(utils.PtrToString(originGroup.Type))).
				SetRecords([]byte(gojson.JsonEncodeNoError(originGroup.Records))).
				SetReferences([]byte(gojson.JsonEncodeNoError(originGroup.References))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新源站组记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有记录
func (s *EdgeOneSync) SyncAllZoneRecords(ctx context.Context) (err error) {
	// 查询全部EdgeOne平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部EdgeOne平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisZoneRecords(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点 所有记录
func (s *EdgeOneSync) SyncThisZoneRecords(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEdgeOneZoneIDs, err := s.db.EdgeOneZone.Query().
		Where(edgeonezone.ConfigIDEQ(config_id)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询所有站点失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemEdgeOneZoneID := range allEdgeOneZoneIDs {
		itemEdgeOneZoneID := itemEdgeOneZoneID
		eg.Go(func() error {
			return s.SyncThisZoneRecord(egCtx, config_id, itemEdgeOneZoneID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 某个站点 所有记录
func (s *EdgeOneSync) SyncThisZoneRecord(ctx context.Context, config_id int64, zone_id int64) (err error) {
	// 查询站点记录
	firstEdgeOneZone, err := s.db.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEdgeOneZone.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取EdgeOne提供程序失败: %w", err)
	}

	// 获取所有DNS记录
	dnsRecords, err := cdnProvider.DnsRecordsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		return fmt.Errorf("获取所有DNS记录失败: %w", err)
	}

	// 入库所有DNS记录
	for _, dnsRecord := range dnsRecords {
		// 查询DNS记录
		firstDNSRecord, err := s.db.EdgeOneZoneDNSRecord.Query().
			Where(edgeonezonednsrecord.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezonednsrecord.RecordIDEQ(utils.PtrToString(dnsRecord.RecordId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询DNS记录记录失败: %w", err)
			}
			// 创建DNS记录记录
			err = s.db.EdgeOneZoneDNSRecord.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(dnsRecord.CreatedOn)).Time).
				SetSyncAt(time.Now()).
				SetEdgeoneZoneID(firstEdgeOneZone.ID).
				SetRecordID(utils.PtrToString(dnsRecord.RecordId)).
				SetName(utils.PtrToString(dnsRecord.Name)).
				SetType(edgeonezonednsrecord.Type(utils.PtrToString(dnsRecord.Type))).
				SetLocation(utils.PtrToString(dnsRecord.Location)).
				SetContent(utils.PtrToString(dnsRecord.Content)).
				SetTTL(utils.PtrToInt64(dnsRecord.TTL)).
				SetWeight(utils.PtrToInt64(dnsRecord.Weight)).
				SetPriority(utils.PtrToInt64(dnsRecord.Priority)).
				SetStatus(edgeonezonednsrecord.Status(utils.PtrToString(dnsRecord.Status))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建DNS记录记录失败: %w", err)
			}
		} else {
			// 更新DNS记录记录
			err = s.db.EdgeOneZoneDNSRecord.UpdateOne(firstDNSRecord).
				SetSyncAt(time.Now()).
				SetName(utils.PtrToString(dnsRecord.Name)).
				SetType(edgeonezonednsrecord.Type(utils.PtrToString(dnsRecord.Type))).
				SetLocation(utils.PtrToString(dnsRecord.Location)).
				SetContent(utils.PtrToString(dnsRecord.Content)).
				SetTTL(utils.PtrToInt64(dnsRecord.TTL)).
				SetWeight(utils.PtrToInt64(dnsRecord.Weight)).
				SetPriority(utils.PtrToInt64(dnsRecord.Priority)).
				SetStatus(edgeonezonednsrecord.Status(utils.PtrToString(dnsRecord.Status))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新DNS记录记录失败: %w", err)
			}
		}
	}

	// 获取所有加速域名
	accelerationDomains, err := cdnProvider.AccelerationDomainsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		return fmt.Errorf("获取所有加速域名失败: %w", err)
	}

	// 入库所有加速域名
	for _, accelerationDomain := range accelerationDomains {
		// 查询加速域名记录
		firstAccelerationDomain, err := s.db.EdgeOneZoneAccelerationDomain.Query().
			Where(edgeonezoneaccelerationdomain.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezoneaccelerationdomain.DomainNameEQ(utils.PtrToString(accelerationDomain.DomainName))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询加速域名记录失败: %w", err)
			}
			// 创建加速域名记录
			err = s.db.EdgeOneZoneAccelerationDomain.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(accelerationDomain.CreatedOn)).Time).
				SetSyncAt(time.Now()).
				SetEdgeoneZoneID(firstEdgeOneZone.ID).
				SetDomainName(utils.PtrToString(accelerationDomain.DomainName)).
				SetDomainStatus(edgeonezoneaccelerationdomain.DomainStatus(utils.PtrToString(accelerationDomain.DomainStatus))).
				SetCname(utils.PtrToString(accelerationDomain.Cname)).
				SetIpv6Status(edgeonezoneaccelerationdomain.Ipv6Status(utils.PtrToString(accelerationDomain.IPv6Status))).
				SetIdentificationStatus(edgeonezoneaccelerationdomain.IdentificationStatus(utils.PtrToString(accelerationDomain.IdentificationStatus))).
				SetOwnershipVerification([]byte(gojson.JsonEncodeNoError(accelerationDomain.OwnershipVerification))).
				SetOriginDetail([]byte(gojson.JsonEncodeNoError(accelerationDomain.OriginDetail))).
				SetOriginProtocol(edgeonezoneaccelerationdomain.OriginProtocol(utils.PtrToString(accelerationDomain.OriginProtocol))).
				SetHTTPOriginPort(utils.PtrToUint64(accelerationDomain.HttpOriginPort)).
				SetHTTPSOriginPort(utils.PtrToUint64(accelerationDomain.HttpsOriginPort)).
				SetCertificate([]byte(gojson.JsonEncodeNoError(accelerationDomain.Certificate))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建加速域名记录失败: %w", err)
			}
		} else {
			// 更新加速域名记录
			err = s.db.EdgeOneZoneAccelerationDomain.UpdateOne(firstAccelerationDomain).
				SetSyncAt(time.Now()).
				SetDomainStatus(edgeonezoneaccelerationdomain.DomainStatus(utils.PtrToString(accelerationDomain.DomainStatus))).
				SetCname(utils.PtrToString(accelerationDomain.Cname)).
				SetIpv6Status(edgeonezoneaccelerationdomain.Ipv6Status(utils.PtrToString(accelerationDomain.IPv6Status))).
				SetIdentificationStatus(edgeonezoneaccelerationdomain.IdentificationStatus(utils.PtrToString(accelerationDomain.IdentificationStatus))).
				SetOwnershipVerification([]byte(gojson.JsonEncodeNoError(accelerationDomain.OwnershipVerification))).
				SetOriginDetail([]byte(gojson.JsonEncodeNoError(accelerationDomain.OriginDetail))).
				SetOriginProtocol(edgeonezoneaccelerationdomain.OriginProtocol(utils.PtrToString(accelerationDomain.OriginProtocol))).
				SetHTTPOriginPort(utils.PtrToUint64(accelerationDomain.HttpOriginPort)).
				SetHTTPSOriginPort(utils.PtrToUint64(accelerationDomain.HttpsOriginPort)).
				SetCertificate([]byte(gojson.JsonEncodeNoError(accelerationDomain.Certificate))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新加速域名记录失败: %w", err)
			}
		}
	}

	return nil
}
