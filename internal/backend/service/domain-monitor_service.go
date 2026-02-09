package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/provider/task"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
)

// DomainMonitorService 域名监控服务接口
type DomainMonitorService interface {
	BuildResponse(ctx context.Context, item *ent.DomainMonitor) (*types.GetDomainMonitorResponse, error)
	GetDomainMonitors(ctx context.Context, request *types.GetDomainMonitorsRequest) ([]*types.GetDomainMonitorResponse, error)
	GetDomainMonitor(ctx context.Context, id int64) (*types.GetDomainMonitorResponse, error)
	CreateDomainMonitor(ctx context.Context, request *types.CreateDomainMonitorRequest) error
	UpdateDomainMonitor(ctx context.Context, request *types.UpdateDomainMonitorRequest) error
	UpdateDomainMonitorActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateDomainMonitorActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteDomainMonitor(ctx context.Context, id int64) error
	BatchDeleteDomainMonitor(ctx context.Context, ids []int64) error
	CheckDomainMonitor(ctx context.Context, id int64) error
}

// DomainMonitorServiceImpl 域名监控服务实现
type DomainMonitorServiceImpl struct {
	db                   *ent.Client                     // 数据库客户端
	logDb                *ent_log.Client                 // 日志数据库客户端
	cron                 gocron.Scheduler                // 定时任务调度器
	logger               *slog.Logger                    // 日志记录器
	httpTransport        *http_log.LoggingRoundTripper   // HTTP日志记录器
	domainMonitorHandler *task.DomainMonitorTaskProvider // CDN同步 处理器
	debug                bool                            // 调试模式
}

// NewDomainMonitorService 创建域名监控服务实例
func NewDomainMonitorService(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) DomainMonitorService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	// 创建 域名监控 处理器
	domainMonitorHandler := task.NewDomainMonitorTaskProviderRaw(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &DomainMonitorServiceImpl{
		db:                   db,
		logDb:                logDb,
		cron:                 cron,
		logger:               logger,
		httpTransport:        httpTransport,
		domainMonitorHandler: domainMonitorHandler,
		debug:                debug,
	}
}

