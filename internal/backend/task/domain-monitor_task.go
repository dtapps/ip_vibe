package task

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gorandom"
	"golang.org/x/sync/errgroup"
)

// 域名监控任务
type DomainMonitorTask struct {
	db            *ent.Client                // 业务数据库客户端
	logDb         *ent_log.Client            // 日志数据库客户端
	cron          gocron.Scheduler           // 定时任务调度器
	dmjob         gocron.Job                 // 域名监控任务
	jobHelper     *gocron_helper.JobHelper   // 任务助手
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
	taskName      string                     // 任务名称
}

// 创建实例
func NewDomainMonitorTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, httpTransport *http_log.LoggingTransport, debug bool) *DomainMonitorTask {
	logger := slog.Default()
	if global.GvaGocronSlog != nil {
		logger = global.GvaGocronSlog.WithLogger()
	}
	return &DomainMonitorTask{
		db:            db,
		logDb:         logDb,
		cron:          cron,
		jobHelper:     gocron_helper.NewJobHelper(cron, logger),
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
		taskName:      "域名监控",
	}
}

// Close 关闭服务，释放资源
func (s *DomainMonitorTask) Close() {
	if s.cron != nil && s.dmjob != nil {
		_ = s.cron.RemoveJob(s.dmjob.ID())
	}
}

// 自定义 logger
func (s *DomainMonitorTask) SetLogger(logger *slog.Logger) *DomainMonitorTask {
	s.logger = logger
	return s
}

// 计算下次执行时间
func (m *DomainMonitorTask) CalculateNextRun(domainMonitor *ent.DomainMonitor, now time.Time) time.Time {
	var nextRun time.Time

	// 基于成功时间计算
	if domainMonitor.LastSuccessAt.IsZero() {
		// 如果从未成功过，立即运行
		nextRun = now
	} else {
		// 下次运行时间 = 上次成功时间 + 检查间隔
		nextRun = domainMonitor.LastSuccessAt.Add(time.Duration(domainMonitor.CheckInterval) * time.Second)
	}

	// 如果是失败状态就使用上次失败时间（覆盖上面的计算）
	if domainMonitor.CurrentStatus == domainmonitor.CurrentStatusFailure {
		if !domainMonitor.LastFailureAt.IsZero() {
			nextRun = domainMonitor.LastFailureAt.Add(time.Duration(domainMonitor.CheckInterval) * time.Second)
		}
	}

	return nextRun
}

// 启动域名监控周期任务
func (s *DomainMonitorTask) StartDomainMonitorPeriodicTask(ctx context.Context) (err error) {
	// 判断是否存在任务
	if s.dmjob != nil {
		return nil
	}
	// 每30秒到60秒执行一次
	s.dmjob, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			30*time.Second, // 最小间隔
			60*time.Second, // 最大间隔
		),
		gocron.NewTask(s.RunDomainMonitorPeriodicTask),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName(s.taskName),                          // 任务名称
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

// 运行域名监控周期任务
func (s *DomainMonitorTask) RunDomainMonitorPeriodicTask() {

	ctx := context.Background()

	logger := s.jobHelper.SlogWith(s.logger, s.dmjob.ID().String(), s.dmjob.Name())
	logger.InfoContext(ctx, "开始执行")

	// 分组记录
	var concurrentRecords, serialRecords []*ent.DomainMonitor

	// 查询全部域名监控
	allDomainMonitors, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IsActive(true)).
		All(ctx)
	if err == nil {
		now := time.Now()
		for _, domainMonitor := range allDomainMonitors {
			// 计算下次执行时间
			nextRun := s.CalculateNextRun(domainMonitor, now)

			// 如果已超过或等于下次运行时间，则执行任务
			if now.After(nextRun) || now.Equal(nextRun) {
				// 按策略分组
				concurrentRecords = append(concurrentRecords, domainMonitor)
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

					if err = s.executeDomainMonitorCheck(egCtx, logger, item); err != nil {
						logger.ErrorContext(egCtx, "处理失败",
							slog.Any("error", err),
						)
					}

					// 返回nil而不是错误，这样单个DNS记录更新失败不会导致整个errgroup取消上下文
					return nil
				})
			}

			if err = eg.Wait(); err != nil {
				logger.ErrorContext(ctx, "并发处理错误",
					slog.Any("error", err),
				)
			}

		}
		// 2. 串行处理不支持并发的记录
		for _, item := range serialRecords {

			logger = logger.With(
				slog.String("处理方式", "串行"),
			)

			if err = s.executeDomainMonitorCheck(ctx, logger, item); err != nil {
				logger.ErrorContext(ctx, "处理失败",
					slog.Any("error", err),
				)
			}

		}
	}

	nextRun, err := s.dmjob.NextRun()
	if err == nil {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
			slog.String("下次运行时间", nextRun.Format(time.DateTime)),
		)
	} else {
		logger.InfoContext(ctx, "执行完成",
			slog.Int("并发记录数", len(concurrentRecords)),
		)
	}
}

