package global

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/helper/operation_logs_helper"

	"github.com/go-co-op/gocron/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/contrib/resty_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
	"go.dtapp.net/library/utils/golog"
)

var (
	DefaultTimezone string = "Asia/Shanghai" // 默认时区
)

// GvaConfig 变量
var (
	GvaEnv struct {
		DEBUG bool // [backend][reporter][forwarder] 是否开启调试模式

		HTTP_ADDR string // [backend] HTTP 地址
		GRPC_ADDR string // [backend] gRPC 地址

		HTTP_URL string // [reporter][forwarder] HTTP URL
		GRPC_URL string // [reporter][forwarder] gRPC URL

		BASIC_AUTH_USERNAME string // [backend][reporter][forwarder] 基本认证用户名
		BASIC_AUTH_PASSWORD string // [backend][reporter][forwarder] 基本认证密码
		DEVICE_CODE         string // [backend][reporter][forwarder] 设备代码
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

	GvaMcpServer *mcp.Server // Model Context Protocol 服务器

	GvaEntDb    *ent.Client     // 业务数据库
	GvaEntLogDb *ent_log.Client // 日志数据库

	GvaOperationLogHelper *operation_logs_helper.OperationLogHelper // 操作日志助手

	GvaGinSlog  *golog.SLog                // gin 日志
	GvaGrpcSlog *golog.SLog                // gRPC 日志
	GvaHttpLog  *http_log.LoggingTransport // HTTP框架日志
	GvaRestyLog *resty_log.Logger          // Resty框架日志

	GvaWechatWork *wechatqy.Client // 企业微信
	GvaDingTalk   *dingtalk.Client // 钉钉
	GvaFeishu     *feishu.Client   // 飞书
	GvaBark       *bark.Client     // Bark
	GvaFt07       *ft07.Client     // Server酱
	GvaPushDeer   *pushdeer.Client // PushDeer
)
