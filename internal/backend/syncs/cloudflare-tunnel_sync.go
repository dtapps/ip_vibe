package syncs

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

	"github.com/tidwall/gjson"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"golang.org/x/sync/errgroup"
)

// Cloudflare隧道同步
type CloudflareTunnelSync struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 调试模式
}

// NewCloudflareTunnelSync 创建Cloudflare隧道同步实例
func NewCloudflareTunnelSync(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) *CloudflareTunnelSync {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &CloudflareTunnelSync{
		db:            db,
		logDb:         logDb,
		logger:        logger,
		httpTransport: syncHttpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *CloudflareTunnelSync) SetDebug(debug bool) {
	s.debug = debug
}

// 获取Cloudflare提供程序
func (s *CloudflareTunnelSync) GetApiClient(ctx context.Context, platformConfigJSON json.RawMessage) (*tun.CloudfareTUNProvider, error) {

	// 创建Cloudflare TUN提供程序选项
	var tunProviderOpts = make([]tun.Option, 0, 1)
	if s.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(s.httpTransport))
	}

	// 创建默认的Cloudflare TUN配置
	tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
	err := tunProviderCfg.ParseCustomConfig(platformConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("解析Cloudflare TUN配置失败: %w", err)
	}

	// 创建Cloudflare TUN提供程序
	tunProvider, err := tun.NewCloudfareTUNProviderRaw(ctx, s.logger, tunProviderCfg, tunProviderOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建Cloudflare TUN提供程序失败: %w", err)
	}

	return tunProvider, nil
}

// 同步 全部账号 全部隧道
func (s *CloudflareTunnelSync) SyncAllCloudflareTunnels(ctx context.Context) error {

	// 查询平台配置
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 并发同步
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
		return fmt.Errorf("并发同步失败: %w", err)
	}

	return nil
}

// 同步 某个账号 全部隧道
func (s *CloudflareTunnelSync) SyncThisCloudflareTunnels(ctx context.Context, config_id int64) error {

	// 查询Cloudflare平台配置
	firstPlatformCofing, err := s.db.PlatformConfig.Query().
		Where(platformconfig.IDEQ(config_id)).
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
	}

	// 判断是否存在 账号ID
	if len(firstPlatformCofing.ConfigJSON) == 0 || gjson.GetBytes(firstPlatformCofing.ConfigJSON, "account_id").String() == "" {
		return nil
	}

	// 获取Cloudflare提供程序
	tunProvider, err := s.GetApiClient(ctx, firstPlatformCofing.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取Cloudflare提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 查询隧道列表
	connectors, err := tunProvider.ConnectorsRaw(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "查询隧道列表失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 入库
	for _, connector := range connectors {
		// 查询隧道记录
		firstTunnel, err := s.db.CloudflareTunnel.Query().
			Where(cloudflaretunnel.ConfigIDEQ(config_id)).
			Where(cloudflaretunnel.TunnelIDEQ(connector.ID)).
			First(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("查询隧道记录失败: %w", err)
			}
			// 创建隧道记录
			createOne := s.db.CloudflareTunnel.Create()
			createOne.SetRemoteCreatedAt(connector.CreatedAt)
			createOne.SetSyncAt(time.Now())
			createOne.SetConfigID(firstPlatformCofing.ID)
			createOne.SetTunnelID(connector.ID)
			createOne.SetConnections(json.RawMessage(gojson.JsonEncodeNoError(connector.Connections)))
			createOne.SetConnsActiveAt(connector.ConnsActiveAt)
			createOne.SetConnsInactiveAt(connector.ConnsInactiveAt)
			createOne.SetName(connector.Name)
			createOne.SetStatus(cloudflaretunnel.Status(connector.Status))
			if connector.TunType != "" {
				createOne.SetTunType(cloudflaretunnel.TunType(connector.TunType))
			}
			err = createOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("创建隧道记录失败: %w", err)
			}
		} else {
			// 更新隧道记录
			updateOne := s.db.CloudflareTunnel.UpdateOne(firstTunnel)
			updateOne.SetSyncAt(time.Now())
			updateOne.SetConnections(json.RawMessage(gojson.JsonEncodeNoError(connector.Connections)))
			updateOne.SetConnsActiveAt(connector.ConnsActiveAt)
			updateOne.SetConnsInactiveAt(connector.ConnsInactiveAt)
			updateOne.SetName(connector.Name)
			updateOne.SetStatus(cloudflaretunnel.Status(connector.Status))
			if connector.TunType != "" {
				updateOne.SetTunType(cloudflaretunnel.TunType(connector.TunType))
			}
			err = updateOne.Exec(ctx)
			if err != nil {
				return fmt.Errorf("更新隧道记录失败: %w", err)
			}
		}
	}

	return nil
}

