# IPVibe

IPVibe 是一套面向多机房 / 多设备的 **IP 动态管理与应用交付** 平台：把「设备 IP 上报 → DNS / CDN / TUN 记录同步 → SSL 证书申请续期 → 域名监控 → 端口转发 → 通知告警」串成一条自动化流水线，并提供 Web 控制台、REST API、gRPC、MCP 四种接入方式。

<div align="center">

[![CNB 版本](https://cnb.cool/dtapp/ip_vibe/-/badge/release)](https://cnb.cool/dtapp/ip_vibe/-/badge/release.link) [![CNB Star 数](https://cnb.cool/dtapp/ip_vibe/-/badge/star)](https://cnb.cool/dtapp/ip_vibe) [![CNB Fork 数](https://cnb.cool/dtapp/ip_vibe/-/badge/fork)](https://cnb.cool/dtapp/ip_vibe) [![CNB 构建](https://cnb.cool/dtapp/ip_vibe/-/badge/git/latest/ci/status/push)](https://cnb.cool/dtapp/ip_vibe) [![CNB 覆盖率](https://cnb.cool/dtapp/ip_vibe/-/badge/git/latest/testing/unit/coverage)](https://cnb.cool/dtapp/ip_vibe/-/insights/unit-test) [![CNB PR 覆盖率](https://cnb.cool/dtapp/ip_vibe/-/badge/git/latest/testing/unit/coverage-pr)](https://cnb.cool/dtapp/ip_vibe/-/insights/unit-test) ![CNB 下载请求数](https://img.shields.io/endpoint?label=CNB%20下载请求数&url=https%3A%2F%2Fdl-stats.dtapp.net%2Fshields.json%3Fplatform%3Dcnb.cool%26repository%3D%2Fdtapp%2Fip_vibe)

[![CNB 漏洞](https://cnb.cool/dtapp/ip_vibe/-/badge/security/vulnerability/full/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/vulnerability/full/latest.link) [![CNB 漏洞差异](https://cnb.cool/dtapp/ip_vibe/-/badge/security/vulnerability/diff/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/vulnerability/diff/latest.link) [![CNB 许可证](https://cnb.cool/dtapp/ip_vibe/-/badge/security/license/full/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/license/full/latest.link) [![CNB 许可证差异](https://cnb.cool/dtapp/ip_vibe/-/badge/security/license/diff/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/license/diff/latest.link) [![CNB SBOM](https://cnb.cool/dtapp/ip_vibe/-/badge/security/sbom/full/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/sbom/full/latest.link) [![CNB SBOM 差异](https://cnb.cool/dtapp/ip_vibe/-/badge/security/sbom/diff/latest)](https://cnb.cool/dtapp/ip_vibe/-/badge/security/sbom/diff/latest.link)

[![CNB 访问](https://commit.cool/badge/visit/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 发布频率](https://commit.cool/badge/release/frequency/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 距离上次发布](https://commit.cool/badge/release/last/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 发布总数](https://commit.cool/badge/releases/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe/-/releases) [![CNB 仓库年龄](https://commit.cool/badge/age/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 主要语言](https://commit.cool/badge/language/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 仓库许可证](https://commit.cool/badge/license/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB 存储用量](https://commit.cool/badge/usage/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB Git 存储](https://commit.cool/badge/usage/git/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe) [![CNB LFS 存储](https://commit.cool/badge/usage/lfs/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe)[![CNB Docker 拉取次数](https://commit.cool/badge/package/docker/pulls/dtapp/ip_vibe)](https://docker.cnb.cool/dtapp/ip_vibe)

[![GitHub 版本](https://img.shields.io/github/v/release/dtapps/ip_vibe?style=flat-square)](https://github.com/dtapps/ip_vibe/releases) [![GitHub 下载量](https://img.shields.io/github/downloads/dtapps/ip_vibe/total?style=flat-square)](https://github.com/dtapps/ip_vibe/releases) [![GitHub Star 数](https://img.shields.io/github/stars/dtapps/ip_vibe?style=flat-square)](https://github.com/dtapps/ip_vibe/stargazers) [![GitHub Fork 数](https://img.shields.io/github/forks/dtapps/ip_vibe?style=flat-square)](https://github.com/dtapps/ip_vibe/network/members) [![GitHub 许可证](https://img.shields.io/github/license/dtapps/ip_vibe?style=flat-square)](LICENSE) [![GitHub 支持平台](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-007EC6?style=flat-square)](https://github.com/dtapps/ip_vibe) [![GitHub 构建](https://img.shields.io/github/actions/workflow/status/dtapps/ip_vibe/push.yml?style=flat-square)](https://github.com/dtapps/ip_vibe/actions/workflows/push.yml) [![Codecov](https://img.shields.io/codecov/c/github/dtapps/ip_vibe?style=flat-square&label=Codecov&logo=codecov)](https://codecov.io/gh/dtapps/ip_vibe) [![Coveralls](https://img.shields.io/coveralls/github/dtapps/ip_vibe?style=flat-square&label=Coveralls&logo=coveralls)](https://coveralls.io/github/dtapps/ip_vibe) ![GitHub 下载请求数](https://img.shields.io/endpoint?label=GitHub%20下载请求数&url=https%3A%2F%2Fdl-stats.dtapp.top%2Fshields.json%3Fplatform%3Dgithub.com%26repository%3D%2Fdtapps%2Fip_vibe)

[![GitHub 访问](https://visitor-badge.laobi.icu/badge?page_id=github/dtapps/ip_vibe)](https://github.com/dtapps/ip_vibe) [![GitHub 距离上次发布](https://img.shields.io/github/release-date/dtapps/ip_vibe)](https://github.com/dtapps/ip_vibe) [![GitHub 仓库年龄](https://img.shields.io/github/created-at/dtapps/ip_vibe)](https://github.com/dtapps/ip_vibe) [![GitHub 主要语言](https://img.shields.io/github/languages/top/dtapps/ip_vibe)](https://github.com/dtapps/ip_vibe) [![GitHub 存储用量](https://img.shields.io/github/repo-size/dtapps/ip_vibe)](https://github.com/dtapps/ip_vibe)[![GHCR backend 拉取次数](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fghcr-badge.elias.eu.org%2Fapi%2Fdtapps%2Fip_vibe%252Fbackend&query=downloadCount&label=ghcr%20backend%20pulls&logo=docker&logoColor=white&color=2496ed)](https://github.com/dtapps/ip_vibe/pkgs/container/ip_vibe%2Fbackend) [![GHCR reporter 拉取次数](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fghcr-badge.elias.eu.org%2Fapi%2Fdtapps%2Fip_vibe%252Freporter&query=downloadCount&label=ghcr%20reporter%20pulls&logo=docker&logoColor=white&color=2496ed)](https://github.com/dtapps/ip_vibe/pkgs/container/ip_vibe%2Freporter) [![GHCR forwarder 拉取次数](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fghcr-badge.elias.eu.org%2Fapi%2Fdtapps%2Fip_vibe%252Fforwarder&query=downloadCount&label=ghcr%20forwarder%20pulls&logo=docker&logoColor=white&color=2496ed)](https://github.com/dtapps/ip_vibe/pkgs/container/ip_vibe%2Fforwarder)

</div>

---

## 目录

- [功能特性](#功能特性)
- [架构与组件](#架构与组件)
- [支持的平台厂商](#支持的平台厂商)
- [支持的通知渠道](#支持的通知渠道)
- [快速开始](#快速开始)
- [环境变量](#环境变量)
- [接口入口](#接口入口)
- [目录结构](#目录结构)
- [常用命令](#常用命令)
- [文档](#文档)
- [许可证](#许可证)

---

## 功能特性

**IP 与设备管理**

- 设备自动注册与心跳上报，设备编号（`DEVICE_CODE`）留空时依据 `/etc/machine-id` 自动生成
- 设备分组、IP 分组，按组批量下发与统计
- 设备 IP 变更历史、在线状态与健康检查
- 远程日志隧道：服务端下发文件名，客户端回传该日志文件最后 200 行

**DNS / CDN / TUN 同步**

- DNS 记录同步：A / AAAA / CNAME，IP 变化时自动更新解析
- CDN 回源同步：支持多源站（`origin_values`）批量刷新
- Cloudflare Tunnel（TUN）同步与 Ingress 入站规则管理，服务类型覆盖 http / https / unix / tcp / ssh / rdp / unix_tls / smb / http_status / bastion
- 记录级健康检查、启用开关、批量启停与批量删除

**SSL 证书**

- 基于 Let's Encrypt（lego）自动申请证书，DNS-01 校验
- 自动续期、证书历史记录，支持手动导入既有证书

**域名监控**

- 周期性探测域名可用性与解析结果，留存监控历史

**端口转发**

- TCP / UDP（`tcp4` `tcp6` `udp4` `udp6`）端口转发规则下发到 forwarder
- 转发流量统计上报与图表展示

**定时任务**

- 内置任务：DNS 同步、CDN 同步、TUN 同步、SSL 申请、SSL 刷新、域名监控、数据清理、Webhook 发送、Cloudflare 隧道 / 站点同步、EdgeOne 站点同步、ESA 站点同步
- 支持 cron / duration / random 三种调度，单次（one_off）与周期（recurring）任务，系统内置与用户自定义分类

**可观测与治理**

- 操作审计日志，来源标记 system / task / user / gin / api / mcp / grpc 等
- HTTP / gRPC / Resty / Ent / gocron 分类日志落盘
- Basic Auth（HTTP 中间件 + gRPC 拦截器）、Coraza WAF + OWASP CRS、CORS、Gzip、RequestID、Recovery
- MCP Server 暴露 tool / resource / prompt，可被 AI 客户端直接调用

---

## 架构与组件

| 组件          | 入口            | 职责                                                                               |
|---------------|-----------------|------------------------------------------------------------------------------------|
| **backend**   | `cmd/backend`   | 控制台与 API 服务端。Gin + gRPC 同端口混合服务，三库持久化，定时任务调度，MCP 服务 |
| **reporter**  | `cmd/reporter`  | 部署在被管设备上，周期性上报本机 IP，并保持设备长连接                              |
| **forwarder** | `cmd/forwarder` | 部署在被管设备上，执行端口转发并上报流量统计                                       |

**技术栈**

- 语言：Go 1.27.3（模块名 `vibe_ip`）
- ORM / 存储：Ent + SQLite（`data/ent.db` 业务库、`data/ent_log.db` 日志库、`data/ent_stat.db` 统计库，WAL 模式）
- Web：Gin，HTML 模板 + 静态资源通过 `go:embed` 内置
- RPC：gRPC + protobuf + grpc-gateway（REST 转接）+ protoc-gen-validate
- 调度：gocron v2
- MCP：modelcontextprotocol/go-sdk
- 文档：Swagger（OpenAPI 2 与 OpenAPI 3.1 双版本）
- 前端：TailwindCSS v4 + DaisyUI v5 + jQuery + ECharts + dayjs + Select2，构建产物输出到 `cmd/backend/static`

---

## 支持的平台厂商

**DNS**

阿里云 DNS、腾讯云 DNS、华为云 DNS、百度云 DNS、京东云 DNS、火山引擎 DNS、Cloudflare DNS

**CDN**

阿里云 CDN、腾讯云 CDN、华为云 CDN、百度云 CDN、京东云 CDN、金山云 CDN、七牛云 CDN、阿里云 ESA、腾讯云 EO

**TUN**

Cloudflare Tunnel

**SSL**

Let's Encrypt

---

## 支持的通知渠道

企业微信、钉钉、飞书、Bark、Server 酱（ft07）、PushDeer、自定义 HTTP

通知以 Webhook + 订阅的方式组织：一个 Webhook 可被多个对象订阅（CDN 记录、DNS 记录、设备、设备组、IP 组、域名监控、SSL 记录、TUN 记录），事件触发时统一推送并留存发送历史。

---

## 快速开始

### Docker Compose（推荐）

```bash
cd cmd
docker compose up -d
```

启动 backend / reporter / forwarder 三个服务，数据与日志分别挂载到 `./data`、`./logs`。 控制台默认访问 `http://127.0.0.1:8080`，健康检查 `GET /health`。

> 镜像：`docker.cnb.cool/dtapp/ip_vibe/{backend,reporter,forwarder}:latest`，均使用 `network_mode: host`。 `reporter` / `forwarder` 通过 `BACKEND_ENDPOINT` 指向 backend 地址。
>
> 版本号规则：`v1.2.3` 为正式版（推送 `:v1.2.3` 与 `:latest`，Release 不勾选 prerelease）；`v1.2.3-rc.1`、`v1.2.3-beta.1`、`v1.2.3-alpha.1`、`v1.2.3-nightly.20260101` 为预发布（只推送对应版本标签，不覆盖 `:latest`，Release 勾选 prerelease）；填 `nightly` 会自动推导成 `<最近正式版 +1 补丁>-nightly.<UTC 日期>`（定时任务每天 UTC 00:00 用它，各平台只保留最近 3 个 nightly Release）。
>
> GitHub Actions 侧同一套规则：`Release` 工作流用 GoReleaser 产出全平台压缩包并发布 Release，`Build Image` 工作流把镜像推到 `ghcr.io/dtapps/ip_vibe/{backend,reporter,forwarder}`，再同步一份到 CNB 制品库。两者都支持手动触发与每天 UTC 00:00（北京时间 08:00）的定时触发（定时等价于填 `nightly`），完成后会通过 ServerChan3（方糖）推一条通知。

### 本地开发

```bash
# 一次性初始化：安装代码生成工具 + 拉取 Go 与前端依赖
make setup

# 终端 1：前端样式热更新
make dev_frontend

# 终端 2：后端服务（0.0.0.0:9999）
make dev_backend

# 终端 3：IP 上报服务
make dev_reporter

# 终端 4：端口转发服务
make dev_forwarder
```

需要 Basic Auth 时使用 `make dev_backend_basic_auth` / `dev_reporter_basic_auth` / `dev_forwarder_basic_auth`（默认 `admin` / `123456`，见 `Makefile` 顶部变量）。

---

## 环境变量

| 变量                  | 默认值           | 说明                                        |
|-----------------------|------------------|---------------------------------------------|
| `DEBUG`               | `false`          | 调试模式                                    |
| `ENABLE_LOG`          | —                | 是否启用日志                                |
| `SERVER_ADDR`         | `0.0.0.0:8080`   | backend 的 HTTP / gRPC 混合监听地址         |
| `BACKEND_ENDPOINT`    | `127.0.0.1:8080` | reporter / forwarder 连接的 backend 地址    |
| `BASIC_AUTH_USERNAME` | 空               | Basic Auth 用户名，留空则关闭认证           |
| `BASIC_AUTH_PASSWORD` | 空               | Basic Auth 密码                             |
| `DEVICE_CODE`         | 自动生成         | 设备编号，留空则依据 `/etc/machine-id` 生成 |
| `DEVICE_NAME`         | 空               | 设备名称                                    |

---

## 接口入口

| 入口         | 路径                                                                                                            |
|--------------|-----------------------------------------------------------------------------------------------------------------|
| Web 控制台   | `/dashboard` `/cdn` `/dns` `/ssl` `/tun` `/device` `/port-forward` `/task` `/log` 等                            |
| REST API     | `/api/...`（平台配置、CDN/DNS/TUN 记录、SSL、域名监控、设备、设备组、IP 组、端口转发、Webhook、任务、调度器等） |
| Swagger UI   | `/swagger/:any`                                                                                                 |
| Swagger JSON | `/swagger/docs_2/doc.json`（OpenAPI 2）、`/swagger/docs_3/doc.json`（OpenAPI 3.1）                              |
| qingfeng UI  | `/qingfeng`                                                                                                     |
| MCP          | `/mcp`                                                                                                          |
| gRPC-Gateway | `GET /setting`、`POST /device/ip-report`、`GET /device/{device_code}/port-forwards`                             |

---

## 目录结构

```
.
├── cmd/                    # 可执行程序入口
│   ├── backend/            # 服务端（templates/ static/ 由 go:embed 内置）
│   ├── reporter/           # IP 上报端
│   ├── forwarder/          # 端口转发端
│   ├── frontend/           # 前端源码与构建脚本
│   └── docker-compose.yml  # 一键编排
├── internal/
│   ├── backend/            # 业务 service 与平台数据同步
│   ├── ent/                # 业务库 Ent schema 与生成代码
│   ├── ent_log/            # 日志库 Ent schema 与生成代码
│   ├── ent_stat/           # 统计库 Ent schema 与生成代码
│   ├── gin_server/         # Gin 服务与路由注册
│   ├── gin_middleware/     # 认证、审计、WAF、CORS 等中间件
│   ├── grpc_proto/         # device / open / remote_log 三个 proto
│   ├── grpc_generated/     # protobuf 生成代码
│   ├── grpc_service/       # gRPC 服务实现
│   ├── mcp_server/         # MCP 服务
│   ├── mcp_tool|resource|prompt/
│   ├── provider/           # 厂商适配（dns/ cdn/ tun/ ssl/ task）
│   ├── manager/            # 定时任务与混合服务管理
│   ├── handlers/           # 页面与 API handler
│   ├── initialize/         # 数据库、日志、调度器初始化
│   └── docs_2/ docs_3/     # Swagger 文档
├── sql/                    # DDL 与样例数据
├── data/                   # SQLite 数据文件（运行时生成）
└── logs/                   # 运行日志
```

---

## 常用命令

```bash
make setup             # 默认目标：安装工具 + 拉取依赖
make deps              # 安装 Go 与前端依赖
make update-deps       # 更新依赖

make dev_frontend      # 前端样式热更新
make dev_backend       # 启动后端
make dev_reporter      # 启动 IP 上报端
make dev_forwarder     # 启动端口转发端

make build-frontend    # 构建前端静态资源
make build-all         # 交叉编译 6 个平台（含 UPX 压缩与打包）
make clean             # 清理 bin/

make generate          # 重新生成 Swagger 与 gRPC 代码
make generate_ent      # 重新生成 Ent 代码
make grpc_generate_all # 重新生成 gRPC 代码

make check-go          # go vet + golangci-lint
make check-node        # pnpm audit / outdated / prettier / knip（未使用依赖）
make format            # 格式化 Go / 前端 / YAML / JSON / SQL / Shell / Markdown / Dockerfile
```

构建产物命名：`{backend|reporter|forwarder}-{os}-{arch}[.exe]`，输出到 `bin/`；版本号、Git Commit、构建时间通过 ldflags 注入 `internal/global`。

发布用的全平台产物由 GoReleaser 生成在 `dist/`（`make goreleaser-snapshot` / `make goreleaser-release`）：压缩包命名同上，另带 Linux 系统包 `deb` / `rpm` / `apk` / `pkg.tar.zst`（包名 `backend` / `reporter` / `forwarder`，二进制装在 `/usr/bin`），以及 `SHA256SUMS`。本地非 CI 环境跑 `make goreleaser-snapshot` 只编当前平台（`goreleaser build --single-target`，不生成压缩包与系统包）；要本地出全平台产物用 `CI=true make goreleaser-snapshot`。

---

## 文档

更多细节见 `docs/` 下的专题文档：

| 文档                                    | 内容                                                                                 |
|-----------------------------------------|--------------------------------------------------------------------------------------|
| [架构与模块设计](docs/architecture.md)  | 系统概览、代码分层、数据层、核心业务链路、调度与厂商适配、扩展指南                   |
| [CI/CD 流水线](docs/ci-cd.md)           | CNB 与 GitHub Actions 全部工作流、版本号规则、发布与镜像构建流程、覆盖率分工、已知坑 |
| [CI 基础镜像](docs/ci-image.md)         | GitHub / CNB 统一 CI 镜像的使用口径与流水线配置注意事项                              |
| [Go 覆盖率实现](docs/coverage-go.md)    | Go 单测覆盖率上报与红线门禁方案（可复用到其他项目的实现文案）                        |
| [前端覆盖率实现](docs/coverage-node.md) | 前端（原生 JS）单测覆盖率方案：Node 内置 test runner + c8                            |

---

## 仓库

| 平台    | 地址                              |
|---------|-----------------------------------|
| CNB     | https://cnb.cool/dtapp/ip_vibe    |
| GitHub  | https://github.com/dtapps/ip_vibe |
| Gitea   | https://gitea.com/dtapps/ip_vibe  |
| GitLab  | https://gitlab.com/dtapps/ip_vibe |
| Gitee   | https://gitee.com/dtapps/ip_vibe  |
| GitCode | https://gitcode.com/dtapp/ip_vibe |

## 许可证

见 [LICENSE](LICENSE) 文件。

## Star 走势

<details>
<summary><b>点击展开查看图表</b></summary>

**CNB**

[![CNB Star 走势](https://commit.cool/stars/dtapp/ip_vibe)](https://cnb.cool/dtapp/ip_vibe)

**GitHub**

[![Star 走势](https://api.star-history.com/chart?repos=dtapps/ip_vibe&type=date&legend=top-left)](https://www.star-history.com/?repos=dtapps%2Fip_vibe&type=date&legend=top-left)

</details>

## 下载量统计图

<details>
<summary><b>点击展开查看图表</b></summary>

[![下载量统计图](https://release-monitor.com/chart/dtapps/ip_vibe.svg?stable=5)](https://release-monitor.com/#/dtapps/ip_vibe)

</details>
