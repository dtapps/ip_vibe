package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gotime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// StatsReportGRPCTask 统计上报任务
type StatsReportGRPCTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 统计上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	jobOne    sync.Once                // 确保只创建一次任务
	logger    *slog.Logger             // 日志记录器
	debug     bool                     // 是否开启调试模式

	forwardAgent *forward.Agent // 端口转发器

	deviceData *core.GetDeviceDataResult // 设备数据

	clientBaseMD      metadata.MD                      // gRPC 客户端基础 metadata
	grpcClient        *grpc.ClientConn                 // gRPC 客户端
	openServiceClient grpc_generated.OpenServiceClient // gRPC 业务服务
}

// NewStatsReportGRPCTask 创建统计上报实例
func NewStatsReportGRPCTask(ctx context.Context, cron gocron.Scheduler, logger *slog.Logger, forwardAgent *forward.Agent, grpcClient *grpc.ClientConn) (*StatsReportGRPCTask, error) {
	debug := sys_env.GvaEnv.DEBUG

	// 获取设备数据
	deviceData, err := core.GetDeviceData(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &StatsReportGRPCTask{
		cron:      cron,
		taskName:  "gRPC 统计上报",
		jobHelper: gocron_helper.NewJobHelper(cron, logger),
		logger:    logger,
		debug:     debug,

		forwardAgent: forwardAgent,

		deviceData: deviceData,

		grpcClient:        grpcClient,
		openServiceClient: grpc_generated.NewOpenServiceClient(grpcClient),
	}, nil
}

// Close 关闭统计上报实例，释放资源
func (s *StatsReportGRPCTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// StartPeriodicTask 启动周期任务
func (s *StatsReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
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
		getSetting.PortForwardReportTime,
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
			gocron.WithSingletonMode(gocron.LimitModeReschedule),                        // 单例模式
			gocron.WithStartAt(gocron.WithStartDateTime(time.Now().Add(1*time.Minute))), // 首次延迟1分钟
			gocron.WithName(s.taskName),                                                 // 任务名称
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
		slog.String("执行间隔", getSetting.PortForwardReportTime),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *StatsReportGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 检查 forwardAgent 是否初始化
	if s.forwardAgent == nil {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
		return
	}

	// 日志记录器
	logger = logger.With(
		slog.String("设备编号", s.deviceData.DeviceCode),
		slog.String("设备名称", s.deviceData.DeviceName),
	)

	// 端口转发规则统计信息
	stats := s.forwardAgent.Stats()
	for ruleID, stat := range stats {
		logger.InfoContext(ctx, "端口转发规则统计信息",
			slog.Time("启动时间", stat.StartTime),
			slog.String("规则ID", ruleID),
			slog.Int64("活跃连接数", stat.ActiveConn),
			slog.Int64("累计连接数", stat.TotalConn),
			slog.Int64("接收字节数", stat.BytesIn),
			slog.Int64("发送字节数", stat.BytesOut),
			slog.Int64("错误次数", stat.Errors),
		)
		// 请求
		_ = s.postApi(ctx, logger, ruleID, stat)
	}
}

// postApi 请求
func (s *StatsReportGRPCTask) postApi(ctx context.Context, logger *slog.Logger, port_forward_id string, stats *forward.Stats) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.CreatePortForwardStatsRequest{
		DeviceCode:    s.deviceData.DeviceCode,
		PortForwardId: cast.ToInt64(port_forward_id),
		StartTime:     gotime.SetCurrent(stats.StartTime).Timestamp(),
		ActiveConn:    stats.ActiveConn,
		TotalConn:     stats.TotalConn,
		BytesIn:       stats.BytesIn,
		BytesOut:      stats.BytesOut,
		UdpPacketsIn:  stats.UDPPacketsIn,
		UdpPacketsOut: stats.UDPPacketsOut,
		Errors:        stats.Errors,
	}

	// 调用gRPC服务
	resp, err := s.openServiceClient.CreatePortForwardStats(metadata.NewOutgoingContext(ctx, s.clientBaseMD), req)
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
