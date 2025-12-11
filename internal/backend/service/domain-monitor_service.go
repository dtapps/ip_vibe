package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global" // 仅用于访问 GvaOperationLogHelper 和类型定义
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
)

// DomainMonitorService 域名监控服务接口
type DomainMonitorService interface {
	GetDomainMonitors(ctx context.Context) ([]types.GetDomainMonitorResponse, error)
	GetDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorResponse, error)
	CreateDomainMonitor(ctx context.Context, request *types.CreateDomainMonitorRequest) (*types.GetDomainMonitorRequest, error)
	UpdateDomainMonitor(ctx context.Context, request *types.UpdateDomainMonitorRequest) (*types.GetDomainMonitorRequest, error)
	UpdateDomainMonitorActive(ctx context.Context, request *types.UpdateDomainMonitorActiveRequest) (*types.GetDomainMonitorRequest, error)
	DeleteDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorRequest, error)
	CheckDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorRequest, error)
}

// DomainMonitorServiceImpl 域名监控服务实现
type DomainMonitorServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
	cron      gocron.Scheduler
}

// NewDomainMonitorService 创建域名监控服务实例
func NewDomainMonitorService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper, cron gocron.Scheduler) DomainMonitorService {
	return &DomainMonitorServiceImpl{db: db, cron: cron, debug: debug, logHelper: logHelper}
}

// GetDomainMonitors 获取全部域名监控
func (s *DomainMonitorServiceImpl) GetDomainMonitors(ctx context.Context) ([]types.GetDomainMonitorResponse, error) {
	domainMonitors := make([]types.GetDomainMonitorResponse, 0)
	// 查询全部域名监控
	allDomainMonitor, err := s.db.DomainMonitor.Query().
		Order(ent.Desc(domainmonitor.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}
	for _, item := range allDomainMonitor {
		list := types.GetDomainMonitorResponse{
			ID:            item.ID,
			Name:          item.Name,
			Domain:        item.Domain,
			CheckInterval: item.CheckInterval,
			CheckProtocol: item.CheckProtocol.String(),
			IsActive:      item.IsActive,
			CurrentStatus: item.CurrentStatus.String(),
			LastSuccessAt: item.LastSuccessAt,
			LastFailureAt: item.LastFailureAt,
		}
		// 查询通知订阅 (保持原始查询逻辑)
		allWebhookSub, _ := s.db.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			All(ctx)
		for _, sub := range allWebhookSub {
			list.Webhooks = append(list.Webhooks, sub.ID)
		}
		domainMonitors = append(domainMonitors, list)
	}
	return domainMonitors, nil
}

// GetDomainMonitor 获取单个域名监控
func (s *DomainMonitorServiceImpl) GetDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorResponse, error) {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}
	response := types.GetDomainMonitorResponse{
		ID:            firstDomainMonitor.ID,
		Name:          firstDomainMonitor.Name,
		Domain:        firstDomainMonitor.Domain,
		CheckInterval: firstDomainMonitor.CheckInterval,
		CheckProtocol: firstDomainMonitor.CheckProtocol.String(),
		IsActive:      firstDomainMonitor.IsActive,
		CurrentStatus: firstDomainMonitor.CurrentStatus.String(),
		LastSuccessAt: firstDomainMonitor.LastSuccessAt,
		LastFailureAt: firstDomainMonitor.LastFailureAt,
	}
	// 查询通知订阅 (保持原始查询逻辑)
	allWebhookSub, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
		Where(webhooksubscription.ObjIDEQ(id)).
		All(ctx)
	for _, item := range allWebhookSub {
		response.Webhooks = append(response.Webhooks, item.ID)
	}
	return &response, nil
}

// CreateDomainMonitor 创建域名监控
func (s *DomainMonitorServiceImpl) CreateDomainMonitor(ctx context.Context, request *types.CreateDomainMonitorRequest) (*types.GetDomainMonitorRequest, error) {
	// 创建域名监控
	saveDomainMonitor, err := s.db.DomainMonitor.Create().
		SetName(request.Name).
		SetDomain(request.Domain).
		SetCheckInterval(request.CheckInterval).
		SetCheckProtocol(domainmonitor.CheckProtocol(request.CheckProtocol)).
		Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("创建域名监控: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建域名监控失败: %w", err)
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(ctx, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "创建域名监控")

	// 创建通知订阅
	_ = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDomainMonitor, saveDomainMonitor.ID)

	return &types.GetDomainMonitorRequest{
		ID: saveDomainMonitor.ID,
	}, nil
}

// UpdateDomainMonitor 更新域名监控
func (s *DomainMonitorServiceImpl) UpdateDomainMonitor(ctx context.Context, request *types.UpdateDomainMonitorRequest) (*types.GetDomainMonitorRequest, error) {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}
	// 更新域名监控
	saveDomainMonitor, err := s.db.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetName(request.Name).
		SetCheckInterval(request.CheckInterval).
		Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("更新域名监控: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新域名监控失败: %w", err)
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "更新域名监控")

	// 更新通知订阅
	_ = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDomainMonitor, saveDomainMonitor.ID)

	return &types.GetDomainMonitorRequest{
		ID: saveDomainMonitor.ID,
	}, nil
}

// UpdateDomainMonitorActive 更新域名监控启用状态
func (s *DomainMonitorServiceImpl) UpdateDomainMonitorActive(ctx context.Context, request *types.UpdateDomainMonitorActiveRequest) (*types.GetDomainMonitorRequest, error) {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}
	// 更新域名监控状态
	saveDomainMonitor, err := s.db.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("更新域名监控启用状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新域名监控启用状态失败: %w", err)
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "更新域名监控启用状态")

	// 删除通知订阅 (注意：Handler 中的这部分逻辑是删除通知订阅，与名称不符，但保持 Handler 原始逻辑)
	_ = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDomainMonitor, saveDomainMonitor.ID)

	return &types.GetDomainMonitorRequest{
		ID: saveDomainMonitor.ID,
	}, nil
}

// DeleteDomainMonitor 删除域名监控
func (s *DomainMonitorServiceImpl) DeleteDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorRequest, error) {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}
	// 删除域名监控
	err = s.db.DomainMonitor.DeleteOne(firstDomainMonitor).Exec(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("删除域名监控: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除域名监控失败: %w", err)
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(ctx, firstDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "删除域名监控")

	return &types.GetDomainMonitorRequest{
		ID: firstDomainMonitor.ID,
	}, nil
}

// CheckDomainMonitor 检查域名监控（执行一次任务）
func (s *DomainMonitorServiceImpl) CheckDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorRequest, error) {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}

	// 域名监控任务 (保持 Handler 中创建依赖的原始逻辑，但使用 Service 内部的注入依赖)
	domainMonitorService := task.NewDomainMonitorTask(s.db, s.cron)

	// 执行单个监控任务
	// 注意：这里需要传入 slog.Logger，Service 无法直接访问 gin.Context
	// 为了保持 Handler 逻辑的迁移，我们直接使用 slog.With 构造 Logger
	logger := slog.With(
		slog.Int64("id", firstDomainMonitor.ID),
		slog.String("domain", firstDomainMonitor.Domain),
		slog.String("protocol", firstDomainMonitor.CheckProtocol.String()),
	)

	err = domainMonitorService.ExecuteMonitorTask(ctx, logger, firstDomainMonitor)
	if err != nil {
		return nil, fmt.Errorf("检查域名监控失败: %w", err)
	}

	return &types.GetDomainMonitorRequest{
		ID: firstDomainMonitor.ID,
	}, nil
}
