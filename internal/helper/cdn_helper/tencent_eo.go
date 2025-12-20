package cdn_helper

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
	"go.dtapp.net/library/contrib/http_log"
)

type TencentEOConfig struct {
	SecretID   string `json:"secret_id"`   // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥Key
	ZoneID     string `json:"zone_id"`     // 区域ID
	DomainName string `json:"domain_name"` // 主域名
}

type TencentEOClient struct {
	config       *TencentEOConfig // 配置
	debug        bool             // 是否开启调试模式
	credential   *common.Credential
	clientConfig *profile.ClientProfile
	client       *teo.Client
}

// 边缘安全加速平台 EO
// https://cloud.tencent.com/document/product/1552
func NewTencentEOClient(ctx context.Context, config *TencentEOConfig) (*TencentEOClient, error) {

	var err error

	c := &TencentEOClient{}
	c.config = config

	// 密钥信息
	c.credential = common.NewCredential(
		c.config.SecretID,
		c.config.SecretKey,
	)

	// 实例化一个client选项
	c.clientConfig = profile.NewClientProfile()
	c.clientConfig.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	c.client, err = teo.NewClient(c.credential, "", c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *TencentEOClient) SetDebug(debug bool) *TencentEOClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *TencentEOClient) SetLog(loggingTransport *http_log.LoggingTransport) *TencentEOClient {
	if loggingTransport == nil {
		return c
	}
	c.client.WithHttpTransport(loggingTransport)
	return c
}

// 查询加速域名
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
// fullDomain: 完整域名
func (c *TencentEOClient) QueryDomain(ctx context.Context, fullDomain string) (TencentEOQueryDomainResponse, error) {

	// 请求
	requestParams := teo.NewDescribeAccelerationDomainsRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID)
	requestParams.Offset = common.Int64Ptr(0)
	requestParams.Limit = common.Int64Ptr(200)
	requestParams.Filters = make([]*teo.AdvancedFilter, 0)
	requestParams.Filters = append(requestParams.Filters, &teo.AdvancedFilter{
		Name:   common.StringPtr("domain-name"),
		Values: common.StringPtrs([]string{fullDomain}),
	})
	response, err := c.client.DescribeAccelerationDomains(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return TencentEOQueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return TencentEOQueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询加速域名", requestParams, response)
	}

	for _, item := range response.Response.AccelerationDomains {
		if strings.EqualFold(utils.PtrToString(item.DomainName), fullDomain) {
			return TencentEOQueryDomainResponse{
				// RecordID:         utils.PtrToString(item.DomainId),
				FullDomain:      utils.PtrToString(item.DomainName),
				DomainCname:     utils.PtrToString(item.Cname),
				RecordType:      utils.PtrToString(item.OriginDetail.OriginType),
				OriginValue:     utils.PtrToString(item.OriginDetail.Origin),
				OriginProtocol:  utils.PtrToString(item.OriginProtocol),
				HttpOriginPort:  cast.ToInt(utils.PtrToUint64(item.HttpOriginPort)),
				HttpsOriginPort: cast.ToInt(utils.PtrToUint64(item.HttpsOriginPort)),
			}, nil
		}
	}

	return TencentEOQueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateAccelerationDomain
