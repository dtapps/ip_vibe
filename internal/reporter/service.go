package reporter

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

type Service struct {
	cron   gocron.Scheduler // 定时任务调度器
	job    gocron.Job       // 任务
	log    *slog.Logger     // 日志记录器
	client *resty.Client    // Resty HTTP 客户端
}

// NewService 创建一个新的服务实例
func NewService(cron gocron.Scheduler) *Service {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 启用请求跟踪以获取延迟信息
	client.EnableTrace()

	return &Service{
		cron: cron,
		log:  slog.Default(),
		// log:    slog.With("component", "reporter"), // 日志记录器
		client: client,
	}
}

// EnableDebug 开启调试模式
func (s *Service) EnableDebug() *Service {
	s.client.EnableDebug()
	return s
}

// SetBaseURL 设置基础URL
func (s *Service) SetBaseURL(url string) *Service {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *Service) SetBasicAuth(username, password string) *Service {
	s.client.SetBasicAuth(username, password)
	return s
}

func (s *Service) StartJob(ctx context.Context) (err error) {
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
		gocron.WithTags("ip_report"),                         // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// 全局互斥锁
var globalMutex sync.Mutex

func (s *Service) runTask() {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	var err error
	ctx := context.Background()

	// 请求参数
	params := types.ReportDeviceIpRequest{
		DeviceCode: global.GvaEnv.DEVICE_CODE,
		DeviceName: global.GvaEnv.DEVICE_NAME,
	}
	// 检查设备代码是否为空
	if params.DeviceCode == "" {
		// 获取设备ID
		params.DeviceCode, err = machineid.ID()
		if err != nil {
			s.log.ErrorContext(ctx, "获取设备ID失败",
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
	s.log.InfoContext(ctx, "IP上报开始",
		slog.String("device_code", params.DeviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
	)
	err = s.httpApi(ctx, params)
	if err != nil {
		s.log.ErrorContext(ctx, "IP上报失败",
			slog.String("device_code", params.DeviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.Any("err", err),
		)
		return
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", params.DeviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", params.DeviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
		)
	}

}

func (s *Service) httpApi(ctx context.Context, params types.ReportDeviceIpRequest) (err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetBody(params).
		Post("")

	// 打印响应状态和内容
	s.log.InfoContext(ctx, "响应状态码和内容",
		slog.Int("status_code", resp.StatusCode()),
		slog.Any("response", resp.String()),
	)

	// 根据响应状态码判断请求是否成功
	if resp.IsError() {
		s.log.ErrorContext(ctx, "请求失败",
			slog.Any("response", resp.String()),
		)
	} else {
		s.log.InfoContext(ctx, "请求成功",
			slog.Any("response", resp.String()),
		)
	}

	return nil
}
