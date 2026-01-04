package schemahook

import (
	"context"
	"vibe_ip/internal/ent_log/operationlog"
)

type sourceKey struct{}

// WithSource 指定操作来源并返回新 context
func WithSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sourceKey{}, source)
}

// GetSource 从 context 中获取来源，默认为 user
func GetSource(ctx context.Context) string {
	if v := ctx.Value(sourceKey{}); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return operationlog.SourceUser.String()
}
