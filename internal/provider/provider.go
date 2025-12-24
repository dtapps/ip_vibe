package provider

import (
	"context"
	"log/slog"
)

// ---------- DNS 模块 ----------

// DNSProvider 定义了 DNS 操作的接口
type DNSProvider interface {
	// Check 检查 DNS 记录是否存在
	Check(ctx context.Context, domain string, subdomain string, recordType string) (id string, ip string, err error)
	// Update 更新 DNS 记录
	Update(ctx context.Context, id string, domain string, subdomain string, ip string, recordType string) (err error)
	// Save 保存 DNS 记录
	Save(ctx context.Context, id string, domain string, subdomain string, ip string, recordType string) (err error)
	// Name 返回 DNS 提供程序的名称
	Name() (name string)
}

// DNSTask 定义了 DNS 更新任务
type DNSTask struct {
	logger    *slog.Logger // 日志记录器
	provider  DNSProvider  // DNS 提供程序
	domain    string       // 域名
	subdomain string       // 子域名
}

// NewDNSTask 创建一个新的 DNS 更新任务
func NewDNSTask(logger *slog.Logger, provider DNSProvider, domain, subdomain string) *DNSTask {
	return &DNSTask{
		logger:    logger,
		provider:  provider,
		domain:    domain,
		subdomain: subdomain,
	}
}

func (t *DNSTask) Execute(ctx context.Context, recordType string, newIP string) error {
	if t.provider == nil {
		return nil
	}
	t.logger.Info("【DNS】开始更新",
		slog.String("provider", t.provider.Name()),
		slog.String("domain", t.domain),
		slog.String("subdomain", t.subdomain),
		slog.String("recordType", recordType),
		slog.String("newIP", newIP),
	)

	recordID, oldIP, err := t.provider.Check(ctx, t.domain, t.subdomain, recordType)
	if err != nil {
		t.logger.Error("【DNS】检查失败",
			slog.String("recordID", recordID),
			slog.Any("err", err),
		)
		return err
	}
	if oldIP == newIP {
		t.logger.Info("【DNS】IP 未改变，无需更新")
		return nil
	}

	if err := t.provider.Update(ctx, recordID, t.domain, t.subdomain, newIP, recordType); err != nil {
		t.logger.Error("【DNS】更新失败",
			slog.Any("err", err),
		)
		return err
	}

	if err := t.provider.Save(ctx, recordID, t.domain, t.subdomain, newIP, recordType); err != nil {
		t.logger.Error("【DNS】保存失败",
			slog.Any("err", err),
		)
		return err
	}

	t.logger.Info("【DNS】更新成功")
	return nil
}

// ---------- CDN 模块 ----------

type CDNProvider interface {
	PurgeCache(ctx context.Context, domains []string) error
	Name() string
}

type CDNTask struct {
	logger   *slog.Logger
	provider CDNProvider
	domains  []string
}

func NewCDNTask(logger *slog.Logger, provider CDNProvider, domains []string) *CDNTask {
	return &CDNTask{
		logger:   logger,
		provider: provider,
		domains:  domains,
	}
}

func (t *CDNTask) Execute(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	t.logger.Info("【CDN】开始清理缓存",
		slog.String("provider", t.provider.Name()),
	)

	if err := t.provider.PurgeCache(ctx, t.domains); err != nil {
		t.logger.Error("【CDN】清理失败（非致命）",
			slog.String("err", err.Error()),
		)
		return nil // 不中止整体流程
	}
	t.logger.Info("【CDN】清理成功")
	return nil
}

// ---------- SSL 模块 ----------

type SSLProvider interface {
	Renew(ctx context.Context) error
	Present(domain, token, keyAuth string) error
	CleanUp(domain, token, keyAuth string) error
	Name() string
}

type SSLTask struct {
	logger   *slog.Logger
	provider SSLProvider
}

func NewSSLTask(logger *slog.Logger, provider SSLProvider) *SSLTask {
	return &SSLTask{logger: logger, provider: provider}
}

func (t *SSLTask) Execute(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	t.logger.Info("【SSL】开始续期证书",
		slog.String("provider", t.provider.Name()),
	)

	if err := t.provider.Renew(ctx); err != nil {
		t.logger.Error("【SSL】续期失败",
			slog.String("err", err.Error()),
		)
		return err
	}
	t.logger.Info("【SSL】续期成功")
	return nil
}

// ---------- 通知模块 ----------

type Notifier interface {
	Notify(ctx context.Context, message string) error
	Name() string
}

type NotifyTask struct {
	logger   *slog.Logger
	notifier Notifier
}

func NewNotifyTask(logger *slog.Logger, notifier Notifier) *NotifyTask {
	return &NotifyTask{logger: logger, notifier: notifier}
}

func (t *NotifyTask) Send(message string) {
	if t.notifier == nil {
		return
	}
	ctx := context.Background()
	if err := t.notifier.Notify(ctx, message); err != nil {
		t.logger.Error("通知发送失败",
			slog.String("err", err.Error()),
		)
		return
	}
	t.logger.Info("通知已发送",
		slog.String("message", message),
	)
}
