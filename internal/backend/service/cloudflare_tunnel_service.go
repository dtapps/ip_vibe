package service

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/syncs"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflaretunnel"
	"vibe_ip/internal/ent/cloudflaretunnelingress"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gomd5"
)

// CloudflareTunnelService Cloudflare隧道服务接口
type CloudflareTunnelService interface {
	GetCloudflareTunnelIngress(ctx context.Context, tunnel_id int64) ([]*types.GetCloudflareTunnelIngressDetailResponse, error)

	GetCloudflareTunnelIngressDetail(ctx context.Context, tunnel_id int64, ingress_id int64) (*types.GetCloudflareTunnelIngressDetailResponse, error)
	CreateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, request *types.CreateCloudflareTunnelIngressRequest) error
	UpdateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64, request *types.UpdateCloudflareTunnelIngressRequest) error
	DeleteCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64) error
}

// CloudflareTunnelServiceImpl Cloudflare隧道服务实现
type CloudflareTunnelServiceImpl struct {
	db                   *ent.Client                   // 数据库客户端
	logDb                *ent_log.Client               // 日志数据库客户端
	logger               *slog.Logger                  // 日志记录器
	httpTransport        *http_log.LoggingRoundTripper // HTTP日志记录器
	cloudflareTunnelSync *syncs.CloudflareTunnelSync   // CloudFlare Tunnel同步
	debug                bool                          // 调试模式
}

// NewCloudflareTunnelService 创建Cloudflare隧道服务实例
func NewCloudflareTunnelService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) CloudflareTunnelService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	cloudflareTunnelSync := syncs.NewCloudflareTunnelSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &CloudflareTunnelServiceImpl{
		db:                   db,
		logDb:                logDb,
		logger:               logger,
		httpTransport:        httpTransport,
		cloudflareTunnelSync: cloudflareTunnelSync,
		debug:                debug,
	}
}

// SetDebug 设置调试模式
func (s *CloudflareTunnelServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// GetCloudflareTunnelIngress 获取Cloudflare隧道全部入站规则
func (s *CloudflareTunnelServiceImpl) GetCloudflareTunnelIngress(ctx context.Context, tunnel_id int64) ([]*types.GetCloudflareTunnelIngressDetailResponse, error) {
	// 查询Cloudflare隧道
	firstCloudflareTunnel, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.IDEQ(tunnel_id)).
		WithIngresses().
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询Cloudflare隧道失败: %w", err)
	}

	ingresses := make([]*types.GetCloudflareTunnelIngressDetailResponse, 0, len(firstCloudflareTunnel.Edges.Ingresses))
	for _, ingress := range firstCloudflareTunnel.Edges.Ingresses {
		ingresses = append(ingresses, &types.GetCloudflareTunnelIngressDetailResponse{
			ID:            ingress.ID,
			Hostname:      ingress.Hostname,
			Service:       ingress.Service,
			Path:          ingress.Path,
			OriginRequest: ingress.OriginRequest,
		})
	}

	return ingresses, nil
}

// 获取Cloudflare隧道入站配置详情
func (s *CloudflareTunnelServiceImpl) GetCloudflareTunnelIngressDetail(ctx context.Context, tunnel_id int64, ingress_id int64) (*types.GetCloudflareTunnelIngressDetailResponse, error) {

	// 查询Cloudflare隧道入站规则
	firstCloudflareTunnelIngress, err := s.db.CloudflareTunnelIngress.Query().
		Where(cloudflaretunnelingress.CloudflareTunnelID(tunnel_id)).
		Where(cloudflaretunnelingress.IDEQ(ingress_id)).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询Cloudflare隧道入站规则失败: %w", err)
	}

	return &types.GetCloudflareTunnelIngressDetailResponse{
		ID:            firstCloudflareTunnelIngress.ID,
		Hostname:      firstCloudflareTunnelIngress.Hostname,
		Service:       firstCloudflareTunnelIngress.Service,
		Path:          firstCloudflareTunnelIngress.Path,
		OriginRequest: firstCloudflareTunnelIngress.OriginRequest,
	}, nil
}

// 创建Cloudflare隧道入站配置
func (s *CloudflareTunnelServiceImpl) CreateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, request *types.CreateCloudflareTunnelIngressRequest) error {

	// 查询Cloudflare隧道及关联配置
	firstPlatformCofing, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.IDEQ(tunnel_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare隧道及关联配置失败: %w", err)
	}

	// 获取Cloudflare Tunnel提供程序
	tunProvider, err := s.cloudflareTunnelSync.GetApiClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取Cloudflare Tunnel提供程序失败: %w", err)
	}

	// 处理新数据：根据请求参数生成新的 Hostname 和 Service 字符串
	newHostname, newService := tunProvider.HandleHostnameAndService(
		request.SubDomain,
		request.DomainName,
		request.ServiceType,
		request.ServiceContent,
	)

	// 获取远程隧道入站规则列表
	ingresses, err := tunProvider.IngresssRaw(ctx, firstPlatformCofing.TunnelID)
	if err != nil {
		return fmt.Errorf("获取远程隧道入站规则列表失败: %w", err)
	}

	// 增加新的入站规则
	ingresses = append(ingresses,
		zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress{
			Hostname: newHostname,
			Service:  newService,
			Path:     request.DomainPath,
		},
	)

	// 调用 API 更新 Cloudflare 远程配置
	err = tunProvider.UpdateIngress(ctx, firstPlatformCofing.TunnelID, ingresses)
	if err != nil {
		return fmt.Errorf("同步到 Cloudflare 远程失败: %w", err)
	}

	return nil
}

