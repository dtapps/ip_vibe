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
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/ssl_helper"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gorandom"
	"go.dtapp.net/library/utils/gostring"
	"golang.org/x/sync/errgroup"
)

// SSL更新任务
type SSLUpdateTask struct {
	db            *ent.Client                // 数据库客户端
	logDb         *ent_log.Client            // 日志数据库客户端
	cron          gocron.Scheduler           // 定时任务调度器
	sslJob        gocron.Job                 // SSL刷新任务
	jobHelper     *gocron_helper.JobHelper   // 任务助手
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
	taskName      string                     // 任务名称
}

// 创建实例
func NewSSLUpdateTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, httpTransport *http_log.LoggingTransport, debug bool) *SSLUpdateTask {
	logger := slog.Default()
	if global.GvaGocronSlog != nil {
		logger = global.GvaGocronSlog.WithLogger()
	}
	return &SSLUpdateTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
		taskName:      "SSL更新",
	}
}

// Close 关闭服务，释放资源
func (s *SSLUpdateTask) Close() {
	if s.cron != nil && s.sslJob != nil {
		_ = s.cron.RemoveJob(s.sslJob.ID())
	}
}

// 自定义 logger
func (s *SSLUpdateTask) SetLogger(logger *slog.Logger) *SSLUpdateTask {
	s.logger = logger
	return s
}

// 计算下次执行时间
func (s *SSLUpdateTask) CalculateNextRun(sslRecord *ent.SSLRecord, now time.Time) time.Time {
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

// 启动SSL周期任务
func (s *SSLUpdateTask) StartSSLPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.sslJob != nil {
		return nil
	}
	// 每10小时执行一次
	s.sslJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Hour),
		gocron.NewTask(s.RunSSLPeriodicTask),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		// gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName(s.taskName), // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicPeriodicListener(),
			s.jobHelper.BeforePeriodicListener(),
			s.jobHelper.AfterPeriodicListener(),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// 运行SSL周期任务
func (s *SSLUpdateTask) RunSSLPeriodicTask() {

	ctx := context.Background()

	logger := s.jobHelper.SlogWith(s.logger, s.sslJob.ID().String(), s.sslJob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*PendingSSLRecord

	// 查询全部SSL记录
	allSSLRecords, err := s.db.SSLRecord.Query().
		Where(sslrecord.IsActiveEQ(true)).
		Where(sslrecord.IsAutoRenewalEQ(true)).
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

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					logger = logger.With(
						slog.String("处理方式", "并发"),
					)

					if err = s.executeSSLUpdate(egCtx, logger, item); err != nil {
						logger.ErrorContext(egCtx, "处理失败",
							slog.Any("错误信息", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
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

			logger = logger.With(
				slog.String("处理方式", "串行"),
			)

			if err = s.executeSSLUpdate(ctx, logger, item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("错误信息", err),
				)
			}

		}

	}

	nextRun, err := s.sslJob.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
		)
	}
}

// 待处理
type PendingSSLRecord struct {
	Item     *ent.SSLRecord               // SSL记录
	Platform platformconfig.Platform      // 平台
	Params   *ssl_helper.SSLProcessParams // 处理参数
}

// 执行SSL更新
func (s *SSLUpdateTask) executeSSLUpdate(ctx context.Context, logger *slog.Logger, pendingRecord *PendingSSLRecord) (err error) {

	// 日志
	logger = logger.With(
		slog.String("处理编号", gorandom.Alphanumeric(10)),
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
func (s *SSLUpdateTask) parseCertificateValidity(res *certificate.Resource) (*CertificateValidity, error) {
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

// 添加SSL一次性任务
func (s *SSLUpdateTask) AddSSLDisposableTask(ctx context.Context, pendingRecord *PendingSSLRecord) (err error) {

	// 任务名称
	taskName := fmt.Sprintf("%s-%v", s.taskName, pendingRecord.Params.SSLRecordID)

	// 检查是否已存在同名任务
	isExist := s.jobHelper.IsJob("", taskName)
	if isExist {
		return fmt.Errorf("任务 %s 已存在，无法重复添加", taskName)
	}

	// 不存在同名任务，添加新任务
	_, err = s.cron.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(5*time.Second)), // 5秒后开始
		),
		gocron.NewTask(
			s.RunSSLDisposableTask,
			ctx,
			s.logger,
			pendingRecord,
		),
		gocron.WithSingletonMode(gocron.LimitModeWait), // 单例模式
		gocron.WithName(taskName),                      // 任务名称
		gocron.WithEventListeners(
			s.jobHelper.PanicDisposableListener(),
			s.jobHelper.BeforeDisposableListener(),
			s.jobHelper.AfterDisposableListener(),
		),
	)
	return err
}

// 运行SSL一次性任务
func (s *SSLUpdateTask) RunSSLDisposableTask(ctx context.Context, logger *slog.Logger, taskPendingRecord *PendingSSLRecord) {
	err := s.executeSSLUpdate(ctx, logger, taskPendingRecord)
	if err != nil {
		logger.ErrorContext(ctx, "处理失败",
			slog.Any("错误信息", err),
		)
	}
}
