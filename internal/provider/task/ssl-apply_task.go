package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent/sslrecord"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/sslcertificatehistory"
	"vibe_ip/internal/helper/ssl_helper"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/utils/gostring"
)

// SSLApplyTaskProvider 实现 provider.TaskProvider 接口
type SSLApplyTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
}

// NewSSLApplyTaskProvider 创建一个新的 SSL申请 提供者
func NewSSLApplyTaskProvider(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewSSLApplyTaskProviderRaw(db, logDb, logger, httpTransport)
}

// NewSSLApplyTaskProvider 创建一个新的 SSL申请 提供者
func NewSSLApplyTaskProviderRaw(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *SSLApplyTaskProvider {
	return &SSLApplyTaskProvider{
		db:            db,
		logDb:         logDb,
		logger:        logger.With(slog.String("handler", "ssl-apply")),
		httpTransport: httpTransport,
	}
}

// SSLApplyTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	SSLApplyTaskName = "ssl_apply"
)

// SSLApplyTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type SSLApplyTaskArgs struct {
	ID int64 `json:"id"` // SSLRecord 的数据库 ID
}

// Process 执行 SSL申请 逻辑
func (p *SSLApplyTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params SSLApplyTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 查询单条记录及其关联信息
	record, err := p.db.SSLRecord.Query().
		Where(sslrecord.IDEQ(params.ID)).
		Where(sslrecord.ProviderNEQ(sslrecord.ProviderUnknown)).
		// 预加载 配置
		WithConfig().
		First(ctx)
	if err != nil {
		return "", fmt.Errorf("获取记录失败: %w", err)
	}

	// 判断是否已禁用
	if meta.TaskType == cronjob.TaskTypeRecurring {
		if !record.IsActive {
			return "跳过: 记录已禁用", nil
		}
	}

	// 如果存在过期时间，判断是否需要续期
	if !record.NotAfter.IsZero() {
		// 计算距离过期时间还有多少天
		daysUntilExpire := record.NotAfter.Sub(time.Now()).Hours() / 24
		// 如果大于过期时间就跳过
		if daysUntilExpire > float64(record.RenewalDays) {
			return fmt.Sprintf("距离过期时间还有%.0f天，不需要续期", daysUntilExpire), nil
		}
		// 关闭了自动续期，也跳过
		if !record.IsAutoRenewal {
			return fmt.Sprintf("距离过期时间还有%.0f天，但是关闭了自动续期，跳过", daysUntilExpire), nil
		}
	}

	// 获取关联的平台配置
	platform, err := record.Edges.ConfigOrErr()
	if err != nil {
		return "", fmt.Errorf("获取关联的平台配置失败: %w", err)
	}

	// 创建SSL处理程序
	var sslProcess ssl_helper.SSLProcess

	switch platform.Platform {
	case platformconfig.PlatformCloudflare:
		sslProcess = ssl_helper.NewSSLProcessCloudflare(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformTencentDNS:
		sslProcess = ssl_helper.NewSSLProcessTencentDNS(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformTencentEo:
		sslProcess = ssl_helper.NewSSLProcessTencentEO(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformAliyunDNS:
		sslProcess = ssl_helper.NewSSLProcessAliyunDNS(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformAliyunEsa:
		sslProcess = ssl_helper.NewSSLProcessAliyunEsa(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformBaiduDNS:
		sslProcess = ssl_helper.NewSSLProcessBaiduDNS(
			p.logger,
			p.httpTransport,
			false,
		)
	case platformconfig.PlatformHuaweiDNS:
		sslProcess = ssl_helper.NewSSLProcessHuaweiDNS(
			p.logger,
			p.httpTransport,
			false,
		)
	default:
		return "", fmt.Errorf("暂不支持该平台: %s", platform.String())
	}

	// 配置
	err = sslProcess.Configure(ctx, &ssl_helper.SSLProcessParams{
		SSLRecordID: cast.ToString(record.ID),
		ConfigJSON:  platform.ConfigJSON,
		Domains:     gostring.Split(record.Domain, ","),
	})
	if err != nil {
		return "", fmt.Errorf("配置失败: %w", err)
	}

	// 申请
	resApply, err := sslProcess.Apply(ctx, gostring.Split(record.Domain, ","))
	if err != nil {
		return "", fmt.Errorf("申请失败: %w", err)
	}

	// 解析证书
	parsedCert, err := utils.ParseCertificate(resApply.Certificate)
	if err != nil {
		return "", fmt.Errorf("解析证书失败: %w", err)
	}

	// 准备保存
	updateOne := p.db.SSLRecord.UpdateOne(record)
	createLogOne := p.logDb.SSLCertificateHistory.Create()

	// 更新SSL记录
	updateOne.SetCertDomain(resApply.Domain)
	updateOne.SetCertURL(resApply.CertURL)
	updateOne.SetCertStableURL(resApply.CertStableURL)
	updateOne.SetCertPrivateKey(string(resApply.PrivateKey))
	updateOne.SetCertCertificate(string(resApply.Certificate))
	updateOne.SetCertIssuerCertificate(string(resApply.IssuerCertificate))
	updateOne.SetCertCsr(string(resApply.CSR))
	updateOne.SetNotBefore(parsedCert.NotBefore)
	updateOne.SetNotAfter(parsedCert.NotAfter)
	updateOne.SetStatus(sslrecord.StatusValid)
	updateOne.SetLastRefreshAt(time.Now())
	updateOne.SetLastAPIResponse("ok")
	updateOne.SetLastAPIResponseAt(time.Now())
	err = updateOne.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("更新记录失败: %w", err)
	}

	// 保存到SSL证书申请历史记录
	createLogOne.SetSslRecordID(record.ID)
	createLogOne.SetCertDomain(resApply.Domain)
	createLogOne.SetCertURL(resApply.CertURL)
	createLogOne.SetCertStableURL(resApply.CertStableURL)
	createLogOne.SetCertPrivateKey(string(resApply.PrivateKey))
	createLogOne.SetCertCertificate(string(resApply.Certificate))
	createLogOne.SetCertIssuerCertificate(string(resApply.IssuerCertificate))
	createLogOne.SetCertCsr(string(resApply.CSR))
	createLogOne.SetNotBefore(parsedCert.NotBefore)
	createLogOne.SetNotAfter(parsedCert.NotAfter)
	createLogOne.SetProvider(sslcertificatehistory.Provider(record.Provider))
	err = createLogOne.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("创建历史记录失败: %w", err)
	}

	return fmt.Sprintf("域名 %s 申请成功", record.Domain), nil
}
