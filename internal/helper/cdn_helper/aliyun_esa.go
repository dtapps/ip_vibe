package cdn_helper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"vibe_ip/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gomd5"
)

// AliyunEsaHttpHandler 适配 github.com/alibabacloud-go/tea/tea 里的 HttpClient 接口
type AliyunEsaHttpHandler struct {
	httpClient *http.Client
}

func NewAliyunEsaHttpHandler(log *http_log.LoggingTransport) *AliyunEsaHttpHandler {
	return &AliyunEsaHttpHandler{
		httpClient: &http.Client{
			Transport: log,
		},
	}
}

// Call 实现了 dara.HttpClient 接口要求的方法
func (a *AliyunEsaHttpHandler) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	return a.httpClient.Do(request)
}

type AliyunEsaConfig struct {
	AccessKeyID     string `json:"access_key_id"`     // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret"` // 访问密钥Key
	SiteID          int64  `json:"site_id"`           // 站点ID
	DomainName      string `json:"domain_name"`       // 主域名
}

type AliyunEsaClient struct {
	config       *AliyunEsaConfig // 配置
	debug        bool             // 是否开启调试模式
	credential   credential.Credential
	clientConfig *openapi.Config
	client       *esa20240910.Client
}

func NewAliyunEsaClient(ctx context.Context, config *AliyunEsaConfig) (*AliyunEsaClient, error) {

	var err error

	c := &AliyunEsaClient{}
	c.config = config

	// 密钥信息
	c.credential, err = credential.NewCredential(&credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(c.config.AccessKeyID),
		AccessKeySecret: tea.String(c.config.AccessKeySecret),
	})
	if err != nil {
		return nil, err
	}

	// 实例化client选项
	c.clientConfig = &openapi.Config{
		Credential: c.credential,
		Endpoint:   tea.String("esa.cn-hangzhou.aliyuncs.com"),
	}

	// 实例化client对象
	c.client, err = esa20240910.NewClient(c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *AliyunEsaClient) SetDebug(debug bool) *AliyunEsaClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *AliyunEsaClient) SetLog(loggingTransport *http_log.LoggingTransport) *AliyunEsaClient {
	if loggingTransport == nil {
		return c
	}
	// 实例化client选项
	clientConfig := c.clientConfig
	clientConfig.HttpClient = NewAliyunEsaHttpHandler(loggingTransport)
	// 实例化client对象
	client, err := esa20240910.NewClient(clientConfig)
	if err != nil {
		return c
	}
	c.clientConfig = clientConfig
	c.client = client
	return c
}

// 查询加速域名
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListRecords
// fullDomain: 完整域名
func (c *AliyunEsaClient) QueryDomain(ctx context.Context, fullDomain string) (AliyunEsaQueryDomainResponse, error) {

	// 请求
	requestParams := &esa20240910.ListRecordsRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	requestParams.RecordName = tea.String(fullDomain)
	requestParams.PageNumber = tea.Int32(1)
	requestParams.PageSize = tea.Int32(500)
	response, err := c.client.ListRecords(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return AliyunEsaQueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return AliyunEsaQueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询加速域名", requestParams, response)
	}

	for _, item := range response.Body.Records {
		if tea.StringValue(item.RecordName) == fullDomain {
			return AliyunEsaQueryDomainResponse{
				RecordID:    cast.ToString(tea.Int64Value(item.RecordId)),
				FullDomain:  tea.StringValue(item.RecordName),
				DomainCname: utils.TrimDots(tea.StringValue(item.RecordCname)),
				RecordType:  tea.StringValue(item.RecordType),
				OriginValue: tea.StringValue(item.Data.Value),
			}, nil
		}
	}

	return AliyunEsaQueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateRecord
// fullDomain: 完整域名
// recordType: A/AAAA
// recordType: CNAME
// originValue: 源站地址或源站组ID
func (c *AliyunEsaClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &esa20240910.CreateRecordRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	requestParams.RecordName = tea.String(fullDomain) // 记录名称
	requestParams.Data = &esa20240910.CreateRecordRequestData{
		Value: tea.String(originValue), // 记录值或部分内容
	}
	requestParams.Type = tea.String(recordType) // 记录的DNS类型
	requestParams.Ttl = tea.Int32(1)            // 记录的过期时间
	requestParams.Proxied = tea.Bool(true)      // 记录是否开启代理加速
	requestParams.BizName = tea.String("web")   // 用于标记DNS记录的业务场景
	response, err := c.client.CreateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建加速域名", requestParams, response)
	}

	return nil
}

