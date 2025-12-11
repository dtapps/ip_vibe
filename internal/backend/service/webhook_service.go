package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
)

// WebhookService 通知服务接口
type WebhookService interface {
	GetWebhooks(ctx context.Context, params *types.GetWebhooksRequest) ([]*types.GetWebhookResponse, error)
	GetWebhook(ctx context.Context, id int64) (*types.GetWebhookResponse, error)
	CreateWebhook(ctx context.Context, request *types.CreateWebhookRequest) (*types.GetWebhookRequest, error)
	UpdateWebhook(ctx context.Context, request *types.UpdateWebhookRequest) (*types.GetWebhookRequest, error)
	UpdateWebhookActive(ctx context.Context, request *types.UpdateWebhookActiveRequest) (*types.GetWebhookRequest, error)
	DeleteWebhook(ctx context.Context, id int64) (*types.GetWebhookRequest, error)
}

// WebhookServiceImpl 通知服务实现
type WebhookServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewWebhookService 创建通知服务实例
func NewWebhookService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) WebhookService {
	return &WebhookServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// buildResponse 将实体转换为响应
func (s *WebhookServiceImpl) buildResponse(ctx context.Context, item *ent.Webhook) (*types.GetWebhookResponse, error) {
	resp := &types.GetWebhookResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		Platform:        item.Platform.String(),
		URL:             item.URL,
		Secret:          item.Secret,
		Events:          item.Events,
		IsActive:        item.IsActive,
		TemplateTitle:   item.TemplateTitle,
		TemplateContent: item.TemplateContent,
	}
	return resp, nil
}

// GetWebhooks 获取全部通知
func (s *WebhookServiceImpl) GetWebhooks(ctx context.Context, request *types.GetWebhooksRequest) ([]*types.GetWebhookResponse, error) {
	lists := make([]*types.GetWebhookResponse, 0)
	// 查询全部通知
	allWebhook, err := s.db.Webhook.Query().
		WhereIf(request.IsActive != nil, webhook.IsActiveEQ(cast.ToBool(request.IsActive))).
		Order(ent.Desc(webhook.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	for _, item := range allWebhook {
		// 将实体转换为响应
		list, err := s.buildResponse(ctx, item)
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
	list, err := s.buildResponse(ctx, firstWebhook)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWebhook 创建通知
func (s *WebhookServiceImpl) CreateWebhook(ctx context.Context, request *types.CreateWebhookRequest) (*types.GetWebhookRequest, error) {
	// 创建通知
	saveWebhook, err := s.db.Webhook.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetPlatform(webhook.Platform(request.Platform)).
		SetURL(request.URL).
		SetSecret(request.Secret).
		SetEvents(request.Events).
		SetTemplateTitle(request.TemplateTitle).
		SetTemplateContent(request.TemplateContent).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建通知失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataCreate(ctx, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "创建通知")
	return &types.GetWebhookRequest{
		ID: saveWebhook.ID,
	}, nil
}

// UpdateWebhook 更新通知
func (s *WebhookServiceImpl) UpdateWebhook(ctx context.Context, request *types.UpdateWebhookRequest) (*types.GetWebhookRequest, error) {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	// 更新通知
	saveWebhook, err := s.db.Webhook.UpdateOne(firstWebhook).
		SetName(request.Name).
		SetDescription(request.Description).
		// Platform 和 URL 不允许通过 UpdateWebhook 更新，保持原始 Handler 逻辑
		SetSecret(request.Secret).
		SetEvents(request.Events).
		SetTemplateTitle(request.TemplateTitle).
		SetTemplateContent(request.TemplateContent).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新通知失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "更新通知")
	return &types.GetWebhookRequest{
		ID: saveWebhook.ID,
	}, nil
}

// UpdateWebhookActive 更新通知启用状态
func (s *WebhookServiceImpl) UpdateWebhookActive(ctx context.Context, request *types.UpdateWebhookActiveRequest) (*types.GetWebhookRequest, error) {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	// 更新通知状态
	saveWebhook, err := s.db.Webhook.UpdateOne(firstWebhook).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新通知状态失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "更新通知状态")
	return &types.GetWebhookRequest{
		ID: saveWebhook.ID,
	}, nil
}

// DeleteWebhook 删除通知
func (s *WebhookServiceImpl) DeleteWebhook(ctx context.Context, id int64) (*types.GetWebhookRequest, error) {
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
	// 删除通知
	err = s.db.Webhook.DeleteOne(firstWebhook).Exec(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除通知失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataDelete(ctx, firstWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "删除通知")
	return &types.GetWebhookRequest{
		ID: firstWebhook.ID,
	}, nil
}
