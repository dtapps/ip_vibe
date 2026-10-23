// 统一测试入口：按序 require 各测试文件，全部在主线程注册由 node:test 执行。
// 不能用 `node --test`（多文件会被塞进 worker 线程，CJS 源不会被 c8 插桩 → 覆盖率归零），
// 因此以单个 node 进程 require 聚合，保证 c8 能对 require 的 CJS 源码正确插桩。
require("./uuid.test.js");
require("./schema.test.js");
require("./components.test.js");
