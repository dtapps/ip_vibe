// 定义字段 Schema
const DNS_SCHEMA = {
  cloudflare: {
    label: "CloudFlare",
    fields: [
      {
        id: "zone",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "comment",
          label: "描述",
          type: "text",
          placeholder: "请输入描述",
          required: false,
        },
      ],
    ],
  },
  tencent_dns: {
    label: "腾讯云 DNS",
    fields: [
      {
        id: "domain",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "record_line",
          label: "线路",
          type: "select",
          placeholder: "请选择线路",
          required: true,
          options: [
            {
              value: "默认",
              label: "默认",
            },
          ],
        },
        {
          id: "remark",
          label: "备注",
          type: "text",
          placeholder: "请输入备注",
          required: false,
        },
      ],
    ],
  },
  aliyun_dns: {
    label: "阿里云 DNS",
    fields: [
      {
        id: "domain",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "line",
          label: "线路",
          type: "select",
          placeholder: "请选择线路",
          required: true,
          options: [
            {
              value: "default",
              label: "默认",
            },
          ],
        },
      ],
    ],
  },
  baidu_dns: {
    label: "百度云 DNS",
    fields: [
      {
        id: "zone",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "line",
          label: "线路",
          type: "select",
          placeholder: "请选择线路",
          required: true,
          options: [
            {
              value: "default",
              label: "默认",
            },
            {
              value: "ct",
              label: "电信",
            },
            {
              value: "cmnet",
              label: "移动",
            },
            {
              value: "cnc",
              label: "联通",
            },
            {
              value: "edu",
              label: "教育网",
            },
            {
              value: "search",
              label: "搜索引擎(百度)",
            },
          ],
        },
        {
          id: "description",
          label: "描述",
          type: "text",
          placeholder: "请输入描述",
          required: false,
        },
      ],
    ],
  },
  huawei_dns: {
    label: "华为云 DNS",
    fields: [
      {
        id: "zone",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "description",
          label: "描述",
          type: "text",
          placeholder: "请输入描述",
          required: false,
        },
      ],
    ],
  },
  volcengine_dns: {
    label: "火山引擎 DNS",
    fields: [
      {
        id: "zone",
        label: "域名",
        type: "select",
        placeholder: "请选择域名",
        required: true,
        options: [],
      },
      [
        {
          id: "line",
          label: "线路",
          type: "select",
          placeholder: "请选择线路",
          required: true,
          options: [
            {
              value: "default",
              label: "默认",
            },
          ],
        },
        {
          id: "remark",
          label: "备注",
          type: "text",
          placeholder: "请输入备注",
          required: false,
        },
      ],
    ],
  },
};

/**
 * 从后端接口获取指定平台的动态 Schema，并更新 DNS_SCHEMA
 * @param {string} platform - 平台类型
 * @param {number} configId - 配置 ID
 * @returns {Promise<void>}
 */
async function fetchAndReplaceDnsSchema(platform, configId) {
  try {
    // 确保 DNS_SCHEMA[platform] 存在
    if (!DNS_SCHEMA[platform]) {
      console.warn(`未知平台: ${platform}`);
      return;
    }

    // 从后端获取动态字段
    const { data: resp } = await API.get(
      `/api/schemas/${platform}/dns-record/${configId}`,
    );
    console.log("resp", resp);

    // 判断 resp 是否为空
    if (!resp || !resp.fields || resp.fields.length === 0) {
      console.warn(`平台 ${platform} 动态字段为空`);
      return;
    }

    // 只替换 fields，保留 label 等静态信息
    DNS_SCHEMA[platform].fields = resp.fields[0];
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
function renderDnsFields(platform, initialData = {}, disabledKeys = []) {
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

      html += `<div class="form-group ${groupClassName} grid grid-cols-${item.length}" data-platform="${platform}">`;
      item.forEach((subField) => {
        html += generateDnsFieldHtml(platform, subField, data, disabledKeys);
      });
      html += `</div>`;
    } else {
      // 普通非分组字段
      html += generateDnsFieldHtml(platform, item, data, disabledKeys);
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
function generateDnsFieldHtml(platform, field, data = {}, disabledKeys = []) {
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

  // 处理 CDN 记录 select
  const cdn_record_id = document.getElementById("cdn_record_id");
  if (cdn_record_id) {
    const selectedOption = cdn_record_id.options[cdn_record_id.selectedIndex];
    const value = cdn_record_id.value.trim();
    const text = selectedOption ? selectedOption.text : "";

    if (value !== "") {
      config.cdn_record_id = value;
      config.cdn_record_id_text = text;
      config.cdn_record_id_value = Number(value);
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

  const platform_el = document.getElementById("platform");
  if (platform_el) {
    platform_el.value = platform;
  }

  // 动态更新域名显示
  const domain_name_el = document.getElementById("domain_name");
  // const domain_name_label = document.getElementById("domain_name_label");
  if (domain_name_el && platform && key && text) {
    // 隐藏标题
    // if (domain_name_label) {
    //   domain_name_label.style.display = "none";
    // }
    switch (platform) {
      case "cloudflare":
        if (key === "zone") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      case "tencent_dns":
        if (key === "domain") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      case "aliyun_dns":
        if (key === "domain") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      case "baidu_dns":
        if (key === "zone") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      case "huawei_dns":
        if (key === "zone") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      case "volcengine_dns":
        if (key === "zone") {
          // domain_name_el.innerText = `.${text}`;
          domain_name_el.value = text;
          domain_name_el.disabled = true;
        }
        break;
      default:
        domain_name_el.disabled = false;
        break;
    }
  }
}