// 同步 全部账号 全部隧道入站规则
func (s *CloudflareTunnelSync) SyncAllCloudflareTunnelIngresses(ctx context.Context) error {

	// 查询平台配置
	allPlatformCofingIDs, err := s.db.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("查询平台配置失败: %w", err)
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

// 同步 某个账号 全部隧道入站规则
func (s *CloudflareTunnelSync) SyncThisCloudflareTunnelIngresses(ctx context.Context, config_id int64) error {
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

			// 获取Cloudflare提供程序
			tunProvider, err := s.GetApiClient(ctx, itemCloudflareTunnel.Edges.Config.ConfigJSON)
			if err != nil {
				s.logger.ErrorContext(ctx, "获取Cloudflare提供程序失败",
					slog.Any("err", err),
				)
				return nil
			}

			// 查询隧道入站规则列表
			ingresses, err := tunProvider.IngresssRaw(egCtx, itemCloudflareTunnel.TunnelID)
			if err != nil {
				s.logger.ErrorContext(ctx, "查询隧道入站规则列表失败",
					slog.Any("err", err),
				)
				return nil
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
				createOne := s.db.CloudflareTunnelIngress.Create()
				createOne.SetSyncAt(time.Now())
				createOne.SetCloudflareTunnelID(itemCloudflareTunnel.ID)
				createOne.SetHostname(ingress.Hostname)
				createOne.SetService(ingress.Service)
				createOne.SetPath(ingress.Path)
				createOne.SetOriginRequest([]byte(gojson.JsonEncodeNoError(ingress.OriginRequest)))
				createBulkList = append(
					createBulkList,
					createOne,
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

// 同步 某个账号 某个隧道入站规则
func (s *CloudflareTunnelSync) SyncThisCloudflareThisTunnelIngresses(ctx context.Context, config_id int64, tunnel_id int64) error {
	// 查询Cloudflare隧道
	firstCloudflareTunnel, err := s.db.CloudflareTunnel.Query().
		Where(cloudflaretunnel.ConfigIDEQ(config_id)).
		Where(cloudflaretunnel.IDEQ(tunnel_id)).
		WithConfig().
		WithIngresses().
		First(ctx)
	if err != nil {
		return fmt.Errorf("查询Cloudflare隧道失败: %w", err)
	}

	// 判断是否存在 账号ID
	if len(firstCloudflareTunnel.Edges.Config.ConfigJSON) == 0 || gjson.GetBytes(firstCloudflareTunnel.Edges.Config.ConfigJSON, "account_id").String() == "" {
		return nil
	}

	// 获取Cloudflare提供程序
	tunProvider, err := s.GetApiClient(ctx, firstCloudflareTunnel.Edges.Config.ConfigJSON)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取Cloudflare提供程序失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 查询隧道入站规则列表
	ingresses, err := tunProvider.IngresssRaw(ctx, firstCloudflareTunnel.TunnelID)
	if err != nil {
		s.logger.ErrorContext(ctx, "查询隧道入站规则列表失败",
			slog.Any("err", err),
		)
		return nil
	}

	// 先清空旧的入站规则
	if len(firstCloudflareTunnel.Edges.Ingresses) > 0 {
		_, err = s.db.CloudflareTunnelIngress.Delete().
			Where(cloudflaretunnelingress.CloudflareTunnelIDEQ(firstCloudflareTunnel.ID)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("删除Cloudflare隧道入站规则失败: %w", err)
		}
	}

	// 批量创建Cloudflare隧道入站规则
	createBulkList := make([]*ent.CloudflareTunnelIngressCreate, 0, len(ingresses))
	for _, ingress := range ingresses {
		createOne := s.db.CloudflareTunnelIngress.Create()
		createOne.SetSyncAt(time.Now())
		createOne.SetCloudflareTunnelID(firstCloudflareTunnel.ID)
		createOne.SetHostname(ingress.Hostname)
		createOne.SetService(ingress.Service)
		createOne.SetPath(ingress.Path)
		createOne.SetOriginRequest([]byte(gojson.JsonEncodeNoError(ingress.OriginRequest)))
		createBulkList = append(
			createBulkList,
			createOne,
		)
	}
	if len(createBulkList) > 0 {
		err = s.db.CloudflareTunnelIngress.CreateBulk(createBulkList...).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("批量创建Cloudflare隧道入站规则记录失败: %w", err)
		}
	}

	return nil
}
