# IPVibe CI/CD 流水线

> 仓库同时在 CNB（cnb.cool）与 GitHub 两条平台上跑 CI：CNB 负责日常门禁与国内制品（配置在 `.cnb/`），GitHub 负责质量检查、镜像与 Release 发布（配置在 `.github/`）。本文是全部工作流的索引与口径说明；覆盖率实现细节见 [coverage-go.md](coverage-go.md) 与 [coverage-node.md](coverage-node.md)，CI 基础镜像见 [ci-image.md](ci-image.md)。

---

## 目录

- [1. 总览](#1-总览)
- [2. 版本号规则](#2-版本号规则)
- [3. 发布二进制包（Release）](#3-发布二进制包release)
- [4. 构建运行镜像（Build Image）](#4-构建运行镜像build-image)
- [5. 日常检查（Push / PR）](#5-日常检查push--pr)
- [6. 测试覆盖率](#6-测试覆盖率)
- [7. CI 基础镜像](#7-ci-基础镜像)
- [8. CNB 侧其它自动化](#8-cnb-侧其它自动化)
- [9. 凭证与外部依赖](#9-凭证与外部依赖)
- [10. 已知坑（改流水线前先看）](#10-已知坑改流水线前先看)

---

## 1. 总览

两套平台共用同一份 Makefile 目标（`deps` / `format` / `check` / `build-*` / `cover-*` / `junit-*`），本地、CNB、GitHub 三处命令口径一致，不会各写一套慢慢漂移。

### 1.1 GitHub Actions（`.github/workflows/`）

| 工作流         | 文件              | 触发                                      | 作用                                             |
|----------------|-------------------|-------------------------------------------|--------------------------------------------------|
| PushRequest    | `push.yml`        | push / PR（所有分支）                     | `make deps → format → check`，质量门禁           |
| Build CI Image | `ci-image.yml`    | CI Dockerfile 变更、每周一 UTC 0 点、手动 | 构建预装工具链的基础镜像推 GHCR                  |
| Coverage       | `coverage.yml`    | push / PR                                 | Go + Node 覆盖率与 JUnit 测试报告上传 Codecov    |
| Coveralls      | `coveralls.yml`   | push / PR                                 | 同口径覆盖率以 lcov 上传 Coveralls（独立工作流） |
| Build Image    | `build-image.yml` | 手动（输入版本号）                        | linux/amd64 二进制 → 三服务运行镜像推 GHCR       |
| Release        | `release.yml`     | 手动（输入版本号）                        | 全平台压缩包并发布 GitHub Release                |

### 1.2 CNB（`.cnb/workflows/`，经根目录 `.cnb.yml` include）

| 工作流               | 触发                                      | 作用                                   |
|----------------------|-------------------------------------------|----------------------------------------|
| `push.yml`           | push                                      | deps → format → check                  |
| `push-coverage.yml`  | push                                      | Go + Node 覆盖率上报（全量红线）       |
| `pull-coverage.yml`  | pull_request                              | 覆盖率增量红线（diffLines）            |
| `code-review.yml`    | pull_request                              | AI 代码评审，结果评论到 PR             |
| `cnb-sechook.yml`    | pull_request                              | 敏感信息 / 代码扫描统计                |
| `sca.yml`            | push（全量）、pull_request.target（增量） | 组件漏洞扫描、License 扫描、SBOM       |
| `build-image.yml`    | 分支按钮「打包 Docker 镜像」              | 三服务镜像推 CNB 制品库                |
| `release.yml`        | 分支按钮「打包 Go 项目」                  | 全平台压缩包 + CNB Release + 附件      |
| `codewiki.yml`       | tag_push                                  | 生成 Code Wiki                         |
| `knowledge-base.yml` | push、issue.open、issue.comment           | `docs/**.md` 索引入知识库 + Issue 同步 |

CNB 的两个发布按钮定义在 `.cnb/web_trigger.yml`（权限 owner / master / developer，输入版本号）。

---

## 2. 版本号规则

唯一实现是 [scripts/ci/resolve-prerelease.sh](../scripts/ci/resolve-prerelease.sh)：GitHub 的 `release.yml` / `build-image.yml` 都 source 这一份；CNB 侧在两个发布流水线里内联了同口径的正则。规则：

| 输入               | 归一化 TAG       | 类型   | 镜像标签                            | Release         |
|--------------------|------------------|--------|-------------------------------------|-----------------|
| `1.2.3` / `v1.2.3` | `v1.2.3`         | 正式版 | 推 `:v1.2.3` 并覆盖 `:latest`       | 不勾 prerelease |
| `v1.2.3-rc.1`      | `v1.2.3-rc.1`    | 预发布 | 只推 `:v1.2.3-rc.1`，不动 `:latest` | 勾 prerelease   |
| `v1.2.3-beta.1`    | `v1.2.3-beta.1`  | 预发布 | 同上                                | 勾 prerelease   |
| `v1.2.3-alpha.1`   | `v1.2.3-alpha.1` | 预发布 | 同上                                | 勾 prerelease   |

判定细则：

- 后缀 `-rc` / `-beta` / `-alpha` 才算预发布，大小写不敏感；`rc1`、`rc.1`、裸 `rc` 都能识别（正则 `-(rc|beta|alpha)([.-]?[0-9A-Za-z.-]*)$`）。
- 不带这三种后缀的一律按正式版处理：`v1.2.3-hotfix`、`v2.0.0-core` 不会被误判。
- 输入自动补 `v` 前缀、去空白字符；空输入直接报错（历史上空输入会打出 `:v` 这种非法标签）。
- 预发布的核心目的：镜像不覆盖 `:latest`、Release 标记 prerelease，未经充分验证的版本不会推给线上用户。

改了脚本里的正则先跑自检：

```bash
bash scripts/ci/resolve-prerelease.sh --self-test
```

---

## 3. 发布二进制包（Release）

产物：3 个服务 × 6 个平台 = 18 个压缩包，GitHub 侧额外带 `SHA256SUMS`。

| 服务                           | 平台                                                                               | 压缩格式                    |
|--------------------------------|------------------------------------------------------------------------------------|-----------------------------|
| backend / reporter / forwarder | linux/amd64、linux/arm64、windows/amd64、windows/arm64、darwin/amd64、darwin/arm64 | Windows 为 zip，其余 tar.gz |

命名 `{service}-{os}-{arch}[.exe]`。版本号、Git Commit、构建时间通过 ldflags 注入 `vibe_ip/internal/global`；编译走 `make build-all`（`CGO_ENABLED=0` 静态链接、`-trimpath`、`-buildvcs=false`）。

### 3.1 GitHub（`release.yml`，手动 workflow_dispatch）

三个 job 串行：

1. **prepare**（CI 镜像容器）：source `resolve-prerelease.sh` 归一化版本号，输出 TAG / 是否预发布 / 标题（预发布标题带 `(预发布)` 后缀）。
2. **build**（CI 镜像容器）：三类缓存（Go module、Go build、pnpm store）→ `make deps` → **先 `make build-frontend`**（前端静态资源由 go:embed 打进 backend 二进制，顺序不能反）→ `make build-all` → `make clean-binaries` 兜底清掉裸二进制 → 生成 `SHA256SUMS` → 上传 artifact。
3. **release**：`fetch-depth: 0` 拉全量 tag 历史 → 用 `git describe --tags --abbrev=0 'HEAD^'` 找上一个 tag、`git log` 生成 changelog → 拼出含下载表与校验说明的正文 → `softprops/action-gh-release` 建 tag、上传 19 个附件。

> 为什么不用 `git tag --sort=-v:refname | head -2 | tail -1` 取「上一个 tag」：那取的是全仓库第二新的 tag，跟当前 HEAD 没有关系。重跑发布（tag 已建过）时会把刚发的 tag 自己当成上一个版本，changelog 段落直接变空。

### 3.2 CNB（`release.yml`，分支按钮「打包 Go 项目」）

一条流水线完成：归一化版本号 → `make deps / build-frontend / build-all / clean-binaries` → `cnbcool/changelog` 插件生成日志 → 内置 `git:release` 建 Release（`prerelease: ${INTERIOR_IS_PRERELEASE}`）→ `cnbcool/attachments` 上传 `bin/*`。

与 GitHub 的差异：CNB 不生成 `SHA256SUMS`；changelog 用插件而非 git log。

---

## 4. 构建运行镜像（Build Image）

三个运行镜像 backend / reporter / forwarder，只构建 linux/amd64（其它平台只出现在 Release 压缩包里）。

| 平台   | 镜像地址                                                                                                                   | Dockerfile                    |
|--------|----------------------------------------------------------------------------------------------------------------------------|-------------------------------|
| GitHub | `ghcr.io/<owner>/<repo>/{backend,reporter,forwarder}`                                                                      | `.github/docker/Dockerfile.*` |
| CNB    | `${CNB_DOCKER_REGISTRY}/${CNB_REPO_SLUG_LOWERCASE}/{backend,reporter,forwarder}`（即 `docker.cnb.cool/dtapp/ip_vibe/...`） | `.cnb/docker/Dockerfile.*`    |

两份 Dockerfile 一一对应，构建上下文都是仓库根目录，唯一差异是基础镜像（Docker Hub 的 `alpine:3.22` vs CNB 内部的 `docker.cnb.cool/dtapp/alpine`）：装 tzdata + ca-certificates，把 `bin/{service}-linux-amd64` 放到 `/app/ip_vibe`，建 `/app/data`、`/app/logs` 目录。

### 4.1 GitHub（`build-image.yml`）

- **binaries**（容器 job）：解析版本号 → 缓存 → deps → build-frontend → `make build-linux-amd64` → 上传 artifact。
- **image**（普通 runner——容器 job 里没有 docker daemon）：下载 artifact、校验三个二进制都在 → 三个服务的镜像标签一次性算好写入 `$GITHUB_ENV`（避免三个 build 步骤各写一遍「有后缀就不推 latest」的条件表达式，三处写不一致）→ buildx 构建推送（gha cache 按服务分 scope）。

标签规则：正式版推 `:{TAG}` + `:latest`；预发布只推 `:{TAG}`。GHCR 要求镜像名全小写，代码里用 `tr '[:upper:]' '[:lower:]'` 处理。

> 刻意不做旧版本清理：这里的每个标签都对应一个正式/预发布版本，删掉会让仍在用旧版本的用户无法重新拉取。GHCR 存储配额吃紧时，再按 actions/delete-package-versions 只清 untagged 版本。

### 4.2 CNB（`build-image.yml`）

按钮触发后单条流水线完成：deps → build-frontend → build-linux-amd64 → `docker login`（CNB 内置 token）→ 三个 `docker build` + `docker push`；预发布版本跳过 latest 标签推送。

---

## 5. 日常检查（Push / PR）

两侧口径一致：`make deps → make format → make check`（先格式化，再跑 golangci-lint 等静态检查，实现见 `scripts/check.sh`）。

- GitHub `push.yml`：容器 job 直接用 CI 基础镜像（`ghcr.io/<owner>/<repo>/ci:latest`），免装工具链；带 Go module、Go build、pnpm store 三类缓存。
- CNB `push.yml`：runner 上用 `.cnb/ci/Dockerfile` 现场构建执行环境，缓存靠 copy-on-write 卷（Go 构建缓存、pnpm store、Go 模块目录）。

`concurrency` 策略：`push.yml` / `ci-image.yml` 同分支新运行会取消旧运行（动作幂等）；`coverage.yml` / `coveralls.yml` / `release.yml` / `build-image.yml` 不取消（上传类、发布类动作跑到一半被砍会留半成品）。

---

## 6. 测试覆盖率

### 6.1 职责分工

| 平台   | 工作流                                    | 角色                                                                                                                       |
|--------|-------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| CNB    | `push-coverage.yml` / `pull-coverage.yml` | 硬性红线门禁（`testing:coverage` 内置任务；`key: go` 与 `key: node` 两份并行报告；push 管全量 lines，PR 管增量 diffLines） |
| GitHub | `coverage.yml` → Codecov                  | PR 评论、徽章、patch 状态、Tests 面板（JUnit）                                                                             |
| GitHub | `coveralls.yml` → Coveralls               | 独立上报，与 Codecov 解耦（故障不互相拖垮、凭证隔离）；代价是测试会再跑一遍                                                |

红线现状：CNB 两侧的 `lines` / `diffLines` 目前处于注释状态（等真实基线稳定后按「基线减缓冲」再启用）；Go 报告 `breakIfNoCoverage: false`，Node 报告 `breakIfNoCoverage: true`（报告缺失直接失败，防假绿）。日后启用门禁时，注意两侧数值保持一致。

### 6.2 命令口径唯一在 Makefile

| Makefile 目标   | 产物                                                     | 消费方                               |
|-----------------|----------------------------------------------------------|--------------------------------------|
| `cover-go`      | `coverage.out`（golang 格式）                            | CNB testing:coverage、GitHub Codecov |
| `cover-go-lcov` | `coverage-lcov.info`                                     | Coveralls                            |
| `cover-node`    | `cmd/frontend/coverage/lcov.info`                        | CNB / Codecov / Coveralls            |
| `junit-go`      | `go-test-report.xml`（gotestsum，顺带产出 coverage.out） | Codecov Tests 面板                   |
| `junit-node`    | `cmd/frontend/coverage/junit.xml`                        | Codecov Tests 面板                   |

关键口径（详见 Makefile 注释与两篇 coverage 文档）：

- coverpkg 范围为 `./cmd/... ./internal/... ./pkg/...`，排除 4 处 ent 生成目录（ent 占全树语句约 73%，计入会把业务覆盖率从约 27% 稀释到约 5%）；不用 `./...`（会带上 `tests/` 无 build tag 的集成测试，CI 里跑会挂）。
- Coveralls 不吃 golang 格式：其 parser 用正则吃掉包前缀，243 个文件会坍缩到同一级，所以先转 lcov（gcov2lcov）再上传。
- 前端是原生 JS（非 Vite/Vue/TS），用 Node 内置 test runner + c8（vitest 对 CJS 经转译会丢失插桩）；c8 必须在仓库根执行，lcov 的 SF 路径才是仓库根相对。
- Codecov 未启用 Bundle Analysis：前端无打包器、产物无 sourcemap，将来引入打包器后再加。

---

## 7. CI 基础镜像

GitHub 侧五个工作流的容器 job 共用一个预构建镜像 `ghcr.io/<owner>/<repo>/ci:latest`（golang:1.27.3-trixie + Node 26 + pnpm + golangci-lint、gotestsum、gcov2lcov、markdownfmt、yamlfmt、shfmt、dockerfmt、upx、nfpm、wixl 等）。由 `ci-image.yml` 在 Dockerfile 变更时自动重建并推送，每周一定时拉基础镜像安全更新；每次构建附 sha 标签、只保留最近 5 个版本。

**首次使用前必须先手动跑一次 Build CI Image**，否则消费方 pull 不到镜像。详细内容、版本升级方式与「加回 arm64」见 [ci-image.md](ci-image.md)。

CNB 侧不用预推镜像：每条流水线 `docker.build: .cnb/ci/Dockerfile` 现场构建执行环境（多配了 daocloud / 腾讯云镜像加速；不含 gotestsum、gcov2lcov——CNB 不产 JUnit / lcov）。

---

## 8. CNB 侧其它自动化

| 工作流               | 说明                                                                                                                         |
|----------------------|------------------------------------------------------------------------------------------------------------------------------|
| `code-review.yml`    | PR 时 AI 评审（deepseek-v4.1-flash）：评论上限 10 条、过滤 style/nit 类与低置信度建议；`fail_on_critical: false`，不阻断合并 |
| `cnb-sechook.yml`    | PR 敏感信息与代码扫描统计                                                                                                    |
| `sca.yml`            | push 全量 + PR 增量的组件漏洞 / License 扫描与 SBOM（SPDX 格式）；PR 侧 `failOnSeverity: high` 门禁，push 侧只出报告         |
| `codewiki.yml`       | 推 tag 时生成 Code Wiki（timeout 3h）                                                                                        |
| `knowledge-base.yml` | push 与 Issue 事件时把 `docs/**.md` 索引入知识库（Issue 同步开启）——docs/ 下的文档推送后即可被检索                           |

`.cnb/settings.yml` 配置 Issue 按钮悬浮图、reaction 图与 NPC 角色。

---

## 9. 凭证与外部依赖

| 项                                  | 说明                                                                                  |
|-------------------------------------|---------------------------------------------------------------------------------------|
| `GITHUB_TOKEN`                      | GitHub 内置；Build Image 需 `packages: write` 推 GHCR，Release 需 `contents: write`   |
| `CODECOV_TOKEN`                     | `coverage.yml` 上传 Codecov                                                           |
| `COVERALLS_REPO_TOKEN`              | `coveralls.yml`；缺失时自动跳过上传（`continue-on-error`），不影响其它步骤            |
| `CNB_TOKEN` / `CNB_TOKEN_USER_NAME` | CNB 内置，用于 `docker login` 制品库                                                  |
| CNB imports                         | 多条流水线 `imports: https://cnb.cool/dtapp/envs/-/blob/main/docker.yml` 引入环境变量 |
| GitLab                              | `.gitlab/ci/Dockerfile` 仅为预留，仓库暂无 `.gitlab-ci.yml`                           |

---

## 10. 已知坑（改流水线前先看）

1. **容器 job 的 git「dubious ownership」**：仓库由 runner 用户落盘、容器内以 root 运行，git 因所有权检查直接 exit 128。构建类工作流（`release.yml` / `build-image.yml`）在 checkout 后用一个「让容器内的 git 可用」步骤登记 `git config --system --add safe.directory '*'`（容器里 root 可写 `/etc/gitconfig`，不依赖 `$HOME`；再补一条 `--global` 的工作区路径），该步骤还会把 `git status` 的结果打进日志，便于确认登记有没有生效；同时 Makefile 的 `BUILD_CMD` 加了 `-buildvcs=false`——go build 默认会把 git 状态盖章进二进制，git 不可用即以 `error obtaining VCS status: exit status 128` 中断（go test 不做 VCS 盖章，所以覆盖率流水线一直是绿的，只有构建会挂）。新增容器 job 里要跑 git 或 go build 时，两道保险都要上。
2. **go:embed 顺序**：`build-frontend` 必须先于后端编译，否则打进二进制的是旧静态资源。
3. **GHCR 镜像名必须全小写**。
4. **changelog 的「上一个 tag」**要用 `git describe --tags --abbrev=0 'HEAD^'` 且 `fetch-depth: 0`，见 §3.1 的反例。
5. **前端覆盖率必须单进程直跑**（直接 `node <file>` 或聚合入口）：`node --test` 的 worker 线程里 c8 无法对 CJS 源插桩，行覆盖率恒为 0。
6. **c8 的 include glob 不允许 `../` 逃逸 cwd**，`test:coverage` 脚本要先 `cd` 回仓库根再跑。
7. **CNB 多份覆盖率报告必须用 key/name 区分**（go / node），否则写入同一 meta map 互相覆盖，徽章与数据残缺。
8. **镜像推送 / Release 发布的 concurrency 不设 cancel-in-progress**：宁可排队，不要半路砍——取消留下的「只有部分服务更新」的镜像组合或只有两三个附件的 Release，比排队慢更贵。
