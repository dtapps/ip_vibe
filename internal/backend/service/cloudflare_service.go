package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider/tun"
	"vibe_ip/internal/types"

	"github.com/tidwall/gjson"
	"go.dtapp.net/library/contrib/http_log"
)

// CloudflareService Cloudflare服务接口
type CloudflareService interface {
	BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error)
	GetCloudflareDomains(ctx context.Context, id int64) ([]*types.GetCloudflareDomainResponse, error)
}

// CloudflareServiceImpl Cloudflare服务实现
type CloudflareServiceImpl struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewCloudflareService 创建Cloudflare服务实例
func NewCloudflareService(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) CloudflareService {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &CloudflareServiceImpl{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *CloudflareServiceImpl) SetDebug(debug bool) {
	s.debug = debug
}

// BuildResponse 将实体转换为响应
func (s *CloudflareServiceImpl) BuildResponse(ctx context.Context, item *ent.TUNRecord) (*types.GetTunRecordResponse, error) {
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

// getTunClient 获取Cloudflare Tunnel提供程序
func (s *CloudflareServiceImpl) getTunClient(ctx context.Context, platformConfigJSON json.RawMessage) (*tun.CloudfareTUNProvider, error) {

	// 创建Cloudflare Tunnel提供程序选项
	var tunProviderOpts = make([]tun.Option, 0, 1)
	if s.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的Cloudflare Tunnel配置
	tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
	err := tunProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析Cloudflare配置失败: %w", err)
	}

	// 创建Cloudflare Tunnel提供程序
	tunProvider, err := tun.NewCloudfareTUNProviderRaw(ctx, s.logger, tunProviderCfg, tunProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建Cloudflare提供程序失败: %w", err)
	}

	return tunProvider, nil
}

// GetCloudflareDomains 获取Cloudflare域名
func (s *CloudflareServiceImpl) GetCloudflareDomains(ctx context.Context, config_id int64) ([]*types.GetCloudflareDomainResponse, error) {
	// 查询Cloudflare平台
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(platformconfig.CapabilityAll, platformconfig.CapabilityTun)).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询Cloudflare平台失败: %w", err)
	}

	// 判断是否存在 账号ID
	if len(firstPlatformCofing.ConfigJSON) == 0 || gjson.GetBytes(firstPlatformCofing.ConfigJSON, "account_id").String() == "" {
		return nil, fmt.Errorf("Cloudflare平台配置中不存在账号ID")
	}

	// 获取Cloudflare Tunnel提供程序
	tunProvider, err := s.getTunClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("获取Cloudflare隧道提供程序失败: %w", err)
	}

	// 查询域名列表
	zones, err := tunProvider.ZonesRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询Cloudflare域名列表失败: %w", err)
	}

	// 转换为响应
	domains := make([]*types.GetCloudflareDomainResponse, 0, len(zones))
	for _, zone := range zones {
		domains = append(domains, &types.GetCloudflareDomainResponse{
			ZoneID:   zone.ID,
			ZoneName: zone.Name,
		})
	}

	return domains, nil
}
