// 定义字段 Schema
const CDN_SCHEMA = {
  tencent_eo: {
    label: "腾讯云 EdgeOne",
    fields: [
      [
        {
          id: "ipv6_status",
          label: "IPv6 状态",
          type: "select",
          placeholder: "请选择 IPv6 状态",
          required: true,
          options: [
            {
              value: "follow",
              label: "遵循站点",
            },
            {
              value: "on",
              label: "开启",
            },
            {
              value: "off",
              label: "关闭",
            },
          ],
        },
        {
          id: "origin_protocol",
          label: "回源协议",
          type: "select",
          placeholder: "请选择回源协议",
          required: true,
          options: [
            {
              value: "FOLLOW",
              label: "协议跟随",
            },
            {
              value: "HTTP",
              label: "HTTP协议回源",
            },
            {
              value: "HTTPS",
              label: "HTTPS协议回源",
            },
          ],
        },
      ],
      [
        {
          id: "http_port",
          label: "HTTP 端口",
          type: "number",
          placeholder: "如 80",
          required: true,
        },
        {
          id: "https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: true,
        },
      ],
    ],
  },
  tencent_cdn: {
    label: "腾讯云 CDN",
    fields: [
      [
        {
          id: "service_type",
          label: "加速类型",
          type: "select",
          placeholder: "请选择加速类型",
          required: true,
          options: [
            {
              value: "web",
              label: "网页小文件",
            },
            {
              value: "download",
              label: "下载大文件",
            },
            {
              value: "media",
              label: "音视频点播",
            },
            {
              value: "hybrid",
              label: "动静加速",
            },
            {
              value: "dynamic",
              label: "动态加速",
            },
          ],
        },
        {
          id: "area",
          label: "加速区域",
          type: "select",
          placeholder: "请选择加速区域",
          required: false,
          options: [
            {
              value: "mainland",
              label: "中国境内",
            },
            {
              value: "overseas",
              label: "中国境外",
            },
            {
              value: "global",
              label: "全球",
            },
          ],
        },
      ],
      {
        id: "origin_pull_protocol",
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
      [
        {
          id: "http_port",
          label: "HTTP 端口",
          type: "number",
          placeholder: "如 80",
          required: false,
        },
        {
          id: "https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: false,
        },
      ],
    ],
  },
  aliyun_cdn: {
    label: "阿里云 CDN",
    fields: [
      [
        {
          id: "cdn_type",
          label: "加速类型",
          type: "select",
          placeholder: "请选择加速类型",
          required: true,
          options: [
            {
              value: "web",
              label: "图片小文件",
            },
            {
              value: "download",
              label: "大文件下载",
            },
            {
              value: "video",
              label: "视音频点播",
            },
          ],
        },
        {
          id: "scope",
          label: "加速区域",
          type: "select",
          placeholder: "请选择加速区域",
          required: false,
          options: [
            {
              value: "domestic",
              label: "仅中国内地",
            },
            {
              value: "overseas",
              label: "全球（不包含中国内地）",
            },
            {
              value: "global",
              label: "全球",
            },
          ],
        },
      ],
      [
        {
          id: "http_port",
          label: "HTTP 端口",
          type: "number",
          placeholder: "如 80",
          required: false,
          value: 0,
        },
        {
          id: "https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: false,
          value: 0,
        },
      ],
    ],
  },
  huawei_cdn: {
    label: "华为云 CDN",
    fields: [
      [
        {
          id: "business_type",
          label: "加速类型",
          type: "select",
          placeholder: "请选择加速类型",
          required: true,
          options: [
            {
              value: "web",
              label: "网页加速",
            },
            {
              value: "download",
              label: "文件下载加速",
            },
            {
              value: "video",
              label: "点播加速",
            },
            {
              value: "wholeSite",
              label: "全站加速",
            },
          ],
        },
        {
          id: "service_area",
          label: "服务区域",
          type: "select",
          placeholder: "请选择服务区域",
          required: false,
          options: [
            {
              value: "mainland_china",
              label: "中国大陆",
            },
            {
              value: "outside_mainland_china",
              label: "中国大陆境外",
            },
            {
              value: "global",
              label: "全球",
            },
          ],
        },
      ],
      [
        {
          id: "http_port",
          label: "HTTP 端口",
          type: "number",
          placeholder: "如 80",
          required: false,
          value: 0,
        },
        {
          id: "https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: false,
          value: 0,
        },
      ],
      {
        id: "remark",
        label: "备注",
        type: "text",
        placeholder: "请输入备注",
        required: false,
      },
    ],
  },
};

/**
 * 动态渲染字段
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {Object} initialData - 编辑模式下的初始数据 (可选)
 */
function renderCdnFields(platform, initialData = {}) {
  const container = document.getElementById("dynamic_fields_container");
  const schema = CDN_SCHEMA[platform];

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
        html += generateCdnFieldHtml(subField, data);
      });
      html += `</div>`;
    } else {
      // 普通非分组字段
      html += generateCdnFieldHtml(item, data);
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
function generateCdnFieldHtml(field, data) {
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
                  }>${opt.label}</option>`,
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
function buildConfigFromCdnFields(platform = "") {
  const config = {};
  const fields = CDN_SCHEMA[platform]?.fields || [];

  // 处理嵌套数组，打平后查找字段定义
  const flatFields = fields.flat();

  document.querySelectorAll(".dynamic-input").forEach((el) => {
    const parent = el.closest(".ui-form-control");
    if (parent && parent.classList.contains("hidden")) return;

    const key = el.getAttribute("data-key");
    const fieldSchema = flatFields.find((f) => f.id === key);
    let val = el.value.trim();

    // if (fieldSchema?.type === "number" && val !== "") {
    if (fieldSchema?.type === "number") {
      val = Number(val);
    }
    config[key] = val;
  });

  document.getElementById("config_json").value = JSON.stringify(config);
}
