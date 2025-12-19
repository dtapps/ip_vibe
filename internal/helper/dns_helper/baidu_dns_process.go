package dns_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
)

// 实现 DNSProcess 接口
type DNSProcessBaiduDnsImpl struct {
	debug         bool                       // 调试模式
	logger        *slog.Logger               // 日志
	db            *ent.Client                // 数据库
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
}

// NewDNSProcessBaiduDns 创建实例
func NewDNSProcessBaiduDns(debug bool, logger *slog.Logger, db *ent.Client, httpTransport *http_log.LoggingTransport) DNSProcess {
	return &DNSProcessBaiduDnsImpl{
		debug:         debug,
		logger:        logger,
		db:            db,
		httpTransport: httpTransport,
	}
}

// 处理记录
func (d *DNSProcessBaiduDnsImpl) UpsertRecord(ctx context.Context, params *DNSProcessParams) error {

	var platformConfig struct {
		AccessKeyID     string `json:"access_key_id"`     // 密钥ID
		SecretAccessKey string `json:"secret_access_key"` // 密钥Key
		DomainName      string `json:"domain_name"`       // 主域名
	}
	err := json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	d.logger.InfoContext(ctx, "配置",
		slog.String("主域名", platformConfig.DomainName),
	)
	d.logger.InfoContext(ctx, "入参",
		slog.String("DNS记录ID", params.DNSRecordID),
		slog.String("子域名", params.SubDomain),
		slog.String("源站值", params.OriginValue),
	)

	// 解析域名
	fullDomain, subDomain, domainName := utils.ParseDomain(
		params.SubDomain,
		platformConfig.DomainName,
	)
	// 解析域名类型
	recordType, originValue, err := utils.ParseDnsDomainType(params.OriginValue)
	if err != nil {
		return fmt.Errorf("解析域名类型失败: %w", err)
	}
	d.logger.InfoContext(ctx, "解析域名",
		slog.String("完整域名", fullDomain),
		slog.String("子域名", subDomain),
		slog.String("主域名", domainName),
		slog.String("记录类型", recordType),
		slog.String("源站值", originValue),
	)

	// 初始化客户端
	client, err := NewBaiduDnsClient(
		ctx,
		platformConfig.AccessKeyID,
		platformConfig.SecretAccessKey,
	)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %w", err)
	}
	// 设置调试模式
	client.SetDebug(d.debug)
	// 设置HTTP日志记录器
	if d.httpTransport != nil {
		client.SetLog(d.httpTransport)
	}

	// 查询域名记录
	queryDomain, _ := client.QueryDomain(
		ctx,
		domainName,
		subDomain,
		recordType,
	)
	d.logger.InfoContext(ctx, "查询域名记录",
		slog.String("记录ID", queryDomain.RecordID),
		slog.String("完整域名", queryDomain.FullDomain),
		slog.String("记录类型", queryDomain.RecordType),
		slog.String("源站值", queryDomain.OriginValue),
	)

	if queryDomain.RecordID == "" {

		// 创建域名记录
		err = client.CreateDomain(
			ctx,
			domainName,
			subDomain,
			recordType,
			originValue,
		)
		if err != nil {
			d.logger.ErrorContext(ctx, "创建域名记录错误",
				slog.Any("error", err),
			)
			return fmt.Errorf("创建域名记录失败: %w", err)
		}

		// 重新查询域名记录
		queryDomain, _ = client.QueryDomain(
			ctx,
			domainName,
			subDomain,
			recordType,
		)
		d.logger.InfoContext(ctx, "重新查询域名记录",
			slog.String("记录ID", queryDomain.RecordID),
			slog.String("完整域名", queryDomain.FullDomain),
			slog.String("记录类型", queryDomain.RecordType),
			slog.String("源站值", queryDomain.OriginValue),
		)

		// 更新DNS记录
		if d.db != nil && params.DNSRecordID != "" {
			_ = d.db.DNSRecord.UpdateOneID(cast.ToInt64(params.DNSRecordID)).
				SetRecordID(queryDomain.RecordID).
				SetLastRefreshAt(time.Now()).
				Exec(ctx)
		}

		d.logger.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新DNS记录
		if d.db != nil && params.DNSRecordID != "" {
			_ = d.db.DNSRecord.UpdateOneID(cast.ToInt64(params.DNSRecordID)).
				SetRecordID(queryDomain.RecordID).
				SetLastRefreshAt(time.Now()).
				Exec(ctx)
		}

		// 是否更新
		var isUpdate = false

		if utils.DNSValuesEqual(queryDomain.OriginValue, params.OriginValue) {
			isUpdate = true
		}

		// 判断是否需要更新
		if !isUpdate {
			d.logger.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新域名记录
		err = client.UpdateDomain(
			ctx,
			queryDomain.RecordID,
			domainName,
			subDomain,
			recordType,
			originValue,
		)
		if err != nil {
			d.logger.ErrorContext(ctx, "更新域名记录错误",
				slog.Any("error", err),
			)
			return fmt.Errorf("更新域名记录失败: %w", err)
		}

		d.logger.InfoContext(ctx, "更新成功")
		return nil
	}
}
