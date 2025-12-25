package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/reporter/service"
	"vibe_ip/internal/types"

	"github.com/go-co-op/gocron/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// IPReportGRPCTask IP上报任务
type IPReportGRPCTask struct {
	cron   gocron.Scheduler    // 定时任务调度器
	job    gocron.Job          // IP上报任务
	logger *slog.Logger        // 日志记录器
	config *IPReportTaskConfig // 配置

	service *service.IPReportService // IP报告服务

	clientTarget   string                           // gRPC 客户端目标地址
	clientOpts     []grpc.DialOption                // gRPC 客户端连接选项
	clientBaseMD   metadata.MD                      // gRPC 客户端基础 metadata
	client         *grpc.ClientConn                 // gRPC 客户端连接
	ipReportClient grpc_generated.OpenServiceClient // gRPC 业务服务客户端
}

// NewIPReportGRPCTask 创建一个新的服务实例
func NewIPReportGRPCTask(cron gocron.Scheduler, logger *slog.Logger, service *service.IPReportService) *IPReportGRPCTask {
	return &IPReportGRPCTask{
		cron:    cron,
		logger:  logger.With("component", "ip-report-grpc"),
		service: service,
		clientOpts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// Close 关闭服务，释放资源
func (s *IPReportGRPCTask) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetConfig 设置任务配置
func (s *IPReportGRPCTask) SetConfig(config *IPReportTaskConfig) *IPReportGRPCTask {
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
func (s *IPReportGRPCTask) SetBaseURL(url string) *IPReportGRPCTask {
	s.clientTarget = url
	return s
}

// 设置基本认证
func (s *IPReportGRPCTask) SetBasicAuth(username, password string) *IPReportGRPCTask {
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
func (s *IPReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 连接gRPC服务器
	s.client, err = grpc.NewClient(s.clientTarget, s.clientOpts...)
	if err != nil {
		return fmt.Errorf("连接gRPC服务器失败: %w", err)
	}

	// 注册业务服务
	s.ipReportClient = grpc_generated.NewOpenServiceClient(s.client)

	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
		),
		gocron.NewTask(s.RunPeriodicTask),                    // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("gRPC IP上报"),                         // 任务名称
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, "IP上报gRPC任务已启动",
		slog.String("interval", "8-12分钟随机"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *IPReportGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 请求参数
	params, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设备数据失败",
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger := s.logger.With(
		slog.String("device_code", params.DeviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
	)

	logger.InfoContext(ctx, "IP上报开始")
	err = s.reportIP(ctx, logger, params.DeviceCode, params)
	if err != nil {
		logger.ErrorContext(ctx, "IP上报失败",
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

// reportIP 发送IP上报请求
func (s *IPReportGRPCTask) reportIP(ctx context.Context, logger *slog.Logger, deviceCode string, params *types.DeviceReportDeviceIpRequest) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.ReportDeviceIPRequest{
		DeviceCode: deviceCode,
		DeviceName: params.DeviceName,
		Ipv4Addr:   params.IPV4Addr,
		Ipv6Addr:   params.IPV6Addr,
	}

	// 调用gRPC服务
	resp, err := s.ipReportClient.ReportDeviceIP(metadata.NewOutgoingContext(ctx, s.clientBaseMD), req)
	if err != nil {
		return fmt.Errorf("gRPC调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, "gRPC IP上报响应",
		slog.Any("response", resp.String()),
	)

	return nil
}
