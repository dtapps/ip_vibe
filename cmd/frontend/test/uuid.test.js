const test = require("node:test");
const assert = require("node:assert");
const vm = require("node:vm");
const fs = require("node:fs");
const path = require("node:path");
// 直接 require 源文件：本项目前端为 CJS/UMD 经典脚本（无 type:module），
// c8 通过 require hook 原生支持 CJS，可正确统计行覆盖率。
const uuid = require("../../backend/static/js/uuid.js");

const UUID_SRC = path.resolve(__dirname, "../../backend/static/js/uuid.js");

// UUID v4：版本位=4，变体位=RFC4122(8/9/a/b)，位于第 4 段首 nibble（标准布局）
const UUID_V4_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
// UUID v7：版本位=7，位于第 3 段首 nibble。
// 注意：uuid.js 的 v7 实现非标准——变体位被放在第 3 段第 3 个 nibble（byte[7] 高 2 位），
// 第 4 段（bytes[8..9]）是 14-bit 计数器而非标准 rand_b；标准 RFC 9562 变体位应在 byte[8]。
const UUID_V7_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f][89ab][0-9a-f]-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// 去掉连字符得到 32 位十六进制串；byteAt 取第 i 个字节（0 起）
const hexOf = (s) => s.replace(/-/g, "");
const byteAt = (hex, i) => parseInt(hex.substr(i * 2, 2), 16);

test.describe("uuid", () => {
  test.it("v4 符合格式且版本位为 4", () => {
    assert.match(uuid.v4(), UUID_V4_RE);
  });

  test.it("v7 符合格式且版本位为 7", () => {
    assert.match(uuid.v7(), UUID_V7_RE);
  });

  test.it("v4 字节级：version=4、variant 高位为 10（RFC4122）", () => {
    for (let i = 0; i < 100; i++) {
      const h = hexOf(uuid.v4());
      assert.strictEqual(byteAt(h, 6) >> 4, 0x4, "version nibble 应为 4");
      assert.strictEqual(byteAt(h, 8) & 0xc0, 0x80, "variant 高 2 位应为 10");
    }
  });

  test.it("v7 字节级：version=7、variant 位于 byte[7]", () => {
    const h = hexOf(uuid.v7());
    assert.strictEqual(byteAt(h, 6) >> 4, 0x7, "version nibble 应为 7");
    // 本实现变体在 byte[7]（非标准 RFC9562 的 byte[8]）
    assert.strictEqual(byteAt(h, 7) & 0xc0, 0x80, "variant 高 2 位应为 10");
    // 已知缺陷（不在本测试断言）：bytes[0..5] 的时间戳用 `ts >>> 40` 等 32-bit 位移编码，
    // JS 位运算会先把 ts 截断为 uint32，导致 48-bit 毫秒时间戳的高位丢失、低两字节重复，
    // 无法还原真实 Date.now()。详见 docs/coverage-node.md「已知问题：v7 时间戳编码」。
  });

  test.it("v7 同毫秒内 counter 从 0 起递增（bytes[8..9]）", () => {
    const realNow = Date.now;
    // 固定到一个「大于此前所有真实时间戳」的毫秒值，确保首次进入即重置 counter=0，
    // 且远小于 2^48（48-bit 时间戳上限），不会溢出编码。
    const base = realNow() + 1_000_000;
    Date.now = () => base;
    try {
      const c0 = parseInt(hexOf(uuid.v7()).slice(16, 20), 16);
      const c1 = parseInt(hexOf(uuid.v7()).slice(16, 20), 16);
      const c2 = parseInt(hexOf(uuid.v7()).slice(16, 20), 16);
      assert.strictEqual(c0, 0, "首条 counter 应为 0");
      assert.strictEqual(c1, 1, "次条 counter 应为 1");
      assert.strictEqual(c2, 2, "第三条 counter 应为 2");
    } finally {
      Date.now = realNow;
    }
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

// loadInSandbox 在指定宿主的沙箱里重新加载 uuid.js：
// 该文件是 UMD 风格，会按 crypto / msCrypto / module 的存在与否走不同分支，
// 换宿主分别加载才能覆盖这些兼容分支（沙箱内没有 module，因此 uuid 挂在沙箱全局上）。
function loadInSandbox(extraGlobals) {
  const src = fs.readFileSync(UUID_SRC, "utf8");
  const ctx = vm.createContext({ ...extraGlobals });
  new vm.Script(src, { filename: UUID_SRC }).runInContext(ctx);
  return ctx;
}

test.describe("uuid 宿主兼容分支", () => {
  test.it("既无 crypto 也无 msCrypto 时降级到 Math.random", () => {
    const ctx = loadInSandbox({});

    assert.strictEqual(typeof ctx.crypto, "undefined");
    assert.strictEqual(typeof ctx.msCrypto, "undefined");
    assert.match(ctx.uuid.v4(), UUID_V4_RE, "降级实现仍应产出合法 v4");
  });

  test.it("存在 msCrypto 时优先使用其 getRandomValues（IE11 兼容）", () => {
    let calls = 0;
    const ctx = loadInSandbox({
      msCrypto: {
        getRandomValues(bytes) {
          calls += 1;
          for (let i = 0; i < bytes.length; i++) bytes[i] = i;
          return bytes;
        },
      },
    });

    assert.match(ctx.uuid.v4(), UUID_V4_RE);
    assert.ok(calls > 0, "应调用 msCrypto.getRandomValues");
  });

  // 注：文件末尾 `else if (typeof window !== "undefined") { window.uuid = uuid }`
  // 实际不可达——IIFE 的形参就叫 `global`，永远满足前一个分支，uuid 会挂到调用方全局对象上。
  // 这里只断言「无 module 时挂到宿主全局」，不假设具体挂在 window 上。
  test.it("无 module 时挂到宿主全局对象", () => {
    const ctx = loadInSandbox({ window: {} });

    assert.strictEqual(typeof ctx.module, "undefined");
    assert.strictEqual(typeof ctx.uuid.v4, "function");
    assert.match(ctx.uuid.v7(), UUID_V7_RE);
  });
});
