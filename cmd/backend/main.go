package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/docs_2"
	"vibe_ip/internal/docs_3"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_client"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/grpc_interceptor"
	"vibe_ip/internal/grpc_service"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/helper/gin_slog"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/mcp_prompt"
	"vibe_ip/internal/mcp_resource"
	"vibe_ip/internal/mcp_tool"
	"vibe_ip/internal/thrift_generated/open"
	"vibe_ip/internal/thrift_process"
	"vibe_ip/internal/thrift_service"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/denisbrodbeck/machineid"
	"github.com/dromara/carbon/v2"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	grpcGatewayRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wdcbot/qingfeng"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 全局资源管理
var (
	grpcServerInstance   *grpc.Server
	grpcListenerInstance net.Listener
	thriftServerInstance *thrift.TSimpleServer
	thriftTransport      thrift.TServerTransport
	httpServerInstance   *http.Server
	grpcConnInstance     *grpc.ClientConn
	shutdownOnce         sync.Once
	shutdownChan         = make(chan struct{})
)

//go:embed templates static
var embeddedFiles embed.FS

var (
	version = "dev" // 版本号
	// version   = "v1.0.0"  // 版本号
	commit    = "unknown" // Git 提交 ID
	buildTime = "unknown" // 构建时间
)

// 初始化版本信息到全局变量
func init() {
	global.GvaVersion = version
	global.GvaGitCommit = commit
	global.GvaBuildTime = buildTime
	global.GvaGoVersion = runtime.Version()
	global.GvaOS = runtime.GOOS
	global.GvaArch = runtime.GOARCH
}

