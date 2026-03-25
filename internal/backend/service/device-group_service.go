package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/devicegroup"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/portforwardstats"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gomd5"
	"golang.org/x/sync/errgroup"
)

// DeviceGroupService 设备组服务接口
type DeviceGroupService interface {
	BuildResponse(ctx context.Context, item *ent.DeviceGroup) (*types.GetDeviceGroupResponse, error)
	GetDeviceGroups(ctx context.Context) ([]*types.GetDeviceGroupResponse, error)
	GetDeviceGroup(ctx context.Context, id int64) (*types.GetDeviceGroupResponse, error)
	CreateDeviceGroup(ctx context.Context, request *types.CreateDeviceGroupRequest) error
	UpdateDeviceGroup(ctx context.Context, request *types.UpdateDeviceGroupRequest) error
	DeleteDeviceGroup(ctx context.Context, id int64) error
	BatchDeleteDeviceGroup(ctx context.Context, ids []int64) error
	GetDeviceGroupDevices(ctx context.Context, id int64) ([]*types.GetDeviceResponse, error)
	HealthcheckDeviceGroup(ctx context.Context, id int64) ([][]ToPairsStruct, error)
	GetDeviceGroupDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDeviceGroupDevicePortForwardStatsRequest) (*types.GetDeviceGroupDevicePortForwardStatsResponse, error)
}

// DeviceGroupServiceImpl 设备组服务实现
type DeviceGroupServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式

	deviceService DeviceService // 设备服务
}

// NewDeviceGroupService 创建 IP 组服务实例
func NewDeviceGroupService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) DeviceGroupService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &DeviceGroupServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,

		deviceService: NewDeviceService(db, logDb, logger),
	}
}