// 执行域名监控检查
func (s *DomainMonitorTask) executeDomainMonitorCheck(ctx context.Context, logger *slog.Logger, domainMonitor *ent.DomainMonitor) (err error) {

	// 日志
	logger = logger.With(
		slog.String("处理编号", gorandom.Alphanumeric(10)),
		slog.Int64("记录", domainMonitor.ID),
		slog.String("域名", domainMonitor.Domain),
		slog.String("协议", domainMonitor.CheckProtocol.String()),
	)
	logger.InfoContext(ctx, "开始处理")
	defer logger.InfoContext(ctx, "处理完成")

	// 执行检查
	result := s.runDomainCheck(ctx, domainMonitor.Domain, domainMonitor.CheckProtocol.String())

	// 历史记录创建
	historyCreator := s.logDb.DomainMonitorHistory.Create().
		SetMonitorID(domainMonitor.ID).
		SetCheckTime(time.Now()).
		SetIsSuccess(result.IsSuccess).
		SetCheckDuration(int(result.Duration.Milliseconds())).
		SetHTTPStatus(result.HTTPStatus).
		SetStatusMessage(result.StatusMsg).
		SetErrorLog(result.ErrorLog)

	// 填充 SSL 字段
	if domainMonitor.CheckProtocol == domainmonitor.CheckProtocolHTTPS && result.SSLExpiryDate.After(time.Time{}) {
		historyCreator.SetSslExpiryDate(result.SSLExpiryDate).
			SetSslCommonName(result.SSLCommonName).
			SetSslIssuerName(result.SSLIssuerName).
			SetSslSignatureAlg(result.SSLSignatureAlg).
			SetSslSanList(result.SANList).
			SetSslCertificateChain(result.CertificateChain)
	}

	_, err = historyCreator.Save(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "保存历史记录失败",
			slog.Any("err", err),
		)
		return err
	}

	// 更新状态
	updaterOne := s.db.DomainMonitor.UpdateOne(domainMonitor)
	if result.IsSuccess {
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusSuccess)
	} else {
		updaterOne.SetCurrentStatus(domainmonitor.CurrentStatusFailure)
	}

	if result.IsSuccess {
		updaterOne.SetLastSuccessAt(time.Now())
	} else {
		updaterOne.SetLastFailureAt(time.Now())
	}

	err = updaterOne.Exec(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "更新状态失败",
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// 添加域名监控一次性任务
func (s *DomainMonitorTask) AddDomainMonitorDisposableTask(ctx context.Context, domainMonitor *ent.DomainMonitor) (err error) {

	// 任务名称
	taskName := fmt.Sprintf("%s-%v", s.taskName, domainMonitor.ID)

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
			s.RunDomainMonitorDisposableTask,
			ctx,
			s.logger,
			domainMonitor,
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

// RunDomainMonitorDisposableTask 运行域名监控一次性任务
func (s *DomainMonitorTask) RunDomainMonitorDisposableTask(ctx context.Context, logger *slog.Logger, domainMonitor *ent.DomainMonitor) {
	// 日志
	logger = logger.With(
		slog.String("处理方式", "手动"),
	)
	// 处理
	err := s.executeDomainMonitorCheck(ctx, logger, domainMonitor)
	if err != nil {
		logger.ErrorContext(ctx, "处理失败",
			slog.Any("error", err),
		)
	}
}

// CheckResult 包含本次检查的所有结果
type CheckResult struct {
	IsSuccess  bool          // 是否成功
	Duration   time.Duration // 检查耗时
	StatusMsg  string        // 状态信息
	HTTPStatus int           // HTTP状态码
	ErrorLog   string        // 错误日志

	// SSL 证书信息 (HTTP 检查时为空)
	SSLExpiryDate    time.Time // 证书过期日期
	SSLCommonName    string    // 证书通用名
	SSLIssuerName    string    // 证书颁发者
	SSLSignatureAlg  string    // 签名算法
	SANList          string    // SAN别名列表
	CertificateChain string    // 证书链内容 (PEM格式)
}

// RunDomainCheck 执行实际的网络检查并返回结果
func (s *DomainMonitorTask) runDomainCheck(ctx context.Context, domain string, protocol string) *CheckResult {
	result := &CheckResult{
		StatusMsg: "未执行",
	}

	url := protocol + "://" + domain
	start := time.Now()

	// 确保域名包含端口 (默认 80/443)
	addr := domain
	if !strings.Contains(addr, ":") {
		if protocol == "https" {
			addr += ":443"
		} else {
			addr += ":80"
		}
	}

	// --- 自定义 Transport/Dialer 以获取 SSL 信息 ---
	// 必须使用 net.Dialer + tls.Client 来手动控制连接和获取证书

	dialer := &net.Dialer{Timeout: 10 * time.Second}

	if protocol == "https" {
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: domain, // 确保 SNI (Server Name Indication) 正确
			// 允许非验证连接以获取证书信息，否则如果证书无效，Go 会直接返回错误
			InsecureSkipVerify: true,
		})

		result.Duration = time.Since(start)

		if err != nil {
			result.ErrorLog = fmt.Sprintf("TLS 握手失败: %v", err)
			result.StatusMsg = "连接或 TLS 握手失败"
			return result
		}
		defer conn.Close()

		// 1. 获取 SSL 证书信息
		state := conn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			cert := state.PeerCertificates[0] // 始终取第一个证书 (叶证书)

			result.SSLExpiryDate = cert.NotAfter
			result.SSLCommonName = cert.Subject.CommonName
			result.SSLIssuerName = cert.Issuer.CommonName
			result.SSLSignatureAlg = cert.SignatureAlgorithm.String()

			// SANs 列表平铺为逗号分隔字符串
			result.SANList = strings.Join(cert.DNSNames, ",")

			// 证书链信息 (PEM 格式文本)
			var chainBuilder strings.Builder
			for _, c := range state.PeerCertificates {
				pemBlock := &pem.Block{
					Type:  "CERTIFICATE",
					Bytes: c.Raw,
				}
				chainBuilder.WriteString(string(pem.EncodeToMemory(pemBlock)))
			}
			result.CertificateChain = chainBuilder.String()
		}

		// 2. 尝试执行 HTTP 请求 (如果需要检查内容)
		// 简化：这里只检查了 SSL 握手成功和证书获取，实际可能还需要再发一个 HTTP GET

		// 假设 SSL 握手成功且证书信息获取，则继续尝试 GET
		result.StatusMsg = "TLS 握手成功"
	}

	// --- 执行 HTTP/HTTPS GET 请求 ---
	// 为了获取 HTTP 状态码和 IP 地址，无论是否是 HTTPS，都需要执行一次请求

	client := http.Client{
		// 再次设置 10 秒超时
		Timeout: 10 * time.Second,
		// 如果是 HTTPS 且证书已获取，需要确保 http.Client 也能处理
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略 Go 自己的证书校验，我们已经手动获取
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// 获取 IP 解析列表 (如果您需要保存 ip_resolved 字段)
				// ips, _ := net.LookupIP(domain)
				// result.IPResolved = ...

				// 使用默认 Dialer
				return (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext(ctx, network, addr)
			},
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := client.Do(req)
	result.Duration = time.Since(start) // 重新计算总耗时

	if err != nil {
		// 如果在 TLS 阶段已经失败，则这个 error 会是连接或超时错误
		if result.ErrorLog == "" { // 如果之前没有记录错误
			result.ErrorLog = err.Error()
		}
		if result.StatusMsg == "未执行" {
			result.StatusMsg = "HTTP/HTTPS 请求失败"
		}
		return result
	}
	defer resp.Body.Close()

	// 3. 处理 HTTP 状态
	result.HTTPStatus = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsSuccess = true
		result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode)
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	return result
}
