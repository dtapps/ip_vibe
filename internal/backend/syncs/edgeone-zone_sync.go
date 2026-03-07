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
	"vibe_ip/internal/ent_stat"
	"vibe_ip/internal/ent_stat/edgeonezoneanalysisdata"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/cdn"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gotime"
	"golang.org/x/sync/errgroup"
)

// EdgeOneZoneSync
type EdgeOneZoneSync struct {
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewEdgeOneZoneSync 创建EdgeOne站点同步实例
func NewEdgeOneZoneSync(logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *EdgeOneZoneSync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &EdgeOneZoneSync{
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *EdgeOneZoneSync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取EdgeOne提供程序
func (s *EdgeOneZoneSync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*cdn.TencentEOProvider, error) {

	// 创建EdgeOne CDN提供程序选项
	var cdnProviderOpts = make([]cdn.Option, 0, 1)
	if s.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的EdgeOne CDN配置
	cdnProviderCfg := cdn.NewDefaultTencentEOConfig()
	err := cdnProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析EdgeOne CDN配置失败: %w", err)
	}

	// 创建EdgeOne CDN提供程序
	cdnProvider, err := cdn.NewTencentEOProviderRaw(ctx, s.logger, cdnProviderCfg, cdnProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建EEdgeOne CDN提供程序失败: %w", err)
	}

	return cdnProvider, nil
}

// 同步 全部账号 所有站点
func (s *EdgeOneZoneSync) SyncAllZones(ctx context.Context) (err error) {

	// 查询平台配置
	allPlatformCofingIDs, err := syncDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
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
func (s *EdgeOneZoneSync) SyncThisZones(ctx context.Context, config_id int64) (err error) {

	// 查询平台配置
	firstPlatformCofing, err := syncDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取EdgeOne提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 获取所有站点
	zones, err := cdnProvider.ZonesRaw(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取所有站点失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库所有站点
	for _, zone := range zones {

		// 站点序号
		var edgeoneZoneID int64

		// 查询站点记录
		firstZone, err := syncDb.EdgeOneZone.Query().
			Where(edgeonezone.ConfigIDEQ(config_id)).
			Where(edgeonezone.ZoneIDEQ(utils.PtrToString(zone.ZoneId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询站点记录失败: %w", err)
			}
			// 创建站点记录
			createOne := syncDb.EdgeOneZone.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(zone.CreatedOn)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			createOne.SetZoneID(utils.PtrToString(zone.ZoneId))
			createOne.SetZoneName(utils.PtrToString(zone.ZoneName))
			createOne.SetAliasZoneName(utils.PtrToString(zone.AliasZoneName))
			createOne.SetArea(edgeonezone.Area(utils.PtrToString(zone.Area)))
			createOne.SetType(edgeonezone.Type(utils.PtrToString(zone.Type)))
			createOne.SetResources(json.RawMessage(gojson.JsonEncodeNoError(zone.Resources)))
			createOne.SetNsDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.NSDetail)))
			createOne.SetCnameDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.CNAMEDetail)))
			createOne.SetDnspodDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.DNSPodDetail)))
			createOne.SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status)))
			createOne.SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus)))
			createOne.SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus)))
			createOne.SetLockStatus(edgeonezone.LockStatus(utils.PtrToString(zone.LockStatus)))
			createOne.SetPaused(utils.PtrToBool(zone.Paused))
			saveZone, err := createOne.Save(ctx)
			if err != nil {
				return fmt.Errorf("创建站点记录失败: %w", err)
			}
			edgeoneZoneID = saveZone.ID
		} else {
			// 更新站点记录
			updateOne := syncDb.EdgeOneZone.UpdateOne(firstZone)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetZoneName(utils.PtrToString(zone.ZoneName))
			updateOne.SetAliasZoneName(utils.PtrToString(zone.AliasZoneName))
			updateOne.SetArea(edgeonezone.Area(utils.PtrToString(zone.Area)))
			updateOne.SetType(edgeonezone.Type(utils.PtrToString(zone.Type)))
			updateOne.SetResources(json.RawMessage(gojson.JsonEncodeNoError(zone.Resources)))
			updateOne.SetNsDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.NSDetail)))
			updateOne.SetCnameDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.CNAMEDetail)))
			updateOne.SetDnspodDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.DNSPodDetail)))
			updateOne.SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status)))
			updateOne.SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus)))
			updateOne.SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus)))
			updateOne.SetLockStatus(edgeonezone.LockStatus(utils.PtrToString(zone.LockStatus)))
			updateOne.SetPaused(utils.PtrToBool(zone.Paused))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新站点记录失败: %w", err)
			}
			edgeoneZoneID = firstZone.ID
		}

		s.logger.DebugContext(ctx, "查询流量分析时序数据",
			slog.String("zone", utils.PtrToString(zone.ZoneId)),
		)

		// 查询流量分析时序数据
		analysisDatas, err := cdnProvider.DescribeTimingL7AnalysisData(ctx, utils.PtrToString(zone.ZoneId))
		if err != nil {
			s.logger.ErrorContext(ctx, "查询流量分析时序数据失败",
				slog.Any("err", err),
			)
			return err
		}

		// 按天聚合：date -> map[metricName]value
		dayMetrics := make(map[string]map[string]int64)
		for _, data := range analysisDatas {
			for _, tv := range data.TypeValue {
				for _, detail := range tv.Detail {
					date := gotime.SetCurrentUnix(utils.PtrToInt64(detail.Timestamp)).StartOfDay().SetFormat(gotime.DateFormat)
					if dayMetrics[date] == nil {
						dayMetrics[date] = make(map[string]int64)
					}
					// 直接存 Value
					dayMetrics[date][utils.PtrToString(tv.MetricName)] = utils.PtrToInt64(detail.Value)
				}
			}
		}

		// 构建批量 Create 对象
		createAnalysisDatas := make([]*ent_stat.EdgeOneZoneAnalysisDataCreate, 0, len(dayMetrics))
		for date, metrics := range dayMetrics {
			createAnalysisDatas = append(
				createAnalysisDatas,
				syncStatDb.EdgeOneZoneAnalysisData.Create().
					SetEdgeoneZoneID(edgeoneZoneID).
					SetDate(gotime.SetCurrentParse(date).Time).
					SetL7FlowOutFlux(metrics["l7Flow_outFlux"]).
					SetL7FlowInFlux(metrics["l7Flow_inFlux"]).
					SetL7FlowFlux(metrics["l7Flow_flux"]).
					SetL7FlowOutBandwidth(metrics["l7Flow_outBandwidth"]).
					SetL7FlowInBandwidth(metrics["l7Flow_inBandwidth"]).
					SetL7FlowBandwidth(metrics["l7Flow_bandwidth"]).
					SetL7FlowRequest(metrics["l7Flow_request"]).
					SetL7FlowAvgResponseTime(metrics["l7Flow_responseTime"]).
					SetL7FlowAvgFirstByteResponseTime(metrics["l7Flow_firstByteResponseTime"]),
			)
		}

		// 批量入库统计数据
		err = syncStatDb.EdgeOneZoneAnalysisData.CreateBulk(createAnalysisDatas...).
			OnConflictColumns(
				edgeonezoneanalysisdata.FieldEdgeoneZoneID, edgeonezoneanalysisdata.FieldDate,
			).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			s.logger.ErrorContext(ctx, "批量入库统计数据失败",
				slog.Any("err", err),
			)
			return fmt.Errorf("批量入库统计数据失败: %w", err)
		}

		s.logger.DebugContext(ctx, "批量入库统计数据成功",
			slog.String("zone", utils.PtrToString(zone.ZoneId)),
		)
	}

	return nil
}

