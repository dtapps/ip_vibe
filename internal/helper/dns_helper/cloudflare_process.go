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
	ConfigJSON   json.RawMessage // 配置
	DomainName   string          // 加速域名
	DomainOrigin string          // IP地址
}

// 处理单IP
func CloudflareProcessSingleIP(ctx context.Context, service ProcessService, params CloudflareProcessSingleIPParams) (err error) {

	var platformConfig struct {
		ApiKey string `json:"api_key"`
		ZoneID string `json:"zone_id"`
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewCloudflareClient(
		ctx,
		platformConfig.ApiKey,
		platformConfig.ZoneID,
	)
	if err != nil {
		return fmt.Errorf("初始化Cloudflare客户端失败: %w", err)
	}
	// 开启调试模式
	if service.Debug {
		client.EnableDebug()
	}

	// 查询域名记录
	service.Log.InfoContext(ctx, "查询域名记录 请求",
		slog.String("domain_name", params.DomainName),
		slog.String("domain_origin", params.DomainOrigin),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		params.DomainName,
		params.DomainOrigin,
	)
	service.Log.InfoContext(ctx, "查询域名记录 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_type", queryDomain.DomainType),
		slog.String("domain_origin", queryDomain.DomainOrigin),
	)

	if queryDomain.DomainID == "" {

		// 创建域名记录
		service.Log.InfoContext(ctx, "创建域名记录 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("domain_origin", params.DomainOrigin),
		)
		err = client.CreateDomain(
			ctx,
			params.DomainName,
			params.DomainOrigin,
		)
		if err != nil {
			return fmt.Errorf("创建域名记录失败: %w", err)
		}

		// 重新查询域名记录
		service.Log.InfoContext(ctx, "重新查询域名记录 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("domain_origin", params.DomainOrigin),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			params.DomainName,
			params.DomainOrigin,
		)
		service.Log.InfoContext(ctx, "重新查询域名记录 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_type", queryDomain.DomainType),
			slog.String("domain_origin", queryDomain.DomainOrigin),
		)

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetLastRefresh(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(queryDomain.DomainID).
			SetLastRefresh(time.Now()).
			Exec(ctx)

		if utils.DNSValuesEqual(queryDomain.DomainOrigin, params.DomainOrigin) {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新域名记录
		service.Log.InfoContext(ctx, "更新域名记录 请求",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", params.DomainName),
			slog.String("domain_origin", params.DomainOrigin),
		)
		err = client.UpdateDomain(
			ctx,
			queryDomain.DomainID,
			params.DomainName,
			params.DomainOrigin,
		)
		if err != nil {
			return fmt.Errorf("更新域名记录失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}
