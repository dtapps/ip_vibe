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
}

// NewCloudflareTunnelSyncTaskProvider 创建一个新的 Cloudflare隧道同步 提供者
func NewCloudflareTunnelSyncTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewCloudflareTunnelSyncTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewCloudflareTunnelSyncTaskProvider 创建一个新的 Cloudflare隧道同步 提供者
func NewCloudflareTunnelSyncTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *CloudflareTunnelSyncTaskProvider {
	cloudflareTunnelSync := syncs.NewCloudflareTunnelSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &CloudflareTunnelSyncTaskProvider{
		db:                   db,
		logDb:                logDb,
		logger:               logger.With(slog.String("handler", "cloudflare-tunnel-sync")),
		httpTransport:        httpTransport,
		cloudflareTunnelSync: cloudflareTunnelSync,
	}
}

// CloudflareTunnelSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	CloudflareTunnelSyncTaskName = "cloudflare_tunnel_sync"
)

// CloudflareTunnelSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type CloudflareTunnelSyncTaskArgs struct {
	ConfigID string `json:"config_id,omitempty"` // 平台配置序号
	TunnelID string `json:"tunnel_id,omitempty"` // 隧道序号
}

// Process 执行 Cloudflare隧道同步 逻辑
func (p *CloudflareTunnelSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params CloudflareTunnelSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.ConfigID == "" && params.TunnelID == "" {

		// 调用 同步全部账号Cloudflare隧道
		err := p.cloudflareTunnelSync.SyncAllCloudflareTunnels(ctx)
		if err != nil {
			return "", err
		}

		// 调用 同步全部账号Cloudflare隧道入站规则
		err = p.cloudflareTunnelSync.SyncAllCloudflareTunnelIngresses(ctx)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("同步成功"), nil
}

// LoadTask 加载 Cloudflare隧道同步 任务
func (p *CloudflareTunnelSyncTaskProvider) LoadTask(ctx context.Context, businessKey string) error {

	// 查找任务
	_, err := p.db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(cronjob.TaskTypeRecurring)).
		Where(cronjob.BusinessTypeEQ(cronjob.BusinessTypeSystem)).
		Where(cronjob.BusinessKeyEQ(businessKey)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = p.db.CronJob.Create().
			SetName("CloudFlare隧道同步: " + businessKey).
			SetHandlerName(CloudflareTunnelSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(CloudflareTunnelSyncTaskArgs{}))).
			SetTaskType(cronjob.TaskTypeRecurring).
			SetJobType(cronjob.JobTypeDuration).
			SetJobCategory(cronjob.JobCategorySystem).
			SetBusinessType(cronjob.BusinessTypeSystem).
			SetBusinessKey(businessKey).
			SetMinInterval(6).
			SetMaxInterval(8).
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
