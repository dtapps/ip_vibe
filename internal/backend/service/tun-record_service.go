package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/tunrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

// TUNRecordService TUN记录服务接口
type TUNRecordService interface {
	BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error)
	GetTUNRecords(ctx context.Context, request *types.GetTunRecordsRequest) ([]*types.GetTunRecordResponse, error)
	GetTUNRecord(ctx context.Context, id int64) (*types.GetTunRecordResponse, error)
	CreateTUNRecord(ctx context.Context, request *types.CreateTunRecordRequest) error
	UpdateTUNRecord(ctx context.Context, request *types.UpdateTunRecordRequest) error
	UpdateTUNRecordActive(ctx context.Context, id int64, isActive bool) error
	BatchUpdateTUNRecordActive(ctx context.Context, ids []int64, isActive bool) error
	DeleteTUNRecord(ctx context.Context, id int64) error
	BatchDeleteTUNRecord(ctx context.Context, ids []int64) error
}

// TUNRecordServiceImpl DNS 记录服务实现
type TUNRecordServiceImpl struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器
	debug  bool            // 调试模式
}

// NewTUNRecordService 创建 DNS 记录服务实例
func NewTUNRecordService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) TUNRecordService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &TUNRecordServiceImpl{
		db:     db,
		logDb:  logDb,
		logger: logger,
		debug:  debug,
	}
}

// SetDebug 设置调试模式
func (s *TUNRecordServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *TUNRecordServiceImpl) BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error) {
	resp := &types.GetTunRecordResponse{
		ID:            item.ID,
		ConfigID:      item.ConfigID,
		Platform:      item.Edges.Config.Platform.String(),
		RecordID:      item.RecordID,
		SubDomain:     item.SubDomain,
		DomainPath:    item.DomainPath,
		ServeType:     item.ServeType.String(),
		ServeURI:      item.ServeURI,
		ConfigJSON:    item.ConfigJSON,
		IsActive:      item.IsActive,
		LastRefreshAt: item.LastRefreshAt,

		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		LastAPIResponse:   item.LastAPIResponse,
		LastAPIResponseAt: item.LastAPIResponseAt,
	}

	// 平台配置
	if config, err := item.Edges.ConfigOrErr(); err == nil {
		resp.ConfigName = config.Name
		resp.Platform = config.Platform.String()
		// 从平台配置中获取 ConfigJSON 里的 Domain
		if len(config.ConfigJSON) > 0 {
			resp.ConfigJSONAccountID = gjson.GetBytes(config.ConfigJSON, "account_id").String()
		}

		switch config.Platform {
		case platformconfig.PlatformCloudflare:
			// 从配置中获取 ConfigJSON 里的 TunnelName 和 ZoneName
			if len(item.ConfigJSON) > 0 {
				resp.ConfigJSONTunnelID = gjson.GetBytes(item.ConfigJSON, "tunnel_id").String()
				resp.ConfigJSONTunnelName = gjson.GetBytes(item.ConfigJSON, "tunnel_name").String()
				resp.ConfigJSONZoneID = gjson.GetBytes(item.ConfigJSON, "zone_id").String()
				resp.ConfigJSONZoneName = gjson.GetBytes(item.ConfigJSON, "zone_name").String()
			}
		}
	}

	// 查询通知订阅
	webhookIDs, _ := s.db.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeTunRecord)).
		Where(webhooksubscription.ObjIDEQ(item.ID)).
		QueryWebhook().
		IDs(ctx)

	resp.Webhooks = webhookIDs

	return resp, nil
}

// GetTUNRecords 获取全部TUN记录
func (s *TUNRecordServiceImpl) GetTUNRecords(ctx context.Context, request *types.GetTunRecordsRequest) ([]*types.GetTunRecordResponse, error) {

	// 查询全部TUN记录
	allTUNRecord, err := s.db.TUNRecord.Query().
		WhereIf(request.IsActive != nil, tunrecord.IsActiveEQ(cast.ToBool(request.IsActive))).
		WithConfig().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询TUN记录失败: %w", err)
	}

	lists := make([]*types.GetTunRecordResponse, 0, len(allTUNRecord))
	for _, item := range allTUNRecord {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	return lists, nil
}

