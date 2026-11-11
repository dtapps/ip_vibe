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

| 工作流      | 文件              | 触发                  | 作用                                             |
|-------------|-------------------|-----------------------|--------------------------------------------------|
| PushRequest | `push.yml`        | push / PR（所有分支） | `make deps → format → check`，质量门禁           |
| Coverage    | `coverage.yml`    | push / PR             | Go + Node 覆盖率与 JUnit 测试报告上传 Codecov    |
| Coveralls   | `coveralls.yml`   | push / PR             | 同口径覆盖率以 lcov 上传 Coveralls（独立工作流） |
| Build Image | `build-image.yml` | 手动（输入版本号）    | linux/amd64 二进制 → 三服务运行镜像推 GHCR       |
| Release     | `release.yml`     | 手动（输入版本号）    | 全平台压缩包并发布 GitHub Release                |

另有 Dependabot（配置在 `.github/dependabot.yml`，不在 workflows 目录）：每日检查三类依赖并自动提升级 PR（label `dependencies`）——Go 模块（根 `go.mod`）、前端 npm（`cmd/frontend`）、GitHub Actions 用到的第三方 action 版本。

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

唯一实现是 [scripts/ci/resolve-prerelease.sh](../scripts/ci/resolve-prerelease.sh)：GitHub 的 `release.yml` / `build-image.yml` 与 CNB 的 `release.yml` 都 source 这一份；CNB 的 `build-image.yml` 内联了同口径的正则。规则：

| 输入               | 归一化 TAG       | 类型   | 镜像标签                            | Release         |
|--------------------|------------------|--------|-------------------------------------|-----------------|
| `1.2.3` / `v1.2.3` | `v1.2.3`         | 正式版 | 推 `:v1.2.3` 并覆盖 `:latest`       | 不勾 prerelease |
| `v1.2.3-rc.1`      | `v1.2.3-rc.1`    | 预发布（rc） | 只推 `:v1.2.3-rc.1`，不动 `:latest` | 勾 prerelease   |
| `v1.2.3-beta.1`    | `v1.2.3-beta.1`  | 预发布（beta） | 同上                                | 勾 prerelease   |
| `v1.2.3-alpha.1`   | `v1.2.3-alpha.1` | 预发布（alpha） | 同上                                | 勾 prerelease   |
| `v1.2.3-nightly.20260101` | `v1.2.3-nightly.20260101` | 预发布（nightly） | 同上 | 勾 prerelease |

判定细则：

- 后缀 `-rc` / `-beta` / `-alpha` / `-nightly` 才算预发布，大小写不敏感；`rc1`、`rc.1`、裸 `rc`、`nightly.20260101` 都能识别（正则 `-(rc|beta|alpha|nightly)([.-]?[0-9A-Za-z.-]*)$`）。
- 不带这四种后缀的一律按正式版处理：`v1.2.3-hotfix`、`v2.0.0-core` 不会被误判。
- `nightly` 是独立渠道：`v1.2.3-nightly.20260101` 勾 prerelease、不覆盖 `:latest`，changelog 只跟 nightly 的 tag 对比，不会跨到 rc / beta / alpha。
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

1. **prepare**（CI 镜像容器）：source `scripts/ci/resolve-prerelease.sh` 归一化版本号，输出 TAG / 是否预发布 / 标题。GitHub 与 CNB 的标题同口径：`IPVibe <TAG>`（预发布再加 `(预发布)` 后缀）；tag 本身仍是纯版本号。
2. **build**（CI 镜像容器）：三类缓存（Go module、Go build、pnpm store）→ `make deps` → **先 `make build-frontend`**（前端静态资源由 go:embed 打进 backend 二进制，顺序不能反）→ `make build-all` → `make clean-binaries` 兜底清掉裸二进制 → 生成 `SHA256SUMS` → 上传 artifact。
3. **release**：`fetch-depth: 0` 拉全量 tag 历史 → **校验版本号与 tag**（§3.4）→ `scripts/ci/release-changelog.sh` 生成更新内容（§3.3）→ `scripts/ci/render-release-body.sh` 渲染正文 → `softprops/action-gh-release` 建 tag、上传 19 个附件。

