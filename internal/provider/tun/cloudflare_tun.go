package tun

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cast"
)

// CloudfareTUNConfig Cloudfare TUN 配置
type CloudfareTUNConfig struct {
	ApiToken  string `json:"api_token"`  // API令牌
	AccountID string `json:"account_id"` // 账号ID
}

// Validate 验证Cloudfare TUN 配置
func (cfg *CloudfareTUNConfig) Validate() error {
	if cfg.ApiToken == "" {
		return fmt.Errorf("ApiToken 不能为空")
	}
	if cfg.AccountID == "" {
		return fmt.Errorf("AccountID 不能为空")
	}
	return nil
}

// NewDefaultCloudfareTUNConfig 创建默认的Cloudfare TUN 配置
func NewDefaultCloudfareTUNConfig() *CloudfareTUNConfig {
	return &CloudfareTUNConfig{}
}

// ParseCustomConfig 解析 Cloudfare TUN 自定义配置
func (p *CloudfareTUNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// CloudfareTUNProvider Cloudfare TUN 提供程序
type CloudfareTUNProvider struct {
	provider.UnimplementedTUNProvider // 嵌入未实现的

	cfg    *CloudfareTUNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *CloudfareTUNCustomConfig
	Provider     struct {
		Opts   []option.RequestOption
		Client *cloudflare.Client
	}

	// 内部数据
	InternalData *provider.TUNProviderInternalData

	// 记录数据
	ItemData []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress
}

// NewCloudfareTUNProvider 创建Cloudfare TUN 提供程序
func NewCloudfareTUNProvider(ctx context.Context, logger *slog.Logger, cfg *CloudfareTUNConfig, opts ...Option) (provider.TUNProvider, error) {
	return NewCloudfareTUNProviderRaw(ctx, logger, cfg, opts...)
}

// NewCloudfareTUNProviderRaw 创建Cloudfare TUN 提供程序
func NewCloudfareTUNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *CloudfareTUNConfig, opts ...Option) (*CloudfareTUNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &CloudfareTUNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &CloudfareTUNCustomConfig{},
		InternalData: &provider.TUNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化client选项
	p.Provider.Opts = append(p.Provider.Opts, option.WithAPIToken(p.cfg.ApiToken))

	// 定义最大重试次数
	p.Provider.Opts = append(p.Provider.Opts, option.WithMaxRetries(3))

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Opts = append(p.Provider.Opts, option.WithHTTPClient(&http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}))
	}

	// 实例化
	p.Provider.Client = cloudflare.NewClient(p.Provider.Opts...)

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *CloudfareTUNProvider) Name() string {
	return "Cloudflare TUN"
}

// 处理主机名和服务
func (p *CloudfareTUNProvider) handleHostnameAndService(sub_domain string, serve_type string, serve_uri string) (hostname string, service string) {
	// 转小写
	serve_type = strings.ToLower(serve_type)

	// unix+tls需要额外处理
	if strings.EqualFold(serve_type, "unix_tls") {
		serve_type = "unix+tls"
	}

	// 主机名和服务
	switch serve_type {
	case "http_status":
		hostname = ""
		service = fmt.Sprintf("%s:%s", serve_type, serve_uri)
	default:
		hostname = fmt.Sprintf("%s.%s", sub_domain, p.CustomConfig.ZoneName)
		service = fmt.Sprintf("%s://%s", serve_type, serve_uri)
	}

	return hostname, service
}

