package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-co-op/gocron/v2"
	"resty.dev/v3"
)

type PortForwardService struct {
	cron   gocron.Scheduler // 定时任务调度器
	job    gocron.Job       // IP上报任务
	client *resty.Client    // Resty HTTP 客户端
	logger *slog.Logger     // 日志记录器
}

// NewPortForwardService 创建一个新的服务实例
func NewPortForwardService(cron gocron.Scheduler) *PortForwardService {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 请求跟踪
	client.EnableTrace()

	// 设置默认超时
	client.SetTimeout(30 * time.Second)

	// 设置请求头
	client.SetHeader("X-Version", global.GvaVersion)
	client.SetHeader("X-Git-Commit", global.GvaGitCommit)
	client.SetHeader("X-Build-Time", global.GvaBuildTime)
	client.SetHeader("X-Go-Version", global.GvaGoVersion)
	client.SetHeader("X-OS", global.GvaOS)
	client.SetHeader("X-Arch", global.GvaArch)

	// 重试机制
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &PortForwardService{
		cron:   cron,
		client: client,
		logger: slog.With("component", "port-forwards"),
	}
}

// Close 关闭服务，释放资源
func (s *PortForwardService) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.cron != nil && s.job != nil {
		_ = s.cron.RemoveJob(s.job.ID())
	}
}

// SetDebug 设置调试模式
func (s *PortForwardService) SetDebug(debug bool) *PortForwardService {
	s.client.SetDebug(debug)
	return s
}

// SetBaseURL 设置基础URL
func (s *PortForwardService) SetBaseURL(url string) *PortForwardService {
	s.client.SetBaseURL(url)
	return s
}

// 设置基本认证
func (s *PortForwardService) SetBasicAuth(username, password string) *PortForwardService {
	if username != "" && password != "" {
		s.client.SetBasicAuth(username, password)
	}
	return s
}

func (s *PortForwardService) StartJob(ctx context.Context) (err error) {
	// 请求参数
	deviceCode := global.GvaEnv.DEVICE_CODE
	// 检查设备代码是否为空
	if deviceCode == "" {
		// 获取设备ID
		deviceCode, err = machineid.ID()
		if err != nil {
			s.logger.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}
	s.logger.InfoContext(ctx, "端口转发获取开始",
		slog.String("device_code", deviceCode),
	)
	response, err := s.httpApi(ctx, deviceCode)
	if err != nil {
		s.logger.ErrorContext(ctx, "端口转发获取失败",
			slog.String("device_code", deviceCode),
			slog.Any("err", err),
		)
		return
	}
	if len(response.Data) == 0 {
		s.logger.InfoContext(ctx, "端口转发获取成功, 但无数据",
			slog.String("device_code", deviceCode),
		)
		return
	}
	startForwarding(response.Data)
	s.logger.InfoContext(ctx, "端口转发获取成功",
		slog.String("device_code", deviceCode),
	)
	return
}

type HttpApiResponse struct {
	Status int                            `json:"status"`
	Data   []types.GetPortForwardResponse `json:"data"`
}

func (s *PortForwardService) httpApi(ctx context.Context, deviceCode string) (response HttpApiResponse, err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetPathParams(map[string]string{
			"device_code": deviceCode,
		}).
		SetResult(&response).
		Get("api/open/device/{device_code}/port-forwards")
	if err != nil {
		return HttpApiResponse{}, err
	}

	// 打印响应状态和内容
	s.logger.InfoContext(ctx, "响应状态码和内容",
		slog.Int("status_code", resp.StatusCode()),
		slog.Any("response", resp.String()),
	)

	if resp.IsError() {
		return HttpApiResponse{}, fmt.Errorf("响应状态码错误: %d", resp.StatusCode())
	}

	return response, nil
}

var activeForwarders = make(map[int64]*ForwarderGroup)
var mu sync.Mutex

type ForwarderGroup struct {
	rule   types.GetPortForwardResponse
	closer []io.Closer
}

func (fg *ForwarderGroup) Close() {
	for _, c := range fg.closer {
		c.Close()
	}
}

// 启动所有转发
func startForwarding(rules []types.GetPortForwardResponse) {
	mu.Lock()
	defer mu.Unlock()

	// 先停止所有旧转发（满足“全局只能一个任务”）
	for _, fg := range activeForwarders {
		fg.Close()
	}
	activeForwarders = make(map[int64]*ForwarderGroup)

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		fg := &ForwarderGroup{rule: rule}
		for _, proto := range rule.Protocols {
			var err error
			switch proto {
			case "tcp4":
				err = startTCPForward("tcp4", rule.LocalPort, rule.RemoteHost, rule.RemotePort, fg)
			case "tcp6":
				err = startTCPForward("tcp6", rule.LocalPort, rule.RemoteHost, rule.RemotePort, fg)
			case "udp4":
				err = startUDPForward("udp4", rule.LocalPort, rule.RemoteHost, rule.RemotePort, fg)
			case "udp6":
				err = startUDPForward("udp6", rule.LocalPort, rule.RemoteHost, rule.RemotePort, fg)
			default:
				log.Printf("Unknown protocol: %s", proto)
			}
			if err != nil {
				log.Printf("Failed to start %s forward for rule %d: %v", proto, rule.ID, err)
			}
		}
		activeForwarders[rule.ID] = fg
	}
}

func startTCPForward(network string, localPort int, remoteHost string, remotePort int, fg *ForwarderGroup) error {
	localAddr := fmt.Sprintf(":%d", localPort)
	listener, err := net.Listen(network, localAddr)
	if err != nil {
		return err
	}
	fg.closer = append(fg.closer, listener)

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // listener closed
			}
			go handleTCP(conn, remoteHost, remotePort)
		}
	}()
	log.Printf("Started %s forward: %s -> %s:%d", network, localAddr, remoteHost, remotePort)
	return nil
}

func handleTCP(localConn net.Conn, remoteHost string, remotePort int) {
	defer localConn.Close()
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remoteHost, remotePort))
	if err != nil {
		log.Printf("Dial remote error: %v", err)
		return
	}
	defer remoteConn.Close()

	go io.Copy(remoteConn, localConn)
	io.Copy(localConn, remoteConn)
}

func startUDPForward(network string, localPort int, remoteHost string, remotePort int, fg *ForwarderGroup) error {
	localAddr := fmt.Sprintf(":%d", localPort)
	conn, err := net.ListenPacket(network, localAddr)
	if err != nil {
		return err
	}
	fg.closer = append(fg.closer, conn)

	go handleUDP(conn, remoteHost, remotePort)
	log.Printf("Started %s forward: %s -> %s:%d", network, localAddr, remoteHost, remotePort)
	return nil
}

func handleUDP(localConn net.PacketConn, remoteHost string, remotePort int) {
	defer localConn.Close()
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteHost, remotePort))
	if err != nil {
		log.Printf("Resolve UDP addr error: %v", err)
		return
	}

	buf := make([]byte, 65536)
	for {
		n, _, err := localConn.ReadFrom(buf)
		if err != nil {
			return
		}
		// 异步转发（注意：UDP 是无连接的，通常需记录 clientAddr 用于回包，但这里为简单转发）
		go func() {
			_, err := localConn.WriteTo(buf[:n], remoteAddr)
			if err != nil {
				log.Printf("UDP write error: %v", err)
			}
		}()
	}
}
