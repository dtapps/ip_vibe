package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/cdn"

	"go.dtapp.net/library/contrib/http_log"
)

// CDNTaskProvider 实现 provider.TaskProvider 接口
type CDNTaskProvider struct {
	db             *ent.Client                   // 数据库客户端
	logDb          *ent_log.Client               // 日志数据库客户端
	logger         *slog.Logger                  // 日志记录器
	httpTransport  *http_log.LoggingRoundTripper // HTTP日志记录器
	webhookService *webhookService.Send          // Webhook服务
}

func NewCDNTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *CDNTaskProvider {
	webhookService := webhookService.NewSend(db, logDb)
	return &CDNTaskProvider{
		db:             db,
		logDb:          logDb,
		logger:         logger.With(slog.String("handler", "cdn_sync")),
		httpTransport:  httpTransport,
		webhookService: webhookService,
	}
}

// CDNTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type CDNTaskArgs struct {
	ID int64 `json:"id"` // CDNRecord 的数据库 ID
}

// Process 执行单条 CDN 记录的更新逻辑
func (p *CDNTaskProvider) Process(ctx context.Context, args json.RawMessage) (string, error) {
	// 解析参数
	var params CDNTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	record, err := p.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(params.ID), cdnrecord.IsActiveEQ(true)).
		WithConfig().
		WithIPGroup(func(q *ent.IpGroupQuery) {
			q.WithDevices(func(dq *ent.DeviceQuery) {
				invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime
				if invalidTime > 0 {
					dq.Where(device.LastHeartbeatAtGTE(time.Now().Add(-time.Duration(invalidTime) * time.Minute)))
				}
			})
		}).
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("获取记录失败或记录已禁用: %w", err)
	}

	// 提取源站 IP
	var originValues []string
	if record.Edges.IPGroup != nil {
		for _, dev := range record.Edges.IPGroup.Edges.Devices {
			if dev.IPV4 != "" && record.Type == cdnrecord.TypeA {
				originValues = append(originValues, dev.IPV4)
			}
			if dev.IPV6 != "" && record.Type == cdnrecord.TypeAAAA {
				originValues = append(originValues, dev.IPV6)
			}
		}
	}

	if len(originValues) == 0 {
		return "跳过: 无可用源站IP", nil
	}

	// 平台配置
	platform := record.Edges.Config

	var cdnProvider provider.CDNProvider
	var cdnProviderOpts = make([]cdn.Option, 0, len(schematype.GetConfigPlatforms))
	if p.httpTransport != nil {
		cdnProviderOpts = append(cdnProviderOpts, cdn.WithHttpTransport(p.httpTransport))
	}

	switch platform.Platform {
	case platformconfig.PlatformTencentCdn:
		cdnProviderCfg := cdn.NewDefaultTencentCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewTencentCDNProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformTencentEo:
		cdnProviderCfg := cdn.NewDefaultTencentEOConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewTencentEOProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunCdn:
		cdnProviderCfg := cdn.NewDefaultAliyunCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewAliyunCDNProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunEsa:
		cdnProviderCfg := cdn.NewDefaultAliyunESAConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewAliyunESAProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformBaiduCdn:
		cdnProviderCfg := cdn.NewDefaultBaiduCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewBaiduCDNProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformHuaweiCdn:
		cdnProviderCfg := cdn.NewDefaultHuaweiCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewHuaweiCDNProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformQiniuCdn:
		cdnProviderCfg := cdn.NewDefaultQiniuCDNConfig()
		err = cdnProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		cdnProvider, err = cdn.NewQiniuCDNProvider(
			ctx,
			p.logger,
			cdnProviderCfg,
			cdnProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return "", fmt.Errorf("暂不支持该平台: %s", platform.String())
	}

	// 创建 CDN 更新任务
	cdnTask, err := provider.NewCDNTask(
		p.logger,
		cdnProvider,
		platform.ConfigJSON,
		record.ConfigJSON,
	)
	if err != nil {
		return "", fmt.Errorf("创建CDN任务失败: %w", err)
	}

	// 配置 CDN 参数
	cdnTask.Config(record.SubDomain, record.Type.String(), originValues)

	// 执行 DNS 更新
	if err = cdnTask.Execute(ctx); err != nil {
		// 保存 DNS 记录响应
		_ = cdnTask.SaveResponse(ctx, p.db, record.ID, err.Error())
		return "", err
	}
	defer func() {
		// 保存 DNS 记录响应
		_ = cdnTask.SaveResponse(ctx, p.db, record.ID, "ok")
	}()

	// 保存 DNS 记录刷新
	_ = cdnTask.SaveRefresh(ctx, p.db, record.ID)

	// 执行 DNS 通知
	if p.webhookService != nil {
		_ = cdnTask.Notify(ctx, p.webhookService, record.ID)
	}

	return fmt.Sprintf("域名 %s 同步成功 (IP:%v)", record.SubDomain, originValues), nil
}
