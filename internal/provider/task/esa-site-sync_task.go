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

// ESASiteSyncTaskProvider 实现 provider.TaskProvider 接口
type ESASiteSyncTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	esaSiteSync   *syncs.ESASiteSync            // ESA站点同步
}

// NewESASiteSyncTaskProvider 创建一个新的 ESA站点同步 提供者
func NewESASiteSyncTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewESASiteSyncTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewESASiteSyncTaskProvider 创建一个新的 ESA站点同步 提供者
func NewESASiteSyncTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *ESASiteSyncTaskProvider {
	esaSiteSync := syncs.NewESASiteSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &ESASiteSyncTaskProvider{
		db:            db,
		logDb:         logDb,
		logger:        logger.With(slog.String("handler", "esa-site-sync")),
		httpTransport: httpTransport,
		esaSiteSync:   esaSiteSync,
	}
}

// ESASiteSyncTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	ESASiteSyncTaskName = "esa_site_sync"
)

// ESASiteSyncTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type ESASiteSyncTaskArgs struct {
	ConfigID string `json:"config_id,omitempty"` // 平台配置序号
	SiteID   string `json:"site_id,omitempty"`   // 站点序号
	PoolID   string `json:"pool_id,omitempty"`   // 源地址池序号
}

// Process 执行 ESA站点同步 逻辑
func (p *ESASiteSyncTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params ESASiteSyncTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.ConfigID == "" && params.SiteID == "" && params.PoolID == "" {

		// 调用 同步 全部账号 所有站点
		err := p.esaSiteSync.SyncAllSites(ctx)
		if err != nil {
			return "", err
		}

		// 调用 同步 全部账号 所有站点 所有源地址池
		err = p.esaSiteSync.SyncAllSiteOriginPools(ctx)
		if err != nil {
			return "", err
		}

		// 调用 同步 全部账号 所有站点 所有记录
		err = p.esaSiteSync.SyncAllSiteRecords(ctx)
		if err != nil {
			return "", err
		}

	}

	return fmt.Sprintf("同步成功"), nil
}

// LoadTask 加载 ESA站点同步 任务
func (p *ESASiteSyncTaskProvider) LoadTask(ctx context.Context, businessKey string) error {

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
			SetName("ESA站点同步: " + businessKey).
			SetHandlerName(ESASiteSyncTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(ESASiteSyncTaskArgs{}))).
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
