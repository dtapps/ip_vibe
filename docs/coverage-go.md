# Go 单元测试覆盖率实现文案（可复制到其他项目交给 AI 实现）

> 来源：CertFlow 项目实战方案（CNB 云原生构建 + testing:coverage 内置任务）。

# 任务：为 Go 项目实现单元测试覆盖率上报与红线门禁（CNB 云原生构建）

## 背景与目标

- 构建平台是 CNB（cnb.cool），流水线配置在仓库根的 `.cnb.yml`（或 `.cnb/workflows/*.yml`，由 .cnb.yml include）。
- 使用 CNB 内置任务 `type: testing:coverage` 上报覆盖率：push 事件上报全量覆盖率并设红线，pull_request 事件额外计算增量覆盖率并设红线（diffLines 仅 PR 类事件生效）。
- 红线采用「棘轮策略」：初始值贴着当前基线往下留余量，防止误伤，后续随覆盖率提升逐步上调。

## 实现步骤

### 第 1 步：先测量当前覆盖率基线（定红线前必须做）

在本地运行并记录 total 覆盖率（按需调整测试范围）：

```bash
go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out | tail -1
```

### 第 2 步：覆盖率命令直接内联到 CI（不新增 Makefile 目标）

覆盖率是 CI 质量门禁，不属于本地 Dev 工作流，因此**不在 Makefile 里新增 coverage 目标**，把生成命令直接写进流水线 stage（即原 Makefile coverage 的内容，`$$` 还原为 shell 的 `$`）：

```bash
# 覆盖 cmd/internal/pkg 业务代码，排除 4 处 ent 生成目录与 tests/ 集成测试，产出 coverage.out + 终端按函数摘要
go test -vet=off -coverpkg=$(go list ./cmd/... ./internal/... ./pkg/... | grep -vE '/(ent|ent_log|ent_stat)(/|$)' | paste -sd, -) -coverprofile=coverage.out ./cmd/... ./internal/... ./pkg/...
go tool cover -func=coverage.out
```

要点：

- `-vet=off`：若项目存在「i18n.T() 等函数作 printf 参数」类 vet 告警会导致 go test 失败，需关闭（按项目实际取舍）。
- `-coverpkg` 排除生成代码目录（如 ent/mocks/protobuf）：生成代码占语句块比例极高（本项目实测占 73%），计入会把总覆盖率从 27% 稀释到 5%，指标失去意义，与排除 vendor 同理。
- `go list ./cmd/... ./internal/... ./pkg/... | grep -vE '/(ent|ent_log|ent_stat)(/|$)' | paste -sd, -`：动态生成排除后的包清单，避免硬编码。ent 生成代码有 4 处（`cmd/ent`、`internal/ent`、`internal/ent_log`、`internal/ent_stat`），用整段匹配 `/(ent|ent_log|ent_stat)(/|$)` 精确排除，避免误伤 `grpc_client`/`ent_slog` 等含 "ent" 子串的正常包。
- 范围用 `./cmd/... ./internal/... ./pkg/...` 而非 `./...`：`./...` 会带上 `tests/`（无 build tag 的集成测试，需真实环境，CI 里跑会挂），显式列三个业务目录更稳。本项目实测：`./...` 共 125 包，排除 ent 后 55，再去掉 `tests/` 剩 54 个业务包。

### 第 3 步：.gitignore 增加

```
coverage.out
coverage.html
```

### 第 4 步：流水线配置（push 与 pull_request 各用独立并行流水线，不并入既有 CI）

CNB 语义：事件键（如 `'**'`）下的值是 **Pipeline 数组，同事件多个数组条目 = 多条流水线并行执行**（并入同一个 pipeline 的 stages 会变串行，不要这样做）。

`.cnb.yml`（或 workflows 文件）中新增：

