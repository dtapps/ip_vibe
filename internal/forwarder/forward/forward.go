package forward

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Rule 定义转发规则
type Rule struct {
	ID                 string   // 唯一标识
	ListenAddr         string   // 本地监听地址
	TargetAddr         string   // 目标地址
	Protocols          []string // tcp4 / tcp6 / udp4 / udp6
	ImplementationType string   // 实现类型: "standard" / "tcpproxy"
}

// Stats 统计信息
type Stats struct {
	StartTime     time.Time `json:"start_time"`      // 转发实例启动时间 (UTC)
	ActiveConn    int64     `json:"active_conn"`     // TCP 活跃连接 / UDP session
	TotalConn     int64     `json:"total_conn"`      // TCP 累计连接
	BytesIn       int64     `json:"bytes_in"`        // 入站流量
	BytesOut      int64     `json:"bytes_out"`       // 出站流量
	UDPPacketsIn  int64     `json:"udp_packets_in"`  // 入站 UDP 包数
	UDPPacketsOut int64     `json:"udp_packets_out"` // 出站 UDP 包数
	Errors        int64     `json:"errors"`          // 错误数
}

// ForwardInstance 单协议实例
type ForwardInstance struct {
	rule     *Rule
	proto    string
	listener net.Listener
	udpConn  *net.UDPConn
	tcpProxy *tcpProxyHandler // tcpproxy 实现（仅 ImplementationType=tcpproxy 时使用）
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

// 热更新规则
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

	// 添加 / 更新新规则
	for _, r := range rules {
		if _, ok := a.forwards[r.ID]; !ok {
			a.addRule(ctx, &r)
		}
	}
}

// 添加单条规则
func (a *Agent) addRule(ctx context.Context, rule *Rule) {
	startTime := time.Now().UTC()
	ctxCancel, cancel := context.WithCancel(ctx)
	var instances []*ForwardInstance
	// 如果没有成功创建任何实例，确保 cancel 被调用以避免 context 泄漏
	defer func() {
		if len(instances) == 0 {
			cancel()
		}
	}()
	stats := &Stats{
		StartTime: startTime,
	}

	hasTCP4, hasTCP6 := false, false
	hasUDP4, hasUDP6 := false, false
	for _, proto := range rule.Protocols {
		switch proto {
		case "tcp4":
			hasTCP4 = true
		case "tcp6":
			hasTCP6 = true
		case "udp4":
			hasUDP4 = true
		case "udp6":
			hasUDP6 = true
		}
	}

	for _, proto := range rule.Protocols {
		if !isProtoMatchIP(proto, rule.TargetAddr) {
			a.logger.Info("跳过协议，因为 IP 与协议不匹配",
				slog.String("规则ID", rule.ID),
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

		// 自动双栈逻辑：同时开启 v4/v6 时合并为双栈监听，并跳过冗余迭代
		// 由匹配目标 IP 版本的协议负责合并，另一个跳过
		skipProto := false
		targetIPv4 := net.ParseIP(strings.Split(rule.TargetAddr, ":")[0]).To4() != nil
		switch proto {
		case "tcp4":
			if hasTCP6 && !targetIPv4 {
				skipProto = true // IPv6 目标，tcp6 负责合并
			} else if hasTCP6 {
				a.logger.Info("检测到 TCP4/TCP6 同时开启，使用自动双栈监听 TCP", slog.String("规则ID", rule.ID))
				proto = "tcp"
			}
		case "tcp6":
			if hasTCP4 && targetIPv4 {
				skipProto = true // IPv4 目标，tcp4 负责合并
			} else if hasTCP4 {
				a.logger.Info("检测到 TCP4/TCP6 同时开启，使用自动双栈监听 TCP", slog.String("规则ID", rule.ID))
				proto = "tcp"
			}
		case "udp4":
			if hasUDP6 && !targetIPv4 {
				skipProto = true
			} else if hasUDP6 {
				a.logger.Info("检测到 UDP4/UDP6 同时开启，使用自动双栈监听 UDP", slog.String("规则ID", rule.ID))
				proto = "udp"
			}
		case "udp6":
			if hasUDP4 && targetIPv4 {
				skipProto = true
			} else if hasUDP4 {
				a.logger.Info("检测到 UDP4/UDP6 同时开启，使用自动双栈监听 UDP", slog.String("规则ID", rule.ID))
				proto = "udp"
			}
		}
		if skipProto {
			continue
		}

		switch proto {
		case "tcp", "tcp4", "tcp6":
			if rule.ImplementationType == "tcpproxy" {
				// tcpproxy 实现：所有 TCP 协议共用一个 proxy 实例
				// 仅在第一个 TCP 协议时创建，后续复用
				var existing *tcpProxyHandler
				for _, inst := range instances {
					if inst.tcpProxy != nil {
						existing = inst.tcpProxy
						break
					}
				}
				if existing != nil {
					fi.tcpProxy = existing
					fi.listener = existing // 复用，确保 Close 正常工作
				} else {
					tp, err := startTCPProxy(laddr, raddr, stats, ctxCancel, a.logger, rule.ID)
					if err != nil {
						atomic.AddInt64(&stats.Errors, 1)
						a.logger.Warn("TCPProxy 监听启动失败",
							slog.String("协议", proto),
							slog.String("本地地址", laddr),
							slog.String("目标地址", raddr),
							slog.Any("错误", err),
						)
						continue
					}
					fi.tcpProxy = tp
					fi.listener = tp
				}
			} else {
				// standard 实现
				ln, err := startTCP(proto, laddr, raddr, stats, ctxCancel, a.logger, rule.ID)
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
			}
		case "udp", "udp4", "udp6":
			if rule.ImplementationType == "tcpproxy" {
				// tcpproxy 不支持 UDP，回退到 standard 实现
				a.logger.Info("TCPProxy 不支持 UDP，回退到 standard 实现",
					slog.String("规则ID", rule.ID),
					slog.String("协议", proto),
				)
			}
			c, err := startUDP(proto, laddr, raddr, stats, ctxCancel, a.logger, rule.ID)
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

// 判断协议是否与目标 IP 类型匹配
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
