package forwarder

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"vibe_ip/internal/types"
)

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