// Check 检查子域名是否存在
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/subresources/cloudflared/subresources/configurations/methods/get/
func (p *CloudfareTUNProvider) Check(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (data *provider.TUNProviderInternalData, err error) {

	// 主机名和服务
	hostname, service := p.handleHostnameAndService(sub_domain, serve_type, serve_uri)

	// 查询记录
	requestTunQuery := zero_trust.TunnelCloudflaredConfigurationGetParams{}
	requestTunQuery.AccountID = cloudflare.String(p.cfg.AccountID)
	responseTunQuery, err := p.Provider.Client.ZeroTrust.Tunnels.Cloudflared.Configurations.Get(ctx, p.CustomConfig.TunnelID, requestTunQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 ZeroTrust.Tunnels.Cloudflared.Configurations.Get 失败 %w", err)
	}
	// 响应参数
	if responseTunQuery.Config.Ingress == nil || len(responseTunQuery.Config.Ingress) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	// 记录数据
	p.ItemData = responseTunQuery.Config.Ingress
	for _, item := range responseTunQuery.Config.Ingress {

		// 查找主机名
		itemHostname := item.Hostname

		if utils.EqualDomainIgnoreCase(itemHostname, hostname) {

			// 存在主机名
			p.InternalData.IsExist = true

			// 查找服务
			itemService := item.Service

			// 判断服务
			if !utils.EqualDomainIgnoreCase(itemService, service) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("服务不匹配，期望 %v，实际 %v", service, itemService)
				return p.InternalData, nil
			}

			p.InternalData.IsContinue = false
			return p.InternalData, nil
		}
	}

	// 未找到记录
	p.InternalData.IsContinue = true
	return p.InternalData, nil
}

// Update 更新 记录
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/subresources/cloudflared/subresources/configurations/methods/update/
func (p *CloudfareTUNProvider) Update(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (err error) {

	// 主机名和服务
	hostname, service := p.handleHostnameAndService(sub_domain, serve_type, serve_uri)

	// 从 ItemData 复制到 ingress
	configIngress := p.mapIngressToUpdate(p.ItemData)

	if p.InternalData.IsExist {
		for i, item := range configIngress {
			if utils.EqualDomainIgnoreCase(cast.ToString(item.Hostname), hostname) {
				if hostname != "" {
					configIngress[i].Hostname = cloudflare.String(hostname)
				}
				if service != "" {
					configIngress[i].Service = cloudflare.String(service)
				}
				if domain_path != "" {
					configIngress[i].Path = cloudflare.String(domain_path)
				}
			}
		}
	} else {
		configIngres := zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{}
		if hostname != "" {
			configIngres.Hostname = cloudflare.String(hostname)
		}
		if service != "" {
			configIngres.Service = cloudflare.String(service)
		}
		if domain_path != "" {
			configIngres.Path = cloudflare.String(domain_path)
		}
		configIngress = append(configIngress, configIngres)
	}

	// 重排规则，把兜底规则（如 http_status）移到最后
	configIngress = p.moveCatchAllRulesToLast(configIngress)

	// 更新记录
	requestTunUpdate := zero_trust.TunnelCloudflaredConfigurationUpdateParams{}
	requestTunUpdate.AccountID = cloudflare.String(p.cfg.AccountID)
	requestTunUpdate.Config = cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfig{
		Ingress: cloudflare.F(configIngress),
	})
	_, err = p.Provider.Client.ZeroTrust.Tunnels.Cloudflared.Configurations.Update(ctx, p.CustomConfig.TunnelID, requestTunUpdate)
	if err != nil {
		return fmt.Errorf("接口 ZeroTrust.Tunnels.Cloudflared.Configurations.Update 失败 %w", err)
	}

	return nil
}

