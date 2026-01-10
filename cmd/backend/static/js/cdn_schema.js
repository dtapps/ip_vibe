// 定义字段 Schema
const CDN_SCHEMA = {
  tencent_eo: {
    label: "腾讯云 EdgeOne",
    fields: [
      {
        id: "origin_protocol",
        label: "回源协议",
        type: "select",
        placeholder: "请选择回源协议",
        required: true,
        options: [
          {
            value: "follow",
            label: "协议跟随",
          },
          {
            value: "http",
            label: "HTTP协议回源",
          },
          {
            value: "https",
            label: "HTTPS协议回源",
          },
        ],
      },
      {
        id: "http_port",
        label: "HTTP 端口",
        type: "number",
        placeholder: "如 80",
        required: true,
        group: "common",
      },
      {
        id: "https_port",
        label: "HTTPS 端口",
        type: "number",
        placeholder: "如 443",
        required: true,
        group: "common",
      },
    ],
  },
};

/**
 * 动态渲染字段
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {Object} initialData - 编辑模式下的初始数据 (可选)
 */
function renderCDNFields(platform = "", initialData = {}) {
  const container = document.getElementById("dynamic_fields_container");

  // 安全检查
  if (!platform || !CDN_SCHEMA[platform]) {
    container.innerHTML =
      '<div class="text-warning text-xs text-center py-4">该平台暂不支持自定义配置</div>';
    return;
  }

  const schema = CDN_SCHEMA[platform];

  // 确保数据是对象（处理后端传回 JSON 字符串的情况）
  let data = initialData;
  if (typeof initialData === "string" && initialData !== "") {
    try {
      data = JSON.parse(initialData);
    } catch (e) {
      data = {};
    }
  }

  let html = `<div class="divider text-[10px] opacity-30 uppercase tracking-[0.2em] mb-6">${schema.label} 参数配置</div>`;

  schema.fields.forEach((field) => {
    // 处理显隐逻辑
    let isHidden = false;

    const value = data[field.id] !== undefined ? data[field.id] : "";
    const requiredAttr = field.required && !isHidden ? "required" : "";

    let inputHtml = "";
    switch (field.type) {
      case "password":
        inputHtml = `
          <div class="flex items-center gap-2">
            <input type="password" class="ui-input dynamic-input" data-key="${field.id}" value="${value}" placeholder="${field.placeholder}" ${requiredAttr} autocomplete="new-password" />
            <button type="button" class="btn btn-square btn-sm btn-ghost" onclick="togglePasswordVisibilityByPrev(this)">
              <span class="i-lucide-eye w-4 h-4"></span>
            </button>
          </div>`;
        break;
      case "select":
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
        break;
      default:
        inputHtml = `
          <input type="${
            field.type
          }" class="ui-input dynamic-input" data-key="${
          field.id
        }" value="${value}" 
                 placeholder="${field.placeholder}" ${
          field.pattern ? `pattern="${field.pattern}"` : ""
        } 
                 ${
                   field.title ? `title="${field.title}"` : ""
                 } ${requiredAttr} />`;
    }

    html += `
      <div class="ui-form-control ${isHidden ? "hidden" : ""}" data-field-id="${
      field.id
    }">
        <label class="ui-label-text">${field.label} ${
      field.required ? '<span class="text-error">*</span>' : ""
    }</label>
        ${inputHtml}
      </div>`;
  });

  container.innerHTML = html;
}

/**
 * 自动从动态字段中组装 JSON
 */
function buildConfigFromCDNFields() {
  const config = {};
  document.querySelectorAll(".dynamic-input").forEach((el) => {
    // 隐藏的字段不参与组装（除非需要保存隐藏字段的值）
    if (el.closest(".ui-form-control").classList.contains("hidden")) return;

    const key = el.getAttribute("data-key");
    const type = CDN_SCHEMA[
      document.getElementById("platform").value
    ]?.fields.find((f) => f.id === key)?.type;

    let val = el.value.trim();
    // 如果是数字类型，转为 Number
    if (type === "number" && val !== "") {
      val = Number(val);
    }
    config[key] = val;
  });
  document.getElementById("config_json").value = JSON.stringify(config);
}
