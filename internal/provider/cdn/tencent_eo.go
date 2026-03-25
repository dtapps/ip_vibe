package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
	"go.dtapp.net/library/utils/gotime"
)

// TencentEOConfig 腾讯云 EO 配置
type TencentEOConfig struct {
	Endpoint   string `json:"endpoint,omitempty"`    // 接口地址
	RegionType string `json:"region_type,omitempty"` // 区域类型
	SecretID   string `json:"secret_id"`             // 密钥ID
	SecretKey  string `json:"secret_key"`            // 密钥Key
}

// DefaultValidate 验证配置
func (cfg *TencentEOConfig) DefaultValidate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.RegionType == "" {
		return fmt.Errorf("RegionType 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	return nil
}

// NewDefaultTencentEOConfig 创建默认的腾讯云 EO 配置
// https://cloud.tencent.com/document/product/1552/80723
// https://www.tencentcloud.com/zh/document/product/1145/50456
// https://edgeone.ai/zh/document/50456?product=api
func NewDefaultTencentEOConfig() *TencentEOConfig {
	return &TencentEOConfig{
		Endpoint:   "teo.tencentcloudapi.com",
		RegionType: "cn",
	}
}

// ParseCustomConfig 解析 腾讯云 EO 自定义配置
func (p *TencentEOConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// TencentEOProvider 腾讯云 EO 提供程序
type TencentEOProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *TencentEOConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *TencentEOCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *teo.Client
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData teo.AccelerationDomain
}

// NewTencentEOProvider 创建腾讯云 EO 提供程序
func NewTencentEOProvider(ctx context.Context, logger *slog.Logger, cfg *TencentEOConfig, opts ...Option) (provider.CDNProvider, error) {
	return NewTencentEOProviderRaw(ctx, logger, cfg, opts...)
}

// NewTencentEOProviderRaw 创建腾讯云 EO 提供程序
func NewTencentEOProviderRaw(ctx context.Context, logger *slog.Logger, cfg *TencentEOConfig, opts ...Option) (*TencentEOProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentEOProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &TencentEOCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth = common.NewCredential(
		p.cfg.SecretID,
		p.cfg.SecretKey,
	)

	// 选项
	p.Provider.Cfg = profile.NewClientProfile()
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint // 指定域名
	p.Provider.Cfg.NetworkFailureMaxRetries = 3          // 定义最大重试次数

	// 区域类型
	if strings.EqualFold(p.cfg.RegionType, "intl") {
		p.Provider.Cfg.HttpProfile.Endpoint = "teo.intl.tencentcloudapi.com"
	} else if strings.EqualFold(p.cfg.RegionType, "cn") {
		p.Provider.Cfg.HttpProfile.Endpoint = "teo.tencentcloudapi.com"
	} else {
		p.Provider.Cfg.HttpProfile.Endpoint = "teo.tencentcloudapi.com"
	}

	// 开启日志
	p.Provider.Cfg.Debug = sys_env.GvaEnv.DEBUG

	// 实例化
	p.Provider.Client, err = teo.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = cdnHttpTransport
	}
	if p.opts.httpTransport != nil {
		p.Provider.Client.WithHttpTransport(p.opts.httpTransport.Instance(nil))
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *TencentEOProvider) Name() string {
	return "腾讯云 EO"
}

// Check 检查子域名是否存在
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
func (p *TencentEOProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (data *provider.CDNProviderInternalData, err error) {

	// 源站信息
	originValues := make([]string, 0, len(origin_values))
	if len(origin_values) == 1 {
		originValues = origin_values
	} else {
		// 获取源站组信息
		responseOriginInfo, err := p.getGroupOrigin(ctx, origin_values)
		if err != nil {
			return nil, fmt.Errorf("查询源站组失败: %w", err)
		}
		originValues = append(originValues, responseOriginInfo.ID)
	}

	// 查询加速域名
	requestCdnQuery := teo.NewDescribeAccelerationDomainsRequest()
	requestCdnQuery.ZoneId = new(p.CustomConfig.ZoneValue)
	requestCdnQuery.Offset = common.Int64Ptr(0)
	requestCdnQuery.Limit = common.Int64Ptr(200)
	requestCdnQuery.Filters = make([]*teo.AdvancedFilter, 0, 1)
	requestCdnQuery.Filters = append(requestCdnQuery.Filters, &teo.AdvancedFilter{
		Name:   new("domain-name"),
		Values: common.StringPtrs([]string{full_domain}),
	})
	responseCdnQuery, err := p.Provider.Client.DescribeAccelerationDomainsWithContext(ctx, requestCdnQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeAccelerationDomainsWithContext 失败 %w", err)
	}
	// 响应参数
	if responseCdnQuery.Response == nil || len(responseCdnQuery.Response.AccelerationDomains) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}

	for _, item := range responseCdnQuery.Response.AccelerationDomains {

		// 查找域名
		itemFullDomain := utils.PtrToString(item.DomainName)

		if utils.EqualDomainIgnoreCase(itemFullDomain, full_domain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录值列表
			itemOriginValues := make([]string, 0, 1)
			itemOriginValues = append(itemOriginValues, utils.PtrToString(item.OriginDetail.Origin))
			// 查找回源协议
			itemOriginProtocol := utils.PtrToString(item.OriginProtocol)
			// 查找HTTP回源端口
			itemHttpPort := utils.PtrToUint64(item.HttpOriginPort)
			// 查找HTTPS回源端口
			itemHttpsPort := utils.PtrToUint64(item.HttpsOriginPort)
			// 查找IPv6状态
			itemIPv6Status := utils.PtrToString(item.IPv6Status)

			// 域名ID
			// p.InternalData.RecordID = utils.TrimDotsPrt(item.DomainId)

			// CDN域名
			p.InternalData.CdnDomain = utils.TrimDotsPrt(item.Cname)

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, originValues) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", originValues, itemOriginValues)
				return p.InternalData, nil
			}

			// 判断回源协议
			if !utils.EqualDomainIgnoreCase(itemOriginProtocol, p.CustomConfig.OriginProtocolValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("回源协议不匹配，期望 %v，实际 %v", p.CustomConfig.OriginProtocolValue, itemOriginProtocol)
				return p.InternalData, nil
			}

			// 判断HTTP回源端口
			if !utils.EqualIgnoreCase(itemHttpPort, p.CustomConfig.HttpPort) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("HTTP回源端口不匹配，期望 %d，实际 %d", p.CustomConfig.HttpPort, itemHttpPort)
				return p.InternalData, nil
			}

			// 判断HTTPS回源端口
			if !utils.EqualIgnoreCase(itemHttpsPort, p.CustomConfig.HttpsPort) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("HTTPS回源端口不匹配，期望 %d，实际 %d", p.CustomConfig.HttpsPort, itemHttpsPort)
				return p.InternalData, nil
			}

			// 判断IPv6状态
			if !utils.EqualDomainIgnoreCase(itemIPv6Status, p.CustomConfig.IPv6StatusValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("IPv6状态不匹配，期望 %v，实际 %v", p.CustomConfig.IPv6StatusValue, itemIPv6Status)
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

// Update 更新 CDN 记录
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateAccelerationDomain
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyAccelerationDomain
func (p *TencentEOProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (err error) {

	// 回源协议，转大写
	p.CustomConfig.OriginProtocolValue = strings.ToUpper(p.CustomConfig.OriginProtocolValue)

	// 源站信息
	originInfo := new(teo.OriginInfo{})
	if len(origin_values) == 1 {
		originInfo.OriginType = new("IP_DOMAIN")  // 源站类型
		originInfo.Origin = new(origin_values[0]) // 源站地址
	} else {
		// 获取源站组信息
		responseOriginInfo, err := p.getGroupOrigin(ctx, origin_values)
		if err != nil {
			return fmt.Errorf("查询源站组失败: %w", err)
		}
		originInfo.OriginType = new(responseOriginInfo.OriginType) // 源站类型
		originInfo.Origin = new(responseOriginInfo.ID)             // 源站地址
	}

	// HTTP回源端口，默认80
	if p.CustomConfig.HttpPort == 0 {
		p.CustomConfig.HttpPort = 80
	}

	// HTTPS回源端口，默认443
	if p.CustomConfig.HttpsPort == 0 {
		p.CustomConfig.HttpsPort = 443
	}

	if p.InternalData.CdnDomain == "" {

		// 创建加速域名
		requestCdnCreate := teo.NewCreateAccelerationDomainRequest()
		requestCdnCreate.ZoneId = new(p.CustomConfig.ZoneValue)                   // 加速域名所属站点 ID
		requestCdnCreate.DomainName = new(full_domain)                            // 加速域名
		requestCdnCreate.OriginInfo = originInfo                                  // 源站信息
		requestCdnCreate.OriginProtocol = new(p.CustomConfig.OriginProtocolValue) // 回源协议
		requestCdnCreate.HttpOriginPort = new(uint64(p.CustomConfig.HttpPort))    // HTTP回源端口
		requestCdnCreate.HttpsOriginPort = new(uint64(p.CustomConfig.HttpsPort))  // HTTPS回源端口
		requestCdnCreate.IPv6Status = new(p.CustomConfig.IPv6StatusValue)         // IPv6状态
		_, err = p.Provider.Client.CreateAccelerationDomainWithContext(ctx, requestCdnCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateAccelerationDomainWithContext 失败 %w", err)
		}

	} else if p.InternalData.CdnDomain != "" {

		// 更新加速域名
		requestCdnUpdate := teo.NewModifyAccelerationDomainRequest()
		requestCdnUpdate.ZoneId = new(p.CustomConfig.ZoneValue)                   // 加速域名所属站点 ID
		requestCdnUpdate.DomainName = new(full_domain)                            // 加速域名
		requestCdnUpdate.OriginInfo = originInfo                                  // 源站信息
		requestCdnUpdate.OriginProtocol = new(p.CustomConfig.OriginProtocolValue) // 回源协议
		requestCdnUpdate.HttpOriginPort = new(uint64(p.CustomConfig.HttpPort))    // HTTP回源端口
		requestCdnUpdate.HttpsOriginPort = new(uint64(p.CustomConfig.HttpsPort))  // HTTPS回源端口
		requestCdnUpdate.IPv6Status = new(p.CustomConfig.IPv6StatusValue)         // IPv6状态
		_, err = p.Provider.Client.ModifyAccelerationDomainWithContext(ctx, requestCdnUpdate)
		if err != nil {
			return fmt.Errorf("接口 ModifyAccelerationDomainWithContext 失败 %w", err)
		}

	}

	return nil
}

// TencentEOProviderGroupOriginInfo 腾讯云 EO 源站组信息
type TencentEOProviderGroupOriginInfo struct {
	ID         string   // 源站组ID
	List       []string // 源站组记录值列表
	OriginType string   // 类型
}

// getGroupOrigin 获取源站组信息
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeOriginGroup
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateOriginGroup
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyOriginGroup
func (p *TencentEOProvider) getGroupOrigin(ctx context.Context, origin_values []string) (originInfo *TencentEOProviderGroupOriginInfo, err error) {

	const originType = "ORIGIN_GROUP"
	originInfo = new(TencentEOProviderGroupOriginInfo)
	originInfo.List = origin_values
	originInfo.OriginType = originType

	// 查询源站组
	requestGroupQuery := teo.NewDescribeOriginGroupRequest()
	requestGroupQuery.ZoneId = new(p.CustomConfig.ZoneValue) // 站点ID
	requestGroupQuery.Offset = common.Uint64Ptr(0)           // 分页查询偏移量
	requestGroupQuery.Limit = common.Uint64Ptr(1000)         // 分页查询限制数目
	requestGroupQuery.Filters = make([]*teo.AdvancedFilter, 0, 1)
	requestGroupQuery.Filters = append(requestGroupQuery.Filters, &teo.AdvancedFilter{
		Name:   new("origin-group-name"), // 按照源站组名称进行过滤
		Values: common.StringPtrs([]string{p.CustomConfig.IPGroupIDCode}),
	})
	responseGroupQuery, err := p.Provider.Client.DescribeOriginGroupWithContext(ctx, requestGroupQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeOriginGroupWithContext 失败 %w", err)
	}

	// 是否需要创建源站组
	needCreate := true
	// 是否需要更新源站组记录值
	needUpdate := false

	// 检查源站组是否存在和记录值是否匹配
	for _, item := range responseGroupQuery.Response.OriginGroups {
		// 查找源站组名称
		itemGroupName := utils.PtrToString(item.Name)
		if utils.EqualDomainIgnoreCase(itemGroupName, p.CustomConfig.IPGroupIDCode) {

			// 查找记录值列表
			itemOriginValues := make([]string, 0, len(item.Records))
			for _, v := range item.Records {
				itemOriginValues = append(itemOriginValues, utils.PtrToString(v.Record))
			}
			// 查找源站组ID
			originInfo.ID = utils.PtrToString(item.GroupId)

			// 判断记录值列表
			if utils.EqualStringSliceUnordered(itemOriginValues, origin_values) {
				return originInfo, nil
			} else {
				p.logger.ErrorContext(ctx, "源站组记录值列表不匹配",
					slog.String("期望", strings.Join(origin_values, ",")),
					slog.String("实际", strings.Join(itemOriginValues, ",")),
				)
			}

			needCreate = false
			needUpdate = true
		}
	}

	if needCreate {

		// 创建源站组
		requestGroupCreate := teo.NewCreateOriginGroupRequest()
		requestGroupCreate.ZoneId = new(p.CustomConfig.ZoneValue)                     // 站点ID
		requestGroupCreate.Records = make([]*teo.OriginRecord, 0, len(origin_values)) // 源站记录信息
		for _, originValue := range origin_values {
			requestGroupCreate.Records = append(requestGroupCreate.Records, &teo.OriginRecord{
				Record: new(originValue),
				Type:   new("IP_DOMAIN"),
			})
		}
		requestGroupCreate.Name = new(p.CustomConfig.IPGroupIDCode) // 源站组名称
		requestGroupCreate.Type = new("GENERAL")                    // 源站组类型
		responseGroupCreate, err := p.Provider.Client.CreateOriginGroupWithContext(ctx, requestGroupCreate)
		if err != nil {
			return nil, fmt.Errorf("接口 CreateOriginGroupWithContext 失败 %w", err)
		}

		// 查找源站组ID
		originInfo.ID = utils.PtrToString(responseGroupCreate.Response.OriginGroupId)

	} else if needUpdate {

		// 更新源站组
		requestGroupUpdate := teo.NewModifyOriginGroupRequest()
		requestGroupUpdate.ZoneId = new(p.CustomConfig.ZoneValue)                     // 站点ID
		requestGroupUpdate.GroupId = new(originInfo.ID)                               // 源站组ID
		requestGroupUpdate.Records = make([]*teo.OriginRecord, 0, len(origin_values)) // 源站记录信息
		for _, originValue := range origin_values {
			requestGroupUpdate.Records = append(requestGroupUpdate.Records, &teo.OriginRecord{
				Record: new(originValue),
				Type:   new("IP_DOMAIN"),
			})
		}
		_, err = p.Provider.Client.ModifyOriginGroupWithContext(ctx, requestGroupUpdate)
		if err != nil {
			return nil, fmt.Errorf("接口 ModifyOriginGroupWithContext 失败 %w", err)
		}

	}

	return originInfo, nil
}

// 腾讯云 EO 自定义配置
type TencentEOCustomConfig struct {
	ZoneValue           string `json:"zone_value"`                      // 域名编号
	IPv6StatusValue     string `json:"ipv6_status_value,omitempty"`     // IPv6状态
	OriginProtocolValue string `json:"origin_protocol_value,omitempty"` // 回源协议
	HttpPort            uint64 `json:"http_port,omitempty"`             // HTTP回源端口
	HttpsPort           uint64 `json:"https_port,omitempty"`            // HTTPS回源端口
	IPGroupIDCode       string `json:"ip_group_id_code,omitempty"`      // 设备IP组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *TencentEOProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// ZonesRaw 获取所有站点
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeZones
func (p *TencentEOProvider) ZonesRaw(ctx context.Context) (zones []*teo.Zone, err error) {
	request := teo.NewDescribeZonesRequest()
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(100)
	request.Order = new("create-time")
	request.Direction = new("asc")
	response, err := p.Provider.Client.DescribeZonesWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeZonesWithContext 失败 %w", err)
	}
	return response.Response.Zones, nil
}

// OriginGroupsRaw 获取所有源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeOriginGroup
func (p *TencentEOProvider) OriginGroupsRaw(ctx context.Context, zone_id string) (groups []*teo.OriginGroup, err error) {
	request := teo.NewDescribeOriginGroupRequest()
	request.ZoneId = new(zone_id)
	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(100)
	response, err := p.Provider.Client.DescribeOriginGroupWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeOriginGroupWithContext 失败 %w", err)
	}
	return response.Response.OriginGroups, nil
}

// DnsRecordsRaw 获取所有DNS记录
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeDnsRecords
func (p *TencentEOProvider) DnsRecordsRaw(ctx context.Context, zone_id string) (records []*teo.DnsRecord, err error) {
	request := teo.NewDescribeDnsRecordsRequest()
	request.ZoneId = new(zone_id)
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(1000)
	request.SortBy = new("created-on")
	request.SortOrder = new("asc")
	response, err := p.Provider.Client.DescribeDnsRecordsWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeDnsRecordsWithContext 失败 %w", err)
	}
	return response.Response.DnsRecords, nil
}

