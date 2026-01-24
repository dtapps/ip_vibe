package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/pkg/utils"

	"google.golang.org/grpc"
)

// RemoteLogService 远程日志服务
type RemoteLogService struct {
	logger *slog.Logger // 日志记录器

	grpcClient      *grpc.ClientConn                      // gRPC 客户端
	remoteLogClient grpc_generated.RemoteLogServiceClient // gRPC 业务服务
}

func NewRemoteLogService(logger *slog.Logger, grpcClient *grpc.ClientConn) *RemoteLogService {
	return &RemoteLogService{
		logger:          logger.With("component", "remote-log-service"),
		grpcClient:      grpcClient,
		remoteLogClient: grpc_generated.NewRemoteLogServiceClient(grpcClient),
	}
}

// StartTunnel 启动日志隧道
func (s *RemoteLogService) StartTunnel(ctx context.Context, deviceID string) {
	// 1. 启动时立即调用服务端的 LogTunnel
	stream, err := s.remoteLogClient.LogTunnel(ctx)
	if err != nil {
		log.Printf("建立日志隧道失败: %v", err)
		return
	}

	// 2. 关键一步：连上后立马发第一个包，告诉服务端我是谁
	// 这样服务端才会执行 s.streams[deviceID] = stream
	err = stream.Send(&grpc_generated.DeviceResponse{
		DeviceId: deviceID,
	})
	if err != nil {
		return
	}

	// 初始化成功
	fmt.Println("远程日志隧道连接成功，准备就绪")

	// 3. 开启循环，死等服务端的“文件名”指令
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("隧道断开: %v", err)
			break
		}

		// 4. 收到服务端传来的文件名
		lines := make([]string, 0, 200)
		if req.FileName != "" {
			// 强制提取纯文件名，防止 ../.. 攻击
			safeFileName := filepath.Base(req.FileName)

			// 拼接成绝对路径
			absLogDir, _ := filepath.Abs("logs")

			// 获取 logs 目录的绝对路径
			targetPath := filepath.Join(absLogDir, safeFileName)

			// 二次检查：确保最终生成的路径依然在 logs 目录下
			if !strings.HasPrefix(targetPath, absLogDir) {
				lines = []string{"[Security] 拒绝跨目录访问"}
				continue
			}

			// 日志存储在项目根目录的 logs 文件夹下
			// targetPath := filepath.Join("logs", requestParams.FileName)

			// 读取文件末尾行 (例如读取最后 200 行)
			lines, err = utils.ReadLastLines(targetPath, 200)
			if err != nil {
				// 如果文件不存在或读取失败，返回空列表或错误提示
				lines = []string{"[System] 暂无日志文件或读取失败: " + err.Error()}
				continue
			}
		} else {
			lines = []string{"[System] 请输入日志文件名"}
			continue
		}

		// 5. 将结果塞回流里发给服务端
		resp := &grpc_generated.DeviceResponse{
			DeviceId: deviceID,
			Lines:    lines,
		}
		if err := stream.Send(resp); err != nil {
			s.logger.Error("回传日志数据失败",
				slog.Any("err", err),
			)
			break
		}
	}
}
