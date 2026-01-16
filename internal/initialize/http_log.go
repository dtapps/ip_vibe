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

	go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *http_log.LogData) {
		err := asyncDb.HttpLog.Create().
			SetHostname(syncData.Hostname).
			SetMethod(syncData.Method).
			SetURL(syncData.URL).
			SetRequestHeaders(syncData.RequestHeaders).
			SetRequestBody(syncData.RequestBody).
			SetStatusCode(syncData.StatusCode).
			SetResponseHeaders(syncData.ResponseHeaders).
			SetResponseBody(syncData.ResponseBody).
			SetElapseTime(syncData.ElapseTime).
			SetIsError(syncData.IsError).
			SetGoVersion(syncData.GoVersion).
			SetPluginVersion(syncData.PluginVersion).
			Exec(asyncCtx)
		if err != nil {
			s.logger.ErrorContext(asyncCtx, "[接口方式]执行保存日志失败",
				slog.Any("err", err),
			)
		}
	}(context.WithoutCancel(ctx), s.db, data)

	return nil
}

var httpLogMutex sync.Once

// InitEntHttpLog 初始化 ent http 日志
func InitEntHttpLog(logger *slog.Logger) {

	httpLogMutex.Do(func() {
		originalTransport := http.DefaultTransport

		// 定义回调函数
		callbackFunc := func(ctx context.Context, data *http_log.LogData) error {
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

			go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *http_log.LogData) {
				err := asyncDb.HttpLog.Create().
					SetHostname(syncData.Hostname).
					SetMethod(syncData.Method).
					SetURL(syncData.URL).
					SetRequestHeaders(syncData.RequestHeaders).
					SetRequestBody(syncData.RequestBody).
					SetStatusCode(syncData.StatusCode).
					SetResponseHeaders(syncData.ResponseHeaders).
					SetResponseBody(syncData.ResponseBody).
					SetElapseTime(syncData.ElapseTime).
					SetIsError(syncData.IsError).
					SetGoVersion(syncData.GoVersion).
					SetPluginVersion(syncData.PluginVersion).
					Exec(asyncCtx)
				if err != nil {
					logger.ErrorContext(asyncCtx, "[回调方式]执行保存日志失败",
						slog.Any("err", err),
					)
				}
			}(context.WithoutCancel(ctx), global.GvaEntLogDb, data)

			return nil
		}

		// 创建 LogSaver 实现
		saver := &EntHttpLogSaver{
			logger: logger,
		}
		if global.GvaEntLogDb != nil {
			saver.db = global.GvaEntLogDb
		}

		// 初始化 Logger，传入回调和接口（回调优先）
		global.GvaHttpLog = http_log.NewLoggingRoundTripper(originalTransport, saver, callbackFunc)

		// 设置全局默认的 Transport
		http.DefaultTransport = global.GvaHttpLog

		// 设置全局 DefaultClient
		http.DefaultClient = &http.Client{
			Transport: global.GvaHttpLog,
		}

		// // 创建 Http 客户端，并带日志记录
		// clone := global.GvaHttpLog.Clone()
		// global.GvaHttpClientWithLog = &http.Client{
		// 	Transport: clone,
		// }

		// // 创建 Http 客户端，并带日志记录不包含请求体/响应体
		// cloneNoBody := global.GvaHttpLog.CloneNoBody()
		// global.GvaHttpClientWithLogNoBody = &http.Client{
		// 	Transport: cloneNoBody,
		// }
	})
}
