package cdn_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/goarray"
)

// 实现 CDNProcess 接口
type CDNProcessTencentEOImpl struct {
	db            *ent.Client                // 数据库
	logger        *slog.Logger               // 日志
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 调试模式
}

// 创建实例
func NewCDNProcessTencentEO(db *ent.Client, logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) CDNProcess {
	return &CDNProcessTencentEOImpl{
		db:            db,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// 处理记录
func (d *CDNProcessTencentEOImpl) UpsertRecord(ctx context.Context, params *CDNProcessParams) error {

	var platformConfig TencentEOConfig
	err := json.Unmarshal(params.ConfigJSON, &platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	d.logger.InfoContext(ctx, "配置",
		slog.String("区域ID", platformConfig.ZoneID),
		slog.String("主域名", platformConfig.DomainName),
	)
	d.logger.InfoContext(ctx, "入参",
		slog.String("CDN记录ID", params.CDNRecordID),
		slog.String("源站组记录ID", params.GroupRecordID),

		slog.String("源站组名称", params.GroupName),
		slog.String("子域名", params.SubDomain),
		slog.Any("源站地址列表", params.OriginValues),
		slog.String("回源协议", params.OriginProtocol),
		slog.Int("HTTP回源端口", params.HttpPort),
		slog.Int("HTTPS回源端口", params.HttpsPort),
	)

	// 解析域名
	fullDomain, subDomain, domainName := utils.ParseDomain(
		params.SubDomain,
		platformConfig.DomainName,
	)
	// 解析域名类型
	recordType, originValues, err := utils.ParseCdnDomainType(schematype.ConfigPlatformTencentEO, params.OriginValues)
	if err != nil {
		return fmt.Errorf("解析域名类型失败: %w", err)
	}
	d.logger.InfoContext(ctx, "解析域名",
		slog.String("完整域名", fullDomain),
		slog.String("子域名", subDomain),
		slog.String("主域名", domainName),
		slog.String("记录类型", recordType),
		slog.Any("源站地址列表", originValues),
	)

	// 初始化客户端
	client, err := NewTencentEOClient(
		ctx,
		&platformConfig,
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

	// 判断是否需要组
	isGroup := len(originValues) >= 2
	if isGroup {

		if params.GroupName == "" {
			return fmt.Errorf("源站组名称不存在")
		}

		// 查询源站组
		queryGroup, _ := client.QueryOriginGroup(
			ctx,
			params.GroupName,
		)
		d.logger.InfoContext(ctx, "查询源站组",
			slog.String("源站组ID", queryGroup.GroupID),
			slog.String("源站组名称", queryGroup.GroupName),
			slog.String("源站组类型", queryGroup.GroupType),
			slog.Any("源站组值列表", queryGroup.GroupOrigins),
		)
		if queryGroup.GroupID == "" {

			// 创建源站组
			err = client.CreateOriginGroup(
				ctx,
				params.GroupName,
				params.OriginValues,
			)
			if err != nil {
				d.logger.ErrorContext(ctx, "创建源站组错误",
					slog.Any("error", err),
				)
				return fmt.Errorf("创建源站组失败: %w", err)
			}

			// 重新查询源站组
			queryGroup, _ = client.QueryOriginGroup(
				ctx,
				params.GroupName,
			)
			d.logger.InfoContext(ctx, "重新查询源站组",
				slog.String("源站组ID", queryGroup.GroupID),
				slog.String("源站组名称", queryGroup.GroupName),
				slog.String("源站组类型", queryGroup.GroupType),
				slog.Any("源站组值列表", queryGroup.GroupOrigins),
			)

			// 更新源站组记录
			if d.db != nil && params.GroupRecordID != "" {
				_ = d.db.IpGroup.UpdateOneID(cast.ToInt64(params.GroupRecordID)).
					SetRecordID(queryGroup.GroupID).
					Exec(ctx)
			}

		} else {

			// 更新源站组记录
			if d.db != nil && params.GroupRecordID != "" {
				_ = d.db.IpGroup.UpdateOneID(cast.ToInt64(params.GroupRecordID)).
					SetRecordID(queryGroup.GroupID).
					Exec(ctx)
			}

			// 是否更新
			var isUpdate = false

			// 比较
			if !goarray.CompareStringSlicesIgnoreOrder(queryGroup.GroupOrigins, originValues) {
				isUpdate = true
			}

			// 更新源站组
			if isUpdate {
				err = client.UpdateOriginGroup(
					ctx,
					queryGroup.GroupID,
					originValues,
				)
				if err != nil {
					d.logger.ErrorContext(ctx, "更新源站组错误",
						slog.Any("error", err),
					)
					return fmt.Errorf("更新源站组失败: %w", err)
				}
			}

		}
		if queryGroup.GroupID == "" {
			return fmt.Errorf("源站组不存在")
		}
		params.originValue = queryGroup.GroupID

	}

	if !isGroup {
		params.originValue = originValues[0]
	}

	// 查询加速域名
	queryDomain, _ := client.QueryDomain(
		ctx,
		fullDomain,
	)
	d.logger.InfoContext(ctx, "查询加速域名",
		slog.String("域名ID", queryDomain.RecordID),
		slog.String("完整域名", queryDomain.FullDomain),
		slog.String("域名别名", queryDomain.DomainCname),
		slog.String("记录类型", queryDomain.RecordType),
		slog.String("源站值", queryDomain.OriginValue),
		slog.String("回源协议", queryDomain.OriginProtocol),
		slog.Int("HTTP回源端口", queryDomain.HttpOriginPort),
		slog.Int("HTTPS回源端口", queryDomain.HttpsOriginPort),
	)

	if queryDomain.DomainCname == "" {

		// 创建加速域名
		err = client.CreateDomain(
			ctx,
			fullDomain,
			recordType,
			params.originValue,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			d.logger.ErrorContext(ctx, "创建加速域名错误",
				slog.Any("error", err),
			)
			return fmt.Errorf("创建加速域名失败: %w", err)
		}

		// 重新查询加速域名
		queryDomain, _ = client.QueryDomain(
			ctx,
			fullDomain,
		)
		d.logger.InfoContext(ctx, "重新查询加速域名",
			slog.String("域名ID", queryDomain.RecordID),
			slog.String("完整域名", queryDomain.FullDomain),
			slog.String("域名别名", queryDomain.DomainCname),
			slog.String("记录类型", queryDomain.RecordType),
			slog.String("源站值", queryDomain.OriginValue),
			slog.String("回源协议", queryDomain.OriginProtocol),
			slog.Int("HTTP回源端口", queryDomain.HttpOriginPort),
			slog.Int("HTTPS回源端口", queryDomain.HttpsOriginPort),
		)

		// 更新CDN记录的RecordID和CDN别名
		_ = d.updateDbOfApiInfo(ctx, params.CDNRecordID, queryDomain.RecordID, queryDomain.DomainCname)

		d.logger.InfoContext(ctx, "创建成功")
		return nil
	} else {

		// 更新CDN记录的RecordID和CDN别名
		_ = d.updateDbOfApiInfo(ctx, params.CDNRecordID, queryDomain.RecordID, queryDomain.DomainCname)

		// 是否更新
		var isUpdate = false

		// 比较源站组
		if !strings.EqualFold(params.originValue, queryDomain.OriginValue) {
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
			d.logger.InfoContext(ctx, "无需更新")
			return nil
		}

		// 更新加速域名
		err = client.UpdateDomain(
			ctx,
			fullDomain,
			recordType,
			params.originValue,
			params.OriginProtocol,
			params.HttpPort,
			params.HttpsPort,
		)
		if err != nil {
			d.logger.ErrorContext(ctx, "更新加速域名错误",
				slog.Any("error", err),
			)
			return fmt.Errorf("更新加速域名失败: %w", err)
		}

		d.logger.InfoContext(ctx, "更新成功")
		return nil
	}
}

// 更新CDN记录的RecordID和CDN别名
func (d *CDNProcessTencentEOImpl) updateDbOfApiInfo(ctx context.Context, id string, recordID string, cdnDomain string) error {
	if d.db != nil && id != "" {
		return d.db.CDNRecord.UpdateOneID(cast.ToInt64(id)).
			SetRecordID(recordID).
			SetCdnDomain(cdnDomain).
			SetLastRefreshAt(time.Now()).
			Exec(ctx)
	}
	return nil
}
