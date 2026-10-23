# 前端单元测试覆盖率实现文案（本项目实际方案）

> 本项目前端是**原生 JS（jQuery/daisyUI/Tailwind）**，不是 Vite/Vue/TS。本文按本项目**真实栈**记录覆盖率实现，可直接作为本项目的实现说明/复现依据。

# 任务：为前端实现单元测试覆盖率上报与红线门禁（CNB 云原生构建）

## 背景与目标

- 构建平台是 CNB（cnb.cool），流水线配置在仓库根 `.cnb.yml` include 的 `.cnb/workflows/*.yml`。
- 使用 CNB 内置任务 `type: testing:coverage` 上报覆盖率：push 事件全量红线（`lines`），pull_request 事件额外计算增量红线（`diffLines`，仅 PR 类事件生效）。
- 红线采用「棘轮策略」：初始值贴当前基线下沿留余量，防止误伤，后续随覆盖率提升逐步上调。

## 本项目前端的真实形态（决定工具选型）

- `cmd/frontend/` 只是**资产构建管道**：把 jquery/echarts/dayjs 等复制到 `../backend/static/libs/`，再用 Tailwind CLI 编译 CSS。**没有应用 JS 源码、没有 Vite、没有 src/ 目录**。
- 真正的应用 JS 在 `cmd/backend/static/js/`（约 33 个文件），是 **CJS/UMD 经典脚本**（`module.exports` 或全局函数声明），通过 `<script src>`（**非 `type=module`**）加载。
- 因此**不能把源码改成 ESM**：一加 `export`，浏览器按经典脚本加载会直接语法错误。

## 工具选型：Node 内置 test runner + c8（不用 vitest）

这是本项目实测后的关键决策：

- vitest 底层是 vite-node，对 **CJS 模块走原生 `require`** 加载，绕过了插桩管线 → 覆盖率**恒为 0%**（哪怕测试执行了上千次）。
- 对照实验（同一套配置）：ESM 探针文件 **100%**，CJS 的 `uuid.js` **0%**；随后换 v8/istanbul provider、把 `root` 抬到 `cmd`/仓库根、`singleThread`、`deps.inline` 全部无效——证明问题在 **CJS 格式本身**，不是配置。
- `c8` 通过 Node 的 **require hook（V8 覆盖率）**工作，**原生支持 CJS**：只要模块被 `require` 就被插桩统计，能正确反映行覆盖。

## 实现步骤

### 第 1 步：安装 c8

```bash
pnpm --dir cmd/frontend add -D c8
```

### 第 2 步：编写测试（node:test + require 源文件）

测试放 `cmd/frontend/test/*.test.js`，直接 `require` 源码：

```js
const test = require("node:test");
const assert = require("node:assert");
// 相对「测试文件位置」解析；cmd/frontend/test → ../../ 到仓库根 → 再进 backend
const uuid = require("../../backend/static/js/uuid.js");

test.describe("uuid", () => {
  test.it("v4 符合格式且版本位为 4", () => {
    assert.match(uuid.v4(), /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i);
  });
  test.it("v4 批量生成不重复", () => {
    const set = new Set();
    for (let i = 0; i < 1000; i++) set.add(uuid.v4());
    assert.strictEqual(set.size, 1000);
  });
});
```

优先覆盖**纯逻辑/UMD 模块**（如 `uuid.js`，不依赖 DOM）；顶层直接操作 `document` 的全局脚本需 jsdom，作为第二步。

当前三类测试的加载方式（都不引 jsdom / jQuery）：

| 测试文件             | 被测对象                            | 加载方式                                                                                   |
|----------------------|-------------------------------------|--------------------------------------------------------------------------------------------|
| `uuid.test.js`       | `uuid.js`（CJS）                    | 直接 `require`；另用 `vm` 沙箱造「无 crypto」「有 msCrypto」宿主，覆盖兼容分支             |
| `schema.test.js`     | 4 个 `*_schema.js`                  | `vm` 沙箱注入 `document/window/console/API/timer` 桩，取顶层 `const XXX_SCHEMA` 做结构校验 |
| `components.test.js` | `components/{pagination,switch}.js` | `vm` 沙箱注入 `localStorage/document.cookie/location/URL/API/alert` 假实现，直接驱动函数   |

关键技巧：脚本用 `new vm.Script(src, { filename: 绝对路径 })` 编译，V8 覆盖率才能按真实文件归因；若用 `vm.runInContext(字符串)` 则这些文件会统计成 0%。`*_schema.js` 是数据文件，只能覆盖到各平台定义行（函数体未执行，故行覆盖 25%~76% 属正常）。

### 第 3 步：package.json 脚本

测试文件拆多个（`uuid.test.js` / `schema.test.js` / `components.test.js`），但**入口只有一个** `test/index.test.js`，由它按序 `require` 各文件，保证都在主线程注册：

```js
// cmd/frontend/test/index.test.js
require("./uuid.test.js");
require("./schema.test.js");
require("./components.test.js");
```

```json
"scripts": {
  "test": "node test/index.test.js",
  "test:coverage": "cd ../.. && c8 --reporter=lcov --reporter=text --all --include='cmd/backend/static/js/**/*.js' --reports-dir=cmd/frontend/coverage node cmd/frontend/test/index.test.js"
}
```

三个实测踩过的易错点：

