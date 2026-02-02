package syncs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/esasite"
	"vibe_ip/internal/ent/esasiteedgecontainerrecord"
	"vibe_ip/internal/ent/esasiteedgeroutinerecord"
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

// ESASiteSync ESA站点同步
type ESASiteSync struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewESASiteSync 创建ESA站点同步实例
func NewESASiteSync(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *ESASiteSync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &ESASiteSync{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *ESASiteSync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取ESA提供程序
func (s *ESASiteSync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*cdn.AliyunESAProvider, error) {

	// 创建ESA CDN提供程序选项
	var cdnProviderOpts = make([]cdn.Option, 0, 1)
	if s.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的ESA CDN配置
	cdnProviderCfg := cdn.NewDefaultAliyunESAConfig()
	err := cdnProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析ESA CDN配置失败: %w", err)
	}

	// 创建ESA CDN提供程序
	cdnProvider, err := cdn.NewAliyunESAProviderRaw(ctx, s.logger, cdnProviderCfg, cdnProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建ESA CDN提供程序失败: %w", err)
	}

	return cdnProvider, nil
}

// 同步 全部账号 所有站点
func (s *ESASiteSync) SyncAllSites(ctx context.Context) (err error) {
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
func (s *ESASiteSync) SyncThisSites(ctx context.Context, config_id int64) (err error) {
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
		return err
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
			createOne := s.db.EsaSite.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(site.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			if tea.StringValue(site.AccessType) != "" {
				createOne.SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType)))
			}
			createOne.SetCnameZone(tea.StringValue(site.CnameZone))
			if tea.StringValue(site.Coverage) != "" {
				createOne.SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage)))
			}
			createOne.SetInstanceID(tea.StringValue(site.InstanceId))
			createOne.SetNameServerList(tea.StringValue(site.NameServerList))
			createOne.SetPlanName(tea.StringValue(site.PlanName))
			createOne.SetPlanSpecName(tea.StringValue(site.PlanSpecName))
			createOne.SetSiteID(tea.Int64Value(site.SiteId))
			createOne.SetSiteName(tea.StringValue(site.SiteName))
			if tea.StringValue(site.Status) != "" {
				createOne.SetStatus(esasite.Status(tea.StringValue(site.Status)))
			}
			createOne.SetOfflineReason(tea.StringValue(site.OfflineReason))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			updateOne := s.db.EsaSite.UpdateOne(firstSite)
			updateOne.SetSyncAt(time.Now())
			if tea.StringValue(site.AccessType) != "" {
				updateOne.SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType)))
			}
			updateOne.SetCnameZone(tea.StringValue(site.CnameZone))
			if tea.StringValue(site.Coverage) != "" {
				updateOne.SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage)))
			}
			updateOne.SetInstanceID(tea.StringValue(site.InstanceId))
			updateOne.SetNameServerList(tea.StringValue(site.NameServerList))
			updateOne.SetPlanName(tea.StringValue(site.PlanName))
			updateOne.SetPlanSpecName(tea.StringValue(site.PlanSpecName))
			updateOne.SetSiteName(tea.StringValue(site.SiteName))
			if tea.StringValue(site.Status) != "" {
				updateOne.SetStatus(esasite.Status(tea.StringValue(site.Status)))
			}
			updateOne.SetOfflineReason(tea.StringValue(site.OfflineReason))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 某个账号 某个站点
