package forward

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// TCP 转发
func startTCP(proto, laddr, raddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID string) (net.Listener, error) {
	ln, err := net.Listen(proto, laddr)
	if err != nil {
		return nil, err
	}

	logger.Info("TCP 转发已启动",
		slog.String("规则ID", ruleID),
		slog.String("协议", proto),
		slog.String("监听", laddr),
		slog.String("目标", raddr),
	)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if ctx.Err() != nil {
					logger.Info("TCP 监听已关闭",
						slog.String("规则ID", ruleID),
						slog.String("协议", proto),
					)
					return
				}
				logger.Warn("TCP Accept 错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", err),
				)
				return
			}

			atomic.AddInt64(&stats.ActiveConn, 1)
			atomic.AddInt64(&stats.TotalConn, 1)

			go handleTCPConn(conn, proto, raddr, stats, logger, ruleID, ctx)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	return ln, nil
}

func handleTCPConn(c net.Conn, proto, raddr string, stats *Stats, logger *slog.Logger, ruleID string, ctx context.Context) {
	defer func() { _ = c.Close() }()
	defer atomic.AddInt64(&stats.ActiveConn, -1)

	rc, err := net.Dial(proto, raddr)
	if err != nil {
		atomic.AddInt64(&stats.Errors, 1)
		logger.Warn("TCP 连接目标失败",
			slog.String("规则ID", ruleID),
			slog.String("协议", proto),
			slog.String("目标", raddr),
			slog.Any("错误", err),
		)
		return
	}
	defer func() { _ = rc.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		proxyTCP(c, rc, &stats.BytesIn, logger, ruleID, proto)
		// 半关闭：通知远端不再发送数据
		if tc, ok := rc.(*net.TCPConn); ok {
			_ = tc.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		proxyTCP(rc, c, &stats.BytesOut, logger, ruleID, proto)
		// 半关闭：通知客户端不再发送数据
		if tc, ok := c.(*net.TCPConn); ok {
			_ = tc.CloseWrite()
		}
	}()

	wg.Wait()
}

func proxyTCP(src, dst net.Conn, counter *int64, logger *slog.Logger, ruleID, proto string) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			wn, werr := dst.Write(buf[:n])
			if wn > 0 {
				atomic.AddInt64(counter, int64(wn))
			}
			if werr != nil {
				logger.Warn("TCP 写入错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", werr),
				)
				return
			}
			logger.Debug("TCP 转发数据",
				slog.String("规则ID", ruleID),
				slog.String("协议", proto),
				slog.Int("字节数", n),
			)
		}
		if err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				logger.Warn("TCP 转发错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", err),
				)
			}
			return
		}
	}
}

// ---------------- UDP ----------------

func startUDP(proto, laddr, raddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID string) (*net.UDPConn, error) {
	la, err := net.ResolveUDPAddr(proto, laddr)
	if err != nil {
		return nil, err
	}
	ra, err := net.ResolveUDPAddr(proto, raddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP(proto, la)
	if err != nil {
		return nil, err
	}

	logger.Info("UDP 转发已启动",
		slog.String("规则ID", ruleID),
		slog.String("协议", proto),
		slog.String("监听", laddr),
		slog.String("目标", raddr),
	)

	go handleUDPWithLog(conn, ra, stats, logger, ruleID, proto, ctx)

	return conn, nil
}

func handleUDPWithLog(conn *net.UDPConn, raddr *net.UDPAddr, stats *Stats, logger *slog.Logger, ruleID, proto string, ctx context.Context) {
	var sessions sync.Map

	// 监听 context 取消，主动关闭监听连接以中断阻塞的 ReadFromUDP
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("UDP 监听已关闭",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
				)
			} else {
				logger.Warn("UDP 读取错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", err),
				)
			}
			return
		}

		atomic.AddInt64(&stats.UDPPacketsIn, 1)
		atomic.AddInt64(&stats.BytesIn, int64(n))

		key := src.String()

		// 获取或创建该客户端的远端连接（用于接收响应）
		val, ok := sessions.Load(key)
		if !ok {
			rc, dialErr := net.DialUDP(proto, nil, raddr)
			if dialErr != nil {
				atomic.AddInt64(&stats.Errors, 1)
				logger.Warn("UDP 连接远端失败",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.String("目标", raddr.String()),
					slog.Any("错误", dialErr),
				)
				continue
			}
			sessions.Store(key, rc)

			// 后台 goroutine：将远端响应回传给客户端
			go func(clientAddr *net.UDPAddr, remoteConn *net.UDPConn) {
				defer func() { _ = remoteConn.Close() }()
				relayUDPResponse(remoteConn, conn, clientAddr, stats, logger, ruleID, proto)
				// 超时后自动清理会话
				time.AfterFunc(60*time.Second, func() {
					sessions.CompareAndDelete(clientAddr.String(), remoteConn)
				})
			}(src, rc)
		}

		rc := val.(*net.UDPConn)

		_, werr := rc.Write(buf[:n])
		if werr != nil {
			logger.Warn("UDP 写入远端错误",
				slog.String("规则ID", ruleID),
				slog.String("协议", proto),
				slog.Any("错误", werr),
			)
		} else {
			atomic.AddInt64(&stats.UDPPacketsOut, 1)
			atomic.AddInt64(&stats.BytesOut, int64(n))
		}

		logger.Debug("UDP 转发数据",
			slog.String("规则ID", ruleID),
			slog.String("协议", proto),
			slog.Int("字节数", n),
			slog.String("来源", src.String()),
		)
	}
}

// relayUDPResponse 从远端连接读取响应，写回客户端
func relayUDPResponse(rc *net.UDPConn, listener *net.UDPConn, clientAddr *net.UDPAddr, stats *Stats, logger *slog.Logger, ruleID, proto string) {
	buf := make([]byte, 65535)
	for {
		n, err := rc.Read(buf)
		if n > 0 {
			_, werr := listener.WriteToUDP(buf[:n], clientAddr)
			if werr != nil {
				logger.Warn("UDP 回传客户端错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", werr),
				)
			} else {
				atomic.AddInt64(&stats.UDPPacketsIn, 1)
				atomic.AddInt64(&stats.BytesIn, int64(n))
			}
		}
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logger.Warn("UDP 读取远端响应错误",
					slog.String("规则ID", ruleID),
					slog.String("协议", proto),
					slog.Any("错误", err),
				)
			}
			return
		}
	}
}
