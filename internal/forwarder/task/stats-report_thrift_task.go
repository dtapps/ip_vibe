package task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/thrift_generated/open"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gostring"
	"go.dtapp.net/library/utils/gotime"
)

// StatsReportThriftTask  统计上报任务
type StatsReportThriftTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 统计上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	config    *PortForwardTaskConfig   // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	clientBase        *open.Base              // Thrift 客户端基础
	thriftClient      *thrift.TStandardClient // Thrift 客户端
	openServiceClient *open.OpenServiceClient // Thrift 业务服务
}

// NewStatsReportThriftTask  创建统计上报实例
func NewStatsReportThriftTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent, thriftClient *thrift.TStandardClient, clientBase *open.Base) *StatsReportThriftTask {
	return &StatsReportThriftTask{
		cron:              cron,
		taskName:          "Thrift 统计上报",
		jobHelper:         gocron_helper.NewJobHelper(cron, logger),
		logger:            logger.With("component", "stats-report-thrift"),
		service:           service,
		forwardAgent:      forwardAgent,
		clientBase:        clientBase,
		thriftClient:      thriftClient,
		openServiceClient: open.NewOpenServiceClient(thriftClient),
	}
}

// Close 关闭统计上报实例，释放资源
func (s *StatsReportThriftTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetConfig 设置任务配置
func (s *StatsReportThriftTask) SetConfig(config *PortForwardTaskConfig) *StatsReportThriftTask {
	s.config = config
	return s
}

// parseTaskTime 解析任务时间
func (s *StatsReportThriftTask) parseTaskTime(taskTime string, timeUnit time.Duration) (jobDuration gocron.JobDefinition, isStart bool, err error) {
	if taskTime == "" {
		return nil, false, fmt.Errorf("时间未配置")
	}
	if strings.Contains(taskTime, "-") {
		// 解析时间范围
		times := gostring.Split(taskTime, "-")
		if len(times) != 2 {
			return nil, false, fmt.Errorf("时间格式错误：%s", taskTime)
		}
		jobDuration = gocron.DurationRandomJob(
			time.Duration(cast.ToInt64(times[0]))*timeUnit, // 最小间隔
			time.Duration(cast.ToInt64(times[1]))*timeUnit, // 最大间隔
		)
	} else {
		jobDuration = gocron.DurationJob(
			time.Duration(cast.ToInt64(taskTime)) * timeUnit,
		)
	}
	isStart = true
	return
}

// StartPeriodicTask 启动周期任务
func (s *StatsReportThriftTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 获取设置
	getSetting, err := s.openServiceClient.GetSetting(ctx, s.clientBase)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设置",
			slog.Any("err", err),
		)
		return err
	}
	// 解析设置的时间
	jobDuration, isStart, err := s.parseTaskTime(
		getSetting.PortForwardReportTime,
		time.Minute,
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
	// 确保只创建一次任务
	s.jobOne.Do(func() {
		// 创建周期任务
		s.job, err = s.cron.NewJob(
			jobDuration,
			gocron.NewTask(s.RunPeriodicTask), // 执行任务
			gocron.WithSingletonMode(gocron.LimitModeReschedule),                        // 单例模式
			gocron.WithStartAt(gocron.WithStartDateTime(time.Now().Add(1*time.Minute))), // 首次延迟1分钟
			gocron.WithName(s.taskName),                                                 // 任务名称
			gocron.WithEventListeners(
				s.jobHelper.PanicPeriodicListener(),
				s.jobHelper.BeforePeriodicListener(),
				s.jobHelper.AfterPeriodicListener(),
			), // 事件监听
		)
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "创建定时任务失败",
			slog.Any("err", err),
		)
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, s.taskName+"任务已启动",
		slog.String("执行间隔", getSetting.PortForwardReportTime),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *StatsReportThriftTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 检查 forwardAgent 是否初始化
	if s.forwardAgent == nil {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
		return
	}

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		logger.ErrorContext(ctx, "获取设备数据失败",
			slog.String("设备编号", s.config.DeviceCode),
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", deviceData.DeviceCode),
		slog.String("设备名称", deviceData.DeviceName),
	)

	// 端口转发规则统计信息
	stats := s.forwardAgent.Stats()
	for ruleID, stat := range stats {
		logger.InfoContext(ctx, "端口转发规则统计信息",
			slog.Time("启动时间", stat.StartTime),
			slog.String("规则ID", ruleID),
			slog.Int64("活跃连接数", stat.ActiveConn),
			slog.Int64("累计连接数", stat.TotalConn),
			slog.Int64("接收字节数", stat.BytesIn),
			slog.Int64("发送字节数", stat.BytesOut),
			slog.Int64("错误次数", stat.Errors),
		)
		// 请求
		_ = s.postApi(ctx, logger, deviceData.DeviceCode, ruleID, stat)
	}
}

// postApi 请求
func (s *StatsReportThriftTask) postApi(ctx context.Context, logger *slog.Logger, deviceCode string, port_forward_id string, stats *forward.Stats) (err error) {

	// 构建Thrift请求
	req := &open.CreatePortForwardStatsRequest{
		DeviceCode:    deviceCode,
		PortForwardID: cast.ToInt64(port_forward_id),
		StartTime:     gotime.SetCurrent(stats.StartTime).Timestamp(),
		ActiveConn:    &stats.ActiveConn,
		TotalConn:     &stats.TotalConn,
		BytesIn:       &stats.BytesIn,
		BytesOut:      &stats.BytesOut,
		UDPPacketsIn:  &stats.UDPPacketsIn,
		UDPPacketsOut: &stats.UDPPacketsOut,
		Errors:        &stats.Errors,
	}

	// 调用Thrift服务
	resp, err := s.openServiceClient.CreatePortForwardStats(ctx, req, s.clientBase)
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return nil
}
