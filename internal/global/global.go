package global

import (
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/contrib/resty_log"
	"go.dtapp.net/library/service/bark"
	"go.dtapp.net/library/service/dingtalk"
	"go.dtapp.net/library/service/feishu"
	"go.dtapp.net/library/service/ft07"
	"go.dtapp.net/library/service/pushdeer"
	"go.dtapp.net/library/service/wechatqy"
)

var (
	DefaultTimezone string = "Asia/Shanghai" // 默认时区
)

// GvaConfig 变量
var (
	GvaEnv struct {
		DEBUG bool // [backend][reporter][forwarder] 是否开启调试模式

		HTTP_ADDR   string // [backend] HTTP 地址
		GRPC_ADDR   string // [backend] gRPC 地址
		THRIFT_ADDR string // [backend] Thrift 地址

		HTTP_URL   string // [reporter][forwarder] HTTP URL
		GRPC_URL   string // [reporter][forwarder] gRPC URL
		THRIFT_URL string // [reporter][forwarder] Thrift URL

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

	GvaSlog       *slog.Logger // slog 日志
	GvaEntSlog    *slog.Logger // ent 日志
	GvaEntLogSlog *slog.Logger // ent log 日志
	GvaGinSlog    *slog.Logger // gin 日志
	GvaGrpcSlog   *slog.Logger // gRPC 日志
	GvaThriftSlog *slog.Logger // Thrift 日志
	GvaGocronSlog *slog.Logger // gocron 日志
	GvaMcpSlog    *slog.Logger // mcp 日志
	GvaRestySlog  *slog.Logger // resty 日志
	GvaHttpSlog   *slog.Logger // http 日志

	GvaEntDb    *ent.Client     // 业务数据库
	GvaEntLogDb *ent_log.Client // 日志数据库

	GvaGoCron        gocron.Scheduler // 定时任务
	GvaGoCronRunning int32            // 定时任务状态 0=停 1=运行

	GvaHttpLog  *http_log.LoggingTransport // HTTP框架日志
	GvaRestyLog *resty_log.Logger          // Resty框架日志

	GvaWechatWork *wechatqy.Client // 企业微信
	GvaDingTalk   *dingtalk.Client // 钉钉
	GvaFeishu     *feishu.Client   // 飞书
	GvaBark       *bark.Client     // Bark
	GvaFt07       *ft07.Client     // Server酱
	GvaPushDeer   *pushdeer.Client // PushDeer
)
