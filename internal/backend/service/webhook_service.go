package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gojson"
)

// WebhookService 通知服务接口
type WebhookService interface {
	BuildResponse(ctx context.Context, item *ent.Webhook) (*types.GetWebhookResponse, error)
	GetWebhooks(ctx context.Context, params *types.GetWebhooksRequest) ([]*types.GetWebhookResponse, error)
	GetWebhook(ctx context.Context, id int64) (*types.GetWebhookResponse, error)
	CreateWebhook(ctx context.Context, request *types.CreateWebhookRequest) error
	UpdateWebhook(ctx context.Context, request *types.UpdateWebhookRequest) error
	UpdateWebhookActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateWebhookActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteWebhook(ctx context.Context, id int64) error
	BatchDeleteWebhook(ctx context.Context, ids []int64) error
	TestWebhook(ctx context.Context, id int64) error
}

// WebhookServiceImpl 通知服务实现
type WebhookServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewWebhookService 创建通知服务实例
func NewWebhookService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) WebhookService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &WebhookServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// SetDebug 设置调试模式
func (s *WebhookServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *WebhookServiceImpl) BuildResponse(ctx context.Context, item *ent.Webhook) (*types.GetWebhookResponse, error) {
	resp := &types.GetWebhookResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		Platform:        item.Platform.String(),
		URL:             item.URL,
		Secret:          item.Secret,
		IsActive:        item.IsActive,
		TemplateTitle:   item.TemplateTitle,
		TemplateContent: item.TemplateContent,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	return resp, nil
}

// GetWebhooks 获取全部通知
func (s *WebhookServiceImpl) GetWebhooks(ctx context.Context, request *types.GetWebhooksRequest) ([]*types.GetWebhookResponse, error) {

	// 查询全部通知
	allWebhook, err := s.db.Webhook.Query().
		WhereIf(request.IsActive != nil, webhook.IsActiveEQ(cast.ToBool(request.IsActive))).
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}

	lists := make([]*types.GetWebhookResponse, 0, len(allWebhook))
	for _, item := range allWebhook {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetWebhook 获取单个通知
func (s *WebhookServiceImpl) GetWebhook(ctx context.Context, id int64) (*types.GetWebhookResponse, error) {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstWebhook)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWebhook 创建通知
func (s *WebhookServiceImpl) CreateWebhook(ctx context.Context, request *types.CreateWebhookRequest) error {
	// 创建通知
	err := s.db.Webhook.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetPlatform(webhook.Platform(request.Platform)).
		SetURL(request.URL).
		SetSecret(request.Secret).
		SetTemplateTitle(request.TemplateTitle).
		SetTemplateContent(request.TemplateContent).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建通知失败: %w", err)
	}

	return nil
}

// UpdateWebhook 更新通知
func (s *WebhookServiceImpl) UpdateWebhook(ctx context.Context, request *types.UpdateWebhookRequest) error {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知失败: %w", err)
	}

	// 更新通知
	err = s.db.Webhook.UpdateOne(firstWebhook).
		SetName(request.Name).
		SetDescription(request.Description).
		SetURL(request.URL).
		SetSecret(request.Secret).
		SetTemplateTitle(request.TemplateTitle).
		SetTemplateContent(request.TemplateContent).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知失败: %w", err)
	}

	return nil
}

// UpdateWebhookActive 更新通知启用状态
func (s *WebhookServiceImpl) UpdateWebhookActive(ctx context.Context, id int64, isActive bool) error {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知失败: %w", err)
	}

	// 更新通知状态
	err = s.db.Webhook.UpdateOne(firstWebhook).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知状态失败: %w", err)
	}

	return nil
}

// BatchUpdateWebhookActive 更新通知启用状态
func (s *WebhookServiceImpl) BatchUpdateWebhookActive(ctx context.Context, ids []int64, isActive bool) error {
	// 查询通知
	allWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知失败: %w", err)
	}
	if len(allWebhook) != len(ids) {
		return fmt.Errorf("查询通知失败: 数量 %d 与请求数量 %d 不匹配", len(allWebhook), len(ids))
	}

	// 更新通知状态
	err = s.db.Webhook.Update().
		Where(webhook.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知状态失败: %w", err)
	}

	return nil
}

// DeleteWebhook 删除通知
func (s *WebhookServiceImpl) DeleteWebhook(ctx context.Context, id int64) error {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知失败: %w", err)
	}
	// 删除通知
	err = s.db.Webhook.DeleteOne(firstWebhook).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}

	return nil
}

// BatchDeleteWebhook 批量删除通知
func (s *WebhookServiceImpl) BatchDeleteWebhook(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询通知
	allWebhook, err := tx.Webhook.Query().
		Where(webhook.IDIn(ids...)).
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询通知失败: %w", err)
	}
	if len(allWebhook) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询通知失败: 数量 %d 与请求数量 %d 不匹配", len(allWebhook), len(ids))
	}

	// 检查是否存在启用的通知
	for _, item := range allWebhook {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除通知失败: 存在启用的通知")
		}
	}

	// 删除通知
	_, err = tx.Webhook.Delete().
		Where(webhook.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除通知失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// TestWebhook 测试通知
func (s *WebhookServiceImpl) TestWebhook(ctx context.Context, id int64) error {

	// 通知模板数据
	notifyData := types.TemplateData{
		EventTitle: "测试事件标题",
		EventType:  "TEST",
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		EventBody:  "测试通知事件",
		DeviceCode: "测试设备代码",
		DeviceName: "测试设备名称",
		IPv4Addr:   "0.0.0.0",
		IPv6Addr:   "0000:0000:0000:0000:0000:0000:0000:0000",
	}

	// 创建任务
	err := s.db.CronJob.Create().
		SetName("测试通知").
		SetHandlerName("webhook_send").
		SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(notifyData))).
		SetTaskType(cronjob.TaskTypeOneOff).
		SetJobType(cronjob.JobTypeDuration).
		SetJobCategory(cronjob.JobCategorySystem).
		SetBusinessType(cronjob.BusinessTypeWebhook).
		SetBusinessID(0).
		SetIntervalValue(1).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(cronjob.StatusRunning).
		SetIsSingleton(true). // 单例模式
		SetRunOnStart(true).  // 启动时立即运行
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}
