package cdn_helper

import (
	"context"
	"vibe_ip/internal/helper/tencent_helper"
	"vibe_ip/internal/utils"
)

type EdgeOneClient struct {
	config struct {
		SecretID  string // 密钥ID
		SecretKey string // 密钥密钥
		ZoneID    string // 站点 ID
	}
	client *tencent_helper.Client
}

// 边缘安全加速平台 EO
// https://cloud.tencent.com/document/api/1552/80730
// https://cloud.tencent.com/document/api/1552/80732
func NewEdgeOneClient(ctx context.Context, secretID string, secretKey string, zoneID string) (c *EdgeOneClient, err error) {

	c = &EdgeOneClient{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey
	c.config.ZoneID = zoneID

	c.client, err = tencent_helper.NewClient(
		"https://teo.tencentcloudapi.com",
		c.config.SecretID,
		c.config.SecretKey,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 开启调试
func (c *EdgeOneClient) EnableDebug() *EdgeOneClient {
	c.client.WithDebug()
	return c
}

// 查询加速域名列表
type ResponseEdgeOneQueryDomain struct {
	DomainID        string // 域名ID
	DomainName      string // 域名名称
	DomainCname     string // 域名别名
	DomainType      string // 源站类型
	DomainOrigin    string // 源站类型值
	OriginProtocol  string // 回源协议
	HttpOriginPort  int    // HTTP回源端口
	HttpsOriginPort int    // HTTPS回源端口
}

// 查询加速域名列表
// https://cloud.tencent.com/document/api/1552/86336
// domainName: 域名名称
func (c *EdgeOneClient) QueryDomain(ctx context.Context, domainName string) (resp ResponseEdgeOneQueryDomain, err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return resp, err
	}

	// 请求
	apiResp := tencent_helper.ResponseQueryDomain{}
	_, err = c.client.R().
		SetXTCAction("DescribeAccelerationDomains").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId": c.config.ZoneID, // 加速域名所属站点 ID
			"Offset": 0,               // 分页查询偏移量
			"Limit":  200,             // 分页查询限制数目
			"Filters": []map[string]any{
				{
					"Name":   "domain-name", // 按照加速域名进行过滤
					"Values": []string{domainFull},
				},
			},
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return resp, err
	}

	for _, item := range apiResp.Response.AccelerationDomains {
		if item.DomainName == domainFull {
			return ResponseEdgeOneQueryDomain{
				DomainID:        item.DomainId,
				DomainName:      item.DomainName,
				DomainCname:     item.Cname,
				DomainType:      item.OriginDetail.OriginType,
				DomainOrigin:    item.OriginDetail.Origin,
				OriginProtocol:  item.OriginProtocol,
				HttpOriginPort:  item.HttpOriginPort,
				HttpsOriginPort: item.HttpsOriginPort,
			}, nil
		}
	}

	return resp, nil
}

// 创建加速域名
// https://cloud.tencent.com/document/api/1552/86338
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeOneClient) CreateDomain(ctx context.Context, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) (err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	switch originProtocol {
	case "follow":
		originProtocol = "FOLLOW"
	case "http":
		originProtocol = "HTTP"
	case "https":
		originProtocol = "HTTPS"
	}
	if httpPort == 0 {
		httpPort = 80
	}
	if httpsPort == 0 {
		httpsPort = 443
	}

	// 请求
	apiResp := tencent_helper.ResponseCreateDomain{}
	_, err = c.client.R().
		SetXTCAction("CreateAccelerationDomain").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId":     c.config.ZoneID, // 加速域名所属站点 ID
			"DomainName": domainFull,      // 加速域名
			"OriginInfo": map[string]any{
				"OriginType": originType,  // 源站类型
				"Origin":     originValue, // 源站地址
			}, // 源站信息
			"OriginProtocol":  originProtocol, // 回源协议
			"HttpOriginPort":  httpPort,       // HTTP回源端口
			"HttpsOriginPort": httpsPort,      // HTTPS回源端口
			"IPv6Status":      "on",           // IPv6状态
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return err
	}

	return nil
}

