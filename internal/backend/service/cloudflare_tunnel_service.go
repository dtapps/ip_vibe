package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflaretunnel"
	"vibe_ip/internal/ent/cloudflaretunnelingress"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/tun"
	"vibe_ip/internal/types"

	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"github.com/tidwall/gjson"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gomd5"
	"golang.org/x/sync/errgroup"
)

// CloudflareTunnelService Cloudflare隧道服务接口
type CloudflareTunnelService interface {
	SyncAllCloudflareTunnels(ctx context.Context) error
	SyncThisCloudflareTunnels(ctx context.Context, config_id int64) error
	SyncAllCloudflareTunnelIngresses(ctx context.Context) error
	SyncThisCloudflareTunnelIngresses(ctx context.Context, config_id int64) error

	GetCloudflareTunnelIngress(ctx context.Context, tunnel_id int64) ([]*types.GetCloudflareTunnelIngressDetailResponse, error)

	GetCloudflareTunnelIngressDetail(ctx context.Context, tunnel_id int64, ingress_id int64) (*types.GetCloudflareTunnelIngressDetailResponse, error)
	CreateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, request *types.CreateCloudflareTunnelIngressRequest) error
	UpdateCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64, request *types.UpdateCloudflareTunnelIngressRequest) error
	DeleteCloudflareTunnelIngress(ctx context.Context, tunnel_id int64, ingress_id int64) error
}

// CloudflareTunnelServiceImpl Cloudflare隧道服务实现
type CloudflareTunnelServiceImpl struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewCloudflareTunnelService 创建Cloudflare隧道服务实例
func NewCloudflareTunnelService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) CloudflareTunnelService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &CloudflareTunnelServiceImpl{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *CloudflareTunnelServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// getTunClient 获取Cloudflare Tunnel提供程序
func (s *CloudflareTunnelServiceImpl) getTunClient(ctx context.Context, platformConfigJSON json.RawMessage) (*tun.CloudfareTUNProvider, error) {

	// 创建Cloudflare Tunnel提供程序选项
	var tunProviderOpts = make([]tun.Option, 0, 1)
	if s.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的Cloudflare Tunnel配置
	tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
	err := tunProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析Cloudflare隧道配置失败: %w", err)
	}

	// 创建Cloudflare Tunnel提供程序
	tunProvider, err := tun.NewCloudfareTUNProviderRaw(ctx, s.logger, tunProviderCfg, tunProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建Cloudflare隧道提供程序失败: %w", err)
	}

	return tunProvider, nil
}

// SyncAllCloudflareTunnels 同步全部Cloudflare隧道
func (s *CloudflareTunnelServiceImpl) SyncAllCloudflareTunnels(ctx context.Context) error {
	// 查询全部Cloudflare平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(platformconfig.CapabilityAll, platformconfig.CapabilityTun)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部Cloudflare平台失败: %w", err)
	}

	// 并发同步Cloudflare隧道
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisCloudflareTunnels(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发失败: %w", err)
	}

	return nil
}

// SyncThisCloudflareTunnels 同步某个Cloudflare隧道
func (s *CloudflareTunnelServiceImpl) SyncThisCloudflareTunnels(ctx context.Context, config_id int64) error {
	// 查询Cloudflare平台
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(platformconfig.CapabilityAll, platformconfig.CapabilityTun)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare平台失败: %w", err)
	}

	// 判断是否存在 账号ID
	if len(firstPlatformCofing.ConfigJSON) == 0 || gjson.GetBytes(firstPlatformCofing.ConfigJSON, "account_id").String() == "" {
		return nil
	}

	// 获取Cloudflare Tunnel提供程序
	tunProvider, err := s.getTunClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("获取Cloudflare隧道提供程序失败: %w", err)
	}

	// 查询隧道列表
	connectors, err := tunProvider.ConnectorsRaw(ctx)
	if err != nil {
		return err
	}

	// 入库
	for _, connector := range connectors {
		// 检查是否已存在
		firstTunnel, err := s.db.CloudflareTunnel.Query().
			Where(cloudflaretunnel.TunnelIDEQ(connector.ID)).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询Cloudflare隧道记录失败: %w", err)
			}
			// 创建记录
			err = s.db.CloudflareTunnel.Create().
				SetConfigID(firstPlatformCofing.ID).
				SetTunnelID(connector.ID).
				SetName(connector.Name).
				SetStatus(cloudflaretunnel.Status(connector.Status)).
				SetConnections([]byte(gojson.JsonEncodeNoError(connector.Connections))).
				SetConnsActiveAt(connector.ConnsActiveAt).
				SetRemoteCreatedAt(connector.CreatedAt).
				SetSyncAt(time.Now()).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建Cloudflare隧道记录失败: %w", err)
			}
		} else {
			// 更新记录
			err = s.db.CloudflareTunnel.UpdateOne(firstTunnel).
				SetName(connector.Name).
				SetStatus(cloudflaretunnel.Status(connector.Status)).
				SetConnections([]byte(gojson.JsonEncodeNoError(connector.Connections))).
				SetConnsActiveAt(connector.ConnsActiveAt).
				SetSyncAt(time.Now()).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新Cloudflare隧道记录失败: %w", err)
			}
		}
	}

	return nil
}

