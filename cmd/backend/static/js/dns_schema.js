// 定义字段 Schema
const DNS_SCHEMA = {};

/**
 * 动态渲染字段
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {Object} initialData - 编辑模式下的初始数据 (可选)
 */
function renderDnsFields(platform, initialData = {}) {
  const container = document.getElementById("dynamic_fields_container");
  const schema = DNS_SCHEMA[platform];

  if (!schema) {
    // container.innerHTML =
    //   '<div class="text-warning text-xs text-center py-4">请选择有效的平台类型</div>';
    container.innerHTML = "";
    return;
  }

  // 数据预处理
  let data = initialData;
  if (typeof initialData === "string" && initialData !== "") {
    try {
      data = JSON.parse(initialData);
    } catch (e) {
      data = {};
    }
  }

  let html = `<div class="divider text-[10px] opacity-30 uppercase tracking-[0.2em] mb-6">${schema.label} 参数配置</div>`;

  // 遍历 fields 定义
  schema.fields.forEach((item) => {
    // 检查是否是嵌套数组（即：分组容器）
    if (Array.isArray(item)) {
      // 直接使用 ID 拼接类名，例如：domain_name-or-email-group
      const groupClassName = item.map((f) => f.id).join("-or-") + "-group";

      html += `<div class="form-group ${groupClassName}" data-platform="${platform}">`;
      item.forEach((subField) => {
        html += generateDnsFieldHtml(subField, data);
      });
      html += `</div>`;
    } else {
      // 普通非分组字段
      html += generateDnsFieldHtml(item, data);
    }
  });

  container.innerHTML = html;
}

/**
 * 生成单个字段 HTML，通过 ID 自动映射特殊属性
 * @param {Object} field - 字段定义对象
 * @param {Object} data - 表单数据对象
 * @returns {string} - 生成的 HTML 字符串
 */
function generateDnsFieldHtml(field, data) {
  // 1. 自动属性映射 (根据 ID 直接映射，不写死)
  let specialClass = "";
  let dataTypeAttr = "";
  let isHidden = false;

  const value = data[field.id] !== undefined ? data[field.id] : "";
  const requiredAttr = field.required && !isHidden ? "required" : "";

  let inputHtml = "";
  if (field.type === "password") {
    inputHtml = `
      <div class="flex items-center gap-2">
        <input type="password" class="ui-input dynamic-input" data-key="${field.id}" value="${value}" placeholder="${field.placeholder}" ${requiredAttr} autocomplete="new-password" />
        <button type="button" class="btn btn-square btn-sm btn-ghost" onclick="togglePasswordVisibilityByPrev(this)">
          <span class="i-lucide-eye w-4 h-4"></span>
        </button>
      </div>`;
  } else if (field.type === "select") {
    inputHtml = `
          <select class="ui-select dynamic-input" data-key="${
            field.id
          }" ${requiredAttr}>
            ${(field.options || [])
              .map(
                (opt) =>
                  `<option value="${opt.value}" ${
                    opt.value == value ? "selected" : ""
                  }>${opt.label}</option>`
              )
              .join("")}
          </select>`;
  } else {
    inputHtml = `<input type="${
      field.type
    }" class="ui-input dynamic-input ${specialClass}" data-key="${
      field.id
    }" value="${value}" 
                  placeholder="${field.placeholder}" ${
      field.pattern ? `pattern="${field.pattern}"` : ""
    } 
                  ${
                    field.title ? `title="${field.title}"` : ""
                  } ${requiredAttr} />`;
  }

  return `
    <div class="ui-form-control ${isHidden ? "hidden" : ""}" ${dataTypeAttr}>
      <label class="ui-label-text">${field.label} ${
    field.required ? '<span class="text-error">*</span>' : ""
  }</label>
      ${inputHtml}
    </div>`;
}

/**
 * 自动从动态字段中组装 JSON
 * @param {string} platform - 平台 key (如 cloudflare)
 */
function buildConfigFromDnsFields(platform = "") {
  const config = {};
  const fields = DNS_SCHEMA[platform]?.fields || [];

  // 处理嵌套数组，打平后查找字段定义
  const flatFields = fields.flat();

  document.querySelectorAll(".dynamic-input").forEach((el) => {
    const parent = el.closest(".ui-form-control");
    if (parent && parent.classList.contains("hidden")) return;

    const key = el.getAttribute("data-key");
    const fieldSchema = flatFields.find((f) => f.id === key);
    let val = el.value.trim();

    if (fieldSchema?.type === "number" && val !== "") {
      val = Number(val);
    }
    config[key] = val;
  });

  document.getElementById("config_json").value = JSON.stringify(config);
}
