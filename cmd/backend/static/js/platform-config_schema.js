// 定义字段 Schema
const PLATFORM_SCHEMA = {
  cloudflare: {
    label: "CloudFlare",
    fields: [
      {
        id: "api_token",
        label: "API Token",
        type: "password",
        placeholder: "your-api-token",
        required: true,
      },
      {
        id: "zone_id",
        label: "域名ID",
        type: "text",
        placeholder: "your-zone-id",
        required: true,
      },
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
      {
        id: "domain_grade",
        label: "域名套餐（配置即可获取线路）",
        type: "select",
        placeholder: "请选择域名套餐等级",
        required: false,
        options: [
          {
            value: "D_FREE",
            label: "免费套餐",
          },
          {
            value: "D_PLUS",
            label: "个人豪华",
          },
          {
            value: "D_EXTRA",
            label: "企业 I",
          },
          {
            value: "D_EXPERT",
            label: "企业 II",
          },
          {
            value: "D_ULTRA",
            label: "企业 III",
          },
          {
            value: "DP_EXTRA",
            label: "企业基础版",
          },
          {
            value: "DP_FREE",
            label: "新免费",
          },
          {
            value: "DP_PLUS",
            label: "专业版",
          },
          {
            value: "DP_EXPERT",
            label: "企业版",
          },
          {
            value: "DP_ULTRA",
            label: "尊享版",
          },
        ],
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
        label: "域名ID",
        type: "text",
        placeholder: "zone-...",
        required: true,
      },
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
    ],
  },
  aliyun_cdn: {
    label: "阿里云 CDN",
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
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
        label: "域名ID",
        type: "number",
        placeholder: "请输入域名ID",
        required: true,
      },
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
        label: "域名ID",
        type: "text",
        placeholder: "zone-...",
        required: true,
      },
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
        },
      ],
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
        label: "域名ID",
        type: "number",
        placeholder: "Zone ID",
        required: true,
      },
      [
        {
          id: "domain_name",
          label: "域名名称",
          type: "text",
          placeholder: "example.com",
          required: true,
          pattern: "^(([a-zA-Z0-9][\\-a-zA-Z0-9]{0,62})\\.)+([a-zA-Z]{2,63})$",
          title: "格式错误",
          group: "common",
        },
        {
          id: "email",
          label: "邮箱",
          type: "email",
          placeholder: "your@email.com",
          required: true,
          group: "common",
        },
      ],
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
  const schema = PLATFORM_SCHEMA[platform];

  if (!schema) {
    container.innerHTML =
      '<div class="text-warning text-xs text-center py-4">请选择有效的平台类型</div>';
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
        html += generatePlatformFieldHtml(subField, data, capability);
      });
      html += `</div>`;
    } else {
      // 普通非分组字段
      html += generatePlatformFieldHtml(item, data, capability);
    }
  });

  container.innerHTML = html;
}

/**
 * 生成单个字段 HTML，通过 ID 自动映射特殊属性
 * @param {Object} field - 字段定义对象
 * @param {Object} data - 表单数据对象
 * @param {string} capability - 服务能力 (all, cdn, dns, ssl)
 * @returns {string} - 生成的 HTML 字符串
 */
function generatePlatformFieldHtml(field, data, capability) {
  // 1. 自动属性映射 (根据 ID 直接映射，不写死)
  let specialClass = "";
  let dataTypeAttr = "";
  let isHidden = false;

  if (field.id === "domain_name") {
    specialClass = "domain-field";
    dataTypeAttr = 'data-type="domain_name"';
    isHidden = !(
      capability === "all" ||
      capability === "cdn" ||
      capability === "dns"
    );
  } else if (field.id === "email") {
    specialClass = "email-field";
    dataTypeAttr = 'data-type="email"';
    isHidden = !(capability === "all" || capability === "ssl");
  }

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
function buildConfigFromPlatformFields(platform = "") {
  const config = {};
  const fields = PLATFORM_SCHEMA[platform]?.fields || [];

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
