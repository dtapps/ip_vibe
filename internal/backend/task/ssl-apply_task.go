package task

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/ssl_helper"
	"vibe_ip/internal/helper/sys_env"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gostring"
	"golang.org/x/sync/errgroup"
)

// SSL申请任务
type SSLApplyTask struct {
	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	cron          gocron.Scheduler              // 定时任务调度器
	job           gocron.Job                    // SSL申请任务
	taskName      string                        // 任务名称
	jobHelper     *gocron_helper.JobHelper      // 任务助手
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 是否开启调试模式
}

// 创建实例
func NewSSLApplyTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *SSLApplyTask {
	debug := false
	if sys_env.GvaEnv != nil {
		debug = sys_env.GvaEnv.DEBUG
	}
	return &SSLApplyTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		taskName:      "SSL申请",
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// SetDebug 设置调试模式
func (s *SSLApplyTask) SetDebug(debug bool) *SSLApplyTask {
	s.debug = debug
	return s
}

// Close 关闭服务，释放资源
func (s *SSLApplyTask) Close() {
	if s.cron != nil && s.job != nil && s.jobHelper != nil {
		_ = s.jobHelper.DeleteJob(s.job.ID().String())
	}
}

// 自定义 logger
func (s *SSLApplyTask) SetLogger(logger *slog.Logger) *SSLApplyTask {
	s.logger = logger
	return s
}

// 计算下次执行时间
func (s *SSLApplyTask) CalculateNextRun(sslRecord *ent.SSLRecord, now time.Time) time.Time {
	// 如果是首次申请（没有过期时间），立即执行
	if sslRecord.NotAfter.IsZero() {
		return now
	}

	// 计算理想的续期时间：过期前 RenewalDays 天
	idealRenewalTime := sslRecord.NotAfter.Add(-time.Duration(sslRecord.RenewalDays) * 24 * time.Hour)

	// 如果理想续期时间已经过去，说明需要立即续期
	if idealRenewalTime.Before(now) {
		return now
	}

	// 添加一些随机性，避免所有证书同时续期
	// 在理想时间前后随机偏移最多6小时
	randomOffset := time.Duration(rand.Intn(7200)-3600) * time.Second // -1小时到+1小时
	nextRun := idealRenewalTime.Add(randomOffset)

	// 确保不会设置到过去
	if nextRun.Before(now) {
		nextRun = now
	}

	return nextRun
}

// StartPeriodicTask 启动周期任务
func (s *SSLApplyTask) StartPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.job != nil {
		return nil
	}
	// 解析设置的时间
	jobDuration, isStart, err := gocron_helper.ParseTaskTime(
		db_cache.GvaSetting.SSLRefreshTaskTime,
		time.Hour,
	)
	if !isStart {
		return nil
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "解析设置的时间",
			slog.Any("err", err),
		)
		return err
	}
	// 创建周期任务
	s.job, err = s.cron.NewJob(
		jobDuration,
		gocron.NewTask(s.RunPeriodicTask),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithName(s.taskName),                          // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		), // 事件监听
	)
	if err != nil {
		return fmt.Errorf("创建定时任务失败: %w", err)
	}

	s.logger.InfoContext(ctx, s.taskName+"任务已启动",
		slog.String("执行间隔", fmt.Sprintf("%s小时", db_cache.GvaSetting.SSLRefreshTaskTime)),
	)
	return nil
}

