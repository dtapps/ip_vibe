package grpc_service

import (
	"context"
	"log/slog"
	"maps"
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
	Stream   grpc_generated.DeviceService_StreamServer
}

// DeviceService 实现 gRPC 服务
type DeviceService struct {
	grpc_generated.UnimplementedDeviceServiceServer              // 嵌入未实现的服务
	logger                                          *slog.Logger // 日志记录器
	mu                                              sync.RWMutex
	clients                                         map[string]*ClientState // device_code|role -> ClientState

	// 新增：用于优雅关闭
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewDeviceService 创建一个新的 DeviceService 实例
func NewDeviceService(logger *slog.Logger) *DeviceService {
	ctx, cancel := context.WithCancel(context.Background())
	return &DeviceService{
		logger:         logger,
		clients:        make(map[string]*ClientState),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

// RegisterService 注册服务到 gRPC
func (s *DeviceService) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterDeviceServiceServer(server, s)
}

// makeClientKey 生成客户端键（device_code|role）
func makeClientKey(deviceCode, role string) string {
	return deviceCode + "|" + role
}

// GetClients 返回当前在线客户端（用于调试或管理）
func (s *DeviceService) GetClients() map[string]*ClientState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 深拷贝或只读视图，避免外部修改
	result := make(map[string]*ClientState, len(s.clients))
	maps.Copy(result, s.clients)
	return result
}

// CloseAllStreams 主动关闭所有客户端流（用于优雅关闭）
func (s *DeviceService) CloseAllStreams() {
	s.shutdownCancel()
}

// Stream 处理客户端流式连接
func (s *DeviceService) Stream(stream grpc_generated.DeviceService_StreamServer) error {

	// 客户端键（device_code|role）
	var clientKey string

	// 清理函数：当连接断开时移除客户端
	defer func() {
		if clientKey != "" {
			s.mu.Lock()
			delete(s.clients, clientKey)
			s.mu.Unlock()
			s.logger.Error("客户端 已断开",
				slog.String("client_key", clientKey),
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
					slog.String("client_key", clientKey),
					slog.String("err", err.Error()),
				)
				return err
			}

			switch p := msg.Payload.(type) {
			case *grpc_generated.ClientMessage_Info:
				info := p.Info

				s.logger.Info("客户端 发送 ClientInfo",
					slog.String("client_key", clientKey),
					slog.Any("info", info),
				)

				// 手动验证
				if info.DeviceCode == "" || info.Role == "" {
					return status.Error(codes.InvalidArgument, "device_code and role are required")
				}

				// 生成客户端键
				clientKey = makeClientKey(info.DeviceCode, info.Role)

				// 注册客户端
				s.mu.Lock()
				s.clients[clientKey] = &ClientState{
					Info:     info,
					LastSeen: time.Now(),
					Stream:   stream,
				}
				s.logger.Info("客户端 已连接",
					slog.String("client_key", clientKey),
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
					slog.String("client_key", clientKey),
				)

				if clientKey != "" {
					s.mu.Lock()
					if state, ok := s.clients[clientKey]; ok {
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

// RequestFullReportToDevice 向指定设备的所有角色实例发送“请重新上报完整信息”指令
func (s *DeviceService) RequestFullReportToDevice(deviceCode string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefix := deviceCode + "|"
	count := 0

	for key, state := range s.clients {
		if key == deviceCode || (len(key) > len(prefix) && key[:len(prefix)] == prefix) {
			err := state.Stream.Send(&grpc_generated.ServerMessage{
				Command: &grpc_generated.ServerMessage_RequestFullReport{
					RequestFullReport: &grpc_generated.RequestFullReport{},
				},
			})
			if err != nil {
				s.logger.Warn("向客户端发送 RequestFullReport 失败",
					slog.String("client_key", key),
					slog.Any("err", err),
				)
				continue
			}
			count++
		}
	}

	if count == 0 {
		return status.Errorf(codes.NotFound, "no online client found for device %q", deviceCode)
	}

	s.logger.Info("已向设备的所有客户端发送 RequestFullReport",
		slog.String("device_code", deviceCode),
		slog.Int("count", count),
	)
	return nil
}

// RequestFullReportToAll 向所有在线客户端发送“请重新上报完整信息”指令
func (s *DeviceService) RequestFullReportToAll() error {
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
