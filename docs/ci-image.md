# CI 基础镜像

> GitHub 侧工作流以容器 job 复用预构建镜像 `ghcr.io/<owner>/<repo>/ci:latest`，免去每次重装系统库、Node 与十几个 Go 工具。镜像的构建与推送由 `.github/workflows/ci-image.yml` 负责，Dockerfile 在 `.github/ci/Dockerfile`。

---

## 目录

- [1. 用途与消费方](#1-用途与消费方)
- [2. 镜像内容与版本升级](#2-镜像内容与版本升级)
- [3. 平台范围](#3-平台范围)
  - [3.1 为什么只构建 linux/amd64](#31-为什么只构建-linuxamd64)
  - [3.2 加回 arm64](#32-加回-arm64)
- [4. 触发、缓存与旧版本清理](#4-触发缓存与旧版本清理)
- [5. 使用前提与故障排查](#5-使用前提与故障排查)
- [6. 与 CNB 侧的差异](#6-与-cnb-侧的差异)

---

## 1. 用途与消费方

| 消费方（容器 job）               | 用法                                                           |
|----------------------------------|----------------------------------------------------------------|
| `push.yml`（PushRequest）        | `container: image: ghcr.io/${{ github.repository }}/ci:latest` |
| `coverage.yml`                   | 同上                                                           |
| `coveralls.yml`                  | 同上                                                           |
| `release.yml`（prepare / build） | 同上                                                           |
| `build-image.yml`（binaries）    | 同上                                                           |

镜像内 `GOPATH=/go`，工具在 `/go/bin` 且已在 PATH。工作流里只需把 `GO_PATH_DIR` / `GO_BIN_DIR` 写入 `$GITHUB_ENV` 供 Makefile 引用；`permissions` 需声明 `packages: read`（拉 GHCR 镜像；公开镜像匿名亦可拉，写上以防仓库转私有）。

## 2. 镜像内容与版本升级

基础镜像 `golang:1.27.3-trixie`；apt 安装 wget / curl / git / make / zip / unzip / yamllint / wixl（wixl 用于 Windows msi 打包）等；预装工具全部通过 Dockerfile 顶部的 `ARG` 写死版本：

| 工具                                      | ARG（当前版本）                 | 用途                                                   |
|-------------------------------------------|---------------------------------|--------------------------------------------------------|
| Node.js                                   | `NODE_VERSION=26.12.1`          | 前端构建 / 测试                                        |
| pnpm                                      | `PNPM_VERSION=12.15.0`          | 前端依赖管理                                           |
| golangci-lint                             | `GOLANGCI_LINT_VERSION=v2.15.0` | `make check`                                           |
| gotestsum                                 | `GOTESTSUM_VERSION=v1.13.0`     | `make junit-go`（v1.12.0 在 Go 1.27 编译失败，勿降级） |
| gcov2lcov                                 | `GCOV2LCOV_VERSION=v1.1.1`      | `make cover-go-lcov`                                   |
| markdownfmt / yamlfmt / shfmt / dockerfmt | 见 Dockerfile                   | `make format`                                          |
| UPX                                       | `UPX_VERSION=5.2.1`             | 二进制压缩                                             |
| nfpm                                      | `NFPM_VERSION=2.48.0`           | 包打包                                                 |

要点：

- **改版本 = 改 Dockerfile**：版本都写死在 ARG 里，push 到改动的 Dockerfile 会自然触发 ci-image.yml 重建（工作流按 paths 过滤 `.github/ci/Dockerfile`）。
- Go 工具链在同一层内 `go install` 后清缓存，不留独立缓存层；最后一步卸载 wget / xz-utils 并清 apt 缓存压体积。
- UPX / nfpm / Node 的安装脚本已按 `uname -m` 写了 x86_64 与 aarch64 双分支，加回 arm64 时无需改 Dockerfile（见 §3.2）。

## 3. 平台范围

### 3.1 为什么只构建 linux/amd64

2026-10-23 起只构建 linux/amd64，且不再 `setup-qemu`：

- 消费方（push.yml / coverage.yml / coveralls.yml / release.yml / build-image.yml）的 `runs-on` 全是 `ubuntu-latest`（amd64），arm64 镜像建出来没有任何人拉；
- 多平台构建的 arm64 腿走 QEMU 模拟，apt 装系统库 + `go install` 十几个工具会慢到离谱，纯粹白烧 Actions 额度。

### 3.2 加回 arm64

确认真有 arm64 消费方（例如 self-hosted arm64 runner）时，按以下步骤加回：

1. `ci-image.yml` 里恢复 `docker/setup-qemu-action` 步骤（放在 setup-buildx 之前）。
2. `build-push-action` 的 `platforms` 改为 `linux/amd64,linux/arm64`。
3. Dockerfile 无需改动：UPX / nfpm / Node 安装逻辑已内置 aarch64 分支（见 §2）。
4. 预期代价：arm64 腿全程 QEMU 模拟，单次构建预计 30 分钟以上；如只是想留个 arm64 备用，建议改为定时构建而非每次触发。

## 4. 触发、缓存与旧版本清理

- **触发**：push / PR 改动 `.github/ci/Dockerfile`；每周一 UTC 0 点定时重建（拉基础镜像的 apt 安全更新）；手动 workflow_dispatch。
- **标签**：默认分支推送 `latest` + `sha-` 短标签（便于回滚）；PR 无推送权限，仅构建校验。
- **缓存**：buildx 的 gha cache（`cache-from: type=gha` / `cache-to: type=gha,mode=max`）加速层重建。
- **旧版本清理**：每次构建推一个新 sha 标签，`actions/delete-package-versions` 只保留最近 5 个版本，避免撑爆 GHCR 存储配额。

  > 注意：GHCR 包名是 `<repo>/ci`（含仓库名那段），不是末尾的 `ci`，`package-name` 写错会找不到包。

## 5. 使用前提与故障排查

- **首次使用前必须先手动跑一次 ci-image.yml**，否则消费方容器 job 会因 pull 不到镜像而失败（换 owner / 改仓库名后同样要先跑一次）。
- 容器 job 拉镜像需要 `packages: read` 权限，各工作流已声明。
- 升级工具版本后如 `make check` / `make junit-go` 行为异常，先比对工具与 Go 版本的兼容性（前车之鉴：gotestsum v1.12.0 在 Go 1.27 下编译失败，需 v1.13.0）。
- 构建缓存未命中导致变慢属正常，gha cache 有 7 天未用即淘汰的策略。

## 6. 与 CNB 侧的差异

CNB 流水线不消费预推镜像，而是每条流水线 `docker.build: .cnb/ci/Dockerfile` 现场构建执行环境：

- 基础同为 `golang:1.27.3-trixie`，工具版本与 GitHub 侧基本一致，但**不含 gotestsum / gcov2lcov**（CNB 不产 JUnit / lcov 报告）；
- 配置了 daocloud 文件镜像（`GITHUB_MIRROR` ARG）与腾讯云 GOPROXY 加速下载；
- 缓存靠 copy-on-write 卷（`/root/.cache/go-build`、pnpm store、`/go/pkg/mod`），并特意只缓存模块下载，不覆盖镜像内预装的 `/go/bin` 工具。
