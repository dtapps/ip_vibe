package initialize

import (
	"path/filepath"

	"go.dtapp.net/library/utils/golog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitSLog 初始化 日志
func InitSLog(name string) {
	golog.NewSlog(
		golog.WithSLogLumberjack(&lumberjack.Logger{
			Filename:   filepath.Join("./logs/slog_" + name + ".log"), // 日志文件的位置
			MaxSize:    10,                                            // 每个日志文件最大尺寸（MB）
			MaxAge:     15,                                            // 留旧文件的最大天数
			MaxBackups: 10 * 30,                                       // 保留的最大旧文件数量
			LocalTime:  true,                                          // 使用本地时间创建时间戳
			Compress:   true,                                          // 是否压缩/归档旧文件
		}),
		golog.WithSLogSetDefault(),
		golog.WithSLogSetDefaultCtx(),
	)
}
