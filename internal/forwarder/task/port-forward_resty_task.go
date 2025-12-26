package task

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"resty.dev/v3"
)

// PortForwardRestyTask 端口转发任务
type PortForwardRestyTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 端口转发任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	config    *PortForwardTaskConfig   // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	client *resty.Client // Resty HTTP 客户端
}

// NewPortForwardRestyTask 创建端口转发实例
func NewPortForwardRestyTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent) *PortForwardRestyTask {

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

	return &PortForwardRestyTask{
		cron:         cron,
		taskName:     "Resty 端口转发",
		jobHelper:    gocron_helper.NewJobHelper(cron, logger),
		logger:       logger.With("component", "port-forward-resty"),
		service:      service,
		forwardAgent: forwardAgent,
		client:       client,
	}
}

// Close 关闭端口转发实例，释放资源
func (s *PortForwardRestyTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// SetDebug 设置调试模式
func (s *PortForwardRestyTask) SetDebug(debug bool) *PortForwardRestyTask {
	s.client.SetDebug(debug)
	return s
}

// SetConfig 设置任务配置
func (s *PortForwardRestyTask) SetConfig(config *PortForwardTaskConfig) *PortForwardRestyTask {
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
func (s *PortForwardRestyTask) SetBaseURL(url string) *PortForwardRestyTask {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *PortForwardRestyTask) SetBasicAuth(username, password string) *PortForwardRestyTask {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

// StartPeriodicTask 启动周期任务
func (s *PortForwardRestyTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 确保只创建一次任务
	s.jobOne.Do(func() {
		// 创建周期任务
		s.job, err = s.cron.NewJob(
			gocron.DurationJob(
				30*time.Minute, // 每30分钟执行一次
			),
			gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
			gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
			gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
			gocron.WithName(s.taskName),                          // 任务名称
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
		slog.String("interval", "30分钟"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *PortForwardRestyTask) RunPeriodicTask() {

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

	// 获取数据
	list, err := s.getApi(ctx, logger, deviceData)
	if err != nil {
		return
	}

	// 转换类型：types.GetPortForwardResponse -> forward.Rule
	var forwardRules []forward.Rule
	for _, item := range list {

		// 构建监听地址
		listenAddr := ":" + string(rune(item.LocalPort))

		// 更准确的构建方式：
		listenAddr = net.JoinHostPort("", fmt.Sprintf("%d", item.LocalPort))

		// 构建目标地址
		targetAddr := net.JoinHostPort(item.RemoteHost, fmt.Sprintf("%d", item.RemotePort))

		rule := forward.Rule{
			ID:         item.ID,    // 唯一标识
			ListenAddr: listenAddr, // 本地监听地址
			TargetAddr: targetAddr, // 目标地址
			Protocols:  item.Protocols,
		}
		forwardRules = append(forwardRules, rule)
	}

	// 同步端口转发规则，热更新
	s.forwardAgent.Sync(ctx, forwardRules)
	logger.InfoContext(ctx, "同步端口转发规则，热更新",
		slog.Int("规则数量", len(forwardRules)),
	)
}

// getApi 请求
func (s *PortForwardRestyTask) getApi(ctx context.Context, logger *slog.Logger, deviceData *service.GetDeviceDataResult) (list []*PortForwardTaskApiList, err error) {

	var response struct {
		Status int                            `json:"status"`
		Data   []types.GetPortForwardResponse `json:"data"`
	}

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetPathParams(map[string]string{
			"device_code": deviceData.DeviceCode,
		}).
		SetResult(&response).
		Get("api/open/device/{device_code}/port-forwards")
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return nil, fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 转换响应
	for _, rule := range response.Data {
		list = append(list, &PortForwardTaskApiList{
			ID:         cast.ToString(rule.ID),
			LocalPort:  rule.LocalPort,
			RemoteHost: rule.RemoteHost,
			RemotePort: rule.RemotePort,
			Protocols:  rule.Protocols,
		})
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return list, nil
}
