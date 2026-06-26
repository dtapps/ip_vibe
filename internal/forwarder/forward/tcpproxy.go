package forward

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/inetaf/tcpproxy"
)

// tcpProxyHandler 包装 tcpproxy.Proxy，实现 net.Listener 接口
// tcpproxy 内部管理连接，Accept 不会被外部调用（proxy 自行处理连接）
type tcpProxyHandler struct {
	proxy  *tcpproxy.Proxy
	doneCh chan struct{}
}

// Accept 不会被调用（tcpproxy 内部处理连接），阻塞直到 Close
func (h *tcpProxyHandler) Accept() (net.Conn, error) {
	<-h.doneCh
	return nil, io.EOF
}

// Close 关闭代理
func (h *tcpProxyHandler) Close() error {
	if h.proxy != nil {
		err := h.proxy.Close()
		close(h.doneCh)
		return err
	}
	return nil
}

// Addr 返回 nil（tcpproxy 不暴露内部 listener 的地址）
func (h *tcpProxyHandler) Addr() net.Addr {
	return nil
}

// startTCPProxy 使用 github.com/inetaf/tcpproxy 启动 TCP 转发
// 仅支持 TCP 协议，不支持 UDP
func startTCPProxy(laddr, raddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID string) (*tcpProxyHandler, error) {
	proxy := &tcpproxy.Proxy{}

	// 添加默认路由：所有连接转发到目标地址
	proxy.AddRoute(laddr, tcpproxy.To(raddr))

	// 启动代理（内部会创建 TCP listener 并开始接受连接）
	if err := proxy.Start(); err != nil {
		return nil, err
	}

	logger.Info("TCPProxy 转发已启动",
		slog.String("规则ID", ruleID),
		slog.String("监听", laddr),
		slog.String("目标", raddr),
	)

	// 监听 context 取消，关闭代理
	go func() {
		<-ctx.Done()
		proxy.Close()
		atomic.AddInt64(&stats.Errors, 1)
		logger.Info("TCPProxy 监听已关闭",
			slog.String("规则ID", ruleID),
		)
	}()

	return &tcpProxyHandler{
		proxy:  proxy,
		doneCh: make(chan struct{}),
	}, nil
}
