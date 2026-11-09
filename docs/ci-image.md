# CI 基础镜像

> GitHub 与 CNB 两侧流水线统一使用同一套 CI 基础镜像：GitHub 用 `ghcr.io/dtapps/golang:latest`，CNB 用 `docker.cnb.cool/dtapp/golang`。运行镜像（backend / reporter / forwarder）的 `FROM` 同样统一到 dtapps 维护的 alpine：GitHub 用 `ghcr.io/dtapps/alpine:latest`，CNB 用 `docker.cnb.cool/dtapp/alpine:latest`。本仓库不再维护任何 CI Dockerfile，也不再构建、推送 CI 镜像。

---

## 1. 消费方

| 平台   | 消费方（流水线 job）                                                                                                        | 配置                                            |
|--------|-----------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------|
| GitHub | `push.yml`、`codecov.yml`、`coveralls.yml`、`release.yml`（prepare / build）、`build-image.yml`（binaries）                 | `container.image: ghcr.io/dtapps/golang:latest` |
| CNB    | `push.yml`、`push-coverage.yml`（go / node）、`pull-coverage.yml`（go / node）、`build-image.yml`、`release.yml`（共 7 处） | `docker.image: docker.cnb.cool/dtapp/golang`    |

两侧共用同一份 Makefile 目标，因此镜像按同一口径准备：Go / Node / pnpm 与 `make check`、`junit-go`、`cover-go-lcov`（gcov2lcov）等目标所需的工具。工具链版本由上游镜像统一维护，升级时无需改本仓库。

## 2. 流水线侧假设

- GitHub：`container.image` 直接拉镜像；`GO_PATH_DIR` / `GO_BIN_DIR` 写入 `$GITHUB_ENV` 供 Makefile 与缓存步骤引用（按镜像内 `GOPATH=/go`、工具在 `/go/bin` 编写）；`pnpm config get store-dir` 定位 pnpm 缓存；`permissions` 声明 `packages: read`（拉 GHCR 镜像；公开镜像可匿名拉，写上以防仓库转私有）。
- CNB：`docker.image` 直接拉镜像（`docker.build` 才需要 Dockerfile，两者可同时写、此时 `image` 是构建失败时的回退）；`docker.volumes` 用 copy-on-write 卷缓存 `/root/.cache/go-build`、`/root/.local/share/pnpm/store`、`/go/pkg/mod`，其中只缓存模块下载、不覆盖镜像内预装的 `/go/bin` 工具。

若上游镜像调整了这些路径，或不再预装所需工具，需同步调整流水线：修正缓存路径，或补回 `setup-go` / `setup-node` / `make tool-deps`（GitHub 侧）与相应的依赖安装步骤（CNB 侧）。

## 3. 平台差异

- 镜像来源：GitHub 从 GHCR 拉 `ghcr.io/dtapps/golang:latest`，CNB 从内部制品库拉 `docker.cnb.cool/dtapp/golang`；两侧都不再现场构建 Dockerfile。
- 认证：CNB 侧镜像若设为私有，需确保流水线具备该制品库的读取权限；GitHub 侧保留 `packages: read`。
- 已知坑：GitHub 容器 job 的 git「dubious ownership」等见 [ci-cd.md](ci-cd.md) §10；CNB 流水线无此问题。

## 4. 运行镜像的基础镜像

三个运行镜像（backend / reporter / forwarder）的 `FROM` 同样是 dtapps 维护的镜像，只是换成 [dtapps/alpine](https://github.com/dtapps/alpine)：

| 平台   | 基础镜像                              | 消费方                                                   |
|--------|---------------------------------------|----------------------------------------------------------|
| GitHub | `ghcr.io/dtapps/alpine:latest`        | `.github/docker/Dockerfile.{backend,reporter,forwarder}` |
| CNB    | `docker.cnb.cool/dtapp/alpine:latest` | `.cnb/docker/Dockerfile.{backend,reporter,forwarder}`    |

两侧 Dockerfile 只差这一行 `FROM` 的 registry，其余（`apk add tzdata ca-certificates`、`COPY bin/<服务>-linux-amd64`、`ENV TZ=Asia/Shanghai`、`ENTRYPOINT`）完全一致。上游镜像已经预装 tzdata、ca-certificates、curl、drill、git、nc 并把 `TZ` 设为 `Asia/Shanghai`，所以 `apk add` 这一步是幂等兜底，不依赖上游是否保留这些包。标签契约（`<alpine 版本>` / `<major.minor>` / `latest` / `<版本>-linux-{amd64,arm64}`）见上游 [images-and-tags.md](https://github.com/dtapps/alpine/blob/main/docs/alpine/images-and-tags.md)；两个 registry 各自构建、digest 不通用，需要固定到某一次构建时按 digest 拉取。
