package initialize

import (
	"log/slog"
	"path/filepath"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/utils/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getLumberjack 获取 lumberjack 日志
func getLumberjack(prefix string, name string) *lumberjack.Logger {
	if global.GvaVersion == "dev" {
		return &lumberjack.Logger{
			Filename:   filepath.Join("logs", prefix+"_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                             // 每个日志文件最大尺寸（MB）
			MaxAge:     1,                                             // 留旧文件的最大天数
			MaxBackups: 5,                                             // 保留的最大旧文件数量
			LocalTime:  true,                                          // 使用本地时间创建时间戳
			Compress:   true,                                          // 是否压缩/归档旧文件
		}
	} else {
		return &lumberjack.Logger{
			Filename:   filepath.Join("logs", prefix+"_"+name+".log"), // 日志文件的位置
			MaxSize:    2,                                             // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                            // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                       // 保留的最大旧文件数量
			LocalTime:  true,                                          // 使用本地时间创建时间戳
			Compress:   true,                                          // 是否压缩/归档旧文件
		}
	}
}

// InitSLog 初始化 日志
func InitSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	global.GvaSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack("slog", name)),
		golog.WithSLogSetLevel(level),
		golog.WithSLogSetDefault(true),
		golog.WithSLogSetDefaultCtx(true),
	)
}

// InitEntSLog 初始化 ent 日志
func InitEntSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "ent"
	global.GvaEntSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
	global.GvaEntLogSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix+"_log", name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitGinSLog 初始化 gin 日志
func InitGinSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "gin"
	global.GvaGinSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitGrpcSLog 初始化 gRPC 日志
func InitGrpcSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "grpc"
	global.GvaGrpcSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitThriftSLog 初始化 Thrift 日志
func InitThriftSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "thrift"
	global.GvaThriftSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitGocronSLog 初始化 gocron 日志
func InitGocronSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "gocron"
	global.GvaGocronSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitMcpSLog 初始化 mcp 日志
func InitMcpSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "mcp"
	global.GvaMcpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitRestySLog 初始化 resty 日志
func InitRestySLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "resty"
	global.GvaRestySLog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}

// InitHttpSLog 初始化 http 日志
func InitHttpSLog(name string) {
	level := slog.LevelInfo
	if global.GvaEnv.DEBUG {
		level = slog.LevelDebug
	}
	prefix := "http"
	global.GvaHttpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(level),
	)
}
