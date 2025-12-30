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

	go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
		err := asyncDb.RestyLog.Create().
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
			SetRestyVersion(data.RestyVersion).
			Exec(asyncCtx)
		if err != nil {
			s.logger.ErrorContext(ctx, "[接口方式]执行保存日志失败",
				slog.Any("err", err),
			)
		}
	}(ctx, s.db)

	return nil
}

func InitEntRestyLog(logger *slog.Logger) {
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

		go func(asyncCtx context.Context, asyncDb *ent_log.Client) {
			err := asyncDb.RestyLog.Create().
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
				SetRestyVersion(data.RestyVersion).
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
	saver := &EntRestyLogSaver{
		logger: logger,
		db:     global.GvaEntLogDb,
	}

	// 初始化 Logger，传入回调和接口（回调优先）
	global.GvaRestyLog = resty_log.NewLogger(saveFunc, saver, false)

}
