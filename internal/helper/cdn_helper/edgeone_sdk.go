package cdn_helper

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type EdgeOneClientSdk struct {
	config struct {
		SecretID  string // 密钥ID
		SecretKey string // 密钥密钥
		ZoneID    string // 站点 ID
	}

	credential *common.Credential
	client     *teo.Client
}

// 边缘安全加速平台 EO
// https://cloud.tencent.com/document/api/1552/80730
// https://cloud.tencent.com/document/api/1552/80732
func NewEdgeOneClientSdk(ctx context.Context, secretID string, secretKey string, zoneID string) (*EdgeOneClientSdk, error) {

	var err error

	c := &EdgeOneClientSdk{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey
	c.config.ZoneID = zoneID

	// 密钥信息
	c.credential = common.NewCredential(
		secretID,
		secretKey,
	)

	// 实例化一个client选项
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	c.client, err = teo.NewClient(c.credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *EdgeOneClientSdk) SetDebug(debug bool) *EdgeOneClientSdk {
	c.client.WithDebug(debug)
	return c
}

// 查询加速域名列表
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
// domainName: 域名名称
func (c *EdgeOneClientSdk) QueryDomain(ctx context.Context, domainName string) (ResponseEdgeOneQueryDomain, error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return ResponseEdgeOneQueryDomain{}, err
	}

	// 请求
	requestParams := teo.NewDescribeAccelerationDomainsRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID)
	requestParams.Offset = common.Int64Ptr(0)
	requestParams.Limit = common.Int64Ptr(200)
	requestParams.Filters = []*teo.AdvancedFilter{
		&teo.AdvancedFilter{
			Name:   common.StringPtr("domain-name"),
			Values: common.StringPtrs([]string{domainFull}),
		},
	}
	response, err := c.client.DescribeAccelerationDomains(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return ResponseEdgeOneQueryDomain{}, fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return ResponseEdgeOneQueryDomain{}, err
	}

	for _, item := range response.Response.AccelerationDomains {
		if *item.DomainName == domainFull {
			return ResponseEdgeOneQueryDomain{
				// DomainID:        *item.DomainId,
				DomainName:      *item.DomainName,
				DomainCname:     *item.Cname,
				DomainType:      *item.OriginDetail.OriginType,
				DomainOrigin:    *item.OriginDetail.Origin,
				OriginProtocol:  *item.OriginProtocol,
				HttpOriginPort:  int(*item.HttpOriginPort),
				HttpsOriginPort: int(*item.HttpsOriginPort),
			}, nil
		}
	}

	return ResponseEdgeOneQueryDomain{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateAccelerationDomain
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeOneClientSdk) CreateDomain(ctx context.Context, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	// 转大写
	originProtocol = strings.ToUpper(originProtocol)

	// 默认值
	if httpPort == 0 {
		httpPort = 80
	}
	if httpsPort == 0 {
		httpsPort = 443
	}

	// 请求
	requestParams := teo.NewCreateAccelerationDomainRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 加速域名所属站点 ID
	requestParams.DomainName = common.StringPtr(domainFull)  // 加速域名
	requestParams.OriginInfo = &teo.OriginInfo{
		OriginType: common.StringPtr(originType),  // 源站类型
		Origin:     common.StringPtr(originValue), // 源站地址
	}
	requestParams.OriginProtocol = common.StringPtr(originProtocol)     // 回源协议
	requestParams.HttpOriginPort = common.Uint64Ptr(uint64(httpPort))   // HTTP回源端口
	requestParams.HttpsOriginPort = common.Uint64Ptr(uint64(httpsPort)) // HTTPS回源端口
	requestParams.IPv6Status = common.StringPtr("on")                   // IPv6状态
	_, err = c.client.CreateAccelerationDomain(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 修改加速域名信息
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyAccelerationDomain
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeOneClientSdk) UpdateDomain(ctx context.Context, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	// 转大写
	originProtocol = strings.ToUpper(originProtocol)

	// 默认值
	if httpPort == 0 {
		httpPort = 80
	}
	if httpsPort == 0 {
		httpsPort = 443
	}

	// 请求
	requestParams := teo.NewModifyAccelerationDomainRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 加速域名所属站点 ID
	requestParams.DomainName = common.StringPtr(domainFull)  // 加速域名
	requestParams.OriginInfo = &teo.OriginInfo{
		OriginType: common.StringPtr(originType),  // 源站类型
		Origin:     common.StringPtr(originValue), // 源站地址
	}
	requestParams.OriginProtocol = common.StringPtr(originProtocol)     // 回源协议
	requestParams.HttpOriginPort = common.Uint64Ptr(uint64(httpPort))   // HTTP回源端口
	requestParams.HttpsOriginPort = common.Uint64Ptr(uint64(httpsPort)) // HTTPS回源端口
	requestParams.IPv6Status = common.StringPtr("on")                   // IPv6状态
	_, err = c.client.ModifyAccelerationDomain(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 获取源站组列表
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeOriginGroup
// domain: 源站组名称
func (c *EdgeOneClientSdk) QueryOriginGroup(ctx context.Context, originGroupName string) (ResponseEdgeOneQueryOriginGroup, error) {

	// 请求
	requestParams := teo.NewDescribeOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.Offset = common.Uint64Ptr(0)               // 分页查询偏移量
	requestParams.Limit = common.Uint64Ptr(1000)             // 分页查询限制数目
	requestParams.Filters = []*teo.AdvancedFilter{
		&teo.AdvancedFilter{
			Name:   common.StringPtr("origin-group-name"), // 按照源站组名称进行过滤
			Values: common.StringPtrs([]string{originGroupName}),
		},
	}
	response, err := c.client.DescribeOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return ResponseEdgeOneQueryOriginGroup{}, fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return ResponseEdgeOneQueryOriginGroup{}, err
	}

	for _, item := range response.Response.OriginGroups {
		if *item.Name == originGroupName {
			origins := []string{}
			for _, v := range item.Records {
				origins = append(origins, *v.Record)
			}
			return ResponseEdgeOneQueryOriginGroup{
				GroupID:      *item.GroupId,
				GroupName:    *item.Name,
				GroupType:    *item.Type,
				GroupOrigins: origins,
			}, nil
		}
	}

	return ResponseEdgeOneQueryOriginGroup{}, fmt.Errorf("未找到记录")
}

// 创建源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateOriginGroup
// domain: 源站组名称
// origins: 源站组IP列表
func (c *EdgeOneClientSdk) CreateOriginGroup(ctx context.Context, name string, origins []string) error {

	// 请求
	requestParams := teo.NewCreateOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.Records = make([]*teo.OriginRecord, 0)     // 源站记录信息
	for _, v := range origins {
		requestParams.Records = append(requestParams.Records, &teo.OriginRecord{
			Record: common.StringPtr(v),
			Type:   common.StringPtr("IP_DOMAIN"),
		})
	}
	requestParams.Name = common.StringPtr(name)      // 源站组名称
	requestParams.Type = common.StringPtr("GENERAL") // 源站组类型
	_, err := c.client.CreateOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 修改源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyOriginGroup
// domain: 源站组ID
// origins: 源站组IP列表
func (c *EdgeOneClientSdk) UpdateOriginGroup(ctx context.Context, groupId string, origins []string) error {

	records := []map[string]any{}
	for _, v := range origins {
		records = append(records, map[string]any{
			"Record": v,
			"Type":   "IP_DOMAIN",
		})
	}

	// 请求
	requestParams := teo.NewModifyOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.GroupId = common.StringPtr(groupId)        // 源站组ID
	requestParams.Records = make([]*teo.OriginRecord, 0)     // 源站记录信息
	for _, v := range origins {
		requestParams.Records = append(requestParams.Records, &teo.OriginRecord{
			Record: common.StringPtr(v),
			Type:   common.StringPtr("IP_DOMAIN"),
		})
	}
	_, err := c.client.ModifyOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}
