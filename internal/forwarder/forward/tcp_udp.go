package forward

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
)

// TCP 转发
func startTCPForward(ln net.Listener, targetAddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID, proto string) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				atomic.AddInt64(&stats.Errors, 1)
				logger.Warn("TCP accept error", slog.String("rule_id", ruleID), slog.String("proto", proto), slog.Any("err", err))
				continue
			}
		}

		atomic.AddInt64(&stats.ActiveConn, 1)
		atomic.AddInt64(&stats.TotalConn, 1)

		go func(c net.Conn) {
			defer c.Close()
			defer atomic.AddInt64(&stats.ActiveConn, -1)

			remote, err := net.Dial(proto, targetAddr)
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				logger.Warn("TCP dial target failed", slog.String("rule_id", ruleID), slog.String("proto", proto), slog.String("target", targetAddr), slog.Any("err", err))
				return
			}
			defer remote.Close()

			done := make(chan struct{}, 2)
			go proxyTCP(c, remote, &stats.BytesOut, done)
			go proxyTCP(remote, c, &stats.BytesIn, done)
			<-done
			<-done
		}(conn)
	}
}

func proxyTCP(src, dst net.Conn, counter *int64, done chan struct{}) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
			atomic.AddInt64(counter, int64(n))
		}
		if err != nil {
			done <- struct{}{}
			return
		}
	}
}

// UDP 转发
func startUDPForward(conn *net.UDPConn, targetAddr string, stats *Stats, ctx context.Context, logger *slog.Logger, ruleID, proto string) {
	clientMap := sync.Map{} // key: clientAddr string, value: *net.UDPAddr

	targetUDP := resolveUDPAddr(targetAddr, proto)

	buf := make([]byte, 65535)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				logger.Warn("UDP read error", slog.String("rule_id", ruleID), slog.String("proto", proto), slog.Any("err", err))
				continue
			}

			atomic.AddInt64(&stats.UDPPacketsIn, 1)
			atomic.AddInt64(&stats.BytesIn, int64(n))

			// 保存客户端地址
			clientMap.Store(clientAddr.String(), clientAddr)

			// 发给目标
			_, err = conn.WriteToUDP(buf[:n], targetUDP)
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				continue
			}
			atomic.AddInt64(&stats.UDPPacketsOut, 1)
			atomic.AddInt64(&stats.BytesOut, int64(n))

			// 异步回包处理
			go func(data []byte) {
				n2, err := conn.WriteToUDP(data[:n], clientAddr)
				if err == nil {
					atomic.AddInt64(&stats.UDPPacketsOut, 1)
					atomic.AddInt64(&stats.BytesOut, int64(n2))
				}
			}(append([]byte(nil), buf[:n]...))
		}
	}
}

func resolveUDPAddr(addr, proto string) *net.UDPAddr {
	udpAddr, _ := net.ResolveUDPAddr(proto, addr)
	return udpAddr
}
