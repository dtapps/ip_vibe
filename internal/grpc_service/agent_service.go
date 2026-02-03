package grpc_service

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/grpc_generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientState 存储每个客户端的状态
type ClientState struct {
	Info     *grpc_generated.ClientInfo
	LastSeen time.Time
	Stream   grpc_generated.AgentService_StreamServer
}

// AgentService 实现 gRPC 服务
type AgentService struct {
	grpc_generated.UnimplementedAgentServiceServer              // 嵌入未实现的服务
	logger                                         *slog.Logger // 日志记录器
	mu                                             sync.RWMutex
	clients                                        map[string]*ClientState // device_code -> ClientState

	// 新增：用于优雅关闭
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewAgentService 创建一个新的 AgentService 实例
func NewAgentService(logger *slog.Logger) *AgentService {
	ctx, cancel := context.WithCancel(context.Background())
	return &AgentService{
		logger:         logger,
		clients:        make(map[string]*ClientState),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

// RegisterService 注册服务到 gRPC
func (s *AgentService) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterAgentServiceServer(server, s)
}

// GetClients 返回当前在线客户端（用于调试或管理）
func (s *AgentService) GetClients() map[string]*ClientState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 深拷贝或只读视图，避免外部修改
	result := make(map[string]*ClientState, len(s.clients))
	for k, v := range s.clients {
		result[k] = v
	}
	return result
}

// CloseAllStreams 主动关闭所有客户端流（用于优雅关闭）
func (s *AgentService) CloseAllStreams() {
	s.shutdownCancel()
}

// Stream 处理客户端流式连接
func (s *AgentService) Stream(stream grpc_generated.AgentService_StreamServer) error {

	// 设备代码
	var deviceCode string

	// 清理函数：当连接断开时移除客户端
	defer func() {
		if deviceCode != "" {
			s.mu.Lock()
			delete(s.clients, deviceCode)
			s.mu.Unlock()
			s.logger.Error("客户端 已断开",
				slog.String("device_code", deviceCode),
			)
		}
	}()

	// 合并上下文：支持 shutdown 和客户端断开
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// 监听全局关闭信号
	go func() {
		select {
		case <-s.shutdownCtx.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// 被主动关闭或客户端断开
			return ctx.Err()
		default:
			msg, err := stream.Recv()
			if err != nil {
				// 客户端断开或出错
				s.logger.Error("客户端 连接错误",
					slog.String("device_code", deviceCode),
					slog.String("err", err.Error()),
				)
				return err
			}

			switch p := msg.Payload.(type) {
			case *grpc_generated.ClientMessage_Info:
				info := p.Info

				s.logger.Info("客户端 发送 ClientInfo",
					slog.String("device_code", info.DeviceCode),
					slog.Any("info", info),
				)

				// 可选：手动验证（如果 protoc-gen-validate 未在生成时启用）
				if info.DeviceCode == "" {
					return status.Error(codes.InvalidArgument, "device_code is required")
				}

				deviceCode = info.DeviceCode

				// 注册客户端
				s.mu.Lock()
				s.clients[deviceCode] = &ClientState{
					Info:     info,
					LastSeen: time.Now(),
					Stream:   stream,
				}
				s.logger.Info("客户端 已连接",
					slog.String("device_code", deviceCode),
				)
				s.mu.Unlock()

				// 回复 Ack
				err = stream.Send(&grpc_generated.ServerMessage{
					Command: &grpc_generated.ServerMessage_Ack{
						Ack: &grpc_generated.Ack{},
					},
				})
				if err != nil {
					return err
				}

			case *grpc_generated.ClientMessage_Heartbeat:

				s.logger.Info("客户端 发送 ClientHeartbeat",
					slog.String("device_code", deviceCode),
				)

				if deviceCode != "" {
					s.mu.Lock()
					if state, ok := s.clients[deviceCode]; ok {
						state.LastSeen = time.Now()
					}
					s.mu.Unlock()
				}
				// 可选：回复心跳 ACK（通常不需要）

			default:
				// 未知消息类型，可忽略或报错
				continue
			}
		}
	}
}

// RequestFullReportToDevice 向指定设备发送“请重新上报完整信息”指令
func (s *AgentService) RequestFullReportToDevice(deviceCode string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.clients[deviceCode]
	if !ok {
		return status.Errorf(codes.NotFound, "device %q is not online", deviceCode)
	}

	err := state.Stream.Send(&grpc_generated.ServerMessage{
		Command: &grpc_generated.ServerMessage_RequestFullReport{
			RequestFullReport: &grpc_generated.RequestFullReport{},
		},
	})
	if err != nil {
		s.logger.Warn("向指定客户端发送 RequestFullReport 失败",
			slog.String("device_code", deviceCode),
			slog.Any("err", err),
		)
		return status.Errorf(codes.Internal, "failed to send request to device %q: %v", deviceCode, err)
	}

	s.logger.Info("已向指定客户端发送 RequestFullReport",
		slog.String("device_code", deviceCode),
	)
	return nil
}

// RequestFullReportToAll 向所有在线客户端发送“请重新上报完整信息”指令
func (s *AgentService) RequestFullReportToAll() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for deviceCode, state := range s.clients {
		err := state.Stream.Send(&grpc_generated.ServerMessage{
			Command: &grpc_generated.ServerMessage_RequestFullReport{
				RequestFullReport: &grpc_generated.RequestFullReport{},
			},
		})
		if err != nil {
			s.logger.Warn("向客户端发送 RequestFullReport 失败",
				slog.String("device_code", deviceCode),
				slog.Any("err", err),
			)
			continue
		}
		count++
	}
	s.logger.Info("已向所有客户端发送 RequestFullReport",
		slog.Int("count", count),
	)
	return nil
}
