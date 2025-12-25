package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/reporter/service"

	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

// IPReportRestyTask IP上报任务
type IPReportRestyTask struct {
	cron   gocron.Scheduler    // 定时任务调度器
	job    gocron.Job          // IP上报任务
	logger *slog.Logger        // 日志记录器
	config *IPReportTaskConfig // 配置

	service *service.IPReportService // IP报告服务

	client *resty.Client // Resty HTTP 客户端
}

// NewIPReportRestyTask 创建一个新的服务实例
func NewIPReportRestyTask(cron gocron.Scheduler, logger *slog.Logger, service *service.IPReportService) *IPReportRestyTask {

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

	return &IPReportRestyTask{
		cron:    cron,
		logger:  logger.With("component", "ip-report-resty"),
		service: service,
		client:  client,
	}
}

// Close 关闭服务，释放资源
func (s *IPReportRestyTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *IPReportRestyTask) SetDebug(debug bool) *IPReportRestyTask {
	s.client.SetDebug(debug)
	return s
}

// SetConfig 设置任务配置
func (s *IPReportRestyTask) SetConfig(config *IPReportTaskConfig) *IPReportRestyTask {
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
func (s *IPReportRestyTask) SetBaseURL(url string) *IPReportRestyTask {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *IPReportRestyTask) SetBasicAuth(username, password string) *IPReportRestyTask {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

func (s *IPReportRestyTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("Resty IP上报"),                        // 任务名称
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, "Resty IP上报任务已启动",
		slog.String("interval", "8-12分钟随机"),
	)
	return nil
}

func (s *IPReportRestyTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	s.logger.InfoContext(ctx, "IP上报开始")

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设备数据失败",
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger := s.logger.With(
		slog.String("device_code", deviceData.DeviceCode),
		slog.String("device_name", deviceData.DeviceName),
		slog.String("ipv4_addr", deviceData.IPV4Addr),
		slog.String("ipv6_addr", deviceData.IPV6Addr),
	)

	// 上传数据
	err = s.postApi(ctx, logger, deviceData)
	if err != nil {
		logger.ErrorContext(ctx, "上传数据失败",
			slog.Any("err", err),
		)
		return
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "IP上报成功",
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "IP上报成功")
	}
}

// postApi 请求
func (s *IPReportRestyTask) postApi(ctx context.Context, logger *slog.Logger, params *service.GetDeviceDataResult) (err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetBody(params).
		SetPathParams(map[string]string{
			"device_code": params.DeviceCode,
		}).
		Post("api/open/device/{device_code}/ip-report")
	if err != nil {
		return fmt.Errorf("Resty请求失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, "Resty IP上报响应",
		slog.Any("response", resp.String()),
	)

	return nil
}
