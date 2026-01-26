package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/helper/ssl_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gostring"
)

// SSLRecordService SSL 记录服务接口
type SSLRecordService interface {
	BuildResponse(ctx context.Context, item *ent.SSLRecord) (*types.GetSSLRecordResponse, error)
	GetSSLRecords(ctx context.Context, request *types.GetSSLRecordsRequest) ([]*types.GetSSLRecordResponse, error)
	GetSSLRecord(ctx context.Context, id int64) (*types.GetSSLRecordResponse, error)
	ImportSSLRecord(ctx context.Context, request *types.ImportSSLRecordRequest) error
	CreateSSLRecord(ctx context.Context, request *types.CreateSSLRecordRequest) error
	UpdateSSLRecord(ctx context.Context, request *types.UpdateSSLRecordRequest) error
	UpdateSSLRecordAutoRenewal(ctx context.Context, request *types.UpdateSSLRecordAutoRenewalRequest) error
	UpdateSSLRecordActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateSSLRecordActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteSSLRecord(ctx context.Context, id int64) error
	BatchDeleteSSLRecord(ctx context.Context, ids []int64) error
	ApplySSLRecord(ctx context.Context, id int64) error
	GetSSLRecordHistoryDetail(ctx context.Context, id int64, historyID int64) (*types.GetSSLRecordHistoryDetailResponse, error)
}

// SSLRecordServiceImpl SSL 记录服务实现
type SSLRecordServiceImpl struct {
	db           *ent.Client        // 数据库客户端
	logDb        *ent_log.Client    // 日志数据库客户端
	logger       *slog.Logger       // 日志记录器
	sslApplyTask *task.SSLApplyTask // SSL申请任务
	debug        bool               // 调试模式
}

// NewSSLRecordService 创建 SSL 记录服务实例
func NewSSLRecordService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, sslApplyTask *task.SSLApplyTask) SSLRecordService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &SSLRecordServiceImpl{
		db:           db,
		logDb:        logDb,
		logger:       logger,
		sslApplyTask: sslApplyTask,
		debug:        debug,
	}
}

// SetDebug 设置调试模式
func (s *SSLRecordServiceImpl) SetDebug(debug bool) {
	s.debug = debug
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
		RenewalDays:           item.RenewalDays,
		IsAutoRenewal:         item.IsAutoRenewal,
		LastRefreshAt:         item.LastRefreshAt,

		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		LastAPIResponse:   item.LastAPIResponse,
		LastAPIResponseAt: item.LastAPIResponseAt,
	}

	// 平台配置
	if config, err := item.Edges.ConfigOrErr(); err == nil {
		resp.ConfigName = config.Name
		resp.Platform = config.Platform.String()
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

	// 查询全部SSL记录
	allSSLRecord, err := s.db.SSLRecord.Query().
		WhereIf(request.IsActive != nil, sslrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询SSL记录失败: %w", err)
	}

	lists := make([]*types.GetSSLRecordResponse, 0, len(allSSLRecord))
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

// ImportSSLRecord 导入 SSL 记录
func (s *SSLRecordServiceImpl) ImportSSLRecord(ctx context.Context, request *types.ImportSSLRecordRequest) error {

	// 解析证书
	parsedCert, err := utils.ParseCertificate([]byte(request.CertCertificate))
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 证书状态
	status := sslrecord.StatusUnknown
	if parsedCert.NotAfter.After(time.Now()) {
		status = sslrecord.StatusValid
	} else {
		status = sslrecord.StatusExpired
	}

	// 创建SSL记录
	err = s.db.SSLRecord.Create().
		SetConfigID(request.ConfigID).
		SetDomain(request.Domain).
		SetCertDomain(parsedCert.Domain).
		SetCertPrivateKey(request.CertPrivateKey).
		SetCertCertificate(request.CertCertificate).
		SetNotBefore(parsedCert.NotBefore).
		SetNotAfter(parsedCert.NotAfter).
		SetStatus(status).
		SetProvider(sslrecord.ProviderUnknown).
		SetIsActive(false).
		SetRenewalDays(0).
		SetIsAutoRenewal(false).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建SSL记录失败: %w", err)
	}

	return nil
}

// CreateSSLRecord 创建 SSL 记录
func (s *SSLRecordServiceImpl) CreateSSLRecord(ctx context.Context, request *types.CreateSSLRecordRequest) error {
	// 创建SSL记录
	saveSSLRecord, err := s.db.SSLRecord.Create().
		SetConfigID(request.ConfigID).
		SetDomain(request.Domain).
		SetStatus(sslrecord.StatusPending).
		SetProvider(sslrecord.Provider(request.Provider)).
		SetRenewalDays(request.RenewalDays).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("创建SSL记录失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeSslRecord, saveSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateSSLRecord 更新 SSL 记录
func (s *SSLRecordServiceImpl) UpdateSSLRecord(ctx context.Context, request *types.UpdateSSLRecordRequest) error {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 更新SSL记录
	saveSSLRecord, err := s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetProvider(sslrecord.Provider(request.Provider)).
		SetRenewalDays(request.RenewalDays).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新CDN记录失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeSslRecord, saveSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateSSLRecordAutoRenewal 更新 SSL 记录自动续期状态
func (s *SSLRecordServiceImpl) UpdateSSLRecordAutoRenewal(ctx context.Context, request *types.UpdateSSLRecordAutoRenewalRequest) error {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}

	if firstSSLRecord.Provider == sslrecord.ProviderUnknown {
		return fmt.Errorf("未配置提供方, 无法更新自动续期状态")
	}

	// 更新SSL记录状态
	err = s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetIsAutoRenewal(request.IsAutoRenewal).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新SSL记录状态失败: %w", err)
	}

	return nil
}

// UpdateSSLRecordActive 更新 SSL 记录启用状态
func (s *SSLRecordServiceImpl) UpdateSSLRecordActive(ctx context.Context, id int64, isActive bool) error {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}

	if firstSSLRecord.Provider == sslrecord.ProviderUnknown {
		return fmt.Errorf("未配置提供方, 无法更新启用状态")
	}

	// 更新SSL记录状态
	err = s.db.SSLRecord.UpdateOne(firstSSLRecord).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新SSL记录状态失败: %w", err)
	}

	return nil
}

