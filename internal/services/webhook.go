package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/webhook"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/utils/gorequest"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

type WebhookService struct {
	db     *ent.Client   // 数据库客户端
	log    *slog.Logger  // 日志记录器
	client *resty.Client // Resty HTTP 客户端
}

func NewWebhookService(db *ent.Client) *WebhookService {

	// 创建一个新的 Resty 客户端
	client := resty.New()

	// 设置默认超时
	client.SetTimeout(30 * time.Second)

	return &WebhookService{
		db:     db,
		log:    slog.With("component", "webhook"),
		client: client,
	}
}

func (s *WebhookService) Send(ctx context.Context, objType webhooksubscription.ObjType, objID int64, payload utils.TemplateData) (err error) {
	allWebhook, err := s.db.WebhookSubscription.Query().
		Where(
			webhooksubscription.ObjTypeEQ(objType),
			webhooksubscription.ObjIDEQ(objID),
		).
		Where(webhooksubscription.IsActiveEQ(true)).
		WithWebhook(func(wq *ent.WebhookQuery) {
			wq.Where(webhook.IsActiveEQ(true))
		}).
		All(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "查询Webhook订阅失败",
			slog.Any("err", err),
			slog.String("obj_type", objType.String()),
			slog.Int64("obj_id", objID),
		)
		return fmt.Errorf("查询Webhook订阅失败: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // 最多 10 个并发

	for _, vWebhook := range allWebhook {
		item := vWebhook // 创建副本以避免闭包问题
		eg.Go(func() (err error) {

			if item.Edges.Webhook != nil {

				templateTitle, err := utils.RenderTemplate(item.Edges.Webhook.TemplateTitle, payload)
				if err != nil {
					s.log.ErrorContext(egCtx, "渲染模板失败",
						slog.Any("err", err),
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
					)
					return nil
				}

				templateContent, err := utils.RenderTemplate(item.Edges.Webhook.TemplateContent, payload)
				if err != nil {
					s.log.ErrorContext(egCtx, "渲染模板失败",
						slog.Any("err", err),
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
					)
					return nil
				}

				switch item.Edges.Webhook.Platform {
				case webhook.PlatformWechatWork:
					err = s.SendWechatWork(egCtx, item.Edges.Webhook.URL, templateTitle, templateContent)
				case webhook.PlatformDingtalk:
					err = s.SendDingtalk(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFeishu:
					err = s.SendFeishu(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformBark:
					err = s.SendBark(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformFt07:
					err = s.SendFt07(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				case webhook.PlatformPushDeer:
					err = s.SendPushDeer(egCtx, item.Edges.Webhook.URL, item.Edges.Webhook.Secret, templateTitle, templateContent)
				default:
					err = s.SendHttp(egCtx, item.Edges.Webhook.URL, templateTitle, templateContent)
				}
				if err != nil {
					s.log.ErrorContext(egCtx, "发送Webhook失败",
						slog.Any("err", err),
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
					)
				} else {
					s.log.InfoContext(egCtx, "Webhook发送成功",
						slog.Int64("webhook_id", item.Edges.Webhook.ID),
						slog.String("webhook_platform", item.Edges.Webhook.Platform.String()),
					)
				}
			}

			// 返回nil而不是错误，这样单个Webhook订阅更新失败不会导致整个errgroup取消上下文
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		s.log.ErrorContext(ctx, "并发错误",
			slog.Any("err", err),
		)
	}

	return nil
}

// 企业微信 Webhook 发送
// https://developer.work.weixin.qq.com/document/path/99110
func (s *WebhookService) SendWechatWork(ctx context.Context, url string, title string, content any) (err error) {

	client := global.GvaWechatWork

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", map[string]any{
		"content": content,
	})

	// 请求
	_, err = client.WebhookSendURL(
		ctx,
		url,
		"markdown",
		params,
	)
	if err != nil {
		return err
	}

	return nil
}

// 钉钉 Webhook 发送
// https://open.dingtalk.com/document/group/custom-robot-access
func (s *WebhookService) SendDingtalk(ctx context.Context, url string, secret string, title string, content any) (err error) {

	client := global.GvaDingTalk

	// 参数
	params := gorequest.NewParams()
	params.Set("msgtype", "markdown")
	params.Set("markdown", content)

	// 请求
	if secret != "" {
		_, err = client.RobotSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
		if err != nil {
			return err
		}
	} else {
		_, err = client.RobotSendURL(
			ctx,
			url,
			params,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// 飞书 Webhook 发送
// https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot
func (s *WebhookService) SendFeishu(ctx context.Context, url string, secret string, title string, content any) (err error) {

	client := global.GvaFeishu

	// 参数
	params := gorequest.NewParams()
	params.Set("msg_type", "post")
	params.Set("content", map[string]any{
		"post": map[string]any{
			"zh_cn": content,
		}})

	// 请求
	if secret != "" {
		_, err = client.WebhookSendSignURL(
			ctx,
			url,
			secret,
			params,
		)
		if err != nil {
			return err
		}
	} else {
		_, err = client.WebhookSendURL(
			ctx,
			url,
			params,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Bark Webhook 发送
// - 第一行
// - 第二个
func (s *WebhookService) SendBark(ctx context.Context, url string, key string, title string, content any) (err error) {

	client := global.GvaBark

	if url != "" {
		client.SetUrl(url)
	}
	if key != "" {
		client.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("body", content)
	params.Set("group", "ip_vibe")
	params.Set("isArchive", 1)

	// 请求
	_, err = client.Push(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// SendFt07 Webhook 发送
// - 第一行
// - 第二个
func (s *WebhookService) SendFt07(ctx context.Context, url string, key string, title string, content any) (err error) {

	client := global.GvaFt07

	if url != "" {
		client.SetUrl(url)
	}
	if key != "" {
		client.SetSendKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("title", title)
	params.Set("desp", content)
	params.Set("tags", "ip_vibe")

	// 请求
	_, err = client.Push(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// PushDeer Webhook 发送
// https://www.pushdeer.com/dev.html
func (s *WebhookService) SendPushDeer(ctx context.Context, url string, key string, title string, content any) (err error) {

	client := global.GvaPushDeer

	if url != "" {
		client.SetUrl(url)
	}
	if key != "" {
		client.SetPushKey(key)
	}

	// 参数
	params := gorequest.NewParams()
	params.Set("desp", content)
	params.Set("type", "markdown")

	// 请求
	_, err = client.MessagePush(ctx, title, params)
	if err != nil {
		return err
	}

	return nil
}

// Webhook 发送
func (s *WebhookService) SendHttp(ctx context.Context, url string, title string, content any) (err error) {

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
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode())
	}

	return nil
}