// SyncAllCloudflareTunnelIngresses 同步全部Cloudflare隧道入站规则
func (s *CloudflareTunnelServiceImpl) SyncAllCloudflareTunnelIngresses(ctx context.Context) error {
	// 查询全部Cloudflare平台
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(platformconfig.CapabilityAll, platformconfig.CapabilityTun)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询全部Cloudflare平台失败: %w", err)
	}

	// 并发同步Cloudflare隧道
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemPlatformCofingID := range allPlatformCofingIDs {
		itemPlatformCofingID := itemPlatformCofingID
		eg.Go(func() error {
			return s.SyncThisCloudflareTunnelIngresses(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发失败: %w", err)
	}

	return nil
}

// SyncThisCloudflareTunnels 同步某个Cloudflare隧道入站规则
func (s *CloudflareTunnelServiceImpl) SyncThisCloudflareTunnelIngresses(ctx context.Context, config_id int64) error {
	// 查询Cloudflare隧道
	allCloudflareTunnel, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.ConfigIDEQ(config_id)).
		WithConfig().
		WithIngresses().
		All(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare隧道失败: %w", err)
	}

	// 并发同步Cloudflare隧道
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(5) // 最多 5 个并发
	for _, itemCloudflareTunnel := range allCloudflareTunnel {
		itemCloudflareTunnel := itemCloudflareTunnel
		eg.Go(func() error {

			// 判断是否存在 账号ID
			if len(itemCloudflareTunnel.Edges.Config.ConfigJSON) == 0 || gjson.GetBytes(itemCloudflareTunnel.Edges.Config.ConfigJSON, "account_id").String() == "" {
				return nil
			}

			// 获取Cloudflare Tunnel提供程序
			tunProvider, err := s.getTunClient(ctx, itemCloudflareTunnel.Edges.Config.ConfigJSON)
			if err != nil {
				return fmt.Errorf("获取Cloudflare隧道提供程序失败: %w", err)
			}

			// 查询隧道入站规则列表
			ingresses, err := tunProvider.IngresssRaw(egCtx, itemCloudflareTunnel.TunnelID)
			if err != nil {
				return err
			}

			// 先清空旧的入站规则
			if len(itemCloudflareTunnel.Edges.Ingresses) > 0 {
				_, err = s.db.CloudflareTunnelIngress.Delete().
					Where(cloudflaretunnelingress.CloudflareTunnelIDEQ(itemCloudflareTunnel.ID)).
					Exec(egCtx)
				if err != nil {
					return fmt.Errorf("删除Cloudflare隧道入站规则失败: %w", err)
				}
			}

			// 批量创建Cloudflare隧道入站规则
			createBulkList := make([]*ent.CloudflareTunnelIngressCreate, 0, len(ingresses))
			for _, ingress := range ingresses {
				createBulkList = append(createBulkList,
					s.db.CloudflareTunnelIngress.Create().
						SetCloudflareTunnelID(itemCloudflareTunnel.ID).
						SetHostname(ingress.Hostname).
						SetService(ingress.Service).
						SetPath(ingress.Path).
						SetOriginRequest([]byte(gojson.JsonEncodeNoError(ingress.OriginRequest))).
						SetSyncAt(time.Now()),
				)
			}
			if len(createBulkList) > 0 {
				err = s.db.CloudflareTunnelIngress.CreateBulk(createBulkList...).
					Exec(egCtx)
				if err != nil {
					return fmt.Errorf("批量创建Cloudflare隧道入站规则记录失败: %w", err)
				}
			}

			return nil
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发失败: %w", err)
	}

	return nil
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
	tunProvider, err := s.getTunClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
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
	tunProvider, err := s.getTunClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
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
	tunProvider, err := s.getTunClient(ctx, firstPlatformCofing.Edges.Config.ConfigJSON)
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