// AccelerationDomainsRaw 获取所有加速域名
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
func (p *TencentEOProvider) AccelerationDomainsRaw(ctx context.Context, zone_id string) (records []*teo.AccelerationDomain, err error) {
	request := teo.NewDescribeAccelerationDomainsRequest()
	request.ZoneId = new(zone_id)
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(200)
	request.Order = new("created_on")
	request.Direction = new("asc")
	response, err := p.Provider.Client.DescribeAccelerationDomainsWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeAccelerationDomainsWithContext 失败 %w", err)
	}
	return response.Response.AccelerationDomains, nil
}

// DescribeTimingL7AnalysisData 查询流量分析时序数据
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeTimingL7AnalysisData
func (p *TencentEOProvider) DescribeTimingL7AnalysisData(ctx context.Context, zone_id string) (records []*teo.TimingDataRecord, err error) {
	// 当前时间
	currentTime := gotime.Current()
	// 开始时间
	startTime := currentTime.BeforeDay(30).StartOfDay().SetFormat(time.RFC3339)
	// 结束时间
	endTime := currentTime.SetFormat(time.RFC3339)

	request := teo.NewDescribeTimingL7AnalysisDataRequest()
	request.StartTime = new(startTime)
	request.EndTime = new(endTime)
	request.MetricNames = common.StringPtrs([]string{
		"l7Flow_outFlux",                  // L7 EdgeOne 响应流量，单位：Byte
		"l7Flow_inFlux",                   // L7 客户端请求流量，单位：Byte
		"l7Flow_flux",                     // L7 访问总流量（EdgeOne 响应+客户端请求），单位：Byte
		"l7Flow_outBandwidth",             // L7 EdgeOne 响应带宽，单位：bps
		"l7Flow_inBandwidth",              // L7 客户端请求带宽，单位：bps
		"l7Flow_bandwidth",                // L7 访问总带宽（EdgeOne 响应+客户端请求），单位：bps
		"l7Flow_request",                  // L7 访问请求数，单位：次
		"l7Flow_avgResponseTime",          // L7 访问平均响应耗时，单位：ms
		"l7Flow_avgFirstByteResponseTime", // L7 访问平均首字节响应耗时，单位：ms
	})
	request.ZoneIds = common.StringPtrs([]string{zone_id})
	request.Interval = new("day")
	response, err := p.Provider.Client.DescribeTimingL7AnalysisDataWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeTimingL7AnalysisDataWithContext 失败 %w", err)
	}
	return response.Response.Data, nil
}