// 更新Cloudflare隧道入站配置
func (s *CloudflareTunnelServiceImpl) UpdateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64, request *types.UpdateCloudflareTunnelIngressRequest) error {

	// 查询Cloudflare隧道及关联配置
	firstPlatformCofing, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.IDEQ(tunnel_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare隧道及关联配置失败: %w", err)
	}

	// 查询Cloudflare隧道入站规则
	oldRuleRecord, err := s.db.CloudflareTunnelIngress.Query().
		Where(cloudflaretunnelingress.IDEQ(ingress_id)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare隧道入站规则失败: %w", err)
	}

	// 获取Cloudflare Tunnel提供程序
	tunProvider, err := s.cloudflareTunnelSync.GetApiClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取Cloudflare Tunnel提供程序失败: %w", err)
	}

	// 处理新数据：根据请求参数生成新的 Hostname 和 Service 字符串
	newHostname, newService := tunProvider.HandleHostnameAndService(
		request.SubDomain,
		request.DomainName,
		request.ServiceType,
		request.ServiceContent,
	)

	// 获取远程隧道入站规则列表
	remoteIngresses, err := tunProvider.IngresssRaw(ctx, firstPlatformCofing.TunnelID)
	if err != nil {
		return fmt.Errorf("获取远程隧道入站规则列表失败: %w", err)
	}

	// 新的隧道入站规则列表
	var updatedIngresses []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress

	// 旧的入站规则哈希值
	oldHash := gomd5.Get16MD5Encode(oldRuleRecord.Service + oldRuleRecord.Hostname + oldRuleRecord.Path)

	// 是否存在旧的入站规则
	foundInRemote := false

	for _, remoteIngresse := range remoteIngresses {
		// 当前入站规则的哈希值
		currentHash := gomd5.Get16MD5Encode(remoteIngresse.Service + remoteIngresse.Hostname + remoteIngresse.Path)

		// 查找需要旧的规则
		if currentHash == oldHash {
			updatedIngresses = append(updatedIngresses, zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress{
				Hostname: newHostname,
				Service:  newService,
				Path:     request.DomainPath,
			})
			foundInRemote = true
		} else {
			updatedIngresses = append(updatedIngresses, remoteIngresse)
		}
	}

	// 没有找到旧的入站规则
	if !foundInRemote {
		return fmt.Errorf("更新失败：远程规则库中未匹配到原始规则，请同步后再试")
	}

	// 调用 API 更新 Cloudflare 远程配置
	err = tunProvider.UpdateIngress(ctx, firstPlatformCofing.TunnelID, updatedIngresses)
	if err != nil {
		return fmt.Errorf("同步到 Cloudflare 远程失败: %w", err)
	}

	return nil
}

// DeleteCloudflareTunnelIngress 删除Cloudflare隧道入站规则
func (s *CloudflareTunnelServiceImpl) DeleteCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64) error {

	// 创建事务
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 查询Cloudflare隧道及关联配置
	firstPlatformCofing, err := tx.CloudflareTunnel.Query().
		Where(cloudflaretunnel.IDEQ(tunnel_id)).
		WithConfig().
		First(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询Cloudflare隧道及关联配置失败: %w", err)
	}

	// 查询Cloudflare隧道入站规则
	firstCloudflareTunnelIngress, err := tx.CloudflareTunnelIngress.Query().
		Where(cloudflaretunnelingress.IDEQ(ingress_id)).
		First(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询Cloudflare隧道入站规则失败: %w", err)
	}

	// 获取Cloudflare Tunnel提供程序
	tunProvider, err := s.cloudflareTunnelSync.GetApiClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取Cloudflare Tunnel提供程序失败: %w", err)
	}

	// 获取远程隧道入站规则列表
	remoteIngresses, err := tunProvider.IngresssRaw(ctx, firstPlatformCofing.TunnelID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取远程隧道入站规则列表失败: %w", err)
	}

	// 新的隧道入站规则列表
	var updatedIngresses []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress

	// 需要删除的入站规则哈希值
	targetHash := gomd5.Get16MD5Encode(firstCloudflareTunnelIngress.Service + firstCloudflareTunnelIngress.Hostname + firstCloudflareTunnelIngress.Path)

	// 是否存在删除的入站规则
	foundInRemote := false

	for _, remoteIngresse := range remoteIngresses {
		// 当前入站规则的哈希值
		currentHash := gomd5.Get16MD5Encode(remoteIngresse.Service + remoteIngresse.Hostname + remoteIngresse.Path)

		// 查找需要删除的规则
		if currentHash == targetHash {
			foundInRemote = true
			continue
		} else {
			updatedIngresses = append(updatedIngresses, remoteIngresse)
		}
	}

	// 没有找到删除的入站规则
	if !foundInRemote {
		tx.Rollback()
		return fmt.Errorf("删除失败，请先同步入站规则再确认是否存在")
	}

	// 删除Cloudflare隧道入站规则
	err = tx.CloudflareTunnelIngress.DeleteOne(firstCloudflareTunnelIngress).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除Cloudflare隧道入站规则失败: %w", err)
	}

	// 调用 API 更新 Cloudflare 远程配置
	err = tunProvider.UpdateIngress(ctx, firstPlatformCofing.TunnelID, updatedIngresses)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("同步到 Cloudflare 远程失败: %w", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