// 更新加速域名
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateRecord
// domainID: 域名ID
// recordType: A/AAAA = IPV4、IPV6 或域名类型源站
// recordType: CNAME = 源站组类型源站
// originValue: 源站地址或源站组ID
func (c *AliyunEsaClient) UpdateDomain(ctx context.Context, domainID string, recordType string, originValue string) error {

	// 请求
	requestParams := &esa20240910.UpdateRecordRequest{}
	requestParams.RecordId = tea.Int64(cast.ToInt64(domainID)) // 站点ID
	requestParams.Data = &esa20240910.UpdateRecordRequestData{
		Value: tea.String(originValue), // 记录值或部分内容
	}
	requestParams.Type = tea.String(recordType) // 记录的DNS类型
	requestParams.Ttl = tea.Int32(1)            // 记录的过期时间
	requestParams.Proxied = tea.Bool(true)      // 记录是否开启代理加速
	response, err := c.client.UpdateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新加速域名", requestParams, response)
	}

	return nil
}

// 查询源站组
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateList
// originGroupName: 源站组名称
func (c *AliyunEsaClient) QueryOriginGroup(ctx context.Context, originGroupName string) (AliyunEsaQueryOriginGroupResponse, error) {

	// 请求
	requestParams := &esa20240910.ListOriginPoolsRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	requestParams.Name = tea.String(originGroupName)  // 源地址池名称
	requestParams.MatchType = tea.String("exact")     // 按名称匹配查询类型
	response, err := c.client.ListOriginPools(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return AliyunEsaQueryOriginGroupResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return AliyunEsaQueryOriginGroupResponse{}, err
	}
	if c.debug {
		printJSON("查询源站组", requestParams, response)
	}

	for _, item := range response.Body.OriginPools {
		if strings.EqualFold(tea.StringValue(item.Name), originGroupName) {
			if tea.BoolValue(item.Enabled) {
				origins := []string{}
				for _, subItem := range item.Origins {
					origins = append(origins, tea.StringValue(subItem.Address))
				}
				return AliyunEsaQueryOriginGroupResponse{
					GroupID:      cast.ToString(tea.Int64Value(item.Id)),
					GroupName:    tea.StringValue(item.Name),
					GroupCname:   tea.StringValue(item.RecordName),
					GroupOrigins: origins,
				}, nil
			}
		}
	}

	return AliyunEsaQueryOriginGroupResponse{}, fmt.Errorf("未找到记录")
}

// 创建源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateOriginGroup
// originGroupName: 源站组名称
// originGroupOrigins: 源站组值列表
func (c *AliyunEsaClient) CreateOriginGroup(ctx context.Context, originGroupName string, originGroupOrigins []string) error {

	// 请求
	requestParams := &esa20240910.CreateOriginPoolRequest{}
	requestParams.Name = tea.String(originGroupName)                               // 列表名称
	requestParams.Enabled = tea.Bool(true)                                         // 源地址池是否启用
	requestParams.SiteId = tea.Int64(c.config.SiteID)                              // 站点ID
	requestParams.Origins = make([]*esa20240910.CreateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
	for _, item := range originGroupOrigins {
		requestParams.Origins = append(requestParams.Origins, &esa20240910.CreateOriginPoolRequestOrigins{
			Address: tea.String(item),
			Type:    tea.String("ip_domain"),
			Name:    tea.String(gomd5.Get16MD5Encode(item)),
			Enabled: tea.Bool(true),
		})
	}
	response, err := c.client.CreateOriginPool(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建源站组", requestParams, response)
	}

	return nil
}

// 更新源站组
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateList
// originGroupID: 源站组ID
// originGroupOrigins: 源站组值列表
func (c *AliyunEsaClient) UpdateOriginGroup(ctx context.Context, originGroupID string, originGroupOrigins []string) error {

	// 请求
	requestParams := &esa20240910.UpdateOriginPoolRequest{}
	requestParams.Id = tea.Int64(cast.ToInt64(originGroupID))                      // 源站组ID
	requestParams.Enabled = tea.Bool(true)                                         // 源地址池是否启用
	requestParams.Origins = make([]*esa20240910.UpdateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
	for _, item := range originGroupOrigins {
		requestParams.Origins = append(requestParams.Origins, &esa20240910.UpdateOriginPoolRequestOrigins{
			Name:    tea.String(gomd5.Get16MD5Encode(item)),
			Address: tea.String(item),
			Type:    tea.String("ip_domain"),
			Enabled: tea.Bool(true),
		})
	}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	response, err := c.client.UpdateOriginPool(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新源站组", requestParams, response)
	}

	return nil
}