// 同步 某个账号 某个站点
func (s *EdgeOneZoneSync) SyncThisZone(ctx context.Context, config_id int64, zone_id int64) (err error) {

	// 查询平台配置
	firstPlatformCofing, err := syncDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 查询站点记录
	firstZone, err := syncDb.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取EdgeOne提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 获取所有站点
	zones, err := cdnProvider.ZonesRaw(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取所有站点失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库所有站点
	for _, zone := range zones {
		if firstZone.ZoneID != utils.PtrToString(zone.ZoneId) {
			continue
		}
		// 更新站点记录
		updateOne := syncDb.EdgeOneZone.UpdateOne(firstZone)
		updateOne.SetSyncAt(time.Now())
		updateOne.SetZoneName(utils.PtrToString(zone.ZoneName))
		updateOne.SetAliasZoneName(utils.PtrToString(zone.AliasZoneName))
		updateOne.SetArea(edgeonezone.Area(utils.PtrToString(zone.Area)))
		updateOne.SetType(edgeonezone.Type(utils.PtrToString(zone.Type)))
		updateOne.SetResources(json.RawMessage(gojson.JsonEncodeNoError(zone.Resources)))
		updateOne.SetNsDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.NSDetail)))
		updateOne.SetCnameDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.CNAMEDetail)))
		updateOne.SetDnspodDetail(json.RawMessage(gojson.JsonEncodeNoError(zone.DNSPodDetail)))
		updateOne.SetStatus(edgeonezone.Status(utils.PtrToString(zone.Status)))
		updateOne.SetCnameStatus(edgeonezone.CnameStatus(utils.PtrToString(zone.CnameStatus)))
		updateOne.SetActiveStatus(edgeonezone.ActiveStatus(utils.PtrToString(zone.ActiveStatus)))
		updateOne.SetLockStatus(edgeonezone.LockStatus(utils.PtrToString(zone.LockStatus)))
		updateOne.SetPaused(utils.PtrToBool(zone.Paused))
		err = updateOne.Exec(ctx)
		if err != nil {
			return fmt.Errorf("更新站点记录失败: %w", err)
		}

		s.logger.DebugContext(ctx, "查询流量分析时序数据",
			slog.String("zone", utils.PtrToString(zone.ZoneId)),
		)

		// 查询流量分析时序数据
		analysisDatas, err := cdnProvider.DescribeTimingL7AnalysisData(ctx, utils.PtrToString(zone.ZoneId))
		if err != nil {
			s.logger.ErrorContext(ctx, "查询流量分析时序数据失败",
				slog.Any("err", err),
			)
			return err
		}

		// 按天聚合：date -> map[metricName]value
		dayMetrics := make(map[string]map[string]int64)
		for _, data := range analysisDatas {
			for _, tv := range data.TypeValue {
				for _, detail := range tv.Detail {
					date := gotime.SetCurrentUnix(utils.PtrToInt64(detail.Timestamp)).StartOfDay().SetFormat(gotime.DateFormat)
					if dayMetrics[date] == nil {
						dayMetrics[date] = make(map[string]int64)
					}
					// 直接存 Value
					dayMetrics[date][utils.PtrToString(tv.MetricName)] = utils.PtrToInt64(detail.Value)
				}
			}
		}

		// 构建批量 Create 对象
		createAnalysisDatas := make([]*ent_stat.EdgeOneZoneAnalysisDataCreate, 0, len(dayMetrics))
		for date, metrics := range dayMetrics {
			createAnalysisDatas = append(
				createAnalysisDatas,
				syncStatDb.EdgeOneZoneAnalysisData.Create().
					SetEdgeoneZoneID(firstZone.ID).
					SetDate(gotime.SetCurrentParse(date).Time).
					SetL7FlowOutFlux(metrics["l7Flow_outFlux"]).
					SetL7FlowInFlux(metrics["l7Flow_inFlux"]).
					SetL7FlowFlux(metrics["l7Flow_flux"]).
					SetL7FlowOutBandwidth(metrics["l7Flow_outBandwidth"]).
					SetL7FlowInBandwidth(metrics["l7Flow_inBandwidth"]).
					SetL7FlowBandwidth(metrics["l7Flow_bandwidth"]).
					SetL7FlowRequest(metrics["l7Flow_request"]).
					SetL7FlowAvgResponseTime(metrics["l7Flow_responseTime"]).
					SetL7FlowAvgFirstByteResponseTime(metrics["l7Flow_firstByteResponseTime"]),
			)
		}

		// 批量入库统计数据
		err = syncStatDb.EdgeOneZoneAnalysisData.CreateBulk(createAnalysisDatas...).
			OnConflictColumns(
				edgeonezoneanalysisdata.FieldEdgeoneZoneID, edgeonezoneanalysisdata.FieldDate,
			).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			s.logger.ErrorContext(ctx, "批量入库统计数据失败",
				slog.Any("err", err),
			)
			return fmt.Errorf("批量入库统计数据失败: %w", err)
		}

		s.logger.DebugContext(ctx, "批量入库统计数据成功",
			slog.String("zone", utils.PtrToString(zone.ZoneId)),
		)
	}

	return nil
}