1. **require 路径要 `../../backend/...`**：`require` 相对测试文件位置解析，`cmd/frontend/test` 到 `cmd/backend/static/js` 需上跳两级（不是一级 `../`）。
2. **c8 必须在仓库根跑（`cd ../..`）**：c8 的 `--include` glob 不允许 `../` 逃逸 cwd；到仓库根后用 `cmd/backend/static/js/**/*.js`。
3. **直接 `node <testfile>`，不要 `node --test`**：`node --test` 用 worker 线程跑测试文件，worker 里的 CJS 同样不被 c8 插桩（与 vitest 的 worker 问题同源）。

产出 `cmd/frontend/coverage/lcov.info`。

### 第 4 步：.gitignore 增加

```
cmd/frontend/coverage
```

### 第 5 步：流水线配置（push 与 pull_request 各用独立并行流水线）

CNB 语义：事件键下的值是 Pipeline 数组，同事件多个条目 = 多流水线并行；并入同一 pipeline 的 stages 会串行，不要这样做。

```yaml
"**":
  push:
    - runner:
        tags: cnb:arch:amd64
      docker:
        build: .cnb/Dockerfile # 已预装 Node + pnpm
        volumes:
          - /root/.local/share/pnpm/store:copy-on-write
      stages:
        - name: 安装前端依赖
          script: pnpm --dir ./cmd/frontend install --frozen-lockfile
        - name: 生成前端测试覆盖率
          script: pnpm --dir ./cmd/frontend test:coverage # 产出 cmd/frontend/coverage/lcov.info
        - name: 上报覆盖率（全量红线）
          type: testing:coverage
          options:
            lines: 1 # 全量红线(%)：当前基线≈1%（仅 uuid 已测），随测试增多逐步上调
            pattern: "cmd/frontend/coverage/lcov.info"
            breakIfNoCoverage: true # 报告缺失直接失败，避免测试挂了却静默跳过
            key: node # 多报告数据键：与 Go 侧（key: go）区分，各自写入 meta map 互不覆盖
            name: Node Coverage # 徽章显示名，与 key 一对一对应
          exports:
            lines: NODE_COVERAGE_LINES
        - name: 输出覆盖率
          script: 'echo "node coverage: lines=$NODE_COVERAGE_LINES"'
  pull_request:
    # 与 push 相同，上报阶段 options 增加：diffLines: 50（本 PR 变更代码覆盖率红线）
```

## 关键决策（实现时遵循）

1. **c8 而非 vitest**：CJS 前端下 vitest 覆盖率恒 0（原因见上），c8 原生支持 CJS。
2. **直接 `node <file>` 不用 `node --test`**：worker 线程不插桩 CJS。
3. **c8 从仓库根跑**：`--include` glob 不能 `../` 逃逸 cwd。
4. **`--all` 全量度量**：未测文件按 0% 计入分母，指标不虚高（当前全量≈1%，仅 uuid 已测）；真正驱动质量的是 PR 增量红线。
5. **不设 `lang`**：lcov 由插件自动识别；`lang: go` 仅 golang 格式报告需要。
6. **红线棘轮**：`lines` 初始贴基线下沿；`diffLines: 50`（变更代码过半覆盖）长期不动或逐步上调。
7. **多份覆盖率报告用 key/name 区分**：同一仓库 Go + Node 各一条并行流水线，必须设不同 `key`（`go`/`node`），否则会写入同一 meta map 互相覆盖、只保留最后跑完那份；`name` 为徽章显示名，与 `key` 一对一。仅一份报告的项目可省略 `key`。
8. **测试从纯函数/UMD 模块开始**：DOM 组件需 jsdom/@vue/test-utils 之类，作为第二步。

## 已知问题

1. `cmd/backend/static/js/uuid.js` 的 **v7 实现不符合 RFC 9562**：变体位（variant）被放在 `byte[7]`（第 3 段第 3 个 nibble）而非 `byte[8]`，第 4 段（bytes[8..9]）是 14-bit 计数器而非标准 `rand_b`；v4 为标准布局。测试已按实际实现校验；若需对外提供合规 UUIDv7，需修正源码。
2. **v7 时间戳编码错误**：`bytes[0..5]` 用 `ts >>> 40`、`ts >>> 32` 等位移写 48 位毫秒时间戳，但 JS 位运算会先把操作数转成 **uint32**（`ts >>> 40` 实际等价 `ts >>> 8`），导致高字节丢失、`bytes[0]==bytes[4]`、`bytes[1]==bytes[5]`，无法从 UUID 还原时间、也失去「按时间有序」的意义。正确写法是用除法取字节，如 `Math.floor(ts / 2 ** 40) & 0xff`。当前测试只断言版本位/变体位/计数器，未断言时间戳，避免把缺陷固化成契约。
3. **导出分支 `else if (typeof window !== "undefined") { window.uuid = ... }` 不可达**：IIFE 的形参就叫 `global`（实参是 `globalThis`），永远满足前一个分支。测试按「无 `module` 时挂到宿主全局对象」断言，不假设挂在 `window` 上。

## 验收标准

- `pnpm test` 全绿（当前 21 例）；`pnpm test:coverage` 产出 `cmd/frontend/coverage/lcov.info`：`uuid.js` 行覆盖≈98%（仅剩不可达的 window 导出分支）、`js/components/*` 100%、`*_schema.js` 25%~76%（仅数据定义行）。
- YAML 解析正确；push/pull_request 各一条独立并行流水线；徽章显示 "Node Coverage"。