// @title			IPVibe
// @version		dev
// @description	IPVibe
// @host			localhost:9999
// @BasePath		/api/
func main() {

	// 全局 Panic 捕获
	defer func() {
		if r := recover(); r != nil {
			// 获取堆栈信息
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackInfo := string(buf[:n])

			// 格式化错误消息
			errMsg := fmt.Sprintf("%v", r)

			// 使用标准 slog 记录到控制台/文件（此时 global 变量可能不可靠，用标准 slog 最稳）
			slog.Error("🔥 系统触发致命崩溃 (Panic)",
				slog.String("stack", stackInfo),
				slog.String("err", errMsg),
			)

			// 给日志留一点写入时间
			time.Sleep(time.Second)
			os.Exit(1)
		}
	}()

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceSystem.String())

	initialize.InitSLog("backend")       // 初始化 日志
	initialize.InitEntSLog("backend")    // 初始化 ent 日志
	initialize.InitGinSLog("backend")    // 初始化 gin 日志
	initialize.InitGrpcSLog("backend")   // 初始化 gRPC 日志
	initialize.InitThriftSLog("backend") // 初始化 Thrift 日志
	initialize.InitGocronSLog("backend") // 初始化 gocron 日志
	initialize.InitMcpSLog("backend")    // 初始化 mcp 日志
	initialize.InitRestySLog("backend")  // 初始化 resty 日志
	initialize.InitHttpSLog("backend")   // 初始化 http 日志
	initialize.EnsureDefaultSLog()       // 确保日志不为 nil

	initialize.InitCoraza(ctx, global.GvaSlog) // 初始化 Coraza 防火墙

	initialize.InitEntDb(ctx, global.GvaEntSlog)       // 初始化 业务数据库
	initialize.InitEntLogDb(ctx, global.GvaEntLogSlog) // 初始化 日志数据库
	initialize.InitGoCron(ctx, global.GvaGocronSlog)   // 初始化 GoCron 调度器
	initialize.InitMCP(ctx, global.GvaMcpSlog)         // 初始化 MCP

	initialize.InitEntHttpLog(global.GvaHttpSlog)   // 初始化 HTTP 日志
	initialize.InitEntRestyLog(global.GvaRestySlog) // 初始化 Resty 日志

	initialize.InitThirdParty(ctx, global.GvaSlog) // 初始化 第三方

	initialize.InitOperationLogHelper() // 初始化操作日志助手

	// 打印
	slog.InfoContext(ctx, "服务启动",
		slog.Bool("debug", global.GvaEnv.DEBUG),
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
		slog.String("os", global.GvaOS),
		slog.String("arch", global.GvaArch),
	)

	if global.GvaGinSlog != nil {
		// 禁用控制台颜色
		gin.DisableConsoleColor()
		// 配置gin日志输出到slog
		gin.DefaultWriter = gin_slog.NewSlogWriter(global.GvaGinSlog)
		gin.DefaultErrorWriter = gin_slog.NewSlogWriter(global.GvaGinSlog)
	} else {
		// 强制控制台颜色
		gin.ForceConsoleColor()
	}

	// 配置调试模式
	if global.GvaEnv.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	// router := gin.Default()
	router := gin.New()

	// 配置 日志 中间件
	router.Use(gin.Logger())

	// 配置 恢复 中间件
	router.Use(gin_middleware.RecoveryMiddleware(global.GvaGinSlog, true))

	// 配置 审计上下文 中间件
	router.Use(gin_middleware.AuditContextMiddleware(
		global.GvaGinSlog,
		operationlog.SourceUser,
	))

	// 配置 请求ID 中间件
	router.Use(gin_middleware.RequestIDMiddleware(
		global.GvaGinSlog,
	))

	// 配置 Coraza 中间件
	router.Use(gin_middleware.CorazaMiddleware(
		global.GvaCoraza,
		global.GvaGinSlog,
	))

	// 配置 Gzip 压缩 中间件
	router.Use(gin_middleware.GzipMiddleware(
		global.GvaGinSlog,
	))

	// 配置 CORS 中间件
	router.Use(gin_middleware.CORSMiddleware(
		global.GvaGinSlog,
	))

	// health check
	router.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.APIResponse{
			Code:    http.StatusOK,
			Message: "ok",
		})
	})

	// Web 资源
	web_resource(ctx, global.GvaGinSlog, router)

	// 配置 BasicAuth 中间件
	router.Use(gin_middleware.BasicAuthMiddleware(
		global.GvaGinSlog,
		global.GvaEnv.BASIC_AUTH_USERNAME,
		global.GvaEnv.BASIC_AUTH_PASSWORD,
	))

	// Web 路由
	web_router(ctx, global.GvaGinSlog, router)

	// API 路由
	api_router(ctx, global.GvaGinSlog, router)

	// Swagger 路由
	swagger_router(ctx, global.GvaGinSlog, router)

	// 新 Swagger 路由
	qingfeng_router(ctx, global.GvaGinSlog, router)

	// MCP 启动
	mcp_run(ctx, global.GvaGinSlog, global.GvaMcpSlog, router)

	// gRPC 启动
	go grpc_run(ctx, global.GvaGrpcSlog)

	// gRPC-Gateway 启动
	go grpc_gw_run(ctx, global.GvaGrpcSlog, router)

	// Thrift 启动
	go thrift_run(ctx, global.GvaThriftSlog)

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 启动 CDN、DNS 刷新任务
	ipUpdate := task.NewIPUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := ipUpdate.StartCdnJob(ctx); err != nil {
		slog.Error("启动CDN、DNS刷新任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 SSL 刷新任务
	sslUpdate := task.NewSSLUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := sslUpdate.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动SSL刷新任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动域名监控任务
	domainMonitor := task.NewDomainMonitorTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := domainMonitor.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动域名监控任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 创建 HTTP 服务实例（用于优雅关闭）
	httpServerInstance = &http.Server{
		Addr:    global.GvaEnv.HTTP_ADDR,
		Handler: router,
	}

	// 启动 Gin 服务
	global.GvaGinSlog.Info("Gin 服务正在启动",
		slog.String("port", global.GvaEnv.HTTP_ADDR),
	)

	// 使用 ListenAndServe 在 goroutine 中运行，以便能够捕获关闭信号
	go func() {
		if err := httpServerInstance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			global.GvaGinSlog.Error("Gin 服务启动失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		global.GvaGinSlog.InfoContext(ctx, "Gin 服务已停止")
	}()

	global.GvaGinSlog.Info("Gin 服务启动成功",
		slog.String("port", global.GvaEnv.HTTP_ADDR),
	)

	// 等待中断信号进行优雅关闭
	waitForShutdown(global.GvaGinSlog, global.GvaGocronSlog, global.GvaEntSlog, global.GvaEntLogSlog)
}

// 优雅关闭函数
func waitForShutdown(ginLogger *slog.Logger, gocronLogger *slog.Logger, entLogger *slog.Logger, entLogLogger *slog.Logger) {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，开始优雅关闭...")

	// 执行关闭流程
	gracefulShutdown(ginLogger, gocronLogger, entLogger, entLogLogger)
}

// 优雅关闭实现
func gracefulShutdown(ginLogger *slog.Logger, gocronLogger *slog.Logger, entLogger *slog.Logger, entLogLogger *slog.Logger) {

	shutdownOnce.Do(func() {
		// 1. 首先关闭 HTTP 服务（停止接收新请求）
		ginLogger.Info("开始关闭 HTTP 服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if httpServerInstance != nil {
			if err := httpServerInstance.Shutdown(ctx); err != nil {
				ginLogger.Error("HTTP 服务关闭失败",
					slog.Any("err", err),
				)
			} else {
				ginLogger.Info("HTTP 服务已关闭")
			}
		}

		// 2. 关闭 gRPC-Gateway 到上游的连接
		slog.Info("开始关闭 gRPC-Gateway 连接...")
		if grpcConnInstance != nil {
			grpcConnInstance.Close()
			slog.Info("gRPC-Gateway 连接已关闭")
		}

		// 3. 通知 gRPC、Thrift 服务开始优雅关闭
		close(shutdownChan)

		// 4. 等待一段时间让 goroutine 完成
		time.Sleep(1 * time.Second)

		// 5. 关闭数据库连接和其他资源（现有的 defer 逻辑）
		slog.Info("开始关闭数据库连接...")

		// 6. 关闭 第三方服务
		// initialize.CloseThirdParty()

		// 7. 关闭 GoCron 调度器
		initialize.CloseGoCron(ctx, gocronLogger)

		// 8. 等待一段时间让 goroutine 完成
		time.Sleep(1 * time.Second)

		// 9. 关闭 业务数据库
		initialize.CloseEntDb(ctx, entLogger)

		// 10. 关闭 日志数据库
		initialize.CloseEntLogDb(ctx, entLogLogger)

		slog.Info("所有资源已优雅关闭")
	})
}

// 开发环境
func loadTemplates(ginLogger *slog.Logger, templatePath string, funcMap template.FuncMap) multitemplate.Renderer {

	r := multitemplate.NewRenderer()

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		ginLogger.Debug("开发环境",
			slog.String("layout", layout),
			slog.String("index", index),
		)
	}

	// 所有业务页面
	pages, err := filepath.Glob(filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		ginLogger.Error("开发环境",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := filepath.Glob(filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		ginLogger.Error("开发环境",
			slog.Any("err", err),
		)
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("开发环境",
				slog.String("components", name),
			)
		}
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := os.Stat(path)
		// if err != nil {
		// 	ginLogger.Error("开发环境",
		// 		slog.Any("err", err),
		// 	)
		// }
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("开发环境",
				slog.String("pages", name),
			)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html 和 _view.html 和 _filter.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_filter.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件
		files = append(files, p)             // 当前页面
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("开发环境",
				slog.String("page", strings.Join(files, " -> ")),
			)
		}
		r.AddFromFilesFuncs(
			name,
			funcMap,
			files...,
		)
	}

	return r
}

// 生产环境
func loadTemplatesFS(ginLogger *slog.Logger, efs embed.FS, funcMap template.FuncMap) multitemplate.Renderer {

	r := multitemplate.NewRenderer()

	templatePath := "templates"

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		ginLogger.Debug("生产环境",
			slog.String("layout", layout),
			slog.String("index", index),
		)
	}

	// 所有业务页面
	pages, err := fs.Glob(efs, filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		ginLogger.Error("生产环境",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := fs.Glob(efs, filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		ginLogger.Error("生产环境",
			slog.Any("err", err),
		)
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("生产环境",
				slog.String("components", name),
			)
		}
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := embeddedFiles.Open(path)
		// if err != nil {
		// 	ginLogger.Error("生产环境",
		// 		slog.Any("err", err),
		// 	)
		// }
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("生产环境",
				slog.String("pages", name),
			)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html 和 _view.html 和 _filter.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_filter.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件

		files = append(files, p) // 当前页面
		if global.GvaEnv.DEBUG {
			ginLogger.Debug("生产环境",
				slog.String("page", strings.Join(files, " -> ")),
			)
		}
		r.AddFromFSFuncs(
			name,
			funcMap,
			efs,
			files...,
		)
	}

	return r
}

// Web 资源
func web_resource(ctx context.Context, ginLogger *slog.Logger, router *gin.Engine) {

	// EnumItem 表示单个枚举项
	type EnumItem struct {
		Label string
		Value string
	}

	// 创建带 FuncMap 的 template
	funcMap := template.FuncMap{
		// 判断是否为调试模式
		"isDebug": func() bool {
			return global.GvaEnv.DEBUG
		},
		// 格式化时间为字符串，若为空则返回 "—"
		"formatTimeOrNil": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format(time.DateTime)
		},
		// 格式化时间为 "X 前" 格式，如 "2 小时前"
		"diffForHumans": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return carbon.NewCarbon(t).DiffForHumans()
		},
		// 获取当前设备 Code
		"selfDeviceCode": func() string {
			code, err := machineid.ID()
			if err != nil {
				ginLogger.Error("[selfDeviceCode] 获取当前设备 Code 失败",
					slog.Any("err", err),
				)
				return "—"
			}
			return code
		},
		// pagination 组件
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"div":   func(a, b int) int { return a / b },
		"toInt": func(i int64) int { return int(i) },
		// switch / icons / button 组件
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				err := fmt.Errorf("期望有偶数个参数")
				ginLogger.Error("[dict] 函数参数错误",
					slog.Any("err", err),
				)
				return nil, err
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					err := fmt.Errorf("键必须是字符串")
					ginLogger.Error("[dict] 函数参数错误",
						slog.Any("err", err),
					)
					return nil, err
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		// 获取字段的枚举项
		"fieldEnumMap": func(fieldType string) (enumItems []EnumItem) {
			switch fieldType {
			case "cdn_type":
				// CDN记录 记录类型
				for _, value := range schematype.GetCDNRecordTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetCDNRecordTypeDescriptions[value],
						Value: value,
					})
				}
			case "cdn_origin_protocol":
				// CDN记录 回源协议
				for _, value := range schematype.GetOriginProtocols {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetOriginProtocolDescriptions[value],
						Value: value,
					})
				}
			case "dns_type":
				// DNS记录 记录类型
				for _, value := range schematype.GetDNSRecordTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDNSRecordTypeDescriptions[value],
						Value: value,
					})
				}
			case "domain_monitor_check_protocol":
				// 域名监控 检查类型
				for _, value := range schematype.GetDomainMonitorCheckProtocols {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDomainMonitorCheckProtocolDescriptions[value],
						Value: value,
					})
				}
			case "domain_monitor_current_status":
				// 域名监控 当前运行状态
				for _, value := range schematype.GetDomainMonitorStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDomainMonitorStatusDescriptions[value],
						Value: value,
					})
				}
			case "platform_config_capability":
				// 平台配置 平台支持的服务能力
				for _, value := range schematype.GetConfigCapabilitys {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetConfigCapabilityDescriptions[value],
						Value: value,
					})
				}
			case "platform_config_platform":
				// 平台配置 平台类型
				for _, value := range schematype.GetConfigPlatforms {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetConfigPlatformDescriptions[value],
						Value: value,
					})
				}
			case "ssl_record_status":
				// SSL记录 证书状态
				for _, value := range schematype.GetSSLRecordStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordStatusDescriptions[value],
						Value: value,
					})
				}
			case "ssl_record_provider":
				// SSL记录 CA提供商
				for _, value := range schematype.GetSSLRecordProviders {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordProviderDescriptions[value],
						Value: value,
					})
				}
			case "webhook_platform":
				// 通知 通知平台类型
				for _, value := range schematype.GetWebhookPlatforms {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetWebhookPlatformDescriptions[value],
						Value: value,
					})
				}
			case "webhook_subscription_obj_type":
				// 通知订阅 对象类型
				for _, value := range schematype.GetWebhookSubscriptionObjTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetWebhookSubscriptionObjTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_source":
				// 操作日志 操作来源
				for _, value := range schematype.GetLogSources {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogSourceDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_op_type":
				// 操作日志 操作类型
				for _, value := range schematype.GetLogOpTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogOpTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_obj_type":
				// 操作日志 对象类型
				for _, value := range schematype.GetLogObjTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogObjTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_status":
				// 操作日志 状态
				for _, value := range schematype.GetLogStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogStatusDescriptions[value],
						Value: value,
					})
				}
			case "ssl_certificate_history_provider":
				// SSL证书申请历史记录 CA提供商
				for _, value := range schematype.GetSSLRecordProviders {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordProviderDescriptions[value],
						Value: value,
					})
				}
			}
			return enumItems
		},
		// 将多个字符串参数拼接为一个字符串
		"join": func(s ...string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.Join(s, "")
		},
		// 将字符串按指定分隔符分割为字符串切片
		"slice": func(s ...string) []string {
			if len(s) == 0 {
				return []string{}
			}
			return s
		},
		// 将字符串按指定分隔符分割为字符串切片
		"split": func(s, sep string) []string {
			if s == "" || sep == "" {
				return []string{}
			}
			return strings.Split(s, sep)
		},
		// 判断字符串是否包含指定子字符串
		"contains": func(s, substr string) bool {
			if s == "" || substr == "" {
				return false
			}
			return strings.Contains(s, substr)
		},
	}

	// 判断是否为开发模式：version == "dev"
	devMode := global.GvaVersion == "dev"
	if devMode {
		// 自定义分隔符
		// router.Delims("{{", "}}")

		// 自定义模板功能
		// router.SetFuncMap(funcMap)

		// 开发：直接读本地 templates
		// router.LoadHTMLGlob("cmd/backend/templates/**/*")

		// 开发环境
		router.HTMLRender = loadTemplates(ginLogger, "cmd/backend/templates", funcMap)

		// 静态文件服务 - 开发：直接读本地 static
		router.StaticFS("/static", http.Dir("cmd/backend/static"))
	} else {
		// 生产：使用 embed
		// templ := template.Must(
		// 	template.New("").
		// 		Funcs(funcMap).
		// 		ParseFS(embeddedFiles, "templates/**/*"),
		// )
		// router.SetHTMLTemplate(templ)

		// 生产环境
		router.HTMLRender = loadTemplatesFS(ginLogger, embeddedFiles, funcMap)

		// 静态文件服务 - 使用嵌入的文件系统
		staticFS, _ := fs.Sub(embeddedFiles, "static")
		router.StaticFS("/static", http.FS(staticFS))
	}
}

