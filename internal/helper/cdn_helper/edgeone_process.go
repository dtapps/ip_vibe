package cdn_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/utils/goarray"
)

// 处理单IP参数
type EdgeOneProcessSingleIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	DomainName     string          // 加速域名
	DomainOrigin   string          // IP地址
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
}

// 处理单IP
func EdgeOneProcessSingleIP(ctx context.Context, service ProcessService, params EdgeOneProcessSingleIPParams) (err error) {

	var originType = "IP_DOMAIN"

	var platformConfig struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		ZoneID    string `json:"zone_id"`
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewEdgeOneClient(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化EdgeOne客户端失败: %w", err)
	}
	// 开启调试模式
	if service.Debug {
		client.EnableDebug()
	}

	// 查询加速域名列表
	service.Log.InfoContext(ctx, "查询加速域名列表 请求",
		slog.String("domain_name", params.DomainName),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		params.DomainName,
	)
	service.Log.InfoContext(ctx, "查询加速域名列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_cname", queryDomain.DomainCname),
		slog.String("domain_type", queryDomain.DomainType),
		slog.String("domain_origin", queryDomain.DomainOrigin),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		service.Log.InfoContext(ctx, "创建加速域名 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.DomainOrigin),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.CreateDomain(
			ctx,
			params.DomainName,
			originType,
			params.DomainOrigin,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			return fmt.Errorf("创建加速域名失败: %w", err)
		}

		// 重新查询加速域名列表
		service.Log.InfoContext(ctx, "重新查询加速域名列表 请求",
			slog.String("domain_name", params.DomainName),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			params.DomainName,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_cname", queryDomain.DomainCname),
			slog.String("domain_type", queryDomain.DomainType),
			slog.String("domain_origin", queryDomain.DomainOrigin),
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

		if utils.DNSValuesEqual(queryDomain.DomainOrigin, params.DomainOrigin) {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 修改加速域名信息
		service.Log.InfoContext(ctx, "修改加速域名信息 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("origin_type", originType),
			slog.String("origin_value", params.DomainOrigin),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.UpdateDomain(
			ctx,
			params.DomainName,
			originType,
			params.DomainOrigin,
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
type EdgeOneProcessMultiIPParams struct {
	ConfigJSON     json.RawMessage // 配置
	DomainName     string          // 加速域名
	DomainOrigins  []string        // IP地址列表
	OriginProtocol string          // 回源协议
	HttpPort       int             // HTTP回源端口
	HttpsPort      int             // HTTPS回源端口
	GroupName      string          // 源站组名
}

// 处理多IP
func EdgeOneProcessMultiIP(ctx context.Context, service ProcessService, params EdgeOneProcessMultiIPParams) (err error) {

	var originType = "ORIGIN_GROUP"

	var platformConfig struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		ZoneID    string `json:"zone_id"`
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewEdgeOneClient(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化EdgeOne客户端失败: %w", err)
	}
	// 开启调试模式
	if service.Debug {
		client.EnableDebug()
	}

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
			slog.Any("group_origins", params.DomainOrigins),
		)
		err = client.CreateOriginGroup(
			ctx,
			params.GroupName,
			params.DomainOrigins,
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

		if !goarray.CompareStringSlicesIgnoreOrder(queryGroup.GroupOrigins, params.DomainOrigins) {

			// 修改源站组
			service.Log.InfoContext(ctx, "修改源站组 请求",
				slog.String("group_id", queryGroup.GroupID),
				slog.Any("group_records", params.DomainOrigins),
			)
			err = client.UpdateOriginGroup(
				ctx,
				queryGroup.GroupID,
				params.DomainOrigins,
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
		slog.String("domain_name", params.DomainName),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		params.DomainName,
	)
	service.Log.InfoContext(ctx, "查询加速域名列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_cname", queryDomain.DomainCname),
		slog.String("domain_type", queryDomain.DomainType),
		slog.String("domain_origin", queryDomain.DomainOrigin),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		service.Log.InfoContext(ctx, "创建加速域名 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("origin_type", originType),
			slog.String("origin_value", queryGroup.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.CreateDomain(
			ctx,
			params.DomainName,
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
			slog.String("domain_name", params.DomainName),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			params.DomainName,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_cname", queryDomain.DomainCname),
			slog.String("domain_type", queryDomain.DomainType),
			slog.String("domain_origin", queryDomain.DomainOrigin),
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

		if queryGroup.GroupID == queryDomain.DomainOrigin {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 修改加速域名信息
		service.Log.InfoContext(ctx, "修改加速域名信息 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("origin_type", originType),
			slog.String("origin_value", queryGroup.GroupID),
			slog.String("origin_protocol", params.OriginProtocol),
			slog.Int("http_port", params.HttpPort),
			slog.Int("https_port", params.HttpsPort),
		)
		err = client.UpdateDomain(
			ctx,
			params.DomainName,
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
