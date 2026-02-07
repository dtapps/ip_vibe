package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/helper/domain_helper"
	"vibe_ip/internal/provider"

	"go.dtapp.net/library/contrib/http_log"
)

// DomainMonitorTaskProvider 实现 provider.TaskProvider 接口
type DomainMonitorTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
}

// NewDomainMonitorTaskProvider 创建一个新的域名监控任务提供者
func NewDomainMonitorTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewDomainMonitorTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewDomainMonitorTaskProvider 创建一个新的域名监控任务提供者
func NewDomainMonitorTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *DomainMonitorTaskProvider {
	return &DomainMonitorTaskProvider{
		db:            db,
		logDb:         logDb,
		logger:        logger.With(slog.String("handler", "domain-monitor")),
		httpTransport: httpTransport,
	}
}

// DomainMonitorTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	DomainMonitorTaskName = "domain_monitor"
)

// DomainMonitorTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type DomainMonitorTaskArgs struct {
	ID int64 `json:"id"` // DomainMonitor 的数据库 ID
}

// Process 执行单条 DomainMonitor 记录的域名监控逻辑
func (p *DomainMonitorTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params DomainMonitorTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	domain, err := p.db.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(params.ID)).
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("获取记录失败: %w", err)
	}

	// 判断是否已禁用
	if !domain.IsActive && meta.TaskType == cronjob.TaskTypeRecurring {
		return "跳过: 记录已禁用", nil
	}

	// 执行检查
	domainCheck := domain_helper.NewDomainCheck(p.logger, p.httpTransport)
	result := domainCheck.Check(
		ctx,
		domain.CheckProtocol.String(),
		domain.Domain,
		domain.RequestHeaders,
	)

	// 历史记录创建
	historyCreator := p.logDb.DomainMonitorHistory.Create()
	historyCreator.SetMonitorID(domain.ID)
	historyCreator.SetCheckTime(time.Now())
	historyCreator.SetIsSuccess(result.IsSuccess)
	historyCreator.SetCheckDuration(int(result.Duration.Milliseconds()))
	historyCreator.SetStatusMessage(result.StatusMsg)
	historyCreator.SetHTTPStatus(result.HTTPStatus)
	historyCreator.SetCheckType(domainmonitorhistory.CheckType(result.CheckType))
	historyCreator.SetIPResolved(result.IPResolved)
	historyCreator.SetErrorLog(result.ErrorLog)

	// 填充 SSL 字段
	if domain.CheckProtocol == domainmonitor.CheckProtocolHTTPS {
		historyCreator.SetSslNotDate(result.SSLNotDate)
		historyCreator.SetSslExpiryDate(result.SSLExpiryDate)
		historyCreator.SetSslCommonName(result.SSLCommonName)
		historyCreator.SetSslIssuerName(result.SSLIssuerName)
		historyCreator.SetSslSignatureAlg(result.SSLSignatureAlg)
		historyCreator.SetSslSanList(result.SSLSANList)
		historyCreator.SetSslCertificateChain(result.SSLCertificateChain)
		historyCreator.SetSslSerialNumber(result.SSLSerialNumber)
		historyCreator.SetSslVersion(result.SSLVersion)
	}

	// 更新域名监控状态
	updaterOne := p.db.DomainMonitor.UpdateOne(domain)
	if result.IsSuccess {
		updaterOne.SetLastSuccessAt(time.Now())
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusSuccess)
	} else {
		updaterOne.SetLastFailureAt(time.Now())
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusFailure)
	}
	err = updaterOne.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("更新域名监控状态失败: %w", err)
	}

	return fmt.Sprintf("域名监控 %s 成功", domain.Domain), nil
}
