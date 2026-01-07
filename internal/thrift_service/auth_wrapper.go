package thrift_service

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/global"
	"vibe_ip/internal/thrift_generated/open" // 请根据实际路径修改
)

// AuthWrapper 包装原始 Handler，实现日志和认证的解耦
type AuthWrapper struct {
	validToken string
	logger     *slog.Logger
}

func NewAuthWrapper(logger *slog.Logger) *AuthWrapper {

	// 获取基本认证 token
	validToken := core.GetBasicAuthToken(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)

	return &AuthWrapper{
		logger:     logger,
		validToken: validToken,
	}
}

// check 统一处理日志打印和认证
func (w *AuthWrapper) check(ctx context.Context, base *open.Base, req any) (func(), error) {

	// 判断是否需要认证
	if w.validToken == "" {
		return func() {}, nil
	}

	start := time.Now() // 记录开始时间

	token := ""
	if base != nil {
		token = base.GetToken()
	}

	// 认证逻辑
	if token != w.validToken {
		w.logger.WarnContext(ctx, "🔒 认证失败")
		return nil, errors.New("unauthorized")
	}

	// 返回一个延迟执行的函数，用于统计耗时
	return func() {
		w.logger.InfoContext(ctx, "--- [Thrift Call Finished] ---",
			slog.Duration("cost", time.Since(start)), // 计算并打印耗时
			slog.Any("payload", req),
		)
	}, nil
}
