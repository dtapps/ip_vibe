package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
)

// InitThirdParty 初始化 第三方
func InitThirdParty(ctx context.Context, logger *slog.Logger) {

	var err error

	// 企业微信
	global.GvaWechatWork, err = wechatqy.NewClient(
		ctx,
		wechatqy.WithRestyClient(global.GvaRestyClientWithLog),
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
		dingtalk.WithRestyClient(global.GvaRestyClientWithLog),
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
		feishu.WithRestyClient(global.GvaRestyClientWithLog),
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
		bark.WithRestyClient(global.GvaRestyClientWithLog),
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
		ft07.WithRestyClient(global.GvaRestyClientWithLog),
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
		pushdeer.WithRestyClient(global.GvaRestyClientWithLog),
		pushdeer.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		logger.Error("[初始化 pushdeer] 错误",
			slog.Any("err", err),
		)
		panic(err)
	}

}
