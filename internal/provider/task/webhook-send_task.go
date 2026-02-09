package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/provider"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
	"go.dtapp.net/library/utils/gojson"
	"go.dtapp.net/library/utils/gorequest"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

// WebhookSendTaskProvider 实现 provider.TaskProvider 接口
type WebhookSendTaskProvider struct {
	provider.UnimplementedTaskProvider // 嵌入未实现的

	db            *ent.Client                   // 数据库客户端
	logDb         *ent_log.Client               // 日志数据库客户端
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器

	client           *resty.Client    // Resty HTTP 客户端
	wechatWorkClient *wechatqy.Client // 企业微信
	dingTalkClient   *dingtalk.Client // 钉钉
	feishuClient     *feishu.Client   // 飞书
	barkClient       *bark.Client     // Bark
	ft07Client       *ft07.Client     // Server酱
	pushDeerClient   *pushdeer.Client // PushDeer
}

// NewWebhookSendTaskProvider 创建一个新的 Webhook发送 提供者
func NewWebhookSendTaskProvider(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) provider.TaskProvider {
	return NewWebhookSendTaskProviderRaw(ctx, db, logDb, logger, httpTransport)
}

// NewWebhookSendTaskProvider 创建一个新的 Webhook发送 提供者
func NewWebhookSendTaskProviderRaw(ctx context.Context, db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *WebhookSendTaskProvider {

	// 创建 Resty 客户端
	client := resty.New()

	// 添加 HTTP 日志记录器
	if httpTransport != nil {
		client.SetTransport(httpTransport.Instance(client.Transport()))
	}

	// 企业微信
	wechatWorkClient, _ := wechatqy.NewClient(
		ctx,
		wechatqy.WithRestyClient(client),
	)

	// 钉钉
	dingTalkClient, _ := dingtalk.NewClient(
		ctx,
		dingtalk.WithRestyClient(client),
	)

	// 飞书
	feishuClient, _ := feishu.NewClient(
		ctx,
		feishu.WithRestyClient(client),
	)

	// bark
	barkClient, _ := bark.NewClient(
		ctx,
		bark.WithRestyClient(client),
	)

	// ft07
	ft07Client, _ := ft07.NewClient(
		ctx,
		ft07.WithRestyClient(client),
	)

	// pushdeer
	pushDeerClient, _ := pushdeer.NewClient(
		ctx,
		pushdeer.WithRestyClient(client),
	)

	return &WebhookSendTaskProvider{
		db:               db,
		logDb:            logDb,
		logger:           logger.With(slog.String("handler", "webhook-send")),
		httpTransport:    httpTransport,
		client:           client,
		wechatWorkClient: wechatWorkClient,
		dingTalkClient:   dingTalkClient,
		feishuClient:     feishuClient,
		barkClient:       barkClient,
		ft07Client:       ft07Client,
		pushDeerClient:   pushDeerClient,
	}
}

// WebhookSendTaskName 定义了存放在 CronJob.handler_name 中的任务名称
const (
	WebhookSendTaskName = "webhook_send"
)

// WebhookSendTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type WebhookSendTaskArgs struct {
	EventTitle     string `json:"event_title,omitempty"`      // 事件标题
	EventType      string `json:"event_type,omitempty"`       // 事件类型
	EventTime      string `json:"event_time,omitempty"`       // 事件时间
	EventBody      string `json:"event_body,omitempty"`       // 事件内容
	DeviceCode     string `json:"device_code,omitempty"`      // 设备编码 Device->Code
	DeviceName     string `json:"device_name,omitempty"`      // 设备名称 Device->Name
	IPv4Addr       string `json:"ipv4_addr,omitempty"`        // IP地址V4 Device->IPv4
	IPv6Addr       string `json:"ipv6_addr,omitempty"`        // IP地址V6 Device->IPv6
	RecordDomain   string `json:"record_domain,omitempty"`    // 记录域名 CDNRecord/DNSRecord->Domain
	RecordType     string `json:"record_type,omitempty"`      // 记录类型 CDNRecord/DNSRecord->Type
	RecordOldValue string `json:"record_old_value,omitempty"` // 旧记录值 CDNRecord/DNSRecord->Record
	RecordNewValue string `json:"record_new_value,omitempty"` // 新记录值 CDNRecord/DNSRecord->Record
	PlatformName   string `json:"platform_name,omitempty"`    // 平台名称 PlatformConfig->Name
	PlatformType   string `json:"platform_type,omitempty"`    // 平台类型 PlatformConfig->Platform
}