// SetDebug 设置调试模式
func (s *DeviceGroupServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *DeviceGroupServiceImpl) BuildResponse(ctx context.Context, item *ent.DeviceGroup) (*types.GetDeviceGroupResponse, error) {
	resp := &types.GetDeviceGroupResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	if item.Code == "" {
		resp.Code = gomd5.Get16MD5Encode(cast.ToString(item.ID))
	}

	// 设备列表
	if devices, err := item.Edges.DevicesOrErr(); err == nil {
		for _, device := range devices {
			resp.Devices = append(resp.Devices, device.ID)
		}
	}

	// CDN记录列表
	if records, err := item.Edges.CdnRecordsOrErr(); err == nil {
		for _, record := range records {
			resp.CdnRecords = append(resp.CdnRecords, record.ID)
		}
	}

	// DNS记录列表
	if records, err := item.Edges.DNSRecordsOrErr(); err == nil {
		for _, record := range records {
			resp.DnsRecords = append(resp.DnsRecords, record.ID)
		}
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDeviceGroup)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetDeviceGroups 获取全部设备组
func (s *DeviceGroupServiceImpl) GetDeviceGroups(ctx context.Context) ([]*types.GetDeviceGroupResponse, error) {

	// 查询全部设备组
	allDeviceGroup, err := s.db.DeviceGroup.Query().
		OrderByCreatedAtDesc().
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询设备组失败: %w", err)
	}

	lists := make([]*types.GetDeviceGroupResponse, 0, len(allDeviceGroup))
	for _, item := range allDeviceGroup {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetDeviceGroup 获取单个设备组
func (s *DeviceGroupServiceImpl) GetDeviceGroup(ctx context.Context, id int64) (*types.GetDeviceGroupResponse, error) {
	// 查询设备组
	firstDeviceGroup, err := s.db.DeviceGroup.Query().
		Where(devicegroup.IDEQ(id)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备组失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstDeviceGroup)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateDeviceGroup 创建设备组
func (s *DeviceGroupServiceImpl) CreateDeviceGroup(ctx context.Context, request *types.CreateDeviceGroupRequest) error {
	// 创建设备组
	saveDeviceGroup, err := s.db.DeviceGroup.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		AddDeviceIDs(request.Devices...).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("创建设备组失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDeviceGroup, saveDeviceGroup.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateDeviceGroup 更新设备组
func (s *DeviceGroupServiceImpl) UpdateDeviceGroup(ctx context.Context, request *types.UpdateDeviceGroupRequest) error {
	// 查询设备组
	firstDeviceGroup, err := s.db.DeviceGroup.Query().
		Where(devicegroup.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询设备组失败: %w", err)
	}
	// 更新设备组
	saveDeviceGroup, err := s.db.DeviceGroup.UpdateOne(firstDeviceGroup).
		SetName(request.Name).
		SetDescription(request.Description).
		ClearDevices().
		AddDeviceIDs(request.Devices...).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新设备组失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDeviceGroup, saveDeviceGroup.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// DeleteDeviceGroup 删除设备组
func (s *DeviceGroupServiceImpl) DeleteDeviceGroup(ctx context.Context, id int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询设备组
	firstDeviceGroup, err := tx.DeviceGroup.Query().
		Where(devicegroup.IDEQ(id)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		First(ctx)
	if err != nil {
		_ = tx.Rollback()
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询设备组失败: %w", err)
	}

	// 判断是否有关联的CDN记录或DNS记录
	if records, err := firstDeviceGroup.Edges.CdnRecordsOrErr(); err == nil && len(records) > 0 {
		_ = tx.Rollback()
		return fmt.Errorf("设备组 %d 有关联的CDN记录，无法删除", firstDeviceGroup.ID)
	}
	if records, err := firstDeviceGroup.Edges.DNSRecordsOrErr(); err == nil && len(records) > 0 {
		_ = tx.Rollback()
		return fmt.Errorf("设备组 %d 有关联的DNS记录，无法删除", firstDeviceGroup.ID)
	}

	// 先删除关联的设备
	if devices, err := firstDeviceGroup.Edges.DevicesOrErr(); err == nil && len(devices) > 0 {
		err = tx.DeviceGroup.UpdateOne(firstDeviceGroup).
			ClearDevices().
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("删除关联设备失败: %w", err)
		}
	}

	// 删除设备组
	err = tx.DeviceGroup.DeleteOne(firstDeviceGroup).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除设备组失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).DeleteTx(ctx, tx, webhooksubscription.ObjTypeDeviceGroup, firstDeviceGroup.ID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除通知订阅失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// BatchDeleteDeviceGroup 批量删除设备组
func (s *DeviceGroupServiceImpl) BatchDeleteDeviceGroup(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询设备组
	allDeviceGroup, err := tx.DeviceGroup.Query().
		Where(devicegroup.IDIn(ids...)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询设备组失败: %w", err)
	}
	if len(allDeviceGroup) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询设备组失败: 数量 %d 与请求数量 %d 不匹配", len(allDeviceGroup), len(ids))
	}

	for _, item := range allDeviceGroup {
		// 判断是否有关联的CDN记录或DNS记录
		if records, err := item.Edges.CdnRecordsOrErr(); err == nil && len(records) > 0 {
			_ = tx.Rollback()
			return fmt.Errorf("设备组 %d 有关联的CDN记录，无法删除", item.ID)
		}
		if records, err := item.Edges.DNSRecordsOrErr(); err == nil && len(records) > 0 {
			_ = tx.Rollback()
			return fmt.Errorf("设备组 %d 有关联的DNS记录，无法删除", item.ID)
		}

		// 先删除关联的设备
		if devices, err := item.Edges.DevicesOrErr(); err == nil && len(devices) > 0 {
			err = tx.DeviceGroup.UpdateOne(item).ClearDevices().Exec(ctx)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("删除关联设备失败: %w", err)
			}
		}
	}

	// 删除设备组
	_, err = tx.DeviceGroup.Delete().
		Where(devicegroup.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除设备组失败: %w", err)
	}

	// 删除通知订阅
	for _, item := range allDeviceGroup {
		err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).DeleteTx(ctx, tx, webhooksubscription.ObjTypeDeviceGroup, item.ID)
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

// GetDeviceGroupDevices 获取设备组的设备列表
func (s *DeviceGroupServiceImpl) GetDeviceGroupDevices(ctx context.Context, id int64) ([]*types.GetDeviceResponse, error) {
	// 查询设备组
	firstDeviceGroup, err := s.db.DeviceGroup.Query().
		Where(devicegroup.IDEQ(id)).
		WithDevices().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备组失败: %w", err)
	}
	// 设备列表
	deviceGroupDevices := make([]*types.GetDeviceResponse, 0, len(firstDeviceGroup.Edges.Devices))
	if devices, err := firstDeviceGroup.Edges.DevicesOrErr(); err == nil {
		for _, item := range devices {
			itemResp, err := s.deviceService.BuildResponse(ctx, item)
			if err != nil {
				return nil, err
			}
			deviceGroupDevices = append(deviceGroupDevices, itemResp)
		}
	}
	return deviceGroupDevices, nil
}

// HealthcheckDeviceGroup 检查设备组健康状态
func (s *DeviceGroupServiceImpl) HealthcheckDeviceGroup(ctx context.Context, id int64) ([][]ToPairsStruct, error) {

	// 查询设备组
	firstDeviceGroup, err := s.db.DeviceGroup.Query().
		Where(devicegroup.IDEQ(id)).
		WithDevices().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备组失败: %w", err)
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

	for _, item := range firstDeviceGroup.Edges.Devices {
		if item.IPV4 != "" {
			// 检测 IP 是否可通过 HTTP/HTTPS 访问
			eg.Go(func() error {
				// 注意：CheckIPHealth 依赖于 backend 包，假设其已注入或作为全局/静态函数可用
				healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV4, 80, 443)
				results = append(results, healthCheckResult.ToPairsForIP())
				return nil
			})
		}
		if item.IPV6 != "" {
			// 检测 IP 是否可通过 HTTP/HTTPS 访问
			eg.Go(func() error {
				healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV6, 80, 443)
				results = append(results, healthCheckResult.ToPairsForIP())
				return nil
			})
		}
	}

	if err = eg.Wait(); err != nil {
		return nil, fmt.Errorf("设备组健康状态检查失败: %w", err)
	}

	return results, nil
}

// GetDeviceGroupDevicePortForwardStats 获取设备组设备端口转发统计数据
func (s *DeviceGroupServiceImpl) GetDeviceGroupDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDeviceGroupDevicePortForwardStatsRequest) (*types.GetDeviceGroupDevicePortForwardStatsResponse, error) {
	// 查询设备组
	firstDeviceGroup, err := s.db.DeviceGroup.Query().
		Where(devicegroup.IDEQ(id)).
		WithDevices().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询设备组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询设备组失败: %w", err)
	}

	// 设备列表
	deviceIDs := make([]int64, 0, len(firstDeviceGroup.Edges.Devices))
	if devices, err := firstDeviceGroup.Edges.DevicesOrErr(); err == nil {
		for _, item := range devices {
			deviceIDs = append(deviceIDs, item.ID)
		}
	}

	// 计算时间范围
	startTime := s.getStartTimeFromRange(request.Range)

	s.logger.InfoContext(ctx, "获取设备组设备端口转发统计数据",
		slog.Int64("ID", id),
		slog.String("Range", request.Range),
		slog.String("StartTime", startTime.Format(time.DateTime)),
		slog.Any("DeviceIDs", deviceIDs),
	)

	// 查询端口转发统计记录
	allStats, err := s.logDb.PortForwardStats.Query().
		Where(portforwardstats.DeviceIDIn(deviceIDs...)).
		Where(portforwardstats.CreatedAtGTE(startTime)).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询统计记录失败: %w", err)
	}

	// 计算统计数据
	result := s.calculatePortForwardStats(allStats)

	result.DeviceIDs = deviceIDs

	return result, nil
}