// DescribeTimingL4Data 查询四层流量时序数据
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeTimingL7AnalysisData
func (p *TencentEOProvider) DescribeTimingL4Data(ctx context.Context, zone_id string) (records []*teo.TimingDataRecord, err error) {
	// 当前时间
	currentTime := gotime.Current()
	// 开始时间
	startTime := currentTime.BeforeDay(30).StartOfDay().SetFormat(time.RFC3339)
	// 结束时间
	endTime := currentTime.SetFormat(time.RFC3339)

	request := teo.NewDescribeTimingL4DataRequest()
	request.StartTime = new(startTime)
	request.EndTime = new(endTime)
	request.MetricNames = common.StringPtrs([]string{
		"l4Flow_flux",               // L4 访问总流量，单位：Byte
		"l4Flow_inFlux",             // L4 访问入流量，单位：Byte
		"l4Flow_outFlux",            // L4 访问出流量，单位：Byte
		"l4Flow_inBandwidth",        // L4 访问入向带宽峰值，单位：bps
		"l4Flow_outBandwidth",       // L4 访问出向带宽峰值，单位：bps
		"l4Flow_connections",        // L4 访问并发连接数，单位：个
		"l4Flow_newConnectionsRate", // L4 新建连接数速率，单位：个/秒，指标值类型： Float，保留两位小数
	})
	request.ZoneIds = common.StringPtrs([]string{zone_id})
	request.Interval = new("day")
	response, err := p.Provider.Client.DescribeTimingL4DataWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeTimingL4DataWithContext 失败 %w", err)
	}
	return response.Response.Data, nil
}
