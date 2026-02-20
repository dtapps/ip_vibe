package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/ent_log/httplog"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/ent_log/portforwardstats"
	"vibe_ip/internal/ent_log/reqlog"
	"vibe_ip/internal/ent_log/restylog"
	"vibe_ip/internal/ent_log/webhookhistory"
	"vibe_ip/internal/provider"

	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gotime"
)

// ClearDataTaskProvider 实现 provider.TaskProvider 接口
type ClearDataTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	logger *slog.Logger    // 日志记录器

	businessType cronjob.BusinessType // 业务类型
}

// NewClearDataTaskProvider 创建一个新的 数据清理 提供者
func NewClearDataTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) provider.TaskProvider {
	return NewClearDataTaskProviderRaw(db, logDb, logger)
}

// NewClearDataTaskProvider 创建一个新的 数据清理 提供者
func NewClearDataTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger) *ClearDataTaskProvider {
	return &ClearDataTaskProvider{
		db:     db,
		logDb:  logDb,
		logger: logger.With(slog.String("handler", "clear-data")),

		businessType: cronjob.BusinessTypeSystem,
	}
}

// ClearDataTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	ClearDataTaskName = "clear_data"
)

// ClearDataTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type ClearDataTaskArgs struct {
	Table string `json:"table"` // 表名
	Day   int    `json:"day"`   // 保留天数
}

// Process 执行 数据清理 新逻辑
func (p *ClearDataTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params ClearDataTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	var (
		err          error
		beforeCount  int
		cleanedCount int
		afterCount   int
	)

	switch params.Table {
	case ent_log.TypeHttpLog:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.HttpLog.Query().
			Where(httplog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.HttpLog.Delete().
			Where(httplog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.HttpLog.Query().
			Where(httplog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeRestyLog:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.RestyLog.Query().
			Where(restylog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.RestyLog.Delete().
			Where(restylog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.RestyLog.Query().
			Where(restylog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeReqLog:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.ReqLog.Query().
			Where(reqlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.ReqLog.Delete().
			Where(reqlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.ReqLog.Query().
			Where(reqlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeDomainMonitorHistory:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.DomainMonitorHistory.Query().
			Where(domainmonitorhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.DomainMonitorHistory.Delete().
			Where(domainmonitorhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.DomainMonitorHistory.Query().
			Where(domainmonitorhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeOperationLog:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.OperationLog.Query().
			Where(operationlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.OperationLog.Delete().
			Where(operationlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.OperationLog.Query().
			Where(operationlog.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeWebhookHistory:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.WebhookHistory.Query().
			Where(webhookhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.WebhookHistory.Delete().
			Where(webhookhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.WebhookHistory.Query().
			Where(webhookhistory.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypeDeviceIpReport:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.DeviceIpReport.Query().
			Where(deviceipreport.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.DeviceIpReport.Delete().
			Where(deviceipreport.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.DeviceIpReport.Query().
			Where(deviceipreport.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	case ent_log.TypePortForwardStats:
		// 查询删除前的记录数
		beforeCount, err = p.logDb.PortForwardStats.Query().
			Where(portforwardstats.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除前的记录数失败: %w", err)
		}
		// 执行删除操作
		cleanedCount, err = p.logDb.PortForwardStats.Delete().
			Where(portforwardstats.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Exec(ctx)
		if err != nil {
			return "", fmt.Errorf("执行删除操作失败: %w", err)
		}
		// 查询删除后的记录数
		afterCount, err = p.logDb.PortForwardStats.Query().
			Where(portforwardstats.CreatedAtLT(gotime.Current().BeforeDay(params.Day).Time)).
			Count(ctx)
		if err != nil {
			return "", fmt.Errorf("查询删除后的记录数失败: %w", err)
		}
	default:
		return "", fmt.Errorf("不支持的表名 %s", params.Table)
	}

	return fmt.Sprintf("表名 %s 清理成功，保留 %d 天，清理前 %d 条数据，清理后 %d 条数据，清理 %d 条数据", params.Table, params.Day, beforeCount, afterCount, cleanedCount), nil
}

// LoadTask 加载 数据清理 周期任务
func (p *ClearDataTaskProvider) LoadTask(ctx context.Context, db *ent.Client, table string) error {

	var day = 7
	switch table {
	case ent_log.TypeHttpLog:
		day = 90
	case ent_log.TypeRestyLog:
		day = 90
	case ent_log.TypeReqLog:
		day = 90
	case ent_log.TypeDomainMonitorHistory:
		day = 365
	case ent_log.TypeOperationLog:
		day = 365
	case ent_log.TypeWebhookHistory:
		day = 365
	case ent_log.TypeDeviceIpReport:
		day = 365
	case ent_log.TypePortForwardStats:
		day = 365
	default:
		return fmt.Errorf("不支持的表名 %s", table)
	}

	// 查找任务
	_, err := db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(cronjob.TaskTypeRecurring)).
		Where(cronjob.BusinessTypeEQ(p.businessType)).
		Where(cronjob.BusinessKeyEQ(table)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = db.CronJob.Create().
			SetName("数据清理: " + table).
			SetHandlerName(ClearDataTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(ClearDataTaskArgs{
				Table: table,
				Day:   day,
			}))).
			SetTaskType(cronjob.TaskTypeRecurring).
			SetJobType(cronjob.JobTypeCron).
			SetJobCategory(cronjob.JobCategorySystem).
			SetBusinessType(p.businessType).
			SetBusinessKey(table).
			SetCronExpr("5 0 * * *").
			SetStatus(cronjob.StatusRunning).
			SetIsSingleton(true). // 单例模式
			SetRunOnStart(false). // 启动时不立即运行
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("创建任务失败: %w", err)
		}

		return nil
	}

	return nil
}
