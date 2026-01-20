package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"vibe_ip/internal/helper/aliyun_helper"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gojson"
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
		return err
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
		Auth   credential.Credential
		Cfg    *openapi.Config
		Client *esa20240910.Client
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
		return nil, err
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
		return nil, err
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

	// 请求参数
	request := &esa20240910.ListRecordsRequest{}
	request.SiteId = tea.Int64(p.cfg.SiteID) // 站点ID
	request.RecordName = tea.String(fullDomain)
	request.PageNumber = tea.Int32(1)
	request.PageSize = tea.Int32(500)
	// 请求
	response, err := p.Provider.Client.ListRecordsWithContext(ctx, request, &dara.RuntimeOptions{})
	printJSON("查询加速域名", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Body.Records == nil || len(response.Body.Records) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range response.Body.Records {

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
			p.InternalData.CdnDomain = utils.TrimDotsPrt(item.GetRecordCname())

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", newOriginValues, itemOriginValues)
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
func (p *AliyunESAProvider) Update(ctx context.Context, domain string, subdomain string, originValues []string, recordType string) (err error) {

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

	switch recordType {
	case "A", "AAAA":
		Type = "A/AAAA"
	default:
		Type = recordType
	}

	if p.InternalData.RecordID == "" {
		// 请求参数
		request := &esa20240910.CreateRecordRequest{}
		request.SiteId = tea.Int64(p.cfg.SiteID)             // 站点ID
		request.RecordName = tea.String(fullDomain)          // 记录名称
		request.Proxied = tea.Bool(true)                     // 记录是否开启代理加速
		request.Type = tea.String(Type)                      // 记录的DNS类型
		request.SourceType = tea.String(sourceType)          // CNAME记录的源站类型
		request.BizName = tea.String(p.CustomConfig.BizName) // 用于标记DNS记录的业务场景
		request.Ttl = tea.Int32(1)                           // 记录的过期时间
		request.Data = &esa20240910.CreateRecordRequestData{
			Value: tea.String(strings.Join(originValues, ",")), // 记录值或部分内容
		}
		request.Comment = tea.String(p.CustomConfig.Comment) // 记录的备注信息
		// 请求
		response, err := p.Provider.Client.CreateRecordWithContext(ctx, request, &dara.RuntimeOptions{})
		printJSON("创建加速域名", request, response)
		if err != nil {
			return err
		}
	}

	if p.InternalData.RecordID != "" {
		// 请求参数
		request := &esa20240910.UpdateRecordRequest{}
		request.RecordId = tea.Int64(cast.ToInt64(p.InternalData.RecordID)) // 站点ID
		request.Ttl = tea.Int32(1)                                          // 记录的过期时间
		request.Proxied = tea.Bool(true)                                    // 记录是否开启代理加速
		request.Type = tea.String(Type)                                     // 记录的DNS类型
		request.Data = &esa20240910.UpdateRecordRequestData{
			Value: tea.String(strings.Join(originValues, ",")), // 记录值或部分内容
		}
		request.SourceType = tea.String(sourceType)          // CNAME记录的源站类型
		request.BizName = tea.String(p.CustomConfig.BizName) // 记录加速时的业务场景
		request.Comment = tea.String(p.CustomConfig.Comment) // 记录的备注信息
		// 请求
		response, err := p.Provider.Client.UpdateRecordWithContext(ctx, request, &dara.RuntimeOptions{})
		printJSON("更新加速域名", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 阿里云 ESA 自定义配置
type AliyunESACustomConfig struct {
	BizName        string `json:"biz_name"`                // 用于标记DNS记录的业务场景
	Comment        string `json:"comment,omitempty"`       // 记录的备注信息
	OriginProtocol string `json:"origin_protocol"`         // 源站协议
	HttpPort       int    `json:"http_port"`               // HTTP 端口
	HttpsPort      int    `json:"https_port"`              // HTTPS 端口
	IPGroupCode    string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *AliyunESAProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
