package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/deviceipreport"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"golang.org/x/sync/errgroup"
)

// DeviceService 设备服务接口
type DeviceService interface {
	GetDevices(ctx context.Context) ([]*types.GetDeviceResponse, error)
	GetDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error)
	CreateDevice(ctx context.Context, request *types.CreateDeviceRequest) (*types.GetDeviceRequest, error)
	UpdateDevice(ctx context.Context, request *types.UpdateDeviceRequest) (*types.GetDeviceRequest, error)
	DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error)
	GetDeviceIpHistory(ctx context.Context, id int64) ([]*types.GetDeviceIpHistoryResponse, error)
	HealthcheckDevice(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// DeviceServiceImpl 设备服务实现
type DeviceServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewDeviceService 创建设备服务实例
func NewDeviceService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) DeviceService {
	return &DeviceServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// buildResponse 将实体转换为响应
func (s *DeviceServiceImpl) buildResponse(ctx context.Context, item *ent.Device) (*types.GetDeviceResponse, error) {
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
		Order(ent.Desc(device.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询设备失败: %w", err)
	}
	for _, item := range allDevice {
		// 将实体转换为响应
		list, err := s.buildResponse(ctx, item)
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
	list, err := s.buildResponse(ctx, firstDevice)
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
	// 记录操作日志
	_ = s.logHelper.RecordDataCreate(ctx, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "创建设备")

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
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "更新设备")

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
func (s *DeviceServiceImpl) DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error) {
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

	// 提取原始数据用于返回
	response := types.GetDeviceResponse{
		ID:              firstDevice.ID,
		Code:            firstDevice.Code,
		Name:            firstDevice.Name,
		Description:     firstDevice.Description,
		IPV4:            firstDevice.IPV4,
		IPV6:            firstDevice.IPV6,
		LastReportedAt:  firstDevice.LastReportedAt,
		LastHeartbeatAt: firstDevice.LastHeartbeatAt,
	}

	// 删除设备
	err = s.db.Device.DeleteOne(firstDevice).Exec(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("删除设备: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除设备失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataDelete(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "删除设备")

	// 删除通知订阅 (保持原始逻辑)
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDevice, firstDevice.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return &response, nil
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
	firstDeviceReports, err := firstDevice.QueryIPReports().
		Order(ent.Desc(deviceipreport.FieldCreatedAt)).
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
