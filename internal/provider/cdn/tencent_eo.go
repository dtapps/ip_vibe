package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
	"go.dtapp.net/library/utils/gojson"
)

// TencentEOConfig 腾讯云 EO 配置
type TencentEOConfig struct {
	Endpoint   string `json:"endpoint,omitempty"` // 接口地址
	SecretID   string `json:"secret_id"`          // 密钥ID
	SecretKey  string `json:"secret_key"`         // 密钥Key
	ZoneID     string `json:"zone_id"`            // 区域ID
	DomainName string `json:"domain_name"`        // 主域名
}

// Validate 验证腾讯云 EO 配置
func (cfg *TencentEOConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.ZoneID == "" {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultTencentEOConfig 创建默认的腾讯云 EO 配置
func NewDefaultTencentEOConfig() *TencentEOConfig {
	return &TencentEOConfig{
		Endpoint: "teo.tencentcloudapi.com",
	}
}

// ParseCustomConfig 解析 腾讯云 EO 自定义配置
func (p *TencentEOConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// TencentEOProvider 腾讯云 EO 提供程序
type TencentEOProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *TencentEOConfig
	opts   *cdnOptions
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
func NewTencentEOProvider(ctx context.Context, logger *slog.Logger, cfg *TencentEOConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewTencentEOProviderRaw(ctx, logger, cfg, opts...)
}

// NewTencentEOProviderRaw 创建腾讯云 EO 提供程序
func NewTencentEOProviderRaw(ctx context.Context, logger *slog.Logger, cfg *TencentEOConfig, opts ...CdnOption) (*TencentEOProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentEOProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
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
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint

	// 实例化
	p.Provider.Client, err = teo.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Client.WithHttpTransport(p.opts.httpTransport.Instance(nil))
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *TencentEOProvider) Name() string {
	return "腾讯云 EO"
}

// printData 打印数据
func (p *TencentEOProvider) printData(ctx context.Context) {
	p.logger.DebugContext(ctx, "数据",
		slog.String("提供商", p.Name()),
		slog.String("内部", gojson.JsonEncodeNoError(p.InternalData)),
	)
	p.logger.DebugContext(ctx, "数据",
		slog.String("提供商", p.Name()),
		slog.String("记录", gojson.JsonEncodeNoError(p.ItemData)),
	)
}

// Check 检查子域名是否存在
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
func (p *TencentEOProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 源站信息
	originValues := make([]string, 0)
	if len(newOriginValues) == 1 {
		originValues = newOriginValues
	} else {
		originInfo, err := p.getGroupOrigin(ctx, newOriginValues)
		if err != nil {
			return nil, fmt.Errorf("查询源站组失败: %w", err)
		}
		originValues = append(originValues, utils.PtrToString(originInfo.Origin))
	}

	// 查询加速域名
	requestCdnQuery := teo.NewDescribeAccelerationDomainsRequest()
	requestCdnQuery.ZoneId = common.StringPtr(p.cfg.ZoneID)
	requestCdnQuery.Offset = common.Int64Ptr(0)
	requestCdnQuery.Limit = common.Int64Ptr(200)
	requestCdnQuery.Filters = make([]*teo.AdvancedFilter, 0)
	requestCdnQuery.Filters = append(requestCdnQuery.Filters, &teo.AdvancedFilter{
		Name:   common.StringPtr("domain-name"),
		Values: common.StringPtrs([]string{fullDomain}),
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

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
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
			if !utils.EqualDomainIgnoreCase(itemOriginProtocol, p.CustomConfig.OriginProtocol) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("回源协议不匹配，期望 %v，实际 %v", p.CustomConfig.OriginProtocol, itemOriginProtocol)
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
			if !utils.EqualDomainIgnoreCase(itemIPv6Status, p.CustomConfig.IPv6Status) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("IPv6状态不匹配，期望 %v，实际 %v", p.CustomConfig.IPv6Status, itemIPv6Status)
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
func (p *TencentEOProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 打印数据
	p.printData(ctx)

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 回源协议，转大写
	p.CustomConfig.OriginProtocol = strings.ToUpper(p.CustomConfig.OriginProtocol)

	// 源站信息
	originInfo := new(teo.OriginInfo{})
	if len(newOriginValues) == 1 {
		originInfo.OriginType = common.StringPtr("IP_DOMAIN")    // 源站类型
		originInfo.Origin = common.StringPtr(newOriginValues[0]) // 源站地址
	} else {
		originInfo, err = p.getGroupOrigin(ctx, newOriginValues)
		if err != nil {
			return fmt.Errorf("查询源站组失败: %w", err)
		}
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
		requestCdnCreate.ZoneId = common.StringPtr(p.cfg.ZoneID)                              // 加速域名所属站点 ID
		requestCdnCreate.DomainName = common.StringPtr(fullDomain)                            // 加速域名
		requestCdnCreate.OriginInfo = originInfo                                              // 源站信息
		requestCdnCreate.OriginProtocol = common.StringPtr(p.CustomConfig.OriginProtocol)     // 回源协议
		requestCdnCreate.HttpOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpPort))   // HTTP回源端口
		requestCdnCreate.HttpsOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpsPort)) // HTTPS回源端口
		requestCdnCreate.IPv6Status = common.StringPtr(p.CustomConfig.IPv6Status)             // IPv6状态
		_, err = p.Provider.Client.CreateAccelerationDomainWithContext(ctx, requestCdnCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateAccelerationDomainWithContext 失败 %w", err)
		}
	}

	if p.InternalData.CdnDomain != "" {
		// 更新加速域名
		requestCdnUpdate := teo.NewModifyAccelerationDomainRequest()
		requestCdnUpdate.ZoneId = common.StringPtr(p.cfg.ZoneID)                              // 加速域名所属站点 ID
		requestCdnUpdate.DomainName = common.StringPtr(fullDomain)                            // 加速域名
		requestCdnUpdate.OriginInfo = originInfo                                              // 源站信息
		requestCdnUpdate.OriginProtocol = common.StringPtr(p.CustomConfig.OriginProtocol)     // 回源协议
		requestCdnUpdate.HttpOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpPort))   // HTTP回源端口
		requestCdnUpdate.HttpsOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpsPort)) // HTTPS回源端口
		requestCdnUpdate.IPv6Status = common.StringPtr(p.CustomConfig.IPv6Status)             // IPv6状态
		_, err = p.Provider.Client.ModifyAccelerationDomainWithContext(ctx, requestCdnUpdate)
		if err != nil {
			return fmt.Errorf("接口 ModifyAccelerationDomainWithContext 失败 %w", err)
		}
	}

	return nil
}

