package task

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// PortForwardGRPCTask 端口转发任务
type PortForwardGRPCTask struct {
	cron   gocron.Scheduler       // 定时任务调度器
	job    gocron.Job             // 端口转发任务
	logger *slog.Logger           // 日志记录器
	config *PortForwardTaskConfig // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	clientTarget      string                           // gRPC 客户端目标地址
	clientOpts        []grpc.DialOption                // gRPC 客户端连接选项
	clientBaseMD      metadata.MD                      // gRPC 客户端基础 metadata
	client            *grpc.ClientConn                 // gRPC 客户端连接
	portForwardClient grpc_generated.OpenServiceClient // gRPC 业务服务客户端
}

// NewPortForwardGRPCTask 创建一个新的服务实例
func NewPortForwardGRPCTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent) *PortForwardGRPCTask {

	return &PortForwardGRPCTask{
		cron:         cron,
		logger:       logger.With("component", "port-forward-grpc"),
		service:      service,
		forwardAgent: forwardAgent,
		clientOpts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// Close 关闭服务，释放资源
func (s *PortForwardGRPCTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetConfig 设置任务配置
func (s *PortForwardGRPCTask) SetConfig(config *PortForwardTaskConfig) *PortForwardGRPCTask {
	s.config = config
	// 设置基础 metadata（相当于HTTP headers）
	s.clientBaseMD = metadata.Pairs(
		"x-version", config.Version,
		"x-git-commit", config.GitCommit,
		"x-build-time", config.BuildTime,
		"x-go-version", config.GoVersion,
		"x-os", config.OS,
		"x-arch", config.Arch,
	)
	return s
}

// SetBaseURL 设置基础URL
func (s *PortForwardGRPCTask) SetBaseURL(url string) *PortForwardGRPCTask {
	s.clientTarget = url
	return s
}

// 设置基本认证
func (s *PortForwardGRPCTask) SetBasicAuth(username, password string) *PortForwardGRPCTask {
	if username != "" && password != "" {
		s.clientOpts = append(
			s.clientOpts,
			grpc.WithPerRPCCredentials(&grpc_credential.BasicAuth{
				Username: username,
				Password: password,
			}),
		)
	}
	return s
}

// StartPeriodicTask 启动任务
func (s *PortForwardGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 连接gRPC服务器
	s.client, err = grpc.NewClient(s.clientTarget, s.clientOpts...)
	if err != nil {
		return fmt.Errorf("连接gRPC服务器失败: %w", err)
	}

	// 注册业务服务
	s.portForwardClient = grpc_generated.NewOpenServiceClient(s.client)

	// 每30分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationJob(30*time.Minute),
		gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("gRPC 端口转发"),                         // 任务名称
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, "gRPC 端口转发任务已启动",
		slog.String("interval", "30分钟"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *PortForwardGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	s.logger.InfoContext(ctx, "端口转发开始")

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设备数据失败",
			slog.String("device_code", s.config.DeviceCode),
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger := s.logger.With(
		slog.String("device_code", deviceData.DeviceCode),
		slog.String("device_name", deviceData.DeviceName),
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
			slog.Int("rule_count", len(forwardRules)),
		)

		// 同步端口转发规则统计信息
		stats := s.forwardAgent.Stats()
		for ruleID, stat := range stats {
			logger.InfoContext(ctx, "同步端口转发规则统计信息",
				slog.String("rule_id", ruleID),
				slog.Int64("active_conn", stat.ActiveConn),
				slog.Int64("total_conn", stat.TotalConn),
				slog.Int64("bytes_in", stat.BytesIn),
				slog.Int64("bytes_out", stat.BytesOut),
				slog.Int64("errors", stat.Errors),
			)
		}
	} else {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "端口转发同步完成",
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "端口转发同步完成")
	}
}

// getApi 请求
func (s *PortForwardGRPCTask) getApi(ctx context.Context, logger *slog.Logger, deviceData *service.GetDeviceDataResult) (list []*PortForwardTaskApiList, err error) {

	// 构建gRPC请求
	req := &grpc_generated.GetDevicePortForwardsRequest{
		DeviceCode: deviceData.DeviceCode,
	}

	// 调用gRPC服务
	resp, err := s.portForwardClient.GetDevicePortForwards(metadata.NewOutgoingContext(ctx, s.clientBaseMD), req)
	if err != nil {
		return nil, fmt.Errorf("gRPC调用失败: %w", err)
	}

	// 转换响应
	for _, rule := range resp.Rules {
		list = append(list, &PortForwardTaskApiList{
			ID:         cast.ToString(rule.Id),
			LocalPort:  cast.ToInt(rule.LocalPort),
			RemoteHost: rule.RemoteHost,
			RemotePort: cast.ToInt(rule.RemotePort),
			Protocols:  rule.Protocols,
		})
	}

	// 记录响应
	logger.InfoContext(ctx, "gRPC 端口转发响应",
		slog.Any("response", resp.String()),
	)

	return list, nil
}
