package service

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/deviceipreport"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"golang.org/x/sync/errgroup"
)

// DeviceService 设备服务接口
type DeviceService interface {
	GetDevices(ctx context.Context) ([]types.GetDeviceResponse, error)
	GetDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error)
	CreateDevice(ctx context.Context, request *types.CreateDeviceRequest) (*types.GetDeviceRequest, error)
	UpdateDevice(ctx context.Context, request *types.UpdateDeviceRequest) (*types.GetDeviceRequest, error)
	DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error)
	GetDeviceIpHistory(ctx context.Context, id int64) ([]types.GetDeviceIpHistoryResponse, error)
	HealthcheckDevice(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// DeviceServiceImpl 设备服务实现
type DeviceServiceImpl struct{}

// NewDeviceService 创建设备服务实例
func NewDeviceService() DeviceService {
	return &DeviceServiceImpl{}
}

// GetDevices 获取全部设备
func (s *DeviceServiceImpl) GetDevices(ctx context.Context) ([]types.GetDeviceResponse, error) {
	devices := make([]types.GetDeviceResponse, 0)

	// 查询全部设备
	allDevice, err := global.GvaEntDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range allDevice {
		list := types.GetDeviceResponse{
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
		list.Webhooks, _ = global.GvaEntDb.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDevice)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			QueryWebhook().
			IDs(ctx)

		devices = append(devices, list)
	}

	return devices, nil
}

// GetDevice 获取设备
func (s *DeviceServiceImpl) GetDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error) {
	// 查询设备
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	response := &types.GetDeviceResponse{
		ID:              firstDevice.ID,
		Code:            firstDevice.Code,
		Name:            firstDevice.Name,
		Description:     firstDevice.Description,
		IPV4:            firstDevice.IPV4,
		IPV6:            firstDevice.IPV6,
		LastReportedAt:  firstDevice.LastReportedAt,
		LastHeartbeatAt: firstDevice.LastHeartbeatAt,
	}

	// 查询通知订阅
	response.Webhooks, _ = global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDevice)).
		Where(webhooksubscription.ObjIDEQ(id)).
		QueryWebhook().
		IDs(ctx)

	return response, nil
}

// CreateDevice 创建设备
func (s *DeviceServiceImpl) CreateDevice(ctx context.Context, request *types.CreateDeviceRequest) (*types.GetDeviceRequest, error) {
	// 创建设备
	saveDevice, err := global.GvaEntDb.Device.Create().
		SetCode(request.Code).
		SetName(request.Name).
		SetDescription(request.Description).
		SetIPV4(request.IPV4).
		SetIPV6(request.IPV6).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(ctx, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "创建设备")

	// 创建通知订阅
	_ = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDevice, saveDevice.ID)

	return &types.GetDeviceRequest{
		ID: saveDevice.ID,
	}, nil
}

// UpdateDevice 更新设备
func (s *DeviceServiceImpl) UpdateDevice(ctx context.Context, request *types.UpdateDeviceRequest) (*types.GetDeviceRequest, error) {
	// 查询设备
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 更新设备
	saveDevice, err := global.GvaEntDb.Device.UpdateOne(firstDevice).
		SetName(request.Name).
		SetDescription(request.Description).
		SetIPV4(request.IPV4).
		SetIPV6(request.IPV6).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "更新设备")

	// 更新通知订阅
	_ = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDevice, saveDevice.ID)

	return &types.GetDeviceRequest{
		ID: saveDevice.ID,
	}, nil
}

// DeleteDevice 删除设备
func (s *DeviceServiceImpl) DeleteDevice(ctx context.Context, id int64) (*types.GetDeviceResponse, error) {
	// 查询设备
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 删除设备
	err = global.GvaEntDb.Device.DeleteOne(firstDevice).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(ctx, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "删除设备")

	// 删除通知订阅
	_ = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDevice, firstDevice.ID)

	return &types.GetDeviceResponse{
		ID: firstDevice.ID,
	}, nil
}

// GetDeviceIpHistory 获取设备的IP历史记录
func (s *DeviceServiceImpl) GetDeviceIpHistory(ctx context.Context, id int64) ([]types.GetDeviceIpHistoryResponse, error) {
	// 查询设备
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 查询设备IP历史记录
	firstDeviceReports, err := firstDevice.QueryIPReports().
		Order(ent.Desc(deviceipreport.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 设备列表
	ips := make([]types.GetDeviceIpHistoryResponse, 0)
	for _, item := range firstDeviceReports {
		ips = append(ips, types.GetDeviceIpHistoryResponse{
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
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

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
		return nil, err
	}

	return results, nil
}