```yaml
"**":
  push:
    # ...既有流水线保持不动...
    # 新增：测试覆盖率流水线（与既有流水线并行）
    - runner:
        tags: cnb:arch:amd64
      docker:
        build: .cnb/Dockerfile # 复用项目构建环境
        volumes:
          - /root/.cache/go-build:copy-on-write # Go 构建缓存
          - /go/pkg/mod:copy-on-write # Go 模块缓存
      stages:
        - name: 生成测试覆盖率
          script:
            - go mod download
            # 内联：覆盖 cmd/internal/pkg 业务代码；排除 4 处 ent 生成目录（ent 占全树语句约 73%，计入会把约 27% 稀释到约 5%）与 tests/ 集成测试
            - go test -vet=off -coverpkg=$(go list ./cmd/... ./internal/... ./pkg/... | grep -vE '/(ent|ent_log|ent_stat)(/|$)' | paste -sd, -) -coverprofile=coverage.out ./cmd/... ./internal/... ./pkg/...
            - go tool cover -func=coverage.out
        - name: 上报覆盖率（全量红线）
          type: testing:coverage
          options:
            lines: 20 # 全量红线(%)：低于此值阻断流水线。初始值=当前基线-5~7个百分点
            lang: go # golang 格式报告必须指定，否则计算有误差
            pattern: "coverage.out"
            breakIfNoCoverage: true # 报告缺失直接失败，避免测试挂了却静默跳过
            key: go # 多报告数据键：与同仓库其它报告（如 Node: key: node）区分，各自写入 meta map 互不覆盖
            name: Go Coverage # 徽章显示名，与 key 一对一对应
          exports:
            lines: COVERAGE_LINES
            diff_pct: COVERAGE_DIFF_PCT
        - name: 输出覆盖率
          script: 'echo "coverage: lines=$COVERAGE_LINES diff=$COVERAGE_DIFF_PCT"'
  pull_request:
    # ...既有 PR 流水线保持不动...
    - runner:
        tags: cnb:arch:amd64
      docker:
        build: .cnb/Dockerfile
        volumes:
          - /root/.cache/go-build:copy-on-write
          - /go/pkg/mod:copy-on-write
      stages:
        - name: 生成测试覆盖率
          script:
            - go mod download
            # 内联：覆盖 cmd/internal/pkg 业务代码；排除 4 处 ent 生成目录（ent 占全树语句约 73%，计入会把约 27% 稀释到约 5%）与 tests/ 集成测试
            - go test -vet=off -coverpkg=$(go list ./cmd/... ./internal/... ./pkg/... | grep -vE '/(ent|ent_log|ent_stat)(/|$)' | paste -sd, -) -coverprofile=coverage.out ./cmd/... ./internal/... ./pkg/...
            - go tool cover -func=coverage.out
        - name: 上报覆盖率（增量红线）
          type: testing:coverage
          options:
            diffLines: 50 # 增量红线(%)：本 PR 变更代码的覆盖率，低于此值阻断
            lines: 20 # 全量底线与 push 一致，防总量被单 PR 拉低
            lang: go
            pattern: "coverage.out"
            breakIfNoCoverage: true
            key: go # 多报告数据键：与同仓库其它报告（如 Node: key: node）区分，各自写入 meta map 互不覆盖
            name: Go Coverage # 徽章显示名，与 key 一对一对应
          exports:
            lines: COVERAGE_LINES
            diff_pct: COVERAGE_DIFF_PCT
        - name: 输出覆盖率
          script: 'echo "coverage: lines=$COVERAGE_LINES diff=$COVERAGE_DIFF_PCT"'
```

## 关键决策（实现时遵循，有冲突按项目实际调整并说明）

1. **排除生成代码**：ent/mock/protobuf 等生成目录必须排除，否则指标被稀释到失去意义。
2. **红线初始值 = 当前基线减去 5~7 个百分点**，防止新增少量未覆盖代码就误触红线；每季度上调一次（棘轮）。
3. **PR 用 diffLines 管增量**：新代码必须过半覆盖，这是长期提升覆盖率的主要驱动力。
4. **breakIfNoCoverage: true**：防止测试编译失败时覆盖率阶段静默通过造成假绿。
5. **独立并行流水线**：与既有 CI 并行，互不阻塞；不要合并进既有 pipeline 的 stages（会串行）。
6. **YAML 细节**：script 值含 `": "` 等特殊字符时整体用单引号包裹；`go mod download` 前置覆盖冷缓存场景。
7. **多份覆盖率报告用 key/name 区分**：同一仓库若同时存在多份 coverage 报告（如 Go + Node 各一条并行流水线），必须给每处 `testing:coverage` 设置不同的 `key`（如 `go`/`node`）——否则它们会写入同一 meta map 互相覆盖、只保留最后跑完的那份，徽章与数据都会残缺；`name` 为徽章显示名，与 `key` 一对一。仅一份报告的项目可省略 `key`。

## 验收标准

- 内联的 `go test` 覆盖率命令跑通，产出 coverage.out，`go tool cover -func` 有 total 且与第 1 步基线一致（排除生成代码后）。
- YAML 解析正确；push/pull_request 各新增一条独立流水线。
- 首次推送后确认：仓库「洞察 → 单元测试覆盖率」徽章生成，红线数值符合预期。
