package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-co-op/gocron/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// IPReportGRPCTask IP上报任务
type IPReportGRPCTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // IP上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	debug     bool                     // 是否开启调试模式

	deviceData *core.GetDeviceDataResult // 设备数据
	ipData     *core.GetIPDataResult     // IP数据

	clientBaseMD      metadata.MD                      // gRPC 客户端基础 metadata
	grpcClient        *grpc.ClientConn                 // gRPC 客户端
	openServiceClient grpc_generated.OpenServiceClient // gRPC 业务服务
}

// NewIPReportGRPCTask 创建IP上报实例
func NewIPReportGRPCTask(ctx context.Context, cron gocron.Scheduler, logger *slog.Logger, grpcClient *grpc.ClientConn) (task *IPReportGRPCTask, err error) {
	debug := sys_env.GvaEnv.DEBUG

	// 获取设备数据
	deviceData, err := core.GetDeviceData(ctx, logger)
	if err != nil {
		return nil, err
	}

	// 获取IP数据
	ipData, err := core.GetIPData(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &IPReportGRPCTask{
		cron:      cron,
		taskName:  "gRPC IP上报",
		jobHelper: gocron_helper.NewJobHelper(cron, logger),
		logger:    logger,
		debug:     debug,

		deviceData: deviceData,
		ipData:     ipData,

		grpcClient:        grpcClient,
		openServiceClient: grpc_generated.NewOpenServiceClient(grpcClient),
	}, nil
}

// Close 关闭IP上报实例，释放资源
func (s *IPReportGRPCTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// StartPeriodicTask 启动周期任务
func (s *IPReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 获取设置
	getSetting, err := s.openServiceClient.GetSetting(ctx, &emptypb.Empty{})
	if err != nil {
		s.logger.ErrorContext(ctx, "获取设置",
			slog.Any("err", err),
		)
		return err
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		getSetting.IPReportIntervalTime,
		time.Minute,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 确保只创建一次任务
	s.jobOne.Do(func() {
		// 创建周期任务
		s.job, err = s.cron.NewJob(
			jobDuration,
			gocron.NewTask(s.RunPeriodicTask), // 执行任务
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
		slog.String("执行间隔", getSetting.IPReportIntervalTime),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *IPReportGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", s.deviceData.DeviceCode),
		slog.String("设备名称", s.deviceData.DeviceName),
		slog.String("IPV4地址", s.ipData.IPV4Addr),
		slog.String("IPV6地址", s.ipData.IPV6Addr),
	)

	// 上传数据
	_ = s.postApi(ctx, logger)
}

// postApi 请求
func (s *IPReportGRPCTask) postApi(ctx context.Context, logger *slog.Logger) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.ReportDeviceIPRequest{
		DeviceCode: s.deviceData.DeviceCode,
		DeviceName: s.deviceData.DeviceName,
		Ipv4Addr:   s.ipData.IPV4Addr,
		Ipv6Addr:   s.ipData.IPV6Addr,
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