// mapIngressToUpdate 将 TunnelCloudflaredConfigurationGetResponseConfigIngress 转换为 TunnelCloudflaredConfigurationUpdateParamsConfigIngress
func (p *CloudfareTUNProvider) mapIngressToUpdate(items []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress) []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress {
	updateRules := make([]zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress, 0, len(items))
	for _, item := range items {
		// 1. 转换基础字段
		rule := zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{
			Hostname: cloudflare.F(item.Hostname),
			Service:  cloudflare.F(item.Service),
			Path:     cloudflare.F(item.Path),
		}
		// 2. 转换 OriginRequest (只有在原始 JSON 中存在该字段时才处理)
		if item.JSON.OriginRequest.Raw() != "" {
			req := item.OriginRequest
			// 映射 OriginRequest 内部所有字段
			originReq := zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngressOriginRequest{
				// 基础设置
				ConnectTimeout:         cloudflare.F(req.ConnectTimeout),
				TLSTimeout:             cloudflare.F(req.TLSTimeout),
				TCPKeepAlive:           cloudflare.F(req.TCPKeepAlive),
				NoTLSVerify:            cloudflare.F(req.NoTLSVerify),
				NoHappyEyeballs:        cloudflare.F(req.NoHappyEyeballs),
				DisableChunkedEncoding: cloudflare.F(req.DisableChunkedEncoding),

				// HTTP & 代理设置
				HTTPHostHeader:   cloudflare.F(req.HTTPHostHeader),
				OriginServerName: cloudflare.F(req.OriginServerName),
				ProxyType:        cloudflare.F(req.ProxyType),
				HTTP2Origin:      cloudflare.F(req.HTTP2Origin),

				// 连接池管理
				KeepAliveConnections: cloudflare.F(req.KeepAliveConnections),
				KeepAliveTimeout:     cloudflare.F(req.KeepAliveTimeout),

				// 证书与安全
				CAPool:         cloudflare.F(req.CAPool),
				MatchSnItoHost: cloudflare.F(req.MatchSnItoHost),
			}
			// 3. 处理嵌套的 Access 策略
			if req.Access.TeamName != "" || len(req.Access.AUDTag) > 0 {
				originReq.Access = cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngressOriginRequestAccess{
					AUDTag:   cloudflare.F(req.Access.AUDTag),
					TeamName: cloudflare.F(req.Access.TeamName),
					Required: cloudflare.F(req.Access.Required),
				})
			}
			rule.OriginRequest = cloudflare.F(originReq)
		}
		updateRules = append(updateRules, rule)
	}
	return updateRules
}

// moveCatchAllRulesToLast 将所有 catch-all 规则移动到最后
func (p *CloudfareTUNProvider) moveCatchAllRulesToLast(rules []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress) []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress {
	var normal []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress
	var catchAll []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress

	for _, rule := range rules {
		// 如果一条规则既没有 Hostname 也没有 Service，那是之前 make([]T, len) 留下的垃圾数据，必须丢弃
		if rule.Hostname.Value == "" && rule.Service.Value == "" {
			continue
		}

		// 判断是否是兜底规则 (没有 Hostname 的规则)
		if rule.Hostname.Value == "" {
			catchAll = append(catchAll, rule)
		} else {
			normal = append(normal, rule)
		}
	}

	// Cloudflare 只允许最后有一个兜底。如果因为逻辑产生了多个，只取最后一个，或者合并。
	if len(catchAll) > 1 {
		// 比如只保留最后一条真正的 http_status 规则
		catchAll = []zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfigIngress{catchAll[len(catchAll)-1]}
	}

	return append(normal, catchAll...)
}

// Cloudfare TUN 自定义配置
type CloudfareTUNCustomConfig struct {
	TunnelID   string `json:"tunnel_id"`             // 隧道ID
	TunnelName string `json:"tunnel_name,omitempty"` // 隧道名称
	ZoneID     string `json:"zone_id"`               // 域名ID
	ZoneName   string `json:"zone_name"`             // 域名名称
}

