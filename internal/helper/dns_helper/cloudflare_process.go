package dns_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/utils"
)

// 处理单IP参数
type CloudflareProcessSingleIPParams struct {
	ConfigJSON  json.RawMessage // 配置
	SubDomain   string          // 子域名
	OriginValue string          // 源站地址
}

// 处理单IP
func CloudflareProcessSingleIP(ctx context.Context, service ProcessService, params CloudflareProcessSingleIPParams) (err error) {

	var platformConfig struct {
		ApiKey     string `json:"api_key"`
		ZoneID     string `json:"zone_id"`
		DomainName string `json:"domain_name"` // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewCloudflareClientSdk(
		ctx,
		platformConfig.ApiKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化Cloudflare客户端失败: %w", err)
	}
	// 设置调试模式
	client.SetDebug(service.Debug)

	// 查询域名记录
	service.Log.InfoContext(ctx, "查询域名记录 请求",
		slog.String("domain_name", platformConfig.DomainName),
		slog.String("sub_domain", params.SubDomain),
		slog.String("origin_value", params.OriginValue),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		platformConfig.DomainName,
		params.SubDomain,
		params.OriginValue,
	)
	service.Log.InfoContext(ctx, "查询域名记录 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("record_type", queryDomain.RecordType),
		slog.String("origin_value", queryDomain.OriginValue),
	)

	if queryDomain.DomainID == "" {

		// 创建域名记录
		service.Log.InfoContext(ctx, "创建域名记录 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_value", params.OriginValue),
		)
		err = client.CreateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			params.OriginValue,
		)
		if err != nil {
			return fmt.Errorf("创建域名记录失败: %w", err)
		}

		// 重新查询域名记录
		service.Log.InfoContext(ctx, "重新查询域名记录 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_value", params.OriginValue),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			params.OriginValue,
		)
		service.Log.InfoContext(ctx, "重新查询域名记录 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("record_type", queryDomain.RecordType),
			slog.String("origin_value", queryDomain.OriginValue),
		)

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		if utils.DNSValuesEqual(queryDomain.OriginValue, params.OriginValue) {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新域名记录
		service.Log.InfoContext(ctx, "更新域名记录 请求",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("origin_value", params.OriginValue),
		)
		err = client.UpdateDomain(
			ctx,
			queryDomain.DomainID,
			platformConfig.DomainName,
			params.SubDomain,
			params.OriginValue,
		)
		if err != nil {
			return fmt.Errorf("更新域名记录失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}
