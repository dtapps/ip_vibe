package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/reporter/service"

	"github.com/go-co-op/gocron/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// IPReportGRPCTask IP上报任务
type IPReportGRPCTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // IP上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	logger    *slog.Logger             // 日志记录器
	config    *IPReportTaskConfig      // 配置

	service *service.IPReportService // IP报告服务

	clientBaseMD      metadata.MD                      // gRPC 客户端基础 metadata
	grpcClient        *grpc.ClientConn                 // gRPC 客户端
	openServiceClient grpc_generated.OpenServiceClient // gRPC 业务服务
}

// NewIPReportGRPCTask 创建IP上报实例
func NewIPReportGRPCTask(cron gocron.Scheduler, logger *slog.Logger, service *service.IPReportService, grpcClient *grpc.ClientConn) *IPReportGRPCTask {
	return &IPReportGRPCTask{
		cron:              cron,
		taskName:          "gRPC IP上报",
		jobHelper:         gocron_helper.NewJobHelper(cron, logger),
		logger:            logger.With("component", "ip-report-grpc"),
		service:           service,
		grpcClient:        grpcClient,
		openServiceClient: grpc_generated.NewOpenServiceClient(grpcClient),
	}
}

// Close 关闭IP上报实例，释放资源
func (s *IPReportGRPCTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
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

// StartPeriodicTask 启动周期任务
func (s *IPReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 每8到12分钟执行一次
	s.job, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			8*time.Minute,  // 最小间隔
			12*time.Minute, // 最大间隔
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
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, s.taskName+"任务已启动",
		slog.String("执行间隔", "8-12分钟随机"),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *IPReportGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 获取设备数据
	deviceData, err := s.service.GetDeviceData(ctx, s.config.DeviceCode, s.config.DeviceName)
	if err != nil {
		logger.ErrorContext(ctx, "获取设备数据失败",
			slog.Any("err", err),
		)
		return
	}

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", deviceData.DeviceCode),
		slog.String("设备名称", deviceData.DeviceName),
		slog.String("IPv4地址", deviceData.IPV4Addr),
		slog.String("IPv6地址", deviceData.IPV6Addr),
	)

	// 上传数据
	_ = s.postApi(ctx, logger, deviceData)
}

// postApi 请求
func (s *IPReportGRPCTask) postApi(ctx context.Context, logger *slog.Logger, deviceData *service.GetDeviceDataResult) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.ReportDeviceIPRequest{
		DeviceCode: deviceData.DeviceCode,
		DeviceName: deviceData.DeviceName,
		Ipv4Addr:   deviceData.IPV4Addr,
		Ipv6Addr:   deviceData.IPV6Addr,
	}

	// 调用gRPC服务
	resp, err := s.openServiceClient.ReportDeviceIP(metadata.NewOutgoingContext(ctx, s.clientBaseMD), req)
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return fmt.Errorf(s.taskName+" 调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return nil
}
