package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/contrib/resty_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
	"resty.dev/v3"
)

// InitThirdParty 初始化 第三方
func InitThirdParty(ctx context.Context, logger *slog.Logger) {

	var err error

	// Resty 客户端
	var restyClient = resty.New()
	var restyClientBool = false

	// Resty 日志
	var restyLog *resty_log.LoggerMiddleware
	var restyLogBool = false

	if global.GvaRestyLog != nil {
		restyLog = global.GvaRestyLog
		restyLogBool = true
	} else if global.GvaRestyClient != nil {
		restyClient = global.GvaRestyClient
		restyClientBool = true
	}

	// 企业微信
	global.GvaWechatWork, err = wechatqy.NewClient(
		ctx,
		wechatqy.WithRestyClientIf(restyClientBool, restyClient),
		wechatqy.WithRestyLogIf(restyLogBool, restyLog),
		wechatqy.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 企业微信] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 钉钉
	global.GvaDingTalk, err = dingtalk.NewClient(
		ctx,
		dingtalk.WithRestyClientIf(restyClientBool, restyClient),
		dingtalk.WithRestyLogIf(restyLogBool, restyLog),
		dingtalk.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 钉钉] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 飞书
	global.GvaFeishu, err = feishu.NewClient(
		ctx,
		feishu.WithRestyClientIf(restyClientBool, restyClient),
		feishu.WithRestyLogIf(restyLogBool, restyLog),
		feishu.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 飞书] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

	// bark
	global.GvaBark, err = bark.NewClient(
		ctx,
		bark.WithRestyClientIf(restyClientBool, restyClient),
		bark.WithRestyLogIf(restyLogBool, restyLog),
		bark.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 bark] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

	// ft07
	global.GvaFt07, err = ft07.NewClient(
		ctx,
		ft07.WithRestyClientIf(restyClientBool, restyClient),
		ft07.WithRestyLogIf(restyLogBool, restyLog),
		ft07.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 ft07] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

	// pushdeer
	global.GvaPushDeer, err = pushdeer.NewClient(
		ctx,
		pushdeer.WithRestyClientIf(restyClientBool, restyClient),
		pushdeer.WithRestyLogIf(restyLogBool, restyLog),
		pushdeer.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 pushdeer] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

}

// CloseThirdParty 关闭 第三方
func CloseThirdParty() {
	if global.GvaWechatWork != nil {
		_ = global.GvaWechatWork.Close()
	}
	if global.GvaDingTalk != nil {
		_ = global.GvaDingTalk.Close()
	}
	if global.GvaFeishu != nil {
		_ = global.GvaFeishu.Close()
	}
	if global.GvaBark != nil {
		_ = global.GvaBark.Close()
	}
	if global.GvaFt07 != nil {
		_ = global.GvaFt07.Close()
	}
	if global.GvaPushDeer != nil {
		_ = global.GvaPushDeer.Close()
	}
}
