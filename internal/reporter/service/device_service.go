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

// AgentService 代理服务
type AgentService struct {
	logger *slog.Logger // 日志记录器

	role       string                    // 角色
	deviceData *core.GetDeviceDataResult // 设备数据

	grpcClient   *grpc.ClientConn                   // gRPC 客户端
	deviceClient grpc_generated.DeviceServiceClient // gRPC 业务服务
}

// NewAgentService 创建代理服务
func NewAgentService(ctx context.Context, logger *slog.Logger, grpcClient *grpc.ClientConn) (*AgentService, error) {

	// 获取设备数据
	deviceData, err := core.GetDeviceData(ctx, logger)
	if err != nil {
		return nil, err
	}

	return &AgentService{
		logger: logger,

		role:       "reporter",
		deviceData: deviceData,

		grpcClient:   grpcClient,
		deviceClient: grpc_generated.NewDeviceServiceClient(grpcClient),
	}, nil
}

// StartStream 启动流式通信
func (s *AgentService) StartStream(ctx context.Context) {

	// 启动
	stream, err := s.deviceClient.Stream(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "启动流式通信失败",
			slog.Any("err", err),
		)
		return
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
		StartTime:  timestamppb.New(global.GvaStartTime),
		GetTime:    timestamppb.New(time.Now()),
	}

	// 发送 ClientInfo
	infoMsg := &grpc_generated.ClientMessage{
		Payload: &grpc_generated.ClientMessage_Info{
			Info: clientInfo,
		},
	}

	// 发送 ClientInfo
	err = stream.Send(infoMsg)
	if err != nil {
		s.logger.ErrorContext(ctx, "发送 ClientInfo 失败",
			slog.Any("err", err),
		)
		return
	}

	// 等待 Ack
	resp, err := stream.Recv()
	if err != nil {
		s.logger.ErrorContext(ctx, "接收 Ack 失败",
			slog.Any("err", err),
		)
		return
	}

	if _, ok := resp.GetCommand().(*grpc_generated.ServerMessage_Ack); !ok {
		s.logger.ErrorContext(ctx, "接收 Ack 失败",
			slog.Any("err", err),
		)
		return
	} else {
		s.logger.InfoContext(ctx, "成功注册",
			slog.String("device_code", s.deviceData.DeviceCode),
		)
	}

	// 启动 goroutine 监听服务端指令
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				// 流已关闭，退出
				return
			}

			switch cmd := msg.GetCommand().(type) {
			case *grpc_generated.ServerMessage_RequestFullReport:
				s.logger.InfoContext(ctx, "收到服务端指令：重新上报完整信息")
				// 重发 ClientInfo
				clientInfo.GetTime = timestamppb.New(time.Now())
				err := stream.Send(&grpc_generated.ClientMessage{
					Payload: &grpc_generated.ClientMessage_Info{
						Info: clientInfo,
					},
				})
				if err != nil {
					s.logger.ErrorContext(ctx, "重发 ClientInfo 失败", slog.Any("err", err))
					return
				}
				s.logger.DebugContext(ctx, "ClientInfo 重发成功")
			default:
				s.logger.WarnContext(ctx, "收到未知服务端指令", slog.Any("command", cmd))
			}
		}
	}()

	// 启动心跳（每 30 秒）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 启动全量上报（每 10 分钟）
	fullReportTicker := time.NewTicker(10 * time.Minute)
	defer fullReportTicker.Stop()

	// 主循环：发送心跳 + 定期重发 ClientInfo
	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "流上下文已取消，正在退出")
			return
		case <-ticker.C:
			heartbeatMsg := &grpc_generated.ClientMessage{
				Payload: &grpc_generated.ClientMessage_Heartbeat{
					Heartbeat: &grpc_generated.Heartbeat{},
				},
			}
			if err := stream.Send(heartbeatMsg); err != nil {
				s.logger.ErrorContext(ctx, "发送心跳失败",
					slog.Any("err", err),
				)
				return
			}
			s.logger.InfoContext(ctx, "心跳发送成功")

		case <-fullReportTicker.C:
			// 重新发送 ClientInfo
			clientInfo.GetTime = timestamppb.New(time.Now())
			infoMsg := &grpc_generated.ClientMessage{
				Payload: &grpc_generated.ClientMessage_Info{
					Info: clientInfo,
				},
			}
			if err := stream.Send(infoMsg); err != nil {
				s.logger.ErrorContext(ctx, "定期重发 ClientInfo 失败",
					slog.Any("err", err),
				)
				return
			}
			s.logger.DebugContext(ctx, "ClientInfo 定期重发成功")
		}
	}
}
