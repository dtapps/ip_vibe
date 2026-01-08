package grpc_gateway_client

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Config gRPC 客户端所需的配置参数
type Config struct {
	Addr     string // 服务器地址
	Username string // 用户名
	Password string // 密码
}

// GRPCGatewayClientManager gRPC 网关客户端管理器
type GRPCGatewayClientManager struct {
	client    *grpc.ClientConn // gRPC 客户端连接
	logger    *slog.Logger     // 日志记录器
	config    Config           // gRPC 客户端配置参数
	ginRouter *gin.Engine      // Gin 路由
}

var (
	instance *GRPCGatewayClientManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config, ginRouter *gin.Engine) *GRPCGatewayClientManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &GRPCGatewayClientManager{
			logger:    logger,
			config:    cfg,
			ginRouter: ginRouter,
		}
	})
	return instance
}

// Run 启动 gRPC 网关客户端
func (m *GRPCGatewayClientManager) Run(ctx context.Context) (err error) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "gRPC-Gateway 客户端不启动，未配置地址")
		return nil
	}

	// 配置选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 鉴权
		grpc.WithPerRPCCredentials(grpc_credential.NewClientBasicAuth(
			m.logger,
			m.config.Username,
			m.config.Password,
		)),
		// 重试
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
	}

	// 连接 gRPC 服务器
	m.client, err = grpc.NewClient(m.config.Addr, opts...)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC-Gateway 客户端连接服务器失败",
			slog.Any("err", err),
		)
		return err
	}

	gwOpts := []runtime.ServeMuxOption{
		// 错误响应统一包装
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			// 默认错误处理器
			st, _ := status.FromError(err)

			// 映射 HTTP 状态码
			httpCode := runtime.HTTPStatusFromCode(st.Code())

			// 构造响应
			resp := handlers.APIResponse{
				Code:    int(st.Code()),
				Message: st.Message(),
				Data:    st.Details(),
			}

			// 设置响应头
			w.Header().Set("Content-Type", marshaler.ContentType(nil))

			// 设置 HTTP 状态码
			w.WriteHeader(httpCode)

			// 使用 JSON 编码写回
			_ = marshaler.NewEncoder(w).Encode(resp)
		}),
		// 成功响应统一包装
		runtime.WithForwardResponseRewriter(func(ctx context.Context, resp proto.Message) (any, error) {
			if _, ok := resp.(*statuspb.Status); ok {
				return resp, nil
			}
			var data any = resp
			// 判断是否是 Empty 类型，如果是，强行设为 nil 以触发 omitempty
			if _, ok := resp.(*emptypb.Empty); ok || resp == nil {
				data = nil
			}
			return handlers.APIResponse{
				Code:    http.StatusOK,
				Message: handlers.MessageSuccess,
				Data:    data,
			}, nil
		}),
	}

	gwMux := runtime.NewServeMux(gwOpts...)

	// 注册服务
	if err = grpc_generated.RegisterIndexServiceHandler(ctx, gwMux, m.client); err != nil {
		logger.ErrorContext(ctx, "注册 IndexService 网关路由失败",
			slog.Any("err", err),
		)
	}
	if err = grpc_generated.RegisterOpenServiceHandler(ctx, gwMux, m.client); err != nil {
		logger.ErrorContext(ctx, "注册 OpenService 网关路由失败",
			slog.Any("err", err),
		)
	}

	// 挂载网关
	if m.ginRouter != nil {
		grpcGwRouter := m.ginRouter.Group("/grpc").Use(gin_middleware.AuditContextMiddleware(
			m.logger,
			operationlog.SourceGrpc,
		))
		grpcGwRouter.Any("/*any", gin.WrapH(http.StripPrefix("/grpc", gwMux)))
		logger.InfoContext(ctx, "gRPC-Gateway 已挂载到 Gin 路由",
			slog.String("prefix", "/grpc"),
		)
	}

	return nil
}

// GetServer 获取底层的 grpc.Server 实例
// 注意：如果在 Run 之前调用，可能会返回 nil
func (m *GRPCGatewayClientManager) GetClient() *grpc.ClientConn {
	return m.client
}

// Stop 优雅关闭
func (m *GRPCGatewayClientManager) Stop(ctx context.Context) {
	if m.client != nil {
		m.logger.InfoContext(ctx, "gRPC-Gateway 客户端正在优雅关闭...")
		m.client.Close()
	}
	m.logger.InfoContext(ctx, "gRPC-Gateway 客户端已安全停止")
}