// BatchUpdateSSLRecordActive 更新 SSL 记录启用状态
func (s *SSLRecordServiceImpl) BatchUpdateSSLRecordActive(ctx context.Context, ids []int64, isActive bool) error {
	// 查询SSL记录
	allSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}
	if len(allSSLRecord) != len(ids) {
		return fmt.Errorf("查询SSL记录失败: 数量 %d 与请求数量 %d 不匹配", len(allSSLRecord), len(ids))
	}

	for _, item := range allSSLRecord {
		if item.Provider == sslrecord.ProviderUnknown {
			return fmt.Errorf("未配置提供方, 无法更新启用状态")
		}
	}

	// 更新SSL记录状态
	err = s.db.SSLRecord.Update().
		Where(sslrecord.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新SSL记录状态失败: %w", err)
	}

	return nil
}

// DeleteSSLRecord 删除 CDN 记录
func (s *SSLRecordServiceImpl) DeleteSSLRecord(ctx context.Context, id int64) error {
	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 删除SSL记录
	err = s.db.SSLRecord.DeleteOne(firstSSLRecord).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除SSL记录失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeSslRecord, firstSSLRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// BatchDeleteSSLRecord 批量删除 SSL 记录
func (s *SSLRecordServiceImpl) BatchDeleteSSLRecord(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询SSL记录
	allSSLRecord, err := tx.SSLRecord.Query().
		Where(sslrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}
	if len(allSSLRecord) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询SSL记录失败: 数量 %d 与请求数量 %d 不匹配", len(allSSLRecord), len(ids))
	}

	// 检查是否存在启用的SSL记录
	for _, item := range allSSLRecord {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除SSL记录失败: 存在启用的SSL记录")
		}
	}

	// 删除SSL记录
	_, err = tx.SSLRecord.Delete().
		Where(sslrecord.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除SSL记录失败: %w", err)
	}

	// 删除通知订阅
	for _, sslRecord := range allSSLRecord {
		err = webhook_sub_helper.DeleteTx(ctx, tx, webhooksubscription.ObjTypeSslRecord, sslRecord.ID)
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

// ApplySSLRecord 申请新的SSL证书
// sslApplyTask 不能为nil
func (s *SSLRecordServiceImpl) ApplySSLRecord(ctx context.Context, id int64) error {
	if s.sslApplyTask == nil {
		return fmt.Errorf("sslApplyTask 不能为nil")
	}

	// 查询SSL记录
	firstSSLRecord, err := s.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(id)).
		WithConfig().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询SSL记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询SSL记录失败: %w", err)
	}

	// 平台配置
	config, err := firstSSLRecord.Edges.ConfigOrErr()
	if err != nil {
		return fmt.Errorf("平台配置失败: %w", err)
	}

	// 添加SSL申请一次性任务
	err = s.sslApplyTask.AddDisposableTask(ctx, &task.ExecuteSSLApplyParams{
		Item:     firstSSLRecord,
		Platform: config.Platform,
		Params: &ssl_helper.SSLProcessParams{
			SSLRecordID: cast.ToString(firstSSLRecord.ID),

			ConfigJSON: config.ConfigJSON,
			Domains:    gostring.Split(firstSSLRecord.Domain, ","),
		},
	})
	if err != nil {
		return fmt.Errorf("添加SSL一次性任务失败: %w", err)
	}

	return nil
}

// GetSSLRecordHistoryDetail 获取SSL记录历史详情
func (s *SSLRecordServiceImpl) GetSSLRecordHistoryDetail(ctx context.Context, id int64, historyID int64) (*types.GetSSLRecordHistoryDetailResponse, error) {
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

	// 查询SSL记录历史
	firstSSLRecordHistory, err := s.logDb.SSLCertificateHistory.Query().
		Where(sslcertificatehistory.SslRecordIDEQ(firstSSLRecord.ID)).
		Where(sslcertificatehistory.IDEQ(historyID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询SSL记录历史失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询SSL记录历史失败: %w", err)
	}

	return &types.GetSSLRecordHistoryDetailResponse{
		ID:                    firstSSLRecordHistory.ID,
		CreatedAt:             firstSSLRecordHistory.CreatedAt,
		SslRecordID:           firstSSLRecordHistory.SslRecordID,
		CertDomain:            firstSSLRecordHistory.CertDomain,
		CertURL:               firstSSLRecordHistory.CertURL,
		CertStableURL:         firstSSLRecordHistory.CertStableURL,
		CertPrivateKey:        firstSSLRecordHistory.CertPrivateKey,
		CertCertificate:       firstSSLRecordHistory.CertCertificate,
		CertIssuerCertificate: firstSSLRecordHistory.CertIssuerCertificate,
		CertCsr:               firstSSLRecordHistory.CertCsr,
		NotBefore:             firstSSLRecordHistory.NotBefore,
		NotAfter:              firstSSLRecordHistory.NotAfter,
		Provider:              firstSSLRecordHistory.Provider.String(),
	}, nil
}