// RunPeriodicTask 运行任务
func (s *SSLApplyTask) RunPeriodicTask() {

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceTask.String())

	// 日志
	logger := s.jobHelper.SlogWith(s.logger, s.job.ID().String(), s.job.Name())

	// 分组记录
	var concurrentRecords, serialRecords []*PendingSSLRecord

	// 查询全部SSL记录
	allSSLRecords, err := s.db.SSLRecord.Query().
		Where(sslrecord.IsActiveEQ(true)).
		Where(sslrecord.IsAutoRenewalEQ(true)).
		Where(sslrecord.ProviderNEQ(sslrecord.ProviderUnknown)).
		// 预加载 配置
		WithConfig().
		All(ctx)
	if err == nil {

		now := time.Now()

		// 按策略分组
		for _, sslRecord := range allSSLRecords {
			// 计算下次执行时间
			nextRun := s.CalculateNextRun(sslRecord, now)

			// 如果已超过或等于下次运行时间，则执行任务
			if now.After(nextRun) || now.Equal(nextRun) {
				// 创建待处理记录
				pendingRecord := &PendingSSLRecord{
					Item:     sslRecord,
					Platform: sslRecord.Edges.Config.Platform,
					Params: &ssl_helper.SSLProcessParams{
						SSLRecordID: cast.ToString(sslRecord.ID),

						ConfigJSON: sslRecord.Edges.Config.ConfigJSON,
						Domains:    gostring.Split(sslRecord.Domain, ","),
					},
				}

				// 按策略分组
				if allowConcurrent(sslRecord.Edges.Config) {
					concurrentRecords = append(concurrentRecords, pendingRecord)
				} else {
					serialRecords = append(serialRecords, pendingRecord)
				}
			}
		}

		logger.InfoContext(ctx, "准备处理",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					if err = s.executeSSLApply(egCtx, logger.With(
						slog.String("处理方式", "并发"),
					), item); err != nil {
						logger.ErrorContext(egCtx, "处理失败",
							slog.Any("错误信息", err),
						)
					}

					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				logger.ErrorContext(ctx, "并发处理错误",
					slog.Any("错误信息", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			if err = s.executeSSLApply(ctx, logger.With(
				slog.String("处理方式", "串行"),
			), item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("错误信息", err),
				)
			}

		}

	}

}

// 待处理
type PendingSSLRecord struct {
	Item     *ent.SSLRecord               // SSL记录
	Platform platformconfig.Platform      // 平台
	Params   *ssl_helper.SSLProcessParams // 处理参数
}

// 执行SSL申请
func (s *SSLApplyTask) executeSSLApply(ctx context.Context, logger *slog.Logger, pendingRecord *PendingSSLRecord) (err error) {

	// 日志
	logger = logger.With(
		slog.String("记录", pendingRecord.Params.SSLRecordID),
		slog.String("提供商", pendingRecord.Item.Provider.String()),
		slog.String("平台", pendingRecord.Platform.String()),
		slog.Any("域名列表", pendingRecord.Params.Domains),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// 初始化SSL处理
	var sslProcess ssl_helper.SSLProcess

	// 准备保存
	updateOne := s.db.SSLRecord.UpdateOneID(pendingRecord.Item.ID)
	createLogOne := s.logDb.SSLCertificateHistory.Create()

	switch pendingRecord.Platform {
	case platformconfig.PlatformCloudflare:
		sslProcess = ssl_helper.NewSSLProcessCloudflare(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformTencentDNS:
		sslProcess = ssl_helper.NewSSLProcessTencentDNS(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformTencentEo:
		sslProcess = ssl_helper.NewSSLProcessTencentEO(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformAliyunDNS:
		sslProcess = ssl_helper.NewSSLProcessAliyunDNS(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformAliyunEsa:
		sslProcess = ssl_helper.NewSSLProcessAliyunEsa(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformBaiduDNS:
		sslProcess = ssl_helper.NewSSLProcessBaiduDNS(
			logger,
			s.httpTransport,
			s.debug,
		)
	case platformconfig.PlatformHuaweiDNS:
		sslProcess = ssl_helper.NewSSLProcessHuaweiDNS(
			logger,
			s.httpTransport,
			s.debug,
		)
	default:
		err = fmt.Errorf("暂不支持该平台: %s", pendingRecord.Platform.String())
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)
		return err
	}

	// 配置
	err = sslProcess.Configure(ctx, pendingRecord.Params)
	if err != nil {
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)

		logger.ErrorContext(ctx, "配置失败",
			slog.Any("错误信息", err),
		)
		return err
	}

	// 申请
	res, err := sslProcess.Apply(ctx, pendingRecord.Params.Domains)
	if err != nil {
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)

		logger.ErrorContext(ctx, "申请失败",
			slog.Any("错误信息", err),
		)
		return err
	}

	// 解析证书有效期
	certValidity, err := s.parseCertificateValidity(res)
	if err != nil {
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)

		logger.ErrorContext(ctx, "解析证书有效期失败",
			slog.Any("错误信息", err),
		)
		return err
	}

	// 更新SSL记录
	updateOne.SetCertDomain(res.Domain)
	updateOne.SetCertURL(res.CertURL)
	updateOne.SetCertStableURL(res.CertStableURL)
	updateOne.SetCertPrivateKey(string(res.PrivateKey))
	updateOne.SetCertCertificate(string(res.Certificate))
	updateOne.SetCertIssuerCertificate(string(res.IssuerCertificate))
	updateOne.SetCertCsr(string(res.CSR))
	updateOne.SetNotBefore(certValidity.NotBefore)
	updateOne.SetNotAfter(certValidity.NotAfter)
	updateOne.SetStatus(sslrecord.StatusValid)
	updateOne.SetLastRefreshAt(time.Now())
	updateOne.SetLastAPIResponse("ok")
	updateOne.SetLastAPIResponseAt(time.Now())
	_ = updateOne.Exec(ctx)

	// 保存到SSL证书申请历史记录
	createLogOne.SetSslRecordID(pendingRecord.Item.ID)
	createLogOne.SetCertDomain(res.Domain)
	createLogOne.SetCertURL(res.CertURL)
	createLogOne.SetCertStableURL(res.CertStableURL)
	createLogOne.SetCertPrivateKey(string(res.PrivateKey))
	createLogOne.SetCertCertificate(string(res.Certificate))
	createLogOne.SetCertIssuerCertificate(string(res.IssuerCertificate))
	createLogOne.SetCertCsr(string(res.CSR))
	createLogOne.SetNotBefore(certValidity.NotBefore)
	createLogOne.SetNotAfter(certValidity.NotAfter)
	createLogOne.SetProvider(sslcertificatehistory.Provider(pendingRecord.Item.Provider))
	_ = createLogOne.Exec(ctx)

	return nil
}

// CertificateValidity 证书有效期信息
type CertificateValidity struct {
	NotBefore time.Time
	NotAfter  time.Time
}

// 从 lego Resource 解析证书有效期
func (s *SSLApplyTask) parseCertificateValidity(res *certificate.Resource) (*CertificateValidity, error) {
	if res == nil || len(res.Certificate) == 0 {
		return nil, errors.New("certificate data is empty")
	}

	// 解析 PEM 数据
	block, _ := pem.Decode(res.Certificate)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing certificate")
	}

	// 解析 X.509 证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return &CertificateValidity{
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
	}, nil
}