// Process 执行 Webhook发送 逻辑
func (p *WebhookSendTaskProvider) Process(ctx context.Context, args json.RawMessage, meta provider.TaskMetadata) (string, error) {

	// 解析参数
	var params WebhookSendTaskArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	var objType webhooksubscription.ObjType
	if meta.BusinessType == cronjob.BusinessTypeCdn {
		objType = webhooksubscription.ObjTypeCdnRecord
	} else if meta.BusinessType == cronjob.BusinessTypeDNS {
		objType = webhooksubscription.ObjTypeDNSRecord
	} else if meta.BusinessType == cronjob.BusinessTypeDevice {
		objType = webhooksubscription.ObjTypeDevice
	} else if meta.BusinessType == cronjob.BusinessTypeDomain {
		objType = webhooksubscription.ObjTypeDomainMonitor
	} else if meta.BusinessType == cronjob.BusinessTypeSsl {
		objType = webhooksubscription.ObjTypeDomainMonitor
	} else if meta.BusinessType == cronjob.BusinessTypeTun {
		objType = webhooksubscription.ObjTypeTunRecord
	}

	if objType == "" {

		// 查询记录
		firstWebhook, err := p.db.Webhook.Query().
			Where(webhook.IDEQ(meta.BusinessID)).
			First(ctx)
		if err != nil {
			return "", fmt.Errorf("获取记录失败: %w", err)
		}

		templateTitle, err := p.RenderTemplate(firstWebhook.TemplateTitle, params)
		if err != nil {
			return "", fmt.Errorf("渲染标题模板失败: %w", err)
		}

		templateContent, err := p.RenderTemplate(firstWebhook.TemplateContent, params)
		if err != nil {
			return "", fmt.Errorf("渲染内容模板失败: %w", err)
		}

		var resp json.RawMessage
		var rErr error
		switch firstWebhook.Platform {
		case webhook.PlatformWechatWork:
			resp, rErr = p.WechatWork(ctx, firstWebhook.URL, templateTitle, templateContent)
		case webhook.PlatformDingtalk:
			resp, rErr = p.Dingtalk(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
		case webhook.PlatformFeishu:
			resp, rErr = p.Feishu(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
		case webhook.PlatformBark:
			resp, rErr = p.Bark(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
		case webhook.PlatformFt07:
			resp, rErr = p.Ft07(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
		case webhook.PlatformPushDeer:
			resp, rErr = p.PushDeer(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
		default:
			resp, rErr = p.Http(ctx, firstWebhook.URL, templateTitle, templateContent)
		}

		// 记录到域名监控历史
		_ = p.logDb.WebhookHistory.Create().
			SetWebhookID(meta.BusinessID).
			SetURL(firstWebhook.URL).
			SetTitle(templateTitle).
			SetContent(templateContent).
			SetResponse(resp).
			SetSuccess(err == nil).
			SetErrorMessage(fmt.Sprintf("%v", err)).
			Exec(ctx)

		if rErr != nil {
			return "", rErr
		}
	} else {

		// 查询记录及其关联信息
		records, err := p.db.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(objType)).
			Where(webhooksubscription.ObjIDEQ(meta.BusinessID)).
			Where(webhooksubscription.IsActiveEQ(true)).
			WithWebhook(func(withWebhook *ent.WebhookQuery) {
				withWebhook.Where(webhook.IsActiveEQ(true))
			}).
			All(ctx)
		if err != nil {
			return "", fmt.Errorf("获取记录失败: %w", err)
		}

		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(10) // 最多 10 个并发

		for _, record := range records {
			record := record // 创建副本以避免闭包问题
			eg.Go(func() (err error) {

				// 获取关联的通知
				recordWebhook, err := record.Edges.WebhookOrErr()
				if err != nil {
					return nil
				}

				templateTitle, err := p.RenderTemplate(record.Edges.Webhook.TemplateTitle, params)
				if err != nil {
					p.logger.ErrorContext(egCtx, "渲染标题模板失败",
						slog.Int64("webhook_id", recordWebhook.ID),
						slog.String("webhook_platform", recordWebhook.Platform.String()),
						slog.Any("err", err),
					)
					return nil
				}

				templateContent, err := p.RenderTemplate(record.Edges.Webhook.TemplateContent, params)
				if err != nil {
					p.logger.ErrorContext(egCtx, "渲染内容模板失败",
						slog.Int64("webhook_id", recordWebhook.ID),
						slog.String("webhook_platform", recordWebhook.Platform.String()),
						slog.Any("err", err),
					)
					return nil
				}

				var resp json.RawMessage
				var rErr error
				switch recordWebhook.Platform {
				case webhook.PlatformWechatWork:
					resp, rErr = p.WechatWork(egCtx, recordWebhook.URL, templateTitle, templateContent)
				case webhook.PlatformDingtalk:
					resp, rErr = p.Dingtalk(egCtx, recordWebhook.URL, recordWebhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFeishu:
					resp, rErr = p.Feishu(egCtx, recordWebhook.URL, recordWebhook.Secret, templateTitle, templateContent)
				case webhook.PlatformBark:
					resp, rErr = p.Bark(egCtx, recordWebhook.URL, recordWebhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFt07:
					resp, rErr = p.Ft07(egCtx, recordWebhook.URL, recordWebhook.Secret, templateTitle, templateContent)
				case webhook.PlatformPushDeer:
					resp, rErr = p.PushDeer(egCtx, recordWebhook.URL, recordWebhook.Secret, templateTitle, templateContent)
				default:
					resp, rErr = p.Http(egCtx, recordWebhook.URL, templateTitle, templateContent)
				}

				// 记录到域名监控历史
				_ = p.logDb.WebhookHistory.Create().
					SetWebhookID(recordWebhook.ID).
					SetSubscriptionID(record.ID).
					SetURL(recordWebhook.URL).
					SetTitle(templateTitle).
					SetContent(templateContent).
					SetResponse(resp).
					SetSuccess(rErr == nil).
					SetErrorMessage(fmt.Sprintf("%v", rErr)).
					Exec(egCtx)

				if rErr != nil {
					return rErr
				}

				return nil
			})
		}

		err = eg.Wait()
		if err != nil {
			return "", fmt.Errorf("并发错误: %w", err)
		}

	}

	return fmt.Sprintf("发送成功"), nil
}

// RenderTemplate 使用 Go Template 渲染标题和内容
func (p *WebhookSendTaskProvider) RenderTemplate(tmplString string, data WebhookSendTaskArgs) (string, error) {
	// 创建一个新的模板实例
	tmpl := template.New("webhook_template")

	// 将默认的 {{ 和 }} 改为 [[ 和 ]]
	tmpl.Delims("[[", "]]")

	// 解析模板字符串
	tmpl, err := tmpl.Parse(tmplString)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var buf bytes.Buffer

	// 执行渲染
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}

	return buf.String(), nil
}

// WechatWork Webhook 发送
// https://developer.work.weixin.qq.com/document/path/99110
func (p *WebhookSendTaskProvider) WechatWork(ctx context.Context, url string, title string, content any) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", map[string]any{
		"content": content,
	})

	// 请求
	resp, err := p.wechatWorkClient.WebhookSendURL(
		ctx,
		url,
		"markdown",
		params,
	)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Dingtalk Webhook 发送
// https://open.dingtalk.com/document/group/custom-robot-access
func (p *WebhookSendTaskProvider) Dingtalk(ctx context.Context, url string, secret string, title string, content any) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", content)

	// 请求
	var resp dingtalk.RobotSend
	if secret != "" {
		resp, err = p.dingTalkClient.RobotSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
	} else {
		resp, err = p.dingTalkClient.RobotSendURL(
			ctx,
			url,
			params,
		)
	}
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Feishu Webhook 发送
// https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot
func (p *WebhookSendTaskProvider) Feishu(ctx context.Context, url string, secret string, title string, content string) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msg_type", "post")
	params.Set("content", map[string]any{
		"post": map[string]any{
			"zh_cn": map[string]any{
				"title":   title,
				"content": p.StringToTextArray(content),
			},
		}})

	// 请求
	var resp feishu.WebhookSend
	if secret != "" {
		resp, err = p.feishuClient.WebhookSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
	} else {
		resp, err = p.feishuClient.WebhookSendURL(
			ctx,
			url,
			params,
		)
	}
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Bark Webhook 发送
// - 第一行
// - 第二个
func (p *WebhookSendTaskProvider) Bark(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		p.barkClient.SetUrl(url)
	}
	if key != "" {
		p.barkClient.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("body", content)
	params.Set("group", "ip_vibe")
	params.Set("isArchive", 1)

	// 请求
	resp, err := p.barkClient.Push(ctx, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Ft07 Webhook 发送
// - 第一行
// - 第二个
func (p *WebhookSendTaskProvider) Ft07(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		p.ft07Client.SetUrl(url)
	}
	if key != "" {
		p.ft07Client.SetSendKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("desp", content)
	params.Set("tags", "ip_vibe")

	// 请求
	resp, err := p.ft07Client.Push(ctx, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// PushDeer Webhook 发送
// https://www.pushdeer.com/dev.html
func (p *WebhookSendTaskProvider) PushDeer(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		p.pushDeerClient.SetUrl(url)
	}
	if key != "" {
		p.pushDeerClient.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("desp", content)
	params.Set("type", "markdown")

	// 请求
	resp, err := p.pushDeerClient.MessagePush(ctx, title, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Http Webhook 发送
func (p *WebhookSendTaskProvider) Http(ctx context.Context, url string, title string, content any) (response json.RawMessage, err error) {

	// 使用 Resty 发送请求
	resp, err := p.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetBody(map[string]any{
			"title": title,
			"body":  content,
		}).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.Bytes(), nil
}

type TextItem struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

func (p *WebhookSendTaskProvider) StringToTextArray(str string) [][]TextItem {
	lines := strings.Split(str, "\n")
	var result []TextItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// 如果当前行次不是最后一行，添加换行符
			if line != lines[len(lines)-1] {
				line += "\n"
			}
			result = append(result, TextItem{
				Tag:  "text",
				Text: line,
			})
		}
	}

	return [][]TextItem{result}
}

// AddTask 添加 Webhook发送 一次性任务
func (p *WebhookSendTaskProvider) AddTask(ctx context.Context, db *ent.Client, businessType cronjob.BusinessType, businessID int64, args WebhookSendTaskArgs) error {

	var (
		taskType    = cronjob.TaskTypeOneOff
		jobType     = cronjob.JobTypeDuration
		jobCategory = cronjob.JobCategorySystem
	)

	if businessType == cronjob.BusinessTypeWebhook {
		jobCategory = cronjob.JobCategoryCustom
	}

	// 创建任务
	err := db.CronJob.Create().
		SetName("通知: " + args.EventTitle).
		SetHandlerName(WebhookSendTaskName).
		SetHandlerArgs(json.RawMessage(gojson.JsonEncodeNoError(args))).
		SetTaskType(taskType).
		SetJobType(jobType).
		SetJobCategory(jobCategory).
		SetBusinessType(businessType).
		SetBusinessID(businessID).
		SetIntervalValue(1).
		SetIntervalUnit(cronjob.IntervalUnitSeconds).
		SetStatus(cronjob.StatusRunning).
		SetIsSingleton(true). // 单例模式
		SetRunOnStart(true).  // 启动时立即运行
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	return nil
}
