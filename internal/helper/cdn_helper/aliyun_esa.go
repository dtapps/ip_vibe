package cdn_helper

import (
	"context"
	"fmt"
	"net/http"
	"vibe_ip/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gomd5"
)

// AliHttpHandler 适配 github.com/alibabacloud-go/tea/tea 里的 HttpClient 接口
type AliyunHttpHandler struct {
	httpClient *http.Client
}

func NewAliyunHttpHandler(log *http_log.LoggingTransport) *AliyunHttpHandler {
	return &AliyunHttpHandler{
		httpClient: &http.Client{
			Transport: log,
		},
	}
}

// Call 实现了 dara.HttpClient 接口要求的方法
func (a *AliyunHttpHandler) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	return a.httpClient.Do(request)
}

type AliyunEsaClient struct {
	config struct {
		AccessKeyID     string // 密钥ID
		AccessKeySecret string // 密钥密钥
		SiteID          int64  // 站点 ID
	}

	credential   credential.Credential
	clientConfig *openapi.Config
	client       *esa20240910.Client
}

func NewAliyunEsaClient(ctx context.Context, accessKeyID string, accessKeySecret string, siteID int64) (*AliyunEsaClient, error) {

	var err error

	c := &AliyunEsaClient{}
	c.config.AccessKeyID = accessKeyID
	c.config.AccessKeySecret = accessKeySecret
	c.config.SiteID = siteID

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
// func (c *AliyunEsaClient) SetDebug(debug bool) *AliyunEsaClient {
// 	c.client.WithDebug(debug)
// 	return c
// }

// 设置HTTP日志记录器
func (c *AliyunEsaClient) SetLog(log *http_log.LoggingTransport) *AliyunEsaClient {
	if log == nil {
		return c
	}
	// 实例化client选项
	c.clientConfig.HttpClient = NewAliyunHttpHandler(log)
	// 实例化client对象
	c.client, _ = esa20240910.NewClient(c.clientConfig)
	return c
}

// 查询加速域名列表
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListRecords
// domainName: 主域名
// subDomain: 子域名
func (c *AliyunEsaClient) QueryDomain(ctx context.Context, domainName string, subDomain string) (EdgeSecurityAccelerationResponseQueryDomain, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName := utils.GetCdnDomainParts(subDomain, domainName)

	// 请求
	requestParams := &esa20240910.ListRecordsRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	requestParams.RecordName = tea.String(fullDomain)
	requestParams.PageNumber = tea.Int32(1)
	requestParams.PageSize = tea.Int32(500)
	response, err := c.client.ListRecords(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return EdgeSecurityAccelerationResponseQueryDomain{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return EdgeSecurityAccelerationResponseQueryDomain{}, err
	}

	for _, item := range response.Body.Records {
		if tea.StringValue(item.RecordName) == fullDomain {
			return EdgeSecurityAccelerationResponseQueryDomain{
				DomainID:    cast.ToString(tea.Int64Value(item.RecordId)),
				DomainName:  tea.StringValue(item.RecordName),
				DomainCname: tea.StringValue(item.RecordCname),
				OriginType:  tea.StringValue(item.RecordType),
				OriginValue: tea.StringValue(item.Data.Value),
			}, nil
		}
	}

	return EdgeSecurityAccelerationResponseQueryDomain{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateRecord
// domainName: 主域名
// subDomain: 子域名
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *AliyunEsaClient) CreateDomain(ctx context.Context, domainName string, subDomain string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName := utils.GetCdnDomainParts(subDomain, domainName)

	// 请求
	requestParams := &esa20240910.CreateRecordRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID)  // 站点ID
	requestParams.RecordName = tea.String(fullDomain)  // 记录名称
	requestParams.Proxied = tea.Bool(true)             // 记录是否开启代理加速
	requestParams.Type = tea.String("A/AAAA")          // 记录的DNS类型
	requestParams.Ttl = tea.Int32(1)                   // 记录的过期时间
	requestParams.Data.Value = tea.String(originValue) // 记录值或部分内容
	_, err := c.client.CreateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 修改加速域名信息
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateRecord
// domainName: 主域名
// subDomain: 子域名
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
func (c *AliyunEsaClient) UpdateDomain(ctx context.Context, recordId string, originType string, originValue string) error {

	// 请求
	requestParams := &esa20240910.UpdateRecordRequest{}
	requestParams.RecordId = tea.Int64(cast.ToInt64(recordId)) // 站点ID
	requestParams.Ttl = tea.Int32(1)                           // 记录的过期时间，单位秒
	requestParams.Proxied = tea.Bool(true)                     // 记录是否开启代理加速
	requestParams.Data.Value = tea.String(originValue)         // 记录值或部分内容
	_, err := c.client.UpdateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 获取源站组详情
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateList
// originGroupID: 源站组ID
func (c *AliyunEsaClient) QueryOriginGroup(ctx context.Context, originGroupID string) ([]string, error) {

	// 请求
	requestParams := &esa20240910.GetOriginPoolRequest{}
	requestParams.SiteId = tea.Int64(c.config.SiteID)         // 站点ID
	requestParams.Id = tea.Int64(cast.ToInt64(originGroupID)) // 源站组ID
	response, err := c.client.GetOriginPool(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return []string{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return []string{}, err
	}

	origins := []string{}
	for _, item := range response.Body.Origins {
		if tea.BoolValue(item.Enabled) {
			origins = append(origins, tea.StringValue(item.Address))
		}
	}
	if len(origins) == 0 {
		return []string{}, fmt.Errorf("未找到记录")
	}

	return origins, nil
}

// 创建源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateOriginGroup
// originGroupName: 源站组名称
// originValues: 源站组值列表
func (c *AliyunEsaClient) CreateOriginGroup(ctx context.Context, originGroupName string, originValues []string) (originGroupID string, err error) {

	// 请求
	requestParams := &esa20240910.CreateOriginPoolRequest{}
	requestParams.Name = tea.String(originGroupName)                               // 列表名称
	requestParams.Enabled = tea.Bool(true)                                         // 源地址池是否启用
	requestParams.SiteId = tea.Int64(c.config.SiteID)                              // 站点ID
	requestParams.Origins = make([]*esa20240910.CreateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
	for _, v := range originValues {
		requestParams.Origins = append(requestParams.Origins, &esa20240910.CreateOriginPoolRequestOrigins{
			Address: tea.String(v),
			Type:    tea.String("ip_domain"),
			Name:    tea.String(gomd5.Get16MD5Encode(v)),
			Enabled: tea.Bool(true),
		})
	}
	response, err := c.client.CreateOriginPool(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return "", fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return "", err
	}

	return cast.ToString(tea.Int64Value(response.Body.Id)), nil
}

// 修改源站组
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateList
// originGroupID: 源站组ID
// originGroupName: 源站组名称
// originValues: 源站组值列表
func (c *AliyunEsaClient) UpdateOriginGroup(ctx context.Context, originGroupID string, originValues []string) error {

	// 请求
	requestParams := &esa20240910.UpdateOriginPoolRequest{}
	requestParams.Id = tea.Int64(cast.ToInt64(originGroupID))                      // 源站组ID
	requestParams.Enabled = tea.Bool(true)                                         // 源地址池是否启用
	requestParams.Origins = make([]*esa20240910.UpdateOriginPoolRequestOrigins, 0) // 源地址池里添加的源站信息
	for _, v := range originValues {
		requestParams.Origins = append(requestParams.Origins, &esa20240910.UpdateOriginPoolRequestOrigins{
			Name:    tea.String(gomd5.Get16MD5Encode(v)),
			Address: tea.String(v),
			Type:    tea.String("ip_domain"),
			Enabled: tea.Bool(true),
		})
	}
	requestParams.SiteId = tea.Int64(c.config.SiteID) // 站点ID
	_, err := c.client.UpdateOriginPool(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}
