package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/utils"
)

// ---------- TUN 模块 ----------

// TUNProviderInternalData 定义了 TUN 提供程序的内部数据
type TUNProviderInternalData struct {
	IsExist    bool   // 是否存在
	IsContinue bool   // 是否继续
	Tips       string // 提示信息
}

// TUNProvider 定义了 TUN 操作的接口
type TUNProvider interface {
	// Name 返回 TUN 提供程序的名称
	Name() (name string)
	// Check 检查 TUN 记录是否存在
	Check(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (data *TUNProviderInternalData, err error)
	// Update 更新 TUN 记录
	Update(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (err error)
	// ParseCustomConfig 解析 TUN 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// TUNTask 定义了 TUN 更新任务
type TUNTask struct {
	logger   *slog.Logger // 日志记录器
	provider TUNProvider  // TUN 提供程序

	config         *TUNConfigParams // 配置
	isNotification bool             // 是否通知

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置
}

// NewTUNTask 创建 TUN 更新任务
func NewTUNTask(logger *slog.Logger, provider TUNProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *TUNTask, err error) {
	t = &TUNTask{
		logger:               logger,
		provider:             provider,
		config:               &TUNConfigParams{},
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}

	// 优化日志
	t.logger = t.logger.With(
		slog.String("模块", "TUN"),
		slog.String("提供商", t.provider.Name()),
	)

	return t, nil
}

type TUNConfigParams struct {
	sub_domain  string // 子域名
	domain_path string // 域名路径
	serve_type  string // 服务类型
	serve_uri   string // 服务URI
}

// Config 配置 TUN 参数
func (t *TUNTask) Config(sub_domain string, domain_path string, serve_type string, serve_uri string) *TUNTask {
	t.config.sub_domain = sub_domain
	t.config.domain_path = domain_path
	t.config.serve_type = serve_type
	t.config.serve_uri = serve_uri
	return t
}

// Validate 验证 TUN 更新任务配置
func (t *TUNTask) Validate() error {
	if t.config.sub_domain == "" {
		return fmt.Errorf("sub_domain 不能为空")
	}
	if t.config.serve_type == "" {
		return fmt.Errorf("serve_type 不能为空")
	}
	if t.config.serve_uri == "" {
		return fmt.Errorf("serve_uri 不能为空")
	}
	return nil
}

// Execute 执行 TUN 更新
func (t *TUNTask) Execute(ctx context.Context) (err error) {
	if t.provider == nil {
		return nil
	}
	t.logger.InfoContext(ctx, "开始执行",
		slog.String("子域名", t.config.sub_domain),
		slog.String("域名路径", t.config.domain_path),
		slog.String("服务类型", t.config.serve_type),
		slog.String("服务URI", t.config.serve_uri),
		slog.String("自定义配置", string(t.record_config_json)),
	)
	defer t.logger.InfoContext(ctx, "执行完成")

	// 验证配置
	t.logger.DebugContext(ctx, "执行验证配置")
	if err := t.Validate(); err != nil {
		t.logger.ErrorContext(ctx, "验证配置失败",
			slog.Any("err", err),
		)
		return err
	}

	if len(t.record_config_json) > 0 {
		t.logger.DebugContext(ctx, "执行解析自定义配置")
		if err := t.provider.ParseCustomConfig(t.record_config_json); err != nil {
			t.logger.ErrorContext(ctx, "解析自定义配置失败",
				slog.Any("err", err),
			)
			return err
		}
	}

	t.logger.DebugContext(ctx, "执行检查")
	checkData, err := t.provider.Check(
		ctx,
		t.config.sub_domain,
		t.config.domain_path,
		t.config.serve_type,
		t.config.serve_uri,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "检查失败",
			slog.Any("err", err),
		)
		return err
	}

	if !checkData.IsContinue {
		t.logger.WarnContext(ctx, "没有变化无需更新")
		return nil
	}

	t.logger.DebugContext(ctx, "执行更新",
		slog.String("提示", checkData.Tips),
	)
	if err := t.provider.Update(
		ctx,
		t.config.sub_domain,
		t.config.domain_path,
		t.config.serve_type,
		t.config.serve_uri,
	); err != nil {
		t.logger.ErrorContext(ctx, "更新失败",
			slog.Any("err", err),
		)
		return err
	}

	// 可以下发通知
	t.isNotification = true

	return nil
}

// SaveRefresh 保存 TUN 记录刷新
func (t *TUNTask) SaveRefresh(ctx context.Context, db *ent.Client, id int64) (err error) {
	if db == nil || id == 0 {
		return nil
	}

	if !t.isNotification {
		return nil
	}

	updateOne := db.TUNRecord.UpdateOneID(id)
	updateOne.SetLastRefreshAt(time.Now())
	err = updateOne.Exec(ctx)
	if err != nil {
		t.logger.ErrorContext(ctx, "保存刷新失败",
			slog.Any("err", err),
		)
		return err
	}
	return nil
}

// SaveResponse 保存 TUN 记录响应
func (t *TUNTask) SaveResponse(ctx context.Context, db *ent.Client, id int64, last_api_response string) (err error) {
	if db == nil || id == 0 {
		return nil
	}

	updateOne := db.TUNRecord.UpdateOneID(id)
	updateOne.SetLastAPIResponse(last_api_response)
	updateOne.SetLastAPIResponseAt(time.Now())
	err = updateOne.Exec(ctx)
	if err != nil {
		t.logger.ErrorContext(ctx, "保存响应失败",
			slog.Any("err", err),
		)
		return err
	}
	return nil
}

// Notify 执行 TUN 通知
func (t *TUNTask) Notify(ctx context.Context, webhookService *webhookService.Send, id int64) (err error) {
	if webhookService == nil || id == 0 {
		return nil
	}

	if !t.isNotification {
		return nil
	}

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventTitle:     "TUN记录更新",
		EventType:      "TUN_UPDATE",
		EventTime:      time.Now().Format("2006-01-02 15:04:05"),
		EventBody:      "TUN记录更新事件",
		RecordType:     t.config.serve_type,
		RecordOldValue: t.config.serve_uri,
	}

	// 发送通知
	if err := webhookService.Send(
		ctx,
		webhooksubscription.ObjTypeTunRecord,
		id,
		notifyData,
	); err != nil {
		t.logger.ErrorContext(ctx, "通知失败",
			slog.Any("err", err),
		)
	}

	return nil
}
