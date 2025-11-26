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
	global.GvaWechatWork, err = wechatqy.NewClient(ctx)
	if err != nil {
		return err
	}

	// 钉钉
	global.GvaDingTalk, err = dingtalk.NewClient(ctx)
	if err != nil {
		return err
	}

	// 飞书
	global.GvaFeishu, err = feishu.NewClient(ctx)
	if err != nil {
		return err
	}

	// bark
	global.GvaBark, err = bark.NewClient(ctx)
	if err != nil {
		return err
	}

	// ft07
	global.GvaFt07, err = ft07.NewClient(ctx)
	if err != nil {
		return err
	}

	// pushdeer
	global.GvaPushDeer, err = pushdeer.NewClient(ctx)
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
