package global

import (
	"vibe_ip/ent"
	"vibe_ip/internal/helper/operation_logs_helper"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/resty_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
)

// GvaConfig 变量
var (
	GvaEnv struct {
		DEBUG               bool   // [backend][reporter] 是否开启调试模式
		HTTP_PORT           string // [backend] HTTP 端口
		HTTP_URL            string // [reporter] HTTP URL
		BASIC_AUTH_USERNAME string // [backend][reporter] 基本认证用户名
		BASIC_AUTH_PASSWORD string // [backend][reporter] 基本认证密码
		DEVICE_CODE         string // [backend][reporter] 设备代码
		DEVICE_NAME         string // [backend][reporter] 设备名称
	}
	GvaVersion   string // 版本号
	GvaGitCommit string // Git 提交 ID
	GvaBuildTime string // 构建时间
	GvaGoVersion string // Go 版本
	GvaOS        string // 操作系统
	GvaArch      string // CPU 架构

	GvaGoCron        gocron.Scheduler // 定时任务
	GvaGoCronRunning int32            // 定时任务状态 0=停 1=运行

	GvaDb *ent.Client // 数据库

	GvaOperationLogHelper *operation_logs_helper.OperationLogHelper // 操作日志助手

	GvaRestyLog *resty_log.Logger // Resty框架日志

	GvaWechatWork *wechatqy.Client // 企业微信
	GvaDingTalk   *dingtalk.Client // 钉钉
	GvaFeishu     *feishu.Client   // 飞书
	GvaBark       *bark.Client     // Bark
	GvaFt07       *ft07.Client     // Server酱
	GvaPushDeer   *pushdeer.Client // PushDeer
)
