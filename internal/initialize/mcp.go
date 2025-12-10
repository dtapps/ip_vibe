package initialize

import (
	"vibe_ip/internal/global"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 初始化 MCP
func InitMCP() {
	// 创建 Model Context Protocol 服务器
	global.GvaMcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.GvaVersion,
	}, nil)
}
