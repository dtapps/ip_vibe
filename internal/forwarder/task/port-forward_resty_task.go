package task

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"resty.dev/v3"
)

// PortForwardRestyTask 端口转发任务
type PortForwardRestyTask struct {
	cron   gocron.Scheduler       // 定时任务调度器
	job    gocron.Job             // 端口转发任务
	logger *slog.Logger           // 日志记录器
	config *PortForwardTaskConfig // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	client *resty.Client // Resty HTTP 客户端
}

// NewPortForwardRestyTask 创建一个新的服务实例
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
		logger:       logger.With("component", "port-forward-resty"),
		service:      service,
		forwardAgent: forwardAgent,
		client:       client,
	}
}

// Close 关闭服务，释放资源
func (s *PortForwardRestyTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
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

// StartPeriodicTask 启动任务
func (s *PortForwardRestyTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 每30分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationJob(30*time.Minute),
		gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("Resty 端口转发"),                        // 任务名称
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, "Resty 端口转发任务已启动",
		slog.String("interval", "30分钟"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *PortForwardRestyTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	s.logger.InfoContext(ctx, "端口转发开始")

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设备数据失败",
			slog.String("设备编号", s.config.DeviceCode),
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger := s.logger.With(
		slog.String("设备编号", deviceData.DeviceCode),
		slog.String("设备名称", deviceData.DeviceName),
	)

	// 获取数据
	list, err := s.getApi(ctx, logger, deviceData)
	if err != nil {
		logger.ErrorContext(ctx, "获取数据失败",
			slog.Any("err", err),
		)
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

	// 调用 forward.Agent 的热更新
	if s.forwardAgent != nil {

		// 同步端口转发规则
		s.forwardAgent.Sync(ctx, forwardRules)
		logger.InfoContext(ctx, "同步端口转发规则完成",
			slog.Int("规则数量", len(forwardRules)),
		)

		// 同步端口转发规则统计信息
		stats := s.forwardAgent.Stats()
		for ruleID, stat := range stats {
			logger.InfoContext(ctx, "同步端口转发规则统计信息",
				slog.Time("启动时间", stat.StartTime),
				slog.String("规则ID", ruleID),
				slog.Int64("活跃连接数", stat.ActiveConn),
				slog.Int64("累计连接数", stat.TotalConn),
				slog.Int64("接收字节数", stat.BytesIn),
				slog.Int64("发送字节数", stat.BytesOut),
				slog.Int64("错误次数", stat.Errors),
			)
		}
	} else {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "端口转发同步完成",
			slog.String("下一次运行时间", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "端口转发同步完成")
	}
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
		return nil, err
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
	logger.InfoContext(ctx, "Resty 端口转发响应",
		slog.Any("响应", resp.String()),
	)

	return list, nil
}
