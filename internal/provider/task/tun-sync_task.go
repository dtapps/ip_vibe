package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/tunrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/provider/tun"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// TUNSyncTaskProvider 实现 provider.TaskProvider 接口
type TUNSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db                     *ent.Client                   // 数据库客户端
	logDb                  *ent_log.Client               // 日志数据库客户端
	logger                 *slog.Logger                  // 日志记录器
	httpTransport          *http_log.LoggingRoundTripper // HTTP日志记录器
	webhookSendSyncHandler *WebhookSendTaskProvider      // Webhook发送 处理器

	businessType cronjob.BusinessType // 业务类型
}

// NewTUNSyncTaskProvider 创建一个新的 TUN同步 提供者
func NewTUNSyncTaskProvider(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) provider.TaskProvider {
	return NewTUNSyncTaskProviderRaw(ctx, db, logDb, logger)
}

// NewTUNSyncTaskProvider 创建一个新的 TUN同步 提供者
func NewTUNSyncTaskProviderRaw(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) *TUNSyncTaskProvider {
	// 创建 Webhook发送 处理器
	webhookSendSyncHandler := NewWebhookSendTaskProviderRaw(
		ctx,
		db,
		logDb,
		logger,
	)
	return &TUNSyncTaskProvider{
		db:                     db,
		logDb:                  logDb,
		logger:                 logger.With(slog.String("handler", "tun-sync")),
		httpTransport:          taskHttpTransport,
		webhookSendSyncHandler: webhookSendSyncHandler,

		businessType: cronjob.BusinessTypeTun,
	}
}

// TUNSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	TUNSyncTaskName = "tun_sync"
)

// TUNSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type TUNSyncTaskArgs struct {
	ID int64 `json:"id"` // TUNRecord 的数据库 ID
}

// Process 执行 TUN同步 逻辑
func (p *TUNSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params TUNSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	record, err := p.db.TUNRecord.Query().
		Where(tunrecord.IDEQ(params.ID)).
		// 预加载 配置
		WithConfig().
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("获取记录失败: %w", err)
	}

	// 判断是否已禁用
	if !record.IsActive {
		return "跳过: 记录已禁用", nil
	}

	// 获取关联的平台配置
	platform, err := record.Edges.ConfigOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的平台配置失败: %w", err)
	}

	var tunProvider provider.TUNProvider
	var tunProviderOpts = make([]tun.Option, 0, len(schematype.GetConfigPlatforms))
	if p.httpTransport != nil {
		tunProviderOpts = append(tunProviderOpts, tun.WithHttpTransport(p.httpTransport))
	}

	switch platform.Platform {
	case platformconfig.PlatformCloudflare:
		tunProviderCfg := tun.NewDefaultCloudfareTUNConfig()
		err = tunProviderCfg.ParseCustomConfig(platform.ConfigJSON)
		if err != nil {
			return "", fmt.Errorf("解析自定义配置失败: %w", err)
		}
		tunProvider, err = tun.NewCloudfareTUNProvider(
			ctx,
			p.logger,
			tunProviderCfg,
			tunProviderOpts...,
		)
		if err != nil {
			return "", fmt.Errorf("初始提供程序失败: %w", err)
		}
	default:
		return "", fmt.Errorf("暂不支持该平台: %s", platform.String())
	}

	// 创建TUN更新任务
	tunTask, err := provider.NewTUNTask(
		p.logger,
		tunProvider,
		platform.ConfigJSON,
		record.ConfigJSON,
	)
	if err != nil {
		return "", fmt.Errorf("创建TUN更新任务失败: %w", err)
	}

	// 配置TUN参数
	tunTask.Config(record.SubDomain,
		record.DomainPath,
		record.ServeType.String(),
		record.ServeURI,
	)

	// 执行TUN更新
	if err = tunTask.Execute(ctx); err != nil {
		return "", err
	}

	// 执行TUN通知
	if tunTask.IsNotification() {
		_ = p.webhookSendSyncHandler.AddTask(
			ctx,
			p.db,
			p.businessType,
			record.ID,
			WebhookSendTaskArgs{
				EventTitle:     "TUN记录更新",
				EventType:      "TUN_UPDATE",
				EventTime:      time.Now().Format("2006-01-02 15:04:05"),
				EventBody:      "TUN记录更新事件",
				RecordType:     record.ServeType.String(),
				RecordOldValue: record.ServeURI,
			},
		)
	}

	return fmt.Sprintf("域名 %s 同步成功 (地址:%v)", record.SubDomain, record.ServeURI), nil
}

// LoadTask 加载 TUN同步 周期任务
func (p *TUNSyncTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.TUNRecord, isActive bool) error {

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
			SetName("TUN同步: " + record.SubDomain).
			SetHandlerName(TUNSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(TUNSyncTaskArgs{
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

// AddTask 添加 TUN同步 一次性任务
func (p *TUNSyncTaskProvider) AddTask(ctx context.Context, db *ent.Client, record *ent.TUNRecord) error {
	return nil
}
