package grpc_service

import (
	"io"
	"log/slog"
	"sync"
	"vibe_ip/internal/grpc_generated"

	"google.golang.org/grpc"
)

// RemoteLogService 实现 gRPC 服务接口
type RemoteLogService struct {
	grpc_generated.UnimplementedRemoteLogServiceServer                                                            // 嵌入未实现的服务
	logger                                             *slog.Logger                                               // 日志记录器
	mu                                                 sync.RWMutex                                               // 用于存放所有连上来的设备流，方便你以后根据 DeviceID 找连接
	streams                                            map[string]grpc_generated.RemoteLogService_LogTunnelServer // 所有连上来的设备流
}

// NewRemoteLogService 创建 gRPC 服务实例
func NewRemoteLogService(logger *slog.Logger) *RemoteLogService {
	return &RemoteLogService{
		logger:  logger,
		streams: make(map[string]grpc_generated.RemoteLogService_LogTunnelServer),
	}
}

// RegisterService 注册服务到 gRPC
func (s *RemoteLogService) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterRemoteLogServiceServer(server, s)
}

// LogTunnel 实现双向流逻辑
func (s *RemoteLogService) LogTunnel(stream grpc_generated.RemoteLogService_LogTunnelServer) error {
	var deviceID string

	for {
		// 1. 接收来自设备（客户端）的消息
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			s.logger.Error("读取设备隧道失败",
				slog.Any("err", err),
			)
			return err
		}

		// 2. 第一次收到消息时，记录设备 ID 并保存这个流
		if deviceID == "" {
			deviceID = resp.DeviceId
			s.mu.Lock()
			s.streams[deviceID] = stream
			s.mu.Unlock()
			s.logger.Info("设备已建立日志隧道",
				slog.String("deviceID", deviceID),
				slog.String("msg", "连接成功"),
			)
		}

		// 3. 处理设备返回的日志列表内容
		if len(resp.Lines) > 0 {
			s.logger.Info("收到远程日志列表",
				slog.String("deviceID", deviceID),
				slog.Int("count", len(resp.Lines)),
			)
		}

		if resp.ErrorMsg != "" {
			s.logger.Warn("设备返回错误",
				slog.String("deviceID", deviceID),
				slog.String("err", resp.ErrorMsg),
			)
		}
	}
}

// SendRequestToFile 给指定设备发送“读取文件名”的请求
func (s *RemoteLogService) SendRequestToFile(deviceID string, fileName string) error {
	s.mu.RLock()
	stream, ok := s.streams[deviceID]
	s.mu.RUnlock()

	if !ok {
		return any(nil).(error) // 简单示意：设备不在线
	}

	// 服务端主动下发文件名给设备
	return stream.Send(&grpc_generated.ServerRequest{
		FileName: fileName,
	})
}
