const test = require("node:test");
const assert = require("node:assert");
const vm = require("node:vm");
const fs = require("node:fs");
const path = require("node:path");

// components/*.js 也是浏览器经典脚本：函数挂在沙箱全局 / window.components 上，
// 依赖 localStorage、document.cookie、location、API 等宿主对象。
// 这里注入最小假实现，直接驱动函数逻辑，覆盖「Cookie 与跳转参数拼装」「失败回滚」等易错点。
const JS_DIR = path.resolve(__dirname, "../../backend/static/js");

// loadScript 在独立沙箱里加载脚本，返回 { ctx, sandbox }。
function loadScript(file, extra) {
  const filePath = path.join(JS_DIR, file);
  const src = fs.readFileSync(filePath, "utf8");
  const sandbox = {
    URL,
    console: { log() {}, warn() {}, error() {} },
    ...extra,
  };
  sandbox.window = sandbox;
  const ctx = vm.createContext(sandbox);
  new vm.Script(src, { filename: filePath }).runInContext(ctx);
  return { ctx, sandbox };
}

test.describe("components/pagination", () => {
  function setup(initialHref) {
    const store = new Map();
    const sandbox = {
      localStorage: { setItem: (k, v) => store.set(k, v) },
      document: { cookie: "" },
      location: { href: initialHref },
    };
    const { ctx } = loadScript("components/pagination.js", sandbox);
    // ctx 即沙箱全局对象（脚本里写入的 window.components 直接挂在它上面）
    return { ctx, sandbox: ctx, store };
  }

  test.it("写入 localStorage / Cookie 并跳转到带分页参数的 URL", () => {
    const { ctx, sandbox, store } = setup(
      "http://127.0.0.1:8080/list?page=1&page_size=10&q=abc",
    );

    ctx.changePage(3, 50);

    assert.strictEqual(
      store.get("pref_page_size"),
      50,
      "分页大小应持久化到 localStorage",
    );
    assert.strictEqual(
      sandbox.document.cookie,
      "pref_page_size=50; path=/; max-age=2592000",
      "Cookie 应带 path 与 30 天有效期",
    );

    const url = new URL(sandbox.location.href);
    assert.strictEqual(url.searchParams.get("page"), "3");
    assert.strictEqual(url.searchParams.get("page_size"), "50");
    assert.strictEqual(url.searchParams.get("q"), "abc", "其它查询参数应保留");
  });

  test.it("挂载到 window.components.pagination，且不覆盖已有对象", () => {
    const { sandbox } = setup("http://127.0.0.1:8080/list");
    assert.strictEqual(
      typeof sandbox.window.components.pagination.changePage,
      "function",
    );

    const existing = { changePage: () => "original" };
    const { sandbox: again } = loadScript("components/pagination.js", {
      localStorage: { setItem() {} },
      document: { cookie: "" },
      location: { href: "http://127.0.0.1:8080/list" },
      components: { pagination: existing },
    });
    assert.strictEqual(
      again.window.components.pagination,
      existing,
      "已存在时不应覆盖",
    );
  });
});

test.describe("components/switch", () => {
  // setup 组装一次 switch 场景，返回沙箱与各桩的调用记录。
  function setup({ api, checked = false, isNoReload = "true" }) {
    const calls = { patch: [], alerts: [], errors: [], reload: 0 };
    const element = { checked };

    const sandbox = {
      API:
        api === undefined
          ? undefined
          : {
              patch: (url, payload) => {
                calls.patch.push({ url, payload });
                return api.patch(url, payload);
              },
            },
      location: { reload: () => (calls.reload += 1) },
      alert: (msg) => calls.alerts.push(String(msg)),
    };
    sandbox.console = {
      log() {},
      warn() {},
      error: (...args) => calls.errors.push(args.join(" ")),
    };

    const { ctx } = loadScript("components/switch.js", sandbox);
    return { ctx, element, calls, isNoReload };
  }

  test.it("成功 PATCH 且 isNoReload!=false 时刷新页面", async () => {
    const { ctx, element, calls, isNoReload } = setup({
      api: { patch: async () => ({}) },
    });

    element.checked = true;
    await ctx.changeStatus(
      "/api/v1/example/1/status",
      "enabled",
      isNoReload,
      element,
    );

    // 载荷对象来自 vm 沙箱（原型链不同），用 JSON 归一后再比较
    assert.deepStrictEqual(
      JSON.parse(JSON.stringify(calls.patch)),
      [{ url: "/api/v1/example/1/status", payload: { enabled: true } }],
      "应以 dataKey 组装 PATCH 载荷",
    );
    assert.strictEqual(calls.reload, 1, "非 \u201cfalse\u201d 时应刷新页面");
    assert.strictEqual(calls.alerts.length, 0);
  });

  test.it("isNoReload 为字符串 false 时不刷新页面", async () => {
    const { ctx, element, calls } = setup({
      api: { patch: async () => ({}) },
      isNoReload: "false",
    });

    await ctx.changeStatus(
      "/api/v1/example/1/status",
      "enabled",
      "false",
      element,
    );
    assert.strictEqual(calls.reload, 0, "isNoReload=false 时不应刷新页面");
  });

  test.it("API.patch 未定义时提示且不刷新", async () => {
    const { ctx, element, calls } = setup({ api: undefined });

    await ctx.changeStatus(
      "/api/v1/example/1/status",
      "enabled",
      "true",
      element,
    );
    assert.strictEqual(calls.patch.length, 0);
    assert.strictEqual(calls.reload, 0);
    assert.match(calls.alerts[0], /API\.patch 函数未定义/);
  });

  test.it("请求失败时提示错误并回滚 checkbox 状态", async () => {
    const { ctx, element, calls } = setup({
      api: {
        patch: async () => {
          throw new Error("boom");
        },
      },
    });

    element.checked = true;
    await ctx.changeStatus(
      "/api/v1/example/1/status",
      "enabled",
      "true",
      element,
    );

    assert.strictEqual(
      element.checked,
      false,
      "失败后应把 checkbox 恢复为原状态",
    );
    assert.match(calls.alerts[0], /状态更新失败: boom/);
    assert.strictEqual(calls.errors.length, 1, "应记录一次 console.error");
    assert.strictEqual(calls.reload, 0);
  });
});
