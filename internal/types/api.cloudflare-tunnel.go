package types

import (
	"encoding/json"
	"time"
)

// 同步Cloudflare隧道 请求结构
type SyncCloudflareTunnelsRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"平台配置序号"` // 平台配置序号
}

// 获取Cloudflare隧道入站配置 请求结构
type GetCloudflareTunnelIngressRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"隧道序号"` // 隧道序号
}

// 删除Cloudflare隧道入站配置 请求结构
type DeleteCloudflareTunnelIngressRequest struct {
	ID        int64 `uri:"id" json:"id" binding:"required" jsonschema:"隧道序号"`                   // 隧道序号
	IngressID int64 `uri:"ingress_id" json:"ingress_id" binding:"required" jsonschema:"入站配置序号"` // 入站配置序号
}

// 获取CloudflareTunnel 隧道 入站配置
type GetCloudflareTunnelTunnelIngresse struct {
	ID            int64           `json:"id"`             // Cloudflare 隧道入站规则序号
	Hostname      string          `json:"hostname"`       // 主机名
	Service       string          `json:"service"`        // 服务
	Path          string          `json:"path"`           // 路径
	OriginRequest json.RawMessage `json:"origin_request"` // 入站请求配置（JSON）
}

// 获取CloudflareTunnel 隧道
type GetCloudflareTunnelTunnel struct {
	ID              int64                               `json:"id"`                // Cloudflare 隧道序号
	TunnelID        string                              `json:"tunnel_id"`         // Cloudflare 隧道ID
	Name            string                              `json:"name"`              // 隧道名称
	Status          string                              `json:"status"`            // 隧道状态
	Connections     json.RawMessage                     `json:"connections"`       // 隧道连接配置（JSON）
	ConnsActiveAt   time.Time                           `json:"conns_active_at"`   // 隧道连接活跃时间
	RemoteCreatedAt time.Time                           `json:"remote_created_at"` // 隧道创建时间
	Ingresses       []GetCloudflareTunnelTunnelIngresse `json:"ingresses"`         // 隧道入站配置
}

// 获取CloudflareTunnel 响应结构
type GetCloudflareTunnelResponse struct {
	ID          int64                       `json:"id"`          // 平台配置序号
	Capability  string                      `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string                      `json:"name"`        // 平台配置名称
	Description string                      `json:"description"` // 平台配置描述
	Platform    string                      `json:"platform"`    // 平台类型
	Tunnels     []GetCloudflareTunnelTunnel `json:"tunnels"`     // 隧道列表
}
