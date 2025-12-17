package cdn_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/utils/goarray"
)

// 处理单IP参数
type AliyunEsaProcessSingleIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	SubDomain      string          // 子域名
	OriginValue    string          // 记录值
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
}

// 处理单IP
func AliyunEsaProcessSingleIP(ctx context.Context, service ProcessService, params AliyunEsaProcessSingleIPParams) (err error) {

	var originType = "IP_DOMAIN"

	var platformConfig struct {
		AccessKeyID     string `json:"access_key_id"`     // 访问密钥ID
		AccessKeySecret string `json:"access_key_secret"` // 访问密钥Key
		SiteID          int64  `json:"site_id"`           // 站点ID
		DomainName      string `json:"domain_name"`       // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewAliyunEsaClient(
		ctx,
		platformConfig.AccessKeyID,
		platformConfig.AccessKeySecret,
		platformConfig.SiteID,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	// client.SetDebug(service.Debug)
	// 设置HTTP日志记录器
	client.SetLog(service.HttpLog)

	// 查询加速域名列表
	service.Log.InfoContext(ctx, "查询加速域名列表 请求",
		slog.String("domain_name", platformConfig.DomainName),
		slog.String("sub_domain", params.SubDomain),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		platformConfig.DomainName,
		params.SubDomain,
	)
	service.Log.InfoContext(ctx, "查询加速域名列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_cname", queryDomain.DomainCname),
		slog.String("origin_type", queryDomain.OriginType),
		slog.String("origin_value", queryDomain.OriginValue),
		// slog.String("origin_protocol", queryDomain.OriginProtocol),
		// slog.Int("http_origin_port", queryDomain.HttpOriginPort),
		// slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		service.Log.InfoContext(ctx, "创建加速域名 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.OriginValue),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.CreateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			originType,
			params.OriginValue,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("创建加速域名失败: %w", err)
		}

		// 重新查询加速域名列表
		service.Log.InfoContext(ctx, "重新查询加速域名列表 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_cname", queryDomain.DomainCname),
			slog.String("origin_type", queryDomain.OriginType),
			slog.String("origin_value", queryDomain.OriginValue),
			// slog.String("origin_protocol", queryDomain.OriginProtocol),
			// slog.Int("http_origin_port", queryDomain.HttpOriginPort),
			// slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
		)

		// 更新CDN记录
		_ = service.Db.CDNRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetCdnDomain(queryDomain.DomainCname).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新CDN记录
		_ = service.Db.CDNRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetCdnDomain(queryDomain.DomainCname).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		var isUpdate = false

		// 比较源站
		if !utils.DNSValuesEqual(queryDomain.OriginValue, params.OriginValue) {
			isUpdate = true
		}
		// // 比较回源协议
		// if !strings.EqualFold(queryDomain.OriginProtocol, params.OriginProtocol) {
		// 	isUpdate = true
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTP) {
		// 	// 比较HTTP回源端口
		// 	if queryDomain.HttpOriginPort != params.HttpPort {
		// 		isUpdate = true
		// 	}
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTPS) {
		// 	// 比较HTTPS回源端口
		// 	if queryDomain.HttpsOriginPort != params.HttpsPort {
		// 		isUpdate = true
		// 	}
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolFOLLOW) {
		// 	// 比较HTTP回源端口
		// 	if queryDomain.HttpOriginPort != params.HttpPort {
		// 		isUpdate = true
		// 	}
		// 	// 比较HTTPS回源端口
		// 	if queryDomain.HttpsOriginPort != params.HttpsPort {
		// 		isUpdate = true
		// 	}
		// }
		// 判断是否需要更新
		if !isUpdate {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 修改加速域名信息
		service.Log.InfoContext(ctx, "修改加速域名信息 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.OriginValue),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.UpdateDomain(
			ctx,
			queryDomain.DomainID,
			platformConfig.DomainName,
			params.SubDomain,
			// originType,
			// params.OriginValue,
			// params.OriginProtocol,
			// params.HttpPort,
			// params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("修改加速域名信息失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}

// 处理多IP参数
type AliyunEsaProcessMultiIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	SubDomain      string          // 子域名
	OriginValues   []string        // 记录值列表
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
	GroupID        string          // 源站组ID
	GroupName      string          // 源站组名
}

// 处理多IP
func AliyunEsaProcessMultiIP(ctx context.Context, service ProcessService, params AliyunEsaProcessMultiIPParams) (err error) {

	var originType = "CNAME"

	var platformConfig struct {
		AccessKeyID     string `json:"access_key_id"`     // 访问密钥ID
		AccessKeySecret string `json:"access_key_secret"` // 访问密钥Key
		SiteID          int64  `json:"site_id"`           // 站点ID
		DomainName      string `json:"domain_name"`       // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewAliyunEsaClient(
		ctx,
		platformConfig.AccessKeyID,
		platformConfig.AccessKeySecret,
		platformConfig.SiteID,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	// client.SetDebug(service.Debug)
	// 设置HTTP日志记录器
	client.SetLog(service.HttpLog)

	// 获取源站组列表
	service.Log.InfoContext(ctx, "获取源站组列表 请求",
		slog.String("group_name", params.GroupName),
	)
	queryGroupOrigins, _ := client.QueryOriginGroup(
		ctx,
		params.GroupName,
	)
	service.Log.InfoContext(ctx, "获取源站组列表 响应",
		slog.Any("group_origins", queryGroupOrigins),
	)
	if params.GroupID == "" {

		// 创建源站组
		service.Log.InfoContext(ctx, "创建源站组 请求",
			slog.String("group_name", params.GroupName),
			slog.Any("group_origins", params.OriginValues),
		)
		params.GroupID, err = client.CreateOriginGroup(
			ctx,
			params.GroupName,
			params.OriginValues,
		)
		if err != nil {
			return fmt.Errorf("创建源站组失败: %w", err)
		}

		// 重新获取源站组列表
		service.Log.InfoContext(ctx, "重新获取源站组列表 请求",
			slog.String("group_name", params.GroupName),
		)
		queryGroupOrigins, _ = client.QueryOriginGroup(
			ctx,
			params.GroupName,
		)
		service.Log.InfoContext(ctx, "重新获取源站组列表 响应",
			slog.Any("group_origins", queryGroupOrigins),
		)

		// 更新源站组记录
		_ = service.Db.IpGroup.UpdateOneID(service.GroupID).
			SetRecordID(params.GroupID).
			Exec(ctx)

	} else {

		if !goarray.CompareStringSlicesIgnoreOrder(queryGroupOrigins, params.OriginValues) {

			// 修改源站组
			service.Log.InfoContext(ctx, "修改源站组 请求",
				slog.String("group_id", params.GroupID),
				slog.Any("group_records", params.OriginValues),
			)
			err = client.UpdateOriginGroup(
				ctx,
				params.GroupID,
				params.OriginValues,
			)
			if err != nil {
				return fmt.Errorf("修改源站组失败: %w", err)
			}

			// 更新源站组记录
			_ = service.Db.IpGroup.UpdateOneID(service.GroupID).
				SetRecordID(params.GroupID).
				Exec(ctx)

		}

	}
	if params.GroupID == "" {
		return fmt.Errorf("源站组不存在")
	}

	// 查询加速域名列表
	service.Log.InfoContext(ctx, "查询加速域名列表 请求",
		slog.String("domain_name", platformConfig.DomainName),
		slog.String("sub_domain", params.SubDomain),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		platformConfig.DomainName,
		params.SubDomain,
	)
	service.Log.InfoContext(ctx, "查询加速域名列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_cname", queryDomain.DomainCname),
		slog.String("origin_type", queryDomain.OriginType),
		slog.String("origin_value", queryDomain.OriginValue),
		// slog.String("origin_protocol", queryDomain.OriginProtocol),
		// slog.Int("http_origin_port", queryDomain.HttpOriginPort),
		// slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		service.Log.InfoContext(ctx, "创建加速域名 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.CreateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			originType,
			params.GroupID,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("创建加速域名失败: %w", err)
		}

		// 重新查询加速域名列表
		service.Log.InfoContext(ctx, "重新查询加速域名列表 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_cname", queryDomain.DomainCname),
			slog.String("origin_type", queryDomain.OriginType),
			slog.String("origin_value", queryDomain.OriginValue),
			// slog.String("origin_protocol", queryDomain.OriginProtocol),
			// slog.Int("http_origin_port", queryDomain.HttpOriginPort),
			// slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
		)

		// 更新CDN记录
		_ = service.Db.CDNRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetCdnDomain(queryDomain.DomainCname).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新CDN记录
		_ = service.Db.CDNRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetCdnDomain(queryDomain.DomainCname).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		var isUpdate = false

		// 比较源站组
		if !strings.EqualFold(params.GroupID, queryDomain.OriginValue) {
			isUpdate = true
		}
		// 比较回源协议
		// if !strings.EqualFold(queryDomain.OriginProtocol, params.OriginProtocol) {
		// 	isUpdate = true
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTP) {
		// 	// 比较HTTP回源端口
		// 	if queryDomain.HttpOriginPort != params.HttpPort {
		// 		isUpdate = true
		// 	}
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTPS) {
		// 	// 比较HTTPS回源端口
		// 	if queryDomain.HttpsOriginPort != params.HttpsPort {
		// 		isUpdate = true
		// 	}
		// }
		// if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolFOLLOW) {
		// 	// 比较HTTP回源端口
		// 	if queryDomain.HttpOriginPort != params.HttpPort {
		// 		isUpdate = true
		// 	}
		// 	// 比较HTTPS回源端口
		// 	if queryDomain.HttpsOriginPort != params.HttpsPort {
		// 		isUpdate = true
		// 	}
		// }
		// 判断是否需要更新
		if !isUpdate {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 修改加速域名信息
		service.Log.InfoContext(ctx, "修改加速域名信息 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.UpdateDomain(
			ctx,
			queryDomain.DomainID,
			originType,
			params.GroupID,
			// params.OriginProtocol,
			// params.HttpPort,
			// params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("修改加速域名信息失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}