// getGroupOrigin 获取源站组信息
func (p *TencentEOProvider) getGroupOrigin(ctx context.Context, newOriginValues []string) (originInfo *teo.OriginInfo, err error) {

	const originType = "ORIGIN_GROUP"
	originInfo = new(teo.OriginInfo{})
	originInfo.OriginType = common.StringPtr(originType)

	// 查询源站组
	requestGroupQuery := teo.NewDescribeOriginGroupRequest()
	requestGroupQuery.ZoneId = common.StringPtr(p.cfg.ZoneID) // 站点ID
	requestGroupQuery.Offset = common.Uint64Ptr(0)            // 分页查询偏移量
	requestGroupQuery.Limit = common.Uint64Ptr(1000)          // 分页查询限制数目
	requestGroupQuery.Filters = make([]*teo.AdvancedFilter, 0)
	requestGroupQuery.Filters = append(requestGroupQuery.Filters, &teo.AdvancedFilter{
		Name:   common.StringPtr("origin-group-name"), // 按照源站组名称进行过滤
		Values: common.StringPtrs([]string{p.CustomConfig.IPGroupCode}),
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
		if utils.EqualDomainIgnoreCase(itemGroupName, p.CustomConfig.IPGroupCode) {

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
			for _, v := range item.Records {
				itemOriginValues = append(itemOriginValues, utils.PtrToString(v.Record))
			}
			// 查找源站组ID
			itemGroupId := item.GroupId // 源站地址

			// 判断记录值列表
			if utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				originInfo.Origin = itemGroupId // 源站地址
				return originInfo, nil
			}

			needCreate = false
			needUpdate = true
		}
	}

	if needCreate {
		// 创建源站组
		requestGroupCreate := teo.NewCreateOriginGroupRequest()
		requestGroupCreate.ZoneId = common.StringPtr(p.cfg.ZoneID) // 站点ID
		requestGroupCreate.Records = make([]*teo.OriginRecord, 0)  // 源站记录信息
		for _, item := range newOriginValues {
			requestGroupCreate.Records = append(requestGroupCreate.Records, &teo.OriginRecord{
				Record: common.StringPtr(item),
				Type:   common.StringPtr("IP_DOMAIN"),
			})
		}
		requestGroupCreate.Name = common.StringPtr(p.CustomConfig.IPGroupCode) // 源站组名称
		requestGroupCreate.Type = common.StringPtr("GENERAL")                  // 源站组类型
		responseGroupCreate, err := p.Provider.Client.CreateOriginGroupWithContext(ctx, requestGroupCreate)
		if err != nil {
			return nil, fmt.Errorf("接口 CreateOriginGroupWithContext 失败 %w", err)
		}
		originInfo.Origin = responseGroupCreate.Response.OriginGroupId // 源站地址
	}
	if needUpdate {
		// 更新源站组
		requestGroupUpdate := teo.NewModifyOriginGroupRequest()
		requestGroupUpdate.ZoneId = common.StringPtr(p.cfg.ZoneID)                // 站点ID
		requestGroupUpdate.GroupId = common.StringPtr(p.CustomConfig.IPGroupCode) // 源站组ID
		requestGroupUpdate.Records = make([]*teo.OriginRecord, 0)                 // 源站记录信息
		for _, item := range newOriginValues {
			requestGroupUpdate.Records = append(requestGroupUpdate.Records, &teo.OriginRecord{
				Record: common.StringPtr(item),
				Type:   common.StringPtr("IP_DOMAIN"),
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
	OriginProtocol string `json:"origin_protocol"`         // 回源协议
	HttpPort       uint64 `json:"http_port,omitempty"`     // HTTP回源端口
	HttpsPort      uint64 `json:"https_port,omitempty"`    // HTTPS回源端口
	IPv6Status     string `json:"ipv6_status,omitempty"`   // IPv6状态
	IPGroupCode    string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *TencentEOProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
