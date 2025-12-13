package grpc_gateway

import (
	"log"
	"net/http"

	"vibe_ip/internal/grpc_server"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Gateway struct {
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewGateway() *Gateway {
	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer()
	openServer := grpc_server.NewOpenServer()
	openServer.RegisterService(grpcServer)

	return &Gateway{
		grpcServer: grpcServer,
	}
}

// RegisterGinRoutes 在 Gin 中注册 gRPC 网关路由
func (g *Gateway) RegisterGinRoutes(router *gin.Engine) {
	// 创建 gRPC 网关 mux
	gwmux := runtime.NewServeMux()

	// 注册 gRPC 服务到网关
	// opts := []grpc.DialOption{grpc.WithInsecure()}

	// err := grpc_generated.RegisterOpenServiceHandlerFromEndpoint(ctx, gwmux, "localhost:8080", opts)
	// if err != nil {
	// 	panic(err)
	// }

	// 将 gRPC 网关挂载到 Gin
	router.Any("/grpc/*any", gin.WrapH(gwmux))

	log.Println("gRPC gateway registered at /grpc")
}