// Web 路由
func web_router(ctx context.Context, ginLogger *slog.Logger, router *gin.Engine) {

	// 根路径 - 提供HTML页面并渲染数据
	router.GET("/", handlers.Dashboard)
	router.GET("/dashboard", handlers.Dashboard)

	// 平台配置页面 - 提供HTML页面并渲染数据
	router.GET("/platform-config", handlers.PlatformConfig)

	// CDN页面 - 提供HTML页面并渲染数据
	router.GET("/cdn", handlers.Cdn)

	// DNS页面 - 提供HTML页面并渲染数据
	router.GET("/dns", handlers.Dns)

	// 设备页面 - 提供HTML页面并渲染数据
	router.GET("/device", handlers.Device)
	router.GET("/device/:id/port-forward/stats", handlers.DevicePortForwardStats)

	// IP组页面 - 提供HTML页面并渲染数据
	router.GET("/ip-group", handlers.IPGroup)

	// 端口转发页面 - 提供HTML页面并渲染数据
	router.GET("/port-forward", handlers.PortForward)
	router.GET("/port-forward/:id/stats", handlers.PortForwardStats)

	// SSL页面 - 提供HTML页面并渲染数据
	router.GET("/ssl", handlers.Ssl)
	router.GET("/ssl/:id/detail", handlers.SslDetail)
	router.GET("/ssl/:id/history", handlers.SslHistory)

	// 域名监控页面 - 提供HTML页面并渲染数据
	router.GET("/domain-monitor", handlers.DomainMonitor)
	router.GET("/domain-monitor/:id/detail", handlers.DomainMonitorDetail)
	router.GET("/domain-monitor/:id/history", handlers.DomainMonitorHistory)

	// 通知页面 - 提供HTML页面并渲染数据
	router.GET("/webhook", handlers.Webhook)

	// 通知订阅页面 - 提供HTML页面并渲染数据
	router.GET("/webhook-sub", handlers.WebhookSub)

	// 任务页面 - 提供HTML页面并渲染数据
	router.GET("/task", handlers.Task)

	// 日志页面 - 提供HTML页面并渲染数据
	router.GET("/log", handlers.Log)

	// 接口页面 - 提供HTML页面并渲染数据
	router.GET("/interface", handlers.Interface)

	// 日志内容页面 - 提供HTML页面并渲染数据
	router.GET("/log-content", handlers.LogContent)

	// 未匹配到任何路由
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handlers.APIResponse{
			Code:    http.StatusNotFound,
			Message: "path not found",
		})
	})

	// 请求方法不被允许
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, handlers.APIResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: "method not allowed",
		})
	})
}

