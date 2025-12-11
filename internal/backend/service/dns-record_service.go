package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// DNSRecordService DNS记录服务接口
type DNSRecordService interface {
	GetDnsRecords(ctx context.Context, request *types.GetDnsRecordsRequest) ([]types.GetDnsRecordResponse, error)
	GetDnsRecord(ctx context.Context, id int64) (*types.GetDnsRecordResponse, error)
	CreateDnsRecord(ctx context.Context, request *types.CreateDnsRecordRequest) (*types.GetDnsRecordRequest, error)
	UpdateDnsRecord(ctx context.Context, request *types.UpdateDnsRecordRequest) (*types.GetDnsRecordRequest, error)
	UpdateDnsRecordActive(ctx context.Context, request *types.UpdateDnsRecordActiveRequest) (*types.GetDnsRecordRequest, error)
	DeleteDnsRecord(ctx context.Context, id int64) (*types.GetDnsRecordRequest, error)
	HealthcheckDnsRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// DNSRecordServiceImpl DNS 记录服务实现
type DNSRecordServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewDNSRecordService 创建 DNS 记录服务实例
func NewDNSRecordService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) DNSRecordService {
	return &DNSRecordServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// GetDnsRecords 获取全部DNS记录
func (s *DNSRecordServiceImpl) GetDnsRecords(ctx context.Context, request *types.GetDnsRecordsRequest) ([]types.GetDnsRecordResponse, error) {
	records := make([]types.GetDnsRecordResponse, 0)
	// 查询全部DNS记录
	allDnsRecord, err := s.db.DNSRecord.Query().
		WhereIf(request.IsActive != nil, dnsrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		Order(ent.Desc(dnsrecord.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}
	for _, item := range allDnsRecord {
		list := types.GetDnsRecordResponse{
			ID:            item.ID,
			ConfigID:      item.ConfigID,
			Platform:      item.Edges.Config.Platform.String(),
			RecordID:      item.RecordID,
			Domain:        item.Domain,
			Type:          item.Type.String(),
			IsActive:      item.IsActive,
			LastRefreshAt: item.LastRefreshAt,
		}
		if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
			list.IPGroupID = ipGroup.ID
			list.IPGroupName = ipGroup.Name
		}
		if cdnRecord, err := item.Edges.CdnRecordOrErr(); err == nil {
			list.CdnRecordID = cdnRecord.ID
			list.CdnRecordCdnDomain = cdnRecord.CdnDomain
		}
		// 查询通知订阅 (保持原始查询逻辑)
		allWebhookSub, _ := s.db.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDNSRecord)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			All(ctx)
		for _, sub := range allWebhookSub {
			list.Webhooks = append(list.Webhooks, sub.ID)
		}
		records = append(records, list)
	}
	return records, nil
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
	response := types.GetDnsRecordResponse{
		ID:            firstDnsRecord.ID,
		ConfigID:      firstDnsRecord.ConfigID,
		Platform:      firstDnsRecord.Edges.Config.Platform.String(),
		RecordID:      firstDnsRecord.RecordID,
		Domain:        firstDnsRecord.Domain,
		Type:          firstDnsRecord.Type.String(),
		IsActive:      firstDnsRecord.IsActive,
		LastRefreshAt: firstDnsRecord.LastRefreshAt,
	}
	if ipGroup, err := firstDnsRecord.Edges.IPGroupOrErr(); err == nil {
		response.IPGroupID = ipGroup.ID
		response.IPGroupName = ipGroup.Name
	}
	if cdnRecord, err := firstDnsRecord.Edges.CdnRecordOrErr(); err == nil {
		response.CdnRecordID = cdnRecord.ID
		response.CdnRecordCdnDomain = cdnRecord.CdnDomain
	}
	// 查询通知订阅 (保持原始查询逻辑)
	allWebhookSub, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDNSRecord)).
		Where(webhooksubscription.ObjIDEQ(id)).
		All(ctx)
	for _, item := range allWebhookSub {
		response.Webhooks = append(response.Webhooks, item.ID)
	}
	return &response, nil
}

// CreateDnsRecord 创建DNS记录
func (s *DNSRecordServiceImpl) CreateDnsRecord(ctx context.Context, request *types.CreateDnsRecordRequest) (*types.GetDnsRecordRequest, error) {
	// 创建DNS记录 (保持原始的链式设置逻辑)
	createDnsRecord := s.db.DNSRecord.Create().
		SetConfigID(request.ConfigID).
		SetDomain(request.Domain).
		SetType(dnsrecord.Type(request.Type))
	if request.IPGroupID > 0 {
		createDnsRecord.SetIPGroupID(request.IPGroupID)
	}
	if request.CdnRecordID > 0 {
		createDnsRecord.SetCdnRecordID(request.CdnRecordID)
	}

	saveDnsRecord, err := createDnsRecord.Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("创建DNS记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建DNS记录失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataCreate(ctx, saveDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "创建DNS记录")

	// 创建通知订阅
	_ = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDNSRecord, saveDnsRecord.ID)

	return &types.GetDnsRecordRequest{
		ID: saveDnsRecord.ID,
	}, nil
}

// UpdateDnsRecord 更新DNS记录
func (s *DNSRecordServiceImpl) UpdateDnsRecord(ctx context.Context, request *types.UpdateDnsRecordRequest) (*types.GetDnsRecordRequest, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}

	// 更新DNS记录 (保持原始的链式设置逻辑)
	updateDnsRecord := s.db.DNSRecord.UpdateOne(firstDnsRecord)
	if request.IPGroupID > 0 {
		updateDnsRecord.SetIPGroupID(request.IPGroupID)
	}
	if request.CdnRecordID > 0 {
		updateDnsRecord.SetCdnRecordID(request.CdnRecordID)
	}

	saveDnsRecord, err := updateDnsRecord.Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("更新DNS记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新DNS记录失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, saveDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "更新DNS记录")

	// 更新通知订阅
	_ = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeDNSRecord, saveDnsRecord.ID)

	return &types.GetDnsRecordRequest{
		ID: saveDnsRecord.ID,
	}, nil
}

