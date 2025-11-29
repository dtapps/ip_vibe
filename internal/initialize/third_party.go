package initialize

import (
	"context"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
)

// InitThirdParty 初始化 第三方
func InitThirdParty(ctx context.Context) (err error) {

	// 企业微信
	global.GvaWechatWork, err = wechatqy.NewClient(
		ctx,
		wechatqy.WithRestyLog(global.GvaRestyLog),
		wechatqy.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	// 钉钉
	global.GvaDingTalk, err = dingtalk.NewClient(
		ctx,
		dingtalk.WithRestyLog(global.GvaRestyLog),
		dingtalk.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	// 飞书
	global.GvaFeishu, err = feishu.NewClient(
		ctx,
		feishu.WithRestyLog(global.GvaRestyLog),
		feishu.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	// bark
	global.GvaBark, err = bark.NewClient(
		ctx,
		bark.WithRestyLog(global.GvaRestyLog),
		bark.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	// ft07
	global.GvaFt07, err = ft07.NewClient(
		ctx,
		ft07.WithRestyLog(global.GvaRestyLog),
		ft07.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	// pushdeer
	global.GvaPushDeer, err = pushdeer.NewClient(
		ctx,
		pushdeer.WithRestyLog(global.GvaRestyLog),
		pushdeer.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return err
	}

	return nil
}

// CloseThirdParty 关闭 第三方
func CloseThirdParty() {
	_ = global.GvaWechatWork.Close()
	_ = global.GvaDingTalk.Close()
	_ = global.GvaFeishu.Close()
	_ = global.GvaBark.Close()
	_ = global.GvaFt07.Close()
	_ = global.GvaPushDeer.Close()
}
