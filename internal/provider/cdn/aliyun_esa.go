package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"vibe_ip/internal/helper/aliyun_helper"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gomd5"
)

// AliyunESAConfig 阿里云 ESA 配置
type AliyunESAConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret"`  // 访问密钥Key
	SiteID          int64  `json:"site_id"`            // 站点ID
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证阿里云 ESA 配置
func (cfg *AliyunESAConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("AccessKeySecret 不能为空")
	}
	if cfg.SiteID == 0 {
		return fmt.Errorf("SiteID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultAliyunESAConfig 创建默认的阿里云 ESA 配置
func NewDefaultAliyunESAConfig() *AliyunESAConfig {
	return &AliyunESAConfig{
		Endpoint: "esa.cn-hangzhou.aliyuncs.com",
	}
}

// ParseCustomConfig 解析 阿里云 ESA 自定义配置
func (p *AliyunESAConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// AliyunESAProvider 阿里云 ESA 提供程序
type AliyunESAProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *AliyunESAConfig
	opts   *cdnOptions
	logger *slog.Logger

	CustomConfig *AliyunESACustomConfig
	Provider     struct {
		Auth    credential.Credential
		Cfg     *openapi.Config
		Client  *esa20240910.Client
		Runtime *util.RuntimeOptions
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData esa20240910.ListRecordsResponseBodyRecords
}

// NewAliyunESAProvider 创建阿里云 ESA 提供程序
func NewAliyunESAProvider(ctx context.Context, logger *slog.Logger, cfg *AliyunESAConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewAliyunESAProviderRaw(ctx, logger, cfg, opts...)
}

// NewAliyunESAProviderRaw 创建阿里云 ESA 提供程序
func NewAliyunESAProviderRaw(ctx context.Context, logger *slog.Logger, cfg *AliyunESAConfig, opts ...CdnOption) (*AliyunESAProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &AliyunESAProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
		logger:       logger,
		CustomConfig: &AliyunESACustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth, err = credential.NewCredential(&credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(p.cfg.AccessKeyID),
		AccessKeySecret: tea.String(p.cfg.AccessKeySecret),
	})
	if err != nil {
		return nil, fmt.Errorf("创建凭证失败 %w", err)
	}

	// 选项
	p.Provider.Cfg = &openapi.Config{
		Credential: p.Provider.Auth,
		Endpoint:   tea.String(p.cfg.Endpoint),
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		httpHandler := aliyun_helper.NewHttpHandler(p.opts.httpTransport)
		p.Provider.Cfg.HttpClient = httpHandler
	}

	// 实例化
	p.Provider.Client, err = esa20240910.NewClient(p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 运行时
	p.Provider.Runtime = &dara.RuntimeOptions{
		Autoretry:   tea.Bool(true),
		MaxAttempts: tea.Int(3),
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *AliyunESAProvider) Name() string {
	return "阿里云 ESA"
}

// printData 打印数据
func (p *AliyunESAProvider) printData(ctx context.Context) {
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
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListRecords
func (p *AliyunESAProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 源站信息
	originValues := make([]string, 0)
	if len(newOriginValues) == 1 {
		originValues = newOriginValues
	} else {
		// 获取源站组信息
		responseOriginInfo, err := p.getGroupOrigin(ctx, newOriginValues)
		if err != nil {
			return nil, fmt.Errorf("查询源站组失败: %w", err)
		}
		originValues = append(originValues, responseOriginInfo.Cname)
	}

	// 查询加速域名
	requestCdnQuery := &esa20240910.ListRecordsRequest{}
	requestCdnQuery.SiteId = tea.Int64(p.cfg.SiteID) // 站点ID
	requestCdnQuery.RecordName = tea.String(fullDomain)
	requestCdnQuery.PageNumber = tea.Int32(1)
	requestCdnQuery.PageSize = tea.Int32(500)
	responseCdnQuery, err := p.Provider.Client.ListRecordsWithContext(ctx, requestCdnQuery, p.Provider.Runtime)
	if err != nil {
		return
	}
	// 响应参数
	if responseCdnQuery.GetBody().GetRecords() == nil || len(responseCdnQuery.GetBody().GetRecords()) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseCdnQuery.GetBody().GetRecords() {

		// 查找域名
		itemFullDomain := tea.StringValue(item.GetRecordName())

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
			itemOriginValues = append(itemOriginValues, tea.StringValue(item.GetData().GetValue()))
			// 查找用于标记DNS记录的业务场景
			itemBizName := tea.StringValue(item.GetBizName())
			// 查找记录的备注信息
			itemComment := tea.StringValue(item.GetComment())

			// 域名ID
			p.InternalData.RecordID = cast.ToString(tea.Int64Value(item.GetRecordId()))

			// CDN域名
			p.InternalData.CdnDomain = utils.TrimDots(tea.StringValue(item.GetRecordCname()))

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, originValues) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", originValues, itemOriginValues)
				return p.InternalData, nil
			}

			// 判断用于标记DNS记录的业务场景
			if !utils.EqualDomainIgnoreCase(itemBizName, p.CustomConfig.BizName) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("用于标记DNS记录的业务场景不匹配，期望 %v，实际 %v", p.CustomConfig.BizName, itemBizName)
				return p.InternalData, nil
			}

			// 判断记录的备注信息
			if !utils.EqualDomainIgnoreCase(itemComment, p.CustomConfig.Comment) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录的备注信息不匹配，期望 %v，实际 %v", p.CustomConfig.Comment, itemComment)
				return p.InternalData, nil
			}

			// 查询回源规则
			err = p.getRuleInfo(ctx, domain, subdomain)
			if err != nil {
				p.InternalData.IsContinue = true
				return p.InternalData, fmt.Errorf("查询回源规则失败: %w", err)
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
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateRecord
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateRecord
func (p *AliyunESAProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 打印数据
	p.printData(ctx)

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 转大写
	recordType = strings.ToUpper(recordType)

	// 记录的DNS类型
	Type := ""

	// CNAME记录的源站类型
	sourceType := ""

	// 记录值或部分内容
	originValues := make([]string, 0)

	switch recordType {
	case "A", "AAAA":
		Type = "A/AAAA"
		sourceType = ""
		originValues = newOriginValues
	default:
		Type = recordType
	}

	// 如果是多值记录
	if len(originValues) > 1 {
		// 获取源站组信息
		responseOriginInfo, err := p.getGroupOrigin(ctx, newOriginValues)
		if err != nil {
			return fmt.Errorf("查询源站组失败: %w", err)
		}
		Type = "CNAME"
		sourceType = responseOriginInfo.SourceType
		originValues = []string{responseOriginInfo.Cname}
	}

	if p.InternalData.RecordID == "" {

		// 创建加速域名
		requestCdnCreate := &esa20240910.CreateRecordRequest{}
		requestCdnCreate.SiteId = tea.Int64(p.cfg.SiteID)             // 站点ID
		requestCdnCreate.RecordName = tea.String(fullDomain)          // 记录名称
		requestCdnCreate.Proxied = tea.Bool(true)                     // 记录是否开启代理加速
		requestCdnCreate.Type = tea.String(Type)                      // 记录的DNS类型
		requestCdnCreate.SourceType = tea.String(sourceType)          // CNAME记录的源站类型
		requestCdnCreate.BizName = tea.String(p.CustomConfig.BizName) // 用于标记DNS记录的业务场景
		requestCdnCreate.Ttl = tea.Int32(1)                           // 记录的过期时间
		requestCdnCreate.Data = &esa20240910.CreateRecordRequestData{
			Value: tea.String(strings.Join(originValues, ",")), // 记录值或部分内容
		}
		requestCdnCreate.Comment = tea.String(p.CustomConfig.Comment) // 记录的备注信息
		responseCdnCreate, err := p.Provider.Client.CreateRecordWithContext(ctx, requestCdnCreate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 CreateRecordWithContext 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = cast.ToString(tea.Int64Value(responseCdnCreate.GetBody().GetRecordId()))

		// 查询回源规则
		err = p.getRuleInfo(ctx, domain, subdomain)
		if err != nil {
			return fmt.Errorf("查询回源规则失败: %w", err)
		}

	} else if p.InternalData.RecordID != "" {

		// 更新加速域名
		requestCdnUpdate := &esa20240910.UpdateRecordRequest{}
		requestCdnUpdate.RecordId = tea.Int64(cast.ToInt64(p.InternalData.RecordID)) // 站点ID
		requestCdnUpdate.Ttl = tea.Int32(1)                                          // 记录的过期时间
		requestCdnUpdate.Proxied = tea.Bool(true)                                    // 记录是否开启代理加速
		requestCdnUpdate.Type = tea.String(Type)                                     // 记录的DNS类型
		requestCdnUpdate.Data = &esa20240910.UpdateRecordRequestData{
			Value: tea.String(strings.Join(originValues, ",")), // 记录值或部分内容
		}
		requestCdnUpdate.SourceType = tea.String(sourceType)          // CNAME记录的源站类型
		requestCdnUpdate.BizName = tea.String(p.CustomConfig.BizName) // 记录加速时的业务场景
		requestCdnUpdate.Comment = tea.String(p.CustomConfig.Comment) // 记录的备注信息
		_, err = p.Provider.Client.UpdateRecordWithContext(ctx, requestCdnUpdate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 UpdateRecordWithContext 失败 %w", err)
		}

		// 查询回源规则
		err = p.getRuleInfo(ctx, domain, subdomain)
		if err != nil {
			return fmt.Errorf("查询回源规则失败: %w", err)
		}

	}

	return nil
}

// AliyunESAProviderGroupOriginInfo 阿里云 ESA 源站组信息
type AliyunESAProviderGroupOriginInfo struct {
	Cname      string   // 源站组别名
	List       []string // 源站组记录值列表
	SourceType string   // 类型
}

// getGroupOrigin 获取源站组信息
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListOriginPools
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateOriginPool
// https://next.api.aliyun.com/api/ESA/2024-09-10/GetOriginPool
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateOriginPool
func (p *AliyunESAProvider) getGroupOrigin(ctx context.Context, newOriginValues []string) (originInfo *AliyunESAProviderGroupOriginInfo, err error) {

	const sourceType = "OP"
	originInfo = new(AliyunESAProviderGroupOriginInfo)
	originInfo.List = newOriginValues
	originInfo.SourceType = sourceType

	// 查询源站组
	requestGroupQuery := &esa20240910.ListOriginPoolsRequest{}
	requestGroupQuery.SiteId = tea.Int64(p.cfg.SiteID)              // 站点ID
	requestGroupQuery.Name = tea.String(p.CustomConfig.IPGroupCode) // 源地址池名称
	requestGroupQuery.MatchType = tea.String("exact")               // 按名称匹配查询类型
	responseGroupQuery, err := p.Provider.Client.ListOriginPoolsWithContext(ctx, requestGroupQuery, p.Provider.Runtime)
	if err != nil {
		return nil, fmt.Errorf("接口 ListOriginPoolsWithContext 失败 %w", err)
	}

	// 是否需要创建源站组
	needCreate := true
	// 是否需要更新源站组记录值
	needUpdate := false

	// 检查源站组是否存在和记录值是否匹配
	for _, item := range responseGroupQuery.GetBody().GetOriginPools() {
		// 查找源站组名称
		itemGroupName := tea.StringValue(item.GetName())
		if utils.EqualDomainIgnoreCase(itemGroupName, p.CustomConfig.IPGroupCode) {

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
			for _, origin := range item.GetOrigins() {
				itemOriginValues = append(itemOriginValues, tea.StringValue(origin.GetAddress()))
			}
			// 查找源站组别名
			itemCname := tea.StringValue(item.GetRecordName())

			// 判断记录值列表
			if utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				originInfo.Cname = itemCname
				return originInfo, nil
			}

			needCreate = false
			needUpdate = true
		}
	}

	if needCreate {

		// 创建源站组
		requestGroupCreate := &esa20240910.CreateOriginPoolRequest{}
		requestGroupCreate.Name = tea.String(p.CustomConfig.IPGroupCode)                    // 列表名称
		requestGroupCreate.Enabled = tea.Bool(true)                                         // 源地址池是否启用
		requestGroupCreate.SiteId = tea.Int64(p.cfg.SiteID)                                 // 站点ID
		requestGroupCreate.Origins = make([]*esa20240910.CreateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
		for _, item := range newOriginValues {
			requestGroupCreate.Origins = append(requestGroupCreate.Origins, &esa20240910.CreateOriginPoolRequestOrigins{
				Address: tea.String(item),
				Type:    tea.String("ip_domain"),
				Name:    tea.String(gomd5.Get16MD5Encode(item)),
				Enabled: tea.Bool(true),
			})
		}
		responseGroupCreate, err := p.Provider.Client.CreateOriginPoolWithContext(ctx, requestGroupCreate, p.Provider.Runtime)
		if err != nil {
			return nil, fmt.Errorf("接口 CreateOriginPoolWithContext 失败 %w", err)
		}

		// 查询单个源站组
		requestGroupDetail := &esa20240910.GetOriginPoolRequest{}
		requestGroupDetail.Id = responseGroupCreate.GetBody().GetId() // 源地址池ID
		requestGroupDetail.SiteId = tea.Int64(p.cfg.SiteID)           // 站点ID
		responseGroupetail, err := p.Provider.Client.GetOriginPoolWithContext(ctx, requestGroupDetail, p.Provider.Runtime)
		if err != nil {
			return nil, fmt.Errorf("接口 GetOriginPoolWithContext 失败 %w", err)
		}

		// 源站组别名
		originInfo.Cname = tea.StringValue(responseGroupetail.GetBody().GetRecordName())

	} else if needUpdate {

		// 更新源站组
		requestGroupUpdate := &esa20240910.UpdateOriginPoolRequest{}
		requestGroupUpdate.Id = tea.Int64(cast.ToInt64(p.CustomConfig.IPGroupCode))         // 源站组ID
		requestGroupUpdate.Enabled = tea.Bool(true)                                         // 源地址池是否启用
		requestGroupUpdate.Origins = make([]*esa20240910.UpdateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
		for _, item := range newOriginValues {
			requestGroupUpdate.Origins = append(requestGroupUpdate.Origins, &esa20240910.UpdateOriginPoolRequestOrigins{
				Name:    tea.String(gomd5.Get16MD5Encode(item)),
				Address: tea.String(item),
				Type:    tea.String("ip_domain"),
				Enabled: tea.Bool(true),
			})
		}
		requestGroupUpdate.SiteId = tea.Int64(p.cfg.SiteID) // 站点ID
		_, err = p.Provider.Client.UpdateOriginPoolWithContext(ctx, requestGroupUpdate, p.Provider.Runtime)
		if err != nil {
			return nil, fmt.Errorf("接口 UpdateOriginPoolWithContext 失败 %w", err)
		}

	}

	return originInfo, nil
}

// getGroupOrigin 获取回源规则信息
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListOriginRules
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateOriginRule
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateOriginRule
func (p *AliyunESAProvider) getRuleInfo(ctx context.Context, domain string, subdomain string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 配置类型
	const configType = "rule"

	//配置ID
	var configID int64

	// 规则名称
	var ruleName = fmt.Sprintf("%v_%v", p.CustomConfig.OriginHttpPort, p.CustomConfig.OriginHttpsPort)

	// 查询回源规则
	requestRuleQuery := &esa20240910.ListOriginRulesRequest{}
	requestRuleQuery.SiteId = tea.Int64(p.cfg.SiteID)    // 站点ID
	requestRuleQuery.ConfigType = tea.String(configType) // 配置类型
	requestRuleQuery.PageNumber = tea.Int32(1)           // 第几页
	requestRuleQuery.PageSize = tea.Int32(500)           // 每页几条
	requestRuleQuery.RuleName = tea.String(ruleName)     // 回源规则名称
	responseRuleQuery, err := p.Provider.Client.ListOriginRulesWithContext(ctx, requestRuleQuery, p.Provider.Runtime)
	if err != nil {
		return fmt.Errorf("接口 ListOriginRulesWithContext 失败 %w", err)
	}

	// 是否需要创建回源规则
	needCreate := true
	// 是否需要更新回源规则配置
	needUpdate := false

	// 检查源站组是否存在和记录值是否匹配
	for _, item := range responseRuleQuery.GetBody().GetConfigs() {
		// 查找规则名称
		itemRuleName := tea.StringValue(item.GetRuleName())
		if utils.EqualDomainIgnoreCase(itemRuleName, ruleName) {

			// 查找配置ID
			itemConfigID := tea.Int64Value(item.GetConfigId())
			// 查找HTTP端口
			itemHttpPort := tea.StringValue(item.GetOriginHttpPort())
			// 查找HTTPS端口
			itemHttpsPort := tea.StringValue(item.GetOriginHttpsPort())
			// 查找回源规则配置
			itemRule := tea.StringValue(item.GetRule())
			// 解析出所有域名
			itemDomains := ExtractDomainsFromRule(itemRule)

			// 配置ID
			configID = itemConfigID

			// 检查HTTP端口是否匹配
			if !utils.EqualStringIgnoreCase(itemHttpPort, cast.ToString(p.CustomConfig.OriginHttpPort)) {
				needCreate = false
				needUpdate = true
				continue
			}

			// 检查HTTPS端口是否匹配
			if !utils.EqualStringIgnoreCase(itemHttpsPort, cast.ToString(p.CustomConfig.OriginHttpsPort)) {
				needCreate = false
				needUpdate = true
				continue
			}

			// 判断是否包含域名
			for _, itemDomain := range itemDomains {
				if utils.EqualDomainIgnoreCase(itemDomain, fullDomain) {
					return nil
				} else {
					needCreate = false
					needUpdate = true
				}
			}

			needCreate = false
			needUpdate = true
		}
	}

	if needCreate {

		// 创建回源规则
		requestRuleCreate := &esa20240910.CreateOriginRuleRequest{}
		requestRuleCreate.RuleName = tea.String(ruleName)                                             // 规则名称
		requestRuleCreate.SiteId = tea.Int64(p.cfg.SiteID)                                            // 站点ID
		requestRuleCreate.RuleEnable = tea.String("on")                                               // 规则开关
		requestRuleCreate.Rule = tea.String(AddDomainToRule("", fullDomain))                          // 规则内容
		requestRuleCreate.OriginScheme = tea.String(p.CustomConfig.OriginScheme)                      // 回源请求使用的协议
		requestRuleCreate.OriginHttpPort = tea.String(cast.ToString(p.CustomConfig.OriginHttpsPort))  // 使用HTTP协议回源时访问的源站端口
		requestRuleCreate.OriginHttpsPort = tea.String(cast.ToString(p.CustomConfig.OriginHttpsPort)) // 使用HTTPS协议回源时访问的源站端口
		_, err = p.Provider.Client.CreateOriginRuleWithContext(ctx, requestRuleCreate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 CreateOriginRuleWithContext 失败 %w", err)
		}

	} else if needUpdate {

		// 更新回源规则
		requestRuleUpdate := &esa20240910.UpdateOriginRuleRequest{}
		requestRuleUpdate.ConfigId = tea.Int64(configID)                                              // 配置ID
		requestRuleUpdate.SiteId = tea.Int64(p.cfg.SiteID)                                            // 站点ID
		requestRuleUpdate.RuleEnable = tea.String("on")                                               // 规则开关
		requestRuleUpdate.Rule = tea.String(AddDomainToRule("", fullDomain))                          // 规则内容
		requestRuleUpdate.OriginScheme = tea.String(p.CustomConfig.OriginScheme)                      // 回源请求使用的协议
		requestRuleUpdate.OriginHttpPort = tea.String(cast.ToString(p.CustomConfig.OriginHttpsPort))  // 使用HTTP协议回源时访问的源站端口
		requestRuleUpdate.OriginHttpsPort = tea.String(cast.ToString(p.CustomConfig.OriginHttpsPort)) // 使用HTTPS协议回源时访问的源站端口
		requestRuleUpdate.RuleName = tea.String(ruleName)                                             // 规则名称
		_, err = p.Provider.Client.UpdateOriginRuleWithContext(ctx, requestRuleUpdate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 UpdateOriginRuleWithContext 失败 %w", err)
		}

	}

	return nil
}

// ExtractDomainsFromRule 从 Rule 字符串中提取所有域名
func ExtractDomainsFromRule(rule string) []string {
	// 匹配 {"..."} 中的内容
	re := regexp.MustCompile(`\{([^}]*)\}`)
	matches := re.FindStringSubmatch(rule)
	if len(matches) < 2 {
		return nil
	}

	// 提取引号内的内容（支持 \" 转义，但你的例子中是普通双引号）
	content := matches[1]
	// 使用正则匹配所有被双引号包围的字符串
	domainRe := regexp.MustCompile(`"([^"]+)"`)
	allMatches := domainRe.FindAllStringSubmatch(content, -1)

	var domains []string
	for _, m := range allMatches {
		if len(m) > 1 {
			domains = append(domains, m[1])
		}
	}
	return domains
}

// AddDomainToRule 向 Rule 中添加域名（如果不存在）
func AddDomainToRule(rule, newDomain string) string {
	domains := ExtractDomainsFromRule(rule)
	if domains == nil {
		// 如果无法解析，保守返回原 rule
		return rule
	}

	// 检查是否已存在
	for _, d := range domains {
		if d == newDomain {
			return rule // 已存在，无需修改
		}
	}

	// 添加新域名
	domains = append(domains, newDomain)

	// 重新构建 Rule
	var quotedDomains []string
	for _, d := range domains {
		quotedDomains = append(quotedDomains, fmt.Sprintf(`"%s"`, d))
	}
	newList := strings.Join(quotedDomains, " ")
	return fmt.Sprintf("(http.host in {%s})", newList)
}

// 阿里云 ESA 自定义配置
type AliyunESACustomConfig struct {
	BizName         string `json:"biz_name"`                // 用于标记DNS记录的业务场景
	Comment         string `json:"comment,omitempty"`       // 记录的备注信息
	OriginScheme    string `json:"origin_scheme"`           // 回源请求使用的协议
	OriginHttpPort  int    `json:"origin_http_port"`        // 使用HTTP协议回源时访问的源站端口
	OriginHttpsPort int    `json:"origin_https_port"`       // 使用HTTPS协议回源时访问的源站端口
	IPGroupCode     string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *AliyunESAProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
