// 定义字段 Schema
const CDN_SCHEMA = {
  tencent_eo: {
    label: "腾讯云 EdgeOne",
    fields: [
      {
        id: "zone",
        label: "站点",
        type: "select",
        placeholder: "请选择站点",
        required: true,
        options: [],
      },
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
  aliyun_esa: {
    label: "阿里云 ESA",
    fields: [
      {
        id: "site",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "biz_name",
          label: "业务场景",
          type: "select",
          placeholder: "请输入业务场景",
          required: true,
          options: [
            {
              value: "web",
              label: "web 网页",
            },
            {
              value: "api",
              label: "api 接口",
            },
            {
              value: "video_image",
              label: "视频图片",
            },
          ],
        },
        {
          id: "origin_scheme",
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
      ],
      [
        {
          id: "origin_http_port",
          label: "HTTP 端口",
          type: "number",
          placeholder: "如 80",
          required: true,
        },
        {
          id: "origin_https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: true,
        },
      ],
      {
        id: "comment",
        label: "备注",
        type: "text",
        placeholder: "请输入备注",
        required: false,
      },
    ],
  },
  baidu_cdn: {
    label: "百度云 CDN",
    fields: [
      [
        {
          id: "form",
          label: "类型",
          type: "select",
          placeholder: "请选择类型",
          required: true,
          options: [
            {
              value: "default",
              label: "默认",
            },
            {
              value: "image",
              label: "图片小文件",
            },
            {
              value: "download",
              label: "大文件下载",
            },
            {
              value: "media",
              label: "流媒体点播",
            },
            {
              value: "dynamic",
              label: "动静态加速",
            },
          ],
        },
        {
          id: "follow_protocol",
          label: "协议跟随回源",
          type: "select",
          placeholder: "请选择协议跟随回源",
          required: false,
          options: [
            {
              value: "*",
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
      ],
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
        },
        {
          id: "https_port",
          label: "HTTPS 端口",
          type: "number",
          placeholder: "如 443",
          required: false,
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
  qiniu_cdn: {
    label: "七牛云 CDN",
    fields: [
      [
        {
          id: "platform",
          label: "平台类型",
          type: "select",
          placeholder: "请选择平台类型",
          required: true,
          options: [
            {
              value: "web",
              label: "网页",
            },
            {
              value: "download",
              label: "下载",
            },
            {
              value: "vod",
              label: "点播",
            },
            {
              value: "dynamic",
              label: "动态加速",
            },
          ],
        },
        {
          id: "geo_cover",
          label: "地域",
          type: "select",
          placeholder: "请选择地域",
          required: false,
          options: [
            {
              value: "china",
              label: "中国大陆",
            },
            {
              value: "foreign",
              label: "海外",
            },
            {
              value: "global",
              label: "全球",
            },
          ],
        },
      ],
      {
        id: "source_url_scheme",
        label: "回源协议",
        type: "select",
        placeholder: "请选择回源协议",
        required: false,
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
};

/**
 * 从后端接口获取指定平台的动态 Schema，并更新 CDN_SCHEMA
 * @param {string} platform - 平台类型
 * @param {number} configId - 配置 ID
 * @returns {Promise<void>}
 */
async function fetchAndReplaceCdnSchema(platform, configId) {
  try {
    // 确保 CDN_SCHEMA[platform] 存在
    if (!CDN_SCHEMA[platform]) {
      console.warn(`未知平台: ${platform}`);
      return;
    }

    // 从后端获取动态字段
    const { data: resp } = await API.get(
      `/api/schemas/${platform}/cdn-record/${configId}`,
    );
    console.log("resp", resp);

    // 判断 resp 是否为空
    if (!resp || !resp.fields || resp.fields.length === 0) {
      console.warn(`平台 ${platform} 动态字段为空`);
      return;
    }

    // 只替换 fields，保留 label 等静态信息
    CDN_SCHEMA[platform].fields = resp.fields[0];
  } catch (error) {
    console.error("加载动态 Schema 失败:", error);
  }
}

/**
 * 动态渲染字段
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {Object} initialData - 编辑模式下的初始数据 (可选)
 * @param {string[]} disabledKeys - 禁用的字段 key 列表 (可选)
 */
function renderCdnFields(platform, initialData = {}, disabledKeys = []) {
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

      html += `<div class="form-group ${groupClassName} ${groupClassName} grid grid-cols-${item.length}" data-platform="${platform}">`;
      item.forEach((subField) => {
        html += generateCdnFieldHtml(platform, subField, data, disabledKeys);
      });
      html += `</div>`;
    } else {
      // 普通非分组字段
      html += generateCdnFieldHtml(platform, item, data, disabledKeys);
    }
  });

  container.innerHTML = html;

  // 触发所有 select 的初始化处理
  const selects = container.querySelectorAll("select.dynamic-input");
  selects.forEach((select) => {
    handleSelectChange(select);
  });
}

/**
 * 生成单个字段 HTML，通过 ID 自动映射特殊属性
 * @param {string} platform - 平台 key (如 cloudflare)
 * @param {Object} field - 字段定义对象
 * @param {Object} data - 表单数据对象
 * @param {string[]} disabledKeys - 禁用的字段 key 列表
 * @returns {string} - 生成的 HTML 字符串
 */
function generateCdnFieldHtml(platform, field, data = {}, disabledKeys = []) {
  // 自动属性映射 (根据 ID 直接映射)
  let specialClass = "";
  let dataTypeAttr = "";
  let isHidden = false;

  const value = data[field.id] !== undefined ? data[field.id] : "";
  const requiredAttr = field.required && !isHidden ? "required" : "";
  const disabledAttr = disabledKeys.includes(field.id) ? "disabled" : "";

  let inputHtml = "";
  if (field.type === "password") {
    inputHtml = `
      <div class="flex items-center gap-2">
        <input type="password" class="ui-input dynamic-input" data-platform="${platform}" data-key="${field.id}" value="${value}" placeholder="${field.placeholder}" ${requiredAttr} ${disabledAttr} autocomplete="new-password" />
        <button type="button" class="btn btn-square btn-sm btn-ghost" onclick="togglePasswordVisibilityByPrev(this)">
          <span class="i-lucide-eye w-4 h-4"></span>
        </button>
      </div>`;
  } else if (field.type === "select") {
    inputHtml = `
          <select class="ui-select dynamic-input" data-platform="${platform}" data-key="${
            field.id
          }" ${requiredAttr} ${disabledAttr} onchange="handleSelectChange(this)">
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
    }" class="ui-input dynamic-input ${specialClass}" data-platform="${platform}" data-key="${
      field.id
    }" value="${value}" 
                  placeholder="${field.placeholder}" ${
                    field.pattern ? `pattern="${field.pattern}"` : ""
                  } 
                  ${
                    field.title ? `title="${field.title}"` : ""
                  } ${requiredAttr} ${disabledAttr} />`;
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

    // 处理 select，额外保存 text 和 value
    if (fieldSchema?.type === "select") {
      const selectedOption = el.options[el.selectedIndex];
      const text = selectedOption ? selectedOption.text : "";

      config[key] = val;
      config[`${key}_text`] = text;
      config[`${key}_value`] = val;

      // 如果有 valueType，根据类型转换
      if (fieldSchema?.valueType === "number" && val !== "") {
        config[`${key}_value`] = Number(val);
      } else if (fieldSchema?.valueType === "boolean") {
        config[`${key}_value`] = val === "true" || val === "1" || val === "yes";
      }
    } else {
      config[key] = val;
    }
  });

  // 处理 IP 组 select
  const ip_group_el = document.getElementById("ip_group_id");
  if (ip_group_el) {
    const selectedOption = ip_group_el.options[ip_group_el.selectedIndex];
    const value = ip_group_el.value.trim();
    const text = selectedOption ? selectedOption.text : "";
    const code = selectedOption ? selectedOption.dataset.code : "";

    if (value !== "") {
      config.ip_group_id = value;
      config.ip_group_id_text = text;
      config.ip_group_id_value = Number(value);

      // 如果有 code，也保存
      if (code) {
        config.ip_group_id_code = code;
      }
    }
  }

  document.getElementById("config_json").value = JSON.stringify(config);
}

/**
 * 处理选择框变化事件
 * @param {HTMLSelectElement} selectEl - 触发事件的选择框元素
 */
function handleSelectChange(selectEl) {
  console.log("handleSelectChange", selectEl);

  // 获取选择框的平台、键和当前选中值
  const platform = selectEl.getAttribute("data-platform");
  const key = selectEl.getAttribute("data-key");
  const selectedValue = selectEl.value;
  console.log("platform", platform);
  console.log("key", key);
  console.log("selectedValue", selectedValue);

  // 获取选项的 text 和 value
  const opt = selectEl.options[selectEl.selectedIndex];
  const text = opt ? opt.text : "";
  const value = opt ? opt.value : "";
  console.log("text", text);
  console.log("value", value);

  platform_el = document.getElementById("platform");
  if (platform_el) {
    platform_el.value = platform;
  }
}
