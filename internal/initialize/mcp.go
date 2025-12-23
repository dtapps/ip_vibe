package initialize

import (
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 初始化 MCP
func InitMCP() {
	// 创建 Model Context Protocol 服务器
	impl := mcp.Implementation{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.GvaVersion,
	}
	options := mcp.ServerOptions{}
	if global.GvaMcpSlog != nil {
		options.Logger = global.GvaMcpSlog.WithLogger()
	}
	global.GvaMcpServer = mcp.NewServer(&impl, &options)
}
