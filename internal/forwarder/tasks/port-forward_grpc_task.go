package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PortForwardGRPCTask 端口转发任务
type PortForwardGRPCTask struct {
	cron      gocron.Scheduler         // 定时任务调度器
	job       gocron.Job               // 端口转发任务
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

// NewPortForwardGRPCTask 创建端口转发实例
func NewPortForwardGRPCTask(ctx context.Context, cron gocron.Scheduler, logger *slog.Logger, forwardAgent *forward.Agent, grpcClient *grpc.ClientConn) (*PortForwardGRPCTask, error) {
	debug := sys_env.GvaEnv.DEBUG

	// 获取设备数据
	deviceData, err := core.GetDeviceData(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &PortForwardGRPCTask{
		cron:      cron,
		taskName:  "gRPC 端口转发",
		jobHelper: gocron_helper.NewJobHelper(cron, logger),
		logger:    logger,
		debug:     debug,

		forwardAgent: forwardAgent,

		deviceData: deviceData,

		grpcClient:        grpcClient,
		openServiceClient: grpc_generated.NewOpenServiceClient(grpcClient),
	}, nil
}

// Close 关闭端口转发实例，释放资源
func (s *PortForwardGRPCTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// StartPeriodicTask 启动周期任务
func (s *PortForwardGRPCTask) StartPeriodicTask(ctx context.Context) (err error) {
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
		getSetting.PortForwardReloadTime,
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
		slog.String("执行间隔", getSetting.PortForwardReloadTime),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *PortForwardGRPCTask) RunPeriodicTask() {

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

	// 获取数据
	list, err := s.getApi(ctx, logger)
	if err != nil {
		return
	}

	// 转换类型：types.GetPortForwardResponse -> forward.Rule
	var forwardRules []forward.Rule
	for _, item := range list {

		// 构建监听地址
		listenAddr := net.JoinHostPort("", fmt.Sprintf("%d", item.LocalPort))

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
func (s *PortForwardGRPCTask) getApi(ctx context.Context, logger *slog.Logger) (list []*PortForwardTaskApiList, err error) {

	// 构建gRPC请求
	req := &grpc_generated.GetDevicePortForwardsRequest{
		DeviceCode: s.deviceData.DeviceCode,
	}

	// 调用gRPC服务
	resp, err := s.openServiceClient.GetDevicePortForwards(metadata.NewOutgoingContext(ctx, s.clientBaseMD), req)
	if err != nil {
		logger.ErrorContext(ctx, s.taskName+" 调用失败",
			slog.Any("err", err),
		)
		return nil, fmt.Errorf(s.taskName+" 调用失败: %w", err)
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
	logger.InfoContext(ctx, s.taskName+" 响应",
		slog.Any("响应", resp.String()),
	)

	return list, nil
}
