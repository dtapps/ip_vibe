package service

import (
	"context"
	"encoding/json"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/types"
)

// PlatformConfigService 配置服务接口
type PlatformConfigService interface {
	BuildResponse(ctx context.Context, item *ent.PlatformConfig) (*types.GetPlatformConfigResponse, error)
	GetConfigs(ctx context.Context) ([]*types.GetPlatformConfigResponse, error)
	GetConfig(ctx context.Context, id int64) (*types.GetPlatformConfigResponse, error)
	CreateConfig(ctx context.Context, request *types.CreatePlatformConfigRequest) (*types.GetPlatformConfigRequest, error)
	UpdateConfig(ctx context.Context, request *types.UpdatePlatformConfigRequest) (*types.GetPlatformConfigRequest, error)
	DeleteConfig(ctx context.Context, id int64) error
	DeleteConfigs(ctx context.Context, ids []int64) error
}

// PlatformConfigServiceImpl 配置服务实现
type PlatformConfigServiceImpl struct {
	db    *ent.Client // 数据库客户端
	debug bool        // 调试模式
}

// NewPlatformConfigService 创建配置服务实例
func NewPlatformConfigService(db *ent.Client, debug bool) PlatformConfigService {
	return &PlatformConfigServiceImpl{
		db:    db,
		debug: debug,
	}
}

// BuildResponse 将实体转换为响应
func (s *PlatformConfigServiceImpl) BuildResponse(ctx context.Context, item *ent.PlatformConfig) (*types.GetPlatformConfigResponse, error) {
	var configJson struct {
		DomainName string `json:"domain_name,omitempty"` // 域名
		Email      string `json:"email,omitempty"`       // 邮箱
	}
	_ = json.Unmarshal(item.ConfigJSON, &configJson)
	resp := &types.GetPlatformConfigResponse{
		ID:                   item.ID,
		Capability:           item.Capability.String(),
		Name:                 item.Name,
		Description:          item.Description,
		Platform:             item.Platform.String(),
		ConfigJSON:           item.ConfigJSON,
		ConfigJSONDomainName: configJson.DomainName,
		ConfigJSONEmail:      configJson.Email,
	}
	return resp, nil
}

// GetConfigs 获取全部配置
func (s *PlatformConfigServiceImpl) GetConfigs(ctx context.Context) ([]*types.GetPlatformConfigResponse, error) {
	lists := make([]*types.GetPlatformConfigResponse, 0)
	// 查询全部配置
	allConfig, err := s.db.PlatformConfig.Query().
		OrderByCreatedAtDesc().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	// 原始映射逻辑：直接在循环中构建 DTO
	for _, item := range allConfig {
		// 将实体转换为响应
		list, err := s.BuildResponse(ctx, item)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

// GetConfig 获取单个配置
func (s *PlatformConfigServiceImpl) GetConfig(ctx context.Context, id int64) (*types.GetPlatformConfigResponse, error) {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询配置失败: 配置不存在")
		}
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	// 将实体转换为响应
	list, err := s.BuildResponse(ctx, firstConfig)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateConfig 创建配置
func (s *PlatformConfigServiceImpl) CreateConfig(ctx context.Context, request *types.CreatePlatformConfigRequest) (*types.GetPlatformConfigRequest, error) {
	// 创建配置
	saveConfig, err := s.db.PlatformConfig.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetPlatform(platformconfig.Platform(request.Platform)).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetCapability(platformconfig.Capability(request.Capability)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建配置失败: %w", err)
	}

	return &types.GetPlatformConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// UpdateConfig 更新配置
func (s *PlatformConfigServiceImpl) UpdateConfig(ctx context.Context, request *types.UpdatePlatformConfigRequest) (*types.GetPlatformConfigRequest, error) {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("查询配置失败: 配置不存在")
		}
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	// 更新配置
	saveConfig, err := s.db.PlatformConfig.UpdateOne(firstConfig).
		SetName(request.Name).
		SetDescription(request.Description).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetCapability(platformconfig.Capability(request.Capability)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("更新配置失败: %w", err)
	}

	return &types.GetPlatformConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// DeleteConfig 删除配置
func (s *PlatformConfigServiceImpl) DeleteConfig(ctx context.Context, id int64) error {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("查询配置失败: 配置不存在")
		}
		return fmt.Errorf("查询配置失败: %w", err)
	}

	// 删除配置
	err = s.db.PlatformConfig.DeleteOne(firstConfig).Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}

	return nil
}

// DeleteConfigs 批量删除配置
func (s *PlatformConfigServiceImpl) DeleteConfigs(ctx context.Context, ids []int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询配置
	allConfig, err := tx.PlatformConfig.Query().
		Where(platformconfig.IDIn(ids...)).
		All(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("查询配置失败: %w", err)
	}
	if len(allConfig) != len(ids) {
		tx.Rollback()
		return fmt.Errorf("查询配置失败: 数量 %d 与请求数量 %d 不匹配", len(allConfig), len(ids))
	}

	// 删除配置
	_, err = tx.PlatformConfig.Delete().
		Where(platformconfig.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除配置失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
