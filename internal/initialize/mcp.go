package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 初始化 MCP
func InitMCP(ctx context.Context, logger *slog.Logger) {

	logger.InfoContext(ctx, "MCP 服务正在启动")

	// 创建 Model Context Protocol 服务器
	impl := mcp.Implementation{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.GvaVersion,
	}
	options := mcp.ServerOptions{}
	options.Logger = logger
	global.GvaMcpServer = mcp.NewServer(&impl, &options)

	logger.InfoContext(ctx, "MCP 服务启动成功")

}
