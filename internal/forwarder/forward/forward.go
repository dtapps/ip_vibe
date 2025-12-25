package forward

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

// Rule 定义转发规则
type Rule struct {
	ID         string   // 唯一标识
	ListenAddr string   // 本地监听地址
	TargetAddr string   // 目标地址
	Protocols  []string // tcp4 / tcp6 / udp4 / udp6
}

// Stats 统计信息
type Stats struct {
	ActiveConn    int64 // TCP 活跃连接 / UDP session
	TotalConn     int64 // TCP 累计连接
	BytesIn       int64
	BytesOut      int64
	UDPPacketsIn  int64
	UDPPacketsOut int64
	Errors        int64
}

// ForwardInstance 单协议实例
type ForwardInstance struct {
	rule     *Rule
	proto    string
	listener net.Listener
	udpConn  *net.UDPConn
	stats    *Stats
	cancel   context.CancelFunc
}

// Agent 管理所有转发实例
type Agent struct {
	mu       sync.Mutex
	forwards map[string][]*ForwardInstance
	logger   *slog.Logger
}

// 构造函数
func NewAgent(logger *slog.Logger) *Agent {
	return &Agent{
		forwards: make(map[string][]*ForwardInstance),
		logger:   logger,
	}
}

// 热更新规则
func (a *Agent) Sync(ctx context.Context, rules []Rule) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 构建新规则 map
	newMap := make(map[string]Rule)
	for _, r := range rules {
		newMap[r.ID] = r
	}

	// 删除旧规则
	for id := range a.forwards {
		if _, ok := newMap[id]; !ok {
			for _, f := range a.forwards[id] {
				f.cancel()
			}
			delete(a.forwards, id)
		}
	}

	// 添加 / 更新新规则
	for _, r := range rules {
		if _, ok := a.forwards[r.ID]; !ok {
			a.addRule(ctx, &r)
		}
	}
}

// 添加单条规则
func (a *Agent) addRule(ctx context.Context, rule *Rule) {
	ctxCancel, cancel := context.WithCancel(ctx)
	stats := &Stats{}

	var protoSet = make(map[string]bool)
	for _, p := range rule.Protocols {
		protoSet[p] = true
	}

	var instances []*ForwardInstance

	// 自动双栈逻辑：同时需要 4/6 就只开 6
	autoTCPProto := "tcp"
	autoUDPProto := "udp"
	if protoSet["tcp4"] && protoSet["tcp6"] {
		a.logger.Info("检测到 TCP4/TCP6 同时开启，使用自动双栈监听 TCP", slog.String("规则ID", rule.ID))
		autoTCPProto = "tcp6"
	}
	if protoSet["udp4"] && protoSet["udp6"] {
		a.logger.Info("检测到 UDP4/UDP6 同时开启，使用自动双栈监听 UDP", slog.String("规则ID", rule.ID))
		autoUDPProto = "udp6"
	}

	for _, proto := range rule.Protocols {
		if !isProtoMatchIP(proto, rule.TargetAddr) {
			atomic.AddInt64(&stats.Errors, 1)
			a.logger.Warn("跳过协议，因为 IP 与协议不匹配",
				slog.String("协议", proto),
				slog.String("目标", rule.TargetAddr),
			)
			continue
		}

		fi := &ForwardInstance{
			rule:   rule,
			proto:  proto,
			stats:  stats,
			cancel: cancel,
		}

		laddr := rule.ListenAddr
		raddr := rule.TargetAddr

		switch proto {
		case "tcp", "tcp4", "tcp6":
			if protoSet["tcp4"] && protoSet["tcp6"] {
				proto = autoTCPProto
			}
			ln, err := startTCP(proto, laddr, raddr, stats, ctxCancel)
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				a.logger.Warn("TCP 监听启动失败",
					slog.String("协议", proto),
					slog.String("本地地址", laddr),
					slog.String("目标地址", raddr),
					slog.Any("错误", err),
				)
				continue
			}
			fi.listener = ln

		case "udp", "udp4", "udp6":
			if protoSet["udp4"] && protoSet["udp6"] {
				proto = autoUDPProto
			}
			c, err := startUDP(proto, laddr, raddr, stats, ctxCancel)
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				a.logger.Warn("UDP 监听启动失败",
					slog.String("协议", proto),
					slog.String("本地地址", laddr),
					slog.String("目标地址", raddr),
					slog.Any("错误", err),
				)
				continue
			}
			fi.udpConn = c
		default:
			continue
		}

		instances = append(instances, fi)
	}

	if len(instances) > 0 {
		a.forwards[rule.ID] = instances
	}
}

// 返回所有规则统计
func (a *Agent) Stats() map[string]*Stats {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make(map[string]*Stats)
	for id, list := range a.forwards {
		if len(list) > 0 {
			result[id] = list[0].stats
		}
	}
	return result
}

// 协议/IP匹配
func isProtoMatchIP(proto, addr string) bool {
	ip := net.ParseIP(strings.Split(addr, ":")[0])
	if ip == nil {
		return false
	}
	switch proto {
	case "tcp4", "udp4":
		return ip.To4() != nil
	case "tcp6", "udp6":
		return ip.To16() != nil && ip.To4() == nil
	default:
		return true
	}
}
