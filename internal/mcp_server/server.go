package mcp_server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/mcp_prompt"
	"vibe_ip/internal/mcp_resource"
	"vibe_ip/internal/mcp_tool"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config MCP 服务所需的配置参数
type Config struct {
	Name    string
	Title   string
	Version string
}

// MCPServerManager MCP 服务管理器
type MCPServerManager struct {
	server          *mcp.Server  // MCP 服务器实例
	logger          *slog.Logger // 日志记录器
	config          Config       // MCP 服务配置参数
	ginRouter       *gin.Engine  // Gin 路由
	ginRouterPrefix string       // Gin 路由前缀
}

var (
	instance *MCPServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config, ginRouter *gin.Engine) *MCPServerManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &MCPServerManager{
			logger:          logger,
			config:          cfg,
			ginRouter:       ginRouter,
			ginRouterPrefix: "/mcp",
		}
	})
	return instance
}

// GetServer 获取底层的 mcp.Server 服务实例
func (m *MCPServerManager) GetServer() *mcp.Server {
	if m.server == nil {
		return nil
	}
	return m.server
}

// Run 启动 MCP 服务
func (m *MCPServerManager) Run(ctx context.Context) {

	// 创建 Model Context Protocol 服务器
	impl := mcp.Implementation{
		Name:    m.config.Name,
		Title:   m.config.Title,
		Version: m.config.Version,
	}
	options := mcp.ServerOptions{
		Logger: m.logger,
	}
	m.server = mcp.NewServer(&impl, &options)

	m.logger.InfoContext(ctx, "MCP 服务启动成功")

	// 注册 MCP 实现
	mcp_resource.RegisterTimeResources(m.server)
	mcp_tool.RegisterTimeTools(m.server)

	mcp_prompt.RegisterPlatformConfigPrompts(m.server)
	mcp_resource.RegisterPlatformConfigResources(m.server)
	mcp_tool.RegisterPlatformConfigTools(m.server)

	mcp_prompt.RegisterCDNRecordPrompts(m.server)
	mcp_resource.RegisterCDNRecordResources(m.server)
	mcp_tool.RegisterCDNRecordTools(m.server)

	mcp_prompt.RegisterDNSRecordPrompts(m.server)
	mcp_resource.RegisterDNSRecordResources(m.server)
	mcp_tool.RegisterDNSRecordTools(m.server)

	mcp_prompt.RegisterDevicePrompts(m.server)
	mcp_resource.RegisterDeviceResources(m.server)
	mcp_tool.RegisterDeviceTools(m.server)

	mcp_prompt.RegisterIPGroupPrompts(m.server)
	mcp_resource.RegisterIPGroupResources(m.server)
	mcp_tool.RegisterIPGroupTools(m.server)

	mcp_prompt.RegisterPortForwardPrompts(m.server)
	mcp_resource.RegisterPortForwardResources(m.server)
	mcp_tool.RegisterPortForwardTools(m.server)

	mcp_prompt.RegisterDomainMonitorPrompts(m.server)
	mcp_resource.RegisterDomainMonitorResources(m.server)
	mcp_tool.RegisterDomainMonitorTools(m.server)

	mcp_prompt.RegisterWebhookPrompts(m.server)
	mcp_resource.RegisterWebhookResources(m.server)
	mcp_tool.RegisterWebhookTools(m.server)

	mcp_prompt.RegisterWebhookSubPrompts(m.server)
	mcp_resource.RegisterWebhookSubResources(m.server)
	mcp_tool.RegisterWebhookSubTools(m.server)

	// 创建 HTTP 处理程序
	httpTransport := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return m.server
		},
		nil,
	)

	// 挂载到路由
	if m.ginRouter != nil {
		mcpRouter := m.ginRouter.Group(m.ginRouterPrefix).
			Use(gin_middleware.AuditContextMiddleware(
				operationlog.SourceMcp,
			))

		// 挂载 HTTP 处理程序
		mcpRouter.GET("", gin.WrapH(http.StripPrefix(m.ginRouterPrefix, httpTransport)))
		mcpRouter.POST("", gin.WrapH(http.StripPrefix(m.ginRouterPrefix, httpTransport)))
		m.logger.InfoContext(ctx, "MCP HTTP 端点已启用",
			slog.String("endpoint", m.ginRouterPrefix),
			slog.String("methods", "GET, POST"),
		)

		mcpRouter.Any("/*any", func(c *gin.Context) {
			gin.WrapH(http.StripPrefix(m.ginRouterPrefix, httpTransport))(c)
		})
	}

}
