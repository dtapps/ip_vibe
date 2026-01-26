package initialize

import (
	"context"
	"log/slog"
	"sync"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/sys_env"

	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
	"resty.dev/v3"
)

var thirdPartyMutex sync.Once

// InitThirdParty 初始化 第三方
func InitThirdParty(ctx context.Context, logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) {

	thirdPartyMutex.Do(func() {

		var err error

		// 创建 Resty 客户端
		restyClient := resty.New()

		// 开启日志
		restyClient.SetDebug(sys_env.GvaEnv.DEBUG)

		// 添加 HTTP 日志记录器
		if httpTransport != nil {
			restyClient.SetTransport(httpTransport.Instance(restyClient.Transport()))
		}

		// 企业微信
		global.GvaWechatWork, err = wechatqy.NewClient(
			ctx,
			wechatqy.WithRestyClient(restyClient),
			wechatqy.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 企业微信] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

		// 钉钉
		global.GvaDingTalk, err = dingtalk.NewClient(
			ctx,
			dingtalk.WithRestyClient(restyClient),
			dingtalk.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 钉钉] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

		// 飞书
		global.GvaFeishu, err = feishu.NewClient(
			ctx,
			feishu.WithRestyClient(restyClient),
			feishu.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 飞书] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

		// bark
		global.GvaBark, err = bark.NewClient(
			ctx,
			bark.WithRestyClient(restyClient),
			bark.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 bark] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

		// ft07
		global.GvaFt07, err = ft07.NewClient(
			ctx,
			ft07.WithRestyClient(restyClient),
			ft07.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 ft07] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

		// pushdeer
		global.GvaPushDeer, err = pushdeer.NewClient(
			ctx,
			pushdeer.WithRestyClient(restyClient),
			pushdeer.WithDebug(sys_env.GvaEnv.DEBUG),
		)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 pushdeer] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}

	})
}
