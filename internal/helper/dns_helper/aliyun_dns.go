package dns_helper

import (
	"context"
	"fmt"
	"net/http"
	"vibe_ip/internal/utils"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"go.dtapp.net/library/contrib/http_log"
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

type AliyunDnsClient struct {
	config struct {
		Lang            string // 语言类型
		AccessKeyID     string // 密钥ID
		AccessKeySecret string // 密钥密钥
	}

	credential   credential.Credential
	clientConfig *openapi.Config
	client       *alidns20150109.Client
}

func NewAliyunDnsClient(ctx context.Context, accessKeyID string, accessKeySecret string) (*AliyunDnsClient, error) {

	var err error

	c := &AliyunDnsClient{}
	c.config.Lang = "zh"
	c.config.AccessKeyID = accessKeyID
	c.config.AccessKeySecret = accessKeySecret

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
		Endpoint:   tea.String("alidns.aliyuncs.com"),
	}

	// 实例化client对象
	c.client, err = alidns20150109.NewClient(c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
// func (c *AliyunDnsClient) SetDebug(debug bool) *AliyunDnsClient {
// 	c.client.WithDebug(debug)
// 	return c
// }

// 设置HTTP日志记录器
func (c *AliyunDnsClient) SetLog(log *http_log.LoggingTransport) *AliyunDnsClient {
	if log == nil {
		return c
	}

	// 实例化client选项
	c.clientConfig.HttpClient = NewAliyunHttpHandler(log)

	// 实例化client对象
	c.client, _ = alidns20150109.NewClient(c.clientConfig)

	return c
}

// 获取解析记录列表
// https://next.api.aliyun.com/api/Alidns/2015-01-09/DescribeDomainRecords
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *AliyunDnsClient) QueryDomain(ctx context.Context, domainName string, subDomain string, originValue string) (QueryDomainResponse, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return QueryDomainResponse{}, err
	}

	// 请求
	requestParams := &alidns20150109.DescribeDomainRecordsRequest{}
	requestParams.Lang = tea.String(c.config.Lang)       // 设置请求和接收消息的语言类型
	requestParams.SearchMode = tea.String("COMBINATION") // 搜索模式
	requestParams.DomainName = tea.String(domainName)    // 域名名称
	requestParams.RRKeyWord = tea.String(subDomain)      // 主机记录的关键字
	if recordType != "" {
		requestParams.TypeKeyWord = tea.String(recordType) // 解析类型的关键字
	}
	response, err := c.client.DescribeDomainRecords(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return QueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return QueryDomainResponse{}, err
	}

	for _, item := range response.Body.DomainRecords.Record {
		if tea.StringValue(item.RR) == subDomain {
			if recordType != "" && tea.StringValue(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				DomainID:    tea.StringValue(item.RecordId),
				DomainName:  tea.StringValue(item.RR),
				OriginType:  tea.StringValue(item.Type),
				OriginValue: tea.StringValue(item.Value),
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://next.api.aliyun.com/api/Alidns/2015-01-09/AddDomainRecord
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *AliyunDnsClient) CreateDomain(ctx context.Context, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &alidns20150109.AddDomainRecordRequest{}
	requestParams.Lang = tea.String(c.config.Lang)    // 设置请求和接收消息的语言类型
	requestParams.DomainName = tea.String(domainName) // 域名名称
	requestParams.RR = tea.String(subDomain)          // 主机记录
	requestParams.Type = tea.String(recordType)       // 解析记录类型
	requestParams.Value = tea.String(originValue)     // 记录值
	_, err = c.client.AddDomainRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://next.api.aliyun.com/api/Alidns/2015-01-09/UpdateDomainRecord
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *AliyunDnsClient) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &alidns20150109.UpdateDomainRecordRequest{}
	requestParams.Lang = tea.String(c.config.Lang) // 设置请求和接收消息的语言类型
	requestParams.RecordId = tea.String(domainID)  // 解析记录的ID
	requestParams.RR = tea.String(subDomain)       // 主机记录
	requestParams.Type = tea.String(recordType)    // 解析记录类型
	requestParams.Value = tea.String(originValue)  // 记录值
	_, err = c.client.UpdateDomainRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}
