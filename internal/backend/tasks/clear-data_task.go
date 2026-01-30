package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/ent_log/httplog"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/ent_log/portforwardstats"
	"vibe_ip/internal/ent_log/reqlog"
	"vibe_ip/internal/ent_log/restylog"
	"vibe_ip/internal/ent_log/webhookhistory"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/utils/gotime"
)

// 清除数据任务
type ClearDataTask struct {
	db        *ent.Client              // 业务数据库客户端
	logDb     *ent_log.Client          // 日志数据库客户端
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 清除数据任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	logger    *slog.Logger             // 日志记录器
	debug     bool                     // 是否开启调试模式
}

// 创建实例
func NewClearDataTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger) *ClearDataTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &ClearDataTask{
		db:        db,
		logDb:     logDb,
		cron:      cron,
		taskName:  "清除数据",
		jobHelper: gocron_helper.NewJobHelper(cron, logger),
		logger:    logger,
		debug:     debug,
	}
}

// Close 关闭服务，释放资源
func (s *ClearDataTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetDebug 设置调试模式
func (s *ClearDataTask) SetDebug(debug bool) *ClearDataTask {
	s.debug = debug
	return s
}

// 自定义 logger
func (s *ClearDataTask) SetLogger(logger *slog.Logger) *ClearDataTask {
	s.logger = logger
	return s
}

// StartPeriodicTask 启动周期任务
func (s *ClearDataTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.job != nil {
		return nil
	}
	// 创建周期任务
	s.job, err = s.cron.NewJob(
		gocron.CronJob(
			"5 0 * * *", // 每天凌晨 00:05 执行
			false,
		),
		gocron.NewTask(s.RunPeriodicTask),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
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
		slog.String("执行间隔", "每天凌晨 00:05 执行"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *ClearDataTask) RunPeriodicTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	s.executeDomainMonitorCheck(ctx, logger)
}

// 执行清除数据检查
func (s *ClearDataTask) executeDomainMonitorCheck(ctx context.Context, logger *slog.Logger) (err error) {

	// 日志
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// HTTP请求日志
	var httpLogTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HTTPRequestLogRecordTime)).Time
	var restyLogTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HTTPRequestLogRecordTime)).Time
	var reqLogTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HTTPRequestLogRecordTime)).Time

	_, err = s.logDb.HttpLog.Delete().
		Where(httplog.CreatedAtLT(httpLogTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除HTTP日志失败: %w", err)
	}

	_, err = s.logDb.RestyLog.Delete().
		Where(restylog.CreatedAtLT(restyLogTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除Resty日志失败: %w", err)
	}

	_, err = s.logDb.ReqLog.Delete().
		Where(reqlog.CreatedAtLT(reqLogTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除Req日志失败: %w", err)
	}

	// 历史数据保留时间
	var domainMonitorHistoryTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HistoryDataRetainTime)).Time
	var operationLogTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HistoryDataRetainTime)).Time
	var webhookHistoryTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.HistoryDataRetainTime)).Time

	_, err = s.logDb.DomainMonitorHistory.Delete().
		Where(domainmonitorhistory.CreatedAtLT(domainMonitorHistoryTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除域名监控历史记录失败: %w", err)
	}

	_, err = s.logDb.OperationLog.Delete().
		Where(operationlog.CreatedAtLT(operationLogTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除操作日志失败: %w", err)
	}

	_, err = s.logDb.WebhookHistory.Delete().
		Where(webhookhistory.CreatedAtLT(webhookHistoryTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除Webhook历史记录失败: %w", err)
	}

	// 统计数据保留时间
	var deviceIpReportTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.StatisticDataRetainTime)).Time
	var portForwardStatsTime = gotime.Current().BeforeDay(int(db_cache.GvaSetting.StatisticDataRetainTime)).Time

	_, err = s.logDb.DeviceIpReport.Delete().
		Where(deviceipreport.CreatedAtLT(deviceIpReportTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除设备IP报告失败: %w", err)
	}

	_, err = s.logDb.PortForwardStats.Delete().
		Where(portforwardstats.CreatedAtLT(portForwardStatsTime)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除端口转发统计记录失败: %w", err)
	}

	return nil
}
