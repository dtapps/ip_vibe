package initialize

import (
	"log/slog"
	"path/filepath"
	"strings"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/utils/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getLumberjack 获取 lumberjack 日志
func getLumberjack(prefix string, name string) *lumberjack.Logger {
	// 使用 -拼接 转字符串
	elems := make([]string, 0)
	if prefix != "" {
		elems = append(elems, prefix)
	}
	if name != "" {
		elems = append(elems, name)
	}
	elems = append(elems, name+".log")
	completeName := strings.Join(elems, "_")

	if global.GvaVersion == "dev" {
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
	if global.GvaEnv.DEBUG {
		return slog.LevelDebug
	} else {
		return slog.LevelInfo
	}
}

// getShowLine 获取是否显示行号
func getShowLine() bool {
	return true
}

// InitSLog 初始化 日志
func InitSLog(name string) {
	global.GvaSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack("slog", name)),
		golog.WithSLogSetLevel(getLevel()),
		golog.WithSLogSetDefault(true),
		golog.WithSLogSetDefaultCtx(true),
	)
}

// InitEntSLog 初始化 ent 日志
func InitEntSLog(name string) {
	prefix := "ent"
	global.GvaEntSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
	global.GvaEntLogSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix+"_log", name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitGinSLog 初始化 gin 日志
func InitGinSLog(name string) {
	prefix := "gin"
	global.GvaGinSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitGrpcSLog 初始化 gRPC 日志
func InitGrpcSLog(name string) {
	prefix := "grpc"
	global.GvaGrpcSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitThriftSLog 初始化 Thrift 日志
func InitThriftSLog(name string) {
	prefix := "thrift"
	global.GvaThriftSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitGocronSLog 初始化 gocron 日志
func InitGocronSLog(name string) {
	prefix := "gocron"
	global.GvaGocronSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitMcpSLog 初始化 mcp 日志
func InitMcpSLog(name string) {
	prefix := "mcp"
	global.GvaMcpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitRestySLog 初始化 resty 日志
func InitRestySLog(name string) {
	prefix := "resty"
	global.GvaRestySlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}

// InitHttpSLog 初始化 http 日志
func InitHttpSLog(name string) {
	prefix := "http"
	global.GvaHttpSlog = golog.NewSlog(
		golog.WithSLogLumberjack(getLumberjack(prefix, name)),
		golog.WithSLogSetLevel(getLevel()),
	)
}
