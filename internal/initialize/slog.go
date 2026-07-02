package initialize

import (
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/sys_env"

	"go.dtapp.net/library/utils/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getLumberjack 获取 lumberjack 日志
func getLumberjack(prefix string, name string) *lumberjack.Logger {
	// 使用 -拼接 转字符串
	elems := make([]string, 0, 2)
	if prefix != "" {
		elems = append(elems, prefix)
	}
	if name != "" {
		elems = append(elems, name)
	}
	completeName := strings.Join(elems, "_")
	completeName += ".log"

	if global.Version == "dev" {
		return &lumberjack.Logger{
			Filename:   filepath.Join("logs", completeName), // 日志文件的位置
			MaxSize:    3,                                   // 每个日志文件最大尺寸（MB）
			MaxAge:     1,                                   // 留旧文件的最大天数
			MaxBackups: 5,                                   // 保留的最大旧文件数量
			LocalTime:  true,                                // 使用本地时间创建时间戳
			Compress:   true,                                // 是否压缩/归档旧文件
		}
	} else {
		return &lumberjack.Logger{
			Filename:   filepath.Join("logs", completeName), // 日志文件的位置
			MaxSize:    2,                                   // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                  // 留旧文件的最大天数
			MaxBackups: 10 * 15,                             // 保留的最大旧文件数量
			LocalTime:  true,                                // 使用本地时间创建时间戳
			Compress:   true,                                // 是否压缩/归档旧文件
		}
	}
}

// getLevel 获取日志级别
func getLevel() slog.Level {
	if sys_env.GvaEnv.DEBUG {
		return slog.LevelDebug
	} else {
		return slog.LevelInfo
	}
}

// getShowLine 获取是否显示行号
func getShowLine() bool {
	// if sys_env.GvaEnv.DEBUG {
	// 	return true
	// } else {
	return false
	// }
}

var setSlogOnce sync.Once

// InitSLog 初始化 日志
func InitSLog(name string) {
	setSlogOnce.Do(func() {
		global.GvaSlog = golog.NewSlog(
			golog.WithSLogLumberjack(getLumberjack("slog", name)),
			golog.WithSLogSetLevel(getLevel()),
			golog.WithSLogShowLine(getShowLine()),
			golog.WithSLogSetDefault(true),
			golog.WithSLogSetDefaultCtx(true),
		).WithLogger()
	})
}

// InitEntSLog 初始化 ent 日志
func InitEntSLog(name string) {
	prefix := "ent"
	global.GvaEntSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
	global.GvaEntLogSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix+"_log", name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
	global.GvaEntStatSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix+"_stat", name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitGinSLog 初始化 gin 日志
func InitGinSLog(name string) {
	prefix := "gin"
	global.GvaGinSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitGrpcSLog 初始化 gRPC 日志
func InitGrpcSLog(name string) {
	prefix := "grpc"
	global.GvaGrpcSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitGocronSLog 初始化 gocron 日志
func InitGocronSLog(name string) {
	prefix := "gocron"
	global.GvaGocronSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitMcpSLog 初始化 mcp 日志
func InitMcpSLog(name string) {
	prefix := "mcp"
	global.GvaMcpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitHttpSLog 初始化 http 日志
func InitHttpSLog(name string) {
	prefix := "http"
	global.GvaHttpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitRestySLog 初始化 resty 日志
func InitRestySLog(name string) {
	prefix := "resty"
	global.GvaRestySlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

// InitReqSLog 初始化 req 日志
func InitReqSLog(name string) {
	prefix := "req"
	global.GvaReqSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogShowLine(getShowLine()),
	).WithLogger()
}

var setDefaultSlogOnce sync.Once

// EnsureDefaultSLog 若全局日志器为 nil，则统一设置为  global.GvaSlog
func EnsureDefaultSLog() {
	setDefaultSlogOnce.Do(func() {
		if global.GvaEntSlog == nil {
			global.GvaEntSlog = global.GvaSlog
		}
		if global.GvaEntLogSlog == nil {
			global.GvaEntLogSlog = global.GvaSlog
		}
		if global.GvaEntStatSlog == nil {
			global.GvaEntStatSlog = global.GvaSlog
		}
		if global.GvaGinSlog == nil {
			global.GvaGinSlog = global.GvaSlog
		}
		if global.GvaGrpcSlog == nil {
			global.GvaGrpcSlog = global.GvaSlog
		}
		if global.GvaGocronSlog == nil {
			global.GvaGocronSlog = global.GvaSlog
		}
		if global.GvaMcpSlog == nil {
			global.GvaMcpSlog = global.GvaSlog
		}
		if global.GvaRestySlog == nil {
			global.GvaRestySlog = global.GvaSlog
		}
		if global.GvaHttpSlog == nil {
			global.GvaHttpSlog = global.GvaSlog
		}
	})
}
