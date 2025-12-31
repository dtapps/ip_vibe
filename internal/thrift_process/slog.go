package thrift_process

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
)

// CapturingInput 包装输入协议，只重写 ReadMessageBegin 来捕获方法名
// 其他方法通过嵌入自动委托给原始 TProtocol
type CapturingInput struct {
	thrift.TProtocol
	methodName *string // 指针，用于存储捕获的方法名
}

func (ci *CapturingInput) ReadMessageBegin(ctx context.Context) (name string, typeId thrift.TMessageType, seqID int32, err error) {
	// 调用原始协议读取，并捕获 name（注意：传递 ctx）
	name, typeId, seqID, err = ci.TProtocol.ReadMessageBegin(ctx)
	if err == nil && ci.methodName != nil {
		*ci.methodName = name
	}
	return name, typeId, seqID, err
}

type LoggingProcessor struct {
	thrift.TProcessor
	logger *slog.Logger
}

func NewLoggingProcessor(p thrift.TProcessor, logger *slog.Logger) *LoggingProcessor {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingProcessor{
		TProcessor: p,
		logger:     logger,
	}
}

func (p *LoggingProcessor) Process(ctx context.Context, in, out thrift.TProtocol) (bool, thrift.TException) {
	// 记录开始时间
	start := time.Now()

	// 创建包装输入协议，捕获方法名
	var methodName string
	capturingIn := &CapturingInput{
		TProtocol:  in,
		methodName: &methodName,
	}

	// 调用原始处理器
	ok, err := p.TProcessor.Process(ctx, capturingIn, out)

	// 记录耗时
	dur := time.Since(start).Milliseconds()

	// 过滤掉空的 EOF 日志（这是 Thrift 服务器正常结束信号，不是错误调用）
	if methodName == "" && err != nil && errors.Is(err, io.EOF) {
		return ok, err // 不记录日志，直接返回
	}

	attrs := []slog.Attr{
		slog.String("method", methodName),
		slog.Int64("time_ms", dur),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("err", err))
	}

	p.logger.LogAttrs(ctx, slog.LevelInfo, "Thrift调用完成", attrs...)
	return ok, err
}
