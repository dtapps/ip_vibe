package thrift_process

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

// GenericInterceptor 实现通用日志和耗时统计
type GenericInterceptor struct {
	thrift.TProcessor
	logger *slog.Logger
}

func NewGenericInterceptor(p thrift.TProcessor, logger *slog.Logger) thrift.TProcessor {
	return &GenericInterceptor{TProcessor: p, logger: logger}
}

func (p *GenericInterceptor) Process(ctx context.Context, in, out thrift.TProtocol) (bool, thrift.TException) {
	start := time.Now()

	// 执行业务
	success, err := p.TProcessor.Process(ctx, in, out)

	// 如果 err 是 EOF，说明是连接正常断开，不需要打 ERROR 日志
	if err != nil && err.Error() == io.EOF.Error() {
		return success, err
	}

	duration := time.Since(start)

	// 记录日志
	if err != nil {
		p.logger.ErrorContext(ctx, "Thrift 调用错误",
			slog.Duration("duration", duration),
			slog.Any("err", err),
		)
	} else {
		p.logger.InfoContext(ctx, "Thrift 调用完成",
			slog.Duration("duration", duration),
		)
	}

	return success, err
}
