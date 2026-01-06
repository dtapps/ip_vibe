package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"

	"go.dtapp.net/library/contrib/resty_log"
)

// 实现 LogSaver 接口
type EntRestyLogSaver struct {
	logger *slog.Logger
	db     *ent_log.Client
}

func (s *EntRestyLogSaver) SaveLog(ctx context.Context, data *resty_log.LogData) error {
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

	go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *resty_log.LogData) {
		err := asyncDb.RestyLog.Create().
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
			SetRestyVersion(syncData.RestyVersion).
			Exec(asyncCtx)
		if err != nil {
			s.logger.ErrorContext(asyncCtx, "[接口方式]执行保存日志失败",
				slog.Any("err", err),
			)
		}
	}(context.WithoutCancel(ctx), s.db, data)

	return nil
}

// InitEntRestyLog 初始化 ent resty 日志
func InitEntRestyLog(ctx context.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 定义回调函数
	saveFunc := func(ctx context.Context, data *resty_log.LogData) error {
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

		go func(asyncCtx context.Context, asyncDb *ent_log.Client, syncData *resty_log.LogData) {
			err := asyncDb.RestyLog.Create().
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
				SetRestyVersion(syncData.RestyVersion).
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
	saver := &EntRestyLogSaver{
		logger: logger,
		db:     global.GvaEntLogDb,
	}

	// 初始化 Logger，传入回调和接口（回调优先）
	global.GvaRestyLog = resty_log.NewLogger(saveFunc, saver, false)

}