> 正文与 changelog 都**不再内联在工作流里**，统一走 `scripts/ci/` 下的两个脚本，CNB 那份调用同一组脚本，避免两边各写一遍表格/标题/下载链接。

### 3.2 CNB（`release.yml`，分支按钮「打包 Go 项目」）

一条流水线完成：source `resolve-prerelease.sh` 归一化版本号 → `make deps / build-frontend / build-all / clean-binaries` → 生成 `bin/SHA256SUMS` → 校验版本号与 tag（§3.4）→ 调用 `scripts/ci/*.sh` 生成正文 → 生成 `./.cnb-release-options.json` → 内置 `git:release`（`optionsFrom`）建 Release → `cnbcool/attachments` 上传 `bin/*`。

⚠️ **`preRelease` 用 `optionsFrom` 传真布尔值，不用 `${...}` 插值**：`preRelease` 的类型是严格 Boolean，插值产出字符串 `"false"`，字符串在布尔位置上容易被当成真值，正式版会被误勾预发布；`descriptionFromFile` 同理，正文里有反引号/管道符/换行，插值容易被转义毁掉。

附件直链规律（实测可用）：`https://cnb.cool/<repo-slug>/-/releases/download/<tag>/<file>`。

> CNB 的脚本任务不保证跑在 bash 下，三个 `scripts/ci/*.sh` 都按 POSIX sh 写：不用 `pipefail`、进程替换、`[[ ]]`、`BASH_SOURCE`。正文 stage 会把 `ARTIFACT_LAYOUT=triplet` 传给渲染脚本（本仓库是三服务项目，下载表要「组件」列）。

### 3.3 更新内容：按渠道选对比基准（rc / beta / alpha / nightly 的关键）

实现：`scripts/ci/release-changelog.sh`，GitHub 与 CNB 共用一份。只看版本号后缀决定**渠道**，同渠道内对比：

| 场景                            | 对比基准                   | 说明                                      |
|---------------------------------|----------------------------|-------------------------------------------|
| 发 `v1.0.44-rc.2`               | `v1.0.44-rc.1`             | rc 只跟 rc 比                             |
| 发 `v1.0.44-beta.2`             | `v1.0.44-beta.1`           | beta 只跟 beta 比，**不会跨到 rc**        |
| 发 `v1.0.44-nightly.20260102`   | `v1.0.44-nightly.20260101` | nightly 按日期只跟 nightly 比             |
| 发 `v1.0.44-rc.1`（该渠道首次） | 上一个**正式版** `v1.0.43` | 渠道没有历史 tag 时回退                   |
| 发正式版 `v1.0.45`              | 第一个**正式版** `v1.0.43` | 不掺预发布，整轮改动都在                  |
| 取不到基准但仓库里有别的 tag    | **exit 1 报错**            | 不谎报「首次发布」，避免掩盖 clone 缺 tag |

为什么不能用 `git describe --tags --abbrev=0 HEAD^`：它取「最近的可达 tag」，**不区分渠道**。发正式版时会取到刚发的 rc，整轮改动全丢；而且会被 `docker-v1.0.x` 之类的无关 tag 干扰。脚本因此显式按 `^v[0-9]+\.[0-9]+\.[0-9]+(-(rc|beta|alpha|nightly)(\.[0-9]+)?)?$` 白名单 + 渠道过滤，并用 `sort -V` 而不是 `--sort=v:refname`（后者对 `rc.2`/`rc.10` 排序不符合直觉）。

输出格式对齐 `cnbcool/changelog` 插件：按 Conventional Commits type 分组（Features / Bug Fixes / CI / Build System…），行内 `[短号](完整SHA链接) - scope: subject (作者)`，`--no-merges` 去掉 dependabot 的重复合并记录，每组最多 50 条并附溢出提示。

