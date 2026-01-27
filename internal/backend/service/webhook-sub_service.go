package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/ipgroup"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

// WebhookSubService 通知订阅服务接口
type WebhookSubService interface {
	BuildResponse(ctx context.Context, item *ent.WebhookSubscription) (*types.GetWebhookSubResponse, error)
	GetWebhookSubs(ctx context.Context, request *types.GetWebhookSubsRequest) ([]*types.GetWebhookSubResponse, error)
	GetWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubResponse, error)
	CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) error
	UpdateWebhookSub(ctx context.Context, request *types.UpdateWebhookSubRequest) error
	UpdateWebhookSubActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateWebhookSubActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteWebhookSub(ctx context.Context, id int64) error
	BatchDeleteWebhookSub(ctx context.Context, ids []int64) error
}

// WebhookSubServiceImpl 通知订阅服务实现
type WebhookSubServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewWebhookSubService 创建通知订阅服务实例
func NewWebhookSubService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) WebhookSubService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &WebhookSubServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// SetDebug 设置调试模式
func (s *WebhookSubServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *WebhookSubServiceImpl) BuildResponse(ctx context.Context, item *ent.WebhookSubscription) (*types.GetWebhookSubResponse, error) {
	resp := &types.GetWebhookSubResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		WebhookID:   item.WebhookID,
		WebhookName: item.Edges.Webhook.Name,
		ObjType:     item.ObjType.String(),
		ObjID:       item.ObjID,
		IsActive:    item.IsActive,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	// 查询关联对象名称
	switch item.ObjType {
	case webhooksubscription.ObjTypeCdnRecord:
		firstCDNRecord, err := s.db.CDNRecord.Query().
			Where(cdnrecord.IDEQ(item.ObjID)).
			WithConfig().
			Select(
				cdnrecord.FieldSubDomain,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstCDNRecord.SubDomain + "." + gjson.GetBytes(firstCDNRecord.Edges.Config.ConfigJSON, "domain_name").String()
		}
	case webhooksubscription.ObjTypeDNSRecord:
		firstDNSRecord, err := s.db.DNSRecord.Query().
			Where(dnsrecord.IDEQ(item.ObjID)).
			WithConfig().
			Select(
				dnsrecord.FieldSubDomain,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstDNSRecord.SubDomain + "." + gjson.GetBytes(firstDNSRecord.Edges.Config.ConfigJSON, "domain_name").String()
		}
	case webhooksubscription.ObjTypeDevice:
		firstDevice, err := s.db.Device.Query().
			Where(device.IDEQ(item.ObjID)).
			Select(
				device.FieldName,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstDevice.Name
		}
	case webhooksubscription.ObjTypeIPGroup:
		firstIPGroup, err := s.db.IpGroup.Query().
			Where(ipgroup.IDEQ(item.ObjID)).
			Select(
				ipgroup.FieldName,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstIPGroup.Name
		}
	case webhooksubscription.ObjTypeDomainMonitor:
		firstDomainMonitor, err := s.db.DomainMonitor.Query().
			Where(domainmonitor.IDEQ(item.ObjID)).
			Select(
				domainmonitor.FieldName,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstDomainMonitor.Name
		}
	case webhooksubscription.ObjTypeSslRecord:
		firstSSLRecord, err := s.db.SSLRecord.Query().
			Where(sslrecord.IDEQ(item.ObjID)).
			Select(
				sslrecord.FieldDomain,
			).
			First(ctx)
		if err == nil {
			resp.ObjName = firstSSLRecord.Domain
		}
	}

	return resp, nil
}

// GetWebhookSubs 获取全部通知订阅
func (s *WebhookSubServiceImpl) GetWebhookSubs(ctx context.Context, request *types.GetWebhookSubsRequest) ([]*types.GetWebhookSubResponse, error) {

	// 查询全部通知订阅
	allWebhookSubscription, err := s.db.WebhookSubscription.Query().
		WhereIf(request.IsActive != nil, webhooksubscription.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithWebhook().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}

	lists := make([]*types.GetWebhookSubResponse, 0, len(allWebhookSubscription))
	for _, item := range allWebhookSubscription {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetWebhookSub 获取单个通知订阅
func (s *WebhookSubServiceImpl) GetWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubResponse, error) {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(id)).
		WithWebhook().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstWebhookSubscription)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWebhookSub 创建通知订阅
func (s *WebhookSubServiceImpl) CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) error {
	// 循环创建通知订阅
	for _, objID := range request.ObjIDs {
		// 创建通知订阅
		err := s.db.WebhookSubscription.Create().
			SetName(request.Name).
			SetDescription(request.Description).
			SetWebhookID(request.WebhookID).
			SetObjType(webhooksubscription.ObjType(request.ObjType)).
			SetObjID(objID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建通知订阅失败: %w", err)
		}
	}

	return nil
}

// UpdateWebhookSub 更新通知订阅
func (s *WebhookSubServiceImpl) UpdateWebhookSub(ctx context.Context, request *types.UpdateWebhookSubRequest) error {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}

	// 更新通知订阅
	err = s.db.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetName(request.Name).
		SetDescription(request.Description).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知订阅失败: %w", err)
	}

	return nil
}

// UpdateWebhookSubActive 更新通知订阅启用状态
func (s *WebhookSubServiceImpl) UpdateWebhookSubActive(ctx context.Context, id int64, isActive bool) error {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}

	// 更新通知订阅状态
	err = s.db.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知订阅状态失败: %w", err)
	}

	return nil
}

// BatchUpdateWebhookSubActive 更新通知订阅启用状态
func (s *WebhookSubServiceImpl) BatchUpdateWebhookSubActive(ctx context.Context, ids []int64, isActive bool) error {
	// 查询通知订阅
	allWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}
	if len(allWebhookSubscription) != len(ids) {
		return fmt.Errorf("查询通知订阅失败: 数量 %d 与请求数量 %d 不匹配", len(allWebhookSubscription), len(ids))
	}

	// 更新通知订阅状态
	err = s.db.WebhookSubscription.Update().
		Where(webhooksubscription.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新通知订阅状态失败: %w", err)
	}

	return nil
}

// DeleteWebhookSub 删除通知订阅
func (s *WebhookSubServiceImpl) DeleteWebhookSub(ctx context.Context, id int64) error {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}
	// 删除通知订阅
	err = s.db.WebhookSubscription.DeleteOne(firstWebhookSubscription).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除通知订阅失败: %w", err)
	}

	return nil
}

// BatchDeleteWebhookSub 批量删除通知订阅
func (s *WebhookSubServiceImpl) BatchDeleteWebhookSub(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询通知订阅
	allWebhookSubscription, err := tx.WebhookSubscription.Query().
		Where(webhooksubscription.IDIn(ids...)).
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}
	if len(allWebhookSubscription) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询通知订阅失败: 数量 %d 与请求数量 %d 不匹配", len(allWebhookSubscription), len(ids))
	}

	// 检查是否存在启用的通知订阅
	for _, item := range allWebhookSubscription {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除通知订阅失败: 存在启用的通知订阅")
		}
	}

	// 删除通知订阅
	_, err = tx.WebhookSubscription.Delete().
		Where(webhooksubscription.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除通知订阅失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
