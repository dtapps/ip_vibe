package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/helper/domain_helper"
	"vibe_ip/internal/helper/gocron_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.dtapp.net/library/contrib/http_log"
	"golang.org/x/sync/errgroup"
)

// 域名监控任务
type DomainMonitorTask struct {
	db            *ent.Client                // 业务数据库客户端
	logDb         *ent_log.Client            // 日志数据库客户端
	cron          gocron.Scheduler           // 定时任务调度器
	job           gocron.Job                 // 域名监控任务
	taskName      string                     // 任务名称
	jobHelper     *gocron_helper.JobHelper   // 任务助手
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
}

// 创建实例
func NewDomainMonitorTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) *DomainMonitorTask {
	return &DomainMonitorTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		taskName:      "域名监控",
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// Close 关闭服务，释放资源
func (s *DomainMonitorTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// 自定义 logger
func (s *DomainMonitorTask) SetLogger(logger *slog.Logger) *DomainMonitorTask {
	s.logger = logger
	return s
}

// 计算下次执行时间
func (m *DomainMonitorTask) CalculateNextRun(domainMonitor *ent.DomainMonitor, now time.Time) time.Time {
	var nextRun time.Time

	// 基于成功时间计算
	if domainMonitor.LastSuccessAt.IsZero() {
		// 如果从未成功过，立即运行
		nextRun = now
	} else {
		// 下次运行时间 = 上次成功时间 + 检查间隔
		nextRun = domainMonitor.LastSuccessAt.Add(time.Duration(domainMonitor.CheckInterval) * time.Second)
	}

	// 如果是失败状态就使用上次失败时间（覆盖上面的计算）
	if domainMonitor.CurrentStatus == domainmonitor.CurrentStatusFailure {
		if !domainMonitor.LastFailureAt.IsZero() {
			nextRun = domainMonitor.LastFailureAt.Add(time.Duration(domainMonitor.CheckInterval) * time.Second)
		}
	}

	return nextRun
}

// StartPeriodicTask 启动周期任务
func (s *DomainMonitorTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.job != nil {
		return nil
	}
	// 创建周期任务
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			30*time.Second, // 最小间隔 30 秒
			60*time.Second, // 最大间隔 60 秒
		),
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
		slog.String("执行间隔", "30-60秒随机"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *DomainMonitorTask) RunPeriodicTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 分组记录
	var concurrentRecords, serialRecords []*ent.DomainMonitor

	// 查询全部域名监控
	allDomainMonitors, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IsActive(true)).
		All(ctx)
	if err == nil {
		now := time.Now()
		for _, domainMonitor := range allDomainMonitors {
			// 计算下次执行时间
			nextRun := s.CalculateNextRun(domainMonitor, now)

			// 如果已超过或等于下次运行时间，则执行任务
			if now.After(nextRun) || now.Equal(nextRun) {
				// 按策略分组
				concurrentRecords = append(concurrentRecords, domainMonitor)
			}
		}

		logger.InfoContext(ctx, "准备处理",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					if err = s.executeDomainMonitorCheck(egCtx, logger.With(
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
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			if err = s.executeDomainMonitorCheck(ctx, logger.With(
				slog.String("处理方式", "串行"),
			), item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("错误信息", err),
				)
			}

		}
	}

}

// 执行域名监控检查
func (s *DomainMonitorTask) executeDomainMonitorCheck(ctx context.Context, logger *slog.Logger, domainMonitor *ent.DomainMonitor) (err error) {

	// 日志
	logger = logger.With(
		slog.Int64("记录", domainMonitor.ID),
		slog.String("域名", domainMonitor.Domain),
		slog.String("协议", domainMonitor.CheckProtocol.String()),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// 执行检查
	domainCheck := domain_helper.NewDomainCheck(logger, s.httpTransport, s.debug)
	result := domainCheck.Check(ctx, domainMonitor.Domain, domainMonitor.CheckProtocol.String())

	// 历史记录创建
	historyCreator := s.logDb.DomainMonitorHistory.Create()
	historyCreator.SetMonitorID(domainMonitor.ID)
	historyCreator.SetCheckTime(time.Now())
	historyCreator.SetIsSuccess(result.IsSuccess)
	historyCreator.SetCheckDuration(int(result.Duration.Milliseconds()))
	historyCreator.SetStatusMessage(result.StatusMsg)
	historyCreator.SetHTTPStatus(result.HTTPStatus)
	historyCreator.SetCheckType(domainmonitorhistory.CheckType(result.CheckType))
	historyCreator.SetIPResolved(result.IPResolved)
	historyCreator.SetErrorLog(result.ErrorLog)

	// 填充 SSL 字段
	if domainMonitor.CheckProtocol == domainmonitor.CheckProtocolHTTPS {
		historyCreator.SetSslNotDate(result.SSLNotDate)
		historyCreator.SetSslExpiryDate(result.SSLExpiryDate)
		historyCreator.SetSslCommonName(result.SSLCommonName)
		historyCreator.SetSslIssuerName(result.SSLIssuerName)
		historyCreator.SetSslSignatureAlg(result.SSLSignatureAlg)
		historyCreator.SetSslSanList(result.SSLSANList)
		historyCreator.SetSslCertificateChain(result.SSLCertificateChain)
		historyCreator.SetSslSerialNumber(result.SSLSerialNumber)
		historyCreator.SetSslVersion(result.SSLVersion)
	}

	_, err = historyCreator.Save(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "保存历史记录失败",
			slog.Any("err", err),
		)
		return err
	}

	// 更新域名监控状态
	updaterOne := s.db.DomainMonitor.UpdateOne(domainMonitor)
	if result.IsSuccess {
		updaterOne.SetLastSuccessAt(time.Now())
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusSuccess)
	} else {
		updaterOne.SetLastFailureAt(time.Now())
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusFailure)
	}
	err = updaterOne.Exec(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "更新域名监控状态失败",
			slog.Any("错误信息", err),
		)
		return err
	}

	return nil
}

