package dns_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type AlidnsClientSdk struct {
	config struct {
		Lang            string // 语言类型
		AccessKeyID     string // 访问密钥ID
		AccessKeySecret string // 访问密钥密钥
	}

	credential credential.Credential
	client     *alidns20150109.Client
}

func NewAlidnsClientSdk(accessKeyID string, accessKeySecret string) (*AlidnsClientSdk, error) {

	var err error

	c := &AlidnsClientSdk{}
	c.config.Lang = "zh"
	c.config.AccessKeyID = accessKeyID
	c.config.AccessKeySecret = accessKeySecret

	// 密钥信息
	c.credential, err = credential.NewCredential(&credential.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
	})
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	c.client, err = alidns20150109.NewClient(&openapi.Config{
		Credential: c.credential,
		Endpoint:   tea.String("alidns.aliyuncs.com"),
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *AlidnsClientSdk) SetDebug(debug bool) *AlidnsClientSdk {
	// c.client.WithDebug(debug)
	return c
}

type AlidnsQueryResponse struct {
	DomainID     string // 域名ID
	DomainName   string // 域名名称
	DomainType   string // 记录类型
	DomainOrigin string // 记录类型值
}

// 获取解析记录列表
// https://next.api.aliyun.com/api/Alidns/2015-01-09/DescribeDomainRecords
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *AlidnsClientSdk) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (AlidnsQueryResponse, error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return AlidnsQueryResponse{}, err
	}

	// 解析IP
	var domainType = ""
	if domainOrigin != "" {
		domainType, err = utils.GetRecordType(domainOrigin)
		if err != nil {
			return AlidnsQueryResponse{}, err
		}
	}

	// 请求
	requestParams := &alidns20150109.DescribeDomainRecordsRequest{
		Lang:       tea.String(c.config.Lang), // 设置请求和接收消息的语言类型
		SearchMode: tea.String("COMBINATION"), // 搜索模式
	}
	requestParams.DomainName = tea.String(domainMain) // 域名名称
	requestParams.RRKeyWord = tea.String(domainSub)   // 主机记录的关键字
	if domainType != "" {
		requestParams.TypeKeyWord = tea.String(domainType) // 解析类型的关键字
	}
	response, err := c.client.DescribeDomainRecords(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return AlidnsQueryResponse{}, fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return AlidnsQueryResponse{}, err
	}

	for _, item := range response.Body.DomainRecords.Record {
		if *item.RR == domainSub {
			if domainType != "" && *item.Type != domainType {
				continue
			}
			return AlidnsQueryResponse{
				DomainID:     *item.RecordId,
				DomainName:   *item.RR,
				DomainType:   *item.Type,
				DomainOrigin: *item.Value,
			}, nil
		}
	}

	return AlidnsQueryResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://next.api.aliyun.com/api/Alidns/2015-01-09/AddDomainRecord
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *AlidnsClientSdk) CreateDomain(ctx context.Context, domainName string, domainOrigin string) error {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}
	if domainSub == "" {
		domainSub = "@"
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &alidns20150109.AddDomainRecordRequest{
		Lang: tea.String(c.config.Lang), // 设置请求和接收消息的语言类型
	}
	requestParams.DomainName = tea.String(domainMain) // 域名名称
	requestParams.RR = tea.String(domainSub)          // 主机记录
	requestParams.Type = tea.String(domainType)       // 解析记录类型
	requestParams.Value = tea.String(domainOrigin)    // 记录值
	_, err = c.client.AddDomainRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://next.api.aliyun.com/api/Alidns/2015-01-09/UpdateDomainRecord
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *AlidnsClientSdk) UpdateDomain(ctx context.Context, domainID string, domainName string, domainOrigin string) error {

	// 解析域名
	_, domainSub, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}
	if domainSub == "" {
		domainSub = "@"
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &alidns20150109.UpdateDomainRecordRequest{
		Lang: tea.String(c.config.Lang), // 设置请求和接收消息的语言类型
	}
	requestParams.RecordId = tea.String(domainID)  // 解析记录的ID
	requestParams.RR = tea.String(domainSub)       // 主机记录
	requestParams.Type = tea.String(domainType)    // 解析记录类型
	requestParams.Value = tea.String(domainOrigin) // 记录值
	_, err = c.client.UpdateDomainRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}
