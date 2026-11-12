import { configDefaults, defineConfig } from 'vitest/config'

// 扫描 / 排除清单由 Makefile「测试覆盖率」节通过环境变量注入（唯一来源），
// 新增目录只改 Makefile；不经过 make 直接跑 vitest 时回退到默认值，行为不变。
// 多个 glob 用空格或逗号分隔。
const list = (value: string | undefined, fallback: string[]) =>
  value ? value.split(/[,\s]+/).filter(Boolean) : fallback

// 配置说明见 docs/coverage-node.md。三个关键点：
// 1. 本项目无 vite.config（纯 HTML/CSS/JS + htmx），故不需要 mergeConfig。
// 2. 被测代码是「浏览器全局脚本」（挂 window/document/htmx，无模块导出），
//    因此 environment 必须用 happy-dom 而非 node。
// 3. 被测源码在前端包之外（cmd/backend/static/js）。覆盖率 glob 逃不出 vitest 的
//    root（实测 include: ['../backend/...'] 会匹配到 0 个文件），
//    因此把 root 上移到 cmd/，让测试与源码都落在根内。
//    注意：测试文件里的相对 import 以「文件自身」为基准，不受 root 影响。
export default defineConfig({
  root: '..',
  test: {
    environment: 'happy-dom',
    globals: true,
    include: list(process.env.VITEST_TEST_INCLUDE, ['frontend/test/**/*.test.js']),
    // index.test.js 是 node:test 的聚合入口（require 其余 3 个文件）；
    // Vitest 会自动发现所有 *.test.js，再加载它会重复注册，排除掉。
    exclude: [...configDefaults.exclude, '**/index.test.js'],
    coverage: {
      provider: 'v8',
      reporter: ['lcov', 'text', 'json-summary'],
      reportsDirectory: './frontend/coverage',
      include: list(process.env.VITEST_COVER_INCLUDE, ['backend/static/js/**/*.js']),
      exclude: list(process.env.VITEST_COVER_EXCLUDE, ['backend/static/libs/**']),
      // 全量度量：include 显式列出源码目录，未被测试 import 的文件也按 0% 计入分母
      //（Vitest 5 已移除 coverage.all，全量由 coverage.include 实现）
    },
  },
})
