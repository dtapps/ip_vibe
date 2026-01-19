package service

import (
	"context"
	"fmt"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/deviceportforward"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/portforwardstats"
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
	UpdatePortForwardActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdatePortForwardActive(ctx context.Context, ids []int64, isActive bool) error
	DeletePortForward(ctx context.Context, id int64) error
	BatchDeletePortForward(ctx context.Context, ids []int64) error
	HealthcheckPortForward(ctx context.Context, id int64) ([][]ToPairsStruct, error)
	GetPortForwardStats(ctx context.Context, id int64, request *types.GetPortForwardStatsRequest) ([]*types.GetPortForwardStatsResponse, error)
}

// PortForwardServiceImpl 端口转发服务实现
type PortForwardServiceImpl struct {
	db    *ent.Client     // 数据库客户端
	logDb *ent_log.Client // 日志数据库客户端
	debug bool            // 调试模式
}

// NewPortForwardService 创建端口转发服务实例
func NewPortForwardService(db *ent.Client, logDb *ent_log.Client, debug bool) PortForwardService {
	return &PortForwardServiceImpl{
		db:    db,
		logDb: logDb,
		debug: debug,
	}
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
	if device, err := item.Edges.DeviceOrErr(); err == nil {
		resp.DeviceID = device.ID
		resp.DeviceName = device.Name
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
		return nil, fmt.Errorf("创建端口转发失败: %w", err)
	}
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
		return nil, fmt.Errorf("更新端口转发失败: %w", err)
	}
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
		return nil, fmt.Errorf("更新端口转发防火墙状态失败: %w", err)
	}
	return &types.GetPortForwardRequest{
		ID: savePortForward.ID,
	}, nil
}

// UpdatePortForwardActive 更新端口转发启用状态
func (s *PortForwardServiceImpl) UpdatePortForwardActive(ctx context.Context, id int64, isActive bool) error {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询端口转发失败: %w", err)
	}

	// 更新端口转发状态
	err = s.db.DevicePortForward.UpdateOne(firstPortForward).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新端口转发启用状态失败: %w", err)
	}

	return nil
}

// BatchUpdatePortForwardActive 更新端口转发启用状态
func (s *PortForwardServiceImpl) BatchUpdatePortForwardActive(ctx context.Context, ids []int64, isActive bool) error {
	// 查询端口转发
	allPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询端口转发失败: %w", err)
	}
	if len(allPortForward) != len(ids) {
		return fmt.Errorf("查询端口转发失败: 数量 %d 与请求数量 %d 不匹配", len(allPortForward), len(ids))
	}

	// 更新端口转发状态
	err = s.db.DevicePortForward.Update().
		Where(deviceportforward.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新端口转发启用状态失败: %w", err)
	}

	return nil
}

// DeletePortForward 删除端口转发
func (s *PortForwardServiceImpl) DeletePortForward(ctx context.Context, id int64) error {
	// 查询端口转发
	firstPortForward, err := s.db.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询端口转发失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询端口转发失败: %w", err)
	}
	// 删除端口转发
	err = s.db.DevicePortForward.DeleteOne(firstPortForward).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除端口转发失败: %w", err)
	}

	return nil
}

// BatchDeletePortForward 删除端口转发
func (s *PortForwardServiceImpl) BatchDeletePortForward(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询端口转发
	allPortForward, err := tx.DevicePortForward.Query().
		Where(deviceportforward.IDIn(ids...)).
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询端口转发失败: %w", err)
	}
	if len(allPortForward) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询端口转发失败: 数量 %d 与请求数量 %d 不匹配", len(allPortForward), len(ids))
	}

	// 检查是否存在启用的端口转发
	for _, item := range allPortForward {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除端口转发失败: 存在启用的端口转发")
		}
	}

	// 删除端口转发
	_, err = tx.DevicePortForward.Delete().
		Where(deviceportforward.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除端口转发失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
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

// GetPortForwardStats 获取端口转发统计数据
func (s *PortForwardServiceImpl) GetPortForwardStats(ctx context.Context, id int64, request *types.GetPortForwardStatsRequest) ([]*types.GetPortForwardStatsResponse, error) {
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
		Where(portforwardstats.DeviceIDEQ(firstPortForward.Edges.Device.ID)).
		Where(portforwardstats.PortForwardIDEQ(firstPortForward.ID)).
		Where(portforwardstats.CreatedAtGTE(startTime)).
		OrderByCreatedAtDesc().
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("查询统计记录失败: %w", err)
	}

	lists := make([]*types.GetPortForwardStatsResponse, 0)
	for _, stat := range allStats {
		lists = append(lists, &types.GetPortForwardStatsResponse{
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
