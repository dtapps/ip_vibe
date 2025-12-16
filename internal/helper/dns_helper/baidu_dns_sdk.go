package dns_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/baidubce/bce-sdk-go/util"
)

type BaiduDnsClientSdk struct {
	config struct {
		AccessKeyID     string // 密钥ID
		SecretAccessKey string // 密钥密钥
	}

	client *dns.Client
}

func NewBaiduDnsClientSdk(accessKeyID string, secretAccessKey string) (*BaiduDnsClientSdk, error) {

	var err error

	c := &BaiduDnsClientSdk{}
	c.config.AccessKeyID = accessKeyID
	c.config.SecretAccessKey = secretAccessKey

	c.client, err = dns.NewClient(c.config.AccessKeyID, c.config.SecretAccessKey, "dns.baidubce.com")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *BaiduDnsClientSdk) SetDebug(debug bool) *BaiduDnsClientSdk {
	// c.client.WithDebug(debug)
	return c
}

type BaiduDnsQueryResponse struct {
	DomainID     string // 域名ID
	DomainName   string // 域名名称
	DomainType   string // 记录类型
	DomainOrigin string // 记录类型值
}

// 获取解析记录列表
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#查询解析记录列表
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *BaiduDnsClientSdk) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (BaiduDnsQueryResponse, error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return BaiduDnsQueryResponse{}, err
	}

	// 解析IP
	var domainType = ""
	if domainOrigin != "" {
		domainType, err = utils.GetRecordType(domainOrigin)
		if err != nil {
			return BaiduDnsQueryResponse{}, err
		}
	}

	// 请求
	requestParams := &dns.ListRecordRequest{}
	requestParams.Rr = domainSub // 主机记录
	requestParams.MaxKeys = 1000 // 每页包含的最大数量，最大数量通常不超过1000，缺省值为1000
	response, err := c.client.ListRecord(domainMain, requestParams)
	if _, ok := err.(*bce.BceClientError); ok {
		return BaiduDnsQueryResponse{}, fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return BaiduDnsQueryResponse{}, fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return BaiduDnsQueryResponse{}, err
	}

	for _, item := range response.Records {
		if item.Rr == domainSub {
			if domainType != "" && item.Type != domainType {
				continue
			}
			return BaiduDnsQueryResponse{
				DomainID:     item.Id,
				DomainName:   item.Rr,
				DomainType:   item.Type,
				DomainOrigin: item.Value,
			}, nil
		}
	}

	return BaiduDnsQueryResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#添加解析记录
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *BaiduDnsClientSdk) CreateDomain(ctx context.Context, domainName string, domainOrigin string) error {

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
	requestParams := &dns.CreateRecordRequest{}
	requestParams.Rr = domainSub       // 主机记录
	requestParams.Type = domainType    // 解析记录类型
	requestParams.Value = domainOrigin // 记录值
	err = c.client.CreateRecord(domainMain, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#修改解析记录
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *BaiduDnsClientSdk) UpdateDomain(ctx context.Context, domainID string, domainName string, domainOrigin string) error {

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
	requestParams := &dns.UpdateRecordRequest{}
	requestParams.Rr = domainSub       // 主机记录
	requestParams.Type = domainType    // 解析记录类型
	requestParams.Value = domainOrigin // 记录值
	err = c.client.UpdateRecord(domainMain, domainID, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}
