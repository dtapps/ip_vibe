package service

import (
	"context"
	"log/slog"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_generated"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DeviceService 设备服务
type DeviceService struct {
	logger *slog.Logger // 日志记录器

	role       string                    // 角色
	deviceData *core.GetDeviceDataResult // 设备数据
	startTime  time.Time                 // 启动时间

	grpcClient   *grpc.ClientConn                   // gRPC 客户端
	deviceClient grpc_generated.DeviceServiceClient // gRPC 业务服务
}

// NewDeviceService 创建设备服务
func NewDeviceService(ctx context.Context, logger *slog.Logger, grpcClient *grpc.ClientConn) (*DeviceService, error) {

	// 获取设备数据
	deviceData, err := core.GetDeviceData(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &DeviceService{
		logger: logger,

		role:       "forwarder",
		deviceData: deviceData,
		startTime:  time.Now().UTC(),

		grpcClient:   grpcClient,
		deviceClient: grpc_generated.NewDeviceServiceClient(grpcClient),
	}, nil
}

// Run 开始运行服务
func (s *DeviceService) Run(ctx context.Context) {

	// 初始重连等待时间
	retryInterval := 1 * time.Second
	maxRetryInterval := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "DeviceService 停止运行")
			return
		default:
			// 尝试执行流逻辑
			err := s.executeStream(ctx)

			// 如果连接正常返回（没有错误），说明是 context 取消，直接重置并继续/退出
			if err == nil {
				retryInterval = 1 * time.Second
				continue
			}

			s.logger.ErrorContext(ctx, "流连接断开，准备重连",
				slog.Duration("after", retryInterval),
				slog.Any("err", err),
			)

			// 等待重连
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
				// 每次重连失败，增加等待时间，最高 30s
				retryInterval *= 2
				if retryInterval > maxRetryInterval {
					retryInterval = maxRetryInterval
				}
			}
		}
	}
}

// executeStream 处理单次连接
func (s *DeviceService) executeStream(ctx context.Context) error {

	// 启动
	stream, err := s.deviceClient.Stream(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "流式通信失败",
			slog.Any("err", err),
		)
		return err
	}

	// 构建 ClientInfo
	clientInfo := &grpc_generated.ClientInfo{
		DeviceCode: s.deviceData.DeviceCode,
		DeviceName: s.deviceData.DeviceName,
		Role:       s.role,
		Version:    global.GvaVersion,
		GitCommit:  global.GvaGitCommit,
		BuildTime:  global.GvaBuildTime,
		GoVersion:  global.GvaGoVersion,
		Os:         global.GvaOS,
		Arch:       global.GvaArch,
		RunTime:    timestamppb.New(global.GvaStartTime),
		StartTime:  timestamppb.New(s.startTime),
		GetTime:    timestamppb.Now(),
	}

	// 准备 ClientInfo 信息
	infoMsg := &grpc_generated.ClientMessage{
		Payload: &grpc_generated.ClientMessage_Info{
			Info: clientInfo,
		},
	}

	// 发送 ClientInfo 信息
	err = stream.Send(infoMsg)
	if err != nil {
		s.logger.ErrorContext(ctx, "发送 ClientInfo 信息 失败",
			slog.Any("err", err),
		)
		return err
	}

	// 等待 Ack 消息
	resp, err := stream.Recv()
	if err != nil {
		s.logger.ErrorContext(ctx, "等待 Ack 消息 失败",
			slog.Any("err", err),
		)
		return err
	}

	// 接收 Ack 消息
	_, ok := resp.GetCommand().(*grpc_generated.ServerMessage_Ack)
	if !ok {
		s.logger.ErrorContext(ctx, "接收 Ack 消息 失败",
			slog.Any("err", err),
		)
		return err
	}

	// 捕获异步协程的错误
	errChan := make(chan error, 1)

	// 启动监听协程
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				// 流已关闭，退出
				errChan <- err
				s.logger.ErrorContext(ctx, "流已关闭，退出",
					slog.Any("err", err),
				)
				return
			}

			switch cmd := msg.GetCommand().(type) {
			case *grpc_generated.ServerMessage_RequestFullReport:
				// 重发 ClientInfo 信息
				clientInfo.GetTime = timestamppb.Now()
				err := stream.Send(&grpc_generated.ClientMessage{
					Payload: &grpc_generated.ClientMessage_Info{
						Info: clientInfo,
					},
				})
				if err != nil {
					errChan <- err
					s.logger.ErrorContext(ctx, "重发 ClientInfo 信息 失败",
						slog.Any("err", err),
					)
					return
				}
				s.logger.DebugContext(ctx, "重发 ClientInfo 信息 成功")
			default:
				s.logger.WarnContext(ctx, "收到未知服务端指令",
					slog.Any("command", cmd),
				)
			}
		}
	}()

	// 启动心跳（每 30 秒）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 启动全量上报（每 30 分钟）
	fullReportTicker := time.NewTicker(30 * time.Minute)
	defer fullReportTicker.Stop()

	// 主循环：发送心跳 + 定期重发 ClientInfo 信息
	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "流上下文已取消，正在退出",
				slog.Any("err", ctx.Err()),
			)
			return ctx.Err()
		case err := <-errChan:
			// 接收协程出错了（服务端断开）
			s.logger.InfoContext(ctx, "接收协程出错了（服务端断开）",
				slog.Any("err", err),
			)
			return err
		case <-ticker.C:
			heartbeatMsg := &grpc_generated.ClientMessage{
				Payload: &grpc_generated.ClientMessage_Heartbeat{
					Heartbeat: &grpc_generated.Heartbeat{},
				},
			}
			err = stream.Send(heartbeatMsg)
			if err != nil {
				s.logger.ErrorContext(ctx, "发送心跳 失败",
					slog.Any("err", err),
				)
				return err
			}
			s.logger.DebugContext(ctx, "心跳发送 成功")

		case <-fullReportTicker.C:
			// 重新发送 ClientInfo 信息
			clientInfo.GetTime = timestamppb.Now()
			infoMsg := &grpc_generated.ClientMessage{
				Payload: &grpc_generated.ClientMessage_Info{
					Info: clientInfo,
				},
			}
			err = stream.Send(infoMsg)
			if err != nil {
				s.logger.ErrorContext(ctx, "定期重发 ClientInfo 信息 失败",
					slog.Any("err", err),
				)
				return err
			}
			s.logger.DebugContext(ctx, "定期重发 ClientInfo 信息成功")
		}
	}
}
