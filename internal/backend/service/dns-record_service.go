package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// DNSRecordService DNS记录服务接口
type DNSRecordService interface {
	BuildResponse(ctx context.Context, item *ent.DNSRecord) (*types.GetDnsRecordResponse, error)
	GetDnsRecords(ctx context.Context, request *types.GetDnsRecordsRequest) ([]*types.GetDnsRecordResponse, error)
	GetDnsRecord(ctx context.Context, id int64) (*types.GetDnsRecordResponse, error)
	CreateDnsRecord(ctx context.Context, request *types.CreateDnsRecordRequest) error
	UpdateDnsRecord(ctx context.Context, request *types.UpdateDnsRecordRequest) error
	UpdateDnsRecordActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateDnsRecordActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteDnsRecord(ctx context.Context, id int64) error
	BatchDeleteDnsRecord(ctx context.Context, ids []int64) error
	HealthcheckDnsRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// DNSRecordServiceImpl DNS 记录服务实现
type DNSRecordServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewDNSRecordService 创建 DNS 记录服务实例
func NewDNSRecordService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) DNSRecordService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &DNSRecordServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// SetDebug 设置调试模式
func (s *DNSRecordServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *DNSRecordServiceImpl) BuildResponse(ctx context.Context, item *ent.DNSRecord) (*types.GetDnsRecordResponse, error) {
	resp := &types.GetDnsRecordResponse{
		ID:            item.ID,
		ConfigID:      item.ConfigID,
		Platform:      item.Edges.Config.Platform.String(),
		RecordID:      item.RecordID,
		SubDomain:     item.SubDomain,
		Type:          item.Type.String(),
		ConfigJSON:    item.ConfigJSON,
		IsActive:      item.IsActive,
		LastRefreshAt: item.LastRefreshAt,
	}

	// 平台配置
	if config, err := item.Edges.ConfigOrErr(); err == nil {
		resp.ConfigName = config.Name
		resp.Platform = config.Platform.String()
		// 从平台配置中获取 ConfigJSON 里的 Domain
		if len(config.ConfigJSON) > 0 {
			configJSON, err := schematype.ParsePCJODomainName(config.ConfigJSON)
			if err == nil {
				resp.DomainName = configJSON.DomainName
			}
		}
	}

	// IP 组
	if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
		resp.IPGroupID = ipGroup.ID
		resp.IPGroupName = ipGroup.Name
	}

	// CDN 记录
	if cdnRecord, err := item.Edges.CdnRecordOrErr(); err == nil {
		resp.CdnRecordID = cdnRecord.ID
		resp.CdnRecordCdnDomain = cdnRecord.CdnDomain
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDNSRecord)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetDnsRecords 获取全部DNS记录
func (s *DNSRecordServiceImpl) GetDnsRecords(ctx context.Context, request *types.GetDnsRecordsRequest) ([]*types.GetDnsRecordResponse, error) {

	// 查询全部DNS记录
	allDnsRecord, err := s.db.DNSRecord.Query().
		WhereIf(request.IsActive != nil, dnsrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}

	lists := make([]*types.GetDnsRecordResponse, 0, len(allDnsRecord))
	for _, item := range allDnsRecord {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetDnsRecord 获取单个DNS记录
func (s *DNSRecordServiceImpl) GetDnsRecord(ctx context.Context, id int64) (*types.GetDnsRecordResponse, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstDnsRecord)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateDnsRecord 创建DNS记录
func (s *DNSRecordServiceImpl) CreateDnsRecord(ctx context.Context, request *types.CreateDnsRecordRequest) error {
	// 创建DNS记录 (保持原始的链式设置逻辑)
	createDnsRecord := s.db.DNSRecord.Create().
		SetConfigID(request.ConfigID).
		SetSubDomain(request.SubDomain).
		SetType(dnsrecord.Type(request.Type)).
		SetConfigJSON([]byte(request.ConfigJSON))
	if request.IPGroupID > 0 {
		createDnsRecord.SetIPGroupID(request.IPGroupID)
	}
	if request.CdnRecordID > 0 {
		createDnsRecord.SetCdnRecordID(request.CdnRecordID)
	}

	saveDnsRecord, err := createDnsRecord.Save(ctx)
	if err != nil {
		return fmt.Errorf("创建DNS记录失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDNSRecord, saveDnsRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateDnsRecord 更新DNS记录
func (s *DNSRecordServiceImpl) UpdateDnsRecord(ctx context.Context, request *types.UpdateDnsRecordRequest) error {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}

	// 更新DNS记录 (保持原始的链式设置逻辑)
	updateDnsRecord := s.db.DNSRecord.UpdateOne(firstDnsRecord).
		SetType(dnsrecord.Type(request.Type)).
		SetConfigJSON([]byte(request.ConfigJSON))
	if request.IPGroupID > 0 {
		updateDnsRecord.SetIPGroupID(request.IPGroupID)
		updateDnsRecord.ClearCdnRecordID()
	}
	if request.CdnRecordID > 0 {
		updateDnsRecord.ClearIPGroupID()
		updateDnsRecord.SetCdnRecordID(request.CdnRecordID)
	}

	saveDnsRecord, err := updateDnsRecord.Save(ctx)
	if err != nil {
		return fmt.Errorf("更新DNS记录失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDNSRecord, saveDnsRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateDnsRecordActive 更新DNS记录启用状态
func (s *DNSRecordServiceImpl) UpdateDnsRecordActive(ctx context.Context, id int64, isActive bool) error {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}

	// 更新DNS记录状态
	err = s.db.DNSRecord.UpdateOne(firstDnsRecord).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新DNS记录状态失败: %w", err)
	}

	return nil
}

// BatchUpdateDnsRecordActive 批量更新DNS记录启用状态
func (s *DNSRecordServiceImpl) BatchUpdateDnsRecordActive(ctx context.Context, ids []int64, isActive bool) error {
	// 查询DNS记录
	allDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}
	if len(allDnsRecord) != len(ids) {
		return fmt.Errorf("查询DNS记录失败: 数量 %d 与请求数量 %d 不匹配", len(allDnsRecord), len(ids))
	}

	// 更新DNS记录状态
	err = s.db.DNSRecord.Update().
		Where(dnsrecord.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新DNS记录状态失败: %w", err)
	}

	return nil
}

// DeleteDnsRecord 删除DNS记录
func (s *DNSRecordServiceImpl) DeleteDnsRecord(ctx context.Context, id int64) error {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}
	// 删除DNS记录
	err = s.db.DNSRecord.DeleteOne(firstDnsRecord).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDNSRecord, firstDnsRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// BatchDeleteDnsRecord 批量删除 DNS 记录
func (s *DNSRecordServiceImpl) BatchDeleteDnsRecord(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询DNS记录
	allDnsRecord, err := tx.DNSRecord.Query().
		Where(dnsrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询DNS记录失败: %w", err)
	}
	if len(allDnsRecord) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询DNS记录失败: 数量 %d 与请求数量 %d 不匹配", len(allDnsRecord), len(ids))
	}

	// 检查是否存在启用的DNS记录
	for _, item := range allDnsRecord {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除DNS记录失败: 存在启用的DNS记录")
		}
	}

	// 删除DNS记录
	_, err = tx.DNSRecord.Delete().
		Where(dnsrecord.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除DNS记录失败: %w", err)
	}

	// 删除通知订阅
	for _, item := range allDnsRecord {
		err = webhook_sub_helper.DeleteTx(ctx, tx, webhooksubscription.ObjTypeDNSRecord, item.ID)
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

// HealthcheckDnsRecord 检查DNS记录健康状态
func (s *DNSRecordServiceImpl) HealthcheckDnsRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
		WithConfig().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}

	// 检查健康状态
	results := make([][]ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发 (保持原始限制)

	if firstDnsRecord.SubDomain != "" {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, firstDnsRecord)
		if err != nil {
			return nil, err
		}
		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		eg.Go(func() error {
			healthCheckResult, _ := CheckDomainHealth(egCtx, schematype.SplicingPCJODomainName(list.SubDomain, list.DomainName), 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForDomain())
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return nil, fmt.Errorf("DNS记录健康状态检查失败: %w", err)
	}

	return results, nil
}