// 修改加速域名信息
// https://cloud.tencent.com/document/api/1552/86335
// domainName: 域名名称
// originType: IP_DOMAIN = IPV4、IPV6 或域名类型源站
// originType: ORIGIN_GROUP = 源站组类型源站
// originValue: 源站地址或源站组ID
// originProtocol: follow = 遵循源站协议、http = HTTP 协议、https = HTTPS 协议
// httpPort: HTTP回源端口
// httpsPort: HTTPS回源端口
func (c *EdgeOneClient) UpdateDomain(ctx context.Context, domainName string, originType string, originValue string, originProtocol string, httpPort int, httpsPort int) (err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	switch originProtocol {
	case "follow":
		originProtocol = "FOLLOW"
	case "http":
		originProtocol = "HTTP"
	case "https":
		originProtocol = "HTTPS"
	}
	if httpPort == 0 {
		httpPort = 80
	}
	if httpsPort == 0 {
		httpsPort = 443
	}

	// 请求
	apiResp := tencent_helper.ResponseUpdateDomain{}
	_, err = c.client.R().
		SetXTCAction("ModifyAccelerationDomain").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId":     c.config.ZoneID, // 加速域名所属站点ID
			"DomainName": domainFull,      // 加速域名名称
			"OriginInfo": map[string]any{
				"OriginType": originType,  // 源站类型
				"Origin":     originValue, // 源站地址
			}, // 源站信息
			"OriginProtocol":  originProtocol, // 回源协议
			"HttpOriginPort":  httpPort,       // HTTP回源端口
			"HttpsOriginPort": httpsPort,      // HTTPS回源端口
			"IPv6Status":      "on",           // IPv6状态
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return err
	}

	return nil
}

// 获取源站组列表
type ResponseEdgeOneQueryOriginGroup struct {
	GroupID      string   // 源站组ID
	GroupName    string   // 源站组名称
	GroupType    string   // 源站组类型
	GroupOrigins []string // 源站组类型值
}

// 获取源站组列表
// https://cloud.tencent.com/document/api/1552/80594
// domain: 源站组名称
func (c *EdgeOneClient) QueryOriginGroup(ctx context.Context, originGroupName string) (resp ResponseEdgeOneQueryOriginGroup, err error) {

	// 请求
	apiResp := tencent_helper.ResponseQueryOriginGroup{}
	_, err = c.client.R().
		SetXTCAction("DescribeOriginGroup").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId": c.config.ZoneID, // 站点ID
			"Offset": 0,               // 分页查询偏移量
			"Limit":  1000,            // 分页查询限制数目
			"Filters": []map[string]any{
				{
					"Name":   "origin-group-name", // 按照源站组名称进行过滤
					"Values": []string{originGroupName},
				},
			},
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return resp, err
	}

	for _, item := range apiResp.Response.OriginGroups {
		if item.Name == originGroupName {
			origins := []string{}
			for _, v := range item.Records {
				origins = append(origins, v.Record)
			}
			return ResponseEdgeOneQueryOriginGroup{
				GroupID:      item.GroupId,
				GroupName:    item.Name,
				GroupType:    item.Type,
				GroupOrigins: origins,
			}, nil
		}
	}

	return resp, nil
}

// 创建源站组
// https://cloud.tencent.com/document/api/1552/80598
// domain: 源站组名称
// origins: 源站组IP列表
func (c *EdgeOneClient) CreateOriginGroup(ctx context.Context, name string, origins []string) (err error) {

	records := []map[string]any{}
	for _, v := range origins {
		records = append(records, map[string]any{
			"Record": v,
			"Type":   "IP_DOMAIN",
		})
	}

	// 请求
	apiResp := tencent_helper.ResponseCreateOriginGroup{}
	_, err = c.client.R().
		SetXTCAction("CreateOriginGroup").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId":  c.config.ZoneID, // 站点ID
			"Records": records,         // 源站记录信息
			"Name":    name,            // 源站组名称
			"Type":    "GENERAL",       // 源站组类型
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return err
	}

	return nil
}

// 修改源站组
// https://cloud.tencent.com/document/api/1552/80592
// domain: 源站组ID
// origins: 源站组IP列表
func (c *EdgeOneClient) UpdateOriginGroup(ctx context.Context, groupId string, origins []string) (err error) {

	records := []map[string]any{}
	for _, v := range origins {
		records = append(records, map[string]any{
			"Record": v,
			"Type":   "IP_DOMAIN",
		})
	}

	// 请求
	apiResp := tencent_helper.ResponseUpdateOriginGroup{}
	_, err = c.client.R().
		SetXTCAction("ModifyOriginGroup").
		SetXTCVersion("2022-09-01").
		SetBodyMap(map[string]any{
			"ZoneId":  c.config.ZoneID, // 站点ID
			"GroupId": groupId,         // 源站组ID
			"Records": records,         // 源站记录信息
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return err
	}

	return nil
}
