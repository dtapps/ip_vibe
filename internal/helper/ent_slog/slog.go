package ent_slog

import (
	"fmt"
	"log/slog"
	"strings"
)

type SlogLogger struct {
	logger *slog.Logger
	name   string
}

func NewSlogLogger(logger *slog.Logger, name string) *SlogLogger {
	return &SlogLogger{logger: logger, name: name}
}

// Log 实现 ent 的日志回调接口：func(a ...any)
func (l *SlogLogger) Log(args ...any) {
	sqlStr := fmt.Sprint(args...)
	// 将 SQL 里的双引号替换为单引号，或者直接去掉，从而彻底避免 slog 的转义机制
	cleanSql := strings.ReplaceAll(sqlStr, "\"", "'")
	l.logger.Info(l.name, "sql", cleanSql)
}
