package forwarder

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

type Service struct {
	cron   gocron.Scheduler // 定时任务调度器
	job    gocron.Job       // IP上报任务
	log    *slog.Logger     // 日志记录器
	client *resty.Client    // Resty HTTP 客户端
}

// NewService 创建一个新的服务实例
func NewService(cron gocron.Scheduler) *Service {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 设置默认超时
	client.SetTimeout(30 * time.Second)

	// 设置自定义请求头
	client.SetHeader("X-Version", global.GvaVersion)
	client.SetHeader("X-Git-Commit", global.GvaGitCommit)
	client.SetHeader("X-Build-Time", global.GvaBuildTime)
	client.SetHeader("X-Go-Version", global.GvaGoVersion)

	return &Service{
		cron:   cron,
		log:    slog.With("component", "port_forwarder"),
		client: client,
	}
}

// SetDebug 设置调试模式
func (s *Service) SetDebug(debug bool) *Service {
	s.client.SetDebug(debug)
	return s
}

// SetBaseURL 设置基础URL
func (s *Service) SetBaseURL(url string) *Service {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *Service) SetBasicAuth(username, password string) *Service {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

func (s *Service) StartJob(ctx context.Context) (err error) {
	// 请求参数
	deviceCode := global.GvaEnv.DEVICE_CODE
	// 检查设备代码是否为空
	if deviceCode == "" {
		// 获取设备ID
		deviceCode, err = machineid.ID()
		if err != nil {
			s.log.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}
	s.log.InfoContext(ctx, "端口转发获取开始",
		slog.String("device_code", deviceCode),
	)
	response, err := s.httpApi(ctx, deviceCode)
	if err != nil {
		s.log.ErrorContext(ctx, "端口转发获取失败",
			slog.String("device_code", deviceCode),
			slog.Any("err", err),
		)
		return
	}
	startForwarding(response.Data)
	s.log.InfoContext(ctx, "端口转发获取成功",
		slog.String("device_code", deviceCode),
	)
	return
}

type HttpApiResponse struct {
	Status int                             `json:"status"`
	Data   []types.GetPortForwardsResponse `json:"data"`
}

func (s *Service) httpApi(ctx context.Context, deviceCode string) (response HttpApiResponse, err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"device_code": deviceCode,
		}).
		SetResult(&response).
		Get("api/open/device/{device_code}/port-forwards")
	if err != nil {
		return HttpApiResponse{}, err
	}

	// 打印响应状态和内容
	s.log.InfoContext(ctx, "响应状态码和内容",
		slog.Int("status_code", resp.StatusCode()),
		slog.Any("response", resp.String()),
	)

	if resp.IsError() {
		return HttpApiResponse{}, fmt.Errorf("响应状态码错误: %d", resp.StatusCode())
	}

	return response, nil
}
