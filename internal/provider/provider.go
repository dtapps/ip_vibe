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
