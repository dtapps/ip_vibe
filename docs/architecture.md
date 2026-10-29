# IPVibe 架构与模块设计

> 面向维护者与二次开发者。本文说明代码如何分层、请求如何流转、数据存在哪里、定时任务如何运行，以及新增功能应该落在哪一层。功能清单、快速开始与部署方式见 [README](../README.md)；测试与覆盖率方案见 [coverage-go.md](coverage-go.md)、[coverage-node.md](coverage-node.md)。

## 目录

- [1. 系统概览](#1-系统概览)
- [2. 代码分层与依赖方向](#2-代码分层与依赖方向)
- [3. 进程与启动流程](#3-进程与启动流程)
- [4. 通信与接入层](#4-通信与接入层)
- [5. 数据层](#5-数据层)
- [6. 核心业务链路](#6-核心业务链路)
- [7. 调度与任务模型](#7-调度与任务模型)
- [8. 厂商适配层](#8-厂商适配层)
- [9. 模块职责速查](#9-模块职责速查)
- [10. 扩展指南](#10-扩展指南)
- [11. 已知约束与风险](#11-已知约束与风险)
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
2. **backend 单实例**：设备连接表、运行中的 gocron 任务、各 Manager 单例都只存在于进程内存，部署多个副本会出现重复调度与连接不一致。详见[第 11 节](#11-已知约束与风险)。
3. **三库分离**：业务数据、日志/审计、统计指标分别落在三个 SQLite 文件，避免大日志表拖慢业务查询。
4. **数据库驱动调度**：`cron_job` 表是任务事实源，gocron 只是执行器；任务增删改通过表数据 + Ent Hook 同步到调度器，重启后从表恢复。
5. **厂商适配器**：DNS/CDN/TUN 通过统一接口屏蔽厂商差异，业务任务只依赖接口，不直接调用云 SDK。
6. **单端口多协议**：Gin、gRPC、MCP 共用一个 h2c 端口，节点侧只需配置一个 `BACKEND_ENDPOINT`。
7. **拉模型同步**：IP 上报不直接改 DNS/CDN，而是更新设备表；DNS/CDN 任务周期性读取“心跳窗口内”的设备 IP 再同步，解耦上报与厂商 API。

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
2. 初始化 8 个分类 logger（主日志、Ent、Gin、gRPC、gocron、MCP、HTTP、Resty），并确保默认 logger 非空（`main.go:64-72`）。
3. 初始化三个 SQLite 库并自动建表，注入 Ent/审计 Hook/Service/Syncs 依赖（`main.go:74-76`，[initialize/ent.go](../internal/initialize/ent.go)）。
4. 初始化运行时设置 `db_cache`（缺省值就地写入）、gocron 调度器、HTTP/Resty 请求日志钩子（`main.go:77-81`）。
5. 创建 Gin/gRPC Manager；加载模板与静态资源；挂全局 BasicAuth；注册页面/API/Swagger 路由（`main.go:94-125`）。
6. 异步启动 MCP（向 Gin 注册 `/mcp`），异步启动 Gin+gRPC 混合服务（`main.go:128-142`）。
7. 启动 gocron，创建 `cron_manager`，注册 11 个任务处理器；加载系统任务（SSL 刷新、数据清理）；把 `cron_job` 中 `status=running` 的任务装入调度器（`main.go:145-174`）。
8. 把 `SyncJob` / `SyncPlatformConfigJob` 注入 Ent Hook，之后任务表的增删改才会自动同步到调度器（`main.go:177-180`）。
9. 等待 SIGINT/SIGTERM，执行优雅关闭（`main.go:187-249`）。

开发/生产的资源加载差异：`Version == "dev"` 时从 `cmd/backend/templates` 与 `cmd/backend/static` 读磁盘，否则读 `go:embed`（[gin_server/resource.go](../internal/gin_server/resource.go) `InitResource`）。版本号由 ldflags 注入。

关闭顺序：关闭所有设备流 → 异步关闭混合 HTTP 服务 → 停止 gocron → 关闭三库，整体 30 秒超时。注意混合服务关闭是 `go mixedMgr.Stop(ctx)`，并未等待完成就继续关库，见[第 11 节](#11-已知约束与风险)。

### 3.2 reporter

入口 [cmd/reporter/main.go](../cmd/reporter/main.go)：

1. 初始化日志与 gocron；`DEVICE_CODE` 为空时用 `machineid.ID()` 生成。
2. 用 `BACKEND_ENDPOINT` + BasicAuth 建立 gRPC 连接；未配置则直接 panic。
3. 创建 IP 上报任务：首次立即执行，之后按 backend 返回的 `IPReportIntervalTime`（默认 `8-12` 分钟随机）运行。
4. 启动设备长连接：首帧发送 `ClientInfo`，收到 `Ack` 后每 30 秒心跳、每 30 分钟重发全量信息；断线按 1→30 秒指数退避重连。
5. 主 goroutine `select {}` 常驻。

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
- 分流条件：`r.ProtoMajor == 2` 且 `Content-Type` 以 `application/grpc` 开头 → gRPC；否则 → Gin。
- HTTP/2：30 秒读空闲 PING、15 秒 PING 超时、最大并发流 100；连接空闲 120 秒。
- 未配置 TLS，节点与 backend 之间是明文 h2c，鉴权仅靠 BasicAuth。

由于是前缀匹配，`application/grpc-web` 也会进入 gRPC 分支。

### 4.2 HTTP 中间件顺序

注册位置 [gin_server/server.go](../internal/gin_server/server.go) `initServer`：

| 顺序 | 中间件                     | 作用                                            |
|------|----------------------------|-------------------------------------------------|
| 1    | `gin.Logger`               | 访问日志，写入 slog                             |
| 2    | `Recovery`                 | panic 恢复                                      |
| 3    | `AuditContext(SourceUser)` | 把审计来源写入 Context                          |
| 4    | `RequestID`                | 请求 ID                                         |
| 5    | `Coraza`                   | WAF 中间件                                      |
| 6    | `CORS`                     | 跨域                                            |
| 7    | `Gzip`                     | 响应压缩                                        |
| 8    | `BasicAuth`                | 全局认证，在 `main.go` 中追加，位于所有路由之前 |

注意：

- 源码中 CORS/Gzip 两处注释与实际注册顺序相反，实际顺序以上表为准。
- BasicAuth 是全局中间件，因此 `/health`、`/static`、`/swagger`、`/mcp` 与业务 API 使用同一套凭据。
- CORS 当前允许任意 Origin 且允许携带凭据；Coraza 只启用了引擎和请求体读取，没有加载 OWASP CRS 规则文件（见[第 11 节](#11-已知约束与风险)）。

### 4.3 gRPC 服务与拦截器

注册位置 [grpc_server/server.go](../internal/grpc_server/server.go)：

- 一元与流式拦截器顺序一致：审计 `SourceGrpc` → 日志 → BasicAuth。
- 收发消息上限 10 MB。
- 注册三个服务：

| 服务               | Proto                                                       | 方法                                                                              | 说明                                 |
|--------------------|-------------------------------------------------------------|-----------------------------------------------------------------------------------|--------------------------------------|
| `DeviceService`    | [device.proto](../internal/grpc_proto/device.proto)         | `Stream(stream ClientMessage) returns (stream ServerMessage)`                     | 设备注册/心跳长连接                  |
| `OpenService`      | [open.proto](../internal/grpc_proto/open.proto)             | `GetSetting`、`ReportDeviceIP`、`GetDevicePortForwards`、`CreatePortForwardStats` | 节点侧开放接口，均为一元调用         |
| `RemoteLogService` | [remote_log.proto](../internal/grpc_proto/remote_log.proto) | `LogTunnel(stream DeviceResponse) returns (stream ServerRequest)`                 | 服务端下发文件名，客户端回传日志尾部 |

`OpenService` 在 proto 中带有 grpc-gateway 的 HTTP 注解，并生成了 `open.pb.gw.go`，但当前 backend **没有挂载 gateway**（`grpc_gateway_client` 包无调用方）。节点实际走原生 gRPC；README 中列出的 `/setting`、`/device/ip-report` 等 gateway 路径目前不是 Gin 路由。

### 4.4 鉴权模型

- HTTP/MCP/Swagger：全局 BasicAuth。用户名或密码任一为空 → 直接放行（即关闭认证）。
- gRPC 节点侧：客户端通过 `PerRPCCredentials` 注入 `authorization: Basic base64(user:pass)`；服务端一元/流式拦截器解析 metadata 校验。
- 客户端只对一元调用配置了重试，流式调用没有重试。
- 所有通道共用同一组 `BASIC_AUTH_USERNAME` / `BASIC_AUTH_PASSWORD`。

### 4.5 接口入口

| 入口        | 路径                                                                                 | 实现                                                                         |
|-------------|--------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| 控制台页面  | `/dashboard` `/cdn` `/dns` `/ssl` `/tun` `/device` `/port-forward` `/task` `/log` 等 | [gin_server/router.go](../internal/gin_server/router.go) `setupWebRouter`    |
| REST API    | `/api/...`                                                                           | `setupApiRouter`，按资源分组                                                 |
| 健康检查    | `GET /health`                                                                        | [gin_server/server.go](../internal/gin_server/server.go)                     |
| Swagger     | `/swagger/docs_2/doc.json`、`/swagger/docs_3/doc.json`、`/swagger/:any`              | [internal/docs_2](../internal/docs_2)、[internal/docs_3](../internal/docs_3) |
| qingfeng UI | `/qingfeng/*`                                                                        | [gin_server/router.go](../internal/gin_server/router.go)                     |
| MCP         | `/mcp`（GET/POST/任意子路径）                                                        | [mcp_server/server.go](../internal/mcp_server/server.go)                     |
| gRPC        | 同端口，`application/grpc`                                                           | [grpc_server/server.go](../internal/grpc_server/server.go)                   |

REST 资源组包括：settings、schemas、platform-configs、cdn-records、dns-records、ssl-records、ssl-imports、tun-records、domain-monitors、devices、device-groups、ip-groups、port-forwards、webhooks、webhook-subs、tasks、gocrons、cloudflares、cloudflare-tunnels、cloudflare-zones、edgeone-zones、esa-sites。

### 4.6 MCP

- 采用 Streamable HTTP，无 stdio 模式；挂载在 Gin 的 `/mcp` 组，并叠加 `AuditContext(SourceMcp)`。
- 按业务域成组注册 prompt / resource / tool：time、platform-config、cdn-record、dns-record、device、ip-group、port-forward、domain-monitor、webhook、webhook-sub。
- 工具名采用“动词 + 实体驼峰”（如 `getDeviceList`、`updatePortForwardActive`）；参数 schema 由 `internal/utils.AutoSchema` 基于请求结构体反射生成。
- MCP handler 直接构造 backend service + 全局 DB 执行操作，与 REST 共用同一套业务逻辑。

---

## 5. 数据层

### 5.1 三个 SQLite 库

| 库     | 文件               | 用途                                                                                                   | 连接配置                                          |
|--------|--------------------|--------------------------------------------------------------------------------------------------------|---------------------------------------------------|
| 业务库 | `data/ent.db`      | 设备、平台配置、DNS/CDN/SSL/TUN 记录、Webhook、任务定义                                                | WAL、外键、`busy_timeout=5000`、`MaxOpenConns(1)` |
| 日志库 | `data/ent_log.db`  | 操作审计、IP 上报历史、任务执行日志、HTTP/Resty 请求日志、Webhook 发送历史、端口转发统计、域名监控历史 | 同上                                              |
| 统计库 | `data/ent_stat.db` | Cloudflare / EdgeOne / ESA 的站点日聚合指标                                                            | 同上                                              |

初始化见 [initialize/ent.go](../internal/initialize/ent.go)：每个库启动时执行 `Schema.Create` 自动建表/补列，没有版本化 migration；单连接串行写用于规避 SQLite 写锁。

### 5.2 业务库实体（30 张表）

| 分组         | 表                                                                                                                      | 说明                                           |
|--------------|-------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|
| 设备拓扑     | `device`、`device_group`、`device_group_device`、`ip_group`、`ip_group_device`、`device_port_forward`                   | 设备、设备组/IP 组及关联表、端口转发规则       |
| 平台与记录   | `platform_config`、`dns_record`、`cdn_record`、`ssl_record`、`ssl_import`、`tun_record`                                 | 平台凭据、各类同步记录、证书与手工导入         |
| Cloudflare   | `cloudflare_zone`、`cloudflare_zone_record`、`cloudflare_tunnel`、`cloudflare_tunnel_ingress`                           | 站点、DNS 记录、隧道与入站规则                 |
| 腾讯 EdgeOne | `edgeone_zone`、`edgeone_zone_acceleration_domain`、`edgeone_zone_dns_record`、`edgeone_zone_origin_group`              | 站点、加速域名、DNS 记录、源站组               |
| 阿里 ESA     | `esa_site`、`esa_site_record`、`esa_site_edge_container_record`、`esa_site_edge_routine_record`、`esa_site_origin_pool` | 站点、记录、边缘容器/边缘函数、源站池          |
| 运维         | `cron_job`、`cache`、`domain_monitor`、`webhook`、`webhook_subscription`                                                | 任务定义、运行时 KV 设置、域名监控、通知与订阅 |

### 5.3 日志库与统计库

日志库（10 张表，无 Ent 关系边，只有逻辑外键）：

`cron_job_log`、`device_ip_report`、`domain_monitor_history`、`http_log`、`req_log`、`resty_log`、`operation_log`、`port_forward_stats`、`ssl_certificate_history`、`webhook_history`。

统计库（3 张表）：

`cloudflare_zone_analysis_data`、`edge_one_zone_analysis_data`、`esa_site_analysis_data`，均按 `zone/site_id + date` 建索引。

### 5.4 关键关系与约束

- `platform_config` 1→N 关联 DNS/CDN/SSL/TUN 记录、Cloudflare Zone/Tunnel、EdgeOne Zone、ESA Site，删除平台配置级联删除子记录（[platformconfig.go](../internal/ent/schema/platformconfig.go) `Edges`）。
- `device` ↔ `device_group`、`device` ↔ `ip_group` 均为多对多，经关联表连接；删除设备级联删除端口转发规则。
- `dns_record` 的源站三选一：关联 IP 组、设备组或 CDN 记录；`cdn_record_id` 唯一。
- `cdn_record` 也可关联 IP 组或设备组，但当前同步任务只读取 IP 组（见[第 11 节](#11-已知约束与风险)）。
- `webhook_subscription` 用 `obj_type + obj_id` 多态指向任意业务对象，`(webhook_id, obj_type, obj_id)` 唯一，确保同一通知不会重复订阅同一对象。
- Cloudflare Zone→Record、Tunnel→Ingress，EdgeOne/ESA 父表→子表均为级联子表。

### 5.5 Schema 约定

- **时间字段**：`AtMixin` 提供 `created_at` / `updated_at`；日志/统计表用 `AtLogMixin`（仅 `created_at`）。同步类实体额外有 `remote_created_at` / `sync_at`。
- **枚举**：统一放在 [schematype/enum.go](../internal/ent/schema/schematype/enum.go)，Schema 中用 `field.Enum(...).Values(...)` 落库为字符串。
- **JSON**：配置、参数、快照统一用 `field.JSON`；凭据、私钥、平台配置标 `Sensitive()`。
- **主键/编码**：业务表用自增 `int64 id`；`cron_job.entry_id` 用 UUID v7，作为 gocron 的稳定标识。
- **审计 Hook**：除两张多对多关联表外，每个业务实体都挂审计 Hook，异步写 `operation_log`（记录来源、操作类型、对象 ID、变更前后值）。`cron_job` 与 `platform_config` 另有调度同步 Hook。

### 5.6 迁移、代码生成与数据保留

- 运行时迁移：启动时 `Schema.Create` 自动建表/补列；没有 down migration。
- 代码生成：`cmd/ent` 调用 `entc.Generate` 生成三个库的代码，并渲染 `audit.tmpl` 生成审计 Hook（`go run ./cmd/ent` 或 Makefile 的 `generate_ent`）。
- 静态 DDL：`sql/data_ddl.sql`、`sql/log_ddl.sql` 是可读快照，不参与运行时迁移。
- 数据保留：`clear_data` 任务按表删除 N 天前数据，覆盖 8 张日志表（HTTP、Resty、Req、域名监控历史、操作审计、Webhook 历史、IP 上报、端口转发统计）；**统计库目前没有清理任务**。
- 设置项存在 `cache` 表：离线判定 12 分钟、IP 上报 `8-12` 分钟、转发规则刷新 30 分钟、统计上报 `8-12` 分钟（[db_cache/db.go](../internal/helper/db_cache/db.go)）。

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

- 设备在线判定读取 `device.last_heartbeat_at` 与 `device_offline_invalid_time`（默认 12 分钟）；该字段为空显示“未知”。
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
   - A/AAAA：遍历 IP 组内“最后心跳时间在离线窗口内”的设备，按 ID 升序；**DNS 只保留最后一个匹配设备的 IP**。
4. 无可用源站 → 跳过。
5. 按平台类型构造 `DNSProvider`，先 `ParseCustomConfig(platform.config_json)` 解析账号凭据，再解析记录级 `config_json`。
6. `Check` 判断是否需要更新；需要则 `Update`。
7. 回写平台侧记录 ID、刷新时间；发生变化时创建 `webhook_send` 任务。

CDN 同步（[cdn-sync_task.go](../internal/provider/task/cdn-sync_task.go)）流程相同，但会**收集 IP 组内所有匹配设备的 IP** 作为多源站，并可能返回 `cdn_domain` 供 CNAME 使用。

调度节奏：两者均为 10–30 分钟随机；记录创建/更新/启停时由 service 调用 `LoadTask` 维护任务行。

> 注意：记录表的 `device_group_id` 字段与关联边存在，但当前 DNS/CDN 同步任务只读取 IP 组，未实现设备组取源站。

### 6.3 TUN 同步（当前未接线）

设计意图与 DNS/CDN 类似：`tun_sync` 任务读取 `tun_record`，调用 `TUNProvider.Check/Update` 同步 Cloudflare Tunnel 的 ingress，变更时发通知；任务定义为 10–30 分钟随机。

**现状**：`TUNSyncTaskProvider` 已实现，但：

- `cron_manager.InitHandlers` 没有注册 `tun_sync`；
- TUN 记录的 service 没有创建/维护 `tun_sync` 任务（`AddTask` 为空实现）。

因此该任务目前不会被执行。新增 TUN 同步能力时需要同时补注册与任务生命周期。

### 6.4 SSL 证书

- 申请/续期走 [internal/helper/ssl_helper](../internal/helper/ssl_helper)，基于 lego 的 DNS-01 校验；按平台配置选择对应 DNS/ESA/EO 的 challenge 实现。
- `ssl_apply` 任务（8–10 小时随机）：
  - 记录未启用则跳过；
  - 证书未过期且距过期天数大于 `renewal_days`（默认 30）则跳过；
  - 关闭自动续期则跳过；
  - 否则申请/续期，写入证书链、私钥、有效期等字段。
- `ssl_refresh` 任务每天 00:00 执行，刷新 `ssl_record` 与 `ssl_import` 两张表的证书元数据。
- 每次签发/刷新写 `ssl_certificate_history`；支持手工导入证书（`ssl_import`）。
- `provider.SSLProvider` 接口仍保留，但当前没有实现，实际逻辑在 `ssl_helper`。

### 6.5 域名监控

`domain_monitor` 任务按记录自身的 `check_interval`（秒）执行：

1. 按 `check_protocol`（HTTP/HTTPS）发起探测，可带自定义请求头。
2. 并发做 DNS 解析、HTTP/HTTPS 状态与 TCP/UDP 连通性检测。
3. 写 `domain_monitor_history`；HTTPS 额外记录证书生效/过期时间、CN、签发者、SAN、证书链、序列号、TLS 版本等。
4. 更新 `domain_monitor` 的当前状态与最近成功/失败时间。

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

- `tcpproxy` 仅支持 TCP，且不累计流量统计；UDP 回退到标准实现。
- `auto_forward`（防火墙自动放行）在 proto 中存在，但 forwarder 当前未实现。
- 转发实例重启会更换 `start_time`，统计查询会把新周期整段计入，不做跨重启差分。

### 6.7 Webhook 通知

- 一个 `webhook` 可被多个 `webhook_subscription` 订阅；订阅用 `obj_type + obj_id` 指向 CDN/DNS/设备/设备组/IP 组/域名监控/SSL/TUN 等对象。
- 事件发生时（IP 变化、DNS/CDN/TUN 更新、SSL 变化、域名监控状态变化、手工测试等），业务代码调用 `WebhookSendTaskProvider.AddTask` 创建一次性 `webhook_send` 任务（1 秒后立即执行）。
- 执行时：
  - 按 `business_type` 映射到 `obj_type`，查询该对象下所有启用订阅及其启用的 webhook；
  - 用 Go `text/template` 渲染标题/内容，模板分隔符为 `[[` `]]`；
  - 最多 10 个并发发送，支持企业微信、钉钉、飞书、Bark、Server 酱（ft07）、PushDeer、自定义 HTTP；
  - 每次发送写 `webhook_history`（URL、标题、内容、响应、成功/失败）。
- 手工测试直接指定 webhook ID，走单条发送分支。

### 6.8 云平台全量同步（Cloudflare / EdgeOne / ESA）

实现集中在 [internal/backend/syncs](../internal/backend/syncs)，由平台配置 Hook 创建 1–2 小时随机周期任务，也可通过 REST 手工触发一次性同步。

| 平台              | 同步内容                                           | 落库                                                                         |
|-------------------|----------------------------------------------------|------------------------------------------------------------------------------|
| Cloudflare Zone   | 站点列表、DNS 记录、按日统计                       | `cloudflare_zone`、`cloudflare_zone_record`、`cloudflare_zone_analysis_data` |
| Cloudflare Tunnel | 隧道列表、ingress 规则（先删后批量建）             | `cloudflare_tunnel`、`cloudflare_tunnel_ingress`                             |
| 腾讯 EdgeOne      | 站点、L7/L4 时序日聚合、源站组、DNS 记录、加速域名 | `edgeone_zone*`、`edge_one_zone_analysis_data`                               |
| 阿里 ESA          | 站点、时序指标、源站池、记录、边缘函数、边缘容器   | `esa_site*`、`esa_site_analysis_data`                                        |

同步按平台配置并发执行，单个账号内并发上限 5；数据库客户端通过 `syncs.InjectDb/InjectLogDb/InjectStatDb` 注入。

---

## 7. 调度与任务模型

### 7.1 `cron_job`：任务事实源

核心字段（[cronjob.go](../internal/ent/schema/cronjob.go)）：

| 字段                                                                               | 说明                                          |
|------------------------------------------------------------------------------------|-----------------------------------------------|
| `entry_id`                                                                         | UUID v7，gocron 任务标识，唯一且不可变        |
| `handler_name`                                                                     | 任务处理器名，对应 `TaskTask` 注册表的 key    |
| `handler_args`                                                                     | JSON 参数，通常为 `{id}` 或平台/站点 ID       |
| `task_type`                                                                        | `one_off` 一次性 / `recurring` 周期           |
| `job_type`                                                                         | `cron` / `duration` / `random`                |
| `job_category`                                                                     | `system` 系统内置 / `custom` 用户自定义       |
| `business_type` + `business_id` / `business_key`                                   | 任务归属的业务对象                            |
| `cron_expr` / `interval_value` / `min_interval` / `max_interval` / `interval_unit` | 调度参数                                      |
| `status`                                                                           | `running` / `stopped` / `finished` / `failed` |
| `is_singleton`                                                                     | 是否单例执行                                  |
| `run_on_start`                                                                     | 是否启动后立即执行一次                        |

### 7.2 执行链

1. `InitHandlers` 把 11 个 `TaskProvider` 注册到 `provider.TaskTask`（[cron_service.go](../internal/manager/cron_manager/cron_service.go)）。
2. `LoadAllJobs` 读取 `status=running` 的任务，按 `job_type` 创建 gocron Job；配置指纹变化时重建，删除/停用时移除。
3. 触发时 `RunHandler` 写一条 `cron_job_log`（Running），从注册表取处理器执行 `Process(ctx, args, meta)`，回填耗时、输出、Success/Failed。
4. 一次性任务在 `AfterJobRuns` 中被置为 `finished` 并从调度器移除；panic 时置为 `failed`。
5. 任务表的新增/删除通过 Ent Hook 异步调用 `SyncJob`；平台配置变化通过 `SyncPlatformConfigJob` 维护平台同步任务。

配置指纹包含任务类型、调度参数、参数、是否单例等；只有指纹变化才会重启计时器，避免无意义重启。

### 7.3 内置任务清单

| `handler_name`           | 作用                 | 调度                      | 任务来源                                       |
|--------------------------|----------------------|---------------------------|------------------------------------------------|
| `dns_sync`               | 单条 DNS 记录同步    | 10–30 分钟随机，周期/系统 | DNS 记录创建/更新/启停                         |
| `cdn_sync`               | 单条 CDN 记录同步    | 10–30 分钟随机，周期/系统 | CDN 记录创建/更新/启停                         |
| `tun_sync`               | 单条 TUN 记录同步    | 10–30 分钟随机，周期/系统 | **当前未注册、未创建任务**                     |
| `domain_monitor`         | 域名可用性与证书探测 | 记录 `check_interval` 秒  | 域名监控创建/更新                              |
| `ssl_apply`              | 证书申请/续期        | 8–10 小时随机，周期/系统  | SSL 记录创建/更新                              |
| `ssl_refresh`            | 刷新证书元数据       | 每天 00:00                | 启动时自动为 `ssl_record`、`ssl_import` 建任务 |
| `cloudflare_tunnel_sync` | 同步 Cloudflare 隧道 | 1–2 小时随机              | 平台配置 Hook                                  |
| `cloudflare_zone_sync`   | 同步 Cloudflare 站点 | 1–2 小时随机              | 平台配置 Hook                                  |
| `edgeone_zone_sync`      | 同步 EdgeOne 站点    | 1–2 小时随机              | 平台配置 Hook                                  |
| `esa_site_sync`          | 同步 ESA 站点        | 1–2 小时随机              | 平台配置 Hook                                  |
| `clear_data`             | 按表清理过期日志     | 每天 00:05                | 启动时自动为 8 张日志表建任务                  |
| `webhook_send`           | 发送通知             | 一次性，1 秒后立即执行    | 事件触发                                       |

手工同步（Cloudflare/EdgeOne/ESA）会创建一次性任务，1 秒后立即执行，不改变周期任务的配置。

---

## 8. 厂商适配层

### 8.1 接口

| 接口           | 文件                                                      | 方法                                           | 使用方               |
|----------------|-----------------------------------------------------------|------------------------------------------------|----------------------|
| `DNSProvider`  | [dns_provider.go](../internal/provider/dns_provider.go)   | `Name`、`Check`、`Update`、`ParseCustomConfig` | `dns_sync`           |
| `CDNProvider`  | [cdn_provide.go](../internal/provider/cdn_provide.go)     | `Name`、`Check`、`Update`、`ParseCustomConfig` | `cdn_sync`           |
| `TUNProvider`  | [tun_provider.go](../internal/provider/tun_provider.go)   | `Name`、`Check`、`Update`、`ParseCustomConfig` | `tun_sync`（未接线） |
| `TaskProvider` | [task_provider.go](../internal/provider/task_provider.go) | `Process(ctx, args, meta)`                     | `cron_manager`       |
| `SSLProvider`  | [provider.go](../internal/provider/provider.go)           | `Renew`、`Present`、`CleanUp`、`Name`          | 保留接口，暂无实现   |

所有接口都有 `UnimplementedXxxProvider` 安全默认实现，具体厂商实现首行嵌入它，未实现的方法返回明确错误。

### 8.2 配置分两层

1. **平台配置** `platform_config.config_json`：账号级凭据，如 AccessKey、API Token、Account ID、Endpoint。
2. **记录配置** `dns_record.config_json` / `cdn_record.config_json` / `tun_record.config_json`：记录级参数，如线路、Zone、Tunnel ID。

任务执行时先解析平台配置构造 Provider，再在 `Check/Update` 前解析记录配置。凭据字段在 Ent Schema 中标为 `Sensitive()`。

### 8.3 注册与选择

- 任务内按 `platform_config.platform` 做 `switch`，构造对应 Provider；未知平台返回“暂不支持该平台”。
- 任务处理器通过 `TaskTask.RegisterTask(name, handler)` 注册，name 与 `cron_job.handler_name` 一一对应。
- HTTP 客户端统一接入 `http_log.LoggingRoundTripper`，厂商请求会写入 `resty_log` / `http_log`（取决于使用的客户端），并通过 `inject.go` 的 `sync.Once` 注入。

### 8.4 新增厂商的步骤

1. 在 `internal/provider/{dns,cdn,tun}` 下新增 `xxx_provider.go`，实现对应接口并嵌入 `Unimplemented*`。
2. 定义配置结构体与 `ParseCustomConfig`，字段与 `platform_config.config_json` 对齐。
3. 在 `internal/provider/{dns,cdn,tun}/inject.go` 确认 HTTP transport 注入可用。
4. 在目标任务（`dns-sync_task.go` / `cdn-sync_task.go` / `tun-sync_task.go`）的 `switch platform.Platform` 中增加分支。
5. 在 `schematype` 的平台枚举与前端 Schema（`internal/utils` 或 `handlers/api.schemas.go`）中补平台定义。

---

## 9. 模块职责速查

| 目录                                                | 职责                                                  | 关键文件                                         |
|-----------------------------------------------------|-------------------------------------------------------|--------------------------------------------------|
| `cmd/backend`                                       | 服务端入口，embed 模板/静态资源                       | `main.go`                                        |
| `cmd/reporter`                                      | IP 上报与设备长连接                                   | `main.go`                                        |
| `cmd/forwarder`                                     | 端口转发与统计上报                                    | `main.go`                                        |
| `cmd/frontend`                                      | 前端资产构建（复制 libs + Tailwind 编译），无应用源码 | `package.json`、`test/`                          |
| `cmd/ent`                                           | Ent 代码与审计 Hook 生成器                            | `main.go`                                        |
| `internal/ent`                                      | 业务库 Schema 与生成代码                              | `schema/`                                        |
| `internal/ent_log`                                  | 日志/审计库 Schema 与生成代码                         | `schema/`                                        |
| `internal/ent_stat`                                 | 统计库 Schema 与生成代码                              | `schema/`                                        |
| `internal/initialize`                               | 日志、三库、gocron、HTTP/Resty 日志初始化             | `ent.go`、`gocron.go`、`slog.go`                 |
| `internal/global`                                   | 全局单例（logger、三库、gocron、版本信息）            | `global.go`                                      |
| `internal/manager/mixed_manager`                    | Gin + gRPC h2c 混合服务                               | `mixed_server.go`                                |
| `internal/manager/cron_manager`                     | 任务注册、加载、同步、执行、日志                      | `cron_service.go`                                |
| `internal/gin_server`                               | Gin 初始化、模板/静态资源、路由                       | `server.go`、`router.go`、`resource.go`          |
| `internal/gin_middleware`                           | 认证、审计、WAF、CORS、Gzip、Recovery、RequestID      | 各中间件文件                                     |
| `internal/grpc_server`                              | gRPC 引擎、拦截器、服务注册                           | `server.go`                                      |
| `internal/grpc_service`                             | 三个 gRPC 服务实现                                    | `device_service.go` 等                           |
| `internal/grpc_proto`                               | proto 定义                                            | `device.proto`、`open.proto`、`remote_log.proto` |
| `internal/grpc_generated`                           | protobuf/gRPC/gateway 生成代码                        | `*.pb.go`                                        |
| `internal/grpc_client`                              | 节点侧 gRPC 客户端                                    | `client.go`                                      |
| `internal/grpc_gateway_client`                      | grpc-gateway 客户端（当前未接入 backend）             | `client.go`                                      |
| `internal/grpc_credential`                          | 节点侧 BasicAuth 凭据                                 | `basic_auth.go`                                  |
| `internal/grpc_interceptor`                         | gRPC 审计/日志/认证拦截器                             | 各拦截器文件                                     |
| `internal/handlers`                                 | 页面与 REST handler                                   | `api.*.go`、页面 handler                         |
| `internal/types`                                    | API 请求/响应与页面数据 DTO                           | `api.*.go`                                       |
| `internal/backend/service`                          | 业务编排、事务、CRUD                                  | `*_service.go`                                   |
| `internal/backend/syncs`                            | Cloudflare/EdgeOne/ESA 全量同步                       | `*_sync.go`                                      |
| `internal/provider`                                 | 厂商接口与任务注册表                                  | `*_provider.go`、`task_provider.go`              |
| `internal/provider/{dns,cdn,tun}`                   | 厂商实现                                              | `*_dns.go`、`*_cdn.go`、`cloudflare_tun.go`      |
| `internal/provider/task`                            | 定时任务实现                                          | `*_task.go`                                      |
| `internal/helper`                                   | IP/域名/SSL/Webhook/gocron/db_cache/日志辅助          | 各子包                                           |
| `internal/core`                                     | 无状态领域函数                                        | `device.go`、`ip.go`、`token.go`                 |
| `internal/mcp_server`                               | MCP 服务与 `/mcp` 挂载                                | `server.go`                                      |
| `internal/mcp_tool` / `mcp_resource` / `mcp_prompt` | MCP 工具/资源/提示词                                  | 按业务域分文件                                   |
| `internal/reporter`                                 | reporter 服务与任务                                   | `service/`、`tasks/`                             |
| `internal/forwarder`                                | forwarder 转发器、服务与任务                          | `forward/`、`service/`、`tasks/`                 |
| `internal/docs_2` / `docs_3`                        | Swagger OpenAPI 2 / 3.1 数据                          | `swagger.json`、`docs.go`                        |
| `internal/utils`                                    | DNS/域名/SSL/MCP Schema 辅助                          | 各工具文件                                       |
| `pkg/utils`                                         | 可复用文件工具                                        | `file.go`                                        |
| `tests`                                             | 需要真实环境的集成测试（不参与 CI 覆盖率）            | —                                                |

---

## 10. 扩展指南

### 10.1 新增业务实体

1. 在 `internal/ent/schema/` 新增 Schema，定义表名、字段、索引、边、Mixin、审计 Hook。
2. 运行 `go run ./cmd/ent`（或 `make generate_ent`）生成 Ent 代码与审计 Hook。
3. 在 `internal/types` 增加请求/响应 DTO。
4. 在 `internal/backend/service` 增加 service，承载事务与业务规则。
5. 在 `internal/handlers` 增加页面/API handler，并在 `gin_server/router.go` 注册路由。
6. 如需通知，在 `schematype` 的 Webhook 对象类型中增加枚举，并在 `WebhookSendTaskProvider` 中补映射。
7. 如需 MCP 能力，在 `mcp_tool` / `mcp_resource` / `mcp_prompt` 增加注册函数，并在 `mcp_server.Run` 中调用。

### 10.2 新增定时任务

1. 在 `internal/provider/task/` 新建 `xxx_task.go`，实现 `TaskProvider.Process`，首行嵌入 `UnimplementedTaskProvider`。
2. 定义 `XxxTaskName` 常量，作为 `cron_job.handler_name`。
3. 在 `cron_manager.InitHandlers` 中 `RegisterTask`。
4. 如需随业务记录自动建任务，实现 `LoadTask` / `AddTask`，并在对应 service 的创建/更新/启停路径调用。
5. 如需随平台配置自动建任务，在 `SyncPlatformConfigJob` 中增加分支。
6. 任务参数写入 `handler_args`，执行时通过 `meta` 拿到业务类型/ID/分类。

### 10.3 新增 REST API

1. 在 `internal/types` 定义请求/响应结构（含 swag 注解所需的字段说明）。
2. 在 `internal/backend/service` 实现业务方法。
3. 在 `internal/handlers/api.*.go` 绑定参数、调用 service、返回 `types.APIResponse`。
4. 在 `gin_server/router.go` 的 `/api` 分组中注册路由；必要时加缓存中间件。
5. 运行 Swagger 生成命令更新 `internal/docs_2`、`internal/docs_3`。

### 10.4 新增控制台页面

1. 在 `cmd/backend/templates/pages/` 新建页面模板，复用 `layout/base.html` 与 `components/`。
2. 在 `internal/handlers` 增加页面 handler，构造 `types.BasePageData[T]`。
3. 在 `gin_server/router.go` 的 `setupWebRouter` 注册路径。
4. 在 `cmd/backend/static/js/` 增加页面脚本（经典 `<script>` 加载，非 ESM）。

### 10.5 新增 MCP 工具

1. 在 `mcp_tool` 定义入参/出参结构体并注册工具；schema 可由 `utils.AutoSchema` 反射生成。
2. 如需要资源模板，在 `mcp_resource` 注册 URI 模板；如需要提示词，在 `mcp_prompt` 注册。
3. 在 `mcp_server.Run` 的对应业务域中调用注册函数。
4. 工具内部直接复用 backend service 与全局 DB，不复制 REST 逻辑。

---

## 11. 已知约束与风险

以下是当前代码的客观现状，部署或二次开发前需要评估：

1. **只能单实例部署**：设备连接表 `clients`、运行中任务 `runningJobs`、各 Manager 单例都在进程内存；多副本会重复调度、设备连接分散。
2. **SQLite 单写**：三个库都 `MaxOpenConns(1)`，高并发写入会排队；`data/*.db` 路径相对进程工作目录，启动目录变化会打开新库。
3. **启动自动迁移无版本管理**：`Schema.Create` 只做增量补列，复杂变更（改类型、删列、数据回填）需要手工处理；变更期间需避免多进程同时启动。
4. **明文传输**：混合服务未配置 TLS，gRPC 使用 `insecure` 凭据；BasicAuth 在网络上是可解码的，建议置于反向代理/TLS 之后。
5. **认证粒度粗**：HTTP、MCP、Swagger、REST 共用一组 BasicAuth；任一项为空即关闭认证。
6. **CORS 过宽**：允许任意 Origin 且 `AllowCredentials=true`，生产环境应收紧。
7. **WAF 实际未加载规则**：Coraza 只启用引擎与请求体读取，没有 `Include` OWASP CRS，拦截能力接近于零。
8. **优雅关闭存在窗口**：混合服务 `Stop` 在 goroutine 中执行，主流程未等待完成就关闭数据库，极端情况下正在处理的请求会遇到已关闭的 DB。
9. **MCP 路由注册时机**：`/mcp` 在 goroutine 中向 Gin 注册，与混合服务启动并发，理论上存在路由树并发读写风险；启动日志中若出现 MCP 端点缺失，需检查该竞态。
10. **gRPC-Gateway 未挂载**：proto 注解与 `open.pb.gw.go` 存在，但 backend 没有接入 gateway；README 中的 gateway 路径当前不可用，节点使用原生 gRPC。
11. **远程日志隧道未接线**：`RemoteLogService.LogTunnel` 服务端已实现，reporter/forwarder 也定义了 `StartTunnel`，但主程序没有调用，隧道当前不会建立。
12. **TUN 同步未接线**：`tun_sync` 处理器未注册、TUN service 未创建任务，详见 [6.3](#63-tun-同步当前未接线)。
13. **设备组取源站未实现**：DNS/CDN 记录可关联设备组，但同步任务只读取 IP 组。
14. **DNS 单值语义**：A/AAAA 记录只使用 IP 组内最后一个匹配设备的 IP；多设备需要多值解析时应改用 CDN 多源站或拆分记录。
15. **监听端口不一致**：默认 `SERVER_ADDR=0.0.0.0:8080`，本地开发 Makefile 使用 9999；排查连接问题时先确认实际端口。
16. **日志增长**：`ent_log` 与 `ent_stat` 会持续增长，`clear_data` 只覆盖 8 张日志表，统计库没有清理任务；长期运行需外部归档或补清理任务。
17. **审计来源不完整**：HTTP 请求统一记为 `user`，gRPC 记为 `grpc`，MCP 记为 `mcp`；`task`/`gin`/`api` 枚举已定义但当前没有写入点，定时任务执行会落到默认 `user`。
18. **tcpproxy 无流量统计**：`implementation_type=tcpproxy` 仅支持 TCP，且不累计字节数；需要流量统计时使用 `standard`。
19. **一次性任务状态**：一次性任务执行完由 gocron 回调置为 `finished`；失败状态依赖任务行自身状态，排查失败请以 `cron_job_log` 为准。

---

## 附录 A：环境变量与默认路径

| 变量                  | 默认值           | 说明                                            |
|-----------------------|------------------|-------------------------------------------------|
| `DEBUG`               | `false`          | 调试模式（Gin/Ent/gocron 日志级别、模板来源等） |
| `ENABLE_LOG`          | 未设置           | 是否启用 HTTP 请求日志                          |
| `SERVER_ADDR`         | `0.0.0.0:8080`   | backend HTTP/gRPC 混合监听地址                  |
| `BACKEND_ENDPOINT`    | `127.0.0.1:8080` | reporter/forwarder 连接的 backend 地址          |
| `BASIC_AUTH_USERNAME` | 空               | BasicAuth 用户名；为空关闭认证                  |
| `BASIC_AUTH_PASSWORD` | 空               | BasicAuth 密码；为空关闭认证                    |
| `DEVICE_CODE`         | 自动生成         | 设备编码；为空时用 `machineid.ID()`             |
| `DEVICE_NAME`         | 空               | 设备名称                                        |

| 路径                                          | 说明                            |
|-----------------------------------------------|---------------------------------|
| `data/ent.db`                                 | 业务库                          |
| `data/ent_log.db`                             | 日志/审计库                     |
| `data/ent_stat.db`                            | 统计库                          |
| `logs/`                                       | 分类日志文件（lumberjack 轮转） |
| `cmd/backend/templates`、`cmd/backend/static` | 开发模式下的模板与静态资源      |
| `internal/docs_2`、`internal/docs_3`          | Swagger JSON/YAML               |

---

## 附录 B：关键文件索引

| 关注点            | 文件                                                                                                                                                                                      |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| backend 启动/关闭 | [cmd/backend/main.go](../cmd/backend/main.go)                                                                                                                                             |
| Gin+ gRPC 分流    | [internal/manager/mixed_manager/mixed_server.go](../internal/manager/mixed_manager/mixed_server.go)                                                                                       |
| HTTP 中间件       | [internal/gin_server/server.go](../internal/gin_server/server.go)、[internal/gin_middleware](../internal/gin_middleware)                                                                  |
| 路由              | [internal/gin_server/router.go](../internal/gin_server/router.go)                                                                                                                         |
| gRPC 注册/拦截器  | [internal/grpc_server/server.go](../internal/grpc_server/server.go)                                                                                                                       |
| 三库初始化/迁移   | [internal/initialize/ent.go](../internal/initialize/ent.go)                                                                                                                               |
| 业务 Schema       | [internal/ent/schema](../internal/ent/schema)                                                                                                                                             |
| 日志 Schema       | [internal/ent_log/schema](../internal/ent_log/schema)                                                                                                                                     |
| 统计 Schema       | [internal/ent_stat/schema](../internal/ent_stat/schema)                                                                                                                                   |
| 任务注册/调度     | [internal/manager/cron_manager/cron_service.go](../internal/manager/cron_manager/cron_service.go)                                                                                         |
| 任务实现          | [internal/provider/task](../internal/provider/task)                                                                                                                                       |
| 厂商接口          | [internal/provider/dns_provider.go](../internal/provider/dns_provider.go)、[cdn_provide.go](../internal/provider/cdn_provide.go)、[tun_provider.go](../internal/provider/tun_provider.go) |
| 设备 gRPC 服务    | [internal/grpc_service/device_service.go](../internal/grpc_service/device_service.go)                                                                                                     |
| 开放接口实现      | [internal/backend/service/open_service.go](../internal/backend/service/open_service.go)                                                                                                   |
| 平台全量同步      | [internal/backend/syncs](../internal/backend/syncs)                                                                                                                                       |
| 端口转发数据面    | [internal/forwarder/forward/forward.go](../internal/forwarder/forward/forward.go)                                                                                                         |
| MCP 注册          | [internal/mcp_server/server.go](../internal/mcp_server/server.go)                                                                                                                         |
| 运行时设置        | [internal/helper/db_cache/db.go](../internal/helper/db_cache/db.go)                                                                                                                       |
