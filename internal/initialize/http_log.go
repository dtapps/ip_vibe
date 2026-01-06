package initialize

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"

	xml2json "github.com/basgys/goxml2json"
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

	go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *http_log.LogData) {
		var (
			requestBody  json.RawMessage = nil
			responseBody json.RawMessage = nil
		)
		if len(syncData.RequestBodyJSON) > 0 {
			requestBody = json.RawMessage(syncData.RequestBodyJSON)
		}
		if len(syncData.RequestBodyXML) > 0 {
			xj, _ := xml2json.Convert(strings.NewReader(string(syncData.RequestBodyXML)))
			requestBody = json.RawMessage(xj.Bytes())
		}
		if len(syncData.ResponseBodyJSON) > 0 {
			responseBody = json.RawMessage(syncData.ResponseBodyJSON)
		}
		if len(syncData.ResponseBodyXML) > 0 {
			xj, _ := xml2json.Convert(strings.NewReader(string(syncData.ResponseBodyXML)))
			responseBody = json.RawMessage(xj.Bytes())
		}
		err := asyncDb.HttpLog.Create().
			SetHostname(syncData.Hostname).
			SetMethod(syncData.Method).
			SetURL(syncData.URL).
			SetRequestHeaders(syncData.RequestHeaders).
			SetRequestBody(requestBody).
			SetStatusCode(syncData.StatusCode).
			SetResponseHeaders(syncData.ResponseHeaders).
			SetResponseBody(responseBody).
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

			go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *http_log.LogData) {
				var (
					requestBody  json.RawMessage = nil
					responseBody json.RawMessage = nil
				)
				if len(syncData.RequestBodyJSON) > 0 {
					requestBody = json.RawMessage(syncData.RequestBodyJSON)
				}
				if len(syncData.RequestBodyXML) > 0 {
					xj, _ := xml2json.Convert(strings.NewReader(string(syncData.RequestBodyXML)))
					requestBody = json.RawMessage(xj.Bytes())
				}
				if len(syncData.ResponseBodyJSON) > 0 {
					responseBody = json.RawMessage(syncData.ResponseBodyJSON)
				}
				if len(syncData.ResponseBodyXML) > 0 {
					xj, _ := xml2json.Convert(strings.NewReader(string(syncData.ResponseBodyXML)))
					responseBody = json.RawMessage(xj.Bytes())
				}
				err := asyncDb.HttpLog.Create().
					SetHostname(syncData.Hostname).
					SetMethod(syncData.Method).
					SetURL(syncData.URL).
					SetRequestHeaders(syncData.RequestHeaders).
					SetRequestBody(requestBody).
					SetStatusCode(syncData.StatusCode).
					SetResponseHeaders(syncData.ResponseHeaders).
					SetResponseBody(responseBody).
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
