package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/spf13/cast"
	"resty.dev/v3"
)

// CloudfareDNSConfig Cloudfare DNS 配置
type CloudfareDNSConfig struct {
	ApiToken string `json:"api_token"` // API令牌
}

// DefaultValidate 验证配置
func (cfg *CloudfareDNSConfig) DefaultValidate() error {
	if cfg.ApiToken == "" {
		return fmt.Errorf("ApiToken 不能为空")
	}
	return nil
}

// NewDefaultCloudfareDNSConfig 创建默认的Cloudfare DNS 配置
func NewDefaultCloudfareDNSConfig() *CloudfareDNSConfig {
	return &CloudfareDNSConfig{}
}

// ParseCustomConfig 解析 Cloudfare DNS 自定义配置
func (p *CloudfareDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// CloudfareDNSProvider Cloudfare DNS 提供程序
type CloudfareDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *CloudfareDNSConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *CloudfareDNSCustomConfig
	Provider     struct {
		Opts   []option.RequestOption
		Client *cloudflare.Client
		Resty  *resty.Client
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData dns.RecordResponse
}

// NewCloudfareDNSProvider 创建Cloudfare DNS 提供程序
func NewCloudfareDNSProvider(ctx context.Context, logger *slog.Logger, cfg *CloudfareDNSConfig, opts ...Option) (provider.DNSProvider, error) {
	return NewCloudfareDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewCloudfareDNSProviderRaw 创建Cloudfare DNS 提供程序
func NewCloudfareDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *CloudfareDNSConfig, opts ...Option) (*CloudfareDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &CloudfareDNSProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &CloudfareDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
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
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = dnsHttpTransport
	}
	if p.opts.httpTransport != nil {
		p.Provider.Opts = append(p.Provider.Opts, option.WithHTTPClient(&http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}))
	}

	// 实例化
	p.Provider.Client = cloudflare.NewClient(p.Provider.Opts...)

	// 实例化 Resty 客户端
	p.Provider.Resty = resty.New()

	// 设置 请求头 Authorization
	p.Provider.Resty.SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.cfg.ApiToken))

	// 设置 请求头 Content-Type
	p.Provider.Resty.SetHeader("Content-Type", "application/json")

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = dnsHttpTransport
	}
	if p.opts.httpTransport != nil {
		p.Provider.Resty.SetTransport(p.opts.httpTransport.Instance(p.Provider.Resty.Transport()))
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *CloudfareDNSProvider) Name() string {
	return "Cloudflare DNS"
}

