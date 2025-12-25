package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

type IpReportService struct {
	cron   gocron.Scheduler // 定时任务调度器
	job    gocron.Job       // IP上报任务
	client *resty.Client    // Resty HTTP 客户端
	logger *slog.Logger     // 日志记录器
}

// NewIpReportService 创建一个新的服务实例
func NewIpReportService(cron gocron.Scheduler) *IpReportService {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 请求跟踪
	client.EnableTrace()

	// 设置默认超时
	client.SetTimeout(30 * time.Second)

	// 设置请求头
	client.SetHeader("X-Version", global.GvaVersion)
	client.SetHeader("X-Git-Commit", global.GvaGitCommit)
	client.SetHeader("X-Build-Time", global.GvaBuildTime)
	client.SetHeader("X-Go-Version", global.GvaGoVersion)
	client.SetHeader("X-OS", global.GvaOS)
	client.SetHeader("X-Arch", global.GvaArch)

	// 重试机制
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &IpReportService{
		cron:   cron,
		client: client,
		logger: slog.With("component", "ip-report"),
	}
}

// Close 关闭服务，释放资源
func (s *IpReportService) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *IpReportService) SetDebug(debug bool) *IpReportService {
	s.client.SetDebug(debug)
	return s
}

// SetBaseURL 设置基础URL
func (s *IpReportService) SetBaseURL(url string) *IpReportService {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *IpReportService) SetBasicAuth(username, password string) *IpReportService {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

func (s *IpReportService) StartJob(ctx context.Context) (err error) {
	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(s.runTask),                            // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("IP上报"),                              // 任务名称
		gocron.WithTags("ip-report"),                         // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// httpApi 发送IP上报请求
func (s *IpReportService) httpApi(ctx context.Context, deviceCode string, params types.DeviceReportDeviceIpRequest) (err error) {

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
		return err
	}

	// 打印响应状态和内容
	s.logger.InfoContext(ctx, "响应状态码和内容",
		slog.Int("status_code", resp.StatusCode()),
		slog.Any("response", resp.String()),
	)

	if resp.IsError() {
		return fmt.Errorf("响应状态码错误: %d", resp.StatusCode())
	}

	return nil
}

func (s *IpReportService) runTask() {

	var err error
	ctx := context.Background()

	// 请求参数
	deviceCode := global.GvaEnv.DEVICE_CODE
	params := types.DeviceReportDeviceIpRequest{
		DeviceName: global.GvaEnv.DEVICE_NAME,
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
	err = s.httpApi(ctx, deviceCode, params)
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
