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
type DnsPodProcessSingleIPParams struct {
	ConfigJSON   json.RawMessage // 配置
	DomainName   string          // 加速域名
	DomainOrigin string          // IP地址
}

// 处理单IP
func DnsPodProcessSingleIP(ctx context.Context, service ProcessService, params DnsPodProcessSingleIPParams) (err error) {

	var platformConfig struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewDnsPodClient(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
	)
	if err != nil {
		return fmt.Errorf("初始化DnsPod客户端失败: %w", err)
	}
	// 开启调试模式
	if service.Debug {
		client.EnableDebug()
	}

	// 获取域名的解析记录列表
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 请求",
		slog.String("domain_name", params.DomainName),
		slog.String("domain_origin", params.DomainOrigin),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		params.DomainName,
		params.DomainOrigin,
	)
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 响应",
		slog.Int64("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("domain_type", queryDomain.DomainType),
		slog.String("domain_origin", queryDomain.DomainOrigin),
	)

	if queryDomain.DomainID == 0 {

		// 添加记录
		service.Log.InfoContext(ctx, "添加记录 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("domain_origin", params.DomainOrigin),
		)
		err = client.CreateDomain(
			ctx,
			params.DomainName,
			params.DomainOrigin,
		)
		if err != nil {
			return fmt.Errorf("添加记录失败: %w", err)
		}

		// 重新获取域名的解析记录列表
		service.Log.InfoContext(ctx, "重新获取域名的解析记录列表 请求",
			slog.String("domain_name", params.DomainName),
			slog.String("domain_origin", params.DomainOrigin),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			params.DomainName,
			params.DomainOrigin,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.Int64("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("domain_type", queryDomain.DomainType),
			slog.String("domain_origin", queryDomain.DomainOrigin),
		)

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(fmt.Sprintf("%d", queryDomain.DomainID)).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(fmt.Sprintf("%d", queryDomain.DomainID)).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		if utils.DNSValuesEqual(queryDomain.DomainOrigin, params.DomainOrigin) {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新域名记录
		service.Log.InfoContext(ctx, "更新域名记录 请求",
			slog.Int64("domain_id", queryDomain.DomainID),
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
