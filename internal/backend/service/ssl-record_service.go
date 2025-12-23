package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/helper/ssl_helper"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
)

// SSLRecordService SSL 记录服务接口
type SSLRecordService interface {
	BuildResponse(ctx context.Context, item *ent.SSLRecord) (*types.GetSSLRecordResponse, error)
	GetSSLRecords(ctx context.Context, request *types.GetSSLRecordsRequest) ([]*types.GetSSLRecordResponse, error)
	GetSSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordResponse, error)
	CreateSSLRecord(ctx context.Context, request *types.CreateSSLRecordRequest) (*types.GetSSLRecordRequest, error)
	UpdateSSLRecord(ctx context.Context, request *types.UpdateSSLRecordRequest) (*types.GetSSLRecordRequest, error)
	UpdateSSLRecordAutoRenewal(ctx context.Context, request *types.UpdateSSLRecordAutoRenewalRequest) (*types.GetSSLRecordRequest, error)
	UpdateSSLRecordActive(ctx context.Context, request *types.UpdateSSLRecordActiveRequest) (*types.GetSSLRecordRequest, error)
	DeleteSSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordRequest, error)
	ApplySSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordRequest, error)
}

// SSLRecordServiceImpl SSL 记录服务实现
type SSLRecordServiceImpl struct {
	db            *ent.Client
	debug         bool
	logHelper     *operation_logs_helper.OperationLogHelper
	sslUpdateTask *task.SSLUpdateTask
}

// NewSSLRecordService 创建 SSL 记录服务实例
func NewSSLRecordService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper, sslUpdateTask *task.SSLUpdateTask) SSLRecordService {
	return &SSLRecordServiceImpl{
		db:            db,
		debug:         debug,
		logHelper:     logHelper,
		sslUpdateTask: sslUpdateTask,
	}
}

// BuildResponse 将实体转换为响应
func (s *SSLRecordServiceImpl) BuildResponse(ctx context.Context, item *ent.SSLRecord) (*types.GetSSLRecordResponse, error) {
	resp := &types.GetSSLRecordResponse{
		ID:                    item.ID,
		ConfigID:              item.ConfigID,
		Domain:                item.Domain,
		CertDomain:            item.CertDomain,
		CertURL:               item.CertURL,
		CertStableURL:         item.CertStableURL,
		CertPrivateKey:        item.CertPrivateKey,
		CertCertificate:       item.CertCertificate,
		CertIssuerCertificate: item.CertIssuerCertificate,
		CertCsr:               item.CertCsr,
		NotBefore:             item.NotBefore,
		NotAfter:              item.NotAfter,
		Status:                item.Status.String(),
		Provider:              item.Provider.String(),
		IsActive:              item.IsActive,
		IsAutoRenewal:         item.IsAutoRenewal,
		LastRefreshAt:         item.LastRefreshAt,
	}

	// 平台配置
	if item.Edges.Config != nil {
		resp.ConfigName = item.Edges.Config.Name
		resp.Platform = item.Edges.Config.Platform.String()
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeSslRecord)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetSSLRecords 获取全部 SSL 记录
func (s *SSLRecordServiceImpl) GetSSLRecords(ctx context.Context, request *types.GetSSLRecordsRequest) ([]*types.GetSSLRecordResponse, error) {
	lists := make([]*types.GetSSLRecordResponse, 0)
	// 查询全部SSL记录
	allSSLRecord, err := s.db.SSLRecord.Query().
		WhereIf(request.IsActive != nil, sslrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		Order(ent.Desc(sslrecord.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	for _, item := range allSSLRecord {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetSSLRecord 获取单个 SSL 记录
func (s *SSLRecordServiceImpl) GetSSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordResponse, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		WithConfig().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstSSLRecord)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateSSLRecord 创建 SSL 记录
func (s *SSLRecordServiceImpl) CreateSSLRecord(ctx context.Context, request *types.CreateSSLRecordRequest) (*types.GetSSLRecordRequest, error) {
	// 创建SSL记录
	saveSSLRecord, err := s.db.SSLRecord.Create().
		SetConfigID(request.ConfigID).
		SetDomain(request.Domain).
		SetProvider(sslrecord.Provider(request.Provider)).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("创建SSL记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建SSL记录失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataCreate(ctx, saveSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "创建SSL记录")

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeSslRecord, saveSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetSSLRecordRequest{
		ID: saveSSLRecord.ID,
	}, nil
}

// UpdateSSLRecord 更新 SSL 记录
func (s *SSLRecordServiceImpl) UpdateSSLRecord(ctx context.Context, request *types.UpdateSSLRecordRequest) (*types.GetSSLRecordRequest, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 更新SSL记录
	saveSSLRecord, err := s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetProvider(sslrecord.Provider(request.Provider)).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新SSL记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新CDN记录失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, firstSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "更新SSL记录")

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeSslRecord, saveSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetSSLRecordRequest{
		ID: saveSSLRecord.ID,
	}, nil
}

// UpdateSSLRecordAutoRenewal 更新 SSL 记录自动续期状态
func (s *SSLRecordServiceImpl) UpdateSSLRecordAutoRenewal(ctx context.Context, request *types.UpdateSSLRecordAutoRenewalRequest) (*types.GetSSLRecordRequest, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 更新SSL记录状态
	saveSSLRecord, err := s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetIsAutoRenewal(request.IsAutoRenewal).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新SSL记录状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新SSL记录状态失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, saveSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "更新SSL记录状态")

	return &types.GetSSLRecordRequest{
		ID: saveSSLRecord.ID,
	}, nil
}

// UpdateSSLRecordActive 更新 SSL 记录启用状态
func (s *SSLRecordServiceImpl) UpdateSSLRecordActive(ctx context.Context, request *types.UpdateSSLRecordActiveRequest) (*types.GetSSLRecordRequest, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 更新SSL记录状态
	saveSSLRecord, err := s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新SSL记录状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新SSL记录状态失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, saveSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "更新SSL记录状态")

	return &types.GetSSLRecordRequest{
		ID: saveSSLRecord.ID,
	}, nil
}

// DeleteSSLRecord 删除 CDN 记录
func (s *SSLRecordServiceImpl) DeleteSSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordRequest, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 删除SSL记录
	err = s.db.SSLRecord.DeleteOne(firstSSLRecord).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("删除SSL记录失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataDelete(ctx, firstSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "删除SSL记录")

	// 删除通知订阅
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeSslRecord, firstSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetSSLRecordRequest{
		ID: firstSSLRecord.ID,
	}, nil
}

// ApplySSLRecord 申请新的SSL证书
func (s *SSLRecordServiceImpl) ApplySSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordRequest, error) {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 申请新的SSL证书
	err = s.sslUpdateTask.SSLUpdate(
		ctx,
		slog.Default(),
		&task.PendingSSLRecord{
			Item:     firstSSLRecord,
			Platform: firstSSLRecord.Edges.Config.Platform,
			Params: &ssl_helper.SSLProcessParams{
				SSLRecordID: cast.ToString(firstSSLRecord.ID),

				ConfigJSON: firstSSLRecord.Edges.Config.ConfigJSON,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("申请新的SSL证书失败: %w", err)
	}

	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, firstSSLRecord.ID, operationlog.ObjTypeSslRecord, operationlog.StatusSuccess, "申请新的SSL证书")

	return &types.GetSSLRecordRequest{
		ID: firstSSLRecord.ID,
	}, nil
}
