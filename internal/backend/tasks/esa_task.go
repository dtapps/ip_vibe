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

// ESA同步任务
type ESATask struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	job           gocron.Job                    // ESA同步任务
	jobHelper     *gocron_helper.JobHelper      // 任务助手
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	esaSync       *syncs.ESASync                // ESA同步
	debug         bool                          // 是否开启调试模式
}

// 创建 ESA同步任务 实例
func NewESATask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *ESATask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	esaSync := syncs.NewESASync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &ESATask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		esaSync:       esaSync,
		debug:         debug,
	}
}

// Close 关闭服务，释放资源
func (s *ESATask) Close() {
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *ESATask) SetDebug(debug bool) *ESATask {
	s.debug = debug
	return s
}

// 自定义 logger
func (s *ESATask) setLogger(logger *slog.Logger) *ESATask {
	s.logger = logger
	return s
}

// 启动ESA同步任务
func (s *ESATask) StartJob(ctx context.Context) (err error) {
	// 判断是否存在ESA同步任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.ESASyncTaskTime,
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
		gocron.NewTask(s.RunTunTask), // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithName("ESA Sync"),                          // 任务名称
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

// 运行ESA同步任务
func (s *ESATask) RunTunTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())
	logger.InfoContext(ctx, "开始执行")
	defer logger.InfoContext(ctx, "执行完成")

	// 调用 同步 全部账号 所有站点
	err := s.esaSync.SyncAllSites(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点",
			slog.Any("err", err),
		)
		return
	}

	// 调用 同步 全部账号 所有站点 所有源地址池
	err = s.esaSync.SyncAllSiteOriginPools(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点 所有源地址池",
			slog.Any("err", err),
		)
		return
	}

	// 调用 同步 全部账号 所有站点 所有记录
	err = s.esaSync.SyncAllSiteRecords(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步 全部账号 所有站点 所有记录",
			slog.Any("err", err),
		)
		return
	}
}