// ParseCustomConfig 解析 TUN 自定义配置
func (p *CloudfareTUNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// CloudfareTUNProviderConnector Cloudfare TUN 隧道
type CloudfareTUNProviderConnector struct {
	ID   string `json:"id"`   // ID
	Name string `json:"name"` // 名称
}

// Connectors 获取 Cloudfare TUN 隧道列表
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/methods/list/
func (p *CloudfareTUNProvider) Connectors(ctx context.Context) (lines []*CloudfareTUNProviderConnector, err error) {
	list, err := p.ConnectorsRaw(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		lines = append(lines, &CloudfareTUNProviderConnector{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return lines, nil
}

// ConnectorsRaw 获取 Cloudfare TUN 隧道列表
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/methods/list/
func (p *CloudfareTUNProvider) ConnectorsRaw(ctx context.Context) (lines []zero_trust.TunnelListResponse, err error) {
	// 查询隧道列表
	requestQueryList := zero_trust.TunnelListParams{}
	requestQueryList.AccountID = cloudflare.F(p.cfg.AccountID)
	requestQueryList.Page = cloudflare.F(float64(1))
	requestQueryList.PerPage = cloudflare.F(float64(1000))
	responseQueryList, err := p.Provider.Client.ZeroTrust.Tunnels.List(ctx, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 ZeroTrust.Tunnels.List 失败 %w", err)
	}
	return responseQueryList.Result, nil
}

// 获取 Cloudfare TUN 隧道入站规则列表
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/subresources/cloudflared/subresources/configurations/methods/get/
func (p *CloudfareTUNProvider) IngresssRaw(ctx context.Context, tunnelID string) (lines []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress, err error) {
	requestQueryList := zero_trust.TunnelCloudflaredConfigurationGetParams{}
	requestQueryList.AccountID = cloudflare.F(p.cfg.AccountID)
	responseQueryList, err := p.Provider.Client.ZeroTrust.Tunnels.Cloudflared.Configurations.Get(ctx, tunnelID, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 ZeroTrust.Tunnels.Cloudflared.Configurations.Get 失败 %w", err)
	}
	return responseQueryList.Config.Ingress, nil
}

// CloudfareTUNProviderZone Cloudfare TUN 域名
type CloudfareTUNProviderZone struct {
	ID   string `json:"id"`   // ID
	Name string `json:"name"` // 名称
}

// Zones 获取 Cloudfare TUN 域名列表
// https://developers.cloudflare.com/api/go/resources/zones/methods/list/
func (p *CloudfareTUNProvider) Zones(ctx context.Context) (lines []*CloudfareTUNProviderZone, err error) {
	list, err := p.ZonesRaw(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		lines = append(lines, &CloudfareTUNProviderZone{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return lines, nil
}

// ZonesRaw 获取 Cloudfare TUN 域名列表
// https://developers.cloudflare.com/api/go/resources/zones/methods/list/
func (p *CloudfareTUNProvider) ZonesRaw(ctx context.Context) (lines []zones.Zone, err error) {
	// 查询域名列表
	requestQueryList := zones.ZoneListParams{}
	requestQueryList.Page = cloudflare.F(float64(1))
	requestQueryList.PerPage = cloudflare.F(float64(50))
	responseQueryList, err := p.Provider.Client.Zones.List(ctx, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 Zones.List 失败 %w", err)
	}
	return responseQueryList.Result, nil
}

// Update 更新 记录
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/subresources/cloudflared/subresources/configurations/methods/update/
func (p *CloudfareTUNProvider) UpdateIngress(ctx context.Context, tunnelID string, updatedIngresses []zero_trust.TunnelCloudflaredConfigurationGetResponseConfigIngress) (err error) {

	// 从 ItemData 复制到 ingress
	configIngress := p.mapIngressToUpdate(updatedIngresses)

	// 重排规则，把兜底规则（如 http_status）移到最后
	configIngress = p.moveCatchAllRulesToLast(configIngress)

	// 更新记录
	requestTunUpdate := zero_trust.TunnelCloudflaredConfigurationUpdateParams{}
	requestTunUpdate.AccountID = cloudflare.String(p.cfg.AccountID)
	requestTunUpdate.Config = cloudflare.F(zero_trust.TunnelCloudflaredConfigurationUpdateParamsConfig{
		Ingress: cloudflare.F(configIngress),
	})
	_, err = p.Provider.Client.ZeroTrust.Tunnels.Cloudflared.Configurations.Update(ctx, tunnelID, requestTunUpdate)
	if err != nil {
		return fmt.Errorf("接口 ZeroTrust.Tunnels.Cloudflared.Configurations.Update 失败 %w", err)
	}

	return nil
}
