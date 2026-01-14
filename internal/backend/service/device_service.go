package service

import (
	"context"
	"fmt"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/ent_log/portforwardstats"
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
	DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceRequest, error)
	GetDeviceIpHistory(ctx context.Context, id int64) ([]*types.GetDeviceIpHistoryResponse, error)
	HealthcheckDevice(ctx context.Context, id int64) ([][]ToPairsStruct, error)
	GetDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDevicePortForwardStatsRequest) ([]*types.GetDevicePortForwardStatsResponse, error)
}

// DeviceServiceImpl 设备服务实现
type DeviceServiceImpl struct {
	db    *ent.Client     // 数据库客户端
	logDb *ent_log.Client // 日志数据库客户端
	debug bool            // 调试模式
}

// NewDeviceService 创建设备服务实例
func NewDeviceService(db *ent.Client, logDb *ent_log.Client, debug bool) DeviceService {
	return &DeviceServiceImpl{
		db:    db,
		logDb: logDb,
		debug: debug,
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

	return resp, nil
}

// GetDevices 获取全部设备
func (s *DeviceServiceImpl) GetDevices(ctx context.Context) ([]*types.GetDeviceResponse, error) {
	lists := make([]*types.GetDeviceResponse, 0)
	// 查询全部设备
	allDevice, err := s.db.Device.Query().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
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
		if s.debug {
			fmt.Printf("创建设备: %s\n", err.Error())
		}
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
		if s.debug {
			fmt.Printf("更新设备: %s\n", err.Error())
		}
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
func (s *DeviceServiceImpl) DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceRequest, error) {
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

	// 删除设备
	err = s.db.Device.DeleteOne(firstDevice).Exec(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("删除设备: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除设备失败: %w", err)
	}

	// 删除通知订阅 (保持原始逻辑)
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDevice, firstDevice.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetDeviceRequest{
		ID: firstDevice.ID,
	}, nil
}

// GetDeviceIpHistory 获取设备的IP历史记录
func (s *DeviceServiceImpl) GetDeviceIpHistory(ctx context.Context, id int64) ([]*types.GetDeviceIpHistoryResponse, error) {
	// 查询设备 (保持原始逻辑)
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
	ips := make([]*types.GetDeviceIpHistoryResponse, 0)
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
func (s *DeviceServiceImpl) GetDevicePortForwardStats(ctx context.Context, id int64, request *types.GetDevicePortForwardStatsRequest) ([]*types.GetDevicePortForwardStatsResponse, error) {
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

	// 根据时间范围计算起始时间
	startTime := time.Now().Add(-6 * time.Hour) // 默认 6 小时
	if request.Range == "1h" {
		startTime = time.Now().Add(-1 * time.Hour)
	} else if request.Range == "6h" {
		startTime = time.Now().Add(-6 * time.Hour)
	} else if request.Range == "24h" {
		startTime = time.Now().Add(-24 * time.Hour)
	} else if request.Range == "7d" {
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if s.debug {
		fmt.Printf("查询端口转发统计数据: ID=%d, Range=%s, StartTime=%s\n", id, request.Range, startTime.Format(time.DateTime))
	}

	// 查询端口转发统计记录
	allStats, err := s.logDb.PortForwardStats.Query().
		Where(portforwardstats.DeviceIDEQ(firstDevice.ID)).
		Where(portforwardstats.CreatedAtGTE(startTime)).
		OrderByCreatedAtDesc().
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("查询统计记录失败: %w", err)
	}

	lists := make([]*types.GetDevicePortForwardStatsResponse, 0)
	for _, stat := range allStats {
		lists = append(lists, &types.GetDevicePortForwardStatsResponse{
			ID:            stat.ID,
			CreatedAt:     stat.CreatedAt,
			DeviceID:      stat.DeviceID,
			PortForwardID: stat.PortForwardID,
			StartTime:     stat.StartTime,
			ActiveConn:    stat.ActiveConn,
			TotalConn:     stat.TotalConn,
			BytesIn:       stat.BytesIn,
			BytesOut:      stat.BytesOut,
			UDPPacketsIn:  stat.UDPPacketsIn,
			UDPPacketsOut: stat.UDPPacketsOut,
			Errors:        stat.Errors,
		})
	}
	return lists, nil
}
