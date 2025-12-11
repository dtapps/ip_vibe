package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global" // 仅用于访问 GvaOperationLogHelper 和类型定义
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// WebhookSubService 通知订阅服务接口
type WebhookSubService interface {
	GetWebhookSubs(ctx context.Context) ([]types.GetWebhookSubResponse, error)
	GetWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubResponse, error)
	CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) (*gin.H, error)
	UpdateWebhookSub(ctx context.Context, request *types.UpdateWebhookSubRequest) (*types.GetWebhookSubRequest, error)
	UpdateWebhookSubActive(ctx context.Context, request *types.UpdateWebhookSubActiveRequest) (*types.GetWebhookSubRequest, error)
	DeleteWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubRequest, error)
}

// WebhookSubServiceImpl 通知订阅服务实现
type WebhookSubServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewWebhookSubService 创建通知订阅服务实例
func NewWebhookSubService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) WebhookSubService {
	return &WebhookSubServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// getObjNameFromID 根据 ObjType 和 ObjID 查询关联对象的名称
func (s *WebhookSubServiceImpl) getObjNameFromID(ctx context.Context, objType webhooksubscription.ObjType, objID int64) string {
	var objName string
	var err error

	switch objType {
	case webhooksubscription.ObjTypeCdnRecord:
		var firstCDNRecord *ent.CDNRecord
		firstCDNRecord, err = s.db.CDNRecord.Query().
			Where(cdnrecord.IDEQ(objID)).
			First(ctx)
		if err == nil {
			objName = firstCDNRecord.Domain
		}
	case webhooksubscription.ObjTypeDNSRecord:
		var firstDNSRecord *ent.DNSRecord
		firstDNSRecord, err = s.db.DNSRecord.Query().
			Where(dnsrecord.IDEQ(objID)).
			First(ctx)
		if err == nil {
			objName = firstDNSRecord.Domain
		}
	case webhooksubscription.ObjTypeDevice:
		var firstDevice *ent.Device
		firstDevice, err = s.db.Device.Query().
			Where(device.IDEQ(objID)).
			First(ctx)
		if err == nil {
			objName = firstDevice.Name
		}
	}

	if err != nil && s.debug {
		fmt.Printf("查询 ObjType: %s, ObjID: %d 的名称失败: %s\n", objType.String(), objID, err.Error())
	}
	return objName
}

// GetWebhookSubs 获取全部通知订阅
func (s *WebhookSubServiceImpl) GetWebhookSubs(ctx context.Context) ([]types.GetWebhookSubResponse, error) {
	webhooks := make([]types.GetWebhookSubResponse, 0)
	// 查询全部通知订阅
	allWebhookSubscription, err := s.db.WebhookSubscription.Query().
		Order(ent.Desc(webhooksubscription.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询通知订阅失败: %w", err)
	}

	for _, item := range allWebhookSubscription {
		list := types.GetWebhookSubResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			WebhookID:   item.WebhookID,
			ObjType:     item.ObjType.String(),
			ObjID:       item.ObjID,
			IsActive:    item.IsActive,
		}
		// 查询关联对象名称 (保持原始逻辑，在循环内进行查询)
		list.ObjName = s.getObjNameFromID(ctx, item.ObjType, item.ObjID)

		// 原始 Handler 逻辑中这里有一个错误：webhooks = append(webhooks) 应该是 webhooks = append(webhooks, list)
		// 迁移时假设意图是追加
		webhooks = append(webhooks, list)
	}
	return webhooks, nil
}

// GetWebhookSub 获取单个通知订阅
func (s *WebhookSubServiceImpl) GetWebhookSub(ctx context.Context, id int64) (*types.GetWebhookSubResponse, error) {
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

	list := types.GetWebhookSubResponse{
		ID:          firstWebhookSubscription.ID,
		Name:        firstWebhookSubscription.Name,
		Description: firstWebhookSubscription.Description,
		WebhookID:   firstWebhookSubscription.WebhookID,
		ObjType:     firstWebhookSubscription.ObjType.String(),
		ObjID:       firstWebhookSubscription.ObjID,
		IsActive:    firstWebhookSubscription.IsActive,
	}
	// 查询关联对象名称
	list.ObjName = s.getObjNameFromID(ctx, firstWebhookSubscription.ObjType, firstWebhookSubscription.ObjID)

	return &list, nil
}

// CreateWebhookSub 创建通知订阅
func (s *WebhookSubServiceImpl) CreateWebhookSub(ctx context.Context, request *types.CreateWebhookSubRequest) (*gin.H, error) {
	// 原始 Handler 逻辑是在循环内部处理错误并返回，导致只有第一个错误会被返回。
	// 这里保持 Handler 的逻辑结构，尽管它会在第一个 ObjID 失败时立即返回。
	for _, objID := range request.ObjIDs {
		// 创建通知订阅
		saveWebhookSubscription, err := s.db.WebhookSubscription.Create().
			SetName(request.Name).
			SetDescription(request.Description).
			SetWebhookID(request.WebhookID).
			SetObjType(webhooksubscription.ObjType(request.ObjType)).
			SetObjID(objID).
			Save(ctx)
		if err != nil {
			if s.debug {
				fmt.Printf("创建通知订阅: %s\n", err.Error())
			}
			// 立即返回错误
			return nil, fmt.Errorf("创建通知订阅失败: %w", err)
		}
		// 记录操作日志
		_ = global.GvaOperationLogHelper.RecordDataCreate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "创建通知订阅")
	}

	// 原始 Handler 返回的是 gin.H{ "ids": requestParams.ObjIDs }
	return &gin.H{
		"ids": request.ObjIDs,
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
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅")
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
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅状态")
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
	_ = global.GvaOperationLogHelper.RecordDataDelete(ctx, firstWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "删除通知订阅")
	return &types.GetWebhookSubRequest{
		ID: firstWebhookSubscription.ID,
	}, nil
}
