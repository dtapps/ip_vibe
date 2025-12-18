package gin_slog

import (
	"log/slog"
	"strings"
)

type SlogWriter struct {
	logger *slog.Logger
}

func NewSlogWriter(logger *slog.Logger) *SlogWriter {
	return &SlogWriter{logger: logger}
}

func (sw *SlogWriter) Write(p []byte) (n int, err error) {
	// 直接在写入 slog 前，把 Gin 字符串里的引号替换掉
	msg := strings.TrimSuffix(string(p), "\n")
	// 移除所有双引号
	msg = strings.ReplaceAll(msg, "\"", "")
	sw.logger.Info(msg)
	return len(p), nil
}