// API 路由
func api_router(ctx context.Context, ginLogger *slog.Logger, router *gin.Engine) {

	// API路由组
	api := router.Group("/api")
	{
		api.Any("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, handlers.APIResponse{
				Code:    http.StatusOK,
				Message: "api ok",
			})
		})

		// 平台配置相关API
		platformConfigs := api.Group("/platform-configs")
		{
			platformConfigs.GET("", handlers.GetPlatformConfigs)          // 获取全部平台配置
			platformConfigs.GET("/:id", handlers.GetPlatformConfigDetail) // 获取平台配置详情
			platformConfigs.POST("", handlers.CreatePlatformConfig)       // 创建平台配置
			platformConfigs.PATCH("/:id", handlers.UpdatePlatformConfig)  // 更新平台配置
			platformConfigs.DELETE("/:id", handlers.DeletePlatformConfig) // 删除平台配置
		}

		// CDN记录管理相关API
		cdnRecords := api.Group("/cdn-records")
		{
			cdnRecords.GET("", handlers.GetCdnRecords)                         // 获取全部CDN记录
			cdnRecords.GET("/:id", handlers.GetCdnRecordDetail)                // 获取CDN记录详情
			cdnRecords.POST("", handlers.CreateCdnRecord)                      // 创建CDN记录
			cdnRecords.PATCH("/:id", handlers.UpdateCdnRecord)                 // 更新CDN记录
			cdnRecords.PATCH("/:id/active", handlers.UpdateCdnRecordActive)    // 更新CDN记录启用状态
			cdnRecords.DELETE("/:id", handlers.DeleteCdnRecord)                // 删除CDN记录
			cdnRecords.GET("/:id/health-check", handlers.HealthCheckCdnRecord) // 检查CDN记录健康状态
		}

		// DNS记录管理相关API
		dnsRecords := api.Group("/dns-records")
		{
			dnsRecords.GET("", handlers.GetDnsRecords)                         // 获取全部DNS记录
			dnsRecords.GET("/:id", handlers.GetDnsRecordDetail)                // 获取DNS记录详情
			dnsRecords.POST("", handlers.CreateDnsRecord)                      // 创建DNS记录
			dnsRecords.PATCH("/:id", handlers.UpdateDnsRecord)                 // 更新DNS记录
			dnsRecords.PATCH("/:id/active", handlers.UpdateDnsRecordActive)    // 更新DNS记录启用状态
			dnsRecords.DELETE("/:id", handlers.DeleteDnsRecord)                // 删除DNS记录
			dnsRecords.GET("/:id/health-check", handlers.HealthCheckDnsRecord) // 检查DNS记录健康状态
		}

		// 设备相关API
		devices := api.Group("/devices")
		{
			devices.GET("", handlers.GetDevices)                                       // 获取全部设备
			devices.GET("/:id", handlers.GetDeviceDetail)                              // 获取设备详情
			devices.POST("", handlers.CreateDevice)                                    // 创建设备
			devices.PATCH("/:id", handlers.UpdateDevice)                               // 更新设备
			devices.DELETE("/:id", handlers.DeleteDevice)                              // 删除设备
			devices.GET("/:id/ip-history", handlers.GetDeviceIpHistory)                // 获取设备IP历史记录
			devices.GET("/:id/health-check", handlers.HealthCheckDevice)               // 检查设备健康状态
			devices.GET("/:id/port-forward/stats", handlers.GetDevicePortForwardStats) // 获取设备端口转发统计数据
		}

		// IP组相关API
		ipGroups := api.Group("/ip-groups")
		{
			ipGroups.GET("", handlers.GetIPGroups)                         // 获取全部IP组
			ipGroups.GET("/:id", handlers.GetIPGroupDetail)                // 获取IP组详情
			ipGroups.POST("", handlers.CreateIPGroup)                      // 创建IP组
			ipGroups.PATCH("/:id", handlers.UpdateIPGroup)                 // 更新IP组
			ipGroups.DELETE("/:id", handlers.DeleteIPGroup)                // 删除IP组
			ipGroups.GET("/:id/devices", handlers.GetIPGroupDevices)       // 获取IP组设备
			ipGroups.GET("/:id/health-check", handlers.HealthCheckIPGroup) // 检查IP组健康状态
		}

		// 端口转发相关API
		portForwards := api.Group("/port-forwards")
		{
			portForwards.GET("", handlers.GetPortForwards)                         // 获取全部端口转发
			portForwards.GET("/:id", handlers.GetPortForwardDetail)                // 获取端口转发详情
			portForwards.POST("", handlers.CreatePortForward)                      // 创建端口转发
			portForwards.PATCH("/:id", handlers.UpdatePortForward)                 // 更新端口转发
			portForwards.PATCH("/:id/forward", handlers.UpdatePortForwardForward)  // 更新端口转发防火墙状态
			portForwards.PATCH("/:id/active", handlers.UpdatePortForwardActive)    // 更新端口转发启用状态
			portForwards.DELETE("/:id", handlers.DeletePortForward)                // 删除端口转发
			portForwards.GET("/:id/health-check", handlers.HealthCheckPortForward) // 检查端口转发健康状态
			portForwards.GET("/:id/stats", handlers.GetPortForwardStats)           // 获取端口转发统计数据
		}

		// SSL记录管理相关API
		sslRecords := api.Group("/ssl-records")
		{
			sslRecords.GET("", handlers.GetSSLRecords)                                     // 获取全部SSL记录
			sslRecords.GET("/:id", handlers.GetSSLRecordDetail)                            // 获取SSL记录详情
			sslRecords.POST("", handlers.CreateSSLRecord)                                  // 创建SSL记录
			sslRecords.PATCH("/:id", handlers.UpdateSSLRecord)                             // 更新SSL记录
			sslRecords.PATCH("/:id/auto-renewal", handlers.UpdateSSLRecordAutoRenewal)     // 更新SSL记录自动续费状态
			sslRecords.PATCH("/:id/active", handlers.UpdateSSLRecordActive)                // 更新SSL记录启用状态
			sslRecords.DELETE("/:id", handlers.DeleteSSLRecord)                            // 删除SSL记录
			sslRecords.POST("/:id/apply", handlers.ApplySSLRecord)                         // 申请新的SSL证书
			sslRecords.GET("/:id/history/:history_id", handlers.GetSSLRecordHistoryDetail) // 获取SSL记录历史详情
		}

		// 域名监控相关API
		domainMonitors := api.Group("/domain-monitors")
		{
			domainMonitors.GET("", handlers.GetDomainMonitors)                      // 获取所有域名监控
			domainMonitors.GET("/:id", handlers.GetDomainMonitorDetail)             // 获取域名监控详情
			domainMonitors.POST("", handlers.CreateDomainMonitor)                   // 创建域名监控
			domainMonitors.PATCH("/:id", handlers.UpdateDomainMonitor)              // 更新域名监控
			domainMonitors.PATCH("/:id/active", handlers.UpdateDomainMonitorActive) // 更新域名监控启用状态
			domainMonitors.DELETE("/:id", handlers.DeleteDomainMonitor)             // 删除域名监控
			domainMonitors.POST("/:id/check", handlers.CheckDomainMonitor)          // 检查域名监控
		}

		// 通知相关API
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", handlers.GetWebhooks)                      // 获取全部通知
			webhooks.GET("/:id", handlers.GetWebhookDetail)             // 获取通知详情
			webhooks.POST("", handlers.CreateWebhook)                   // 创建通知
			webhooks.PATCH("/:id", handlers.UpdateWebhook)              // 更新通知
			webhooks.PATCH("/:id/active", handlers.UpdateWebhookActive) // 更新通知启用状态
			webhooks.DELETE("/:id", handlers.DeleteWebhook)             // 删除通知
			webhooks.POST("/:id/test", handlers.TestWebhook)            // 测试通知
		}

		// 通知订阅相关API
		webhookSubs := api.Group("/webhook-subs")
		{
			webhookSubs.GET("", handlers.GetWebhookSubs)                      // 获取全部通知订阅
			webhookSubs.PATCH("/:id/active", handlers.UpdateWebhookSubActive) // 更新通知订阅启用状态
			webhookSubs.DELETE("/:id", handlers.DeleteWebhookSub)             // 删除通知订阅
		}

		// 任务相关API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)                        // 获取全部任务
			tasks.GET("/:id", handlers.GetTaskDetail)               // 获取任务详情
			tasks.POST("/:id", handlers.RunTask)                    // 运行任务
			tasks.DELETE("/:id", handlers.DeleteTask)               // 删除任务
			tasks.POST("/clean-all", handlers.CleanAllTask)         // 清理所有任务
			tasks.POST("/start-scheduler", handlers.StartScheduler) // 启动调度器
			tasks.POST("/stop-scheduler", handlers.StopScheduler)   // 停止调度器
		}

		// 开放API
		opens := api.Group("/open").Use(gin_middleware.AuditContextMiddleware(
			ginLogger,
			operationlog.SourceAPI,
		))
		{
			opens.POST("/device/:device_code/ip-report", handlers.DeviceReportDeviceIp)                              // 设备上报IP
			opens.GET("/device/:device_code/port-forwards", handlers.DevicePortForwards)                             // 获取设备端口转发
			opens.POST("/device/:device_code/port-forwards-stats/:port_forward_id", handlers.CreatePortForwardStats) // 创建端口转发统计
		}

		// 分组内兜底：未匹配路径返回 JSON 404（不会与 /api/d 冲突）
		api.Any("/:any", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, handlers.APIResponse{
				Code:    http.StatusNotFound,
				Message: "api path not found",
			})
		})
	}
}

