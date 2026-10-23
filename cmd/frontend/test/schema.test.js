const test = require("node:test");
const assert = require("node:assert");
const vm = require("node:vm");
const fs = require("node:fs");
const path = require("node:path");

// 这些 *_schema.js 是浏览器经典脚本：顶层只声明一个 const XXX_SCHEMA 对象与若干函数，
// 不执行任何 DOM/接口调用。这里用 vm 沙箱加载，注入最小桩后把常量取出来做结构校验，
// 从而在不引入 jsdom 的前提下守护「schema 字段/选项漏配」这类静态配置回归。
const SCHEMAS = [
  { file: "cdn_schema.js", name: "CDN_SCHEMA" },
  { file: "dns_schema.js", name: "DNS_SCHEMA" },
  { file: "platform-config_schema.js", name: "PLATFORM_SCHEMA" },
  { file: "tun_schema.js", name: "TUN_SCHEMA" },
];

const JS_DIR = path.resolve(__dirname, "../../backend/static/js");

function loadSchema(file, name) {
  const filePath = path.join(JS_DIR, file);
  const src = fs.readFileSync(filePath, "utf8");
  const sandbox = {
    document: {},
    window: {},
    console: { log() {}, warn() {}, error() {} },
    API: {},
    setTimeout() {},
    clearTimeout() {},
  };
  const ctx = vm.createContext(sandbox);
  // 追加一句把常量导出到沙箱全局，便于取回（顶层 const 不会自动挂到 globalThis）
  // 用 vm.Script + 绝对路径 filename：让 V8 覆盖率按真实文件归因，c8 才能统计到这些脚本。
  new vm.Script(`${src}\n;globalThis.__SCHEMA__ = ${name};`, {
    filename: filePath,
  }).runInContext(ctx);
  return ctx.__SCHEMA__;
}

// fields 可能是「字段对象」或「字段对象数组（一行）」，统一拍平为字段列表
function flattenFields(fields) {
  const out = [];
  for (const item of fields) {
    if (Array.isArray(item)) out.push(...item);
    else out.push(item);
  }
  return out;
}

function validateSchema(schema, label) {
  assert.ok(
    schema && typeof schema === "object" && !Array.isArray(schema),
    `${label} 应为对象`,
  );
  const platforms = Object.keys(schema);
  assert.ok(platforms.length > 0, `${label} 至少应有一个平台`);

  for (const p of platforms) {
    const def = schema[p];
    assert.ok(def && typeof def === "object", `${label}.${p} 应为对象`);
    assert.strictEqual(
      typeof def.label,
      "string",
      `${label}.${p}.label 应为字符串`,
    );
    assert.ok(def.label.length > 0, `${label}.${p}.label 不应为空`);
    assert.ok(Array.isArray(def.fields), `${label}.${p}.fields 应为数组`);

    const fields = flattenFields(def.fields);
    assert.ok(fields.length > 0, `${label}.${p} 应至少有一个字段`);

    for (const f of fields) {
      assert.ok(f && typeof f === "object", `${label}.${p} 字段应为对象`);
      assert.strictEqual(
        typeof f.id,
        "string",
        `${label}.${p} 字段 id 应为字符串`,
      );
      assert.ok(f.id.length > 0, `${label}.${p} 字段 id 不应为空`);
      assert.strictEqual(
        typeof f.label,
        "string",
        `${label}.${p}.${f.id}.label 应为字符串`,
      );
      assert.strictEqual(
        typeof f.type,
        "string",
        `${label}.${p}.${f.id}.type 应为字符串`,
      );
      assert.ok(f.type.length > 0, `${label}.${p}.${f.id}.type 不应为空`);
      assert.strictEqual(
        typeof f.required,
        "boolean",
        `${label}.${p}.${f.id}.required 应为布尔`,
      );
      if (f.type === "select") {
        assert.ok(
          Array.isArray(f.options),
          `${label}.${p} 的 select 字段 ${f.id} 必须有 options 数组`,
        );
      }
    }
  }
}

for (const { file, name } of SCHEMAS) {
  test.it(`${file} 结构合法（${name}）`, () => {
    const schema = loadSchema(file, name);
    validateSchema(schema, name);
  });
}