// 同步 全部账号 所有站点 所有源站组
func (s *EdgeOneZoneSync) SyncAllZoneOriginGroups(ctx context.Context) (err error) {

	// 查询平台配置
	allPlatformCofingIDs, err := syncDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
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
func (s *EdgeOneZoneSync) SyncThisZoneOriginGroups(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEdgeOneZoneIDs, err := syncDb.EdgeOneZone.Query().
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
func (s *EdgeOneZoneSync) SyncThisZoneOriginGroup(ctx context.Context, config_id int64, zone_id int64) (err error) {

	// 查询站点记录
	firstEdgeOneZone, err := syncDb.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEdgeOneZone.Edges.Config.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取EdgeOne提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 获取所有源站组
	originGroups, err := cdnProvider.OriginGroupsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取所有源站组失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库所有源站组
	for _, originGroup := range originGroups {
		// 查询源站组记录
		firstOriginGroup, err := syncDb.EdgeOneZoneOriginGroup.Query().
			Where(edgeonezoneorigingroup.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezoneorigingroup.GroupIDEQ(utils.PtrToString(originGroup.GroupId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询源站组记录失败: %w", err)
			}
			// 创建源站组记录
			createOne := syncDb.EdgeOneZoneOriginGroup.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(originGroup.CreateTime)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEdgeoneZoneID(firstEdgeOneZone.ID)
			createOne.SetGroupID(utils.PtrToString(originGroup.GroupId))
			createOne.SetName(utils.PtrToString(originGroup.Name))
			if utils.PtrToString(originGroup.Type) != "" {
				createOne.SetType(edgeonezoneorigingroup.Type(utils.PtrToString(originGroup.Type)))
			}
			createOne.SetRecords(json.RawMessage(gojson.JsonEncodeNoError(originGroup.Records)))
			createOne.SetReferences(json.RawMessage(gojson.JsonEncodeNoError(originGroup.References)))
			createOne.SetHostHeader(utils.PtrToString(originGroup.HostHeader))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建源站组记录失败: %w", err)
			}
		} else {
			// 更新源站组记录
			updateOne := syncDb.EdgeOneZoneOriginGroup.UpdateOne(firstOriginGroup)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetName(utils.PtrToString(originGroup.Name))
			if utils.PtrToString(originGroup.Type) != "" {
				updateOne.SetType(edgeonezoneorigingroup.Type(utils.PtrToString(originGroup.Type)))
			}
			updateOne.SetRecords(json.RawMessage(gojson.JsonEncodeNoError(originGroup.Records)))
			updateOne.SetReferences(json.RawMessage(gojson.JsonEncodeNoError(originGroup.References)))
			updateOne.SetHostHeader(utils.PtrToString(originGroup.HostHeader))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新源站组记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 所有站点 所有记录
func (s *EdgeOneZoneSync) SyncAllZoneRecords(ctx context.Context) (err error) {

	// 查询平台配置
	allPlatformCofingIDs, err := syncDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
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
func (s *EdgeOneZoneSync) SyncThisZoneRecords(ctx context.Context, config_id int64) (err error) {

	// 查询所有站点
	allEdgeOneZoneIDs, err := syncDb.EdgeOneZone.Query().
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
func (s *EdgeOneZoneSync) SyncThisZoneRecord(ctx context.Context, config_id int64, zone_id int64) (err error) {

	// 查询站点记录
	firstEdgeOneZone, err := syncDb.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(zone_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询站点记录失败: %w", err)
	}

	// 获取EdgeOne提供程序
	cdnProvider, err := s.GetApiClient(ctx, firstEdgeOneZone.Edges.Config.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取EdgeOne提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 获取所有DNS记录
	dnsRecords, err := cdnProvider.DnsRecordsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取所有DNS记录失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库所有DNS记录
	for _, dnsRecord := range dnsRecords {
		// 查询DNS记录
		firstDNSRecord, err := syncDb.EdgeOneZoneDNSRecord.Query().
			Where(edgeonezonednsrecord.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezonednsrecord.RecordIDEQ(utils.PtrToString(dnsRecord.RecordId))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询DNS记录记录失败: %w", err)
			}
			// 创建DNS记录记录
			createOne := syncDb.EdgeOneZoneDNSRecord.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(dnsRecord.CreatedOn)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEdgeoneZoneID(firstEdgeOneZone.ID)
			createOne.SetRecordID(utils.PtrToString(dnsRecord.RecordId))
			createOne.SetName(utils.PtrToString(dnsRecord.Name))
			if utils.PtrToString(dnsRecord.Type) != "" {
				createOne.SetType(edgeonezonednsrecord.Type(utils.PtrToString(dnsRecord.Type)))
			}
			createOne.SetLocation(utils.PtrToString(dnsRecord.Location))
			createOne.SetContent(utils.PtrToString(dnsRecord.Content))
			createOne.SetTTL(utils.PtrToInt64(dnsRecord.TTL))
			createOne.SetWeight(utils.PtrToInt64(dnsRecord.Weight))
			createOne.SetPriority(utils.PtrToInt64(dnsRecord.Priority))
			if utils.PtrToString(dnsRecord.Status) != "" {
				createOne.SetStatus(edgeonezonednsrecord.Status(utils.PtrToString(dnsRecord.Status)))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建DNS记录记录失败: %w", err)
			}
		} else {
			// 更新DNS记录记录
			updateOne := syncDb.EdgeOneZoneDNSRecord.UpdateOne(firstDNSRecord)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetName(utils.PtrToString(dnsRecord.Name))
			if utils.PtrToString(dnsRecord.Type) != "" {
				updateOne.SetType(edgeonezonednsrecord.Type(utils.PtrToString(dnsRecord.Type)))
			}
			updateOne.SetLocation(utils.PtrToString(dnsRecord.Location))
			updateOne.SetContent(utils.PtrToString(dnsRecord.Content))
			updateOne.SetTTL(utils.PtrToInt64(dnsRecord.TTL))
			updateOne.SetWeight(utils.PtrToInt64(dnsRecord.Weight))
			updateOne.SetPriority(utils.PtrToInt64(dnsRecord.Priority))
			if utils.PtrToString(dnsRecord.Status) != "" {
				updateOne.SetStatus(edgeonezonednsrecord.Status(utils.PtrToString(dnsRecord.Status)))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新DNS记录记录失败: %w", err)
			}
		}
	}

	// 获取所有加速域名
	accelerationDomains, err := cdnProvider.AccelerationDomainsRaw(ctx, firstEdgeOneZone.ZoneID)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取所有加速域名失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库所有加速域名
	for _, accelerationDomain := range accelerationDomains {
		// 查询加速域名记录
		firstAccelerationDomain, err := syncDb.EdgeOneZoneAccelerationDomain.Query().
			Where(edgeonezoneaccelerationdomain.EdgeoneZoneIDEQ(firstEdgeOneZone.ID)).
			Where(edgeonezoneaccelerationdomain.DomainNameEQ(utils.PtrToString(accelerationDomain.DomainName))).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询加速域名记录失败: %w", err)
			}
			// 创建加速域名记录
			createOne := syncDb.EdgeOneZoneAccelerationDomain.Create()
			createOne.SetRemoteCreatedAt(gotime.SetCurrentParse(utils.PtrToString(accelerationDomain.CreatedOn)).Time)
			createOne.SetSyncAt(time.Now())
			createOne.SetEdgeoneZoneID(firstEdgeOneZone.ID)
			createOne.SetDomainName(utils.PtrToString(accelerationDomain.DomainName))
			if utils.PtrToString(accelerationDomain.DomainStatus) != "" {
				createOne.SetDomainStatus(edgeonezoneaccelerationdomain.DomainStatus(utils.PtrToString(accelerationDomain.DomainStatus)))
			}
			createOne.SetCname(utils.PtrToString(accelerationDomain.Cname))
			if utils.PtrToString(accelerationDomain.IPv6Status) != "" {
				createOne.SetIpv6Status(edgeonezoneaccelerationdomain.Ipv6Status(utils.PtrToString(accelerationDomain.IPv6Status)))
			}
			if utils.PtrToString(accelerationDomain.IdentificationStatus) != "" {
				createOne.SetIdentificationStatus(edgeonezoneaccelerationdomain.IdentificationStatus(utils.PtrToString(accelerationDomain.IdentificationStatus)))
			}
			createOne.SetOwnershipVerification(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.OwnershipVerification)))
			createOne.SetOriginDetail(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.OriginDetail)))
			if utils.PtrToString(accelerationDomain.OriginProtocol) != "" {
				createOne.SetOriginProtocol(edgeonezoneaccelerationdomain.OriginProtocol(utils.PtrToString(accelerationDomain.OriginProtocol)))
			}
			createOne.SetHTTPOriginPort(utils.PtrToUint64(accelerationDomain.HttpOriginPort))
			createOne.SetHTTPSOriginPort(utils.PtrToUint64(accelerationDomain.HttpsOriginPort))
			createOne.SetCertificate(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.Certificate)))
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建加速域名记录失败: %w", err)
			}
		} else {
			// 更新加速域名记录
			updateOne := syncDb.EdgeOneZoneAccelerationDomain.UpdateOne(firstAccelerationDomain)
			updateOne.SetSyncAt(time.Now())
			if utils.PtrToString(accelerationDomain.DomainStatus) != "" {
				updateOne.SetDomainStatus(edgeonezoneaccelerationdomain.DomainStatus(utils.PtrToString(accelerationDomain.DomainStatus)))
			}
			updateOne.SetCname(utils.PtrToString(accelerationDomain.Cname))
			if utils.PtrToString(accelerationDomain.IPv6Status) != "" {
				updateOne.SetIpv6Status(edgeonezoneaccelerationdomain.Ipv6Status(utils.PtrToString(accelerationDomain.IPv6Status)))
			}
			if utils.PtrToString(accelerationDomain.IdentificationStatus) != "" {
				updateOne.SetIdentificationStatus(edgeonezoneaccelerationdomain.IdentificationStatus(utils.PtrToString(accelerationDomain.IdentificationStatus)))
			}
			updateOne.SetOwnershipVerification(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.OwnershipVerification)))
			updateOne.SetOriginDetail(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.OriginDetail)))
			if utils.PtrToString(accelerationDomain.OriginProtocol) != "" {
				updateOne.SetOriginProtocol(edgeonezoneaccelerationdomain.OriginProtocol(utils.PtrToString(accelerationDomain.OriginProtocol)))
			}
			updateOne.SetHTTPOriginPort(utils.PtrToUint64(accelerationDomain.HttpOriginPort))
			updateOne.SetHTTPSOriginPort(utils.PtrToUint64(accelerationDomain.HttpsOriginPort))
			updateOne.SetCertificate(json.RawMessage(gojson.JsonEncodeNoError(accelerationDomain.Certificate)))
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新加速域名记录失败: %w", err)
			}
		}
	}

	return nil
}