// fullDomain: 完整域名
// recordType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// recordType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *TencentEOClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 转大写
	recordType = strings.ToUpper(recordType)
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
	requestParams.DomainName = common.StringPtr(fullDomain)  // 加速域名
	requestParams.OriginInfo = &teo.OriginInfo{
		OriginType: common.StringPtr(recordType),  // 源站类型
		Origin:     common.StringPtr(originValue), // 源站地址
	}
	requestParams.OriginProtocol = common.StringPtr(originProtocol)     // 回源协议
	requestParams.HttpOriginPort = common.Uint64Ptr(uint64(httpPort))   // HTTP回源端口
	requestParams.HttpsOriginPort = common.Uint64Ptr(uint64(httpsPort)) // HTTPS回源端口
	requestParams.IPv6Status = common.StringPtr("on")                   // IPv6状态
	response, err := c.client.CreateAccelerationDomain(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
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
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyAccelerationDomain
// fullDomain: 完整域名
// recordType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// recordType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *TencentEOClient) UpdateDomain(ctx context.Context, fullDomain string, recordType string, originValue string, originProtocol string, httpPort int, httpsPort int) error {

	// 转大写
	recordType = strings.ToUpper(recordType)
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
	requestParams.DomainName = common.StringPtr(fullDomain)  // 加速域名
	requestParams.OriginInfo = &teo.OriginInfo{
		OriginType: common.StringPtr(recordType),  // 源站类型
		Origin:     common.StringPtr(originValue), // 源站地址
	}
	requestParams.OriginProtocol = common.StringPtr(originProtocol)     // 回源协议
	requestParams.HttpOriginPort = common.Uint64Ptr(uint64(httpPort))   // HTTP回源端口
	requestParams.HttpsOriginPort = common.Uint64Ptr(uint64(httpsPort)) // HTTPS回源端口
	requestParams.IPv6Status = common.StringPtr("on")                   // IPv6状态
	response, err := c.client.ModifyAccelerationDomain(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
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
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeOriginGroup
// originGroupName: 源站组名称
func (c *TencentEOClient) QueryOriginGroup(ctx context.Context, originGroupName string) (TencentEOQueryOriginGroupResponse, error) {

	// 请求
	requestParams := teo.NewDescribeOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.Offset = common.Uint64Ptr(0)               // 分页查询偏移量
	requestParams.Limit = common.Uint64Ptr(1000)             // 分页查询限制数目
	requestParams.Filters = make([]*teo.AdvancedFilter, 0)
	requestParams.Filters = append(requestParams.Filters, &teo.AdvancedFilter{
		Name:   common.StringPtr("origin-group-name"), // 按照源站组名称进行过滤
		Values: common.StringPtrs([]string{originGroupName}),
	})
	response, err := c.client.DescribeOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return TencentEOQueryOriginGroupResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return TencentEOQueryOriginGroupResponse{}, err
	}
	if c.debug {
		printJSON("查询源站组", requestParams, response)
	}

	for _, item := range response.Response.OriginGroups {
		if strings.EqualFold(utils.PtrToString(item.Name), originGroupName) {
			origins := []string{}
			for _, v := range item.Records {
				origins = append(origins, utils.PtrToString(v.Record))
			}
			return TencentEOQueryOriginGroupResponse{
				GroupID:      utils.PtrToString(item.GroupId),
				GroupName:    utils.PtrToString(item.Name),
				GroupType:    utils.PtrToString(item.Type),
				GroupOrigins: origins,
			}, nil
		}
	}

	return TencentEOQueryOriginGroupResponse{}, fmt.Errorf("未找到记录")
}

// 创建源站组
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateOriginGroup
// originGroupName: 源站组名称
// originGroupOrigins: 源站组值列表
func (c *TencentEOClient) CreateOriginGroup(ctx context.Context, originGroupName string, originGroupOrigins []string) error {

	// 请求
	requestParams := teo.NewCreateOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.Records = make([]*teo.OriginRecord, 0)     // 源站记录信息
	for _, item := range originGroupOrigins {
		requestParams.Records = append(requestParams.Records, &teo.OriginRecord{
			Record: common.StringPtr(item),
			Type:   common.StringPtr("IP_DOMAIN"),
		})
	}
	requestParams.Name = common.StringPtr(originGroupName) // 源站组名称
	requestParams.Type = common.StringPtr("GENERAL")       // 源站组类型
	response, err := c.client.CreateOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
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
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyOriginGroup
// originGroupID: 源站组ID
// originGroupOrigins: 源站组值列表
func (c *TencentEOClient) UpdateOriginGroup(ctx context.Context, originGroupID string, originGroupOrigins []string) error {

	// 请求
	requestParams := teo.NewModifyOriginGroupRequest()
	requestParams.ZoneId = common.StringPtr(c.config.ZoneID) // 站点ID
	requestParams.GroupId = common.StringPtr(originGroupID)  // 源站组ID
	requestParams.Records = make([]*teo.OriginRecord, 0)     // 源站记录信息
	for _, item := range originGroupOrigins {
		requestParams.Records = append(requestParams.Records, &teo.OriginRecord{
			Record: common.StringPtr(item),
			Type:   common.StringPtr("IP_DOMAIN"),
		})
	}
	response, err := c.client.ModifyOriginGroup(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
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
