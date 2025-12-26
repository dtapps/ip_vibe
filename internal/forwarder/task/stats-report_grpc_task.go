package task

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/gocron_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// StatsReportGRPCTask 统计上报任务
type StatsReportGRPCTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 统计上报任务
	taskName  string                   // 任务名称
	jobHelper *gocron_helper.JobHelper // 任务助手
	logger    *slog.Logger             // 日志记录器
	config    *PortForwardTaskConfig   // 配置

	service      *service.PortForwardServer // 端口转发服务
	forwardAgent *forward.Agent             // 端口转发器

	clientBaseMD      metadata.MD                      // gRPC 客户端基础 metadata
	grpcClient        *grpc.ClientConn                 // gRPC 客户端
	openServiceClient grpc_generated.OpenServiceClient // gRPC 业务服务
}

// NewStatsReportGRPCTask 创建统计上报实例
func NewStatsReportGRPCTask(cron gocron.Scheduler, logger *slog.Logger, service *service.PortForwardServer, forwardAgent *forward.Agent, grpcClient *grpc.ClientConn) *StatsReportGRPCTask {
	return &StatsReportGRPCTask{
		cron:              cron,
		taskName:          "gRPC 统计上报",
		jobHelper:         gocron_helper.NewJobHelper(cron, logger),
		logger:            logger.With("component", "stats-report-grpc"),
		service:           service,
		forwardAgent:      forwardAgent,
		grpcClient:        grpcClient,
		openServiceClient: grpc_generated.NewOpenServiceClient(grpcClient),
	}
}

// Close 关闭统计上报实例，释放资源
func (s *StatsReportGRPCTask) Close() {
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetConfig 设置任务配置
func (s *StatsReportGRPCTask) SetConfig(config *PortForwardTaskConfig) *StatsReportGRPCTask {
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
func (s *StatsReportGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
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
func (s *StatsReportGRPCTask) RunPeriodicTask() {

	// 创建上下文
	ctx := context.Background()

	s.logger.InfoContext(ctx, "统计上报开始")

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

	// 调用 forward.Agent 的统计
	if s.forwardAgent != nil {
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
			_ = s.postApi(ctx, logger, deviceData.DeviceCode, ruleID, stat)
		}
	} else {
		logger.ErrorContext(ctx, "forwardAgent 未初始化")
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "统计上报同步完成",
			slog.String("下一次运行时间", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "统计上报同步完成")
	}
}

// postApi 请求
func (s *StatsReportGRPCTask) postApi(ctx context.Context, logger *slog.Logger, deviceCode string, port_forward_id string, stats *forward.Stats) (err error) {

	// 构建gRPC请求
	req := &grpc_generated.CreatePortForwardStatsRequest{
		DeviceCode:    deviceCode,
		PortForwardId: cast.ToInt64(port_forward_id),
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
		return fmt.Errorf("gRPC调用失败: %w", err)
	}

	// 记录响应
	logger.InfoContext(ctx, "gRPC 端口转发响应",
		slog.Any("响应", resp.String()),
	)

	return nil
}
