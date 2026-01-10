package service

import (
	"context"
	"encoding/json"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// CDNRecordService CDN 记录服务接口
type CDNRecordService interface {
	BuildResponse(ctx context.Context, item *ent.CDNRecord) (*types.GetCdnRecordResponse, error)
	GetCdnRecords(ctx context.Context, request *types.GetCdnRecordsRequest) ([]*types.GetCdnRecordResponse, error)
	GetCdnRecord(ctx context.Context, id int64) (*types.GetCdnRecordResponse, error)
	CreateCdnRecord(ctx context.Context, request *types.CreateCdnRecordRequest) (*types.GetCdnRecordRequest, error)
	UpdateCdnRecord(ctx context.Context, request *types.UpdateCdnRecordRequest) (*types.GetCdnRecordRequest, error)
	UpdateCdnRecordActive(ctx context.Context, request *types.UpdateCdnRecordActiveRequest) (*types.GetCdnRecordRequest, error)
	DeleteCdnRecord(ctx context.Context, id int64) (*types.GetCdnRecordRequest, error)
	HealthcheckCdnRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error)
}

// CDNRecordServiceImpl CDN 记录服务实现
type CDNRecordServiceImpl struct {
	db    *ent.Client // 数据库客户端
	debug bool        // 调试模式
}

// NewCDNRecordService 创建 CDN 记录服务实例
func NewCDNRecordService(db *ent.Client, debug bool) CDNRecordService {
	return &CDNRecordServiceImpl{
		db:    db,
		debug: debug,
	}
}

// BuildResponse 将实体转换为响应
func (s *CDNRecordServiceImpl) BuildResponse(ctx context.Context, item *ent.CDNRecord) (*types.GetCdnRecordResponse, error) {
	resp := &types.GetCdnRecordResponse{
		ID:            item.ID,
		ConfigID:      item.ConfigID,
		RecordID:      item.RecordID,
		SubDomain:     item.SubDomain,
		Type:          item.Type.String(),
		ConfigJSON:    item.ConfigJSON,
		CdnDomain:     item.CdnDomain,
		IsActive:      item.IsActive,
		LastRefreshAt: item.LastRefreshAt,
	}

	// 平台配置
	if item.Edges.Config != nil {
		resp.ConfigName = item.Edges.Config.Name
		resp.Platform = item.Edges.Config.Platform.String()
		// 从平台配置中获取 ConfigJSON 里的 Domain
		if item.Edges.Config.ConfigJSON != nil {
			configJSON, err := schematype.ParsePCJODomainName(item.Edges.Config.ConfigJSON)
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

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeCdnRecord)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetCdnRecords 获取全部 CDN 记录
func (s *CDNRecordServiceImpl) GetCdnRecords(ctx context.Context, request *types.GetCdnRecordsRequest) ([]*types.GetCdnRecordResponse, error) {
	lists := make([]*types.GetCdnRecordResponse, 0)
	// 查询全部CDN记录
	allCdnRecord, err := s.db.CDNRecord.Query().
		WhereIf(request.IsActive != nil, cdnrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		WithIPGroup().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}

	for _, item := range allCdnRecord {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetCdnRecord 获取单个 CDN 记录
func (s *CDNRecordServiceImpl) GetCdnRecord(ctx context.Context, id int64) (*types.GetCdnRecordResponse, error) {
	// 查询CDN记录
	firstCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(id)).
		WithConfig().
		WithIPGroup().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询CDN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询CDN记录失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstCdnRecord)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateCdnRecord 创建 CDN 记录
func (s *CDNRecordServiceImpl) CreateCdnRecord(ctx context.Context, request *types.CreateCdnRecordRequest) (*types.GetCdnRecordRequest, error) {
	// 创建CDN记录
	saveCdnRecord, err := s.db.CDNRecord.Create().
		SetConfigID(request.ConfigID).
		SetSubDomain(request.SubDomain).
		SetType(cdnrecord.Type(request.Type)).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetIPGroupID(request.IPGroupID).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("创建CDN记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建CDN记录失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeCdnRecord, saveCdnRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	}, nil
}

// UpdateCdnRecord 更新 CDN 记录
func (s *CDNRecordServiceImpl) UpdateCdnRecord(ctx context.Context, request *types.UpdateCdnRecordRequest) (*types.GetCdnRecordRequest, error) {
	// 查询CDN记录
	firstCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询CDN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询CDN记录失败: %w", err)
	}

	// 更新CDN记录
	saveCdnRecord, err := s.db.CDNRecord.UpdateOne(firstCdnRecord).
		SetType(cdnrecord.Type(request.Type)).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetIPGroupID(request.IPGroupID).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新CDN记录: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新CDN记录失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeCdnRecord, saveCdnRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	}, nil
}

// UpdateCdnRecordActive 更新 CDN 记录启用状态
func (s *CDNRecordServiceImpl) UpdateCdnRecordActive(ctx context.Context, request *types.UpdateCdnRecordActiveRequest) (*types.GetCdnRecordRequest, error) {
	// 查询CDN记录
	firstCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询CDN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询CDN记录失败: %w", err)
	}

	// 更新CDN记录状态
	saveCdnRecord, err := s.db.CDNRecord.UpdateOne(firstCdnRecord).
		SetIsActive(request.IsActive).
		Save(ctx)
	if err != nil {
		if s.debug {
			fmt.Printf("更新CDN记录状态: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新CDN记录状态失败: %w", err)
	}

	return &types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	}, nil
}

// DeleteCdnRecord 删除 CDN 记录
func (s *CDNRecordServiceImpl) DeleteCdnRecord(ctx context.Context, id int64) (*types.GetCdnRecordRequest, error) {
	// 查询CDN记录
	firstCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询CDN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询CDN记录失败: %w", err)
	}

	// 删除CDN记录
	err = s.db.CDNRecord.DeleteOne(firstCdnRecord).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("删除CDN记录失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeCdnRecord, firstCdnRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return &types.GetCdnRecordRequest{
		ID: firstCdnRecord.ID,
	}, nil
}

// HealthcheckCdnRecord 检查 CDN 记录健康状态
func (s *CDNRecordServiceImpl) HealthcheckCdnRecord(ctx context.Context, id int64) ([][]ToPairsStruct, error) {
	// 查询CDN记录
	firstCdnRecord, err := s.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(id)).
		WithConfig().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询CDN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询CDN记录失败: %w", err)
	}

	// 解析CDN记录配置
	var cdnRecordConfigJson schematype.CDNRecordConfigJSON
	err = json.Unmarshal([]byte(firstCdnRecord.ConfigJSON), &cdnRecordConfigJson)
	if err != nil {
		return nil, fmt.Errorf("解析CDN记录配置失败: %w", err)
	}
	switch cdnRecordConfigJson.OriginProtocol {
	case schematype.CDNRecordOriginProtocolHTTP:
		cdnRecordConfigJson.HTTPSPort = 0
	case schematype.CDNRecordOriginProtocolHTTPS:
		cdnRecordConfigJson.HTTPPort = 0
	}

	results := make([][]ToPairsStruct, 0)
	// 并发检查 (使用 context.Context)
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

	if firstCdnRecord.SubDomain != "" {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, firstCdnRecord)
		if err != nil {
			return nil, err
		}
		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		eg.Go(func() error {
			healthCheckResult, _ := CheckDomainHealth(egCtx, fmt.Sprintf("%s.%s", list.SubDomain, list.DomainName), 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForDomain())
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return nil, fmt.Errorf("CDN记录健康状态检查失败: %w", err)
	}

	return results, nil
}
