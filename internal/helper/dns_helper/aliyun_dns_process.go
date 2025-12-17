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
type AliyunDnsProcessSingleIPParams struct {
	ConfigJSON  json.RawMessage // 配置
	SubDomain   string          // 子域名
	OriginValue string          // 源站地址
}

// 处理单IP
func AliyunDnsProcessSingleIP(ctx context.Context, service ProcessService, params AliyunDnsProcessSingleIPParams) (err error) {

	var platformConfig struct {
		AccessKeyID     string `json:"access_key_id"`     // 密钥ID
		AccessKeySecret string `json:"access_key_secret"` // 密钥Key
		DomainName      string `json:"domain_name"`       // 主域名
	}
	err = json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 初始化客户端
	client, err := NewAliyunDnsClient(
		ctx,
		platformConfig.AccessKeyID,
		platformConfig.AccessKeySecret,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	// client.SetDebug(service.Debug)
	// 设置HTTP日志记录器
	client.SetLog(service.HttpLog)

	// 获取域名的解析记录列表
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 请求",
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
	service.Log.InfoContext(ctx, "获取域名的解析记录列表 响应",
		slog.String("domain_id", queryDomain.DomainID),
		slog.String("domain_name", queryDomain.DomainName),
		slog.String("origin_type", queryDomain.OriginType),
		slog.String("origin_value", queryDomain.OriginValue),
	)

	if queryDomain.DomainID == "" {

		// 添加记录
		service.Log.InfoContext(ctx, "添加记录 请求",
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
			return fmt.Errorf("添加记录失败: %w", err)
		}

		// 重新获取域名的解析记录列表
		service.Log.InfoContext(ctx, "重新获取域名的解析记录列表 请求",
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
		service.Log.InfoContext(ctx, "重新查询加速域名列表 响应",
			slog.String("domain_id", queryDomain.DomainID),
			slog.String("domain_name", queryDomain.DomainName),
			slog.String("origin_type", queryDomain.OriginType),
			slog.String("origin_value", queryDomain.OriginValue),
		)

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(fmt.Sprintf("%v", queryDomain.DomainID)).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)

		service.Log.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新DNS记录
		_ = service.Db.DNSRecord.UpdateOneID(service.RecordID).
			SetRecordID(fmt.Sprintf("%v", queryDomain.DomainID)).
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