// SetDebug 设置调试模式
func (s *DomainMonitorServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *DomainMonitorServiceImpl) BuildResponse(ctx context.Context, item *ent.DomainMonitor) (*types.GetDomainMonitorResponse, error) {
	resp := &types.GetDomainMonitorResponse{
		ID:             item.ID,
		Name:           item.Name,
		Description:    item.Description,
		Domain:         item.Domain,
		RequestHeaders: item.RequestHeaders,
		CheckInterval:  item.CheckInterval,
		CheckProtocol:  item.CheckProtocol.String(),
		IsActive:       item.IsActive,
		CurrentStatus:  item.CurrentStatus.String(),
		LastSuccessAt:  item.LastSuccessAt,
		LastFailureAt:  item.LastFailureAt,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetDomainMonitors 获取全部域名监控
func (s *DomainMonitorServiceImpl) GetDomainMonitors(ctx context.Context, request *types.GetDomainMonitorsRequest) ([]*types.GetDomainMonitorResponse, error) {

	// 查询全部域名监控
	allDomainMonitor, err := s.db.DomainMonitor.Query().
		WhereIf(request.IsActive != nil, domainmonitor.IsActiveEQ(cast.ToBool(request.IsActive))).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询域名监控失败: %w", err)
	}

	lists := make([]*types.GetDomainMonitorResponse, 0, len(allDomainMonitor))
	for _, item := range allDomainMonitor {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
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

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstDomainMonitor)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateDomainMonitor 创建域名监控
func (s *DomainMonitorServiceImpl) CreateDomainMonitor(ctx context.Context, request *types.CreateDomainMonitorRequest) error {
	// 创建域名监控
	saveDomainMonitor, err := s.db.DomainMonitor.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetDomain(request.Domain).
		SetRequestHeaders(request.RequestHeaders).
		SetCheckInterval(request.CheckInterval).
		SetCheckProtocol(domainmonitor.CheckProtocol(request.CheckProtocol)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("创建域名监控失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDomainMonitor, saveDomainMonitor.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateDomainMonitor 更新域名监控
func (s *DomainMonitorServiceImpl) UpdateDomainMonitor(ctx context.Context, request *types.UpdateDomainMonitorRequest) error {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询域名监控失败: %w", err)
	}
	// 更新域名监控
	saveDomainMonitor, err := s.db.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetName(request.Name).
		SetDescription(request.Description).
		SetRequestHeaders(request.RequestHeaders).
		SetCheckInterval(request.CheckInterval).
		SetCheckProtocol(domainmonitor.CheckProtocol(request.CheckProtocol)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新域名监控失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDomainMonitor, saveDomainMonitor.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateDomainMonitorActive 更新域名监控启用状态
func (s *DomainMonitorServiceImpl) UpdateDomainMonitorActive(ctx context.Context, id int64, isActive bool) error {

	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询域名监控失败: %w", err)
	}

	// 更新域名监控状态
	err = s.db.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新域名监控启用状态失败: %w", err)
	}

	// 加载 域名监控 周期任务
	return s.domainMonitorHandler.LoadTask(ctx, s.db, firstDomainMonitor, isActive)
}

// BatchUpdateDomainMonitorActive 批量更新域名监控启用状态
func (s *DomainMonitorServiceImpl) BatchUpdateDomainMonitorActive(ctx context.Context, ids []int64, isActive bool) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询域名监控
	allDomainMonitor, err := tx.DomainMonitor.Query().
		Where(domainmonitor.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询域名监控失败: %w", err)
	}
	if len(allDomainMonitor) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询域名监控失败: 数量 %d 与请求数量 %d 不匹配", len(allDomainMonitor), len(ids))
	}

	// 更新域名监控状态
	err = tx.DomainMonitor.Update().
		Where(domainmonitor.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新域名监控启用状态失败: %w", err)
	}

	// 加载 域名监控 周期任务
	for _, record := range allDomainMonitor {
		err = s.domainMonitorHandler.LoadTask(ctx, tx.Client(), record, isActive)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("添加域名监控同步周期任务失败: %w", err)
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DeleteDomainMonitor 删除域名监控
func (s *DomainMonitorServiceImpl) DeleteDomainMonitor(ctx context.Context, id int64) error {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询域名监控失败: %w", err)
	}
	// 删除域名监控
	err = s.db.DomainMonitor.DeleteOne(firstDomainMonitor).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除域名监控失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).Delete(ctx, webhooksubscription.ObjTypeDomainMonitor, firstDomainMonitor.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// BatchDeleteDomainMonitor 批量删除域名监控
func (s *DomainMonitorServiceImpl) BatchDeleteDomainMonitor(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询域名监控
	allDomainMonitor, err := tx.DomainMonitor.Query().
		Where(domainmonitor.IDIn(ids...)).
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询域名监控失败: %w", err)
	}
	if len(allDomainMonitor) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询域名监控失败: 数量 %d 与请求数量 %d 不匹配", len(allDomainMonitor), len(ids))
	}

	// 检查是否存在启用的域名监控
	for _, item := range allDomainMonitor {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除域名监控失败: 存在启用的域名监控")
		}
	}

	// 删除域名监控
	_, err = tx.DomainMonitor.Delete().
		Where(domainmonitor.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除域名监控失败: %w", err)
	}

	// 删除通知订阅
	for _, domainMonitor := range allDomainMonitor {
		err = webhook_sub_helper.NewWebhookSubscription(s.db, s.logger).DeleteTx(ctx, tx, webhooksubscription.ObjTypeDomainMonitor, domainMonitor.ID)
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

// CheckDomainMonitor 检查域名监控（执行一次任务）
func (s *DomainMonitorServiceImpl) CheckDomainMonitor(ctx context.Context, id int64) error {
	// 查询域名监控
	firstDomainMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询域名监控失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询域名监控失败: %w", err)
	}

	// 添加 域名监控 一次性任务
	err = s.domainMonitorHandler.AddTask(ctx, s.db, firstDomainMonitor)
	if err != nil {
		return fmt.Errorf("添加域名监控一次性任务失败: %w", err)
	}

	return nil
}