// Swagger 路由
// https://github.com/swaggo/swag/blob/v2/README.md
// https://github.com/swaggo/swag/issues/1588#issuecomment-3360487197
func swagger_router(ctx context.Context, ginLogger *slog.Logger, router *gin.Engine) {

	swaggerGroup := router.Group("/swagger")
	{
		// 配置 Swagger 信息
		docs_2.SwaggerInfo.Version = global.GvaVersion
		docs_2.SwaggerInfo.Host = ""
		docs_3.SwaggerInfo.Version = global.GvaVersion
		docs_3.SwaggerInfo.Host = ""
		swaggerGroup.GET("/docs_2/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_2.SwaggerInfo.ReadDoc())
		})
		swaggerGroup.GET("/docs_3/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_3.SwaggerInfo.ReadDoc())
		})
		// 配置 Swagger 页面
		swaggerGroup.GET("/:any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/swagger/docs_2/doc.json"),
		))
	}
}

// 新 Swagger 路由
// https://gitee.com/xiaowan1997/qingfeng
func qingfeng_router(ctx context.Context, ginLogger *slog.Logger, router *gin.Engine) {

	qingfengGroup := router.Group("/qingfeng")
	{
		// 配置 Swagger 页面
		qingfengGroup.GET("/*any", qingfeng.Handler(qingfeng.Config{
			Title:       docs_2.SwaggerInfo.Title,       // 文档标题
			Description: docs_2.SwaggerInfo.Description, // 文档描述
			Version:     docs_2.SwaggerInfo.Version,     // API 版本号
			BasePath:    "/qingfeng",                    // 文档路由前缀
			DocJSON:     docs_2.GetSwaggerBytes(),       // 直接传入 swagger JSON 内容
			EnableDebug: global.GvaEnv.DEBUG,            // 是否启用在线调试
			Environments: []qingfeng.Environment{
				{
					Name:    "本地开发",
					BaseURL: "/api",
				},
			}, // 多环境配置
		}))
	}
}

