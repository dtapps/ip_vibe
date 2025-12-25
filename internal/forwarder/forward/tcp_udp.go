package forward

import (
	"context"
	"net"
	"sync/atomic"
)

// TCP 转发
func startTCP(proto, laddr, raddr string, stats *Stats, ctx context.Context) (net.Listener, error) {
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
				defer c.Close()
				defer atomic.AddInt64(&stats.ActiveConn, -1)

				rc, err := net.Dial(proto, raddr)
				if err != nil {
					atomic.AddInt64(&stats.Errors, 1)
					return
				}
				defer rc.Close()

				go proxyTCP(c, rc, &stats.BytesIn)
				proxyTCP(rc, c, &stats.BytesOut)
			}(conn)
		}
	}()

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	return ln, nil
}

func proxyTCP(src, dst net.Conn, counter *int64) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
			atomic.AddInt64(counter, int64(n))
		}
		if err != nil {
			return
		}
	}
}

// UDP 转发
func startUDP(proto, laddr, raddr string, stats *Stats, ctx context.Context) (*net.UDPConn, error) {
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

	go func() {
		buf := make([]byte, 65535)
		for {
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			atomic.AddInt64(&stats.UDPPacketsIn, 1)
			atomic.AddInt64(&stats.BytesIn, int64(n))

			conn.WriteToUDP(buf[:n], ra)
			atomic.AddInt64(&stats.UDPPacketsOut, 1)
			atomic.AddInt64(&stats.BytesOut, int64(n))

			_ = src
		}
	}()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return conn, nil
}