// AddDisposableTask 添加一次性任务
func (s *DomainMonitorTask) AddDisposableTask(ctx context.Context, domainMonitor *ent.DomainMonitor) (err error) {

	// 任务编号
	jobID := uuid.New()

	// 任务名称
	jobName := fmt.Sprintf("%s-%v", s.taskName, domainMonitor.ID)

	// 检查是否已存在同名任务
	isExist := s.jobHelper.IsJob("", jobName)
	if isExist {
		return fmt.Errorf("任务 %s 已存在，无法重复添加", jobName)
	}

	// 不存在同名任务，添加一次任务
	_, err = s.cron.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(5*time.Second)), // 5秒后开始
		),
		gocron.NewTask(
			s.RunDisposableTask,
			ctx,
			s.jobHelper.SlogWith(s.logger, jobID.String(), jobName),
			domainMonitor,
		),
		gocron.WithSingletonMode(gocron.LimitModeWait), // 单例模式
		gocron.WithName(jobName),                       // 任务名称
		gocron.WithIdentifier(jobID),                   // 任务编号
		gocron.WithEventListeners(
			s.jobHelper.PanicDisposableListener(),
			s.jobHelper.BeforeDisposableListener(),
			s.jobHelper.AfterDisposableListener(),
		), // 事件监听
	)

	return err
}

// RunDisposableTask 运行一次性任务
func (s *DomainMonitorTask) RunDisposableTask(ctx context.Context, logger *slog.Logger, domainMonitor *ent.DomainMonitor) {
	err := s.executeDomainMonitorCheck(ctx, logger, domainMonitor)
	if err != nil {
		logger.ErrorContext(ctx, "处理失败",
			slog.Any("错误信息", err),
		)
	}
}