// Check 检查子域名是否存在
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
func (p *CloudfareDNSProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (data *provider.DNSProviderInternalData, err error) {

	// 查询域名记录
	requestDnsQuery := dns.RecordListParams{}
	requestDnsQuery.ZoneID = cloudflare.String(p.CustomConfig.ZoneValue)
	requestDnsQuery.Name = cloudflare.F(dns.RecordListParamsName{
		Exact: cloudflare.String(full_domain),
	})
	responseDnsQuery, err := p.Provider.Client.DNS.Records.List(ctx, requestDnsQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DNS.Records.List 失败 %w", err)
	}
	// 响应参数
	if responseDnsQuery.Result == nil || len(responseDnsQuery.Result) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseDnsQuery.Result {

		// 查找域名
		itemFullDomain := item.Name

		if utils.EqualDomainIgnoreCase(itemFullDomain, full_domain) {

			// 记录数据
			p.ItemData = item

			// 查找记录类型
			itemRecordType := cast.ToString(item.Type)
			// 查找记录值
			itemOriginValue := item.Content
			// 查找记录集的描述信息
			itemComment := item.Comment

			// 域名ID
			p.InternalData.RecordID = item.ID

			// 检查记录类型
			if record_type != "" && !utils.EqualDomainIgnoreCase(itemRecordType, record_type) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录类型不匹配，期望 %v，实际 %v", record_type, itemRecordType)
				return p.InternalData, nil
			}

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, origin_value) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", origin_value, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录描述
			if !utils.EqualDomainIgnoreCase(itemComment, p.CustomConfig.Comment) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录描述不匹配，期望 %v，实际 %v", p.CustomConfig.Comment, itemComment)
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

// Update 更新 DNS 记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/edit/
func (p *CloudfareDNSProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (err error) {

	if p.InternalData.RecordID == "" {

		// 创建域名记录
		requestDnsCreateBody := dns.ARecordParam{}
		requestDnsCreateBody.Name = cloudflare.String(full_domain)
		requestDnsCreateBody.Type = cloudflare.F(dns.ARecordType(record_type))
		requestDnsCreateBody.Content = cloudflare.String(origin_value)
		requestDnsCreateBody.Comment = cloudflare.String(p.CustomConfig.Comment)
		requestDnsCreate := dns.RecordNewParams{
			ZoneID: cloudflare.String(p.CustomConfig.ZoneValue),
			Body:   requestDnsCreateBody,
		}
		responseDnsCreate, err := p.Provider.Client.DNS.Records.New(
			ctx,
			requestDnsCreate,
		)
		if err != nil {
			return fmt.Errorf("接口 DNS.Records.New 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = responseDnsCreate.ID

	} else if p.InternalData.RecordID != "" {

		// 更新域名记录
		requestDnsUpdateBody := dns.ARecordParam{}
		requestDnsUpdateBody.Name = cloudflare.String(full_domain)
		requestDnsUpdateBody.Type = cloudflare.F(dns.ARecordType(record_type))
		requestDnsUpdateBody.Content = cloudflare.String(origin_value)
		requestDnsUpdateBody.Comment = cloudflare.String(p.CustomConfig.Comment)
		requestDnsUpdate := dns.RecordEditParams{
			ZoneID: cloudflare.String(p.CustomConfig.ZoneValue),
			Body:   requestDnsUpdateBody,
		}
		responseDnsUpdate, err := p.Provider.Client.DNS.Records.Edit(
			ctx,
			p.InternalData.RecordID,
			requestDnsUpdate,
		)
		if err != nil {
			return fmt.Errorf("接口 DNS.Records.Edit 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = responseDnsUpdate.ID

	}

	return nil
}

// Cloudfare DNS 自定义配置
type CloudfareDNSCustomConfig struct {
	ZoneValue     string `json:"zone_value"`                 // 域名编号
	IPGroupIDCode string `json:"ip_group_id_code,omitempty"` // 设备IP组代码
	Comment       string `json:"comment,omitempty"`          // 备注
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *CloudfareDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// ZonesRaw 获取所有站点
// https://developers.cloudflare.com/api/go/resources/zones/methods/list/
func (p *CloudfareDNSProvider) ZonesRaw(ctx context.Context) (lines []zones.Zone, err error) {
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

// RecordsRaw 获取所有解析记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
func (p *CloudfareDNSProvider) RecordsRaw(ctx context.Context, zone_id string) (lines []dns.RecordResponse, err error) {
	// 查询域名列表
	requestQueryList := dns.RecordListParams{}
	requestQueryList.ZoneID = cloudflare.String(zone_id)
	requestQueryList.Page = cloudflare.F(float64(1))
	requestQueryList.PerPage = cloudflare.F(float64(5000000))
	responseQueryList, err := p.Provider.Client.DNS.Records.List(ctx, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 DNS.Records.List 失败 %w", err)
	}
	return responseQueryList.Result, nil
}

// GraphQLError GraphQL 通用错误结构体
type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []string       `json:"path"`
	Extensions map[string]any `json:"extensions"`
}

// StatsByDay 按天统计的网站访问数据
type StatsByDay struct {
	Date string `json:"date"` // 日期（YYYY-MM-DD）

	Requests int64 `json:"requests"` // 请求总数（单位：次）

	CachedRequests   int64 `json:"cachedRequests"`   // 已缓存请求数（单位：次）
	UncachedRequests int64 `json:"uncachedRequests"` // 未缓存请求数（单位：次）

	Bandwidth int64 `json:"bandwidth"` // 总带宽（单位：Bytes）

	CachedBandwidth   int64 `json:"cachedBandwidth"`   // 已缓存带宽（单位：Bytes）
	UncachedBandwidth int64 `json:"uncachedBandwidth"` // 未缓存带宽（单位：Bytes）

	Uniques int64 `json:"uniques"` // 唯一访问者数量（单位：人）
}

// StatsByDay 按天统计的网站访问数据
func (p *CloudfareDNSProvider) StatsByDay(ctx context.Context, zoneID string) ([]StatsByDay, error) {
	until := time.Now().UTC()
	since := until.AddDate(0, 0, -6)

	var gqlResp struct {
		Data struct {
			Viewer struct {
				Zones []struct {
					HttpRequests1dGroups []struct {
						Dimensions struct {
							Date string `json:"date"` // 日期（YYYY-MM-DD）
						} `json:"dimensions"`
						Sum struct {
							Bytes          int64 `json:"bytes"`          // 总带宽（单位：Bytes）
							CachedBytes    int64 `json:"cachedBytes"`    // 已缓存带宽（单位：Bytes）
							CachedRequests int64 `json:"cachedRequests"` // 已缓存请求数（单位：次）
							Requests       int64 `json:"requests"`       // 请求总数（单位：次）
						} `json:"sum"`
						Uniq struct {
							Uniques int64 `json:"uniques"` // 唯一访客（单位：人）
						} `json:"uniq"`
					} `json:"httpRequests1dGroups"`
				} `json:"zones"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []GraphQLError `json:"errors"`
	}

	// GraphQL 查询语句
	query := fmt.Sprintf(`
query StatsByDay {
  viewer {
    zones(filter: { zoneTag: "%s" }) {
      httpRequests1dGroups(
        limit: 1000
        orderBy: [date_ASC]
        filter: {
          date_geq: "%s"
          date_leq: "%s"
        }
      ) {
        dimensions {
          date
        }
        sum {
          requests
          cachedRequests
          bytes
          cachedBytes
        }
        uniq {
          uniques
        }
      }
    }
  }
}
`,
		zoneID,
		since.Format(time.DateOnly),
		until.Format(time.DateOnly),
	)

	// 发起 GraphQL 请求
	resp, err := p.Provider.Resty.R().
		SetContext(ctx).
		SetBody(map[string]any{
			"query": query,
		}).
		SetResult(&gqlResp).
		Post("https://api.cloudflare.com/client/v4/graphql")
	if err != nil {
		return nil, fmt.Errorf("GraphQL 请求失败: %w", err)
	}

	// 检查 HTTP 状态码
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("GraphQL 请求返回非成功状态 [%d]: %s",
			resp.StatusCode(),
			resp.String(),
		)
	}

	// 处理 GraphQL 业务错误
	if len(gqlResp.Errors) > 0 {
		errMsgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			errMsgs[i] = fmt.Sprintf("%s (path: %v)", e.Message, e.Path)
		}
		return nil, fmt.Errorf("GraphQL 查询错误: %v", strings.Join(errMsgs, "; "))
	}

	if len(gqlResp.Data.Viewer.Zones) == 0 {
		return nil, fmt.Errorf("zone %s 未返回统计数据", zoneID)
	}

	groups := gqlResp.Data.Viewer.Zones[0].HttpRequests1dGroups
	result := make([]StatsByDay, 0, len(groups))
	for _, g := range groups {

		result = append(result, StatsByDay{
			Date:              g.Dimensions.Date,
			Requests:          g.Sum.Requests,
			Bandwidth:         g.Sum.Bytes,
			Uniques:           g.Uniq.Uniques,
			CachedRequests:    g.Sum.CachedRequests,
			UncachedRequests:  g.Sum.Requests - g.Sum.CachedRequests,
			CachedBandwidth:   g.Sum.CachedBytes,
			UncachedBandwidth: g.Sum.Bytes - g.Sum.CachedBytes,
		})
	}

	return result, nil
}

// CreateRecord 创建 DNS 记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
func (p *CloudfareDNSProvider) CreateRecord(ctx context.Context, full_domain string, record_type string, origin_value string, proxied bool) (err error) {
	requestDnsCreateBody := dns.ARecordParam{}
	requestDnsCreateBody.Name = cloudflare.String(full_domain)
	requestDnsCreateBody.Type = cloudflare.F(dns.ARecordType(record_type))
	requestDnsCreateBody.Content = cloudflare.String(origin_value)
	requestDnsCreateBody.Proxied = cloudflare.Bool(proxied)
	requestDnsCreate := dns.RecordNewParams{
		ZoneID: cloudflare.String(p.CustomConfig.ZoneValue),
		Body:   requestDnsCreateBody,
	}
	_, err = p.Provider.Client.DNS.Records.New(
		ctx,
		requestDnsCreate,
	)
	if err != nil {
		return fmt.Errorf("接口 DNS.Records.New 失败 %w", err)
	}
	return nil
}