// MCP 启动
func mcp_run(ctx context.Context, ginLogger *slog.Logger, mcpLogger *slog.Logger, router *gin.Engine) {

	// 判断是否开启 MCP 功能
	if global.GvaMcpServer == nil {
		return
	}

	// 注册 MCP 实现
	mcp_resource.RegisterTimeResources(global.GvaMcpServer)
	mcp_tool.RegisterTimeTools(global.GvaMcpServer)

	mcp_prompt.RegisterPlatformConfigPrompts(global.GvaMcpServer)
	mcp_resource.RegisterPlatformConfigResources(global.GvaMcpServer)
	mcp_tool.RegisterPlatformConfigTools(global.GvaMcpServer)

	mcp_prompt.RegisterCDNRecordPrompts(global.GvaMcpServer)
	mcp_resource.RegisterCDNRecordResources(global.GvaMcpServer)
	mcp_tool.RegisterCDNRecordTools(global.GvaMcpServer)

	mcp_prompt.RegisterDNSRecordPrompts(global.GvaMcpServer)
	mcp_resource.RegisterDNSRecordResources(global.GvaMcpServer)
	mcp_tool.RegisterDNSRecordTools(global.GvaMcpServer)

	mcp_prompt.RegisterDevicePrompts(global.GvaMcpServer)
	mcp_resource.RegisterDeviceResources(global.GvaMcpServer)
	mcp_tool.RegisterDeviceTools(global.GvaMcpServer)

	mcp_prompt.RegisterIPGroupPrompts(global.GvaMcpServer)
	mcp_resource.RegisterIPGroupResources(global.GvaMcpServer)
	mcp_tool.RegisterIPGroupTools(global.GvaMcpServer)

	mcp_prompt.RegisterPortForwardPrompts(global.GvaMcpServer)
	mcp_resource.RegisterPortForwardResources(global.GvaMcpServer)
	mcp_tool.RegisterPortForwardTools(global.GvaMcpServer)

	mcp_prompt.RegisterDomainMonitorPrompts(global.GvaMcpServer)
	mcp_resource.RegisterDomainMonitorResources(global.GvaMcpServer)
	mcp_tool.RegisterDomainMonitorTools(global.GvaMcpServer)

	mcp_prompt.RegisterWebhookPrompts(global.GvaMcpServer)
	mcp_resource.RegisterWebhookResources(global.GvaMcpServer)
	mcp_tool.RegisterWebhookTools(global.GvaMcpServer)

	mcp_prompt.RegisterWebhookSubPrompts(global.GvaMcpServer)
	mcp_resource.RegisterWebhookSubResources(global.GvaMcpServer)
	mcp_tool.RegisterWebhookSubTools(global.GvaMcpServer)

	// 注册 MCP 路由
	httpTransport := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return global.GvaMcpServer
	}, nil)
	mcpRouter := router.Use(gin_middleware.AuditContextMiddleware(
		ginLogger,
		operationlog.SourceMcp,
	))
	mcpRouter.Any("/mcp/*path", gin.WrapH(http.StripPrefix("/mcp", httpTransport)))
	mcpRouter.Any("/mcp", gin.WrapH(http.StripPrefix("/mcp", httpTransport)))
}

