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
	ConfigJSON  json.RawMessage // 配置
	SubDomain   string          // 子域名
	RecordValue string          // 记录值
}

// 处理单IP
func DnsPodProcessSingleIP(ctx context.Context, service ProcessService, params DnsPodProcessSingleIPParams) (err error) {

	var platformConfig struct {
		SecretID   string `json:"secret_id"`
		SecretKey  string `json:"secret_key"`
		DomainName string `json:"domain_name"` // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewDnsPodClientSdk(
		ctx,
		platformConfig.SecretID,
		platformConfig.SecretKey,
	)
	if err != nil {
		return fmt.Errorf("初始化DnsPod客户端失败: %w", err)
	}
	// 设置调试模式
	client.SetDebug(service.Debug)

	// 获取域名的解析记录列表
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 请求",
		slog.String("domain_name", platformConfig.DomainName),
		slog.String("sub_domain", params.SubDomain),
		slog.String("record_value", params.RecordValue),
	)
	queryDomain, _ := client.QueryDomain(
		ctx,
		platformConfig.DomainName,
		params.SubDomain,
		params.RecordValue,
	)
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("record_type", queryDomain.RecordType),
		slog.String("record_value", queryDomain.RecordValue),
	)

	if queryDomain.DomainID == "" {

		// 添加记录
		service.Log.InfoContext(ctx, "添加记录 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("record_value", params.RecordValue),
		)
		err = client.CreateDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			params.RecordValue,
		)
		if err != nil {
			return fmt.Errorf("添加记录失败: %w", err)
		}

		// 重新获取域名的解析记录列表
		service.Log.InfoContext(ctx, "重新获取域名的解析记录列表 请求",
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("record_value", params.RecordValue),
		)
		queryDomain, _ = client.QueryDomain(
			ctx,
			platformConfig.DomainName,
			params.SubDomain,
			params.RecordValue,
		)
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("record_type", queryDomain.RecordType),
			slog.String("record_value", queryDomain.RecordValue),
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

		if utils.DNSValuesEqual(queryDomain.RecordValue, params.RecordValue) {
			service.Log.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新域名记录
		service.Log.InfoContext(ctx, "更新域名记录 请求",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", platformConfig.DomainName),
			slog.String("sub_domain", params.SubDomain),
			slog.String("record_value", params.RecordValue),
		)
		err = client.UpdateDomain(
			ctx,
			queryDomain.DomainID,
			platformConfig.DomainName,
			params.SubDomain,
			params.RecordValue,
		)
		if err != nil {
			return fmt.Errorf("更新域名记录失败: %w", err)
		}

		service.Log.InfoContext(ctx, "更新成功")
		return nil
	}
}
