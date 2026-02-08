package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/sslimport"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gojson"
	"golang.org/x/sync/errgroup"
)

// SSLRefreshTaskProvider 实现 provider.TaskProvider 接口
type SSLRefreshTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
}

// NewSSLRefreshTaskProvider 创建一个新的 SSL刷新 提供者
func NewSSLRefreshTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewSSLRefreshTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewSSLRefreshTaskProvider 创建一个新的 SSL刷新 提供者
func NewSSLRefreshTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *SSLRefreshTaskProvider {
	return &SSLRefreshTaskProvider{
		db:            db,
		logDb:         logDb,
		logger:        logger.With(slog.String("handler", "ssl-refresh")),
		httpTransport: httpTransport,
	}
}

// SSLRefreshTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	SSLRefreshTaskName = "ssl_refresh"
)

// SSLRefreshTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type SSLRefreshTaskArgs struct {
	Table string `json:"table"` // 表名
}

// Process 执行 SSL刷新 逻辑
func (p *SSLRefreshTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params SSLRefreshTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	allRecordCount := 0

	switch params.Table {
	case ent.TypeSSLRecord:

		// 查询所有未处理的SSL记录
		allRecord, err := p.db.SSLRecord.Query().
			Where(sslrecord.ProviderEQ(sslrecord.ProviderUnknown)).
			// 排序
			OrderByCreatedAtAsc().
			All(ctx)
		if err != nil {
			return "", fmt.Errorf("获取未处理的SSL记录失败: %w", err)
		}

		allRecordCount = len(allRecord)

		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(10) // 最多 10 个并发

		for _, record := range allRecord {
			// 创建副本以避免闭包问题
			record := record
			eg.Go(func() error {

				// 解析证书
				parsedCert, err := utils.ParseCertificate([]byte(record.CertCertificate))
				if err != nil {
					return fmt.Errorf("解析证书失败: %w", err)
				}

				// 证书状态
				status := sslrecord.StatusUnknown
				if parsedCert.NotAfter.After(time.Now()) {
					status = sslrecord.StatusValid
				} else {
					status = sslrecord.StatusExpired
				}

				// 准备保存
				updateOne := p.db.SSLRecord.UpdateOneID(record.ID)

				updateOne.SetNotBefore(parsedCert.NotBefore)
				updateOne.SetNotAfter(parsedCert.NotAfter)
				updateOne.SetStatus(status)

				err = updateOne.Exec(egCtx)
				if err != nil {
					return fmt.Errorf("更新记录失败: %w", err)
				}

				return nil
			})
		}

		err = eg.Wait()
		if err != nil {
			return "", fmt.Errorf("并发处理错误: %w", err)
		}

	case ent.TypeSSLImport:

		// 查询所有未处理的SSL导入记录
		allRecord, err := p.db.SSLImport.Query().
			// 排序
			OrderByCreatedAtAsc().
			All(ctx)
		if err != nil {
			return "", fmt.Errorf("获取未处理的SSL导入记录失败: %w", err)
		}

		allRecordCount = len(allRecord)

		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(10) // 最多 10 个并发

		for _, record := range allRecord {
			// 创建副本以避免闭包问题
			record := record
			eg.Go(func() error {

				// 解析证书
				parsedCert, err := utils.ParseCertificate([]byte(record.Certificate))
				if err != nil {
					return fmt.Errorf("解析证书失败: %w", err)
				}

				// 证书状态
				status := sslimport.StatusUnknown
				if parsedCert.NotAfter.After(time.Now()) {
					status = sslimport.StatusValid
				} else {
					status = sslimport.StatusExpired
				}

				// 准备保存
				updateOne := p.db.SSLImport.UpdateOneID(record.ID)

				updateOne.SetNotBefore(parsedCert.NotBefore)
				updateOne.SetNotAfter(parsedCert.NotAfter)
				updateOne.SetStatus(status)

				err = updateOne.Exec(egCtx)
				if err != nil {
					return fmt.Errorf("更新记录失败: %w", err)
				}

				return nil
			})
		}

		err = eg.Wait()
		if err != nil {
			return "", fmt.Errorf("并发处理错误: %w", err)
		}
	default:
		return "", fmt.Errorf("不支持的表名 %s", params.Table)
	}

	return fmt.Sprintf("表名 %s 处理成功 %d 条记录", params.Table, allRecordCount), nil
}

// LoadTask 加载 SSL刷新 任务
func (p *SSLRefreshTaskProvider) LoadTask(ctx context.Context, table string) error {

	switch table {
	case ent.TypeSSLRecord:
	case ent.TypeSSLImport:
	default:
		return fmt.Errorf("不支持的表名 %s", table)
	}

	// 查找任务
	_, err := p.db.CronJob.Query().
		Where(cronjob.TaskTypeEQ(cronjob.TaskTypeRecurring)).
		Where(cronjob.BusinessTypeEQ(cronjob.BusinessTypeSsl)).
		Where(cronjob.BusinessKeyEQ(table)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("查找任务失败: %w", err)
		}

		// 创建任务
		err = p.db.CronJob.Create().
			SetName("SSL刷新: " + table).
			SetHandlerName(SSLRefreshTaskName).
			SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(SSLRefreshTaskArgs{
				Table: table,
			}))).
			SetTaskType(cronjob.TaskTypeRecurring).
			SetJobType(cronjob.JobTypeCron).
			SetJobCategory(cronjob.JobCategorySystem).
			SetBusinessType(cronjob.BusinessTypeSsl).
			SetBusinessKey(table).
			SetCronExpr("0 0 * * *").
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