// gRPC 启动
func grpc_run(ctx context.Context, grpcLogger *slog.Logger) {

	// 判断是否开启 gRPC 功能
	if global.GvaEnv.GRPC_ADDR == "" {
		grpcLogger.InfoContext(ctx, "gRPC 服务不启动，因为没有配置端口",
			slog.String("port", global.GvaEnv.GRPC_ADDR),
		)
		return
	}

	// 创建 gRPC 服务
	opts := make([]grpc.ServerOption, 0)
	usis := make([]grpc.UnaryServerInterceptor, 0)
	// 审计上下文 一元拦截器
	usis = append(
		usis,
		grpc_interceptor.AuditContextInterceptor(
			grpcLogger,
			operationlog.SourceGrpc,
		),
	)
	// BasicAuth 一元拦截器
	usis = append(
		usis,
		grpc_interceptor.BasicAuthUnaryInterceptor(
			grpcLogger,
			global.GvaEnv.BASIC_AUTH_USERNAME,
			global.GvaEnv.BASIC_AUTH_PASSWORD,
		),
	)
	// 日志 一元拦截器
	usis = append(
		usis,
		grpc_interceptor.LoggerUnaryInterceptor(grpcLogger),
	)
	opts = append(
		opts,
		grpc.ChainUnaryInterceptor(usis...),
	)
	opts = append(
		opts,
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
	)
	opts = append(
		opts,
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)
	grpcServerInstance = grpc.NewServer(opts...)

	// 注册服务
	indexServer := grpc_service.NewIndexServer()
	indexServer.RegisterService(grpcServerInstance)
	openServer := grpc_service.NewOpenServer()
	openServer.RegisterService(grpcServerInstance)

	// 启动 gRPC 服务
	var err error
	grpcListenerInstance, err = net.Listen("tcp", global.GvaEnv.GRPC_ADDR)
	if err != nil {
		grpcLogger.ErrorContext(ctx, "gRPC 服务启动失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	grpcLogger.InfoContext(ctx, "gRPC 服务正在启动",
		slog.String("port", global.GvaEnv.GRPC_ADDR),
	)

	// 使用 Serve 但不阻塞，通过 channel 监听关闭信号
	go func() {
		if err := grpcServerInstance.Serve(grpcListenerInstance); err != nil {
			grpcLogger.ErrorContext(ctx, "gRPC 服务异常退出",
				slog.Any("err", err),
			)
		}
		grpcLogger.InfoContext(ctx, "gRPC 服务已停止")
	}()

	grpcLogger.InfoContext(ctx, "gRPC 服务启动成功",
		slog.String("port", global.GvaEnv.GRPC_ADDR),
	)

	<-shutdownChan // 等待关闭信号
	grpcLogger.InfoContext(ctx, "接收到关闭信号，gRPC 服务正在优雅关闭")
	if grpcServerInstance != nil {
		grpcServerInstance.GracefulStop() // 优雅关闭
	}
	if grpcListenerInstance != nil {
		grpcListenerInstance.Close()
	}
	grpcLogger.InfoContext(ctx, "gRPC 服务已优雅关闭")
}

// gRPC-Gateway 启动
func grpc_gw_run(ctx context.Context, grpcLogger *slog.Logger, router *gin.Engine) {

	// 判断是否开启 gRPC-Gateway 功能
	if global.GvaEnv.GRPC_ADDR == "" {
		grpcLogger.InfoContext(ctx, "gRPC-Gateway 不启动，因为没有配置端口")
		return
	}

	// 创建 gRPC 客户端
	grpc_client := grpc_client.NewGRPCClient()
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		// 鉴权
		grpc.WithPerRPCCredentials(grpc_credential.NewClientBasicAuth(
			grpcLogger,
			global.GvaEnv.BASIC_AUTH_USERNAME,
			global.GvaEnv.BASIC_AUTH_PASSWORD,
		)),
	)
	// 连接 gRPC 服务
	err := grpc_client.Dial(global.GvaEnv.GRPC_ADDR)
	if err != nil {
		grpcLogger.ErrorContext(ctx, "gRPC-Gateway 连接上游失败（忽略路由注册）",
			slog.Any("err", err),
		)
		return
	}
	// 获取 gRPC 客户端
	grpcConnInstance = grpc_client.GetClient()

	gwMux := grpcGatewayRuntime.NewServeMux(
		// 错误响应统一包装
		grpcGatewayRuntime.WithErrorHandler(func(ctx context.Context, mux *grpcGatewayRuntime.ServeMux, marshaler grpcGatewayRuntime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			// 默认错误处理器
			st, _ := status.FromError(err)

			// 映射 HTTP 状态码
			httpCode := grpcGatewayRuntime.HTTPStatusFromCode(st.Code())

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
		grpcGatewayRuntime.WithForwardResponseRewriter(func(ctx context.Context, resp proto.Message) (any, error) {
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
				Message: "success",
				Data:    data,
			}, nil
		}),
	)

	// 注册服务
	if err := grpc_generated.RegisterOpenServiceHandler(ctx, gwMux, grpcConnInstance); err != nil {
		grpcLogger.ErrorContext(ctx, "注册 OpenService 网关路由失败",
			slog.Any("err", err),
		)
		return
	}
	if err := grpc_generated.RegisterIndexServiceHandler(ctx, gwMux, grpcConnInstance); err != nil {
		grpcLogger.ErrorContext(ctx, "注册 IndexService 网关路由失败",
			slog.Any("err", err),
		)
		return
	}

	// 关键：把网关挂到 Gin 的 /grpc 前缀
	grpcGwRouter := router.Group("/grpc").Use(gin_middleware.AuditContextMiddleware(
		grpcLogger,
		operationlog.SourceGrpc,
	))
	grpcGwRouter.Any("/*any", gin.WrapH(http.StripPrefix("/grpc", gwMux)))
	grpcLogger.InfoContext(ctx, "gRPC-Gateway 已挂载到 Gin",
		slog.String("prefix", "/grpc"),
		slog.String("upstream", global.GvaEnv.GRPC_ADDR),
	)
}

// Thrift 启动
func thrift_run(ctx context.Context, thriftLogger *slog.Logger) {

	// 判断是否开启 Thrift 功能
	if global.GvaEnv.THRIFT_ADDR == "" {
		thriftLogger.InfoContext(ctx, "Thrift 服务不启动，因为没有配置端口",
			slog.String("port", global.GvaEnv.THRIFT_ADDR),
		)
		return
	}

	// 启动 Thrift 服务
	socket, err := thrift.NewTServerSocket(global.GvaEnv.THRIFT_ADDR)
	if err != nil {
		thriftLogger.ErrorContext(ctx, "Thrift 服务启动失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	thriftLogger.InfoContext(ctx, "Thrift 服务正在启动",
		slog.String("port", global.GvaEnv.THRIFT_ADDR),
	)

	// 传输工厂
	transportFactory := thrift.NewTFramedTransportFactoryConf(
		thrift.NewTTransportFactory(),
		&thrift.TConfiguration{},
	)

	// 协议工厂
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})

	// 业务逻辑层 (Handler)
	openHandler := thrift_service.NewOpenServer(thriftLogger)
	processor := open.NewOpenServiceProcessor(openHandler)

	// 挂载【通用】处理器包装器
	// 无论 IDL 怎么变，这里一行代码搞定所有服务的日志和耗时
	wrappedProcessor := thrift_process.NewGenericInterceptor(processor, thriftLogger)

	// 传入 SimpleServer
	thriftServerInstance = thrift.NewTSimpleServer4(
		wrappedProcessor,
		socket,
		transportFactory,
		protocolFactory,
	)

	// 使用 Serve 但不阻塞，通过 channel 监听关闭信号
	go func() {
		if err := thriftServerInstance.Serve(); err != nil {
			thriftLogger.ErrorContext(ctx, "Thrift 服务启动失败",
				slog.Any("err", err),
			)
		}
		thriftLogger.InfoContext(ctx, "Thrift 服务已停止")
	}()

	thriftLogger.InfoContext(ctx, "Thrift 服务启动成功",
		slog.String("port", global.GvaEnv.THRIFT_ADDR),
	)

	// 等待关闭信号
	<-shutdownChan
	thriftLogger.InfoContext(ctx, "接收到关闭信号，Thrift 服务正在优雅关闭")
	if thriftServerInstance != nil {
		thriftServerInstance.Stop() // 优雅关闭
	}
	if thriftTransport != nil {
		thriftTransport.Close()
	}
	thriftLogger.InfoContext(ctx, "Thrift 服务已优雅关闭")
}
