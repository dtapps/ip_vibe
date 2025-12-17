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
type TencentEOProcessSingleIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	SubDomain      string          // 子域名
	OriginValue    string          // 记录值
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
}

// 处理单IP
func TencentEOProcessSingleIP(ctx context.Context, service ProcessService, params TencentEOProcessSingleIPParams) (err error) {

	var originType = "IP_DOMAIN"

	var platformConfig struct {
		SecretID   string `json:"secret_id"`   // 密钥ID
		SecretKey  string `json:"secret_key"`  // 密钥Key
		ZoneID     string `json:"zone_id"`     // 区域ID
		DomainName string `json:"domain_name"` // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewTencentEOClient(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	client.SetDebug(service.Debug)

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
		slog.String("origin_protocol", queryDomain.OriginProtocol),
		slog.Int("http_origin_port", queryDomain.HttpOriginPort),
		slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
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
			slog.String("origin_protocol", queryDomain.OriginProtocol),
			slog.Int("http_origin_port", queryDomain.HttpOriginPort),
			slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
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
		// 比较回源协议
		if !strings.EqualFold(queryDomain.OriginProtocol, params.OriginProtocol) {
			isUpdate = true
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTP) {
			// 比较HTTP回源端口
			if queryDomain.HttpOriginPort != params.HttpPort {
				isUpdate = true
			}
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTPS) {
			// 比较HTTPS回源端口
			if queryDomain.HttpsOriginPort != params.HttpsPort {
				isUpdate = true
			}
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolFOLLOW) {
			// 比较HTTP回源端口
			if queryDomain.HttpOriginPort != params.HttpPort {
				isUpdate = true
			}
			// 比较HTTPS回源端口
			if queryDomain.HttpsOriginPort != params.HttpsPort {
				isUpdate = true
			}
		}
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
			platformConfig.DomainName,
			params.SubDomain,
			originType,
			params.OriginValue,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("修改加速域名信息失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}

// 处理多IP参数
type TencentEOProcessMultiIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	SubDomain      string          // 子域名
	OriginValues   []string        // 记录值列表
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
	GroupName      string          // 源站组名
}

// 处理多IP
func TencentEOProcessMultiIP(ctx context.Context, service ProcessService, params TencentEOProcessMultiIPParams) (err error) {

	var originType = "ORIGIN_GROUP"

	var platformConfig struct {
		SecretID   string `json:"secret_id"`   // 密钥ID
		SecretKey  string `json:"secret_key"`  // 密钥Key
		ZoneID     string `json:"zone_id"`     // 区域ID
		DomainName string `json:"domain_name"` // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewTencentEOClient(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	client.SetDebug(service.Debug)

	// 获取源站组列表
	service.Log.InfoContext(ctx, "获取源站组列表 请求",
		slog.String("group_name", params.GroupName),
	)
	queryGroup, _ := client.QueryOriginGroup(
		ctx,
		params.GroupName,
	)
	service.Log.InfoContext(ctx, "获取源站组列表 响应",
		slog.String("group_id", queryGroup.GroupID),
		slog.String("group_name", queryGroup.GroupName),
		slog.String("group_type", queryGroup.GroupType),
		slog.Any("group_origins", queryGroup.GroupOrigins),
	)
	if queryGroup.GroupID == "" {

		// 创建源站组
		service.Log.InfoContext(ctx, "创建源站组 请求",
			slog.String("group_name", params.GroupName),
			slog.Any("group_origins", params.OriginValues),
		)
		err = client.CreateOriginGroup(
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
		queryGroup, _ = client.QueryOriginGroup(
			ctx,
			params.GroupName,
		)
		service.Log.InfoContext(ctx, "重新获取源站组列表 响应",
			slog.String("group_id", queryGroup.GroupID),
			slog.String("group_name", queryGroup.GroupName),
			slog.String("group_type", queryGroup.GroupType),
			slog.Any("group_origins", queryGroup.GroupOrigins),
		)

		// 更新源站组记录
		_ = service.Db.IpGroup.UpdateOneID(service.GroupID).
			SetRecordID(queryGroup.GroupID).
			Exec(ctx)

	} else {

		if !goarray.CompareStringSlicesIgnoreOrder(queryGroup.GroupOrigins, params.OriginValues) {

			// 修改源站组
			service.Log.InfoContext(ctx, "修改源站组 请求",
				slog.String("group_id", queryGroup.GroupID),
				slog.Any("group_records", params.OriginValues),
			)
			err = client.UpdateOriginGroup(
				ctx,
				queryGroup.GroupID,
				params.OriginValues,
			)
			if err != nil {
				return fmt.Errorf("修改源站组失败: %w", err)
			}

			// 更新源站组记录
			_ = service.Db.IpGroup.UpdateOneID(service.GroupID).
				SetRecordID(queryGroup.GroupID).
				Exec(ctx)

		}

	}
	if queryGroup.GroupID == "" {
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
		slog.String("origin_protocol", queryDomain.OriginProtocol),
		slog.Int("http_origin_port", queryDomain.HttpOriginPort),
		slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		service.Log.InfoContext(ctx, "创建加速域名 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_type", originType),
			slog.String("origin_value", queryGroup.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.CreateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			originType,
			queryGroup.GroupID,
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
			slog.String("origin_protocol", queryDomain.OriginProtocol),
			slog.Int("http_origin_port", queryDomain.HttpOriginPort),
			slog.Int("https_origin_port", queryDomain.HttpsOriginPort),
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
		if !strings.EqualFold(queryGroup.GroupID, queryDomain.OriginValue) {
			isUpdate = true
		}
		// 比较回源协议
		if !strings.EqualFold(queryDomain.OriginProtocol, params.OriginProtocol) {
			isUpdate = true
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTP) {
			// 比较HTTP回源端口
			if queryDomain.HttpOriginPort != params.HttpPort {
				isUpdate = true
			}
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolHTTPS) {
			// 比较HTTPS回源端口
			if queryDomain.HttpsOriginPort != params.HttpsPort {
				isUpdate = true
			}
		}
		if strings.EqualFold(queryDomain.OriginProtocol, OriginProtocolFOLLOW) {
			// 比较HTTP回源端口
			if queryDomain.HttpOriginPort != params.HttpPort {
				isUpdate = true
			}
			// 比较HTTPS回源端口
			if queryDomain.HttpsOriginPort != params.HttpsPort {
				isUpdate = true
			}
		}
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
			slog.String("origin_value", queryGroup.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.UpdateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			originType,
			queryGroup.GroupID,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("修改加速域名信息失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}
