package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/dns"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// DNSSyncTaskProvider 实现 provider.TaskProvider 接口
type DNSSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db                     *ent.Client                   // 数据库客户端
	logDb                  *ent_log.Client               // 日志数据库客户端
	logger                 *slog.Logger                  // 日志记录器
	httpTransport          *http_log.LoggingRoundTripper // HTTP日志记录器
	webhookSendSyncHandler *WebhookSendTaskProvider      // Webhook发送 处理器

	businessType cronjob.BusinessType // 业务类型
}

// NewDNSSyncTaskProvider 创建一个新的 DNS同步 提供者
func NewDNSSyncTaskProvider(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewDNSSyncTaskProviderRaw(ctx, db, logDb, logger, httpTransport)
}

// NewDNSSyncTaskProvider 创建一个新的 DNS同步 提供者
func NewDNSSyncTaskProviderRaw(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *DNSSyncTaskProvider {
	// 创建 Webhook发送 处理器
	webhookSendSyncHandler := NewWebhookSendTaskProviderRaw(
		ctx,
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &DNSSyncTaskProvider{
		db:                     db,
		logDb:                  logDb,
		logger:                 logger.With(slog.String("handler", "cdn-sync")),
		httpTransport:          httpTransport,
		webhookSendSyncHandler: webhookSendSyncHandler,

		businessType: cronjob.BusinessTypeDNS,
	}
}

// DNSSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	DNSSyncTaskName = "dns_sync"
)

// DNSSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type DNSSyncTaskArgs struct {
	ID int64 `json:"id"` // DNSRecord 的数据库 ID
}

// Process 执行 DNS同步 逻辑
func (p *DNSSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params DNSSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	record, err := p.db.DNSRecord.Query().
		Where(dnsrecord.IDEQ(params.ID)).
		// 预加载 配置
		WithConfig().
		// 预加载 CDN记录
		WithCdnRecord().
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
	var originValue string

	if record.Type == dnsrecord.TypeCNAME {

		// 获取关联的CDN记录
		cdnRecord, err := record.Edges.CdnRecordOrErr()
		if err != nil {
			return "", fmt.Errorf("获取关联的CDN记录失败: %w", err)
		}

		originValue = cdnRecord.CdnDomain

	} else {

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
			if device.IPV4 != "" && record.Type == dnsrecord.TypeA {
				originValue = device.IPV4
			}
			if device.IPV6 != "" && record.Type == dnsrecord.TypeAAAA {
				originValue = device.IPV6
			}
		}

	}

	if len(originValue) == 0 {
		return "跳过: 无可用源站", nil
	}

	// 获取关联的平台配置
	platform, err := record.Edges.ConfigOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的平台配置失败: %w", err)
	}

	var dnsProvider provider.DNSProvider
	var dnsProviderOpts = make([]dns.Option, 0, len(schematype.GetConfigPlatforms))
	if p.httpTransport != nil {
		dnsProviderOpts = append(dnsProviderOpts, dns.WithHttpTransport(p.httpTransport))
	}

	switch platform.Platform {
	case platformconfig.PlatformCloudflare:
		dnsProviderCfg := dns.NewDefaultCloudfareDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewCloudfareDNSProvider(
			ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformTencentDNS:
		dnsProviderCfg := dns.NewDefaultTencentDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewTencentDNSProvider(
			ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformAliyunDNS:
		dnsProviderCfg := dns.NewDefaultAliyunDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewAliyunDNSProvider(
			ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformBaiduDNS:
		dnsProviderCfg := dns.NewDefaultBaiduDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewBaiduDNSProvider(
			ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformHuaweiDNS:
		dnsProviderCfg := dns.NewDefaultHuaweiDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewHuaweiDNSProvider(ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	case platformconfig.PlatformVolcengineDNS:
		dnsProviderCfg := dns.NewDefaultVolcengineDNSConfig()
		err = dnsProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		dnsProvider, err = dns.NewVolcengineDNSProvider(ctx,
			p.logger,
			dnsProviderCfg,
			dnsProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return "", fmt.Errorf("暂不支持该平台: %s", platform.String())
	}

	// 创建DNS更新任务
	dnsTask, err := provider.NewDNSTask(
		p.logger,
		dnsProvider,
		platform.ConfigJSON,
		record.ConfigJSON,
	)
	if err != nil {
		return "", fmt.Errorf("创建DNS更新任务失败: %w", err)
	}

	// 执行DNS更新
	err = dnsTask.Execute(ctx, &provider.DNSConfigParams{
		DomainName:  record.DomainName,
		SubDomain:   record.SubDomain,
		FullDomain:  record.FullDomain,
		RecordType:  record.Type.String(),
		OriginValue: originValue,
	})
	if err != nil {
		return "", err
	}

	// 准备更新
	updateOne := record.Update()
	isUpdate := false

	// 获取远程记录ID
	remoteRecordID := dnsTask.GetRecordID()
	if remoteRecordID != "" {
		if !strings.EqualFold(record.RecordID, remoteRecordID) {
			updateOne.SetRecordID(remoteRecordID)
			isUpdate = true
		}
	}

	// 执行更新
	if isUpdate {
		err = updateOne.Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("更新DNS记录失败: %w", err)
		}
	}

	// 执行DNS通知
	if dnsTask.IsNotification() {
		err = p.webhookSendSyncHandler.AddTask(
			ctx,
			p.db,
			p.businessType,
			record.ID,
			WebhookSendTaskArgs{
				EventTitle:   "DNS记录更新",
				EventType:    "DNS_UPDATE",
				EventTime:    time.Now().Format("2006-01-02 15:04:05"),
				EventBody:    "DNS记录更新事件",
				RecordDomain: record.FullDomain,
				RecordType:   record.Type.String(),
				// RecordOldValue: t.config.oldOriginValue,
				RecordNewValue: originValue,
			},
		)
		if err != nil {
			return "", fmt.Errorf("添加DNS通知任务失败: %w", err)
		}
	}

	return fmt.Sprintf("域名 %s 同步成功 (源站:%v)", record.SubDomain, originValue), nil
}

// LoadTask 加载 DNS同步 周期任务
func (p *DNSSyncTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.DNSRecord, isActive bool) error {

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
			SetName("DNS同步: " + record.FullDomain).
			SetHandlerName(DNSSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(DNSSyncTaskArgs{
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

// AddTask 添加 DNS同步 一次性任务
func (p *DNSSyncTaskProvider) AddTask(ctx context.Context, db *ent.Client, record *ent.DNSRecord) error {
	return nil
}
