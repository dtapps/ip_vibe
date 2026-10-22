const test = require("node:test");
const assert = require("node:assert");
// 直接 require 源文件：本项目前端为 CJS/UMD 经典脚本（无 type:module），
// c8 通过 require hook 原生支持 CJS，可正确统计行覆盖率。
const uuid = require("../../backend/static/js/uuid.js");

// UUID v4：版本位=4，变体位=RFC4122(8/9/a/b)，位于第 4 段首 nibble（标准布局）
const UUID_V4_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
// UUID v7：版本位=7，位于第 3 段首 nibble。
// 注意：uuid.js 的 v7 实现非标准——变体位被放在第 3 段第 3 个 nibble（byte[7] 高 2 位），
// 第 4 段（bytes[8..9]）是 14-bit 计数器而非标准 rand_b；标准 RFC 9562 变体位应在 byte[8]。
const UUID_V7_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f][89ab][0-9a-f]-[0-9a-f]{4}-[0-9a-f]{12}$/i;

test.describe("uuid", () => {
  test.it("v4 符合格式且版本位为 4", () => {
    assert.match(uuid.v4(), UUID_V4_RE);
  });

  test.it("v7 符合格式且版本位为 7", () => {
    assert.match(uuid.v7(), UUID_V7_RE);
  });

  test.it("v4 批量生成不重复", () => {
    const set = new Set();
    for (let i = 0; i < 1000; i++) set.add(uuid.v4());
    assert.strictEqual(set.size, 1000);
  });

  test.it("v7 批量生成不重复（同毫秒内走 14-bit 计数器）", () => {
    const set = new Set();
    for (let i = 0; i < 1000; i++) set.add(uuid.v7());
    assert.strictEqual(set.size, 1000);
  });

  test.it("导出 v4 / v7 两个方法", () => {
    assert.strictEqual(typeof uuid.v4, "function");
    assert.strictEqual(typeof uuid.v7, "function");
  });
});
