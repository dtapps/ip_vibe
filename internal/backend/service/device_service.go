package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/ent_log/portforwardstats"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"golang.org/x/sync/errgroup"
)

// DeviceService 设备服务接口
type DeviceService interface {
	BuildResponse(ctx context.Context, item *ent.Device) (*types.GetDeviceResponse, error)
	GetDevices(ctx context.Context) ([]*types.GetDeviceResponse, error)
	GetDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error)
	CreateDevice(ctx context.Context, request *types.CreateDeviceRequest) (*types.GetDeviceRequest, error)
	UpdateDevice(ctx context.Context, request *types.UpdateDeviceRequest) (*types.GetDeviceRequest, error)
	DeleteDevice(ctx context.Context, id int64) error
	BatchDeleteDevice(ctx context.Context, ids []int64) error
	GetDeviceIpHistory(ctx context.Context, id int64) ([]*types.GetDeviceIpHistoryResponse, error)
	HealthcheckDevice(ctx context.Context, id int64) ([][]ToPairsStruct, error)
	GetDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDevicePortForwardStatsRequest) (*types.GetDevicePortForwardStatsResponse, error)
}

// DeviceServiceImpl 设备服务实现
type DeviceServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewDeviceService 创建设备服务实例
func NewDeviceService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, debug bool) DeviceService {
	return &DeviceServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// BuildResponse 将实体转换为响应
func (s *DeviceServiceImpl) BuildResponse(ctx context.Context, item *ent.Device) (*types.GetDeviceResponse, error) {
	resp := &types.GetDeviceResponse{
		ID:              item.ID,
		Code:            item.Code,
		Name:            item.Name,
		Description:     item.Description,
		IPV4:            item.IPV4,
		IPV6:            item.IPV6,
		LastReportedAt:  item.LastReportedAt,
		LastHeartbeatAt: item.LastHeartbeatAt,
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDevice)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	// 默认离线
	resp.Online = false

	lastHB := item.LastHeartbeatAt
	invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime // 单位：分钟

	// 如果从未上报过心跳，保持离线（false）
	if lastHB.IsZero() {
		resp.Online = false
		return resp, nil
	}

	// 如果未启用离线检测（invalidTime <= 0），则视为始终在线
	if invalidTime <= 0 {
		resp.Online = true
		return resp, nil
	}

	// 启用了离线检测：判断是否在有效时间窗口内
	threshold := time.Duration(invalidTime) * time.Minute
	if time.Since(lastHB) <= threshold {
		resp.Online = true // 在线
	} else {
		resp.Online = false // 离线（超时）
	}

	return resp, nil
}

// GetDevices 获取全部设备
func (s *DeviceServiceImpl) GetDevices(ctx context.Context) ([]*types.GetDeviceResponse, error) {

	// 查询全部设备
	allDevice, err := s.db.Device.Query().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	lists := make([]*types.GetDeviceResponse, 0, len(allDevice))
	for _, item := range allDevice {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetDevice 获取单个设备
func (s *DeviceServiceImpl) GetDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstDevice)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateDevice 创建设备
func (s *DeviceServiceImpl) CreateDevice(ctx context.Context, request *types.CreateDeviceRequest) (*types.GetDeviceRequest, error) {
	// 创建设备
	saveDevice, err := s.db.Device.Create().
		SetCode(request.Code).
		SetName(request.Name).
		SetDescription(request.Description).
		SetIPV4(request.IPV4).
		SetIPV6(request.IPV6).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建设备失败: %w", err)
	}

	// 创建通知订阅 (保持原始逻辑)
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDevice, saveDevice.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetDeviceRequest{
		ID: saveDevice.ID,
	}, nil
}

// UpdateDevice 更新设备
func (s *DeviceServiceImpl) UpdateDevice(ctx context.Context, request *types.UpdateDeviceRequest) (*types.GetDeviceRequest, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	// 更新设备
	saveDevice, err := s.db.Device.UpdateOne(firstDevice).
		SetName(request.Name).
		SetDescription(request.Description).
		SetIPV4(request.IPV4).
		SetIPV6(request.IPV6).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新设备失败: %w", err)
	}

	// 更新通知订阅 (保持原始逻辑)
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDevice, saveDevice.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetDeviceRequest{
		ID: saveDevice.ID,
	}, nil
}

