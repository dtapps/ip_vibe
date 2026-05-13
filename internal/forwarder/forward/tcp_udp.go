package forward

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
)

// TCP 转发
func startTCP(proto, laddr, raddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID string) (net.Listener, error) {
	ln, err := net.Listen(proto, laddr)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			atomic.AddInt64(&stats.ActiveConn, 1)
			atomic.AddInt64(&stats.TotalConn, 1)

			go func(c net.Conn) {
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

				go proxyTCP(c, rc, &stats.BytesIn, logger, ruleID, proto)
				proxyTCP(rc, c, &stats.BytesOut, logger, ruleID, proto)
			}(conn)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	return ln, nil
}

func proxyTCP(src, dst net.Conn, counter *int64, logger *slog.Logger, ruleID, proto string) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			_, _ = dst.Write(buf[:n])
			atomic.AddInt64(counter, int64(n))
			logger.Debug("TCP 转发数据",
				slog.String("规则ID", ruleID),
				slog.String("协议", proto),
				slog.Int("字节数", n),
			)
		}
		if err != nil {
			if err != io.EOF {
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

	go handleUDPWithLog(conn, ra, stats, logger, ruleID, proto, ctx)

	return conn, nil
}

func handleUDPWithLog(conn *net.UDPConn, raddr *net.UDPAddr, stats *Stats, logger *slog.Logger, ruleID, proto string, ctx context.Context) {
	buf := make([]byte, 65535)
	for {
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return
		default:
		}

		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			logger.Warn("UDP 读取错误",
				slog.String("规则ID", ruleID),
				slog.String("协议", proto),
				slog.Any("错误", err),
			)
			return
		}

		atomic.AddInt64(&stats.UDPPacketsIn, 1)
		atomic.AddInt64(&stats.BytesIn, int64(n))

		_, err = conn.WriteToUDP(buf[:n], raddr)
		if err != nil {
			logger.Warn("UDP 写入错误",
				slog.String("规则ID", ruleID),
				slog.String("协议", proto),
				slog.Any("错误", err),
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
