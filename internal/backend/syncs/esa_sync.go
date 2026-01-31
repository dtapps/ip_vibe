package syncs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/esasite"
	"vibe_ip/internal/ent/esasiteoriginpool"
	"vibe_ip/internal/ent/esasiterecord"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/cdn"

	"github.com/alibabacloud-go/tea/tea"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gotime"
	"golang.org/x/sync/errgroup"
)

type ESASync struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewESASync 创建ESA同步实例
func NewESASync(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *ESASync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &ESASync{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *ESASync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取ESA提供程序
func (s *ESASync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*cdn.AliyunESAProvider, error) {

	// 创建ESA提供程序选项
	var cdnProviderOpts = make([]cdn.Option, 0, 1)
	if s.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的ESA配置
	cdnProviderCfg := cdn.NewDefaultAliyunESAConfig()
	err := cdnProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析ESA配置失败: %w", err)
	}

	// 创建ESA提供程序
	cdnProvider, err := cdn.NewAliyunESAProviderRaw(ctx, s.logger, cdnProviderCfg, cdnProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建ESA提供程序失败: %w", err)
	}

	return cdnProvider, nil
}

// 同步 全部账号 所有站点
func (s *ESASync) SyncAllSites(ctx context.Context) (err error) {
	// 查询全部ESA平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部ESA平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisSites(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点
func (s *ESASync) SyncThisSites(ctx context.Context, config_id int64) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取ESA提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取ESA提供程序失败: %w", err)
	}

	// 获取所有站点
	sites, err := cdnProvider.SitesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, site := range sites {
		// 查询站点记录
		firstSite, err := s.db.EsaSite.Query().
			Where(esasite.ConfigIDEQ(config_id)).
			Where(esasite.SiteIDEQ(tea.Int64Value(site.SiteId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			err = s.db.EsaSite.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(site.CreateTime)).Time).
				SetSyncAt(time.Now()).
				SetConfigID(firstPlatformCofing.ID).
				SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType))).
				SetCnameZone(tea.StringValue(site.CnameZone)).
				SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage))).
				SetInstanceID(tea.StringValue(site.InstanceId)).
				SetNameServerList(tea.StringValue(site.NameServerList)).
				SetPlanName(tea.StringValue(site.PlanName)).
				SetPlanSpecName(tea.StringValue(site.PlanSpecName)).
				SetSiteID(tea.Int64Value(site.SiteId)).
				SetSiteName(tea.StringValue(site.SiteName)).
				SetStatus(esasite.Status(tea.StringValue(site.Status))).
				SetOfflineReason(tea.StringValue(site.OfflineReason)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			err = s.db.EsaSite.UpdateOne(firstSite).
				SetSyncAt(time.Now()).
				SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType))).
				SetCnameZone(tea.StringValue(site.CnameZone)).
				SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage))).
				SetInstanceID(tea.StringValue(site.InstanceId)).
				SetNameServerList(tea.StringValue(site.NameServerList)).
				SetPlanName(tea.StringValue(site.PlanName)).
				SetPlanSpecName(tea.StringValue(site.PlanSpecName)).
				SetSiteName(tea.StringValue(site.SiteName)).
				SetStatus(esasite.Status(tea.StringValue(site.Status))).
				SetOfflineReason(tea.StringValue(site.OfflineReason)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 某个账号 某个站点
func (s *ESASync) SyncThisSite(ctx context.Context, config_id int64, site_id int64) (err error) {
	// 查询平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取ESA提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取ESA提供程序失败: %w", err)
	}

	// 获取所有站点
	sites, err := cdnProvider.SitesRaw(ctx)
	if err != nil {
		return err
	}

	// 入库所有站点
	for _, site := range sites {
		if tea.Int64Value(site.SiteId) != site_id {
			continue
		}
		// 查询站点记录
		firstSite, err := s.db.EsaSite.Query().
			Where(esasite.ConfigIDEQ(config_id)).
			Where(esasite.SiteIDEQ(tea.Int64Value(site.SiteId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			err = s.db.EsaSite.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(site.CreateTime)).Time).
				SetSyncAt(time.Now()).
				SetConfigID(firstPlatformCofing.ID).
				SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType))).
				SetCnameZone(tea.StringValue(site.CnameZone)).
				SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage))).
				SetInstanceID(tea.StringValue(site.InstanceId)).
				SetNameServerList(tea.StringValue(site.NameServerList)).
				SetPlanName(tea.StringValue(site.PlanName)).
				SetPlanSpecName(tea.StringValue(site.PlanSpecName)).
				SetSiteID(tea.Int64Value(site.SiteId)).
				SetSiteName(tea.StringValue(site.SiteName)).
				SetStatus(esasite.Status(tea.StringValue(site.Status))).
				SetOfflineReason(tea.StringValue(site.OfflineReason)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			err = s.db.EsaSite.UpdateOne(firstSite).
				SetSyncAt(time.Now()).
				SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType))).
				SetCnameZone(tea.StringValue(site.CnameZone)).
				SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage))).
				SetInstanceID(tea.StringValue(site.InstanceId)).
				SetNameServerList(tea.StringValue(site.NameServerList)).
				SetPlanName(tea.StringValue(site.PlanName)).
				SetPlanSpecName(tea.StringValue(site.PlanSpecName)).
				SetSiteName(tea.StringValue(site.SiteName)).
				SetStatus(esasite.Status(tea.StringValue(site.Status))).
				SetOfflineReason(tea.StringValue(site.OfflineReason)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有源地址池
func (s *ESASync) SyncAllSiteOriginPools(ctx context.Context) (err error) {
	// 查询全部ESA平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部ESA平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisSiteOriginPools(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点 所有源地址池
func (s *ESASync) SyncThisSiteOriginPools(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEsaSiteSiteIDs, err := s.db.EsaSite.Query().
		Where(esasite.ConfigIDEQ(config_id)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询所有站点失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, allEsaSiteSiteID := range allEsaSiteSiteIDs {
		allEsaSiteSiteID := allEsaSiteSiteID
		eg.Go(func() error {
			return s.SyncThisSiteOriginPool(egCtx, config_id, allEsaSiteSiteID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 某个站点 所有源地址池
func (s *ESASync) SyncThisSiteOriginPool(ctx context.Context, config_id int64, site_id int64) (err error) {
	// 查询站点记录
	firstEsaSite, err := s.db.EsaSite.Query().
		Where(esasite.IDEQ(site_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取ESA提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEsaSite.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取ESA提供程序失败: %w", err)
	}

	// 获取所有源地址池
	originPools, err := cdnProvider.OriginPoolsRaw(ctx, firstEsaSite.SiteID)
	if err != nil {
		return fmt.Errorf("获取所有源地址池失败: %w", err)
	}

	// 入库所有源地址池
	for _, originPool := range originPools {
		// 查询源地址池记录
		firstOriginPool, err := s.db.EsaSiteOriginPool.Query().
			Where(esasiteoriginpool.EsaSiteID(firstEsaSite.ID)).
			Where(esasiteoriginpool.PoolIDEQ(tea.Int64Value(originPool.Id))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询源地址池记录失败: %w", err)
			}
			// 创建源地址池记录
			err = s.db.EsaSiteOriginPool.Create().
				SetSyncAt(time.Now()).
				SetPoolID(tea.Int64Value(originPool.Id)).
				SetName(tea.StringValue(originPool.Name)).
				SetRecordName(tea.StringValue(originPool.RecordName)).
				SetReferences([]byte(gojson.JsonEncodeNoError(originPool.References))).
				SetReferenceLBCount(tea.Int32Value(originPool.ReferenceLBCount)).
				SetOrigins([]byte(gojson.JsonEncodeNoError(originPool.Origins))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建源地址池记录失败: %w", err)
			}
		} else {
			// 更新源地址池记录
			err = s.db.EsaSiteOriginPool.UpdateOne(firstOriginPool).
				SetSyncAt(time.Now()).
				SetName(tea.StringValue(originPool.Name)).
				SetRecordName(tea.StringValue(originPool.RecordName)).
				SetReferences([]byte(gojson.JsonEncodeNoError(originPool.References))).
				SetReferenceLBCount(tea.Int32Value(originPool.ReferenceLBCount)).
				SetOrigins([]byte(gojson.JsonEncodeNoError(originPool.Origins))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新源地址池记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有记录
func (s *ESASync) SyncAllSiteRecords(ctx context.Context) (err error) {
	// 查询全部ESA平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部ESA平台失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisSiteRecords(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 所有站点 所有记录
func (s *ESASync) SyncThisSiteRecords(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEsaSiteSiteIDs, err := s.db.EsaSite.Query().
		Where(esasite.ConfigIDEQ(config_id)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询所有站点失败: %w", err)
	}

	// 并发同步
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, allEsaSiteSiteID := range allEsaSiteSiteIDs {
		allEsaSiteSiteID := allEsaSiteSiteID
		eg.Go(func() error {
			return s.SyncThisSiteRecord(egCtx, config_id, allEsaSiteSiteID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 某个站点 所有记录
func (s *ESASync) SyncThisSiteRecord(ctx context.Context, config_id int64, site_id int64) (err error) {
	// 查询站点记录
	firstEsaSite, err := s.db.EsaSite.Query().
		Where(esasite.IDEQ(site_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取ESA提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEsaSite.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取ESA提供程序失败: %w", err)
	}

	// 获取所有解析记录
	records, err := cdnProvider.RecordsRaw(ctx, firstEsaSite.SiteID)
	if err != nil {
		return fmt.Errorf("获取所有解析记录失败: %w", err)
	}

	// 入库所有DNS记录
	for _, record := range records {
		// 查询解析记录
		firstRecord, err := s.db.EsaSiteRecord.Query().
			Where(esasiterecord.EsaSiteIDEQ(firstEsaSite.ID)).
			Where(esasiterecord.RecordIDEQ(tea.Int64Value(record.RecordId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询解析记录记录失败: %w", err)
			}
			// 创建解析记录记录
			err = s.db.EsaSiteRecord.Create().
				SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(record.CreateTime)).Time).
				SetSyncAt(time.Now()).
				SetEsaSiteID(firstEsaSite.ID).
				SetBizName(esasiterecord.BizName(tea.StringValue(record.BizName))).
				SetData([]byte(gojson.JsonEncodeNoError(record.Data))).
				SetProxied(tea.BoolValue(record.Proxied)).
				SetRecordID(tea.Int64Value(record.RecordId)).
				SetRecordSourceType(esasiterecord.RecordSourceType(tea.StringValue(record.RecordSourceType))).
				SetRecordName(tea.StringValue(record.RecordName)).
				SetRecordType(esasiterecord.RecordType(tea.StringValue(record.RecordType))).
				SetTTL(tea.Int64Value(record.Ttl)).
				SetRecordCname(tea.StringValue(record.RecordCname)).
				SetComment(tea.StringValue(record.Comment)).
				SetAuthConf([]byte(gojson.JsonEncodeNoError(record.AuthConf))).
				SetHostPolicy(esasiterecord.HostPolicy(tea.StringValue(record.HostPolicy))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建解析记录记录失败: %w", err)
			}
		} else {
			// 更新解析记录记录
			err = s.db.EsaSiteRecord.UpdateOne(firstRecord).
				SetSyncAt(time.Now()).
				SetBizName(esasiterecord.BizName(tea.StringValue(record.BizName))).
				SetData([]byte(gojson.JsonEncodeNoError(record.Data))).
				SetProxied(tea.BoolValue(record.Proxied)).
				SetRecordSourceType(esasiterecord.RecordSourceType(tea.StringValue(record.RecordSourceType))).
				SetRecordName(tea.StringValue(record.RecordName)).
				SetRecordType(esasiterecord.RecordType(tea.StringValue(record.RecordType))).
				SetTTL(tea.Int64Value(record.Ttl)).
				SetRecordCname(tea.StringValue(record.RecordCname)).
				SetComment(tea.StringValue(record.Comment)).
				SetAuthConf([]byte(gojson.JsonEncodeNoError(record.AuthConf))).
				SetHostPolicy(esasiterecord.HostPolicy(tea.StringValue(record.HostPolicy))).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新解析记录记录失败: %w", err)
			}
		}
	}

	return nil
}
