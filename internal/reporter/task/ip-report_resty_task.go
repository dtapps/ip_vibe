package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

// IPReportRestyTask IP上报任务
type IPReportRestyTask struct {
	cron   gocron.Scheduler    // 定时任务调度器
	job    gocron.Job          // IP上报任务
	logger *slog.Logger        // 日志记录器
	config *IPReportTaskConfig // 配置

	client *resty.Client // Resty HTTP 客户端
}

// NewIPReportRestyTask 创建一个新的服务实例
func NewIPReportRestyTask(cron gocron.Scheduler) *IPReportRestyTask {

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
		cron:   cron,
		logger: slog.With("component", "ip-report"),
		client: client,
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

	var err error
	ctx := context.Background()

	// 请求参数
	deviceCode := s.config.DeviceCode
	params := types.DeviceReportDeviceIpRequest{
		DeviceName: s.config.DeviceName,
	}
	// 检查设备代码是否为空
	if deviceCode == "" {
		// 获取设备ID
		deviceCode, err = machineid.ID()
		if err != nil {
			s.logger.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}

	// 获取设备IP
	iPHelper := ip_helper.NewIPHelper()
	ipv4, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV4)
	if err == nil {
		params.IPV4Addr = ipv4.IP
	}
	ipv6, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV6)
	if err == nil {
		params.IPV6Addr = ipv6.IP
	}
	s.logger.InfoContext(ctx, "IP上报开始",
		slog.String("device_code", deviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
	)
	err = s.reportIP(ctx, deviceCode, params)
	if err != nil {
		s.logger.ErrorContext(ctx, "IP上报失败",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.Any("err", err),
		)
		return
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		s.logger.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		s.logger.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
		)
	}
}

// reportIP 发送IP上报请求
func (s *IPReportRestyTask) reportIP(ctx context.Context, deviceCode string, params types.DeviceReportDeviceIpRequest) (err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetBody(params).
		SetPathParams(map[string]string{
			"device_code": deviceCode,
		}).
		Post("api/open/device/{device_code}/ip-report")
	if err != nil {
		return fmt.Errorf("Resty请求失败: %w", err)
	}

	// 记录响应
	s.logger.InfoContext(ctx, "Resty IP上报响应",
		slog.Any("response", resp.String()),
	)

	return nil
}