// AddDisposableTask 添加一次性任务
func (s *SSLApplyTask) AddDisposableTask(ctx context.Context, pendingRecord *PendingSSLRecord) (err error) {

	// 任务编号
	jobID := uuid.New()

	// 任务名称
	jobName := fmt.Sprintf("%s-%v", s.taskName, pendingRecord.Params.SSLRecordID)

	// 检查是否已存在同名任务
	isExist := s.jobHelper.IsJob("", jobName)
	if isExist {
		return fmt.Errorf("任务 %s 已存在，无法重复添加", jobName)
	}

	// 不存在同名任务，添加一次任务
	_, err = s.cron.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(5*time.Second)), // 5秒后开始
		),
		gocron.NewTask(
			s.RunDisposableTask,
			ctx,
			s.jobHelper.SlogWith(s.logger, jobID.String(), jobName),
			pendingRecord,
		),
		gocron.WithSingletonMode(gocron.LimitModeWait), // 单例模式
		gocron.WithName(jobName),                       // 任务名称
		gocron.WithIdentifier(jobID),                   // 任务编号
		gocron.WithEventListeners(
			s.jobHelper.PanicDisposableListener(),
			s.jobHelper.BeforeDisposableListener(),
			s.jobHelper.AfterDisposableListener(),
		), // 事件监听
	)

	return err
}

// RunDisposableTask 运行一次性任务
func (s *SSLApplyTask) RunDisposableTask(ctx context.Context, logger *slog.Logger, taskPendingRecord *PendingSSLRecord) {
	err := s.executeSSLApply(ctx, logger, taskPendingRecord)
	if err != nil {
		logger.ErrorContext(ctx, "处理失败",
			slog.Any("错误信息", err),
		)
	}
}
