package cdn_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type EdgeSecurityAccelerationClientSdk struct {
	config struct {
		AccessKeyID     string // 密钥ID
		AccessKeySecret string // 密钥密钥
		SiteID          int64  // 站点 ID
	}

	credential credential.Credential
	client     *esa20240910.Client
}

func NewEdgeSecurityAccelerationClientSdk(accessKeyID string, accessKeySecret string, siteID int64) (*EdgeSecurityAccelerationClientSdk, error) {

	var err error

	c := &EdgeSecurityAccelerationClientSdk{}
	c.config.AccessKeyID = accessKeyID
	c.config.AccessKeySecret = accessKeySecret
	c.config.SiteID = siteID

	// 密钥信息
	c.credential, err = credential.NewCredential(&credential.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
	})
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	c.client, err = esa20240910.NewClient(&openapi.Config{
		Credential: c.credential,
		Endpoint:   tea.String("esa.cn-hangzhou.aliyuncs.com"),
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *EdgeSecurityAccelerationClientSdk) SetDebug(debug bool) *EdgeSecurityAccelerationClientSdk {
	// c.client.WithDebug(debug)
	return c
}

// 查询加速域名列表
type EdgeSecurityAccelerationResponseQueryDomain struct {
	DomainID     int64  // 域名ID
	DomainName   string // 域名名称
	DomainCname  string // 域名别名
	DomainType   string // 源站类型
	DomainOrigin string // 源站类型值
}

// 查询加速域名列表
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListRecords
// domainName: 域名名称
func (c *EdgeSecurityAccelerationClientSdk) QueryDomain(ctx context.Context, domainName string) (EdgeSecurityAccelerationResponseQueryDomain, error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return EdgeSecurityAccelerationResponseQueryDomain{}, err
	}

	// 请求
	requestParams := &esa20240910.ListRecordsRequest{
		SiteId: tea.Int64(c.config.SiteID), // 站点ID
	}
	requestParams.RecordName = tea.String(domainFull)
	requestParams.PageNumber = tea.Int32(1)
	requestParams.PageSize = tea.Int32(500)
	response, err := c.client.ListRecords(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return EdgeSecurityAccelerationResponseQueryDomain{}, fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return EdgeSecurityAccelerationResponseQueryDomain{}, err
	}

	for _, item := range response.Body.Records {
		if *item.RecordName == domainFull {
			return EdgeSecurityAccelerationResponseQueryDomain{
				DomainID:     *item.RecordId,
				DomainName:   *item.RecordName,
				DomainCname:  *item.RecordCname,
				DomainType:   *item.RecordType,
				DomainOrigin: *item.Data.Value,
			}, nil
		}
	}

	return EdgeSecurityAccelerationResponseQueryDomain{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://next.api.aliyun.com/api/ESA/2024-09-10/CreateRecord
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeSecurityAccelerationClientSdk) CreateDomain(ctx context.Context, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &esa20240910.CreateRecordRequest{
		SiteId: tea.Int64(c.config.SiteID), // 站点ID
	}
	requestParams.RecordName = tea.String(domainFull)  // 记录名称
	requestParams.Proxied = tea.Bool(true)             // 记录是否开启代理加速
	requestParams.Type = tea.String("A/AAAA")          // 记录的DNS类型
	requestParams.Ttl = tea.Int32(1)                   // 记录的过期时间
	requestParams.Data.Value = tea.String(originValue) // 记录值或部分内容
	_, err = c.client.CreateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 修改加速域名信息
// https://next.api.aliyun.com/api/ESA/2024-09-10/UpdateRecord
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeSecurityAccelerationClientSdk) UpdateDomain(ctx context.Context, recordId int64, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 解析域名
	_, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &esa20240910.UpdateRecordRequest{
		RecordId: tea.Int64(recordId),
	}
	requestParams.Ttl = tea.Int32(1)                   // 记录的过期时间，单位秒
	requestParams.Proxied = tea.Bool(true)             // 记录是否开启代理加速
	requestParams.Data.Value = tea.String(originValue) // 记录值或部分内容
	_, err = c.client.UpdateRecord(requestParams)
	if _, ok := err.(*tea.SDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}
