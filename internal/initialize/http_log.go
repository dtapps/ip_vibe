package initialize

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/contrib/http_log"
)

var httpLogMutex sync.Once

// 实现 LogSaver 接口
type EntHttpLogSaver struct {
	db *ent_log.Client
}

func (s *EntHttpLogSaver) HandleLog(ctx context.Context, data *http_log.LogData) error {
	if global.GvaEnv.DEBUG {
		slog.InfoContext(ctx, "[接口方式]准备保存日志",
			slog.String("Hostname", data.Hostname),
			slog.String("Method", data.Method),
			slog.String("URL", data.URL),
		)
	}
	if s.db == nil || data == nil {
		return nil
	}
	if global.GvaEnv.DEBUG {
		slog.InfoContext(ctx, "[接口方式]执行保存日志",
			slog.String("Hostname", data.Hostname),
			slog.String("Method", data.Method),
			slog.String("URL", data.URL),
			slog.Int("StatusCode", data.StatusCode),
			slog.Int64("ElapseTime", data.ElapseTime),
			slog.Bool("IsError", data.IsError),
		)
	}

	go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
		_ = asyncDb.HttpLog.Create().
			SetHostname(data.Hostname).
			SetMethod(data.Method).
			SetURL(data.URL).
			SetRequestHeaders(data.RequestHeaders).
			SetRequestBody(data.RequestBody).
			SetStatusCode(data.StatusCode).
			SetResponseHeaders(data.ResponseHeaders).
			SetResponseBody(data.ResponseBody).
			SetElapseTime(data.ElapseTime).
			SetIsError(data.IsError).
			SetGoVersion(data.GoVersion).
			SetPluginVersion(data.PluginVersion).
			Exec(asyncCtx)
	}(ctx, s.db)

	return nil
}

func InitEntHttpLog() {
	httpLogMutex.Do(func() {
		originalTransport := http.DefaultTransport

		// 定义回调函数
		saveFunc := func(ctx context.Context, data *http_log.LogData) error {
			if global.GvaEnv.DEBUG {
				slog.InfoContext(ctx, "[回调方式]准备保存日志",
					slog.String("Hostname", data.Hostname),
					slog.String("Method", data.Method),
					slog.String("URL", data.URL),
				)
			}
			if global.GvaEntLogDb == nil || data == nil {
				return nil
			}
			if global.GvaEnv.DEBUG {
				slog.InfoContext(ctx, "[回调方式]执行保存日志",
					slog.String("Hostname", data.Hostname),
					slog.String("Method", data.Method),
					slog.String("URL", data.URL),
					slog.Int("StatusCode", data.StatusCode),
					slog.Int64("ElapseTime", data.ElapseTime),
					slog.Bool("IsError", data.IsError),
				)
			}

			go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
				_ = asyncDb.HttpLog.Create().
					SetHostname(data.Hostname).
					SetMethod(data.Method).
					SetURL(data.URL).
					SetRequestHeaders(data.RequestHeaders).
					SetRequestBody(data.RequestBody).
					SetStatusCode(data.StatusCode).
					SetResponseHeaders(data.ResponseHeaders).
					SetResponseBody(data.ResponseBody).
					SetElapseTime(data.ElapseTime).
					SetIsError(data.IsError).
					SetGoVersion(data.GoVersion).
					SetPluginVersion(data.PluginVersion).
					Exec(asyncCtx)
			}(ctx, global.GvaEntLogDb)

			return nil
		}

		// 创建 LogSaver 实现
		saver := &EntHttpLogSaver{
			db: global.GvaEntLogDb,
		}

		// 初始化 Logger，传入回调和接口（回调优先）
		global.GvaHttpLog = &http_log.LoggingTransport{
			Proxied: originalTransport,
			OnLog:   saveFunc,
			Handler: saver,
		}

		// 设置全局默认的 Transport
		http.DefaultTransport = global.GvaHttpLog

		// 设置全局 DefaultClient
		http.DefaultClient = &http.Client{
			Transport: global.GvaHttpLog,
		}
	})
}