// GetTUNRecord 获取单个TUN记录
func (s *TUNRecordServiceImpl) GetTUNRecord(ctx context.Context, id int64) (*types.GetTunRecordResponse, error) {

	// 查询TUN记录
	firstTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IDEQ(id)).
		WithConfig().
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询TUN记录失败: 记录不存在或 %w", err)
		}
		return nil, fmt.Errorf("查询TUN记录失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstTUNRecord)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// CreateTUNRecord 创建TUN记录
func (s *TUNRecordServiceImpl) CreateTUNRecord(ctx context.Context, request *types.CreateTunRecordRequest) error {

	// 转小写
	request.ServeType = strings.ToLower(request.ServeType)

	// 创建TUN记录
	saveTUNRecord, err := s.db.TUNRecord.Create().
		SetConfigID(request.ConfigID).
		SetSubDomain(request.SubDomain).
		SetDomainPath(request.DomainPath).
		SetServeType(tunrecord.ServeType(request.ServeType)).
		SetServeURI(request.ServeURI).
		SetConfigJSON([]byte(request.ConfigJSON)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("创建TUN记录失败: %w", err)
	}

	// 创建通知订阅
	err = webhook_sub_helper.Create(ctx, request.WebhookIDs, webhooksubscription.ObjTypeTunRecord, saveTUNRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("创建通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateTUNRecord 更新TUN记录
func (s *TUNRecordServiceImpl) UpdateTUNRecord(ctx context.Context, request *types.UpdateTunRecordRequest) error {

	// 转小写
	request.ServeType = strings.ToLower(request.ServeType)

	// 查询TUN记录
	firstTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询TUN记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询TUN记录失败: %w", err)
	}

	// 更新TUN记录
	saveTUNRecord, err := s.db.TUNRecord.UpdateOne(firstTUNRecord).
		SetSubDomain(request.SubDomain).
		SetDomainPath(request.DomainPath).
		SetServeType(tunrecord.ServeType(request.ServeType)).
		SetServeURI(request.ServeURI).
		SetConfigJSON([]byte(request.ConfigJSON)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("更新TUN记录失败: %w", err)
	}

	// 更新通知订阅
	err = webhook_sub_helper.Update(ctx, request.WebhookIDs, webhooksubscription.ObjTypeTunRecord, saveTUNRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("更新通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// UpdateTUNRecordActive 更新TUN记录启用状态
func (s *TUNRecordServiceImpl) UpdateTUNRecordActive(ctx context.Context, id int64, isActive bool) error {

	// 查询TUN记录
	firstTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询TUN记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询TUN记录失败: %w", err)
	}

	// 更新TUN记录状态
	err = s.db.TUNRecord.UpdateOne(firstTUNRecord).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新TUN记录状态失败: %w", err)
	}

	return nil
}

// BatchUpdateTUNRecordActive 批量更新TUN记录启用状态
func (s *TUNRecordServiceImpl) BatchUpdateTUNRecordActive(ctx context.Context, ids []int64, isActive bool) error {

	// 查询TUN记录
	allTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询TUN记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询TUN记录失败: %w", err)
	}
	if len(allTUNRecord) != len(ids) {
		return fmt.Errorf("查询TUN记录失败: 数量 %d 与请求数量 %d 不匹配", len(allTUNRecord), len(ids))
	}

	// 更新TUN记录状态
	err = s.db.TUNRecord.Update().
		Where(tunrecord.IDIn(ids...)).
		SetIsActive(isActive).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新TUN记录状态失败: %w", err)
	}

	return nil
}

// DeleteTUNRecord 删除TUN记录
func (s *TUNRecordServiceImpl) DeleteTUNRecord(ctx context.Context, id int64) error {

	// 查询TUN记录
	firstTUNRecord, err := s.db.TUNRecord.Query().
		Where(tunrecord.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询TUN记录失败: 记录不存在或 %w", err)
		}
		return fmt.Errorf("查询TUN记录失败: %w", err)
	}
	// 删除TUN记录
	err = s.db.TUNRecord.DeleteOne(firstTUNRecord).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除TUN记录失败: %w", err)
	}

	// 删除通知订阅
	err = webhook_sub_helper.Delete(ctx, webhooksubscription.ObjTypeTunRecord, firstTUNRecord.ID)
	if err != nil {
		if s.debug {
			fmt.Printf("删除通知订阅失败: %s\n", err.Error())
		}
	}

	return nil
}

// BatchDeleteTUNRecord 批量删除 DNS 记录
func (s *TUNRecordServiceImpl) BatchDeleteTUNRecord(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询TUN记录
	allTUNRecord, err := tx.TUNRecord.Query().
		Where(tunrecord.IDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询TUN记录失败: %w", err)
	}
	if len(allTUNRecord) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询TUN记录失败: 数量 %d 与请求数量 %d 不匹配", len(allTUNRecord), len(ids))
	}

	// 检查是否存在启用的TUN记录
	for _, item := range allTUNRecord {
		if item.IsActive {
			tx.Rollback()
			return fmt.Errorf("删除TUN记录失败: 存在启用的TUN记录")
		}
	}

	// 删除TUN记录
	_, err = tx.TUNRecord.Delete().
		Where(tunrecord.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除TUN记录失败: %w", err)
	}

	// 删除通知订阅
	for _, item := range allTUNRecord {
		err = webhook_sub_helper.DeleteTx(ctx, tx, webhooksubscription.ObjTypeTunRecord, item.ID)
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
