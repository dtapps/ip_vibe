package initialize

import (
	"path/filepath"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/utils/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitSLog 初始化 日志
func InitSLog(name string) {
	golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("logs", "slog_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                          // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                         // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                    // 保留的最大旧文件数量
			LocalTime:  true,                                       // 使用本地时间创建时间戳
			Compress:   true,                                       // 是否压缩/归档旧文件
		}),
		golog.WithSLogSetDefault(),
		golog.WithSLogSetDefaultCtx(),
	)
}

// InitEntSLog 初始化 ent 日志
func InitEntSLog(name string) {
	global.GvaEntSlog = golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("logs", "ent_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                         // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                        // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                   // 保留的最大旧文件数量
			LocalTime:  true,                                      // 使用本地时间创建时间戳
			Compress:   true,                                      // 是否压缩/归档旧文件
		}),
	)
}

// InitGinSLog 初始化 gin 日志
func InitGinSLog(name string) {
	global.GvaGinSlog = golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("logs", "gin_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                         // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                        // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                   // 保留的最大旧文件数量
			LocalTime:  true,                                      // 使用本地时间创建时间戳
			Compress:   true,                                      // 是否压缩/归档旧文件
		}),
	)
}

// InitGrpcSLog 初始化 gRPC 日志
func InitGrpcSLog(name string) {
	global.GvaGrpcSlog = golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("logs", "grpc_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                          // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                         // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                    // 保留的最大旧文件数量
			LocalTime:  true,                                       // 使用本地时间创建时间戳
			Compress:   true,                                       // 是否压缩/归档旧文件
		}),
	)
}

// InitGocronSLog 初始化 gocron 日志
func InitGocronSLog(name string) {
	global.GvaGocronSlog = golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("logs", "gocron_"+name+".log"), // 日志文件的位置
			MaxSize:    3,                                            // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                           // 留旧文件的最大天数
			MaxBackups: 10 * 15,                                      // 保留的最大旧文件数量
			LocalTime:  true,                                         // 使用本地时间创建时间戳
			Compress:   true,                                         // 是否压缩/归档旧文件
		}),
	)
}
