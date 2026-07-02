package global

import (
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_stat"

	"github.com/go-co-op/gocron/v2"
	"go.dtapp.net/library/contrib/http_log"
	"go.dtapp.net/library/contrib/req_log"
	"go.dtapp.net/library/contrib/resty_log"
)

// 打包时注入
var (
	Version   = "dev"     // 版本号
	GitCommit = "unknown" // Git 提交 ID
	BuildTime = "unknown" // 构建时间
)

var (
	StartTime time.Time // 启动时间
)

var (
	DefaultTimezone string = "Asia/Shanghai" // 默认时区
)

// GvaConfig 变量
var (
	GvaSlog        *slog.Logger // slog 日志
	GvaEntSlog     *slog.Logger // ent 日志
	GvaEntLogSlog  *slog.Logger // ent log 日志
	GvaEntStatSlog *slog.Logger // ent stat 日志
	GvaGinSlog     *slog.Logger // gin 日志
	GvaGrpcSlog    *slog.Logger // gRPC 日志
	GvaGocronSlog  *slog.Logger // gocron 日志
	GvaMcpSlog     *slog.Logger // mcp 日志
	GvaHttpSlog    *slog.Logger // http 日志
	GvaRestySlog   *slog.Logger // resty 日志
	GvaReqSlog     *slog.Logger // req 日志

	GvaEntDb     *ent.Client      // 业务数据库
	GvaEntLogDb  *ent_log.Client  // 日志数据库
	GvaEntStatDb *ent_stat.Client // 统计数据库

	GvaGoCron        gocron.Scheduler // 定时任务
	GvaGoCronRunning int32            // 定时任务状态 0=停 1=运行

	// GvaHttpClientWithLog        *http.Client  // HTTP 客户端，并带日志记录
	// GvaHttpClientWithLogNoBody  *http.Client  // HTTP 客户端，并带日志记录不包含请求体/响应体
	// GvaRestyClientWithLog       *resty.Client // Resty 客户端，并带日志记录
	// GvaRestyClientWithLogNoBody *resty.Client // Resty 客户端，并带日志记录不包含请求体/响应体
	// GvaReqClientWithLog         *req.Client   // Req 客户端，并带日志记录
	// GvaReqClientWithLogNoBody   *req.Client   // Req 客户端，并带日志记录不包含请求体/响应体

	GvaHttpLog  *http_log.LoggingRoundTripper // HTTP框架日志
	GvaRestyLog *resty_log.LoggerMiddleware   // Resty框架日志
	GvaReqLog   *req_log.LoggerMiddleware     // Req框架日志
)

func init() {
	StartTime = time.Now().UTC()
}
