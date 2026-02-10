package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/cdn"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// CDNSyncTaskProvider 实现 provider.TaskProvider 接口
type CDNSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db                     *ent.Client                   // 数据库客户端
	logDb                  *ent_log.Client               // 日志数据库客户端
	logger                 *slog.Logger                  // 日志记录器
	httpTransport          *http_log.LoggingRoundTripper // HTTP日志记录器
	webhookSendSyncHandler *WebhookSendTaskProvider      // Webhook发送 处理器

	businessType cronjob.BusinessType // 业务类型
}

// NewCDNSyncTaskProvider 创建一个新的 CDN同步 提供者
func NewCDNSyncTaskProvider(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewCDNSyncTaskProviderRaw(ctx, db, logDb, logger, httpTransport)
}

// NewCDNSyncTaskProvider 创建一个新的 CDN同步 提供者
func NewCDNSyncTaskProviderRaw(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *CDNSyncTaskProvider {
	// 创建 Webhook发送 处理器
	webhookSendSyncHandler := NewWebhookSendTaskProviderRaw(
		ctx,
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &CDNSyncTaskProvider{
		db:                     db,
		logDb:                  logDb,
		logger:                 logger.With(slog.String("handler", "cdn-sync")),
		httpTransport:          httpTransport,
		webhookSendSyncHandler: webhookSendSyncHandler,

		businessType: cronjob.BusinessTypeCdn,
	}
}

// CDNSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	CDNSyncTaskName = "cdn_sync"
)

// CDNSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type CDNSyncTaskArgs struct {
	ID int64 `json:"id"` // CDNRecord 的数据库 ID
}

// Process 执行 CDN同步 逻辑
func (p *CDNSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params CDNSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	record, err := p.db.CDNRecord.Query().
		Where(cdnrecord.IDEQ(params.ID)).
		// 预加载 配置
		WithConfig().
		// 预加载 IP组
		WithIPGroup(func(withIPGroup *ent.IpGroupQuery) {
			// 在 IP组 里查询设备
			withIPGroup.WithDevices(func(withDevices *ent.DeviceQuery) {
				// 单位：分钟
				invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime
				if invalidTime <= 0 {
					return
				}
				// 设备心跳时间在范围内
				withDevices.Where(device.LastHeartbeatAtGTE(time.Now().Add(-time.Duration(invalidTime) * time.Minute)))
				withDevices.OrderByIDAsc()
			})
		}).
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("获取记录失败: %w", err)
	}

	// 判断是否已禁用
	if !record.IsActive {
		return "跳过: 记录已禁用", nil
	}

	// 存放源站
	var originValues []string

	// 获取关联的IP组
	ipGroup, err := record.Edges.IPGroupOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的IP组失败: %w", err)
	}

	// 获取关联的IP组中的设备
	devices, err := ipGroup.Edges.DevicesOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的IP组中的设备失败: %w", err)
	}

	// 遍历设备，根据记录类型筛选 IP
	for _, device := range devices {
		if device.IPV4 != "" && record.Type == cdnrecord.TypeA {
			originValues = append(originValues, device.IPV4)
		}
		if device.IPV6 != "" && record.Type == cdnrecord.TypeAAAA {
			originValues = append(originValues, device.IPV6)
		}
	}

	if len(originValues) == 0 {
		return "跳过: 无可用源站", nil
	}

	// 获取关联的平台配置
	platform, err := record.Edges.ConfigOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的平台配置失败: %w", err)
	}

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

	// 创建CDN更新任务
	cdnTask, err := provider.NewCDNTask(
		p.logger,
		cdnProvider,
		platform.ConfigJSON,
		record.ConfigJSON,
	)
	if err != nil {
		return "", fmt.Errorf("创建CDN更新任务失败: %w", err)
	}

	// 配置CDN参数
	cdnTask.Config(
		record.SubDomain,
		record.Type.String(),
		originValues,
	)

	// 执行CDN更新
	if err = cdnTask.Execute(ctx); err != nil {
		return "", err
	}

	// 执行CDN通知
	if cdnTask.IsNotification() {
		_ = p.webhookSendSyncHandler.AddTask(
			ctx,
			p.db,
			p.businessType,
			record.ID,
			WebhookSendTaskArgs{
				EventTitle:   "CDN记录更新",
				EventType:    "CDN_UPDATE",
				EventTime:    time.Now().Format("2006-01-02 15:04:05"),
				EventBody:    "CDN记录更新事件",
				RecordDomain: record.FullDomain,
				RecordType:   record.Type.String(),
				// RecordOldValue: strings.Join(t.config.oldOriginValues, ","),
				RecordNewValue: strings.Join(originValues, ","),
			},
		)
	}

	return fmt.Sprintf("域名 %s 同步成功 (源站:%v)", record.SubDomain, originValues), nil
}

// LoadTask 加载 CDN同步 周期任务
func (p *CDNSyncTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.CDNRecord, isActive bool) error {

	var (
		taskType    = cronjob.TaskTypeRecurring
		jobType     = cronjob.JobTypeRandom
		jobCategory = cronjob.JobCategorySystem
	)

	// 状态
	targetStatus := cronjob.StatusStopped
	if isActive {
		targetStatus = cronjob.StatusRunning
	}

	// 查找任务
	job, err := db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(taskType)).
		Where(cronjob.BusinessTypeEQ(p.businessType)).
		Where(cronjob.BusinessIDEQ(record.ID)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = db.CronJob.Create().
			SetName("CDN同步: " + record.FullDomain).
			SetHandlerName(CDNSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(CDNSyncTaskArgs{
				ID: record.ID,
			}))).
			SetTaskType(taskType).
			SetJobType(jobType).
			SetJobCategory(jobCategory).
			SetBusinessType(p.businessType).
			SetBusinessID(record.ID).
			SetMinInterval(10).
			SetMaxInterval(30).
			SetIntervalUnit(cronjob.IntervalUnitMinutes).
			SetStatus(targetStatus).
			SetIsSingleton(true). // 单例模式
			SetRunOnStart(false). // 启动时不立即运行
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建任务失败: %w", err)
		}

		return nil
	}

	err = db.CronJob.UpdateOne(job).
		SetStatus(targetStatus).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	return nil
}

// AddTask 添加 CDN同步 一次性任务
func (p *CDNSyncTaskProvider) AddTask(ctx context.Context, db *ent.Client, record *ent.CDNRecord) error {
	return nil
}
