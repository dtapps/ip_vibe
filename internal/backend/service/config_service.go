package service

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
)

// ConfigService 配置服务接口
type ConfigService interface {
	GetConfigs(ctx context.Context) ([]types.GetConfigResponse, error)
	GetConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error)
	CreateConfig(ctx context.Context, request *types.CreateConfigRequest) (*types.GetConfigRequest, error)
	UpdateConfig(ctx context.Context, request *types.UpdateConfigRequest) (*types.GetConfigRequest, error)
	DeleteConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error)
}

// ConfigServiceImpl 配置服务实现
type ConfigServiceImpl struct {
	db *ent.Client
}

// NewConfigService 创建配置服务实例
func NewConfigService(db *ent.Client) ConfigService {
	return &ConfigServiceImpl{db: db}
}

// GetConfigs 获取全部配置
func (s *ConfigServiceImpl) GetConfigs(ctx context.Context) ([]types.GetConfigResponse, error) {
	configs := make([]types.GetConfigResponse, 0)

	// 查询全部配置
	allConfig, err := s.db.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range allConfig {
		configs = append(configs, types.GetConfigResponse{
			ID:          item.ID,
			Capability:  item.Capability.String(),
			Name:        item.Name,
			Description: item.Description,
			Platform:    item.Platform.String(),
			ConfigJSON:  item.ConfigJSON,
		})
	}

	return configs, nil
}

// GetConfig 获取配置
func (s *ConfigServiceImpl) GetConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error) {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	response := &types.GetConfigResponse{
		ID:          firstConfig.ID,
		Capability:  firstConfig.Capability.String(),
		Name:        firstConfig.Name,
		Description: firstConfig.Description,
		Platform:    firstConfig.Platform.String(),
		ConfigJSON:  firstConfig.ConfigJSON,
	}

	return response, nil
}

// CreateConfig 创建配置
func (s *ConfigServiceImpl) CreateConfig(ctx context.Context, request *types.CreateConfigRequest) (*types.GetConfigRequest, error) {
	// 创建配置
	saveConfig, err := s.db.PlatformConfig.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetPlatform(platformconfig.Platform(request.Platform)).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetCapability(platformconfig.Capability(request.Capability)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(ctx, saveConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "创建配置")

	return &types.GetConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// UpdateConfig 更新配置
func (s *ConfigServiceImpl) UpdateConfig(ctx context.Context, request *types.UpdateConfigRequest) (*types.GetConfigRequest, error) {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(request.ID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 更新配置
	saveConfig, err := s.db.PlatformConfig.UpdateOne(firstConfig).
		SetName(request.Name).
		SetDescription(request.Description).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetCapability(platformconfig.Capability(request.Capability)).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "更新配置")

	return &types.GetConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// DeleteConfig 删除配置
func (s *ConfigServiceImpl) DeleteConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error) {
	// 查询配置
	firstConfig, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	// 删除配置
	err = s.db.PlatformConfig.DeleteOne(firstConfig).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(ctx, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "删除配置")

	return &types.GetConfigResponse{
		ID: firstConfig.ID,
	}, nil
}
