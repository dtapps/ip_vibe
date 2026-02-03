package service

import (
	"context"
	"log/slog"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/sys_env"

	"google.golang.org/grpc"
)

// AgentService 代理服务
type AgentService struct {
	logger *slog.Logger // 日志记录器
	debug  bool         // 是否开启调试模式

	grpcClient  *grpc.ClientConn                  // gRPC 客户端
	agentClient grpc_generated.AgentServiceClient // gRPC 业务服务
}

// NewAgentService 创建代理服务
func NewAgentService(logger *slog.Logger, grpcClient *grpc.ClientConn) *AgentService {
	debug := sys_env.GvaEnv.DEBUG
	return &AgentService{
		logger: logger,
		debug:  debug,

		grpcClient:  grpcClient,
		agentClient: grpc_generated.NewAgentServiceClient(grpcClient),
	}
}

// StartStream 启动流式通信
func (s *AgentService) StartStream(ctx context.Context) {

	// 启动
	stream, err := s.agentClient.Stream(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "启动流式通信失败",
			slog.Any("err", err),
		)
		return
	}

	// 构建 ClientInfo
	clientInfo := &grpc_generated.ClientInfo{
		DeviceCode: sys_env.GvaEnv.DEVICE_CODE,
		Version:    global.GvaVersion,
		GitCommit:  global.GvaGitCommit,
		BuildTime:  global.GvaBuildTime,
		GoVersion:  global.GvaGoVersion,
		Os:         global.GvaOS,
		Arch:       global.GvaArch,
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
			slog.String("device_code", sys_env.GvaEnv.DEVICE_CODE),
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
				err := stream.Send(&grpc_generated.ClientMessage{
					Payload: &grpc_generated.ClientMessage_Info{Info: clientInfo},
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

	// 主循环：发送心跳
	for {
		select {
		case <-ctx.Done():
			s.logger.ErrorContext(ctx, "流上下文已取消，正在退出")
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
		}
	}
}