// UpdateDnsRecordActive 更新DNS记录启用状态
func (s *DNSRecordServiceImpl) UpdateDnsRecordActive(ctx context.Context, request *types.UpdateDnsRecordActiveRequest) (*types.GetDnsRecordRequest, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}
	// 更新DNS记录状态
	saveRecord, err := s.db.DNSRecord.UpdateOne(firstDnsRecord).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("更新DNS记录状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新DNS记录状态失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataUpdate(ctx, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "更新DNS记录状态")

	return &types.GetDnsRecordRequest{
		ID: saveRecord.ID,
	}, nil
}

// DeleteDnsRecord 删除DNS记录
func (s *DNSRecordServiceImpl) DeleteDnsRecord(ctx context.Context, id int64) (*types.GetDnsRecordRequest, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询DNS记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询DNS记录失败: %w", err)
	}
	// 删除DNS记录
	err = s.db.DNSRecord.DeleteOne(firstDnsRecord).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("删除DNS记录失败: %w", err)
	}
	// 记录操作日志
	_ = s.logHelper.RecordDataDelete(ctx, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "删除DNS记录")

	// 删除通知订阅
	_ = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeDNSRecord, firstDnsRecord.ID)

	return &types.GetDnsRecordRequest{
		ID: firstDnsRecord.ID,
	}, nil
}

// HealthcheckDnsRecord 检查DNS记录健康状态
func (s *DNSRecordServiceImpl) HealthcheckDnsRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询DNS记录
	firstDnsRecord, err := s.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(id)).
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

	if firstDnsRecord.Domain != "" {
		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		eg.Go(func() error {
			healthCheckResult, _ := CheckDomainHealth(egCtx, firstDnsRecord.Domain, 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForDomain())
			return nil
		})
	}

	// 以下注释掉的逻辑保持不变 (将 Handler 中的注释原封不动迁移)
	/*
		// 查询Dns的IP组
		// firstDnsIpGroup, err := firstDnsRecord.QueryIPGroup().
		// 	WithDevices().
		// 	First(egCtx)
		// if err == nil {
		// 	for _, item := range firstDnsIpGroup.Edges.Devices {
		// 		item := item // 创建副本以避免闭包问题
		// 		if item.IPV4 != "" && firstDnsRecord.Type == dnsrecord.TypeA {
		// 			// 检测 IP 是否可通过 HTTP/HTTPS 访问
		// 			eg.Go(func() error {
		// 				healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV4, 80, 443)
		// 				results = append(results, healthCheckResult.ToPairsForIP())
		// 				return nil
		// 			})
		// 		}
		// 		if item.IPV6 != "" && firstDnsRecord.Type == dnsrecord.TypeAAAA {
		// 			// 检测 IP 是否可通过 HTTP/HTTPS 访问
		// 			eg.Go(func() error {
		// 				healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV6, 80, 443)
		// 				results = append(results, healthCheckResult.ToPairsForIP())
		// 				return nil
		// 			})
		// 		}
		// 	}
		// }

		// 查询CDN记录
		// firstCdnRecord, err := firstDnsRecord.QueryCdnRecord().
		// 	First(egCtx)
		// if err == nil {
		// 	if firstCdnRecord.CdnDomain != "" {
		// 		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		// 		eg.Go(func() error {
		// 			healthCheckResult, _ := CheckDomainHealth(egCtx, firstCdnRecord.CdnDomain, 80, 443, []string{})
		// 			results = append(results, healthCheckResult.ToPairsForDomain())
		// 			return nil
		// 		})
		// 	}
		// 	// 查询CDN的IP组
		// 	firstCdnIpGroup, err := firstCdnRecord.QueryIPGroup().
		// 		WithDevices().
		// 		First(egCtx)
		// 	if err == nil {
		// 		switch firstCdnRecord.OriginProtocol {
		// 		case cdnrecord.OriginProtocolHTTP:
		// 			firstCdnRecord.HTTPSPort = 0
		// 		case cdnrecord.OriginProtocolHTTPS:
		// 			firstCdnRecord.HTTPPort = 0
		// 		}
		// 		for _, item := range firstCdnIpGroup.Edges.Devices {
		// 			item := item // 创建副本以避免闭包问题
		// 			if item.IPV4 != "" && firstCdnRecord.Type == cdnrecord.TypeA {
		// 				// 检测 IP 是否可通过 HTTP/HTTPS 访问
		// 				eg.Go(func() error {
		// 					healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV4, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
		// 					results = append(results, healthCheckResult.ToPairsForIP())
		// 					return nil
		// 				})
		// 			}
		// 			if item.IPV6 != "" && firstCdnRecord.Type == cdnrecord.TypeAAAA {
		// 				// 检测 IP 是否可通过 HTTP/HTTPS 访问
		// 				eg.Go(func() error {
		// 					healthCheckResult, _ := CheckIPHealth(egCtx, item.IPV6, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
		// 					results = append(results, healthCheckResult.ToPairsForIP())
		// 					return nil
		// 				})
		// 			}
		// 		}
		// 	}
		// }
	*/

	if err = eg.Wait(); err != nil {
		return nil, fmt.Errorf("DNS记录健康状态检查失败: %w", err)
	}

	return results, nil
}
