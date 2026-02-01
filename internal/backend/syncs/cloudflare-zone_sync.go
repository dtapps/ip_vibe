package syncs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflarezone"
	"vibe_ip/internal/ent/cloudflarezonerecord"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/dns"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"golang.org/x/sync/errgroup"
)

// Cloudflare站点同步
type CloudflareZoneSync struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewCloudflareZoneSync 创建Cloudflare站点同步实例
func NewCloudflareZoneSync(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *CloudflareZoneSync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &CloudflareZoneSync{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *CloudflareZoneSync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取ESA提供程序
func (s *CloudflareZoneSync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*dns.CloudfareDNSProvider, error) {

	// 创建Cloudflare DNS提供程序选项
	var dnsProviderOpts = make([]dns.Option, 0, 1)
	if s.httpTransport != nil {
		dnsProviderOpts = append(dnsProviderOpts, dns.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的Cloudflare DNS配置
	dnsProviderCfg := dns.NewDefaultCloudfareDNSConfig()
	err := dnsProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析Cloudflare DNS配置失败: %w", err)
	}

	// 创建Cloudflare DNS提供程序
	dnsProvider, err := dns.NewCloudfareDNSProviderRaw(ctx, s.logger, dnsProviderCfg, dnsProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建Cloudflare DNS提供程序失败: %w", err)
	}

	return dnsProvider, nil
}

// 同步 全部账号 所有站点
func (s *CloudflareZoneSync) SyncAllZones(ctx context.Context) (err error) {
	// 查询全部Cloudflare平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部Cloudflare平台失败: %w", err)
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
func (s *CloudflareZoneSync) SyncThisZones(ctx context.Context, config_id int64) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取Cloudflare提供程序
	dnsProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return err
	}

	// 获取所有站点
	zones, err := dnsProvider.ZonesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, zone := range zones {
		// 查询站点记录
		firstZone, err := s.db.CloudflareZone.Query().
			Where(cloudflarezone.ConfigIDEQ(config_id)).
			Where(cloudflarezone.ZoneIDEQ(zone.ID)).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			createOne := s.db.CloudflareZone.Create()
			createOne.SetRemoteCreatedAt(zone.CreatedOn)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			createOne.SetZoneID(zone.ID)
			createOne.SetAccount(json.RawMessage(gojson.JsonEncodeNoError(zone.Account)))
			createOne.SetName(zone.Name)
			createOne.SetNameServers(zone.NameServers)
			createOne.SetOriginalDnshost(zone.OriginalDnshost)
			createOne.SetOriginalNameServers(zone.OriginalNameServers)
			createOne.SetOriginalRegistrar(zone.OriginalRegistrar)
			createOne.SetPaused(zone.Paused)
			if zone.Status != "" {
				createOne.SetStatus(cloudflarezone.Status(zone.Status))
			}
			if zone.Type != "" {
				createOne.SetType(cloudflarezone.Type(zone.Type))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			updateOne := s.db.CloudflareZone.UpdateOne(firstZone)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetAccount(json.RawMessage(gojson.JsonEncodeNoError(zone.Account)))
			updateOne.SetName(zone.Name)
			updateOne.SetNameServers(zone.NameServers)
			updateOne.SetOriginalDnshost(zone.OriginalDnshost)
			updateOne.SetOriginalNameServers(zone.OriginalNameServers)
			updateOne.SetOriginalRegistrar(zone.OriginalRegistrar)
			updateOne.SetPaused(zone.Paused)
			if zone.Status != "" {
				updateOne.SetStatus(cloudflarezone.Status(zone.Status))
			}
			if zone.Type != "" {
				updateOne.SetType(cloudflarezone.Type(zone.Type))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 某个账号 某个站点
func (s *CloudflareZoneSync) SyncThisSite(ctx context.Context, config_id int64, zone_id string) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取Cloudflare提供程序
	dnsProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return err
	}

	// 获取所有站点
	zones, err := dnsProvider.ZonesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, zone := range zones {
		if zone.ID != zone_id {
			continue
		}
		// 查询站点记录
		firstZone, err := s.db.CloudflareZone.Query().
			Where(cloudflarezone.ConfigIDEQ(config_id)).
			Where(cloudflarezone.ZoneIDEQ(zone.ID)).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			createOne := s.db.CloudflareZone.Create()
			createOne.SetRemoteCreatedAt(zone.CreatedOn)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			createOne.SetZoneID(zone.ID)
			createOne.SetAccount(json.RawMessage(gojson.JsonEncodeNoError(zone.Account)))
			createOne.SetName(zone.Name)
			createOne.SetNameServers(zone.NameServers)
			createOne.SetOriginalDnshost(zone.OriginalDnshost)
			createOne.SetOriginalNameServers(zone.OriginalNameServers)
			createOne.SetOriginalRegistrar(zone.OriginalRegistrar)
			createOne.SetPaused(zone.Paused)
			if zone.Status != "" {
				createOne.SetStatus(cloudflarezone.Status(zone.Status))
			}
			if zone.Type != "" {
				createOne.SetType(cloudflarezone.Type(zone.Type))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			updateOne := s.db.CloudflareZone.UpdateOne(firstZone)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetAccount(json.RawMessage(gojson.JsonEncodeNoError(zone.Account)))
			updateOne.SetName(zone.Name)
			updateOne.SetNameServers(zone.NameServers)
			updateOne.SetOriginalDnshost(zone.OriginalDnshost)
			updateOne.SetOriginalNameServers(zone.OriginalNameServers)
			updateOne.SetOriginalRegistrar(zone.OriginalRegistrar)
			updateOne.SetPaused(zone.Paused)
			if zone.Status != "" {
				updateOne.SetStatus(cloudflarezone.Status(zone.Status))
			}
			if zone.Type != "" {
				updateOne.SetType(cloudflarezone.Type(zone.Type))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有记录
func (s *CloudflareZoneSync) SyncAllZoneRecords(ctx context.Context) (err error) {
	// 查询全部Cloudflare平台配置
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部Cloudflare平台失败: %w", err)
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
func (s *CloudflareZoneSync) SyncThisZoneRecords(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allZoneIDs, err := s.db.CloudflareZone.Query().
		Where(cloudflarezone.ConfigIDEQ(config_id)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询所有站点失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, allZoneID := range allZoneIDs {
		allZoneID := allZoneID
		eg.Go(func() error {
			return s.SyncThisZoneRecord(egCtx, config_id, allZoneID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 某个站点 所有记录
func (s *CloudflareZoneSync) SyncThisZoneRecord(ctx context.Context, config_id int64, zone_id int64) (err error) {
	// 查询站点记录
	firstZone, err := s.db.CloudflareZone.Query().
		Where(cloudflarezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取Cloudflare提供程序
	dnsProvider, err := s.GetApiClient(ctx, firstZone.Edges.Config.ConfigJSON)
	if err != nil {
		return err
	}

	// 获取所有解析记录
	records, err := dnsProvider.RecordsRaw(ctx, firstZone.ZoneID)
	if err != nil {
		return fmt.Errorf("获取所有解析记录失败: %w", err)
	}

	// 入库所有DNS记录
	for _, record := range records {
		// 查询解析记录
		firstRecord, err := s.db.CloudflareZoneRecord.Query().
			Where(cloudflarezonerecord.CloudflareZoneIDEQ(firstZone.ID)).
			Where(cloudflarezonerecord.RecordIDEQ(record.ID)).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询解析记录记录失败: %w", err)
			}
			// 创建解析记录记录
			createOne := s.db.CloudflareZoneRecord.Create()
			createOne.SetRemoteCreatedAt(record.CreatedOn)
			createOne.SetSyncAt(time.Now())
			createOne.SetCloudflareZoneID(firstZone.ID)
			createOne.SetRecordID(record.ID)
			createOne.SetComment(record.Comment)
			createOne.SetContent(record.Content)
			createOne.SetName(record.Name)
			createOne.SetProxiable(record.Proxiable)
			createOne.SetProxied(record.Proxied)
			createOne.SetSettings(json.RawMessage(gojson.JsonEncodeNoError(record.Settings)))
			createOne.SetTTL(float64(record.TTL))
			if record.Type != "" {
				createOne.SetType(cloudflarezonerecord.Type(record.Type))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建解析记录记录失败: %w", err)
			}
		} else {
			// 更新解析记录记录
			updateOne := s.db.CloudflareZoneRecord.UpdateOne(firstRecord)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetComment(record.Comment)
			updateOne.SetContent(record.Content)
			updateOne.SetName(record.Name)
			updateOne.SetProxiable(record.Proxiable)
			updateOne.SetProxied(record.Proxied)
			updateOne.SetSettings(json.RawMessage(gojson.JsonEncodeNoError(record.Settings)))
			updateOne.SetTTL(float64(record.TTL))
			if record.Type != "" {
				updateOne.SetType(cloudflarezonerecord.Type(record.Type))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新解析记录记录失败: %w", err)
			}
		}
	}

	return nil
}