// DeleteDevice 删除设备
func (s *DeviceServiceImpl) DeleteDevice(ctx context.Context, id int64) error {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询设备失败: %w", err)
	}

	// 删除设备
	err = s.db.Device.DeleteOne(firstDevice).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除设备失败: %w", err)
	}

	// 删除通知订阅 (保持原始逻辑)
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDevice, firstDevice.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// BatchDeleteDevice 批量删除设备
func (s *DeviceServiceImpl) BatchDeleteDevice(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询设备
	allDevice, err := tx.Device.Query().
		Where(device.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询设备失败: %w", err)
	}
	if len(allDevice) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询设备失败: 数量 %d 与请求数量 %d 不匹配", len(allDevice), len(ids))
	}

	// 删除设备
	_, err = tx.Device.Delete().
		Where(device.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除设备失败: %w", err)
	}

	// 删除通知订阅
	for _, item := range allDevice {
		err = webhook_sub_helper.DeleteTx(ctx, tx, webhooksubscription.ObjTypeDevice, item.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("删除通知订阅失败: %w", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetDeviceIpHistory 获取设备的IP历史记录
func (s *DeviceServiceImpl) GetDeviceIpHistory(ctx context.Context, id int64) ([]*types.GetDeviceIpHistoryResponse, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 查询设备IP历史记录
	firstDeviceReports, err := s.logDb.DeviceIpReport.Query().
		Where(deviceipreport.DeviceIDEQ(firstDevice.ID)).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询设备IP历史记录失败: %w", err)
	}

	// 设备列表
	ips := make([]*types.GetDeviceIpHistoryResponse, 0, len(firstDeviceReports))
	for _, item := range firstDeviceReports {
		ips = append(ips, &types.GetDeviceIpHistoryResponse{
			IPV4:      item.IPV4,
			IPV6:      item.IPV6,
			CreatedAt: item.CreatedAt,
		})
	}
	return ips, nil
}

// HealthcheckDevice 检查设备健康状态
func (s *DeviceServiceImpl) HealthcheckDevice(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发 (保持原始限制)

	if firstDevice.IPV4 != "" {
		// 检测 IP 是否可通过 HTTP/HTTPS 访问
		eg.Go(func() error {
			healthCheckResult, _ := CheckIPHealth(egCtx, firstDevice.IPV4, 80, 443)
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}
	if firstDevice.IPV6 != "" {
		// 检测 IP 是否可通过 HTTP/HTTPS 访问
		eg.Go(func() error {
			healthCheckResult, _ := CheckIPHealth(egCtx, firstDevice.IPV6, 80, 443)
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return nil, fmt.Errorf("设备健康状态检查失败: %w", err)
	}

	return results, nil
}

// GetDevicePortForwardStats 获取设备端口转发统计数据
func (s *DeviceServiceImpl) GetDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDevicePortForwardStatsRequest) (*types.GetDevicePortForwardStatsResponse, error) {
	// 查询设备
	firstDevice, err := s.db.Device.Query().
		Where(device.IDEQ(id)).
		WithPortForwards().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}

	// 端口列表
	portIDs := make([]int64, 0, len(firstDevice.Edges.PortForwards))
	if portForwards, err := firstDevice.Edges.PortForwardsOrErr(); err == nil {
		for _, item := range portForwards {
			portIDs = append(portIDs, int64(item.ID))
		}
	}

	// 计算时间范围
	startTime := s.getStartTimeFromRange(request.Range)

	s.logger.InfoContext(ctx, "获取设备端口转发统计数据",
		slog.Int64("ID", id),
		slog.String("Range", request.Range),
		slog.String("StartTime", startTime.Format(time.DateTime)),
		slog.Any("PortIDs", portIDs),
	)

	// 查询端口转发统计记录
	allStats, err := s.logDb.PortForwardStats.Query().
		Where(portforwardstats.DeviceIDEQ(firstDevice.ID)).
		Where(portforwardstats.CreatedAtGTE(startTime)).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询统计记录失败: %w", err)
	}

	// 计算统计数据
	result := s.calculatePortForwardStats(allStats)

	result.PortIDs = portIDs

	return result, nil
}

// 根据时间范围计算起始时间
func (s *DeviceServiceImpl) getStartTimeFromRange(rangeStr string) time.Time {
	startTime := time.Now().Add(-6 * time.Hour) // 默认 6 小时
	if rangeStr == "1h" {
		startTime = time.Now().Add(-1 * time.Hour)
	} else if rangeStr == "6h" {
		startTime = time.Now().Add(-6 * time.Hour)
	} else if rangeStr == "24h" {
		startTime = time.Now().Add(-24 * time.Hour)
	} else if rangeStr == "7d" {
		startTime = time.Now().AddDate(0, 0, -7)
	}
	return startTime
}

// 计算端口转发统计数据
func (s *DeviceServiceImpl) calculatePortForwardStats(raw []*ent_log.PortForwardStats) *types.GetDevicePortForwardStatsResponse {
	count := len(raw)
	if count == 0 {
		return &types.GetDevicePortForwardStatsResponse{}
	}

	// 1. 按时间戳聚合：将同一秒内所有端口的数据相加
	type aggregatePoint struct {
		ActiveConn int64
		BytesIn    int64
		BytesOut   int64
		Errors     int64
		CreatedAt  time.Time
	}

	// 使用 Map 按 CreatedAt 分组聚合
	timeMap := make(map[int64]*aggregatePoint)
	for _, item := range raw {
		ts := item.CreatedAt.Unix() // 按秒聚合
		if _, ok := timeMap[ts]; !ok {
			timeMap[ts] = &aggregatePoint{CreatedAt: item.CreatedAt}
		}
		timeMap[ts].ActiveConn += item.ActiveConn
		timeMap[ts].BytesIn += item.BytesIn
		timeMap[ts].BytesOut += item.BytesOut
		timeMap[ts].Errors += item.Errors
	}

	// 2. 将聚合后的数据转为有序切片（升序）
	sortedPoints := make([]*aggregatePoint, 0, len(timeMap))
	for _, p := range timeMap {
		sortedPoints = append(sortedPoints, p)
	}
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].CreatedAt.Before(sortedPoints[j].CreatedAt)
	})

	// 3. 开始计算速率和增量
	var res types.GetDevicePortForwardStatsResponse
	aggCount := len(sortedPoints)

	// 最新快照即为最后一个点
	lastPoint := sortedPoints[aggCount-1]
	res.Latest = types.GetDevicePortForwardStatsLatest{
		ActiveConn: lastPoint.ActiveConn,
		BytesIn:    lastPoint.BytesIn,
		BytesOut:   lastPoint.BytesOut,
		Errors:     lastPoint.Errors,
	}

	rates := make([]types.GetDevicePortForwardStatsRates, 0, aggCount)
	var totalIn, totalOut, totalErr int64

	for i := 0; i < aggCount; i++ {
		curr := sortedPoints[i]
		var inKbps, outKbps float64

		if i > 0 {
			prev := sortedPoints[i-1]
			duration := curr.CreatedAt.Sub(prev.CreatedAt).Seconds()

			if duration > 0 {
				// 注意：在设备维度，由于多个端口重启时间不同，
				// 简单的 BytesIn 相减如果出现负数，说明有端口重启了。
				diffIn := curr.BytesIn - prev.BytesIn
				diffOut := curr.BytesOut - prev.BytesOut
				diffErr := curr.Errors - prev.Errors

				if diffIn >= 0 {
					inKbps = float64(diffIn) / 1024.0 / duration
					totalIn += diffIn
				} else {
					// 如果差值为负，说明本时段内至少有一个大流量端口重启了
					// 此时保守做法是直接累加当前值（代表从0开始的新增）
					totalIn += curr.BytesIn
				}

				if diffOut >= 0 {
					outKbps = float64(diffOut) / 1024.0 / duration
					totalOut += diffOut
				} else {
					totalOut += curr.BytesOut
				}

				if diffErr >= 0 {
					totalErr += diffErr
				}
			}
		}

		rates = append(rates, types.GetDevicePortForwardStatsRates{
			Timestamp:  curr.CreatedAt.UnixMilli(),
			InKBps:     math.Round(inKbps*1000) / 1000,
			OutKBps:    math.Round(outKbps*1000) / 1000,
			ActiveConn: curr.ActiveConn,
		})
	}

	res.Rates = rates
	res.Delta = types.GetDevicePortForwardStatsDelta{
		BytesInIncr:  totalIn,
		BytesOutIncr: totalOut,
		ErrorsIncr:   totalErr,
	}

	return &res
}
