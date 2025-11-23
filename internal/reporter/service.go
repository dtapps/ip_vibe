package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
)

type Service struct {
	cron gocron.Scheduler
	job  gocron.Job
	log  *slog.Logger

	debug  bool
	config struct {
		httpUrl  string
		username string
		password string
	}
}

// NewService 创建一个新的服务实例
func NewService(cron gocron.Scheduler) *Service {
	return &Service{
		cron: cron,
		log:  slog.With(),
	}
}

func (s *Service) EnableDebug() *Service {
	s.debug = true
	return s
}

func (s *Service) SetHttpUrl(url string) *Service {
	s.config.httpUrl = url
	return s
}

func (s *Service) SetBasicAuth(username, password string) *Service {
	s.config.username = username
	s.config.password = password
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
	checker := ip_helper.NewIPChecker()
	ipv4, err := checker.CheckIP(ctx, "v4")
	if err == nil {
		params.IPV4Addr = ipv4.IP
	}
	ipv6, err := checker.CheckIP(ctx, "v6")
	if err == nil {
		params.IPV6Addr = ipv6.IP
	}
	s.log.InfoContext(ctx, "IP上报开始",
		slog.String("device_code", params.DeviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
		slog.String("http_url", s.config.httpUrl),
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

	// 创建HTTP请求
	data, err := json.Marshal(params)
	if err != nil {
		s.log.ErrorContext(ctx, "JSON编码失败",
			slog.Any("err", err),
		)
		return err
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 创建POST请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.httpUrl, bytes.NewBuffer(data))
	if err != nil {
		s.log.ErrorContext(ctx, "创建请求失败",
			slog.Any("err", err),
		)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 判断是否有BasicAuth用户名和密码，如果有则设置认证
	if s.config.username != "" && s.config.password != "" {
		req.SetBasicAuth(s.config.username, s.config.password)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		s.log.ErrorContext(ctx, "发送请求失败",
			slog.Any("err", err),
		)
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.ErrorContext(ctx, "读取响应失败",
			slog.Any("err", err),
		)
		return err
	}

	// 打印响应状态和内容
	s.log.InfoContext(ctx, "响应状态码和内容",
		slog.Int("status_code", resp.StatusCode),
		slog.Any("response", body),
	)

	// 根据响应状态码判断请求是否成功
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.log.InfoContext(ctx, "请求成功",
			slog.Any("response", body),
		)
	} else {
		s.log.ErrorContext(ctx, "请求失败",
			slog.Any("response", body),
		)
	}

	return nil
}
