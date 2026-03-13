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
	"go.dtapp.net/library/utils/gojson"
)

// DomainMonitorTaskProvider 实现 provider.TaskProvider 接口
type DomainMonitorTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器

	businessType cronjob.BusinessType // 业务类型
}

// NewDomainMonitorTaskProvider 创建一个新的 域名监控 提供者
func NewDomainMonitorTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) provider.TaskProvider {
	return NewDomainMonitorTaskProviderRaw(db, logDb, logger)
}

// NewDomainMonitorTaskProvider 创建一个新的 域名监控 提供者
func NewDomainMonitorTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) *DomainMonitorTaskProvider {
	return &DomainMonitorTaskProvider{
		db:            db,
		logDb:         logDb,
		logger:        logger.With(slog.String("handler", "domain-monitor")),
		httpTransport: taskHttpTransport,

		businessType: cronjob.BusinessTypeDomain,
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

// Process 执行 域名监控 逻辑
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
	if meta.TaskType == cronjob.TaskTypeRecurring {
		if !domain.IsActive {
			return "跳过: 记录已禁用", nil
		}
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

// LoadTask 加载 域名监控 周期任务
func (p *DomainMonitorTaskProvider) LoadTask(ctx context.Context, db *ent.Client, record *ent.DomainMonitor, isActive bool) error {

	var (
		taskType    = cronjob.TaskTypeRecurring
		jobType     = cronjob.JobTypeDuration
		jobCategory = cronjob.JobCategorySystem
	)

	// 状态
	targetStatus := cronjob.StatusStopped
	if isActive {
		targetStatus = cronjob.StatusRunning
	}

	// 查找任务
	job, err := db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(taskType)).
		Where(cronjob.BusinessTypeEQ(p.businessType)).
		Where(cronjob.BusinessIDEQ(record.ID)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = db.CronJob.Create().
			SetName("域名监控: " + record.Domain).
			SetHandlerName(DomainMonitorTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(DomainMonitorTaskArgs{
				ID: record.ID,
			}))).
			SetTaskType(taskType).
			SetJobType(jobType).
			SetJobCategory(jobCategory).
			SetBusinessType(p.businessType).
			SetBusinessID(record.ID).
			SetIntervalValue(record.CheckInterval).
			SetIntervalUnit(cronjob.IntervalUnitSeconds).
			SetStatus(targetStatus).
			SetIsSingleton(true). // 单例模式
			SetRunOnStart(false). // 启动时不立即运行
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建任务失败: %w", err)
		}

		return nil
	}

	// 更新任务
	err = db.CronJob.UpdateOne(job).
		SetIntervalValue(record.CheckInterval).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(targetStatus).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	return nil
}

// AddTask 添加 域名监控 一次性任务
func (p *DomainMonitorTaskProvider) AddTask(ctx context.Context, db *ent.Client, record *ent.DomainMonitor) error {

	var (
		taskType    = cronjob.TaskTypeOneOff
		jobType     = cronjob.JobTypeDuration
		jobCategory = cronjob.JobCategoryCustom
	)

	// 创建任务
	err := db.CronJob.Create().
		SetName("域名监控: " + record.Domain).
		SetHandlerName(DomainMonitorTaskName).
		SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(DomainMonitorTaskArgs{
			ID: record.ID,
		}))).
		SetTaskType(taskType).
		SetJobType(jobType).
		SetJobCategory(jobCategory).
		SetBusinessType(p.businessType).
		SetBusinessID(record.ID).
		SetIntervalValue(1).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(cronjob.StatusRunning).
		SetIsSingleton(true). // 单例模式
		SetRunOnStart(true).  // 启动时立即运行
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}
