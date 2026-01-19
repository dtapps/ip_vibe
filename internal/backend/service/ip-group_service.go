package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/ipgroup"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"golang.org/x/sync/errgroup"
)

// IPGroupService IP组服务接口
type IPGroupService interface {
	BuildResponse(ctx context.Context, item *ent.IpGroup) (*types.GetIPGroupResponse, error)
	GetIPGroups(ctx context.Context) ([]*types.GetIPGroupResponse, error)
	GetIPGroup(ctx context.Context, id int64) (*types.GetIPGroupResponse, error)
	CreateIPGroup(ctx context.Context, request *types.CreateIPGroupRequest) (int64, error)
	UpdateIPGroup(ctx context.Context, request *types.UpdateIPGroupRequest) (*types.GetIPGroupRequest, error)
	DeleteIPGroup(ctx context.Context, id int64) (*types.GetIPGroupRequest, error)
	GetIPGroupDevices(ctx context.Context, id int64) ([]*types.GetDeviceResponse, error)
	HealthcheckIPGroup(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// IPGroupServiceImpl IP组服务实现
type IPGroupServiceImpl struct {
	db    *ent.Client // 数据库客户端
	debug bool        // 调试模式
}

// NewIPGroupService 创建 IP 组服务实例
func NewIPGroupService(db *ent.Client, debug bool) IPGroupService {
	return &IPGroupServiceImpl{
		db:    db,
		debug: debug,
	}
}

// BuildResponse 将实体转换为响应
func (s *IPGroupServiceImpl) BuildResponse(ctx context.Context, item *ent.IpGroup) (*types.GetIPGroupResponse, error) {
	resp := &types.GetIPGroupResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
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
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeIPGroup)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetIPGroups 获取全部IP组
func (s *IPGroupServiceImpl) GetIPGroups(ctx context.Context) ([]*types.GetIPGroupResponse, error) {
	lists := make([]*types.GetIPGroupResponse, 0)
	// 查询全部IP组
	allIpGroup, err := s.db.IpGroup.Query().
		OrderByCreatedAtDesc().
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}
	for _, item := range allIpGroup {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetIPGroup 获取单个IP组
func (s *IPGroupServiceImpl) GetIPGroup(ctx context.Context, id int64) (*types.GetIPGroupResponse, error) {
	// 查询IP组
	firstIpGroup, err := s.db.IpGroup.Query().
		Where(ipgroup.IDEQ(id)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询IP组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstIpGroup)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateIPGroup 创建IP组
// 注意：Handler返回的是 gin.H{"id": saveIpGroup.ID}，这里返回 int64 ID
func (s *IPGroupServiceImpl) CreateIPGroup(ctx context.Context, request *types.CreateIPGroupRequest) (int64, error) {
	// 创建IP组
	saveIpGroup, err := s.db.IpGroup.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		AddDeviceIDs(request.Devices...).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("创建IP组: %s\n", err.Error())
		}
		return 0, fmt.Errorf("创建IP组失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeIPGroup, saveIpGroup.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return saveIpGroup.ID, nil
}

// UpdateIPGroup 更新IP组
func (s *IPGroupServiceImpl) UpdateIPGroup(ctx context.Context, request *types.UpdateIPGroupRequest) (*types.GetIPGroupRequest, error) {
	// 查询IP组
	firstIpGroup, err := s.db.IpGroup.Query().
		Where(ipgroup.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询IP组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}
	// 更新IP组
	saveIpGroup, err := s.db.IpGroup.UpdateOne(firstIpGroup).
		SetName(request.Name).
		SetDescription(request.Description).
		ClearDevices().
		AddDeviceIDs(request.Devices...).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新IP组: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新IP组失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeIPGroup, saveIpGroup.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetIPGroupRequest{
		ID: saveIpGroup.ID,
	}, nil
}

// DeleteIPGroup 删除IP组
func (s *IPGroupServiceImpl) DeleteIPGroup(ctx context.Context, id int64) (*types.GetIPGroupRequest, error) {
	// 查询IP组
	firstIpGroup, err := s.db.IpGroup.Query().
		Where(ipgroup.IDEQ(id)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询IP组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}

	// 判断是否有关联的CDN记录或DNS记录
	if records, err := firstIpGroup.Edges.CdnRecordsOrErr(); err == nil && len(records) > 0 {
		return nil, fmt.Errorf("IP组 %d 有关联的CDN记录，无法删除", firstIpGroup.ID)
	}
	if records, err := firstIpGroup.Edges.DNSRecordsOrErr(); err == nil && len(records) > 0 {
		return nil, fmt.Errorf("IP组 %d 有关联的DNS记录，无法删除", firstIpGroup.ID)
	}

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建事务失败: %w", err)
	}

	// 先删除关联的设备
	err = tx.IpGroup.UpdateOne(firstIpGroup).ClearDevices().Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("删除关联设备失败: %w", err)
	}

	// 删除IP组
	err = tx.IpGroup.DeleteOne(firstIpGroup).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("删除IP组失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.DeleteTx(ctx, tx, webhooksubscription.ObjTypeIPGroup, firstIpGroup.ID)
	if err != nil {
		_ = tx.Rollback()
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &types.GetIPGroupRequest{
		ID: firstIpGroup.ID,
	}, nil
}

// GetIPGroupDevices 获取IP组的设备列表
func (s *IPGroupServiceImpl) GetIPGroupDevices(ctx context.Context, id int64) ([]*types.GetDeviceResponse, error) {
	// 查询IP组
	firstIpGroup, err := s.db.IpGroup.Query().
		Where(ipgroup.IDEQ(id)).
		WithDevices().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询IP组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}
	// 设备列表
	ipGroupDevices := make([]*types.GetDeviceResponse, 0)
	if devices, err := firstIpGroup.Edges.DevicesOrErr(); err == nil {
		for _, item := range devices {
			ipGroupDevices = append(ipGroupDevices, &types.GetDeviceResponse{
				ID:              item.ID,
				Code:            item.Code,
				Name:            item.Name,
				Description:     item.Description,
				IPV4:            item.IPV4,
				IPV6:            item.IPV6,
				LastReportedAt:  item.LastReportedAt,
				LastHeartbeatAt: item.LastHeartbeatAt,
			})
		}
	}
	return ipGroupDevices, nil
}

// HealthcheckIPGroup 检查IP组健康状态
func (s *IPGroupServiceImpl) HealthcheckIPGroup(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询IP组
	firstIpGroup, err := s.db.IpGroup.Query().
		Where(ipgroup.IDEQ(id)).
		WithDevices().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询IP组失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询IP组失败: %w", err)
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发 (保持原始限制)

	for _, item := range firstIpGroup.Edges.Devices {
		item := item // 创建副本以避免闭包问题
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
		return nil, fmt.Errorf("IP组健康状态检查失败: %w", err)
	}

	return results, nil
}
