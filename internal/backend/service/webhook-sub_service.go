package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/ipgroup"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
)

// WebhookSubService 通知订阅服务接口
type WebhookSubService interface {
	BuildResponse(ctx context.Context, item *ent.WebhookSubscription) (*types.GetWebhookSubResponse, error)
	GetWebhookSubs(ctx context.Context, request *types.GetWebhookSubsRequest) ([]*types.GetWebhookSubResponse, error)
	GetWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubResponse, error)
	CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) (*types.CreateWebhookSubResponse, error)
	UpdateWebhookSub(ctx context.Context, request *types.UpdateWebhookSubRequest) (*types.GetWebhookSubRequest, error)
	UpdateWebhookSubActive(ctx context.Context, request *types.UpdateWebhookSubActiveRequest) (*types.GetWebhookSubRequest, error)
	DeleteWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubRequest, error)
}

// WebhookSubServiceImpl 通知订阅服务实现
type WebhookSubServiceImpl struct {
	db        *ent.Client // 数据库客户端
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewWebhookSubService 创建通知订阅服务实例
func NewWebhookSubService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) WebhookSubService {
	return &WebhookSubServiceImpl{db: db, debug: debug, logHelper: logHelper}
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
			configJSON, err := schematype.ParsePCJODomainName(firstCDNRecord.Edges.Config.ConfigJSON)
			if err == nil {
				resp.ObjName = firstCDNRecord.SubDomain + "." + configJSON.DomainName
			} else {
				resp.ObjName = firstCDNRecord.SubDomain
			}
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
			configJSON, err := schematype.ParsePCJODomainName(firstDNSRecord.Edges.Config.ConfigJSON)
			if err == nil {
				resp.ObjName = firstDNSRecord.SubDomain + "." + configJSON.DomainName
			} else {
				resp.ObjName = firstDNSRecord.SubDomain
			}
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
	lists := make([]*types.GetWebhookSubResponse, 0)
	// 查询全部通知订阅
	allWebhookSubscription, err := s.db.WebhookSubscription.Query().
		WhereIf(request.IsActive != nil, webhooksubscription.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithWebhook().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}

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
func (s *WebhookSubServiceImpl) CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) (*types.CreateWebhookSubResponse, error) {
	// 原始 Handler 逻辑是在循环内部处理错误并返回，导致只有第一个错误会被返回。
	// 这里保持 Handler 的逻辑结构，尽管它会在第一个 ObjID 失败时立即返回。
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
			if s.debug {
				fmt.Printf("创建通知订阅: %s\n", err.Error())
			}
			// 立即返回错误
			return nil, fmt.Errorf("创建通知订阅失败: %w", err)
		}
		// 记录操作日志
		// _ = s.logHelper.RecordDataCreate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "创建通知订阅")
	}

	return &types.CreateWebhookSubResponse{
		IDs: request.ObjIDs,
	}, nil
}

// UpdateWebhookSub 更新通知订阅
func (s *WebhookSubServiceImpl) UpdateWebhookSub(ctx context.Context, request *types.UpdateWebhookSubRequest) (*types.GetWebhookSubRequest, error) {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}
	// 更新通知订阅
	saveWebhookSubscription, err := s.db.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetName(request.Name).
		SetDescription(request.Description).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新通知订阅失败: %w", err)
	}
	// 记录操作日志
	// _ = s.logHelper.RecordDataUpdate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅")
	return &types.GetWebhookSubRequest{
		ID: saveWebhookSubscription.ID,
	}, nil
}

// UpdateWebhookSubActive 更新通知订阅启用状态
func (s *WebhookSubServiceImpl) UpdateWebhookSubActive(ctx context.Context, request *types.UpdateWebhookSubActiveRequest) (*types.GetWebhookSubRequest, error) {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}
	// 更新通知订阅状态
	saveWebhookSubscription, err := s.db.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新通知订阅状态失败: %w", err)
	}
	// 记录操作日志
	// _ = s.logHelper.RecordDataUpdate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅状态")
	return &types.GetWebhookSubRequest{
		ID: saveWebhookSubscription.ID,
	}, nil
}

// DeleteWebhookSub 删除通知订阅
func (s *WebhookSubServiceImpl) DeleteWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubRequest, error) {
	// 查询通知订阅
	firstWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询通知订阅失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}
	// 删除通知订阅
	err = s.db.WebhookSubscription.DeleteOne(firstWebhookSubscription).Exec(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除通知订阅失败: %w", err)
	}
	// 记录操作日志
	// _ = s.logHelper.RecordDataDelete(ctx, firstWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "删除通知订阅")
	return &types.GetWebhookSubRequest{
		ID: firstWebhookSubscription.ID,
	}, nil
}
