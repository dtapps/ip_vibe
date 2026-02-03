package grpc_service

import (
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
}

// NewAgentService 创建一个新的 AgentService 实例
func NewAgentService(logger *slog.Logger) *AgentService {
	return &AgentService{
		logger:  logger,
		clients: make(map[string]*ClientState),
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

// Stream 处理客户端流式连接
func (s *AgentService) Stream(stream grpc_generated.AgentService_StreamServer) error {
	var deviceCode string

	// 清理函数：当连接断开时移除客户端
	defer func() {
		if deviceCode != "" {
			s.mu.Lock()
			delete(s.clients, deviceCode)
			s.mu.Unlock()
			s.logger.Error("Client %s disconnected", deviceCode)
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			// 客户端断开或出错
			return err
		}

		switch p := msg.Payload.(type) {
		case *grpc_generated.ClientMessage_Info:
			info := p.Info

			s.logger.Info("Received ClientInfo",
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

			s.logger.Info("Received ClientHeartbeat",
				slog.Any("heartbeat", p.Heartbeat),
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
