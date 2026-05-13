package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/backend/syncs"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/provider"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// CloudflareTunnelSyncTaskProvider 实现 provider.TaskProvider 接口
type CloudflareTunnelSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db                   *ent.Client                   // 数据库客户端
	logDb                *ent_log.Client               // 日志数据库客户端
	logger               *slog.Logger                  // 日志记录器
	httpTransport        *http_log.LoggingRoundTripper // HTTP日志记录器
	cloudflareTunnelSync *syncs.CloudflareTunnelSync   // Cloudflare隧道同步

	businessType cronjob.BusinessType // 业务类型
}

// NewCloudflareTunnelSyncTaskProvider 创建一个新的 Cloudflare隧道同步 提供者
func NewCloudflareTunnelSyncTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) provider.TaskProvider {
	return NewCloudflareTunnelSyncTaskProviderRaw(db, logDb, logger)
}

// NewCloudflareTunnelSyncTaskProvider 创建一个新的 Cloudflare隧道同步 提供者
func NewCloudflareTunnelSyncTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) *CloudflareTunnelSyncTaskProvider {
	cloudflareTunnelSync := syncs.NewCloudflareTunnelSync(
		db,
		logDb,
		logger,
	)
	return &CloudflareTunnelSyncTaskProvider{
		db:                   db,
		logDb:                logDb,
		logger:               logger.With(slog.String("handler", "cloudflare-tunnel-sync")),
		httpTransport:        taskHttpTransport,
		cloudflareTunnelSync: cloudflareTunnelSync,

		businessType: cronjob.BusinessTypeCloudflareTunnel,
	}
}

// CloudflareTunnelSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	CloudflareTunnelSyncTaskName = "cloudflare_tunnel_sync"
)

// CloudflareTunnelSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type CloudflareTunnelSyncTaskArgs struct {
	ConfigID int64 `json:"config_id,omitempty"` // 平台配置序号
	TunnelID int64 `json:"tunnel_id,omitempty"` // 隧道序号
}

// Process 执行 Cloudflare隧道同步 逻辑
func (p *CloudflareTunnelSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params CloudflareTunnelSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.ConfigID == 0 && params.TunnelID == 0 {

		// 同步 全部账号 全部隧道
		err := p.cloudflareTunnelSync.SyncAllCloudflareTunnels(ctx)
		if err != nil {
			return "", err
		}

		// 同步 全部账号 全部隧道入站规则
		err = p.cloudflareTunnelSync.SyncAllCloudflareTunnelIngresses(ctx)
		if err != nil {
			return "", err
		}

	} else if params.ConfigID != 0 && params.TunnelID == 0 {

		// 同步 某个账号 全部隧道
		err := p.cloudflareTunnelSync.SyncThisCloudflareTunnels(ctx, params.ConfigID)
		if err != nil {
			return "", err
		}

		// 同步 某个账号 全部隧道入站规则
		err = p.cloudflareTunnelSync.SyncThisCloudflareTunnelIngresses(ctx, params.ConfigID)
		if err != nil {
			return "", err
		}

	} else if params.ConfigID != 0 && params.TunnelID != 0 {

		// 同步 某个账号 某个隧道入站规则
		err := p.cloudflareTunnelSync.SyncThisCloudflareThisTunnelIngresses(ctx, params.ConfigID, params.TunnelID)
		if err != nil {
			return "", err
		}

	} else {
		return "", fmt.Errorf("参数错误")
	}

	return "同步成功", nil
}

// LoadTask 加载 Cloudflare隧道同步 周期任务
func (p *CloudflareTunnelSyncTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.PlatformConfig) error {

	var (
		taskType    = cronjob.TaskTypeRecurring
		jobType     = cronjob.JobTypeRandom
		jobCategory = cronjob.JobCategorySystem
	)

	// 查找任务
	_, err := db.CronJob.Query().
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
			SetName("CloudFlare隧道同步: " + record.Name).
			SetHandlerName(CloudflareTunnelSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(CloudflareTunnelSyncTaskArgs{
				ConfigID: record.ID,
			}))).
			SetTaskType(taskType).
			SetJobType(jobType).
			SetJobCategory(jobCategory).
			SetBusinessType(p.businessType).
			SetBusinessID(record.ID).
			SetMinInterval(1).
			SetMaxInterval(2).
			SetIntervalUnit(cronjob.IntervalUnitHours).
			SetStatus(cronjob.StatusRunning).
			SetIsSingleton(true). // 单例模式
			SetRunOnStart(false). // 启动时不立即运行
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建任务失败: %w", err)
		}

		return nil
	}

	return nil
}

// AddTask 添加 Cloudflare隧道同步 一次性任务
func (p *CloudflareTunnelSyncTaskProvider) AddTask(ctx context.Context, db *ent.Client, configID int64, tunnelID int64) error {

	var (
		taskType    = cronjob.TaskTypeOneOff
		jobType     = cronjob.JobTypeDuration
		jobCategory = cronjob.JobCategoryCustom
	)

	// 任务名称
	var name string

	// 业务键
	var businessID int64

	if configID != 0 && tunnelID == 0 {
		name = fmt.Sprintf("CloudFlare隧道同步: %d", configID)
		businessID = configID
	} else if configID != 0 && tunnelID != 0 {
		name = fmt.Sprintf("CloudFlare隧道同步: %d-%d", configID, tunnelID)
		businessID = tunnelID
	} else {
		return fmt.Errorf("参数错误")
	}

	// 创建任务
	err := db.CronJob.Create().
		SetName(name).
		SetHandlerName(CloudflareTunnelSyncTaskName).
		SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(CloudflareTunnelSyncTaskArgs{
			ConfigID: configID,
			TunnelID: tunnelID,
		}))).
		SetTaskType(taskType).
		SetJobType(jobType).
		SetJobCategory(jobCategory).
		SetBusinessType(p.businessType).
		SetBusinessID(businessID).
		SetIntervalValue(1).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(cronjob.StatusRunning).
		SetIsSingleton(true). // 单例模式
		SetRunOnStart(true).  // 启动时立即运行
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}
