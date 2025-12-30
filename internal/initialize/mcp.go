package initialize

import (
	"context"
	"log/slog"
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 初始化 MCP
func InitMCP(ctx context.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 创建 Model Context Protocol 服务器
	impl := mcp.Implementation{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.GvaVersion,
	}
	options := mcp.ServerOptions{}
	options.Logger = logger
	global.GvaMcpServer = mcp.NewServer(&impl, &options)
}