func (s *ESASiteSync) SyncThisSite(ctx context.Context, config_id int64, site_id int64) (err error) {
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
		return err
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
			createOne := s.db.EsaSite.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(site.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			if tea.StringValue(site.AccessType) != "" {
				createOne.SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType)))
			}
			createOne.SetCnameZone(tea.StringValue(site.CnameZone))
			if tea.StringValue(site.Coverage) != "" {
				createOne.SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage)))
			}
			createOne.SetInstanceID(tea.StringValue(site.InstanceId))
			createOne.SetNameServerList(tea.StringValue(site.NameServerList))
			createOne.SetPlanName(tea.StringValue(site.PlanName))
			createOne.SetPlanSpecName(tea.StringValue(site.PlanSpecName))
			createOne.SetSiteID(tea.Int64Value(site.SiteId))
			createOne.SetSiteName(tea.StringValue(site.SiteName))
			if tea.StringValue(site.Status) != "" {
				createOne.SetStatus(esasite.Status(tea.StringValue(site.Status)))
			}
			createOne.SetOfflineReason(tea.StringValue(site.OfflineReason))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
		} else {
			// 更新站点记录
			updateOne := s.db.EsaSite.UpdateOne(firstSite)
			updateOne.SetSyncAt(time.Now())
			if tea.StringValue(site.AccessType) != "" {
				updateOne.SetAccessType(esasite.AccessType(tea.StringValue(site.AccessType)))
			}
			updateOne.SetCnameZone(tea.StringValue(site.CnameZone))
			if tea.StringValue(site.Coverage) != "" {
				updateOne.SetCoverage(esasite.Coverage(tea.StringValue(site.Coverage)))
			}
			updateOne.SetInstanceID(tea.StringValue(site.InstanceId))
			updateOne.SetNameServerList(tea.StringValue(site.NameServerList))
			updateOne.SetPlanName(tea.StringValue(site.PlanName))
			updateOne.SetPlanSpecName(tea.StringValue(site.PlanSpecName))
			updateOne.SetSiteName(tea.StringValue(site.SiteName))
			if tea.StringValue(site.Status) != "" {
				updateOne.SetStatus(esasite.Status(tea.StringValue(site.Status)))
			}
			updateOne.SetOfflineReason(tea.StringValue(site.OfflineReason))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有源地址池
func (s *ESASiteSync) SyncAllSiteOriginPools(ctx context.Context) (err error) {
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
func (s *ESASiteSync) SyncThisSiteOriginPools(ctx context.Context, config_id int64) (err error) {

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
func (s *ESASiteSync) SyncThisSiteOriginPool(ctx context.Context, config_id int64, site_id int64) (err error) {
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
		return err
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
			createOne := s.db.EsaSiteOriginPool.Create()
			createOne.SetSyncAt(time.Now())
			createOne.SetEsaSiteID(firstEsaSite.ID)
			createOne.SetPoolID(tea.Int64Value(originPool.Id))
			createOne.SetName(tea.StringValue(originPool.Name))
			createOne.SetRecordName(tea.StringValue(originPool.RecordName))
			createOne.SetEnabled(tea.BoolValue(originPool.Enabled))
			createOne.SetReferences(json.RawMessage(gojson.JsonEncodeNoError(originPool.References)))
			createOne.SetReferenceLBCount(tea.Int32Value(originPool.ReferenceLBCount))
			createOne.SetOrigins(json.RawMessage(gojson.JsonEncodeNoError(originPool.Origins)))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建源地址池记录失败: %w", err)
			}
		} else {
			// 更新源地址池记录
			updateOne := s.db.EsaSiteOriginPool.UpdateOne(firstOriginPool)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetName(tea.StringValue(originPool.Name))
			updateOne.SetRecordName(tea.StringValue(originPool.RecordName))
			updateOne.SetEnabled(tea.BoolValue(originPool.Enabled))
			updateOne.SetReferences(json.RawMessage(gojson.JsonEncodeNoError(originPool.References)))
			updateOne.SetReferenceLBCount(tea.Int32Value(originPool.ReferenceLBCount))
			updateOne.SetOrigins(json.RawMessage(gojson.JsonEncodeNoError(originPool.Origins)))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新源地址池记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有记录
func (s *ESASiteSync) SyncAllSiteRecords(ctx context.Context) (err error) {
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
func (s *ESASiteSync) SyncThisSiteRecords(ctx context.Context, config_id int64) (err error) {

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
func (s *ESASiteSync) SyncThisSiteRecord(ctx context.Context, config_id int64, site_id int64) (err error) {
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
		return err
	}

	// 查询解析记录列表
	records, err := cdnProvider.RecordsRaw(ctx, firstEsaSite.SiteID)
	if err != nil {
		return fmt.Errorf("查询解析记录列表失败: %w", err)
	}
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
			createOne := s.db.EsaSiteRecord.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(record.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEsaSiteID(firstEsaSite.ID)
			if tea.StringValue(record.BizName) != "" {
				createOne.SetBizName(esasiterecord.BizName(tea.StringValue(record.BizName)))
			}
			createOne.SetData(json.RawMessage(gojson.JsonEncodeNoError(record.Data)))
			createOne.SetProxied(tea.BoolValue(record.Proxied))
			createOne.SetRecordID(tea.Int64Value(record.RecordId))
			if tea.StringValue(record.RecordSourceType) != "" {
				createOne.SetRecordSourceType(esasiterecord.RecordSourceType(tea.StringValue(record.RecordSourceType)))
			}
			createOne.SetRecordName(tea.StringValue(record.RecordName))
			if tea.StringValue(record.RecordType) != "" {
				createOne.SetRecordType(esasiterecord.RecordType(tea.StringValue(record.RecordType)))
			}
			createOne.SetTTL(tea.Int64Value(record.Ttl))
			createOne.SetRecordCname(tea.StringValue(record.RecordCname))
			createOne.SetComment(tea.StringValue(record.Comment))
			createOne.SetAuthConf(json.RawMessage(gojson.JsonEncodeNoError(record.AuthConf)))
			if tea.StringValue(record.HostPolicy) != "" {
				createOne.SetHostPolicy(esasiterecord.HostPolicy(tea.StringValue(record.HostPolicy)))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建解析记录记录失败: %w", err)
			}
		} else {
			// 更新解析记录记录
			updateOne := s.db.EsaSiteRecord.UpdateOne(firstRecord)
			updateOne.SetSyncAt(time.Now())
			if tea.StringValue(record.BizName) != "" {
				updateOne.SetBizName(esasiterecord.BizName(tea.StringValue(record.BizName)))
			}
			updateOne.SetData(json.RawMessage(gojson.JsonEncodeNoError(record.Data)))
			updateOne.SetProxied(tea.BoolValue(record.Proxied))
			if tea.StringValue(record.RecordSourceType) != "" {
				updateOne.SetRecordSourceType(esasiterecord.RecordSourceType(tea.StringValue(record.RecordSourceType)))
			}
			updateOne.SetRecordName(tea.StringValue(record.RecordName))
			if tea.StringValue(record.RecordType) != "" {
				updateOne.SetRecordType(esasiterecord.RecordType(tea.StringValue(record.RecordType)))
			}
			updateOne.SetTTL(tea.Int64Value(record.Ttl))
			updateOne.SetRecordCname(tea.StringValue(record.RecordCname))
			updateOne.SetComment(tea.StringValue(record.Comment))
			updateOne.SetAuthConf(json.RawMessage(gojson.JsonEncodeNoError(record.AuthConf)))
			if tea.StringValue(record.HostPolicy) != "" {
				updateOne.SetHostPolicy(esasiterecord.HostPolicy(tea.StringValue(record.HostPolicy)))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新解析记录记录失败: %w", err)
			}
		}
	}

	// 查询边缘路由记录列表
	routineRecords, err := cdnProvider.EdgeRoutineRecordsRaw(ctx, firstEsaSite.SiteID)
	if err != nil {
		return fmt.Errorf("查询边缘路由记录列表失败: %w", err)
	}
	for _, routineRecord := range routineRecords {
		// 查询解析记录
		firstRecord, err := s.db.EsaSiteEdgeRoutineRecord.Query().
			Where(esasiteedgeroutinerecord.EsaSiteIDEQ(firstEsaSite.ID)).
			Where(esasiteedgeroutinerecord.RecordNameEQ(tea.StringValue(routineRecord.RecordName))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询解析记录记录失败: %w", err)
			}
			// 创建解析记录记录
			createOne := s.db.EsaSiteEdgeRoutineRecord.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(routineRecord.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEsaSiteID(firstEsaSite.ID)
			createOne.SetRecordName(tea.StringValue(routineRecord.RecordName))
			createOne.SetRecordCname(tea.StringValue(routineRecord.RecordCname))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建解析记录记录失败: %w", err)
			}
		} else {
			// 更新解析记录记录
			updateOne := s.db.EsaSiteEdgeRoutineRecord.UpdateOne(firstRecord)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetRecordCname(tea.StringValue(routineRecord.RecordCname))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新解析记录记录失败: %w", err)
			}
		}
	}

	// 查询边缘容器记录列表
	containerRecords, err := cdnProvider.EdgeContainerRecordsRaw(ctx, firstEsaSite.SiteID)
	if err != nil {
		return fmt.Errorf("查询边缘容器记录列表失败: %w", err)
	}
	for _, containerRecord := range containerRecords {
		// 查询解析记录
		firstRecord, err := s.db.EsaSiteEdgeContainerRecord.Query().
			Where(esasiteedgecontainerrecord.EsaSiteIDEQ(firstEsaSite.ID)).
			Where(esasiteedgecontainerrecord.RecordNameEQ(tea.StringValue(containerRecord.RecordName))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询解析记录记录失败: %w", err)
			}
			// 创建解析记录记录
			createOne := s.db.EsaSiteEdgeContainerRecord.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(tea.StringValue(containerRecord.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEsaSiteID(firstEsaSite.ID)
			createOne.SetRecordName(tea.StringValue(containerRecord.RecordName))
			createOne.SetRecordCname(tea.StringValue(containerRecord.RecordCname))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建解析记录记录失败: %w", err)
			}
		} else {
			// 更新解析记录记录
			updateOne := s.db.EsaSiteEdgeContainerRecord.UpdateOne(firstRecord)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetRecordCname(tea.StringValue(containerRecord.RecordCname))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新解析记录记录失败: %w", err)
			}
		}
	}

	return nil
}
