package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/sslimport"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/utils"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
	"golang.org/x/sync/errgroup"
)

// SSL刷新任务
type SSLRefreshTask struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	job           gocron.Job                    // SSL刷新任务
	taskName      string                        // 任务名称
	jobHelper     *gocron_helper.JobHelper      // 任务助手
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 是否开启调试模式
}

// 创建实例
func NewSSLRefreshTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *SSLRefreshTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &SSLRefreshTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		taskName:      "SSL刷新",
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *SSLRefreshTask) SetDebug(debug bool) *SSLRefreshTask {
	s.debug = debug
	return s
}

// Close 关闭服务，释放资源
func (s *SSLRefreshTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// 自定义 logger
func (s *SSLRefreshTask) SetLogger(logger *slog.Logger) *SSLRefreshTask {
	s.logger = logger
	return s
}

// StartPeriodicTask 启动周期任务
func (s *SSLRefreshTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.SSLRefreshTaskTime,
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
		gocron.NewTask(s.RunPeriodicTask),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName(s.taskName),                          // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		), // 事件监听
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, s.taskName+"任务已启动",
		slog.String("执行间隔", fmt.Sprintf("%s小时", db_cache.GvaSetting.SSLRefreshTaskTime)),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *SSLRefreshTask) RunPeriodicTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 查询全部SSL记录
	allSSLRecords, err := s.db.SSLRecord.Query().
		Where(sslrecord.ProviderEQ(sslrecord.ProviderUnknown)).
		// 排序
		OrderByCreatedAtAsc().
		All(ctx)
	if err == nil {

		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(10) // 最多 10 个并发

		for _, item := range allSSLRecords {
			item := item // 创建副本以避免闭包问题
			eg.Go(func() error {

				if err = s.executeSSLRecordRefresh(egCtx, logger.With(
					slog.String("处理方式", "并发"),
				), item); err != nil {
					logger.ErrorContext(egCtx, "处理失败",
						slog.Any("错误信息", err),
					)
				}

				return nil
			})
		}

		if err = eg.Wait(); err != nil {
			logger.ErrorContext(ctx, "并发处理错误",
				slog.Any("错误信息", err),
			)
		}

	}

	// 查询全部SSL导入
	allSSLImports, err := s.db.SSLImport.Query().
		// 排序
		OrderByCreatedAtAsc().
		All(ctx)
	if err == nil {

		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(10) // 最多 10 个并发

		for _, item := range allSSLImports {
			item := item // 创建副本以避免闭包问题
			eg.Go(func() error {

				if err = s.executeSSLImportRefresh(egCtx, logger.With(
					slog.String("处理方式", "并发"),
				), item); err != nil {
					logger.ErrorContext(egCtx, "处理失败",
						slog.Any("错误信息", err),
					)
				}

				return nil
			})
		}

		if err = eg.Wait(); err != nil {
			logger.ErrorContext(ctx, "并发处理错误",
				slog.Any("错误信息", err),
			)
		}

	}

}

// 执行SSL记录刷新
func (s *SSLRefreshTask) executeSSLRecordRefresh(ctx context.Context, logger *slog.Logger, item *ent.SSLRecord) (err error) {

	// 日志
	logger = logger.With(
		slog.Int64("记录", item.ID),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// 解析证书
	parsedCert, err := utils.ParseCertificate([]byte(item.CertCertificate))
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 证书状态
	status := sslrecord.StatusUnknown
	if parsedCert.NotAfter.After(time.Now()) {
		status = sslrecord.StatusValid
	} else {
		status = sslrecord.StatusExpired
	}

	// 准备保存
	updateOne := s.db.SSLRecord.UpdateOneID(item.ID)

	updateOne.SetNotBefore(parsedCert.NotBefore)
	updateOne.SetNotAfter(parsedCert.NotAfter)
	updateOne.SetStatus(status)

	err = updateOne.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新记录失败: %w", err)
	}

	return nil
}

// 执行SSL导入刷新
func (s *SSLRefreshTask) executeSSLImportRefresh(ctx context.Context, logger *slog.Logger, item *ent.SSLImport) (err error) {

	// 日志
	logger = logger.With(
		slog.Int64("记录", item.ID),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// 解析证书
	parsedCert, err := utils.ParseCertificate([]byte(item.Certificate))
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 证书状态
	status := sslimport.StatusUnknown
	if parsedCert.NotAfter.After(time.Now()) {
		status = sslimport.StatusValid
	} else {
		status = sslimport.StatusExpired
	}

	// 准备保存
	updateOne := s.db.SSLImport.UpdateOneID(item.ID)

	updateOne.SetNotBefore(parsedCert.NotBefore)
	updateOne.SetNotAfter(parsedCert.NotAfter)
	updateOne.SetStatus(status)

	err = updateOne.Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新记录失败: %w", err)
	}

	return nil
}
