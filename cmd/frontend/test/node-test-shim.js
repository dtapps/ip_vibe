// 把 node:test 风格的 test.describe / test.it 桥接到 Vitest 的用例注册函数。
// 现有 *.test.js 是 CJS 经典脚本（require），Vitest 无法直接收集 node:test 的用例；
// 各测试文件只把 require("node:test") 换成本文件，断言与 vm 加载逻辑完全不动。
// Vitest 下 globals: true 会把 describe / it 挂到 globalThis；
// 直接用 node 跑聚合入口时回退到 node:test，保持原来的运行方式可用。
const nodeTest = require("node:test");

module.exports = {
  describe: (...args) => (globalThis.describe || nodeTest.describe)(...args),
  it: (...args) => (globalThis.it || nodeTest.it)(...args),
};
