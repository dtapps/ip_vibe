package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflaretunnel"
	"vibe_ip/internal/ent/cloudflaretunnelingress"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/tun"
	"vibe_ip/internal/types"

	"github.com/tidwall/gjson"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"golang.org/x/sync/errgroup"
)

// CloudflareTunnelService Cloudflare隧道服务接口
type CloudflareTunnelService interface {
	BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error)
	SyncAllCloudflareTunnels(ctx context.Context) error
	SyncCloudflareTunnels(ctx context.Context, configID int64) error
	SyncAllCloudflareTunnelIngresses(ctx context.Context) error
	SyncCloudflareTunnelIngresses(ctx context.Context, configID int64) error
	GetCloudflareTunnelIngresses(ctx context.Context, id int64) ([]*types.GetCloudflareTunnelTunnelIngresse, error)
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

// BuildResponse 将实体转换为响应
func (s *CloudflareTunnelServiceImpl) BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error) {
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
			return s.SyncCloudflareTunnels(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发失败: %w", err)
	}

	return nil
}

// SyncCloudflareTunnels 同步某个Cloudflare隧道
func (s *CloudflareTunnelServiceImpl) SyncCloudflareTunnels(ctx context.Context, configID int64) error {
	// 查询Cloudflare平台
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(configID)).
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

	// 创建Cloudfare TUN 提供程序选项
	var tunProviderOpts = make([]tun.Option, 0, 1)
	if s.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的Cloudfare TUN 配置
	tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
	err = tunProviderCfg.ParseCustomConfig(firstPlatformCofing.ConfigJSON)
	if err != nil {
		return fmt.Errorf("解析Cloudflare隧道配置失败: %w", err)
	}

	// 创建Cloudfare TUN 提供程序
	tunProvider, err := tun.NewCloudfareTUNProviderRaw(ctx, s.logger, tunProviderCfg, tunProviderOpts...)
	if err != nil {
		return fmt.Errorf("创建Cloudflare隧道提供程序失败: %w", err)
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
			return s.SyncCloudflareTunnelIngresses(egCtx, itemPlatformCofingID)
		})
	}

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("并发失败: %w", err)
	}

	return nil
}

// SyncCloudflareTunnels 同步某个Cloudflare隧道入站规则
func (s *CloudflareTunnelServiceImpl) SyncCloudflareTunnelIngresses(ctx context.Context, configID int64) error {
	// 查询Cloudflare隧道
	allCloudflareTunnel, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.ConfigIDEQ(configID)).
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

			// 创建Cloudfare TUN 提供程序选项
			var tunProviderOpts = make([]tun.Option, 0, 1)
			if s.httpTransport != nil {
				tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
			}

			// 创建默认的Cloudfare TUN 配置
			tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
			err = tunProviderCfg.ParseCustomConfig(itemCloudflareTunnel.Edges.Config.ConfigJSON)
			if err != nil {
				return fmt.Errorf("解析Cloudflare隧道配置失败: %w", err)
			}

			// 创建Cloudfare TUN 提供程序
			tunProvider, err := tun.NewCloudfareTUNProviderRaw(egCtx, s.logger, tunProviderCfg, tunProviderOpts...)
			if err != nil {
				return fmt.Errorf("创建Cloudflare隧道提供程序失败: %w", err)
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

// GetCloudflareTunnelIngresses 获取Cloudflare隧道全部入站规则
func (s *CloudflareTunnelServiceImpl) GetCloudflareTunnelIngresses(ctx context.Context, id int64) ([]*types.GetCloudflareTunnelTunnelIngresse, error) {
	// 查询Cloudflare隧道
	firstCloudflareTunnel, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.IDEQ(id)).
		WithIngresses().
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询Cloudflare隧道失败: %w", err)
	}

	ingresses := make([]*types.GetCloudflareTunnelTunnelIngresse, 0, len(firstCloudflareTunnel.Edges.Ingresses))
	for _, ingress := range firstCloudflareTunnel.Edges.Ingresses {
		ingresses = append(ingresses, &types.GetCloudflareTunnelTunnelIngresse{
			ID:            ingress.ID,
			Hostname:      ingress.Hostname,
			Service:       ingress.Service,
			Path:          ingress.Path,
			OriginRequest: ingress.OriginRequest,
		})
	}

	return ingresses, nil
}
