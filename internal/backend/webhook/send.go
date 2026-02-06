package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
	"go.dtapp.net/library/utils/gorequest"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

type Send struct {
	db     *ent.Client     // 数据库客户端
	logDb  *ent_log.Client // 日志数据库客户端
	client *resty.Client   // Resty HTTP 客户端
	logger *slog.Logger    // 日志记录器

	wechatWorkClient *wechatqy.Client // 企业微信
	dingTalkClient   *dingtalk.Client // 钉钉
	feishuClient     *feishu.Client   // 飞书
	barkClient       *bark.Client     // Bark
	ft07Client       *ft07.Client     // Server酱
	pushDeerClient   *pushdeer.Client // PushDeer
}

func NewSend(db *ent.Client, logDb *ent_log.Client) *Send {

	// 创建 Resty 客户端
	client := resty.New()

	// 开启日志
	client.SetDebug(sys_env.GvaEnv.DEBUG)

	// 设置默认超时
	client.SetTimeout(30 * time.Second)

	// 跳过证书验证
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &Send{
		db:     db,
		logDb:  logDb,
		client: client,
		logger: slog.With("component", "webhook"),

		// wechatWorkClient: global.GvaWechatWork,
		// dingTalkClient:   global.GvaDingTalk,
		// feishuClient:     global.GvaFeishu,
		// barkClient:       global.GvaBark,
		// ft07Client:       global.GvaFt07,
		// pushDeerClient:   global.GvaPushDeer,
	}
}

