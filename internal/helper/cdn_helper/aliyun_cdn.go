package cdn_helper

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cdn20180510 "github.com/alibabacloud-go/cdn-20180510/v9/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
)

// AliyunCdnHttpHandler 适配 github.com/alibabacloud-go/tea/tea 里的 HttpClient 接口
type AliyunCdnHttpHandler struct {
	httpClient *http.Client
}

func NewAliyunCdnHttpHandler(log *http_log.LoggingTransport) *AliyunCdnHttpHandler {
	return &AliyunCdnHttpHandler{
		httpClient: &http.Client{
			Transport: log,
		},
	}
}

// Call 实现了 dara.HttpClient 接口要求的方法
func (a *AliyunCdnHttpHandler) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	return a.httpClient.Do(request)
}

type AliyunCdnConfig struct {
	AccessKeyID     string `json:"access_key_id"`     // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret"` // 访问密钥Key
	DomainName      string `json:"domain_name"`       // 主域名
}

type AliyunCdnClient struct {
	config       *AliyunCdnConfig // 配置
	debug        bool             // 是否开启调试模式
	credential   credential.Credential
	clientConfig *openapi.Config
	client       *cdn20180510.Client
}

// 内容分发服务
// https://cloud.tencent.com/document/product/1552
func NewAliyunCdnClient(ctx context.Context, config *AliyunCdnConfig) (*AliyunCdnClient, error) {

	var err error

	c := &AliyunCdnClient{}
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
		Endpoint:   tea.String("cdn.aliyuncs.com"),
	}

	// 实例化client对象
	c.client, err = cdn20180510.NewClient(c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *AliyunCdnClient) SetDebug(debug bool) *AliyunCdnClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *AliyunCdnClient) SetLog(loggingTransport *http_log.LoggingTransport) *AliyunCdnClient {
	if loggingTransport == nil {
		return c
	}
	// 实例化client选项
	clientConfig := c.clientConfig
	clientConfig.HttpClient = NewAliyunCdnHttpHandler(loggingTransport)
	// 实例化client对象
	client, err := cdn20180510.NewClient(clientConfig)
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
func (c *AliyunCdnClient) QueryDomain(ctx context.Context, fullDomain string) (AliyunCdnQueryDomainResponse, error) {

	// 请求
	requestParams := &cdn20180510.DescribeUserDomainsRequest{}
	requestParams.DomainName = tea.String(fullDomain) // 加速域名
	requestParams.PageNumber = tea.Int32(1)           // 取得第几页
	requestParams.PageSize = tea.Int32(500)           // 分页大小
	response, err := c.client.DescribeUserDomains(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return AliyunCdnQueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return AliyunCdnQueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询加速域名", requestParams, response)
	}

	for _, item := range response.Body.Domains.PageData {
		if strings.EqualFold(tea.StringValue(item.DomainName), fullDomain) {
			return AliyunCdnQueryDomainResponse{
				RecordID:    cast.ToString(tea.Int64Value(item.DomainId)),
				FullDomain:  tea.StringValue(item.DomainName),
				DomainCname: tea.StringValue(item.Cname),
				RecordType:  tea.StringValue(item.Sources.Source[0].Type),
				OriginValue: tea.StringValue(item.Sources.Source[0].Content),
			}, nil
		}
	}

	return AliyunCdnQueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://next.api.aliyun.com/api/Cdn/2018-05-10/AddCdnDomain
// fullDomain: 完整域名
// originValue: 源站地址或源站组ID
func (c *AliyunCdnClient) CreateDomain(ctx context.Context, fullDomain string, originValue string) error {

	type Sources struct {
		Type    string `json:"Type"`    // 回源地址类型
		Content string `json:"Content"` // 回源地址
	}
	sources := Sources{
		Type:    "ipaddr",
		Content: originValue,
	}

	// 请求
	requestParams := &cdn20180510.AddCdnDomainRequest{}
	requestParams.CdnType = tea.String("web")                             // 加速域名的业务类型
	requestParams.DomainName = tea.String(fullDomain)                     // 需要接入CDN的加速域名
	requestParams.Sources = tea.String(gojson.JsonEncodeNoError(sources)) // 回源地址列表
	requestParams.Scope = tea.String("domestic")                          // 加速区域
	response, err := c.client.AddCdnDomain(requestParams)
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
// https://next.api.aliyun.com/api/Cdn/2018-05-10/ModifyCdnDomain
// fullDomain: 完整域名
// originValue: 源站地址或源站组ID
func (c *AliyunCdnClient) UpdateDomain(ctx context.Context, fullDomain string, originValue string) error {

	type Sources struct {
		Type    string `json:"Type"`    // 回源地址类型
		Content string `json:"Content"` // 回源地址
	}
	sources := Sources{
		Type:    "ipaddr",
		Content: originValue,
	}

	// 请求
	requestParams := &cdn20180510.ModifyCdnDomainRequest{}
	requestParams.DomainName = tea.String(fullDomain)                     // 需要接入CDN的加速域名
	requestParams.Sources = tea.String(gojson.JsonEncodeNoError(sources)) // 回源地址列表
	response, err := c.client.ModifyCdnDomain(requestParams)
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
