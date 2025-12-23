package task

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/helper/ssl_helper"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gostring"
	"golang.org/x/sync/errgroup"
)

// SSL更新任务
type SSLUpdateTask struct {
	db            *ent.Client                // 数据库客户端
	logDb         *ent_log.Client            // 日志数据库客户端
	cron          gocron.Scheduler           // 定时任务调度器
	sslJob        gocron.Job                 // SSL刷新任务
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
}

// 创建实例
func NewSSLUpdateTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) *SSLUpdateTask {
	return &SSLUpdateTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// Close 关闭服务，释放资源
func (s *SSLUpdateTask) Close() {
	if s.cron != nil && s.sslJob != nil {
		_ = s.cron.RemoveJob(s.sslJob.ID())
	}
}

// 启动 SSL 刷新 任务
func (s *SSLUpdateTask) StartSSLJob(ctx context.Context) (err error) {
	// 判断是否存在SSL任务
	if s.sslJob != nil {
		return nil
	}
	// 每10小时执行一次
	s.sslJob, err = s.cron.NewJob(
		gocron.DurationJob(10*time.Hour),                     // 每10小时执行一次
		gocron.NewTask(s.runSSLTask),                         // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		// gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("SSL记录刷新"),    // 任务名称
		gocron.WithTags("ssl_update"), // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSLUpdateTask) runSSLTask() {

	ctx := context.Background()
	processingType := "SSL"

	s.logger.InfoContext(ctx, "开始执行",
		slog.String("处理类型", processingType),
	)

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

		// 按策略分组
		for _, sslRecord := range allSSLRecords {

			// 创建待处理记录
			pendingRecord := &PendingSSLRecord{
				Item:     sslRecord,
				Platform: sslRecord.Edges.Config.Platform,
				Params: &ssl_helper.SSLProcessParams{
					SSLRecordID: cast.ToString(sslRecord.ID),

					ConfigJSON: sslRecord.Edges.Config.ConfigJSON,
				},
			}

			// 按策略分组
			if allowConcurrent(sslRecord.Edges.Config) {
				concurrentRecords = append(concurrentRecords, pendingRecord)
			} else {
				serialRecords = append(serialRecords, pendingRecord)
			}
		}

		// 1. 并发处理支持并发的记录
		if len(concurrentRecords) > 0 {

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(10) // 最多 10 个并发

			for _, item := range concurrentRecords {
				item := item // 创建副本以避免闭包问题
				eg.Go(func() error {

					newLog := s.logger.With(
						slog.String("处理方式", "并发"),
						slog.String("处理类型", processingType),
						slog.Int64("记录", item.Item.ID),
						slog.String("域名", item.Item.Domain),
						slog.String("平台", item.Platform.String()),
					)

					if err = s.SSLUpdate(egCtx, newLog, item); err != nil {
						newLog.ErrorContext(egCtx, "处理失败",
							slog.Any("处理参数", item.Params),
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				s.logger.ErrorContext(ctx, "并发处理错误",
					slog.String("处理类型", processingType),
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			newLog := s.logger.With(
				slog.String("处理方式", "串行"),
				slog.String("处理类型", processingType),
				slog.Int64("记录", item.Item.ID),
				slog.String("域名", item.Item.Domain),
				slog.String("平台", item.Platform.String()),
			)

			if err = s.SSLUpdate(ctx, newLog, item); err != nil {
				newLog.ErrorContext(ctx, "处理失败",
					slog.Any("处理参数", item.Params),
					slog.Any("error", err),
				)
			}

		}

	}

	nextRun, err := s.sslJob.NextRun()
	if err == nil {
		s.logger.InfoContext(ctx, "执行完成",
			slog.String("处理类型", processingType),
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)
	} else {
		s.logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.Int("串行记录数", len(serialRecords)),
			slog.String("处理类型", processingType),
		)
	}

}

// 待处理
type PendingSSLRecord struct {
	Item     *ent.SSLRecord               // SSL记录
	Platform platformconfig.Platform      // 平台
	Params   *ssl_helper.SSLProcessParams // 处理参数
}

// 更新SSL记录
// pendingRecord：待处理记录
func (s *SSLUpdateTask) SSLUpdate(ctx context.Context, logger *slog.Logger, pendingRecord *PendingSSLRecord) (err error) {

	// 初始化SSL处理
	var sslProcess ssl_helper.SSLProcess

	// 准备域名列表
	domains := gostring.Split(pendingRecord.Item.Domain, ",")

	// 准备保存
	updateOne := s.db.SSLRecord.UpdateOneID(pendingRecord.Item.ID)
	createLogOne := s.logDb.SSLCertificateHistory.Create()

	switch pendingRecord.Platform {
	case platformconfig.PlatformCloudflare:
		sslProcess = ssl_helper.NewSSLProcessCloudflare(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformTencentDNS:
		sslProcess = ssl_helper.NewSSLProcessTencentDNS(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformTencentEo:
		sslProcess = ssl_helper.NewSSLProcessTencentEO(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformAliyunDNS:
		sslProcess = ssl_helper.NewSSLProcessAliyunDNS(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformAliyunEsa:
		sslProcess = ssl_helper.NewSSLProcessAliyunEsa(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformBaiduDNS:
		sslProcess = ssl_helper.NewSSLProcessBaiduDNS(
			s.debug,
			logger,
			s.httpTransport,
		)
	case platformconfig.PlatformHuaweiDNS:
		sslProcess = ssl_helper.NewSSLProcessHuaweiDNS(
			s.debug,
			logger,
			s.httpTransport,
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
			slog.Any("处理参数", pendingRecord.Params),
			slog.Any("error", err),
		)
		return err
	}

	// 申请
	res, err := sslProcess.Apply(ctx, domains)
	if err != nil {
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)
		logger.ErrorContext(ctx, "申请失败",
			slog.Any("处理参数", pendingRecord.Params),
			slog.Any("error", err),
		)
		return err
	}

	// 解析证书有效期
	certValidity, err := s.ParseCertificateValidity(res)
	if err != nil {
		updateOne.SetLastAPIResponse(err.Error())
		updateOne.SetLastAPIResponseAt(time.Now())
		_ = updateOne.Exec(ctx)
		logger.ErrorContext(ctx, "解析证书有效期失败",
			slog.Any("处理参数", pendingRecord.Params),
			slog.Any("error", err),
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

// ParseCertificateValidity 从 lego Resource 解析证书有效期
func (s *SSLUpdateTask) ParseCertificateValidity(res *certificate.Resource) (*CertificateValidity, error) {
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
