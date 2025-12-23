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

	"github.com/go-co-op/gocron/v2"
)

// 域名监控任务
type DomainMonitorTask struct {
	db     *ent.Client      // 业务数据库客户端
	logDb  *ent_log.Client  // 日志数据库客户端
	cron   gocron.Scheduler // 定时任务调度器
	dmjob  gocron.Job       // 域名监控任务
	logger *slog.Logger     // 日志记录器
}

// 创建实例
func NewDomainMonitorTask(db *ent.Client, logDb *ent_log.Client, cron gocron.Scheduler, logger *slog.Logger) *DomainMonitorTask {
	return &DomainMonitorTask{
		cron:   cron,
		db:     db,
		logDb:  logDb,
		logger: logger,
	}
}

// Close 关闭服务，释放资源
func (s *DomainMonitorTask) Close() {
	if s.cron != nil && s.dmjob != nil {
		_ = s.cron.RemoveJob(s.dmjob.ID())
	}
}

func (s *DomainMonitorTask) StartDomainMonitorJob(ctx context.Context) (err error) {
	// 每30秒到60秒执行一次
	s.dmjob, err = s.cron.NewJob(
		gocron.DurationRandomJob(
			30*time.Second, // 最小间隔
			60*time.Second, // 最大间隔
		),
		gocron.NewTask(s.CheckAndRunAllMonitors),             // 执行任务
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // 单例模式
		gocron.WithStartAt(gocron.WithStartImmediately()),    // 立即开始
		gocron.WithName("域名监控任务"),                            // 任务名称
		gocron.WithTags("domain-monitor"),                    // 任务标签
	)
	if err != nil {
		return err
	}
	return nil
}

// CheckAndRunAllMonitors 查找并执行到期的监控任务
func (s *DomainMonitorTask) CheckAndRunAllMonitors() {

	var err error
	ctx := context.Background()

	// 查找所有活跃且需要检查的任务
	// 判断逻辑：当前时间 > (上次成功时间 + 检查间隔) 或者上次失败时间 > (上次成功时间 + 检查间隔)

	// 查找所有活跃的监控记录，让后续逻辑判断是否到期
	allMonitor, err := s.db.DomainMonitor.Query().
		Where(domainmonitor.IsActive(true)).
		All(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "查询 DomainMonitor 失败",
			slog.Any("err", err),
		)
		return
	}

	now := time.Now()
	for _, monitor := range allMonitor {
		newLog := s.logger.With(
			slog.Int64("id", monitor.ID),
			slog.String("domain", monitor.Domain),
			slog.String("protocol", monitor.CheckProtocol.String()),
		)

		// 计算预定的下次执行时间
		var nextRun time.Time
		if monitor.LastSuccessAt.IsZero() {
			// 如果从未成功过，立即运行
			nextRun = now
		} else {
			// 下次运行时间 = 上次成功时间 + 检查间隔
			nextRun = monitor.LastSuccessAt.Add(time.Duration(monitor.CheckInterval) * time.Second)
		}

		// 如果是失败状态就使用上次失败时间
		if monitor.CurrentStatus == domainmonitor.CurrentStatusFailure {
			if !monitor.LastFailureAt.IsZero() {
				nextRun = monitor.LastFailureAt.Add(time.Duration(monitor.CheckInterval) * time.Second)
			}
		}

		// 如果已超过或等于下次运行时间，则执行任务
		if now.After(nextRun) || now.Equal(nextRun) {
			newLog.InfoContext(ctx, "准备处理",
				slog.String("protocol", monitor.CheckProtocol.String()),
				slog.Time("last_success_at", monitor.LastSuccessAt),
				slog.Int("check_interval", monitor.CheckInterval),
			)
			// 异步执行任务，避免阻塞主调度任务
			go s.ExecuteMonitorTask(context.Background(), newLog, monitor)
		}
	}
}

// ExecuteMonitorTask 执行单个监控任务
func (s *DomainMonitorTask) ExecuteMonitorTask(ctx context.Context, newLog *slog.Logger, monitor *ent.DomainMonitor) (err error) {

	newLog.InfoContext(ctx, "开始执行检查")
	defer newLog.InfoContext(ctx, "检查完成")

	// 执行检查
	result := s.runDomainCheck(ctx, monitor.Domain, monitor.CheckProtocol.String())

	// 历史记录创建
	historyCreator := s.logDb.DomainMonitorHistory.Create().
		SetMonitorID(monitor.ID).
		SetCheckTime(time.Now()).
		SetIsSuccess(result.IsSuccess).
		SetCheckDuration(int(result.Duration.Milliseconds())).
		SetHTTPStatus(result.HTTPStatus).
		SetStatusMessage(result.StatusMsg).
		SetErrorLog(result.ErrorLog)

	// 填充 SSL 字段
	if monitor.CheckProtocol == domainmonitor.CheckProtocolHTTPS && result.SSLExpiryDate.After(time.Time{}) {
		historyCreator.SetSslExpiryDate(result.SSLExpiryDate).
			SetSslCommonName(result.SSLCommonName).
			SetSslIssuerName(result.SSLIssuerName).
			SetSslSignatureAlg(result.SSLSignatureAlg).
			SetSslSanList(result.SANList).
			SetSslCertificateChain(result.CertificateChain)
	}

	_, err = historyCreator.Save(ctx)
	if err != nil {
		newLog.ErrorContext(ctx, "保存历史记录失败",
			slog.Any("err", err),
		)
		return err
	}

	// 更新状态
	updater := s.db.DomainMonitor.UpdateOne(monitor).
		SetCurrentStatus(domainmonitor.CurrentStatus(s.getCurrentStatus(result.IsSuccess))) // 假设您有状态转换逻辑

	if result.IsSuccess {
		updater.SetLastSuccessAt(time.Now())
	} else {
		updater.SetLastFailureAt(time.Now())
	}

	err = updater.Exec(ctx)
	if err != nil {
		newLog.ErrorContext(ctx, "更新状态失败",
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// 辅助函数：根据检查结果获取 DomainMonitorStatus
func (s *DomainMonitorTask) getCurrentStatus(success bool) domainmonitor.CurrentStatus {
	if success {
		return domainmonitor.CurrentStatusSuccess
	}
	return domainmonitor.CurrentStatusFailure
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
