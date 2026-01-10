// 定义字段 Schema
const PLATFORM_SCHEMA = {
  cloudflare: {
    label: "CloudFlare",
    fields: [
      {
        id: "api_key",
        label: "API Token",
        type: "password",
        placeholder: "your-api-token",
        required: true,
      },
      {
        id: "zone_id",
        label: "Zone ID",
        type: "text",
        placeholder: "your-zone-id",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  tencent_dns: {
    label: "腾讯云 DNS",
    fields: [
      {
        id: "secret_id",
        label: "Secret ID",
        type: "text",
        placeholder: "AKID...",
        required: true,
      },
      {
        id: "secret_key",
        label: "Secret Key",
        type: "password",
        placeholder: "Secret Key...",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  tencent_eo: {
    label: "腾讯云 EdgeOne",
    fields: [
      {
        id: "secret_id",
        label: "Secret ID",
        type: "text",
        placeholder: "AKID...",
        required: true,
      },
      {
        id: "secret_key",
        label: "Secret Key",
        type: "password",
        placeholder: "Secret Key...",
        required: true,
      },
      {
        id: "zone_id",
        label: "Zone ID",
        type: "text",
        placeholder: "zone-...",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  aliyun_dns: {
    label: "阿里云 DNS",
    fields: [
      {
        id: "access_key_id",
        label: "Access Key ID",
        type: "text",
        placeholder: "AK...",
        required: true,
      },
      {
        id: "access_key_secret",
        label: "Access Key Secret",
        type: "password",
        placeholder: "Secret...",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  aliyun_esa: {
    label: "阿里云 ESA",
    fields: [
      {
        id: "access_key_id",
        label: "Access Key ID",
        type: "text",
        placeholder: "AK...",
        required: true,
      },
      {
        id: "access_key_secret",
        label: "Access Key Secret",
        type: "password",
        placeholder: "Secret...",
        required: true,
      },
      {
        id: "site_id",
        label: "Site ID",
        type: "number",
        placeholder: "Site ID",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  baidu_dns: {
    label: "百度云 DNS",
    fields: [
      {
        id: "access_key_id",
        label: "Access Key ID",
        type: "text",
        placeholder: "AK...",
        required: true,
      },
      {
        id: "secret_access_key",
        label: "Secret Access Key",
        type: "password",
        placeholder: "Secret...",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  huawei_dns: {
    label: "华为云 DNS",
    fields: [
      {
        id: "access_key_id",
        label: "Access Key ID",
        type: "text",
        placeholder: "AK...",
        required: true,
      },
      {
        id: "secret_access_key",
        label: "Secret Access Key",
        type: "password",
        placeholder: "Secret...",
        required: true,
      },
      {
        id: "zone_id",
        label: "Zone ID",
        type: "text",
        placeholder: "zone-...",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
  volcengine_dns: {
    label: "火山引擎 DNS",
    fields: [
      {
        id: "access_key",
        label: "Access Key",
        type: "text",
        placeholder: "AK...",
        required: true,
      },
      {
        id: "secret_key",
        label: "Secret Key",
        type: "password",
        placeholder: "Secret...",
        required: true,
      },
      {
        id: "zone_id",
        label: "Zone ID",
        type: "number",
        placeholder: "Zone ID",
        required: true,
      },
      {
        id: "domain_name",
        label: "Domain Name",
        type: "text",
        placeholder: "example.com",
        required: true,
        pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
        title: "格式错误",
        group: "common",
      },
      {
        id: "email",
        label: "Email",
        type: "email",
        placeholder: "your@email.com",
        required: true,
        group: "common",
      },
    ],
  },
};

/**
 * 动态渲染字段
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {string} capability - 服务能力 (all, cdn, dns, ssl)
 * @param {Object} initialData - 编辑模式下的初始数据 (可选)
 */
function renderPlatformFields(platform, capability, initialData = {}) {
  const container = document.getElementById("dynamic_fields_container");

  // 安全检查
  if (!platform || !PLATFORM_SCHEMA[platform]) {
    container.innerHTML =
      '<div class="text-warning text-xs text-center py-4">请选择有效的平台类型</div>';
    return;
  }

  const schema = PLATFORM_SCHEMA[platform];

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
    if (field.id === "domain_name") {
      isHidden = !(
        capability === "all" ||
        capability === "cdn" ||
        capability === "dns"
      );
    }
    if (field.id === "email") {
      isHidden = !(capability === "all" || capability === "ssl");
    }

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
function buildConfigFromPlatformFields() {
  const config = {};
  document.querySelectorAll(".dynamic-input").forEach((el) => {
    // 隐藏的字段不参与组装（除非需要保存隐藏字段的值）
    if (el.closest(".ui-form-control").classList.contains("hidden")) return;

    const key = el.getAttribute("data-key");
    const type = PLATFORM_SCHEMA[
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