### 3.4 发版前的三道校验

两条流水线都有「校验版本号与 tag」这一步。这三件事出错都**不会让构建失败**，但会发出一个「看起来成功、内容是错的」Release：

1. **tag 没拉全** → changelog 算不出基准，正文退化成「首次发布」→ 直接失败；
2. **tag 已存在** → 重跑会把 tag 挪到新提交上，历史 Release 的对比基准跟着变 → 直接失败；
3. **正式版比已有正式版旧** → 正文里的 diff 是反的 → 直接失败。

### 3.5 正文渲染：`render-release-body.sh`

一份实现供 CNB 与 GitHub 共用，只用环境变量传参。要点：

- **`ARTIFACT_LAYOUT`**：`single`（默认，单二进制项目如 ikuai，下载表 5 列）/ `triplet`（多服务项目如本仓库，多一列「组件」，按 组件→平台→架构 排序）。默认 single 保证与 ikuai 行为一致。
- **「说明」列**：按平台写包类型（Debian/Ubuntu 等 Linux 通用包 / macOS 通用包 / Windows 可执行程序包），均为解压即用的压缩包，不是 `.deb` / `.rpm` / `.msi` 安装包。
- **只把 `.tar.gz` / `.zip` 当附件**：产物目录里混进正文 md、日志时不会误列表。
- **Windows 校验命令按实际产物生成**，不硬编码文件名；PowerShell 输出全大写与 `SHA256SUMS` 的小写等价，正文里显式说明。
- **一律 `printf %s\n` 而非 `echo`**：dash 的 `echo` 会解释反斜杠，`Get-FileHash .\backend-...` 里的 `\b` 会被吞成退格符，显示成 `.ackend-...`。
- 提交号**显示短号、链接用完整 SHA**（缩写长度随仓库规模变化，链接里用缩写部分平台会 404）。

> 脚本自带自检：`sh scripts/ci/render-release-body.sh --self-test`、`sh scripts/ci/resolve-prerelease.sh --self-test`；`sh scripts/ci/release-changelog.sh` 无参数时打印用法。改格式/改判定先跑它们。

---

## 4. 构建运行镜像（Build Image）

三个运行镜像 backend / reporter / forwarder，只构建 linux/amd64（其它平台只出现在 Release 压缩包里）。

| 平台   | 镜像地址                                                                                                                   | Dockerfile                    |
|--------|----------------------------------------------------------------------------------------------------------------------------|-------------------------------|
| GitHub | `ghcr.io/<owner>/<repo>/{backend,reporter,forwarder}`                                                                      | `.github/docker/Dockerfile.*` |
| CNB    | `${CNB_DOCKER_REGISTRY}/${CNB_REPO_SLUG_LOWERCASE}/{backend,reporter,forwarder}`（即 `docker.cnb.cool/dtapp/ip_vibe/...`） | `.cnb/docker/Dockerfile.*`    |

