package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/deviceportforward"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// PortForwardService 端口转发服务接口
type PortForwardService interface {
	BuildResponse(ctx context.Context, item *ent.DevicePortForward) (*types.GetPortForwardResponse, error)
	GetPortForwards(ctx context.Context, request *types.GetPortForwardsRequest) ([]*types.GetPortForwardResponse, error)
	GetPortForward(ctx context.Context, id int64) (*types.GetPortForwardResponse, error)
	CreatePortForward(ctx context.Context, request *types.CreatePortForwardRequest) (*types.GetPortForwardRequest, error)
	UpdatePortForward(ctx context.Context, request *types.UpdatePortForwardRequest) (*types.GetPortForwardRequest, error)
	UpdatePortForwardForward(ctx context.Context, request *types.UpdatePortForwardForwardRequest) (*types.GetPortForwardRequest, error)
	UpdatePortForwardActive(ctx context.Context, request *types.UpdatePortForwardActiveRequest) (*types.GetPortForwardRequest, error)
	DeletePortForward(ctx context.Context, id int64) (*types.GetPortForwardRequest, error)
	HealthcheckPortForward(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// PortForwardServiceImpl 端口转发服务实现
type PortForwardServiceImpl struct {
	db        *ent.Client // 数据库客户端
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewPortForwardService 创建端口转发服务实例
func NewPortForwardService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) PortForwardService {
	return &PortForwardServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// BuildResponse 将实体转换为响应
func (s *PortForwardServiceImpl) BuildResponse(ctx context.Context, item *ent.DevicePortForward) (*types.GetPortForwardResponse, error) {
	resp := &types.GetPortForwardResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		LocalPort:   item.LocalPort,
		RemoteHost:  item.RemoteHost,
		RemotePort:  item.RemotePort,
		Protocols:   item.Protocols,
		AutoForward: item.AutoForward,
		IsActive:    item.IsActive,
	}

	// 设备
	if item.Edges.Device != nil {
		resp.DeviceID = item.Edges.Device.ID
		resp.DeviceName = item.Edges.Device.Name
	}

	return resp, nil
}

// GetPortForwards 获取全部端口转发
func (s *PortForwardServiceImpl) GetPortForwards(ctx context.Context, request *types.GetPortForwardsRequest) ([]*types.GetPortForwardResponse, error) {
	lists := make([]*types.GetPortForwardResponse, 0)
	// 查询全部端口转发
	allPortForward, err := s.db.DevicePortForward.Query().
		WhereIf(request.IsActive != nil, deviceportforward.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithDevice().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}
	for _, item := range allPortForward {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetPortForward 获取单个端口转发
func (s *PortForwardServiceImpl) GetPortForward(ctx context.Context, id int64) (*types.GetPortForwardResponse, error) {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(id)).
		WithDevice().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstPortForward)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreatePortForward 创建端口转发
func (s *PortForwardServiceImpl) CreatePortForward(ctx context.Context, request *types.CreatePortForwardRequest) (*types.GetPortForwardRequest, error) {
	// 创建端口转发
	savePortForward, err := s.db.DevicePortForward.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetLocalPort(request.LocalPort).
		SetRemoteHost(request.RemoteHost).
		SetRemotePort(request.RemotePort).
		SetProtocols(request.Protocols).
		SetDeviceID(request.DeviceID).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("创建端口转发: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建端口转发失败: %w", err)
	}
	// 记录操作日志 (使用全局 GvaOperationLogHelper)
	_ = s.logHelper.RecordDataCreate(ctx, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "创建端口转发")
	return &types.GetPortForwardRequest{
		ID: savePortForward.ID,
	}, nil
}

// UpdatePortForward 更新端口转发
func (s *PortForwardServiceImpl) UpdatePortForward(ctx context.Context, request *types.UpdatePortForwardRequest) (*types.GetPortForwardRequest, error) {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}
	// 更新端口转发
	savePortForward, err := s.db.DevicePortForward.UpdateOne(firstPortForward).
		SetName(request.Name).
		SetDescription(request.Description).
		SetLocalPort(request.LocalPort).
		SetRemoteHost(request.RemoteHost).
		SetRemotePort(request.RemotePort).
		SetProtocols(request.Protocols).
		SetDeviceID(request.DeviceID).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新端口转发: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新端口转发失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发")
	return &types.GetPortForwardRequest{
		ID: savePortForward.ID,
	}, nil
}

// UpdatePortForwardForward 更新端口转发防火墙状态 (AutoForward)
func (s *PortForwardServiceImpl) UpdatePortForwardForward(ctx context.Context, request *types.UpdatePortForwardForwardRequest) (*types.GetPortForwardRequest, error) {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}
	// 更新端口转发状态
	savePortForward, err := s.db.DevicePortForward.UpdateOne(firstPortForward).
		SetAutoForward(request.AutoForward).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新端口转发防火墙状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新端口转发防火墙状态失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发防火墙状态")
	return &types.GetPortForwardRequest{
		ID: savePortForward.ID,
	}, nil
}

// UpdatePortForwardActive 更新端口转发启用状态 (IsActive)
func (s *PortForwardServiceImpl) UpdatePortForwardActive(ctx context.Context, request *types.UpdatePortForwardActiveRequest) (*types.GetPortForwardRequest, error) {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}
	// 更新端口转发状态
	savePortForward, err := s.db.DevicePortForward.UpdateOne(firstPortForward).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新端口转发启用状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新端口转发启用状态失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发启用状态")
	return &types.GetPortForwardRequest{
		ID: savePortForward.ID,
	}, nil
}

// DeletePortForward 删除端口转发
func (s *PortForwardServiceImpl) DeletePortForward(ctx context.Context, id int64) (*types.GetPortForwardRequest, error) {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}
	// 删除端口转发
	err = s.db.DevicePortForward.DeleteOne(firstPortForward).Exec(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("删除端口转发: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除端口转发失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataDelete(ctx, firstPortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "删除端口转发")
	return &types.GetPortForwardRequest{
		ID: firstPortForward.ID,
	}, nil
}

// HealthcheckPortForward 检查端口转发健康状态
func (s *PortForwardServiceImpl) HealthcheckPortForward(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询端口转发并加载关联的设备
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(id)).
		WithDevice().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询端口转发失败: %w", err)
	}

	// 检查关联的设备是否存在 IP 地址
	if firstPortForward.Edges.Device == nil {
		return nil, fmt.Errorf("端口转发未关联设备")
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查 (保持原始 errgroup 逻辑和限制)
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

	localPort := firstPortForward.LocalPort
	deviceIPV4 := firstPortForward.Edges.Device.IPV4
	deviceIPV6 := firstPortForward.Edges.Device.IPV6

	if deviceIPV4 != "" {
		// 检测 IPv4 是否可通过本地端口访问
		eg.Go(func() error {
			// 注意：Handler 中 CheckIPHealth 的 port 参数传递了两次 localPort
			healthCheckResult, _ := CheckIPHealth(egCtx, deviceIPV4, localPort, localPort)
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}
	if deviceIPV6 != "" {
		// 检测 IPv6 是否可通过本地端口访问
		eg.Go(func() error {
			healthCheckResult, _ := CheckIPHealth(egCtx, deviceIPV6, localPort, localPort)
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		// 错误处理也保持原始逻辑，返回内部错误
		return nil, fmt.Errorf("端口转发健康状态检查失败: %w", err)
	}

	return results, nil
}
