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

// CloudFlare Tunnel同步任务
type CloudflareTunnelTask struct {
	db                   *ent.Client                   // 数据库客户端
	logDb                *ent_log.Client               // 日志数据库客户端
	cron                 gocron.Scheduler              // 定时任务调度器
	job                  gocron.Job                    // CloudFlare Tunnel同步任务
	jobHelper            *gocron_helper.JobHelper      // 任务助手
	logger               *slog.Logger                  // 日志记录器
	httpTransport        *http_log.LoggingRoundTripper // HTTP日志记录器
	cloudflareTunnelSync *syncs.CloudflareTunnelSync   // CloudFlare Tunnel同步
	debug                bool                          // 是否开启调试模式
}

// 创建 CloudFlare Tunnel同步任务 实例
func NewCloudflareTunnelTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *CloudflareTunnelTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	cloudflareTunnelSync := syncs.NewCloudflareTunnelSync(
		db,
		logDb,
		logger,
		httpTransport,
	)
	return &CloudflareTunnelTask{
		db:                   db,
		logDb:                logDb,
		cron:                 cron,
		jobHelper:            gocron_helper.NewJobHelper(cron, logger),
		logger:               logger,
		httpTransport:        httpTransport,
		cloudflareTunnelSync: cloudflareTunnelSync,
		debug:                debug,
	}
}

// Close 关闭服务，释放资源
func (s *CloudflareTunnelTask) Close() {
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *CloudflareTunnelTask) SetDebug(debug bool) *CloudflareTunnelTask {
	s.debug = debug
	return s
}

// 自定义 logger
func (s *CloudflareTunnelTask) setLogger(logger *slog.Logger) *CloudflareTunnelTask {
	s.logger = logger
	return s
}

// 启动CloudFlare Tunnel同步任务
func (s *CloudflareTunnelTask) StartJob(ctx context.Context) (err error) {
	// 判断是否存在CloudFlare Tunnel同步任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.CloudflareTunnelSyncTaskTime,
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
		gocron.WithName("CloudFlare Tunnel Sync"),            // 任务名称
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

// 运行CloudFlare Tunnel同步任务
func (s *CloudflareTunnelTask) RunTunTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())
	logger.InfoContext(ctx, "开始执行")
	defer logger.InfoContext(ctx, "执行完成")

	// 调用 同步全部账号Cloudflare隧道
	err := s.cloudflareTunnelSync.SyncAllCloudflareTunnels(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步全部账号Cloudflare隧道",
			slog.Any("err", err),
		)
		return
	}

	// 调用 同步全部账号Cloudflare隧道入站规则
	err = s.cloudflareTunnelSync.SyncAllCloudflareTunnelIngresses(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "同步全部账号Cloudflare隧道入站规则",
			slog.Any("err", err),
		)
		return
	}

}