两份 Dockerfile 一一对应，构建上下文都是仓库根目录，唯一差异是基础镜像的 registry（GitHub 拉 `ghcr.io/dtapps/alpine:latest`，CNB 拉 `docker.cnb.cool/dtapp/alpine:latest`；两者由 [dtapps/alpine](https://github.com/dtapps/alpine) 同一套流水线发布、标签契约一致）：装 tzdata + ca-certificates，把 `bin/{service}-linux-amd64` 放到 `/app/ip_vibe`，建 `/app/data`、`/app/logs` 目录。

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

- GitHub `push.yml`：容器 job 统一使用 CI 基础镜像 `ghcr.io/dtapps/golang:latest`，免装工具链；带 Go module、Go build、pnpm store 三类缓存。
- CNB `push.yml`：流水线统一使用 CI 基础镜像 `docker.cnb.cool/dtapp/golang`（`docker.image`），缓存靠 copy-on-write 卷（Go 构建缓存、pnpm store、Go 模块目录）。

`concurrency` 策略：GitHub 侧工作流统一 `cancel-in-progress: false`（同分支新运行排队，不取消旧运行）——上传类、发布类动作跑到一半被砍会留半成品。

---

## 6. 测试覆盖率

### 6.1 职责分工

| 平台   | 工作流                                    | 角色                                                                                                                       |
|--------|-------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| CNB    | `push-coverage.yml` / `pull-coverage.yml` | 硬性红线门禁（`testing:coverage` 内置任务；`key: go` 与 `key: node` 两份并行报告；push 管全量 lines，PR 管增量 diffLines） |
| GitHub | `coverage.yml` → Codecov                  | PR 评论、徽章、patch 状态、Tests 面板（JUnit）                                                                             |
| GitHub | `coveralls.yml` → Coveralls               | 独立上报，与 Codecov 解耦（故障不互相拖垮、凭证隔离）；代价是测试会再跑一遍                                                |

红线现状：CNB 两侧的 `lines` / `diffLines` 目前处于注释状态（等真实基线稳定后按「基线减缓冲」再启用）；Go 报告 `breakIfNoCoverage: false`，Node 报告 `breakIfNoCoverage: true`（报告缺失直接失败，防假绿）。日后启用门禁时，注意两侧数值保持一致。

Codecov 侧的配置在 `.github/codecov.yml`：go / node 双 flag 独立统计（`carryforward: true`，PR 只跑部分测试时沿用历史数据），`ignore` 列表与 coverpkg 排除口径一致（同样排除 4 处 ent 生成目录）；所有 status（project / patch）均为 `informational: true`——**GitHub 侧只做趋势展示与 PR 评论，不做硬门禁，红线只在 CNB**。想在 GitHub 侧也强约束时，去掉对应 status 块的 `informational: true` 即可。

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
- Coveralls 不吃 golang 格式：其 parser 用正则吃掉包前缀，243 个文件会坍缩到同一级，所以先转 lcov（CI 镜像预装的 gcov2lcov）再上传。
- 前端是原生 JS（非 Vite/Vue/TS），用 Node 内置 test runner + c8（vitest 对 CJS 经转译会丢失插桩）；c8 必须在仓库根执行，lcov 的 SF 路径才是仓库根相对。
- Codecov 未启用 Bundle Analysis：前端无打包器、产物无 sourcemap，将来引入打包器后再加。

---

## 7. CI 基础镜像

GitHub 侧五个工作流的容器 job 统一使用 `ghcr.io/dtapps/golang:latest`；CNB 侧 7 处流水线配置统一使用内部制品库的 `docker.cnb.cool/dtapp/golang`（写在 `docker.image`，与 `docker.build` 二选一）。本仓库不再维护任何 CI Dockerfile，也不再构建 CI 镜像；工具链版本由上游镜像统一维护，升级时无需改本仓库。消费方清单、路径假设与权限说明见 [ci-image.md](ci-image.md)。

CNB 侧另用 `docker.volumes` 保留 copy-on-write 缓存（`/root/.cache/go-build`、pnpm store、`/go/pkg/mod`），并特意只缓存模块下载，不覆盖镜像内预装的 `/go/bin` 工具。

---

## 8. CNB 侧其它自动化

| 工作流               | 说明                                                                                                                         |
|----------------------|------------------------------------------------------------------------------------------------------------------------------|
| `code-review.yml`    | PR 时 AI 评审（deepseek-v4.1-flash）：评论上限 10 条、过滤 style/nit 类与低置信度建议；`fail_on_critical: false`，不阻断合并 |
| `cnb-sechook.yml`    | PR 敏感信息与代码扫描统计                                                                                                    |
| `sca.yml`            | push 全量 + PR 增量的组件漏洞 / License 扫描与 SBOM（SPDX 格式）；PR 侧 `failOnSeverity: high` 门禁，push 侧只出报告         |
| `codewiki.yml`       | 推 tag 时生成 Code Wiki（timeout 3h）                                                                                        |
| `knowledge-base.yml` | push 与 Issue 事件时把 `docs/**.md` 索引入知识库（Issue 同步开启）——docs/ 下的文档推送后即可被检索                           |

敏感信息与 SCA 扫描的忽略清单在仓库根 `.scanignore`，由 `.cnb/security/code_scan_config.yml` 的 secrets 与 software-composition-analysis 两处 `scan.ignoreFrom` 指定。

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
| GitLab                              | `.gitlab/` 仅为预留目录，仓库暂无 `.gitlab-ci.yml`                                    |

---

## 10. 已知坑（改流水线前先看）

1. **容器 job 的 git「dubious ownership」**：仓库由 runner 用户落盘、容器内以 root 运行，git 因所有权检查直接 exit 128。构建类工作流（`release.yml` / `build-image.yml`）在 checkout 后用一个「让容器内的 git 可用」步骤登记 `git config --system --add safe.directory '*'`（容器里 root 可写 `/etc/gitconfig`，不依赖 `$HOME`；再补一条 `--global` 的工作区路径），该步骤还会把 `git status` 的结果打进日志，便于确认登记有没有生效；同时 Makefile 的 `BUILD_CMD` 加了 `-buildvcs=false`——go build 默认会把 git 状态盖章进二进制，git 不可用即以 `error obtaining VCS status: exit status 128` 中断（go test 不做 VCS 盖章，所以覆盖率流水线一直是绿的，只有构建会挂）。新增容器 job 里要跑 git 或 go build 时，两道保险都要上。
2. **go:embed 顺序**：`build-frontend` 必须先于后端编译，否则打进二进制的是旧静态资源。
3. **GHCR 镜像名必须全小写**。
4. **changelog 按渠道选基准**（§3.3），不要再退回 `git describe --tags --abbrev=0 HEAD^`——它不区分渠道，发正式版时整轮预发布改动会丢；`fetch-depth: 0` 仍是前提。
5. **前端覆盖率必须单进程直跑**（直接 `node <file>` 或聚合入口）：`node --test` 的 worker 线程里 c8 无法对 CJS 源插桩，行覆盖率恒为 0。
6. **c8 的 include glob 不允许 `../` 逃逸 cwd**，`test:coverage` 脚本要先 `cd` 回仓库根再跑。
7. **CNB 多份覆盖率报告必须用 key/name 区分**（go / node），否则写入同一 meta map 互相覆盖，徽章与数据残缺。
8. **CNB 的 `git:release` 参数名是驼峰 `preRelease`**，写成小写 `prerelease` 会被**静默忽略**，Release 不会打上预发布标记；且 `preRelease` 是严格 Boolean，必须用 `optionsFrom` 传 JSON 里的真布尔值，不要用 `${...}` 插值（字符串 "false" 易被当真理）。改 `options`/`settings` 时务必对着文档核参数名与大小写。
9. **`scripts/ci/*.sh` 必须保持 POSIX sh**：CNB 的 stage 用 `sh`（可能是 dash/ash）执行，不要引入 `[[ ]]`、数组、`BASH_SOURCE`、`pipefail`、进程替换。判定「是否被 source」用 `$0` 的 basename，不能用 `${BASH_SOURCE[0]}`（dash 会 Bad substitution）。
10. **`render-release-body.sh` 只把 `.tar.gz`/`.zip` 当附件**：GitHub 侧的 `release_body.md` 常和附件同放一个目录，不筛会把它当成可下载文件列进表格。
11. **镜像推送 / Release 发布的 concurrency 不设 cancel-in-progress**：宁可排队，不要半路砍——取消留下的「只有部分服务更新」的镜像组合或只有两三个附件的 Release，比排队慢更贵。
