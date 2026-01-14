package provider

import (
	"context"
	"errors"
	"log/slog"
)

// 未找到记录错误
var ErrRecordNotFound = errors.New("记录不存在")

// PlatformConfigJson 平台配置
type PlatformConfigJson struct {
	DomainName string `json:"domain_name,omitempty"`
	Email      string `json:"email,omitempty"`
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
