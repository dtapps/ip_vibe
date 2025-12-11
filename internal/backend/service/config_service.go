package service

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"
)

// PlatformConfigService 配置服务接口
type PlatformConfigService interface {
	GetConfigs(ctx context.Context) ([]types.GetConfigResponse, error)
	GetConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error)
	CreateConfig(ctx context.Context, request *types.CreateConfigRequest) (*types.GetConfigRequest, error)
	UpdateConfig(ctx context.Context, request *types.UpdateConfigRequest) (*types.GetConfigRequest, error)
	DeleteConfig(ctx context.Context, id int64) (*types.GetConfigRequest, error)
}

// PlatformConfigServiceImpl 配置服务实现
type PlatformConfigServiceImpl struct {
	db        *ent.Client
	debug     bool
	logHelper *operation_logs_helper.OperationLogHelper
}

// NewPlatformConfigService 创建配置服务实例
func NewPlatformConfigService(db *ent.Client, debug bool, logHelper *operation_logs_helper.OperationLogHelper) PlatformConfigService {
	return &PlatformConfigServiceImpl{db: db, debug: debug, logHelper: logHelper}
}

// GetConfigs 获取全部配置
func (s *PlatformConfigServiceImpl) GetConfigs(ctx context.Context) ([]types.GetConfigResponse, error) {
	configs := make([]types.GetConfigResponse, 0)
	// 查询全部配置
	allConfig, err := s.db.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	// ⚠️ 原始映射逻辑：直接在循环中构建 DTO
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

// GetConfig 获取单个配置
func (s *PlatformConfigServiceImpl) GetConfig(ctx context.Context, id int64) (*types.GetConfigResponse, error) {
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
	// ⚠️ 原始映射逻辑：直接构建 DTO
	response := types.GetConfigResponse{
		ID:          firstConfig.ID,
		Capability:  firstConfig.Capability.String(),
		Name:        firstConfig.Name,
		Description: firstConfig.Description,
		Platform:    firstConfig.Platform.String(),
		ConfigJSON:  firstConfig.ConfigJSON,
	}
	return &response, nil
}

// CreateConfig 创建配置
func (s *PlatformConfigServiceImpl) CreateConfig(ctx context.Context, request *types.CreateConfigRequest) (*types.GetConfigRequest, error) {
	// 创建配置
	saveConfig, err := s.db.PlatformConfig.Create().
		SetName(request.Name).
		SetDescription(request.Description).
		SetPlatform(platformconfig.Platform(request.Platform)).
		SetConfigJSON([]byte(request.ConfigJSON)).
		SetCapability(platformconfig.Capability(request.Capability)).
		Save(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("创建配置: %s\n", err.Error())
		}
		return nil, fmt.Errorf("创建配置失败: %w", err)
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(ctx, saveConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "创建配置")

	return &types.GetConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// UpdateConfig 更新配置
func (s *PlatformConfigServiceImpl) UpdateConfig(ctx context.Context, request *types.UpdateConfigRequest) (*types.GetConfigRequest, error) {
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
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("更新配置: %s\n", err.Error())
		}
		return nil, fmt.Errorf("更新配置失败: %w", err)
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(ctx, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "更新配置")

	return &types.GetConfigRequest{
		ID: saveConfig.ID,
	}, nil
}

// DeleteConfig 删除配置
func (s *PlatformConfigServiceImpl) DeleteConfig(ctx context.Context, id int64) (*types.GetConfigRequest, error) {
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

	// 删除配置
	err = s.db.PlatformConfig.DeleteOne(firstConfig).Exec(ctx)
	if err != nil {
		if s.debug { // 修正：使用 s.debug 字段代替全局变量
			fmt.Printf("删除配置: %s\n", err.Error())
		}
		return nil, fmt.Errorf("删除配置失败: %w", err)
	}

	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(ctx, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "删除配置")

	return &types.GetConfigRequest{
		ID: firstConfig.ID,
	}, nil
}
