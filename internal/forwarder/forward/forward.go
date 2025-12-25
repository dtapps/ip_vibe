package forward

import (
	"context"
	"log/slog"
	"net"
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
	ActiveConn    int64
	TotalConn     int64
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

func NewAgent(logger *slog.Logger) *Agent {
	return &Agent{
		forwards: make(map[string][]*ForwardInstance),
		logger:   logger,
	}
}

// Sync 热更新规则
func (a *Agent) Sync(ctx context.Context, rules []Rule) {
	a.mu.Lock()
	defer a.mu.Unlock()

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

	// 添加新规则
	for _, r := range rules {
		if _, ok := a.forwards[r.ID]; !ok {
			a.addRule(ctx, &r)
		}
	}
}

// 添加规则
func (a *Agent) addRule(ctx context.Context, rule *Rule) {
	var instances []*ForwardInstance
	for _, proto := range rule.Protocols {
		ctxCancel, cancel := context.WithCancel(ctx)
		stats := &Stats{}

		fi := &ForwardInstance{
			rule:   rule,
			proto:  proto,
			stats:  stats,
			cancel: cancel,
		}

		laddr := rule.ListenAddr
		raddr := rule.TargetAddr

		switch proto {
		case "tcp4", "tcp6":
			ln, err := net.Listen(proto, laddr)
			if err != nil {
				a.logger.Warn("TCP listen failed", slog.String("proto", proto), slog.String("addr", laddr), slog.Any("err", err))
				atomic.AddInt64(&stats.Errors, 1)
				continue
			}
			fi.listener = ln
			go startTCPForward(ln, raddr, stats, ctxCancel, a.logger, rule.ID, proto)
		case "udp4", "udp6":
			udpConn, err := net.ListenUDP(proto, resolveUDPAddr(laddr, proto))
			if err != nil {
				a.logger.Warn("UDP listen failed", slog.String("proto", proto), slog.String("addr", laddr), slog.Any("err", err))
				atomic.AddInt64(&stats.Errors, 1)
				continue
			}
			fi.udpConn = udpConn
			go startUDPForward(udpConn, raddr, stats, ctxCancel, a.logger, rule.ID, proto)
		default:
			a.logger.Warn("Unknown protocol", slog.String("proto", proto))
			continue
		}

		instances = append(instances, fi)
	}

	if len(instances) > 0 {
		a.forwards[rule.ID] = instances
	}
}

// 获取统计信息
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