// 发送通知
func (s *Send) Send(ctx context.Context, objType webhooksubscription.ObjType, objID int64, payload types.TemplateData) (err error) {
	allWebhook, err := s.db.WebhookSubscription.Query().
		Where(
			webhooksubscription.ObjTypeEQ(objType),
			webhooksubscription.ObjIDEQ(objID),
		).
		Where(webhooksubscription.IsActiveEQ(true)).
		WithWebhook(func(withWebhook *ent.WebhookQuery) {
			withWebhook.Where(webhook.IsActiveEQ(true))
		}).
		All(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "查询通知订阅失败",
			slog.String("obj_type", objType.String()),
			slog.Int64("obj_id", objID),
			slog.Any("err", err),
		)
		return fmt.Errorf("查询通知订阅失败: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

	for _, vWebhook := range allWebhook {
		item := vWebhook // 创建副本以避免闭包问题
		eg.Go(func() (err error) {

			if item.Edges.Webhook != nil {

				templateTitle, err := RenderTemplate(item.Edges.Webhook.TemplateTitle, payload)
				if err != nil {
					s.logger.ErrorContext(egCtx, "渲染标题模板失败",
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
						slog.Any("err", err),
					)
					return nil
				}

				templateContent, err := RenderTemplate(item.Edges.Webhook.TemplateContent, payload)
				if err != nil {
					s.logger.ErrorContext(egCtx, "渲染内容模板失败",
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
						slog.Any("err", err),
					)
					return nil
				}

				var resp json.RawMessage
				var rErr error
				switch item.Edges.Webhook.Platform {
				case webhook.PlatformWechatWork:
					resp, rErr = s.WechatWork(egCtx, item.Edges.Webhook.URL, templateTitle, templateContent)
				case webhook.PlatformDingtalk:
					resp, rErr = s.Dingtalk(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFeishu:
					resp, rErr = s.Feishu(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformBark:
					resp, rErr = s.Bark(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFt07:
					resp, rErr = s.Ft07(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformPushDeer:
					resp, rErr = s.PushDeer(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				default:
					resp, rErr = s.Http(egCtx, item.Edges.Webhook.URL, templateTitle, templateContent)
				}

				// 记录到域名监控历史
				_ = s.logDb.WebhookHistory.Create().
					SetWebhookID(item.Edges.Webhook.ID).
					SetSubscriptionID(item.ID).
					SetURL(item.Edges.Webhook.URL).
					SetTitle(templateTitle).
					SetContent(templateContent).
					SetResponse(resp).
					SetSuccess(rErr == nil).
					SetErrorMessage(fmt.Sprintf("%v", rErr)).
					Exec(egCtx)

				if rErr != nil {
					s.logger.ErrorContext(egCtx, "发送Webhook失败",
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
						slog.Any("err", rErr),
					)
				} else {
					s.logger.InfoContext(egCtx, "Webhook发送成功",
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
					)
				}
			}

			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		s.logger.ErrorContext(ctx, "并发错误",
			slog.String("obj_type", objType.String()),
			slog.Int64("obj_id", objID),
			slog.Any("err", err),
		)
	}

	return nil
}

// 测试通知
func (s *Send) Test(ctx context.Context, id int64, payload types.TemplateData) (err error) {
	// 查询通知
	firstWebhook, err := s.db.Webhook.Query().
		Where(webhook.IDEQ(id)).
		First(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "查询通知失败",
			slog.Int64("webhook_id", id),
			slog.Any("err", err),
		)
		return fmt.Errorf("查询通知失败: %w", err)
	}

	templateTitle, err := RenderTemplate(firstWebhook.TemplateTitle, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "渲染标题模板失败",
			slog.Int64("webhook_id", id),
			slog.Any("err", err),
		)
		return fmt.Errorf("渲染标题模板失败: %w", err)
	}

	templateContent, err := RenderTemplate(firstWebhook.TemplateContent, payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "渲染内容模板失败",
			slog.Int64("webhook_id", id),
			slog.Any("err", err),
		)
		return fmt.Errorf("渲染内容模板失败: %w", err)
	}

	var resp json.RawMessage
	switch firstWebhook.Platform {
	case webhook.PlatformWechatWork:
		resp, err = s.WechatWork(ctx, firstWebhook.URL, templateTitle, templateContent)
	case webhook.PlatformDingtalk:
		resp, err = s.Dingtalk(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
	case webhook.PlatformFeishu:
		resp, err = s.Feishu(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
	case webhook.PlatformBark:
		resp, err = s.Bark(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
	case webhook.PlatformFt07:
		resp, err = s.Ft07(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
	case webhook.PlatformPushDeer:
		resp, err = s.PushDeer(ctx, firstWebhook.URL, firstWebhook.Secret, templateTitle, templateContent)
	default:
		resp, err = s.Http(ctx, firstWebhook.URL, templateTitle, templateContent)
	}

	// 记录到域名监控历史
	_ = s.logDb.WebhookHistory.Create().
		SetWebhookID(id).
		SetURL(firstWebhook.URL).
		SetTitle(templateTitle).
		SetContent(templateContent).
		SetResponse(resp).
		SetSuccess(err == nil).
		SetErrorMessage(fmt.Sprintf("%v", err)).
		Exec(ctx)

	if err != nil {
		s.logger.ErrorContext(ctx, "发送Webhook失败",
			slog.Int64("webhook_id", id),
			slog.Any("err", err),
		)
		return fmt.Errorf("发送Webhook失败: %w", err)
	} else {
		s.logger.InfoContext(ctx, "Webhook发送成功",
			slog.Int64("webhook_id", id),
		)
		return nil
	}
}

// WechatWork Webhook 发送
// https://developer.work.weixin.qq.com/document/path/99110
func (s *Send) WechatWork(ctx context.Context, url string, title string, content any) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", map[string]any{
		"content": content,
	})

	// 请求
	resp, err := s.wechatWorkClient.WebhookSendURL(
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
func (s *Send) Dingtalk(ctx context.Context, url string, secret string, title string, content any) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", content)

	// 请求
	var resp dingtalk.RobotSend
	if secret != "" {
		resp, err = s.dingTalkClient.RobotSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
	} else {
		resp, err = s.dingTalkClient.RobotSendURL(
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
func (s *Send) Feishu(ctx context.Context, url string, secret string, title string, content string) (response json.RawMessage, err error) {

	// 参数
	params := gorequest.NewParams()
	params.Set("msg_type", "post")
	params.Set("content", map[string]any{
		"post": map[string]any{
			"zh_cn": map[string]any{
				"title":   title,
				"content": StringToTextArray(content),
			},
		}})

	// 请求
	var resp feishu.WebhookSend
	if secret != "" {
		resp, err = s.feishuClient.WebhookSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
	} else {
		resp, err = s.feishuClient.WebhookSendURL(
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
func (s *Send) Bark(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		s.barkClient.SetUrl(url)
	}
	if key != "" {
		s.barkClient.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("body", content)
	params.Set("group", "ip_vibe")
	params.Set("isArchive", 1)

	// 请求
	resp, err := s.barkClient.Push(ctx, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Ft07 Webhook 发送
// - 第一行
// - 第二个
func (s *Send) Ft07(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		s.ft07Client.SetUrl(url)
	}
	if key != "" {
		s.ft07Client.SetSendKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("desp", content)
	params.Set("tags", "ip_vibe")

	// 请求
	resp, err := s.ft07Client.Push(ctx, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// PushDeer Webhook 发送
// https://www.pushdeer.com/dev.html
func (s *Send) PushDeer(ctx context.Context, url string, key string, title string, content any) (response json.RawMessage, err error) {

	if url != "" {
		s.pushDeerClient.SetUrl(url)
	}
	if key != "" {
		s.pushDeerClient.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("desp", content)
	params.Set("type", "markdown")

	// 请求
	resp, err := s.pushDeerClient.MessagePush(ctx, title, params)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(gorequest.JsonEncodeNoError(resp)), nil
}

// Http Webhook 发送
func (s *Send) Http(ctx context.Context, url string, title string, content any) (response json.RawMessage, err error) {

	// 使用 Resty 发送请求
	resp, err := s.client.R().
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
