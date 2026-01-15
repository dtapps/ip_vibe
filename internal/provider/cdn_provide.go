package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/utils"
)

// ---------- CDN 模块 ----------

// CDNProvider 定义 CDN 提供程序接口
type CDNProvider interface {
	// Name 返回 CDN 提供程序的名称
	Name() string
	// Check 检查 CDN 记录是否存在
	Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, cdnDomain string, originValues []string, err error)
	// Update 更新 CDN 记录
	Update(ctx context.Context, recordID string, domain string, subdomain string, originValues []string, recordType string) (err error)
	// ParseCustomConfig 解析 CDN 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// CDNTask 定义了 CDN 更新任务
type CDNTask struct {
	logger   *slog.Logger // 日志记录器
	provider CDNProvider  // CDN 提供程序

	config         *CDNConfigParams // 配置
	isNotification bool             // 是否通知

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置
}

func NewCDNTask(logger *slog.Logger, provider CDNProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *CDNTask, err error) {
	t = &CDNTask{
		logger:               logger,
		provider:             provider,
		config:               &CDNConfigParams{},
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}
	if len(t.platform_config_json) > 0 {
		var platformConfig PlatformConfigJson
		if err := json.Unmarshal(t.platform_config_json, &platformConfig); err != nil {
			t.logger.Error("【CDN】解析平台配置失败",
				slog.String("name", t.provider.Name()),
				slog.Any("err", err),
			)
			return nil, err
		}
		t.config.domain = platformConfig.DomainName
	}
	return t, nil
}

type CDNConfigParams struct {
	domain    string // 域名名称
	subdomain string // 子域名

	recordID        string   // 域名ID
	recordType      string   // 记录类型
	cdnDomain       string   // CDN域名
	oldOriginValues []string // 旧记录值列表
	newOriginValues []string // 新记录值列表
}

// Config 配置 DNS 参数
func (t *CDNTask) Config(subdomain string, recordType string, newOriginValues []string) *CDNTask {
	t.config.subdomain = subdomain
	t.config.recordType = recordType
	t.config.newOriginValues = newOriginValues
	return t
}

// Validate 验证 DNS 更新任务配置
func (t *CDNTask) Validate() error {
	if t.config.domain == "" {
		return fmt.Errorf("domain 不能为空")
	}
	if t.config.subdomain == "" {
		return fmt.Errorf("subdomain 不能为空")
	}
	if t.config.recordType == "" {
		return fmt.Errorf("recordType 不能为空")
	}
	if len(t.config.newOriginValues) == 0 {
		return fmt.Errorf("newOriginValues 不能为空")
	}
	return nil
}

// Execute 执行 CDN 更新
func (t *CDNTask) Execute(ctx context.Context) (err error) {
	if t.provider == nil {
		return nil
	}

	// 验证配置
	if err := t.Validate(); err != nil {
		t.logger.Error("【CDN】配置验证失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	t.logger.Info("【CDN】开始更新",
		slog.String("provider", t.provider.Name()),
		slog.String("domain", t.config.domain),
		slog.String("subdomain", t.config.subdomain),
		slog.String("recordType", t.config.recordType),
		slog.String("newValues", strings.Join(t.config.newOriginValues, ",")),
		slog.String("config_json", string(t.record_config_json)),
	)

	if len(t.record_config_json) > 0 {
		if err := t.provider.ParseCustomConfig(t.record_config_json); err != nil {
			t.logger.Error("【CDN】解析自定义配置失败",
				slog.String("name", t.provider.Name()),
				slog.Any("err", err),
			)
			return err
		}
	}

	t.config.recordID, t.config.cdnDomain, t.config.oldOriginValues, err = t.provider.Check(
		ctx,
		t.config.domain,
		t.config.subdomain,
		t.config.recordType,
	)
	if err != nil && !errors.Is(err, ErrRecordNotFound) {
		t.logger.Error("【CDN】检查失败",
			slog.String("recordID", t.config.recordID),
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	if utils.EqualStringSliceUnordered(t.config.oldOriginValues, t.config.newOriginValues) {
		t.logger.Info("【CDN】未改变，无需更新")
		return nil
	}

	if err := t.provider.Update(
		ctx,
		t.config.recordID,
		t.config.domain,
		t.config.subdomain,
		t.config.newOriginValues,
		t.config.recordType,
	); err != nil {
		t.logger.Error("【CDN】更新失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	// 可以下发通知
	t.isNotification = true

	t.logger.Info("【CDN】更新成功")
	return nil
}

// SaveResponse 保存 DNS 记录响应
func (t *CDNTask) SaveResponse(ctx context.Context, db *ent.Client, id int64, last_api_response string) (err error) {
	if db == nil || id == 0 {
		return nil
	}

	err = db.CDNRecord.UpdateOneID(id).
		SetRecordID(t.config.recordID).
		SetCdnDomain(t.config.cdnDomain).
		SetLastAPIResponse(last_api_response).
		SetLastAPIResponseAt(time.Now()).
		Exec(ctx)
	if err != nil {
		t.logger.Error("【CDN】保存响应失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}
	return nil
}

// Notify 执行 DNS 通知
func (t *CDNTask) Notify(ctx context.Context, webhookService *webhookService.Send, id int64) (err error) {
	if webhookService == nil || id == 0 || !t.isNotification {
		return nil
	}

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventTitle:     "CDN记录更新",
		EventType:      "CDN_UPDATE",
		EventTime:      time.Now().Format("2006-01-02 15:04:05"),
		EventBody:      "CDN记录更新事件",
		RecordDomain:   fmt.Sprintf("%s.%s", t.config.subdomain, t.config.domain),
		RecordType:     t.config.recordType,
		RecordOldValue: strings.Join(t.config.oldOriginValues, ","),
		RecordNewValue: strings.Join(t.config.newOriginValues, ","),
	}

	// 发送通知
	if err := webhookService.Send(
		ctx,
		webhooksubscription.ObjTypeCdnRecord,
		id,
		notifyData,
	); err != nil {
		t.logger.Error("【CDN】通知失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
	}

	return nil
}
