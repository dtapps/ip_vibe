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
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gostring"
	"go.dtapp.net/library/utils/gotime"
	"resty.dev/v3"
)

// StatsReportRestyTask 统计上报任务
type StatsReportRestyTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 统计上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	config    *PortForwardTaskConfig   // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	client *resty.Client // Resty HTTP 客户端
}

// NewStatsReportRestyTask 创建一个新的服务实例
func NewStatsReportRestyTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent) *StatsReportRestyTask {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 请求跟踪
	client.EnableTrace()

	// 设置默认超时
	client.SetTimeout(1 * time.Minute)

	// 重试机制
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &StatsReportRestyTask{
		cron:         cron,
		taskName:     "Resty 统计上报",
		jobHelper:    gocron_helper.NewJobHelper(cron, logger),
		logger:       logger.With("component", "stats-report-resty"),
		service:      service,
		forwardAgent: forwardAgent,
		client:       client,
	}
}

// Close 关闭服务，释放资源
func (s *StatsReportRestyTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetDebug 设置调试模式
func (s *StatsReportRestyTask) SetDebug(debug bool) *StatsReportRestyTask {
	s.client.SetDebug(debug)
	return s
}

// SetConfig 设置任务配置
func (s *StatsReportRestyTask) SetConfig(config *PortForwardTaskConfig) *StatsReportRestyTask {
	s.config = config
	// 设置请求头
	s.client.SetHeader("X-Version", config.Version)
	s.client.SetHeader("X-Git-Commit", config.GitCommit)
	s.client.SetHeader("X-Build-Time", config.BuildTime)
	s.client.SetHeader("X-Go-Version", config.GoVersion)
	s.client.SetHeader("X-OS", config.OS)
	s.client.SetHeader("X-Arch", config.Arch)
	return s
}

// SetBaseURL 设置基础URL
func (s *StatsReportRestyTask) SetBaseURL(url string) *StatsReportRestyTask {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *StatsReportRestyTask) SetBasicAuth(username, password string) *StatsReportRestyTask {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

// parseTaskTime 解析任务时间
func (s *StatsReportRestyTask) parseTaskTime(taskTime string, timeUnit time.Duration) (jobDuration gocron.JobDefinition, isStart bool, err error) {
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
func (s *StatsReportRestyTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 获取设置
	getSetting, err := s.getSetting(ctx, s.logger)
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
func (s *StatsReportRestyTask) RunPeriodicTask() {

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
func (s *StatsReportRestyTask) postApi(ctx context.Context, logger *slog.Logger, deviceCode string, portForwardID string, stats *forward.Stats) (err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetPathParams(map[string]string{
			"device_code":     deviceCode,
			"port_forward_id": portForwardID,
		}).
		SetBody(types.CreatePortForwardStatsRequest{
			DevicePortForwardsRequest: types.DevicePortForwardsRequest{
				DeviceCode: deviceCode,
			},
			PortForwardID: cast.ToInt64(portForwardID),
			StartTime:     gotime.SetCurrent(stats.StartTime).Timestamp(),
			ActiveConn:    stats.ActiveConn,
			TotalConn:     stats.TotalConn,
			BytesIn:       stats.BytesIn,
			BytesOut:      stats.BytesOut,
			UDPPacketsIn:  stats.UDPPacketsIn,
			UDPPacketsOut: stats.UDPPacketsOut,
			Errors:        stats.Errors,
		}).
		Post("api/open/device/{device_code}/port-forwards-stats/{port_forward_id}")
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

// getSetting 获取设置
func (s *StatsReportRestyTask) getSetting(ctx context.Context, logger *slog.Logger) (item *types.OpenGetSettingResponse, err error) {

	var response struct {
		Status int                          `json:"status"`
		Data   types.OpenGetSettingResponse `json:"data"`
	}

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetResult(&response).
		Get("api/open/setting")
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return nil, fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return &response.Data, nil
}
