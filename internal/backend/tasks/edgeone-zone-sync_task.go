package tasks

import (
	"context"
	"log/slog"
	"time"
	"vibe_ip/internal/backend/syncs"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
)

// EdgeOne站点同步任务
type EdgeOneZoneSyncTask struct {
	db              *ent.Client                   // 数据库客户端
	logDb           *ent_log.Client               // 日志数据库客户端
	cron            gocron.Scheduler              // 定时任务调度器
	job             gocron.Job                    // EdgeOne同步任务
	jobHelper       *gocron_helper.JobHelper      // 任务助手
	logger          *slog.Logger                  // 日志记录器
	httpTransport   *http_log.LoggingRoundTripper // HTTP日志记录器
	edgeOneZoneSync *syncs.EdgeOneZoneSync        // EdgeOne站点同步
	debug           bool                          // 是否开启调试模式
}

// 创建 EdgeOne站点同步任务 实例
func NewEdgeOneZoneSyncTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *EdgeOneZoneSyncTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	edgeOneZoneSync := syncs.NewEdgeOneZoneSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &EdgeOneZoneSyncTask{
		db:              db,
		logDb:           logDb,
		cron:            cron,
		jobHelper:       gocron_helper.NewJobHelper(cron, logger),
		logger:          logger,
		httpTransport:   httpTransport,
		edgeOneZoneSync: edgeOneZoneSync,
		debug:           debug,
	}
}

// Close 关闭服务，释放资源
func (s *EdgeOneZoneSyncTask) Close() {
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *EdgeOneZoneSyncTask) SetDebug(debug bool) *EdgeOneZoneSyncTask {
	s.debug = debug
	return s
}

// 自定义 logger
func (s *EdgeOneZoneSyncTask) setLogger(logger *slog.Logger) *EdgeOneZoneSyncTask {
	s.logger = logger
	return s
}

// 启动 EdgeOne站点同步任务
func (s *EdgeOneZoneSyncTask) StartJob(ctx context.Context) (err error) {
	// 判断是否存在 EdgeOne站点同步任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.EdgeOneZoneSyncTaskTime,
		time.Hour,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 创建周期任务
	s.job, err = s.cron.NewJob(
		jobDuration,
		gocron.NewTask(s.RunSyncTask), // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithName("EdgeOne Zone Sync"),                 // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// 运行EdgeOne同步任务
func (s *EdgeOneZoneSyncTask) RunSyncTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())
	logger.InfoContext(ctx, "开始执行")
	defer logger.InfoContext(ctx, "执行完成")

	// 调用 同步 全部账号 所有站点
	err := s.edgeOneZoneSync.SyncAllZones(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点",
			slog.Any("err", err),
		)
		return
	}

	// 调用 同步 全部账号 所有站点 所有源站组
	err = s.edgeOneZoneSync.SyncAllZoneOriginGroups(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点 所有源站组",
			slog.Any("err", err),
		)
		return
	}

	// 调用 同步 全部账号 所有站点 所有记录
	err = s.edgeOneZoneSync.SyncAllZoneRecords(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点 所有记录",
			slog.Any("err", err),
		)
		return
	}
}
