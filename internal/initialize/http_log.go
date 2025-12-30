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
	logger *slog.Logger
	db     *ent_log.Client
}

func (s *EntHttpLogSaver) HandleLog(ctx context.Context, data *http_log.LogData) error {
	s.logger.DebugContext(ctx, "[接口方式]准备保存日志",
		slog.String("Hostname", data.Hostname),
		slog.String("Method", data.Method),
		slog.String("URL", data.URL),
	)
	if s.db == nil || data == nil {
		return nil
	}
	s.logger.DebugContext(ctx, "[接口方式]执行保存日志",
		slog.String("Hostname", data.Hostname),
		slog.String("Method", data.Method),
		slog.String("URL", data.URL),
		slog.Int("StatusCode", data.StatusCode),
		slog.Int64("ElapseTime", data.ElapseTime),
		slog.Bool("IsError", data.IsError),
	)

	go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
		var (
			requestBody  = []byte{}
			responseBody = []byte{}
		)
		if len(data.RequestBodyJSON) > 0 {
			requestBody = data.RequestBodyJSON
		}
		if len(data.RequestBodyXML) > 0 {
			requestBody = data.RequestBodyXML
		}
		if len(data.ResponseBodyJSON) > 0 {
			responseBody = data.ResponseBodyJSON
		}
		if len(data.ResponseBodyXML) > 0 {
			responseBody = data.ResponseBodyXML
		}
		err := asyncDb.HttpLog.Create().
			SetHostname(data.Hostname).
			SetMethod(data.Method).
			SetURL(data.URL).
			SetRequestHeaders(data.RequestHeaders).
			SetRequestBody(requestBody).
			SetStatusCode(data.StatusCode).
			SetResponseHeaders(data.ResponseHeaders).
			SetResponseBody(responseBody).
			SetElapseTime(data.ElapseTime).
			SetIsError(data.IsError).
			SetGoVersion(data.GoVersion).
			SetPluginVersion(data.PluginVersion).
			Exec(asyncCtx)
		if err != nil {
			s.logger.ErrorContext(ctx, "[接口方式]执行保存日志失败",
				slog.Any("err", err),
			)
		}
	}(ctx, s.db)

	return nil
}

func InitEntHttpLog(logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	httpLogMutex.Do(func() {
		originalTransport := http.DefaultTransport

		// 定义回调函数
		saveFunc := func(ctx context.Context, data *http_log.LogData) error {
			logger.DebugContext(ctx, "[回调方式]准备保存日志",
				slog.String("Hostname", data.Hostname),
				slog.String("Method", data.Method),
				slog.String("URL", data.URL),
			)
			if global.GvaEntLogDb == nil || data == nil {
				return nil
			}
			logger.DebugContext(ctx, "[回调方式]执行保存日志",
				slog.String("Hostname", data.Hostname),
				slog.String("Method", data.Method),
				slog.String("URL", data.URL),
				slog.Int("StatusCode", data.StatusCode),
				slog.Int64("ElapseTime", data.ElapseTime),
				slog.Bool("IsError", data.IsError),
			)

			go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
				var (
					requestBody  = []byte{}
					responseBody = []byte{}
				)
				if len(data.RequestBodyJSON) > 0 {
					requestBody = data.RequestBodyJSON
				}
				if len(data.RequestBodyXML) > 0 {
					requestBody = data.RequestBodyXML
				}
				if len(data.ResponseBodyJSON) > 0 {
					responseBody = data.ResponseBodyJSON
				}
				if len(data.ResponseBodyXML) > 0 {
					responseBody = data.ResponseBodyXML
				}
				err := asyncDb.HttpLog.Create().
					SetHostname(data.Hostname).
					SetMethod(data.Method).
					SetURL(data.URL).
					SetRequestHeaders(data.RequestHeaders).
					SetRequestBody(requestBody).
					SetStatusCode(data.StatusCode).
					SetResponseHeaders(data.ResponseHeaders).
					SetResponseBody(responseBody).
					SetElapseTime(data.ElapseTime).
					SetIsError(data.IsError).
					SetGoVersion(data.GoVersion).
					SetPluginVersion(data.PluginVersion).
					Exec(asyncCtx)
				if err != nil {
					logger.ErrorContext(ctx, "[回调方式]执行保存日志失败",
						slog.Any("err", err),
					)
				}
			}(ctx, global.GvaEntLogDb)

			return nil
		}

		// 创建 LogSaver 实现
		saver := &EntHttpLogSaver{
			logger: logger,
			db:     global.GvaEntLogDb,
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
