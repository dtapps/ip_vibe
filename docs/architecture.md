# IPVibe 架构与模块设计

> 面向维护者与二次开发者。本文说明代码如何分层、请求如何流转、数据存在哪里、定时任务如何运行，以及新增功能应该落在哪一层。功能清单、快速开始与部署方式见 [README](../README.md)；CI/CD、版本号规则与发布流程见 [ci-cd.md](ci-cd.md)，CI 基础镜像见 [ci-image.md](ci-image.md)；测试与覆盖率方案见 [coverage-go.md](coverage-go.md)、[coverage-node.md](coverage-node.md)。

## 目录

- [1. 系统概览](#1-系统概览)
- [2. 代码分层与依赖方向](#2-代码分层与依赖方向)
- [3. 进程与启动流程](#3-进程与启动流程)
- [4. 通信与接入层](#4-通信与接入层)
- [5. 数据层](#5-数据层)
- [6. 核心业务链路](#6-核心业务链路)
- [7. 调度与任务模型](#7-调度与任务模型)
- [8. 厂商适配层](#8-厂商适配层)
- [9. 配置、构建与部署形态](#9-配置构建与部署形态)
- [10. 日志与可观测性](#10-日志与可观测性)
- [11. 前端与模板约定](#11-前端与模板约定)
- [12. 横切开发约定](#12-横切开发约定)
- [13. 模块职责速查](#13-模块职责速查)
- [14. 扩展指南](#14-扩展指南)
- [15. 已知约束与风险](#15-已知约束与风险)
- [附录 A：环境变量与默认路径](#附录-a环境变量与默认路径)
- [附录 B：关键文件索引](#附录-b关键文件索引)

---

## 1. 系统概览

IPVibe 由三个可执行程序组成，共用一套 `internal/` 业务代码：

| 程序      | 入口                                              | 部署位置 | 职责                                                            |
|-----------|---------------------------------------------------|----------|-----------------------------------------------------------------|
| backend   | [cmd/backend/main.go](../cmd/backend/main.go)     | 中心节点 | 控制台、REST API、gRPC 服务端、MCP、定时任务调度、厂商 API 同步 |
| reporter  | [cmd/reporter/main.go](../cmd/reporter/main.go)   | 被管设备 | 探测本机公网 IPv4/IPv6，周期上报；维持设备长连接                |
| forwarder | [cmd/forwarder/main.go](../cmd/forwarder/main.go) | 被管设备 | 拉取端口转发规则并热更新，上报流量统计；维持设备长连接          |

### 1.1 总体结构

```text
                           ┌────────────────────── backend（cmd/backend）──────────────────────┐
 浏览器 / AI 客户端 ──HTTP──▶│ Gin：控制台页面 / REST API / Swagger / MCP(/mcp)                  │
 reporter ──────────gRPC───▶│ h2c 混合监听 :8080 ─┬─ Content-Type: application/grpc ─▶ gRPC Server │
 forwarder ─────────gRPC───▶│                     └─ 其他请求 ────────────────────────▶ Gin Engine  │
                           │                                                                    │
                           │ cron_manager ──▶ gocron ──▶ provider/task/* ──▶ 厂商 API / DB         │
                           │ backend/syncs ──▶ provider/{dns,cdn} ──▶ Cloudflare / EdgeOne / ESA   │
                           │                                                                    │
                           │ data/ent.db（业务）  data/ent_log.db（日志）  data/ent_stat.db（统计）│
                           └────────────────────────────────────────────────────────────────────┘
```

### 1.2 关键设计取舍

1. **单仓库多入口**：三个程序共享 `internal/`，但只有 backend 初始化数据库与调度器；reporter/forwarder 不落库，配置全部通过 gRPC 从 backend 拉取。
2. **backend 单实例**：设备连接表、运行中的 gocron 任务、各 Manager 单例都只存在于进程内存，部署多个副本会出现重复调度与连接不一致。详见[第 15 节](#15-已知约束与风险)。
3. **三库分离**：业务数据、日志/审计、统计指标分别落在三个 SQLite 文件，避免大日志表拖慢业务查询。
4. **数据库驱动调度**：`cron_job` 表是任务事实源，gocron 只是执行器；任务增删改通过表数据 + Ent Hook 同步到调度器，重启后从表恢复。
5. **厂商适配器**：DNS/CDN/TUN 通过统一接口屏蔽厂商差异，业务任务只依赖接口，不直接调用云 SDK。
6. **单端口多协议**：Gin、gRPC、MCP 共用一个 h2c 端口，节点侧只需配置一个 `BACKEND_ENDPOINT`。
7. **拉模型同步**：IP 上报不直接改 DNS/CDN，而是更新设备表；DNS/CDN 任务周期性读取“心跳窗口内”的设备 IP 再同步，解耦上报与厂商 API。
8. **无框架依赖**：配置只用环境变量，前端不用任何 SPA 框架（原生 JS + jQuery + DaisyUI/Tailwind），后端不引 Ent 的 privacy feature，部署不引 goreleaser。选型偏“可控、可单二进制交付”。

---

## 2. 代码分层与依赖方向

```text
cmd/*                         进程入口、参数/环境变量、启动顺序
  │
  ├─ internal/initialize      日志、三库、gocron 等基础设施初始化
  ├─ internal/manager         混合服务管理、定时任务管理
  ├─ internal/gin_server      Gin 引擎、模板/静态资源、路由
  ├─ internal/grpc_server     gRPC 引擎、拦截器、服务注册
  ├─ internal/mcp_server      MCP 服务与 /mcp 挂载
  │
  ├─ internal/handlers        页面与 REST 入参绑定、响应组装
  │    └─ internal/types      API 请求/响应 DTO
  │
  ├─ internal/backend/service 业务编排、事务、CRUD
  ├─ internal/backend/syncs   平台全量同步（Cloudflare / EdgeOne / ESA）
  ├─ internal/provider/task   可调度任务实现（同步、SSL、监控、清理、通知）
  │
  ├─ internal/provider/{dns,cdn,tun}  厂商适配器
  ├─ internal/helper/*        领域辅助（IP/域名/SSL/gocron/Webhook/db_cache/日志）
  ├─ internal/core            无状态领域函数（设备标识、IP 探测、BasicAuth token）
  │
  └─ internal/ent | ent_log | ent_stat   Ent Schema 与生成代码
```

约定：

- **依赖方向自上而下**。`provider` 不依赖 Gin/Ent 的具体业务表；`handlers` 不承载复杂事务；跨表事务放在 `backend/service`。
- **任务实现统一放在 `provider/task`**，实现 `provider.TaskProvider`。`cron_manager` 只负责注册表、调度与执行日志，不认识具体业务。
- **依赖注入采用“全局单例 + `InjectXxx` + `sync.Once`”**：
  - 全局对象在 [internal/global/global.go](../internal/global/global.go)；
  - service/syncs/provider 各自有 `inject.go`，首次注入生效，后续注入被 `sync.Once` 忽略；
  - 因此 **初始化顺序是硬约束**：必须先初始化 DB/日志，再构造 Manager/任务。
- **HTTP 与 gRPC 共用业务服务**：gRPC 的 `OpenService` 直接转调 `backend/service`，不复制业务逻辑。

---

## 3. 进程与启动流程

### 3.1 backend

入口 [cmd/backend/main.go](../cmd/backend/main.go) 无 CLI flag，配置全部来自环境变量（见[附录 A](#附录-a环境变量与默认路径)）。

启动顺序：

1. 安装全局 panic 兜底；创建带审计来源 `system` 的根 Context（`main.go:62`）。
2. 依次调用 8 个 `Init*SLog`，实际创建 10 个命名 logger（`InitEntSLog` 一次建 ent/ent_log/ent_stat 三个），`EnsureDefaultSLog` 兜底非空（`main.go:64-72`，详见[第 10 节](#10-日志与可观测性)）。
3. 初始化三个 SQLite 库并自动建表，注入 Ent/审计 Hook/Service/Syncs 依赖（`main.go:74-76`，[initialize/ent.go](../internal/initialize/ent.go)）。
4. 初始化运行时设置 `db_cache`（缺省值就地写入）、gocron 调度器、HTTP/Resty 请求日志钩子（`main.go:77-81`）。
5. 创建 Gin/gRPC Manager；加载模板与静态资源（此处会注册 `/static`）；**之后**才挂全局 BasicAuth；注册页面/API/Swagger 路由（`main.go:94-125`）。
6. 异步启动 MCP（向 Gin 注册 `/mcp`），异步启动 Gin+gRPC 混合服务（`main.go:128-142`）。
7. 启动 gocron，创建 `cron_manager`，注册 11 个任务处理器；加载系统任务（SSL 刷新、数据清理）；把 `cron_job` 中 `status=running` 的任务装入调度器（`main.go:145-174`）。
8. 把 `SyncJob` / `SyncPlatformConfigJob` 注入 Ent Hook，之后任务表的增删改才会自动同步到调度器（`main.go:177-180`）。
9. 等待 SIGINT/SIGTERM，执行优雅关闭（`main.go:187-249`）。

两个容易踩的顺序细节：

- `/health`（`gin_server/server.go:112`，建引擎时注册）与 `/static`（`InitResource` 内注册）都在 `router.Use(BasicAuth)` **之前**挂好，因此**不受认证保护**；页面、`/api`、Swagger、`/mcp` 都在其后，受认证保护。
- 模板与静态资源来自磁盘还是 `go:embed` 由 `Version == "dev"` 决定（[gin_server/resource.go](../internal/gin_server/resource.go) `InitResource`），版本号由 ldflags 注入。

关闭顺序：关闭所有设备流 → 异步关闭混合 HTTP 服务 → 停止 gocron → 关闭三库，整体 30 秒超时。注意混合服务关闭是 `go mixedMgr.Stop(ctx)`，并未等待完成就继续关库；MCP server 没有停止步骤；`gracefulShutdown` 里重新 `NewManager(...)` 依赖 `sync.Once` 的单例语义。见[第 15 节](#15-已知约束与风险)。

### 3.2 reporter

入口 [cmd/reporter/main.go](../cmd/reporter/main.go)：

1. 初始化日志与 gocron；`DEVICE_CODE` 为空时用 `machineid.ID()` 生成（读 `/etc/machine-id`）。
2. 用 `BACKEND_ENDPOINT` + BasicAuth 建立 gRPC 连接；未配置则直接 panic。
3. 创建 IP 上报任务：首次立即执行，之后按 backend 返回的 `IPReportIntervalTime`（默认 `8-12` 分钟随机）运行。
4. 启动设备长连接：首帧发送 `ClientInfo`，收到 `Ack` 后每 30 秒心跳、每 30 分钟重发全量信息；断线按 1→30 秒指数退避重连。
5. 主 goroutine `select {}` 常驻。

**reporter 与 forwarder 都没有信号处理**：主线程永久阻塞在 `select {}`，`defer CloseGoCron()`、`defer task.Close()`、`defer grpcConn.Close()` 在收到 SIGTERM 时不会执行，容器停止依赖 SIGKILL 兜底。只有 backend 实现了优雅关闭。

### 3.3 forwarder

入口 [cmd/forwarder/main.go](../cmd/forwarder/main.go)：

1. 初始化日志/gocron，生成设备编码，建立 gRPC 连接。
2. 创建 `forward.Agent`（端口转发实例管理器）。
3. 启动规则刷新任务：启动即拉取一次，之后按 `PortForwardReloadTime`（默认 30 分钟）刷新。
4. 启动统计上报任务：按 `PortForwardReportTime`（默认 `8-12` 分钟）逐规则上报。
5. 启动设备长连接，逻辑与 reporter 相同（`role=forwarder`）。

---

## 4. 通信与接入层

### 4.1 Gin + gRPC 同端口（h2c）

实现见 [mixed_manager/mixed_server.go](../internal/manager/mixed_manager/mixed_server.go)：

- 使用 `h2c.NewHandler` 让明文 HTTP/2 可用。
- 分流条件：`r.ProtoMajor == 2` 且 `Content-Type` 以 `application/grpc` 开头 → gRPC；否则 → Gin（`mixed_server.go:62`）。
- HTTP/2：30 秒读空闲 PING、15 秒 PING 超时、最大并发流 100；连接空闲 120 秒。
- 监听地址只有一处配置：`SERVER_ADDR`（默认 `0.0.0.0:8080`）。启动是先打印“启动成功”日志、再 `ListenAndServe`（`:103`、`:108`），端口占用时打印错误后 `panic`。
- `ReadTimeout` / `WriteTimeout` 在源码里被注释掉（`:84-85`），即**入站连接没有读写超时**。
- 未配置 TLS，节点与 backend 之间是明文 h2c，鉴权仅靠 BasicAuth。

由于是前缀匹配，`application/grpc-web` 也会进入 gRPC 分支。

### 4.2 HTTP 中间件顺序

注册位置 [gin_server/server.go](../internal/gin_server/server.go) `initServer`：

| 顺序 | 中间件         | 作用                 | 关键实现细节                                                                                          |
|------|----------------|----------------------|-------------------------------------------------------------------------------------------------------|
| 1    | `gin.Logger`   | 访问日志             | 写入 slog writer，无采样                                                                              |
| 2    | `Recovery`     | panic 恢复           | 只返回裸 `500` 状态码，没有 JSON body（`recovery.go:59`）                                             |
| 3    | `AuditContext` | 写审计来源到 Context | **入参被忽略，恒写 `operationlog.SourceUser`**（`audit.go:16`）                                       |
| 4    | `RequestID`    | 请求 ID              | 透传或生成 `X-Request-ID`，并回写响应头（`requestid.go:27-47`）                                       |
| 5    | `Coraza`       | WAF 中间件           | 只设 `SecRuleEngine On` + `SecRequestBodyAccess On`，未 Include 任何规则；`io.ReadAll` 读 body 无上限 |
| 6    | `CORS`         | 跨域                 | 任意 Origin 放行、`AllowCredentials=true`、Max-Age 86400                                              |
| 7    | `Gzip`         | 响应压缩             | 默认压缩级别，只排除 `.svg`                                                                           |
| 8    | `BasicAuth`    | 全局认证             | 在 `main.go:118` 追加，位于页面/`/api`/Swagger/`/mcp` 之前，但晚于 `/health`、`/static`               |

注意：

- 源码中 CORS/Gzip 两处注释与实际注册顺序相反，实际顺序以上表为准。
- `/health` 与整棵 `/static` 在 BasicAuth 之前注册，**可匿名访问**；其余入口共用同一套凭据。
- `AuditContextMiddleware` 忽略传入的 source，所以 MCP 组虽然传了 `SourceMcp`，落库仍是 `user`；gRPC 侧的 `SourceGrpc` 由 gRPC interceptor 直接 `WithSource`，不受影响。详见[第 15 节](#15-已知约束与风险)。

### 4.3 gRPC 服务与拦截器

注册位置 [grpc_server/server.go](../internal/grpc_server/server.go)：

- 一元与流式拦截器顺序一致：审计 `SourceGrpc` → 日志 → BasicAuth。
- 收发消息上限 10 MB。未注册 reflection 服务。
- 注册三个服务：

| 服务               | Proto                                                       | 方法                                                                              | 说明                                 |
|--------------------|-------------------------------------------------------------|-----------------------------------------------------------------------------------|--------------------------------------|
| `DeviceService`    | [device.proto](../internal/grpc_proto/device.proto)         | `Stream(stream ClientMessage) returns (stream ServerMessage)`                     | 设备注册/心跳长连接                  |
| `OpenService`      | [open.proto](../internal/grpc_proto/open.proto)             | `GetSetting`、`ReportDeviceIP`、`GetDevicePortForwards`、`CreatePortForwardStats` | 节点侧开放接口，均为一元调用         |
| `RemoteLogService` | [remote_log.proto](../internal/grpc_proto/remote_log.proto) | `LogTunnel(stream DeviceResponse) returns (stream ServerRequest)`                 | 服务端下发文件名，客户端回传日志尾部 |

`OpenService` 在 proto 中带有 grpc-gateway 的 HTTP 注解，并生成了 `open.pb.gw.go`，但当前 backend **没有挂载 gateway**（`grpc_gateway_client` 包无调用方）。节点实际走原生 gRPC；README 中列出的 `/setting`、`/device/ip-report` 等 gateway 路径目前不是 Gin 路由。

### 4.4 鉴权模型

- HTTP/MCP/Swagger/REST：全局 BasicAuth，走 `gin.BasicAuth`（常数时间比较）。用户名或密码任一为空 → 直接放行（即关闭认证）。realm 为默认值，401 只有 `WWW-Authenticate` 头与空 body。
- gRPC 节点侧：客户端通过 `PerRPCCredentials` 注入 `authorization: Basic base64(user:pass)`；服务端拦截器解析 metadata 校验，**密码用 `!=` 直接比较，非常数时间**（`grpc_interceptor/basic_auth.go:64,109`），与 HTTP 侧行为不一致。
- 客户端只对一元调用配置了重试，流式调用没有重试。
- 所有通道共用同一组 `BASIC_AUTH_USERNAME` / `BASIC_AUTH_PASSWORD`，无用户表、无会话、无按角色区分。

### 4.5 接口入口

| 入口        | 路径                                                                                     | 实现                                                                         |
|-------------|------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| 控制台页面  | `/` `/dashboard` `/cdn` `/dns` `/ssl` `/tun` `/device` `/port-forward` `/task` `/log` 等 | [gin_server/router.go](../internal/gin_server/router.go) `setupWebRouter`    |
| REST API    | `/api/...`                                                                               | `setupApiRouter`，按资源分组                                                 |
| 健康检查    | `GET /health`（**无需认证**）                                                            | [gin_server/server.go](../internal/gin_server/server.go)                     |
| Swagger     | `/swagger/docs_2/doc.json`、`/swagger/docs_3/doc.json`、`/swagger/:any`                  | [internal/docs_2](../internal/docs_2)、[internal/docs_3](../internal/docs_3) |
| qingfeng UI | `/qingfeng/*`                                                                            | [gin_server/router.go](../internal/gin_server/router.go)                     |
| MCP         | `/mcp`（GET/POST/任意子路径）                                                            | [mcp_server/server.go](../internal/mcp_server/server.go)                     |
| gRPC        | 同端口，`application/grpc`                                                               | [grpc_server/server.go](../internal/grpc_server/server.go)                   |

- 页面路由实际有 34 条 GET，包含列表页与 `/:id` 详情/表单子页；上表只列主要入口。
- `/api` 下有 22 个资源分组：settings、schemas、platform-configs、cdn-records、dns-records、ssl-records、ssl-imports、tun-records、domain-monitors、devices、device-groups、ip-groups、port-forwards、webhooks、webhook-subs、tasks、gocrons、cloudflares、cloudflare-tunnels、cloudflare-zones、edgeone-zones、esa-sites。另有 `GET /api/` 探活与 `Any /api/:any` 兜底（返回统一信封的 404/405）。
- `/api/schemas/*` 4 条路由挂了 `CacheByRequestURI` 的 30 秒内存缓存（store TTL 2 分钟），改完 Schema 最长 30 秒才生效，且缓存不区分认证用户。
- 没有 pprof、没有 `/metrics`，对外可观测入口只有 `/health`、日志页与 `cron_job_log`。

### 4.6 MCP

- 采用 Streamable HTTP，无 stdio 模式；挂载在 Gin 的 `/mcp` 组，并叠加 `AuditContextMiddleware`（实际写入 `user`，见 4.2）。
- 只注册了 `GET /mcp`、`POST /mcp` 与 `Any /mcp/*any`：**没有 DELETE 路由**，有状态会话无法按规范终止，客户端只能靠超时回收；请求体受 SDK 默认 4 MiB 上限约束；SSE 流会被全局 Gzip 包裹。
- 按业务域成组注册 prompt / resource / tool：time、platform-config、cdn-record、dns-record、device、ip-group、port-forward、domain-monitor、webhook、webhook-sub。
- 工具名采用“动词 + 实体驼峰”（如 `getDeviceList`、`updatePortForwardActive`）；参数 schema 由 `internal/utils.AutoSchema` 基于请求结构体反射生成。
- MCP handler 直接构造 backend service + 全局 DB 执行操作，与 REST 共用同一套业务逻辑。

---

## 5. 数据层

### 5.1 三个 SQLite 库

| 库     | 文件               | 用途                                                                                                   | 连接配置                                                             |
|--------|--------------------|--------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------|
| 业务库 | `data/ent.db`      | 设备、平台配置、DNS/CDN/SSL/TUN 记录、Webhook、任务定义                                                | WAL、外键、`busy_timeout=5000`、`MaxOpenConns(1)`、`MaxIdleConns(1)` |
| 日志库 | `data/ent_log.db`  | 操作审计、IP 上报历史、任务执行日志、HTTP/Resty 请求日志、Webhook 发送历史、端口转发统计、域名监控历史 | 同上                                                                 |
| 统计库 | `data/ent_stat.db` | Cloudflare / EdgeOne / ESA 的站点日聚合指标                                                            | 同上                                                                 |

初始化见 [initialize/ent.go](../internal/initialize/ent.go)：

- 每个库启动时执行 `Schema.Create` 自动建表/补列，没有版本化 migration。
- **DSN 硬编码**（`file:data/ent.db?_pragma=...`），路径、pragma 都不能通过环境变量改；只有 `DEBUG` 可调 Ent 日志级别。
- `MaxOpenConns(1)` 是单写串行化的根因，SQLite 写锁不会排队爆掉，但业务写入会串行。
- 只有业务库与日志库会注入审计 Hook，统计库不注入。
- 没有第四个库：运行时设置是业务库里的 `cache` 表（见 5.6）。

### 5.2 业务库实体（30 张表）

| 分组         | 表                                                                                                                      | 说明                                           |
|--------------|-------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|
| 设备拓扑     | `device`、`device_group`、`device_group_device`、`ip_group`、`ip_group_device`、`device_port_forward`                   | 设备、设备组/IP 组及关联表、端口转发规则       |
| 平台与记录   | `platform_config`、`dns_record`、`cdn_record`、`ssl_record`、`ssl_import`、`tun_record`                                 | 平台凭据、各类同步记录、证书与手工导入         |
| Cloudflare   | `cloudflare_zone`、`cloudflare_zone_record`、`cloudflare_tunnel`、`cloudflare_tunnel_ingress`                           | 站点、DNS 记录、隧道与入站规则                 |
| 腾讯 EdgeOne | `edgeone_zone`、`edgeone_zone_acceleration_domain`、`edgeone_zone_dns_record`、`edgeone_zone_origin_group`              | 站点、加速域名、DNS 记录、源站组               |
| 阿里 ESA     | `esa_site`、`esa_site_record`、`esa_site_edge_container_record`、`esa_site_edge_routine_record`、`esa_site_origin_pool` | 站点、记录、边缘容器/边缘函数、源站池          |
| 运维         | `cron_job`、`cache`、`domain_monitor`、`webhook`、`webhook_subscription`                                                | 任务定义、运行时 KV 设置、域名监控、通知与订阅 |

表名以各 Schema 上的 `entsql.Annotation{Table: ...}` 为准；计数与命名可与 `internal/ent/migrate/schema.go` 对照。

### 5.3 日志库与统计库

日志库（10 张表，无 Ent 关系边，只有逻辑外键）：

`cron_job_log`、`device_ip_report`、`domain_monitor_history`、`http_log`、`req_log`、`resty_log`、`operation_log`、`port_forward_stats`、`ssl_certificate_history`、`webhook_history`。

> `req_log` 目前**只有清理与查询代码，没有任何写入点**（全仓无 `ReqLog.Create`），`logs/req_backend.log` 是同名字系统的残留。

统计库（3 张表）：

`cloudflare_zone_analysis_data`、`edge_one_zone_analysis_data`、`esa_site_analysis_data`。三张表都对 `(zone_id 或 site_id, date)` 建了**唯一索引**，同步任务用 `OnConflictColumns(...).UpdateNewValues` 做 upsert，因此每天每站点一行，可重复回填。

### 5.4 关键关系与约束

- `platform_config` 1→N 关联 DNS/CDN/SSL/TUN 记录、Cloudflare Zone/Tunnel、EdgeOne Zone、ESA Site，删除平台配置级联删除这些子记录（[platformconfig.go](../internal/ent/schema/platformconfig.go) `Edges`）。
- `device` ↔ `device_group`、`device` ↔ `ip_group` 均为多对多，经 `Through` 关联表连接。关联表以 `(device_id, device_group_id)` 为复合主键、只有 `created_at`，两条外键都是 `ON DELETE NO ACTION`（`internal/ent/migrate/schema.go:601-625`）；由于启动时开了 `foreign_keys`，删除仍在组内的设备会撞外键约束，必须先 `Clear`/`Remove` 边。
- 删除设备级联删除其 `device_port_forward` 规则。
- `dns_record` 的源站三选一：关联 IP 组、设备组或 CDN 记录。所谓 `cdn_record_id` 唯一只是 Ent 的 1:1 边语义，**数据库层没有 UNIQUE 约束/索引**（`internal/ent/migrate/schema.go:446,518`）。
- `cdn_record` 也可关联 IP 组或设备组，但当前同步任务只读取 IP 组（见[第 15 节](#15-已知约束与风险)）。
- `webhook_subscription` 用 `obj_type + obj_id` 多态指向任意业务对象，`(webhook_id, obj_type, obj_id)` 唯一；它同时存在两条指向 `webhook` 的外键：未被业务使用的 `webhook_subscriptions` 列是 CASCADE，而真正承载数据的 `webhook_id` 是 `NO ACTION` —— **删除 webhook 不会级联清理订阅行**。
- Cloudflare Zone→Record、Tunnel→Ingress，EdgeOne/ESA 父表→子表均为级联子表；隧道 ingress 采用“先全删再 `CreateBulk`”。
- 外键列名沿用边名而非统一的 `*_id`（如 `device_port_forward.device_port_forwards`），手写 SQL 时注意。

### 5.5 Schema 约定

- **时间字段**：`AtMixin` 提供 `created_at` / `updated_at`（其中 `updated_at` 标了 `Immutable()`，只在 DB 层 `UpdateDefault` 刷新）；日志/统计表多用 `AtLogMixin`（仅 `created_at`），但 `cron_job_log` 例外使用了 `AtMixin`，`http_log`/`req_log`/`resty_log` 直接内联声明 `created_at`。同步类实体统一挂 **`SyncAtMixin`**，提供 `remote_created_at` / `sync_at`（当前 14 个实体使用）。
- **枚举**：统一放在 [schematype/enum.go](../internal/ent/schema/schematype/enum.go)，`ent_log` 与 `ent_stat` 直接复用同一份包；Schema 中用 `field.Enum(...).Values(...)` 落库为字符串。
- **JSON**：配置、参数、快照统一用 `field.JSON`。
- **敏感字段**：`platform_config.config_json`、`ssl_record`/`tun_record` 的 `config_json` 与 `last_api_response`、`ssl_import`、`ssl_certificate_history` 标了 `Sensitive()`。**该标记只影响生成代码的 `String()` 输出，不影响 JSON 序列化**（Ent 的 privacy feature 未启用），因此 REST/页面响应仍会原样返回凭据与私钥，脱敏需在 handler 侧手工处理（见[第 15 节](#15-已知约束与风险)）。
- **主键/编码**：业务表用自增 `int64 id`；`cron_job.entry_id` 用 UUID v7，唯一且不可变，作为 gocron 的稳定标识。
- **审计 Hook**：`schemahook/audit_gen.go` 为 28 个业务实体生成 `hook.On`，只有两张多对多关联表没有；每次 Create/Update/Delete 起一个 goroutine 异步写 `operation_log`（来源、操作类型、对象 ID、变更前后值），用 `context.WithoutCancel` 脱离请求生命周期。
- **额外 Hook**：`CronJobGocronHook`（仅 Create/Delete/DeleteOne）、`PlatformConfigGocronHook`（含 Update）、`EsaSiteRecordDataHook`。前两个用于把表变更同步到调度器。

### 5.6 迁移、代码生成与数据保留

- 运行时迁移：启动时 `Schema.Create` 自动建表/补列；没有 down migration，也没有 migration 版本表。
- **代码生成有两条不等价的路径**，见 [9.3](#93-代码生成有两条分叉路径重要)。
- `sql/` 不是迁移目录：`data_ddl.sql`、`log_ddl.sql` 是手工维护的 DDL 快照（已过期，前者缺 `cron_job`、后者缺 `cron_job_log`），另有 `console.sql` 和若干导出的 JSON 夹具；没有任何 Go 代码引用它们。
- 数据保留：`clear_data` 任务按表删除 N 天前数据，覆盖 8 张日志表——`http_log`/`resty_log`/`req_log`/`domain_monitor_history` 保留 90 天，`operation_log`/`webhook_history`/`device_ip_report`/`port_forward_stats` 保留 365 天。
- **`cron_job_log`、`cron_job`（已完成的 one_off 行）与统计库三张表都不在清理范围内**。
- 设置项存在业务库 `cache` 表，key 带 `conf:` 前缀，只有 4 项：离线判定 12 分钟、IP 上报 `8-12` 分钟、转发规则刷新 30 分钟、统计上报 `8-12` 分钟（[db_cache/db.go](../internal/helper/db_cache/db.go)）。

---

## 6. 核心业务链路

### 6.1 设备注册、心跳与 IP 上报

```text
reporter / forwarder
  └─ gRPC DeviceService.Stream
       ├─ 首帧 ClientInfo(device_code, role, version, os, arch, ...)
       ├─ 收 Ack
       ├─ 每 30s 发 Heartbeat
       └─ 每 30min 重发 ClientInfo

reporter（8-12 分钟随机）
  └─ 外部回显服务探测公网 IPv4/IPv6
       └─ gRPC OpenService.ReportDeviceIP
            ├─ 事务：设备不存在则创建；IP 变化则更新 IP/上报时间/心跳时间；未变化只刷心跳
            ├─ 提交事务
            └─ 异步：写 device_ip_report + 创建 webhook_send 一次性任务（事件 IP_UPDATE）
```

关键点：

- gRPC 入口在 [grpc_service/open_service.go](../internal/grpc_service/open_service.go)，业务体在 [backend/service/open_service.go](../internal/backend/service/open_service.go)。新建设备时同时写入 `last_reported_at` 与 `last_heartbeat_at`。
- 设备在线判定读取 `device.last_heartbeat_at` 与 `device_offline_invalid_time`（默认 12 分钟）；该字段为空显示“未知”；阈值 `<=0` 时恒判定为在线。
- **gRPC 心跳只更新服务端内存中的 `LastSeen`，不写库**；数据库心跳由 `ReportDeviceIP` 更新。因此设备即使长连接在线，只要 IP 上报任务停了，也会被判定为离线。
- 服务端用内存 map `device_code|role` 维护连接，断流即删除；不持久化在线状态。
- 查询设备状态时会先向设备发 `RequestFullReport`，再汇总内存中的各角色版本信息。
- IP 变化只触发通知，不直接触发 DNS/CDN 更新。

### 6.2 DNS / CDN 同步

以 DNS 为例（[dns-sync_task.go](../internal/provider/task/dns-sync_task.go)）：

1. 根据 `handler_args.id` 取记录，预加载平台配置、CDN 记录、IP 组及组内设备。
2. 记录未启用 → 跳过。
3. 计算记录值：
   - CNAME：使用关联 CDN 记录的 `cdn_domain`；
   - A/AAAA：遍历 IP 组内“`last_heartbeat_at >= now - 离线阈值`”的设备，按 ID 升序；**DNS 只保留最后一个匹配设备的 IP**（循环内覆盖赋值）。
4. 无可用源站 → 跳过。
5. 按平台类型构造 `DNSProvider`，先 `ParseCustomConfig(platform.config_json)` 解析账号凭据，再解析记录级 `config_json`。
6. `Check` 返回 `IsContinue=false` 时不调用 `Update`；否则执行 `Update`。
7. 回写平台侧记录 ID（CDN 还会回写 `cdn_domain`）；发生变化时创建 `webhook_send` 任务。**记录上没有“最近同步时间”字段**，判断是否需要同步完全依赖 `Check` 与厂商侧对比。

CDN 同步（[cdn-sync_task.go](../internal/provider/task/cdn-sync_task.go)）流程相同，但会**收集 IP 组内所有匹配设备的 IP** 作为多源站。

调度节奏：两者均为 10–30 分钟随机；记录创建/更新/启停时由 service 调用 `LoadTask` 维护任务行。

> 注意：记录表的 `device_group_id` 字段与关联边存在，但 DNS/CDN/TUN 三个任务的源码里都没有引用设备组，设备组取源站尚未实现。

### 6.3 TUN 同步（当前未接线）

设计意图与 DNS/CDN 类似：`tun_sync` 任务读取 `tun_record`，调用 `TUNProvider.Check/Update` 同步 Cloudflare Tunnel 的 ingress，变更时发通知；任务定义为 10–30 分钟随机。

**现状**：`TUNSyncTaskProvider` 已实现（含 `LoadTask`），但：

- `cron_manager.InitHandlers` 没有注册 `tun_sync`；
- [tun-record_service.go](../internal/backend/service/tun-record_service.go) 全文没有引用任务/cron，`TUNSyncTaskProvider.AddTask` 是空实现。

因此该任务目前不会被执行。新增 TUN 同步能力时需要同时补注册与任务生命周期。

### 6.4 SSL 证书

- 申请/续期走 [internal/helper/ssl_helper](../internal/helper/ssl_helper)，基于 lego v5 的 DNS-01 校验；`SSLProcess` 接口下共有 10 个厂商实现（cloudflare、tencent_dns、tencent_eo、aliyun_dns、aliyun_esa、baidu_dns、huawei_dns、volcengine_dns、rainyun、westcn）。
- `ssl_apply` 任务（8–10 小时随机）：
  - 仅 `task_type=recurring` 时“记录未启用则跳过”；
  - 证书未过期且距过期天数大于 `renewal_days`（默认 30）则跳过；
  - 关闭自动续期则跳过；
  - 否则按平台 `switch` 选择 challenge —— **当前只接线 7 个平台**，`volcengine_dns`/`rainyun`/`westcn` 三个实现存在但任务里没有分支。
  - 签发成功后写入证书链、私钥、有效期等字段，并写 `ssl_certificate_history`。
- `ssl_refresh` 任务每天 00:00 执行，刷新 `ssl_record` 与 `ssl_import` 两张表的证书元数据；**它不写 `ssl_certificate_history`**（写历史只发生在申请与手工操作）。
- 支持手工导入证书（`ssl_import`）。
- `provider.SSLProvider` 接口仍保留，但没有任何实现，实际逻辑在 `ssl_helper`；`internal/provider/ssl/` 只剩 `opts.go`。

### 6.5 域名监控

`domain_monitor` 任务按记录自身的 `check_interval`（秒）执行，检查逻辑在 [domain_helper/check.go](../internal/helper/domain_helper/check.go) `Check`，**顺序执行**：

1. 把 `protocol + domain` 拼成 URL；解析失败即记录失败。
2. DNS 解析：固定使用上游 DNS `119.29.29.29:53`（硬编码），结果写 `ip_resolved`。
3. HTTPS 时额外 `tls.Dial` 抓证书：生效/过期时间、CN、签发者、SAN、证书链、序列号、TLS 版本；TLS 探测设 `InsecureSkipVerify: true`（只采集不校验）。
4. HTTP/HTTPS 取数：在**标准库**与 **resty** 两条实现之间**随机选一条**执行，可带自定义请求头（`k=v` 形式）。
5. 写 `domain_monitor_history`，更新 `domain_monitor` 的 `current_status`（success/failure）与 `last_success_at` / `last_failure_at`。

> 没有 TCP/UDP 连通性探测，也没有并发检测；探测点固定用腾讯 DNS，无法配置。

### 6.6 端口转发

```text
forwarder 规则刷新（启动即跑，之后默认 30 分钟）
  └─ OpenService.GetDevicePortForwards(device_code)
       └─ 仅返回 is_active=true 的规则
            └─ forward.Agent.Sync：按规则 ID 差量
                 ├─ 删除不存在的规则并 cancel
                 └─ 新增规则：按协议创建监听
                      ├─ tcp4/tcp6/udp4/udp6
                      ├─ tcp4+tcp6（或 udp4+udp6）可合并双栈监听
                      └─ implementation_type=standard / tcpproxy

forwarder 统计上报（默认 8-12 分钟，首次延迟 1 分钟）
  └─ 读取每个规则的原子计数快照
       └─ OpenService.CreatePortForwardStats
            └─ backend 校验设备与规则后写 port_forward_stats
                 └─ 查询时按同一 start_time 周期做差分，得到速率/增量
```

- **`Agent.Sync` 只比较规则 ID**：同一 ID 的目标地址、协议、实现类型变更会被静默忽略。改规则必须“停用再启用”（换 ID 语义）或重启 forwarder。
- `tcpproxy` 仅支持 TCP，且只累加 `Errors`，**不统计字节与连接数**；UDP 回退到标准实现。
- `auto_forward`（防火墙自动放行）在 proto 与表字段中存在，但 forwarder 全仓没有引用，未实现。
- 转发实例重启会更换 `start_time`，统计查询会把新周期整段计入，不做跨重启差分。

### 6.7 Webhook 通知

- 一个 `webhook` 可被多个 `webhook_subscription` 订阅；`obj_type` 枚举有 8 个值：`cdn_record`、`dns_record`、`device`、`device_group`、`ip_group`、`domain_monitor`、`ssl_record`、`tun_record`。
- 事件发生时业务代码调用 `WebhookSendTaskProvider.AddTask` 创建一次性 `webhook_send` 任务（1 秒后立即执行）。
- 执行时按 `meta.business_type` 映射到 `obj_type`，再查该对象下所有启用订阅及其启用的 webhook。当前只有 6 个 `business_type` 有映射，**`device_group` / `ip_group` / `ssl_record` 没有触发路径**；且 `BusinessTypeSsl` 被映射成 `ObjTypeDomainMonitor`（[webhook-send_task.go:158-159](../internal/provider/task/webhook-send_task.go)），看起来是复制粘贴导致的错误。
- 用 Go `text/template` 渲染标题/内容，模板分隔符为 `[[` `]]`（避免与 `{{ }}` 混淆）。
- 最多 10 个并发发送，支持企业微信、钉钉、飞书、Bark、Server 酱（ft07）、PushDeer，其余平台落入“自定义 HTTP”。
- 每次发送写 `webhook_history`（URL、标题、内容、响应、成功标志、错误信息）。
- 手工测试走“`objType` 为空”的分支，直接按 webhook ID 发送；该分支把结果写成 `SetSuccess(err == nil)`，而此处外层 `err` 恒为 nil，**手工测试的历史记录恒显示成功**，排查时要看 `response` 字段而不是 `is_success`。

### 6.8 云平台全量同步（Cloudflare / EdgeOne / ESA）

实现集中在 [internal/backend/syncs](../internal/backend/syncs)，由平台配置 Hook 创建 1–2 小时随机周期任务，也可通过 REST 手工触发一次性同步（1 秒后立即执行）。

| 平台              | 同步内容                                           | 落库                                                                         |
|-------------------|----------------------------------------------------|------------------------------------------------------------------------------|
| Cloudflare Zone   | 站点列表、DNS 记录、按日统计                       | `cloudflare_zone`、`cloudflare_zone_record`、`cloudflare_zone_analysis_data` |
| Cloudflare Tunnel | 隧道列表、ingress 规则（先删后批量建）             | `cloudflare_tunnel`、`cloudflare_tunnel_ingress`                             |
| 腾讯 EdgeOne      | 站点、L7/L4 时序日聚合、源站组、DNS 记录、加速域名 | `edgeone_zone*`、`edge_one_zone_analysis_data`                               |
| 阿里 ESA          | 站点、时序指标、源站池、记录、边缘函数、边缘容器   | `esa_site*`、`esa_site_analysis_data`                                        |

- 同步按平台配置并发执行，单个账号内并发上限 5；DB 客户端与 HTTP transport 通过 `syncs.InjectDb/InjectLogDb/InjectStatDb/InjectHttpTransport` 注入。
- 平台配置变更 Hook 里，**只有 Cloudflare 会按 `capability` 决定建 zone 还是 tunnel 任务**，EdgeOne/ESA 无条件建任务。
- 日统计写入统一用 `CreateBulk + OnConflictColumns(...).UpdateNewValues`，配合 5.3 的唯一索引实现幂等回填。

### 6.9 厂商支持矩阵（实际接线情况）

枚举里有 16 个平台（`schematype/enum.go:80-104`），但任务 `switch` 只接了一部分：

| 能力     | 任务里已接线                                                                       | 有适配器但未接线                                 |
|----------|------------------------------------------------------------------------------------|--------------------------------------------------|
| DNS 同步 | cloudflare、tencent_dns、aliyun_dns、baidu_dns、huawei_dns、volcengine_dns         | jdcloud_dns                                      |
| CDN 同步 | tencent_cdn、tencent_eo、aliyun_cdn、aliyun_esa、baidu_cdn、huawei_cdn、qiniu_cdn  | jdcloud_cdn、ksyun_cdn；upyun_cdn 只有枚举无实现 |
| TUN 同步 | cloudflare                                                                         | —                                                |
| SSL 申请 | cloudflare、tencent_dns、tencent_eo、aliyun_dns、aliyun_esa、baidu_dns、huawei_dns | volcengine_dns、rainyun、westcn                  |

前端下拉与表单来自枚举 + `EnumRegistry`（见[第 11 节](#11-前端与模板约定)），**用户能选到尚未接线的平台**，选中后任务返回“暂不支持该平台”。

---

## 7. 调度与任务模型

### 7.1 `cron_job`：任务事实源

核心字段（[cronjob.go](../internal/ent/schema/cronjob.go)）：

| 字段                                                                               | 说明                                                |
|------------------------------------------------------------------------------------|-----------------------------------------------------|
| `entry_id`                                                                         | UUID v7，gocron 任务标识，唯一且不可变              |
| `handler_name`                                                                     | 任务处理器名，对应 `provider.TaskTask` 注册表的 key |
| `handler_args`                                                                     | JSON 参数，通常为 `{id}` 或平台/站点 ID             |
| `task_type`                                                                        | `one_off` 一次性 / `recurring` 周期                 |
| `job_type`                                                                         | `cron` / `duration` / `random`                      |
| `job_category`                                                                     | `system` 系统内置 / `custom` 用户自定义             |
| `business_type` + `business_id` / `business_key`                                   | 任务归属的业务对象                                  |
| `cron_expr` / `interval_value` / `min_interval` / `max_interval` / `interval_unit` | 调度参数                                            |
| `status`                                                                           | `running` / `stopped` / `finished` / `failed`       |
| `is_singleton`                                                                     | 是否单例执行（`LimitModeReschedule`）               |
| `run_on_start`                                                                     | 是否立即开始（`WithStartImmediately`）              |

调度参数的默认值写在**各任务文件的 `AddTask`/`LoadTask`** 里，不在 Schema 中。

### 7.2 执行链

1. `InitHandlers` 把 11 个 `TaskProvider` 注册到 `provider.TaskTask`（[cron_service.go](../internal/manager/cron_manager/cron_service.go)）。
2. `LoadAllJobs` 只读取 `status=running` 的任务，按 `job_type` 创建 gocron Job，以 `entry_id` 作为任务标识。
3. **配置指纹** = `job_type, task_type, interval_value, interval_unit, min_interval, max_interval, cron_expr, handler_args, run_on_start, is_singleton`；只有指纹变化才会重建任务，避免无意义重启。
4. 触发时 `RunHandler` 写一条 `cron_job_log`（Running），从注册表取处理器执行 `Process(ctx, args, meta)`，回填耗时、输出、Success/Failed。**`RunHandler` 从不回写 `cron_job.status`**，任务失败在任务表上看不出来。
5. 一次性任务在 `AfterJobRuns` 中被置为 `finished` 并从调度器移除；panic 监听器把状态置为 `failed`——**两者都只对 `one_off` 生效，周期任务 panic 不会改状态**。
6. 任务表的新增/删除通过 `CronJobGocronHook` 异步调用 `SyncJob`（**该 Hook 不挂 Update**，直接改表字段不会立刻反映到调度器）；平台配置变化通过 `PlatformConfigGocronHook`（含 Update）维护平台同步任务。

### 7.3 内置任务清单

| `handler_name`           | 作用                 | 调度                                   | 任务来源                                       |
|--------------------------|----------------------|----------------------------------------|------------------------------------------------|
| `dns_sync`               | 单条 DNS 记录同步    | 10–30 分钟随机，周期/系统              | DNS 记录创建/更新/启停                         |
| `cdn_sync`               | 单条 CDN 记录同步    | 10–30 分钟随机，周期/系统              | CDN 记录创建/更新/启停                         |
| `tun_sync`               | 单条 TUN 记录同步    | 10–30 分钟随机，周期/系统              | **当前未注册、未创建任务**                     |
| `domain_monitor`         | 域名可用性与证书探测 | `Duration = check_interval` 秒         | 域名监控创建/更新                              |
| `ssl_apply`              | 证书申请/续期        | 8–10 小时随机，周期/系统               | SSL 记录创建/更新                              |
| `ssl_refresh`            | 刷新证书元数据       | Cron `0 0 * * *`                       | 启动时自动为 `ssl_record`、`ssl_import` 建任务 |
| `cloudflare_tunnel_sync` | 同步 Cloudflare 隧道 | 1–2 小时随机                           | 平台配置 Hook                                  |
| `cloudflare_zone_sync`   | 同步 Cloudflare 站点 | 1–2 小时随机                           | 平台配置 Hook                                  |
| `edgeone_zone_sync`      | 同步 EdgeOne 站点    | 1–2 小时随机                           | 平台配置 Hook                                  |
| `esa_site_sync`          | 同步 ESA 站点        | 1–2 小时随机                           | 平台配置 Hook                                  |
| `clear_data`             | 按表清理过期日志     | Cron `5 0 * * *`，`run_on_start=false` | 启动时自动为 8 张日志表建任务                  |
| `webhook_send`           | 发送通知             | 一次性，1 秒后立即执行，单例           | 事件触发                                       |

手工同步（Cloudflare/EdgeOne/ESA/域名监控/SSL 申请）会创建 `Duration=1s` 的一次性任务立即执行，不改变周期任务的配置。

---

## 8. 厂商适配层

### 8.1 接口

| 接口           | 文件                                                      | 方法                                           | 使用方               |
|----------------|-----------------------------------------------------------|------------------------------------------------|----------------------|
| `DNSProvider`  | [dns_provider.go](../internal/provider/dns_provider.go)   | `Name`、`Check`、`Update`、`ParseCustomConfig` | `dns_sync`           |
| `CDNProvider`  | [cdn_provide.go](../internal/provider/cdn_provide.go)     | `Name`、`Check`、`Update`、`ParseCustomConfig` | `cdn_sync`           |
| `TUNProvider`  | [tun_provider.go](../internal/provider/tun_provider.go)   | `Name`、`Check`、`Update`、`ParseCustomConfig` | `tun_sync`（未接线） |
| `TaskProvider` | [task_provider.go](../internal/provider/task_provider.go) | `Process(ctx, args, meta)`                     | `cron_manager`       |
| `SSLProvider`  | [provider.go](../internal/provider/provider.go)           | `Renew`、`Present`、`CleanUp`、`Name`          | 保留接口，无实现     |

[unimplemented.go](../internal/provider/unimplemented.go) 为 DNS/CDN/TUN/Task 四类提供 `UnimplementedXxxProvider` 安全默认实现，具体实现首行嵌入它，未实现的方法返回明确错误。**`SSLProvider` 没有对应的 Unimplemented**。

### 8.2 配置分两层

1. **平台配置** `platform_config.config_json`：账号级凭据，如 AccessKey、API Token、Account ID、Endpoint。
2. **记录配置** `dns_record.config_json` / `cdn_record.config_json` / `tun_record.config_json`：记录级参数，如线路、Zone、Tunnel ID。

任务执行时先解析平台配置构造 Provider，再在 `Check/Update` 前解析记录配置。

### 8.3 注册与选择

- 任务内按 `platform_config.platform` 做 `switch`，构造对应 Provider；未知平台返回“暂不支持该平台”。
- 任务处理器通过 `TaskTask.RegisterTask(name, handler)` 注册，name 与 `cron_job.handler_name` 一一对应。
- HTTP 客户端统一接入 `http_log.LoggingRoundTripper`，出站请求会写 `http_log` / `resty_log`（取决于使用的客户端），由各包的 `inject.go` 用 `sync.Once` 注入。

### 8.4 新增厂商的步骤

1. 在 `internal/provider/{dns,cdn,tun}` 下新增 `xxx_provider.go`，实现对应接口并嵌入 `Unimplemented*`，同时定义 `opts.go` 里的选项。
2. 定义配置结构体与 `ParseCustomConfig`，字段与 `platform_config.config_json` 对齐。
3. 在目标任务（`dns-sync_task.go` / `cdn-sync_task.go` / `tun-sync_task.go` / `ssl-apply_task.go`）的 `switch platform.Platform` 中增加分支——**这一步不能省**，只加 provider 不会被调用。
4. 在 [schematype/enum.go](../internal/ent/schema/schematype/enum.go) 补 `ConfigPlatform*` 常量与枚举值，并在 `internal/handlers/api.schemas.go` + `internal/utils` 补该平台的表单 Schema。
5. 在 `internal/gin_server/resource.go` 的 `EnumRegistry` 补中文描述，否则前端下拉不显示（见[第 11 节](#11-前端与模板约定)）。
6. `internal/provider/{dns,cdn,tun}/inject.go` 只暴露 `InjectHttpTransport`，新增厂商一般无需改动。

---

## 9. 配置、构建与部署形态

### 9.1 配置只有两层

| 层         | 载体                                                                             | 读取时机                                          | 内容                                                                                                                   |
|------------|----------------------------------------------------------------------------------|---------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| 进程配置   | 环境变量，[sys_env/env.go](../internal/helper/sys_env/env.go) 的 `init()`        | 进程启动时一次性，不可热更                        | `DEBUG`、`ENABLE_LOG`、`SERVER_ADDR`、`BACKEND_ENDPOINT`、`BASIC_AUTH_USERNAME/PASSWORD`、`DEVICE_CODE`、`DEVICE_NAME` |
| 运行时设置 | 业务库 `cache` 表（`conf:` 前缀），[db_cache](../internal/helper/db_cache/db.go) | 启动载入 `GvaSetting`，`PATCH /api/settings` 可改 | 离线判定、IP 上报区间、转发规则刷新、统计上报区间                                                                      |

- 没有配置框架：不用 viper/godotenv，没有配置文件；`.env` 只出现在 `.gitignore`。所有默认值硬编码在 `env.go:47-54`。
- 数据库路径、日志目录同样硬编码，**换工作目录等于换库**。
- `DEBUG` 决定 Gin/Ent/gocron/HTTP 级别；`global.Version == "dev"` 同时决定模板/静态资源来源与日志轮转档位（一个变量三处语义）。
- `sys_env` 的 `init()` 用 `fmt.Printf("sys_env:%+v\n", GvaEnv)` 打印整个结构体，**`BASIC_AUTH_PASSWORD` 会以明文进 stdout / 容器日志**（见[第 15 节](#15-已知约束与风险)）。

### 9.2 版本注入与构建产物

- ldflags 只注入 `vibe_ip/internal/global` 的 `Version` / `GitCommit` / `BuildTime`（[Makefile](../Makefile)）。`cmd/reporter/main.go`、`cmd/forwarder/main.go` 里另有一套同名的局部变量，既不被 ldflags 注入也没有使用，日志实际打印的是 `global.Version`。
- 统一构建命令：`CGO_ENABLED=0`、`-trimpath`、`-buildvcs=false`（避免构建机缺 git 元数据时失败）。
- `build_binary` 用 `define` 展开 6 平台矩阵 → `bin/{name}-{os}-{arch}[.exe]` → `archive_binary` 打成 `.tar.gz`（unix）/`.zip`（windows）→ `clean-binaries` 只在存在压缩包时删除裸二进制。
- `ENABLE_UPX` 与 `compress_file` 已定义，但除 `archive_binary` 的调用外没有实际生效路径，UPX 压缩接近死配置。
- 常用目标：`deps`、`format`、`check`、`build-*`、`cover-*`、`junit-*`、`generate_ent`、`ent_new` / `ent_new_log`、`archive_binary`、`clean-binaries`。口径与 CI 完全一致，见 [ci-cd.md](ci-cd.md)。
- `scripts/`：`format.sh`（yaml/json/sql/shell/markdown/docker 格式化，工具缺失静默跳过）、`lib.sh`（`now_ms`/`run_step` 计时与跨平台降级链）、`ci/resolve-prerelease.sh`（预发布版本标记）；项目检查（golangci-lint 多平台 + 前端 audit/lint）在 `Makefile` 的 `check-*` 目标中直接执行，工具缺失即失败，不自动安装。
- 没有 `.goreleaser`；发布 = Makefile + GitHub/CNB 的 `release.yml`。没有 self-update 机制，节点升级靠替换二进制并重启。

### 9.3 代码生成有两条分叉路径（重要）

| 命令                | 覆盖范围                                                                | 是否生成审计 Hook                                                           |
|---------------------|-------------------------------------------------------------------------|-----------------------------------------------------------------------------|
| `go run ./cmd/ent`  | 业务/日志/统计三个库 + 自定义模板（aggregate、pagination、where_if 等） | **是**，渲染 `internal/ent/template/audit.tmpl` → `schemahook/audit_gen.go` |
| `make generate_ent` | 只有 `./internal/ent/schema`                                            | **否**，且 feature 列表与 `cmd/ent` 各自维护                                |

两者都**没有启用 Ent 的 `privacy` feature**（`Makefile:72`、`cmd/ent/main.go:38` 均被注释）。改完 Schema 必须用 `go run ./cmd/ent`，否则审计 Hook 会缺失或过期。

### 9.4 目录身份澄清

- `sql/`：手工 DDL 快照 + 导出 JSON 夹具，无代码引用，不参与迁移（见 5.6）。
- 根 `tests/`：被 `.gitignore` 忽略（`.gitignore:13`），是本地手工探针，**不在仓库里也不参与 CI**；真正的测试是 `internal/**/*_test.go` 与 `cmd/frontend/test/`，见 [coverage-go.md](coverage-go.md)、[coverage-node.md](coverage-node.md)。
- `.cnb/`（主 CI + `docker/Dockerfile.{backend,reporter,forwarder}`；CI 环境统一用 `docker.cnb.cool/dtapp/golang`）、`.github/`（等价流水线 + codecov/dependabot/release）、`.gitlab/`（预留目录，无流水线文件）。

### 9.5 部署形态与设备码链路

- 容器编排在 `cmd/docker-compose.yml` 与 `cmd/{backend,reporter,forwarder}/docker-compose.yml`：使用 `network_mode: host`，backend 用 `GET /health` 做健康检查。
- **节点容器必须只读挂载 `/etc/machine-id`**：`DEVICE_CODE` 为空时取 `machineid.ID()`（[core/device.go](../internal/core/device.go)），拿不到稳定 machine-id 会导致每次启动生成新设备码、库里堆积幽灵设备。前端模板函数 `selfDeviceCode` 也走同一条链路。
- 出站探测依赖外部服务：IP 探测使用 15 个 IPv4 + 13 个 IPv6 回显地址，含自建镜像域名（`eo-ip.*` / `cf-ip.*` / `esa-ip.dtapp.net`）与 `ddns.oray.com`、`myip.ipip.net`、`4.ipw.cn`、`ip.3322.net` 等，`errgroup` 并发取首个成功，5 秒超时，**`InsecureSkipVerify: true`**（[ip_helper](../internal/helper/ip_helper/ip_helper.go)）。

---

## 10. 日志与可观测性

### 10.1 文件日志

- backend 调 8 个 `Init*SLog`，实际产出 **10 个命名 logger**（`InitEntSLog` 一次建 ent/ent_log/ent_stat 三个）。
- 文件名规则 `{前缀}_{进程}.log`：`slog_`、`ent_`、`ent_log_`、`ent_stat_`、`gin_`、`grpc_`、`gocron_`、`mcp_`、`http_`、`resty_`，因此每个进程各有一套。
- 轮转按 `global.Version == "dev"` 分档（[initialize/slog.go:28-46](../internal/initialize/slog.go)）：dev 3 MB / 1 天 / 5 份；生产 2 MB / 15 天 / 150 份；均 `Compress: true`、`LocalTime: true`。
- handler 形态由外部 `golog` 决定：配置 lumberjack 时是 `io.MultiWriter(os.Stdout, 文件)`（stdout 与文件同时写），`TextHandler` 而非 JSON；只有主 logger `InitSLog` 会 `slog.SetDefault`。
- 级别由 `DEBUG` 决定（Debug/Info）；`getShowLine()` 被硬编码 `return false`，**日志里不会带源码行号**。
- `EnsureDefaultSLog` 把未初始化的 logger 回落成 `GvaSlog`：reporter/forwarder 只初始化了部分 logger，缺的不会 panic，但内容会混进 `slog_reporter.log` 而不是独立文件。
- `logs/req_backend.log`、`logs/thrift_*.log` 是已删除子系统留下的残留文件，不要据此判断还有 thrift/req 逻辑。

### 10.2 出站请求日志（重点风险）

[initialize/http_log.go:137-165](../internal/initialize/http_log.go) 做三件事：

1. 用 `http_log.LoggingRoundTripper` **包住 `http.DefaultTransport` 与 `http.DefaultClient`** —— 进程内任意出站 HTTP 都会被记录。
2. 额外把 transport 注入 service、syncs、cdn、dns、task、tun 六个包。
3. resty 侧另接 `resty_log`。

行为边界：

- `ENABLE_LOG` 只在 http 侧的两个写入点判断（saver 与 callback），**关闭时抓包与克隆 body 的开销仍然发生**；resty 侧完全没有 `ENABLE_LOG` 判断，`resty_log` 表无条件写入。
- 写入是无背压 `go func(...)` + `context.WithoutCancel(ctx)`，失败只打错误日志；`http_log`、`resty_log` 都保存**完整请求头**，第三方 API Token、`Authorization: Basic ...` 原文进 `ent_log.db`，无掩码、无 body 大小上限、无采样。
- `req_log` 表没有任何写入点（见 5.3）。

### 10.3 排障入口

| 现象            | 先看                                                                   |
|-----------------|------------------------------------------------------------------------|
| 任务没跑/没生效 | `cron_job.status` + `cron_job_log`（失败信息只在这里）                 |
| 厂商 API 报错   | `resty_log` / `http_log` 的 response body，或 `logs/resty_backend.log` |
| 设备显示离线    | `device.last_heartbeat_at`（由 IP 上报写，不是 gRPC 心跳）             |
| 谁改了配置      | `operation_log`（注意来源恒为 `user`，见 4.2）                         |
| 通知没发出      | `webhook_history` 的 `response`（`is_success` 对手工测试不可信）       |
| 进程启动失败    | stdout（`sys_env` 打印 + panic 兜底栈），以及 `logs/*_{进程}.log`      |

---

## 11. 前端与模板约定

没有 SPA 框架，页面是 Go `html/template` 服务端渲染 + 每页一个经典 `<script>`。

### 11.1 模板结构与命名约定

- `cmd/backend/templates/layout/base.html`（1 个）、`pages/*.html`（77 个）、`components/*.html`（5 个：icons、pagination、popup、setting_form、switch）。
- 多模板引擎（multitemplate）按 `pages/{base}_{后缀}.html` 规则把片段合并进主页面，识别的后缀是 `_form`、`_filter`、`_detail`、`_view`、`_history`、`_stats`（[gin_server/resource.go:656-700](../internal/gin_server/resource.go)）。**新建页面必须遵守这套后缀**，否则片段不会被加载。
- `resource.go` 的 `getTemplateFuncMap` 注册 **36 个模板函数**，常用的有 `isDebug`、`deviceOfflineEnabled`、`formatTimeOrNil`、`diffForHumans`、`formatBytes`、`formatBps`、`formatNumber`、`formatMs`、`jsonify`、`jsonGet`（gjson）、`dict`、`fieldEnumMap/fieldEnumValues/fieldEnumLabel`、`selfDeviceCode`、`getCacheKey`、`parseRecordTTL`、`add/sub/div`、`toInt…toUint64`。注意 `replace` 直接转发 `strings.Replace(old,new,-1)` 的四参签名。
- `resource.go` 里还维护一张 **`EnumRegistry`（51 个枚举键）**，把 `schematype` 枚举值映射到中文描述，是页面下拉框与 `fieldEnumLabel` 的唯一来源。**新增枚举必须同时改 `schematype/enum.go` 和 `EnumRegistry`**。

### 11.2 静态资源构建

- `cmd/frontend` 只做资源组装：`build.sh` / `dev.sh` 先清空 `cmd/backend/static/libs/`，从 `node_modules` 复制 7 个库（loadjs、jquery、clipboard、select2 + select2 的 Tailwind v4 主题、echarts、dayjs 及 relativeTime/zh-cn），再执行 `tailwindcss -i css/input.css -o ../backend/static/css/style.css --minify`（dev 模式 `--watch --poll`）。
- `cmd/backend/static/libs/` 是**构建产物**，不是源码库；版本只在 `cmd/frontend/package.json`。
- Tailwind v4 采用 CSS-first 配置，没有 `tailwind.config.js`：`css/input.css` 里 `@source` 扫描 `backend/templates` 与 `static/js`，`@plugin "daisyui" { darktheme: "auto" }`，`@plugin "@egoist/tailwindcss-icons"` + lucide 图标集。

### 11.3 页面数据加载

- `base.html` 固定注入 `api.js`、`loadjs`、`load.js`、`index.js`；页面自身脚本按需请求 `/api/...`，响应是统一信封（见[第 12 节](#12-横切开发约定)）。
- `load.js` 维护 `CDN_CONFIG`：每个库配置 6–7 个候选源，按“本地 `/static/libs` → 自建镜像（`esa-*` / `eo-*` 域名）→ cdnjs/jsdelivr”逐级回退。**自建镜像域名是隐性外部依赖**，离线部署时首屏加载会走完整回退链。
- `api.js` 的 `post`/`patch` 分支里 `info = {...}` 少了 `let/const`，是隐式全局变量（并发请求会互相覆盖）。
- 没有 i18n 机制，文案硬编码中文；只有 dayjs 的 `zh-cn` 与 select2 的 `i18n` 语言包被动 vendored。

---

## 12. 横切开发约定

### 12.1 REST 响应与错误

- 统一信封 `types.APIResponse{Code, Message, Data}`：`successResponse` 固定 `code=200`、`message="success"`；`errorResponse` 让 **HTTP 状态码与 `code` 同值**，`message` 直接是 `err.Error()` 原文——没有业务错误码表，内部错误文案会透给前端（[handlers/common.go:18-31](../internal/handlers/common.go)）。
- 未匹配路由/方法返回同一信封的 404/405；但 panic 被 `Recovery` 拦下时返回**裸 500、无 body**，前端只能靠 HTTP 状态码判断。
- handler 里普遍是 `service.NewXxxService(global.GvaEntDb, global.GvaEntLogDb, global.GvaGinSlog)` 现场构造，没有 DI 容器；包级依赖走 `InjectDb`/`InjectHttpTransport`。

### 12.2 入参绑定、分页与三态布尔

- 绑定统一用 [gin_helper](../internal/helper/gin_helper/gin_param.go) 的 `BindQueryWithDefaults` / `BindJSONWithDefaults` / `BindUriWithDefaults`：先反射填充结构体 `default:"..."` tag 再绑定（Gin 原生不支持 struct tag 默认值）；JSON 只在 `ContentLength > 0 && Content-Type` 含 `application/json` 时绑定。**项目没有使用 `validator` tag，也没有自定义校验层**，必填/范围校验靠 handler 自查。
- 分页约定：`CommonPageRequest{page=1, page_size=10}` + 泛型 `CommonPageResponse{filter, list, page, page_size, total}`（另有 `WithItem`、`WithItemAndStat` 变体）。隐性规则：URL 未传或 `page_size <= 10` 时会被 `pref_page_size` **Cookie 覆盖**。
- `NullableBool` 实现 `UnmarshalJSON`/`UnmarshalText`/`UnmarshalParam` 三套入口，用来区分“未传 / null / false”，PATCH 语义依赖它。

### 12.3 审计写路径

`audit_gen.go` 为 28 个实体生成 `hook.On`，每次 Create/Update/Delete **各起一个无上限 goroutine** 写 `operation_log`，脱离请求 Context。没有队列、没有批量、关停时不 drain —— `CloseEntLogDb` 立即执行，在途审计会被静默丢弃；高并发时这些写入还会与业务写争抢 `MaxOpenConns(1)` 的同一连接。

### 12.4 敏感数据现状

- Ent 的 `Sensitive()` 只影响生成代码的 `String()`；privacy feature 未启用 ⇒ **`platform_config.config_json`、`webhook.secret`、SSL 私钥都会原样出现在 REST 响应里**，唯一的“防护”是 Coraza 里一条 `pass`（不拦截）的正则。
- 出站请求头原文写进 `ent_log.db`（见 10.2）。
- 二次开发新增接口时，涉及凭据的响应要手工屏蔽，不要假设框架会处理。

---

## 13. 模块职责速查

| 目录                                                               | 职责                                                                    | 关键文件                                                        |
|--------------------------------------------------------------------|-------------------------------------------------------------------------|-----------------------------------------------------------------|
| `cmd/backend`                                                      | 服务端入口，embed 模板/静态资源，docker-compose 样例                    | `main.go`、`docker-compose.yml`                                 |
| `cmd/reporter`                                                     | IP 上报与设备长连接                                                     | `main.go`                                                       |
| `cmd/forwarder`                                                    | 端口转发与统计上报                                                      | `main.go`                                                       |
| `cmd/frontend`                                                     | 前端资产构建（复制 libs + Tailwind 编译），无应用源码                   | `package.json`、`build.sh`、`test/`                             |
| `cmd/ent`                                                          | 三个库的 Ent 代码与审计 Hook 生成器                                     | `main.go`                                                       |
| `internal/ent` \| `ent_log` \| `ent_stat`                          | 三个库的 Schema 与生成代码                                              | `schema/`                                                       |
| `internal/initialize`                                              | 日志、三库、gocron、HTTP/Resty 日志初始化                               | `slog.go`、`ent.go`、`gocron.go`、`http_log.go`、`resty_log.go` |
| `internal/global`                                                  | 全局单例（logger、三库、gocron、版本信息）                              | `global.go`                                                     |
| `internal/helper/sys_env`                                          | 环境变量读取与默认值（唯一配置入口）                                    | `env.go`                                                        |
| `internal/helper/db_cache`                                         | 运行时设置缓存 `GvaSetting`                                             | `db.go`                                                         |
| `internal/manager/mixed_manager`                                   | Gin + gRPC h2c 混合服务                                                 | `mixed_server.go`                                               |
| `internal/manager/cron_manager`                                    | 任务注册、加载、同步、执行、日志                                        | `cron_service.go`                                               |
| `internal/gin_server`                                              | Gin 初始化、模板/静态资源、路由、Coraza 配置                            | `server.go`、`router.go`、`resource.go`、`coraza.go`            |
| `internal/gin_middleware`                                          | 认证、审计、WAF、CORS、Gzip、Recovery、RequestID                        | 各中间件文件                                                    |
| `internal/grpc_server` / `grpc_service` / `grpc_interceptor`       | gRPC 引擎、三个服务、审计/日志/认证拦截器                               | `server.go` 等                                                  |
| `internal/grpc_proto` / `grpc_generated`                           | proto 定义与生成代码（含未被挂载的 gateway）                            | `*.proto`、`*.pb.go`                                            |
| `internal/grpc_client` / `grpc_credential` / `grpc_gateway_client` | 节点侧客户端、BasicAuth 凭据、gateway 客户端（未接入）                  | `client.go`、`basic_auth.go`                                    |
| `internal/handlers`                                                | 页面与 REST handler、`APIResponse` 封装、Schema 表单下发                | `api.*.go`、`common.go`、`api.schemas.go`                       |
| `internal/types`                                                   | API 请求/响应 DTO、分页与 `NullableBool`                                | `common.go`、`api.*.go`                                         |
| `internal/backend/service`                                         | 业务编排、事务、CRUD                                                    | `*_service.go`                                                  |
| `internal/backend/syncs`                                           | Cloudflare/EdgeOne/ESA 全量同步                                         | `*_sync.go`、`inject.go`                                        |
| `internal/provider`                                                | 厂商接口、Unimplemented 默认、任务注册表                                | `*_provider.go`、`unimplemented.go`                             |
| `internal/provider/{dns,cdn,tun}`                                  | 厂商实现与 transport 注入                                               | `*_dns.go`、`*_cdn.go`、`cloudflare_tun.go`                     |
| `internal/provider/task`                                           | 定时任务实现（含调度默认值）                                            | `*_task.go`                                                     |
| `internal/helper/*`                                                | IP 探测、域名检测、SSL(lego)、Webhook、gocron 辅助、 Gin 参数绑定、日志 | `ip_helper`、`domain_helper`、`ssl_helper`、`gin_helper`        |
| `internal/core`                                                    | 无状态领域函数（设备标识、IP 探测、BasicAuth token）                    | `device.go`、`ip.go`、`token.go`                                |
| `internal/mcp_server` / `mcp_tool` / `mcp_resource` / `mcp_prompt` | MCP 服务与按域注册的工具/资源/提示词                                    | `server.go` 等                                                  |
| `internal/reporter` / `internal/forwarder`                         | 节点侧服务与任务、转发数据面                                            | `service/`、`tasks/`、`forward/`                                |
| `internal/docs_2` / `docs_3`                                       | Swagger OpenAPI 2 / 3.1 数据                                            | `swagger.json`、`docs.go`                                       |
| `internal/utils` / `pkg/utils`                                     | DNS/域名/SSL/MCP schema 辅助、文件工具                                  | `mcp_schema.go`、`file.go`                                      |
| `Makefile` / `scripts`                                             | 构建、格式化、检查、发布打包                                            | `format.sh`、`lib.sh`、`ci/*.sh`                                |
| `sql`                                                              | 手工 DDL 快照与导出夹具（不参与迁移）                                   | `data_ddl.sql`、`log_ddl.sql`                                   |
| `logs` / `data` / `bin`                                            | 运行期产物（日志、SQLite、构建输出），均在 `.gitignore`                 | —                                                               |

---

## 14. 扩展指南

### 14.1 新增业务实体

1. 在 `internal/ent/schema/` 新增 Schema，定义表名注解、字段、索引、边、Mixin、`Sensitive()`。
2. **用 `go run ./cmd/ent` 生成**代码与审计 Hook（`make generate_ent` 不会生成审计 Hook，见 9.3）。
3. 在 `internal/types` 增加请求/响应 DTO。
4. 在 `internal/backend/service` 增加 service，承载事务与业务规则。
5. 在 `internal/handlers` 增加页面/API handler，并在 `gin_server/router.go` 注册路由。
6. 如需通知，在 `schematype` 的 Webhook 对象类型中增加枚举，并在 `WebhookSendTaskProvider` 中补映射。
7. 如需 MCP 能力，在 `mcp_tool` / `mcp_resource` / `mcp_prompt` 增加注册函数，并在 `mcp_server.Run` 中调用。

### 14.2 新增定时任务

1. 在 `internal/provider/task/` 新建 `xxx_task.go`，实现 `TaskProvider.Process`，首行嵌入 `UnimplementedTaskProvider`，并把调度默认值写在 `AddTask`/`LoadTask`。
2. 定义 `XxxTaskName` 常量，作为 `cron_job.handler_name`。
3. 在 `cron_manager.InitHandlers` 中 `RegisterTask`——**漏这一步任务永远不跑**（`tun_sync` 就是现状示例）。
4. 如需随业务记录自动建任务，实现 `LoadTask` / `AddTask`，并在对应 service 的创建/更新/启停路径调用。
5. 如需随平台配置自动建任务，在 `SyncPlatformConfigJob` 中增加分支。
6. 任务参数写入 `handler_args`，执行时通过 `meta` 拿到业务类型/ID/分类。
7. 记得给新任务补清理策略：已完成的一次性任务行目前不会被回收（见 15）。

### 14.3 新增 REST API

1. 在 `internal/types` 定义请求/响应结构，用 `default:"..."` tag 声明默认值（Gin 原生不认，靠 `gin_helper`）。
2. 在 `internal/backend/service` 实现业务方法。
3. 在 `internal/handlers/api.*.go` 用 `Bind*WithDefaults` 绑定、调用 service、返回 `types.APIResponse`；返回前确认没有夹带凭据。
4. 在 `gin_server/router.go` 的 `/api` 分组中注册路由。
5. 运行 Swagger 生成命令更新 `internal/docs_2`、`internal/docs_3`。

### 14.4 新增控制台页面

1. 在 `cmd/backend/templates/pages/` 新建页面模板，遵守 `_{form,filter,detail,view,history,stats}` 后缀约定，复用 `layout/base.html` 与 `components/`。
2. 在 `internal/handlers` 增加页面 handler，构造 `types.BasePageData[T]`。
3. 在 `gin_server/router.go` 的 `setupWebRouter` 注册路径。
4. 在 `cmd/backend/static/js/` 增加页面脚本（经典 `<script>`，非 ESM）；生产环境走 embed，本地 `Version=dev` 走磁盘。
5. 若新增了枚举值，同步 `schematype/enum.go` 与 `EnumRegistry`，再 `pnpm --dir cmd/frontend build` 重编译 CSS（新 class 才会被 Tailwind 扫描到）。

### 14.5 新增 MCP 工具

1. 在 `mcp_tool` 定义入参/出参结构体并注册工具；schema 可由 `utils.AutoSchema` 反射生成。
2. 如需要资源模板，在 `mcp_resource` 注册 URI 模板；如需要提示词，在 `mcp_prompt` 注册。
3. 在 `mcp_server.Run` 的对应业务域中调用注册函数。
4. 工具内部直接复用 backend service 与全局 DB，不复制 REST 逻辑。

---

## 15. 已知约束与风险

以下是当前代码的客观现状，部署或二次开发前需要评估：

**部署与并发**

1. **只能单实例部署**：设备连接表 `clients`、运行中任务、各 Manager 单例都在进程内存；多副本会重复调度、设备连接分散。
2. **SQLite 单写**：三个库都 `MaxOpenConns(1)`，写入串行；DSN 与 `data/*.db` 路径硬编码、相对工作目录。
3. **启动自动迁移无版本管理**：`Schema.Create` 只做增量补列，复杂变更（改类型、删列、数据回填）需手工处理；变更期间要避免多进程同时启动。
4. **入站没有读写超时**，Coraza 中间件用 `io.ReadAll` 无上限读请求体（`gin_middleware/coraza.go:77`）→ 慢连接与大 body 可耗尽内存。
5. **审计与日志写入无背压**：每次实体变更起一个 goroutine，出站 HTTP 日志同样异步写库，高并发下会放大连接争抢。

**安全**

6. **明文传输**：混合服务未配置 TLS，gRPC 使用 `insecure` 凭据；BasicAuth 在网络上可解码，建议置于反向代理/TLS 之后。
7. **认证粒度粗**：MCP、Swagger、REST 共用一组 BasicAuth；任一项为空即关闭认证。
8. **`/health` 与整棵 `/static` 不经认证**：它们在 `router.Use(BasicAuth)` 之前注册。
9. **启动横幅泄露密码**：`sys_env` 的 `init()` 打印整个配置结构体，`BASIC_AUTH_PASSWORD` 明文进 stdout/容器日志。
10. **凭据落库与响应不脱敏**：`Sensitive()` 只影响 `String()`，privacy 未启用 → REST 响应含 `config_json`、`webhook.secret`、SSL 私钥；出站请求头（含 `Authorization`）原文写进 `ent_log.db`。
11. **CORS 过宽**：任意 Origin + `AllowCredentials=true`。
12. **WAF 实际未加载规则**：Coraza 只启用引擎与请求体读取，没有 Include OWASP CRS，拦截能力接近于零。
13. **gRPC 侧密码非常数时间比较**，与 HTTP 侧（`gin.BasicAuth`）实现不一致。
14. **IP 探测与证书采集跳过 TLS 校验**（`InsecureSkipVerify: true`），且探测点包含硬编码的第三方回显服务与自建镜像域名。

**协议与接线缺口**

15. **优雅关闭存在窗口**：混合服务 `Stop` 在 goroutine 中执行，主流程未等待完成就关库；MCP 没有停止步骤；在途审计写入会被丢弃。
16. **reporter/forwarder 没有信号处理**：`select{}` 常驻，`defer` 清理不执行，容器停止依赖 SIGKILL。
17. **MCP 路由注册时机**：`/mcp` 在 goroutine 中向 Gin 注册，与混合服务启动并发，理论上存在路由树并发读写风险；同时缺少 DELETE 路由，有状态会话无法正常结束。
18. **gRPC-Gateway 未挂载**：proto 注解与 `open.pb.gw.go` 存在，但 backend 没接入 gateway；README 中的 gateway 路径当前不可用。
19. **远程日志隧道未接线**：`RemoteLogService.LogTunnel` 服务端已实现，节点侧也定义了 `StartTunnel`，但主程序没有调用。
20. **TUN 同步未接线**：处理器未注册、service 未建任务，详见 [6.3](#63-tun-同步当前未接线)。
21. **审计来源不准确**：`AuditContextMiddleware` 忽略入参恒写 `user`，MCP 侧审计也记成 `user`；`task`/`gin`/`api` 枚举定义了但没有写入点；定时任务变更落默认 `user`。
22. **设备组取源站未实现**：DNS/CDN 记录可关联设备组，但同步任务只读取 IP 组。
23. **DNS 单值语义**：A/AAAA 只使用 IP 组内最后一个匹配设备的 IP；多设备多值解析需改用 CDN 多源站或拆记录。
24. **同一 ID 的转发规则变更不会热更新**：`Agent.Sync` 只按 ID 增删，改目标地址/协议需停用再启用或重启 forwarder；`tcpproxy` 只有 TCP 且无流量统计。
25. **平台枚举 > 已接线平台**：用户能选到 `jdcloud_cdn`、`ksyun_cdn`、`upyun_cdn` 等，任务执行时才报“暂不支持该平台”（见 6.9）。
26. **Webhook 映射缺陷**：`BusinessTypeSsl` 被映射到 `ObjTypeDomainMonitor`；`device_group`/`ip_group`/`ssl_record` 无触发路径；手工测试的历史记录 `is_success` 恒为 true。
27. **数据保留不完整**：`clear_data` 只覆盖 8 张日志表，`cron_job_log`、已完成的一次性 `cron_job` 行、统计库三张表都不清理；`req_log` 是空表。
28. **排查任务失败要看 `cron_job_log`**：`RunHandler` 不回写任务状态，周期任务 panic 也不置 `failed`。
29. **改 Schema 后 `make generate_ent` 会丢审计 Hook**：审计 Hook 只由 `go run ./cmd/ent` 生成（见 9.3）。
30. **监听端口不一致**：默认 `SERVER_ADDR=0.0.0.0:8080`，本地开发 Makefile 用 9999，排障先确认实际端口。

---

## 附录 A：环境变量与默认路径

| 变量                  | 默认值           | 说明                                                                |
|-----------------------|------------------|---------------------------------------------------------------------|
| `DEBUG`               | `false`          | 调试模式（Gin/Ent/gocron/HTTP 日志级别为 Debug）                    |
| `ENABLE_LOG`          | `false`          | 是否写入 `http_log` 表；**不影响 `resty_log`**，也不关闭抓包开销    |
| `SERVER_ADDR`         | `0.0.0.0:8080`   | backend HTTP/gRPC 混合监听地址                                      |
| `BACKEND_ENDPOINT`    | `127.0.0.1:8080` | reporter/forwarder 连接的 backend 地址                              |
| `BASIC_AUTH_USERNAME` | 空               | BasicAuth 用户名；为空关闭认证                                      |
| `BASIC_AUTH_PASSWORD` | 空               | BasicAuth 密码；为空关闭认证；启动时会明文打印到 stdout             |
| `DEVICE_CODE`         | 自动生成         | 设备编码；为空时用 `machineid.ID()`（容器需挂载 `/etc/machine-id`） |
| `DEVICE_NAME`         | 空               | 设备名称                                                            |

不可配置（源码内硬编码）：三个 SQLite 路径与 DSN pragma、`logs/` 目录与轮转档位、模板/静态资源路径、域名监控使用的上游 DNS `119.29.29.29`、IP 探测回显地址列表。

| 路径                                          | 说明                                                 |
|-----------------------------------------------|------------------------------------------------------|
| `data/ent.db`                                 | 业务库                                               |
| `data/ent_log.db`                             | 日志/审计库                                          |
| `data/ent_stat.db`                            | 统计库                                               |
| `logs/`                                       | 分类日志文件（lumberjack 轮转，`{前缀}_{进程}.log`） |
| `cmd/backend/templates`、`cmd/backend/static` | 开发模式下的模板与静态资源                           |
| `bin/`                                        | 构建产物与压缩包                                     |
| `internal/docs_2`、`internal/docs_3`          | Swagger JSON/YAML                                    |

运行时设置（`cache` 表，`conf:` 前缀，`PATCH /api/settings` 可改）：设备离线判定 12 分钟、IP 上报 `8-12` 分钟、端口转发规则刷新 30 分钟、端口转发统计上报 `8-12` 分钟。

---

## 附录 B：关键文件索引

| 关注点             | 文件                                                                                                                                                                                                                               |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| backend 启动/关闭  | [cmd/backend/main.go](../cmd/backend/main.go)                                                                                                                                                                                      |
| Gin+ gRPC 分流     | [internal/manager/mixed_manager/mixed_server.go](../internal/manager/mixed_manager/mixed_server.go)                                                                                                                                |
| HTTP 中间件        | [internal/gin_server/server.go](../internal/gin_server/server.go)、[internal/gin_middleware](../internal/gin_middleware)                                                                                                           |
| WAF/CORS/Gzip 配置 | [internal/gin_server/coraza.go](../internal/gin_server/coraza.go)、[internal/gin_middleware/cors.go](../internal/gin_middleware/cors.go)                                                                                           |
| 路由               | [internal/gin_server/router.go](../internal/gin_server/router.go)                                                                                                                                                                  |
| 模板函数与枚举注册 | [internal/gin_server/resource.go](../internal/gin_server/resource.go)                                                                                                                                                              |
| gRPC 注册/拦截器   | [internal/grpc_server/server.go](../internal/grpc_server/server.go)                                                                                                                                                                |
| 三库初始化/迁移    | [internal/initialize/ent.go](../internal/initialize/ent.go)                                                                                                                                                                        |
| 日志初始化与轮转   | [internal/initialize/slog.go](../internal/initialize/slog.go)                                                                                                                                                                      |
| 出站 HTTP 日志     | [internal/initialize/http_log.go](../internal/initialize/http_log.go)、[resty_log.go](../internal/initialize/resty_log.go)                                                                                                         |
| 环境变量读取       | [internal/helper/sys_env/env.go](../internal/helper/sys_env/env.go)                                                                                                                                                                |
| 业务 Schema        | [internal/ent/schema](../internal/ent/schema)                                                                                                                                                                                      |
| 日志 Schema        | [internal/ent_log/schema](../internal/ent_log/schema)                                                                                                                                                                              |
| 统计 Schema        | [internal/ent_stat/schema](../internal/ent_stat/schema)                                                                                                                                                                            |
| 审计 Hook 生成     | [internal/ent/template/audit.tmpl](../internal/ent/template/audit.tmpl)、[schemahook/audit_gen.go](../internal/ent/schema/schemahook/audit_gen.go)                                                                                 |
| 任务注册/调度      | [internal/manager/cron_manager/cron_service.go](../internal/manager/cron_manager/cron_service.go)                                                                                                                                  |
| 任务实现           | [internal/provider/task](../internal/provider/task)                                                                                                                                                                                |
| 厂商接口           | [dns_provider.go](../internal/provider/dns_provider.go)、[cdn_provide.go](../internal/provider/cdn_provide.go)、[tun_provider.go](../internal/provider/tun_provider.go)、[unimplemented.go](../internal/provider/unimplemented.go) |
| 设备 gRPC 服务     | [internal/grpc_service/device_service.go](../internal/grpc_service/device_service.go)                                                                                                                                              |
| 开放接口实现       | [internal/backend/service/open_service.go](../internal/backend/service/open_service.go)                                                                                                                                            |
| 平台全量同步       | [internal/backend/syncs](../internal/backend/syncs)                                                                                                                                                                                |
| 端口转发数据面     | [internal/forwarder/forward/forward.go](../internal/forwarder/forward/forward.go)                                                                                                                                                  |
| SSL 申请           | [internal/helper/ssl_helper](../internal/helper/ssl_helper)                                                                                                                                                                        |
| 域名探测           | [internal/helper/domain_helper/check.go](../internal/helper/domain_helper/check.go)                                                                                                                                                |
| IP 探测            | [internal/helper/ip_helper/ip_helper.go](../internal/helper/ip_helper/ip_helper.go)                                                                                                                                                |
| 请求绑定与分页     | [internal/helper/gin_helper/gin_param.go](../internal/helper/gin_helper/gin_param.go)、[internal/types/common.go](../internal/types/common.go)                                                                                     |
| MCP 注册           | [internal/mcp_server/server.go](../internal/mcp_server/server.go)                                                                                                                                                                  |
| 运行时设置         | [internal/helper/db_cache/db.go](../internal/helper/db_cache/db.go)                                                                                                                                                                |
| 构建与打包         | [Makefile](../Makefile)、[scripts](../scripts)、[cmd/docker-compose.yml](../cmd/docker-compose.yml)                                                                                                                                |
| 前端构建           | [cmd/frontend/build.sh](../cmd/frontend/build.sh)、[cmd/frontend/css/input.css](../cmd/frontend/css/input.css)、[cmd/backend/static/js/load.js](../cmd/backend/static/js/load.js)                                                  |