// 根据时间范围计算起始时间
func (s *DeviceGroupServiceImpl) getStartTimeFromRange(rangeStr string) time.Time {
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
func (s *DeviceGroupServiceImpl) calculatePortForwardStats(raw []*ent_log.PortForwardStats) *types.GetDeviceGroupDevicePortForwardStatsResponse {
	count := len(raw)
	if count == 0 {
		return &types.GetDeviceGroupDevicePortForwardStatsResponse{}
	}

	// 1. 聚合：将同一秒上报的所有记录（不论哪个设备、哪个端口）累加
	type aggPoint struct {
		ActiveConn int64
		BytesIn    int64
		BytesOut   int64
		Errors     int64
		Time       time.Time
	}

	timeMap := make(map[int64]*aggPoint)
	for _, item := range raw {
		ts := item.CreatedAt.Unix()
		if _, ok := timeMap[ts]; !ok {
			timeMap[ts] = &aggPoint{Time: item.CreatedAt}
		}
		timeMap[ts].ActiveConn += item.ActiveConn
		timeMap[ts].BytesIn += item.BytesIn
		timeMap[ts].BytesOut += item.BytesOut
		timeMap[ts].Errors += item.Errors
	}

	// 2. 排序：从 Map 转为切片并按时间升序排列
	sorted := make([]*aggPoint, 0, len(timeMap))
	for _, v := range timeMap {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Time.Before(sorted[j].Time)
	})

	// 3. 计算：遍历排序后的聚合点
	var res types.GetDeviceGroupDevicePortForwardStatsResponse
	aggCount := len(sorted)

	// 最新快照取最后一个点
	last := sorted[aggCount-1]
	res.Latest = types.GetDeviceGroupDevicePortForwardStatsLatest{
		ActiveConn: last.ActiveConn,
		BytesIn:    last.BytesIn,
		BytesOut:   last.BytesOut,
		Errors:     last.Errors,
	}

	var totalIn, totalOut, totalErr int64
	rates := make([]types.GetDeviceGroupDevicePortForwardStatsRates, 0, aggCount)

	for i := range aggCount {
		curr := sorted[i]
		var inKbps, outKbps float64

		if i > 0 {
			prev := sorted[i-1]
			duration := curr.Time.Sub(prev.Time).Seconds()

			if duration > 0 {
				// 计算差值
				diffIn := curr.BytesIn - prev.BytesIn
				diffOut := curr.BytesOut - prev.BytesOut
				diffErr := curr.Errors - prev.Errors

				// 处理逻辑：如果差值大于等于 0，正常计算速率并累加增量
				// 如果差值为负（说明组内有设备重启了，计数器归零），则将当前值直接计为增量
				if diffIn >= 0 {
					inKbps = float64(diffIn) / 1024.0 / duration
					totalIn += diffIn
				} else {
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

		rates = append(rates, types.GetDeviceGroupDevicePortForwardStatsRates{
			Timestamp:  curr.Time.UnixMilli(),
			InKBps:     math.Round(inKbps*1000) / 1000,
			OutKBps:    math.Round(outKbps*1000) / 1000,
			ActiveConn: curr.ActiveConn,
		})
	}

	res.Rates = rates
	res.Delta = types.GetDeviceGroupDevicePortForwardStatsDelta{
		BytesInIncr:  totalIn,
		BytesOutIncr: totalOut,
		ErrorsIncr:   totalErr,
	}

	return &res
}
