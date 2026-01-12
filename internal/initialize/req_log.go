package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"

	"github.com/imroc/req/v3"
	"go.dtapp.net/library/contrib/req_log"
)

// 实现 LogSaver 接口
type EntReqLogSaver struct {
	logger *slog.Logger
	db     *ent_log.Client
}

func (s *EntReqLogSaver) HandleLog(ctx context.Context, data *req_log.LogData) error {
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

	go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *req_log.LogData) {
		err := asyncDb.ReqLog.Create().
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

// InitEntReqLog 初始化 ent req 日志
func InitEntReqLog(logger *slog.Logger) {

	// 定义回调函数
	callbackFunc := func(ctx context.Context, data *req_log.LogData) error {
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

		go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *req_log.LogData) {
			err := asyncDb.ReqLog.Create().
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
	saver := &EntReqLogSaver{
		logger: logger,
	}
	if global.GvaEntLogDb != nil {
		saver.db = global.GvaEntLogDb
	}

	// 初始化 Logger，传入回调和接口（回调优先）
	global.GvaReqLog = req_log.NewLoggerMiddleware(saver, callbackFunc)

	// 创建 Req 客户端，并带日志记录
	global.GvaReqClientWithLog = req.NewClient()
	// 请求中间件
	global.GvaReqClientWithLog.OnBeforeRequest(global.GvaReqLog.OnBeforeRequest)
	// 响应中间件
	global.GvaReqClientWithLog.OnAfterResponse(global.GvaReqLog.OnAfterResponse)
}
