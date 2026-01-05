/**
 * 判断是否需要 domain_name 字段
 * @param {string} capability - 服务能力
 * @returns {boolean} 是否需要 domain_name 字段
 */
function needsDomainField(capability) {
  return capability === "all" || capability === "cdn" || capability === "dns";
}

/**
 * 判断是否需要 email 字段
 * @param {string} capability - 服务能力
 * @returns {boolean} 是否需要 email 字段
 */
function needsEmailField(capability) {
  return capability === "all" || capability === "ssl";
}

/**
 * 获取 domain_name 和 email 字段值及类型
 * @param {string} platform - 平台名称
 * @param {string} capability - 服务能力
 * @returns {Object} { domain_name?: string, email?: string }
 */
function getDomainAndEmailFieldInfo(platform, capability) {
  const result = {};

  // 总是尝试获取 domain_name（如果字段存在）
  const domainField = document.getElementById(`${platform}_domain_name`);
  if (domainField && needsDomainField(capability)) {
    const value = domainField.value.trim();
    if (value) {
      result.domain_name = value;
    }
  }

  // 总是尝试获取 email（如果字段存在）
  const emailField = document.getElementById(`${platform}_email`);
  if (emailField && needsEmailField(capability)) {
    const value = emailField.value.trim();
    if (value) {
      result.email = value;
    }
  }

  return result;
}

/**
 * 直接返回键值对，根据服务能力判断返回 domain_name 和/或 email
 * @param {string} platform - 平台名称
 * @param {string} capability - 服务能力
 * @returns {Object} { domain_name?: string, email?: string }
 */
function getDomainAndEmailKeyValue(platform, capability) {
  return getDomainAndEmailFieldInfo(platform, capability);
}

/**
 * 设置 domain_name 和 email 字段值
 * @param {string} platform - 平台名称
 * @param {string} capability - 服务能力
 * @param {string} domain_name_value - domain_name 值
 * @param {string} email_value - email 值
 * @returns {Object} { domain_name?: string, email?: string }
 */
function setDomainAndEmailField(
  platform,
  capability,
  domain_name_value,
  email_value
) {
  const result = {};

  // 设置 domain_name（如果需要且值不为空）
  if (needsDomainField(capability)) {
    const domainField = document.getElementById(`${platform}_domain_name`);
    if (domainField) {
      domainField.value = domain_name_value || "";
      if (domain_name_value) {
        result.domain_name = domain_name_value;
      }
    }
  }

  // 设置 email（如果需要且值不为空）
  if (needsEmailField(capability)) {
    const emailField = document.getElementById(`${platform}_email`);
    if (emailField) {
      emailField.value = email_value || "";
      if (email_value) {
        result.email = email_value;
      }
    }
  }

  return result;
}

// 请选择服务能力 capability
function changeCapability(sel) {
  console.log("changeCapability", sel);

  const capability = sel.value;
  const domainOrEmailGroups = document.querySelectorAll(
    ".domain-or-email-group"
  );

  domainOrEmailGroups.forEach((group) => {
    const domainField = group.querySelector(".domain-field");
    const emailField = group.querySelector(".email-field");
    const domainLabel = group.querySelector('[data-type="domain"]');
    const emailLabel = group.querySelector('[data-type="email"]');

    if (capability === "all") {
      // 同时显示 domain 和 email
      domainField.classList.remove("hidden");
      emailField.classList.remove("hidden");
      domainLabel.classList.remove("hidden");
      emailLabel.classList.remove("hidden");
    } else if (capability === "ssl") {
      // 只显示 email，隐藏 domain
      domainField.classList.add("hidden");
      emailField.classList.remove("hidden");
      domainLabel.classList.add("hidden");
      emailLabel.classList.remove("hidden");

      // 清空 domain 字段值
      domainField.value = "";
    } else {
      // 只显示 domain，隐藏 email
      domainField.classList.remove("hidden");
      emailField.classList.add("hidden");
      domainLabel.classList.remove("hidden");
      emailLabel.classList.add("hidden");

      // 清空 email 字段值
      emailField.value = "";
    }
  });
}

// 请选择平台 platform
function changePlatform(platform) {
  // 隐藏所有平台字段
  document.querySelectorAll(".platform-fields").forEach((el) => {
    el.classList.add("hidden");
  });

  // 显示选中平台的字段
  const fields = document.getElementById(platform + "_fields");
  if (fields) {
    fields.classList.remove("hidden");

    // 根据当前 capability 设置字段显示
    const currentCapability = document.getElementById("capability").value;
    if (currentCapability) {
      changeCapability({ value: currentCapability });
    }
  }
}

// 组装JSON配置
function buildConfigPlatform() {
  const platform = document.getElementById("platform").value.trim();
  const capability = document.getElementById("capability").value.trim();
  let config = {};

  switch (platform) {
    case "cloudflare":
      config = {
        api_key: document.getElementById(`${platform}_api_key`).value.trim(),
        zone_id: document.getElementById(`${platform}_zone_id`).value.trim(),
      };
      break;
    case "tencent_dns":
      config = {
        secret_id: document
          .getElementById(`${platform}_secret_id`)
          .value.trim(),
        secret_key: document
          .getElementById(`${platform}_secret_key`)
          .value.trim(),
      };
      break;
    case "tencent_eo":
      config = {
        secret_id: document
          .getElementById(`${platform}_secret_id`)
          .value.trim(),
        secret_key: document
          .getElementById(`${platform}_secret_key`)
          .value.trim(),
        zone_id: document.getElementById(`${platform}_zone_id`).value.trim(),
      };
      break;
    case "aliyun_dns":
      config = {
        access_key_id: document
          .getElementById(`${platform}_access_key_id`)
          .value.trim(),
        access_key_secret: document
          .getElementById(`${platform}_access_key_secret`)
          .value.trim(),
      };
      break;
    case "aliyun_esa":
      config = {
        access_key_id: document
          .getElementById(`${platform}_access_key_id`)
          .value.trim(),
        access_key_secret: document
          .getElementById(`${platform}_access_key_secret`)
          .value.trim(),
        site_id: Number(
          document.getElementById(`${platform}_site_id`).value.trim()
        ),
      };
      break;
    case "baidu_dns":
      config = {
        access_key_id: document
          .getElementById(`${platform}_access_key_id`)
          .value.trim(),
        secret_access_key: document
          .getElementById(`${platform}_secret_access_key`)
          .value.trim(),
      };
      break;
    case "huawei_dns":
      config = {
        access_key_id: document
          .getElementById(`${platform}_access_key_id`)
          .value.trim(),
        secret_access_key: document
          .getElementById(`${platform}_secret_access_key`)
          .value.trim(),
        zone_id: document.getElementById(`${platform}_zone_id`).value.trim(),
      };
    case "volcengine_dns":
      config = {
        access_key: document
          .getElementById(`${platform}_access_key`)
          .value.trim(),
        secret_key: document
          .getElementById(`${platform}_secret_key`)
          .value.trim(),
        zone_id: Number(
          document.getElementById(`${platform}_zone_id`).value.trim()
        ),
      };
      break;
  }

  // 合并 domain_name 和 email 字段（可能同时存在）
  Object.assign(config, getDomainAndEmailKeyValue(platform, capability));

  // 设置隐藏字段的值
  document.getElementById("config_json").value = JSON.stringify(config);
}

// 打开 表单 弹窗
async function openFormModal(id) {
  document.getElementById("form").reset();

  // 编辑
  if (id) {
    const { data: config } = await API.get(`/api/platform-configs/${id}`);
    console.log("config", config);
    if (!config) {
      alert("配置数据异常");
      return;
    }

    document.getElementById("id").value = config.id;
    document.getElementById("id").disabled = true;
    document.getElementById("capability").value = config.capability;
    document.getElementById("name").value = config.name;
    document.getElementById("platform").value = config.platform;
    document.getElementById("platform").disabled = true;

    // 解析JSON并填充表单
    changePlatform(config.platform);

    // 根据 capability 设置字段显示
    setTimeout(() => {
      changeCapability({ value: config.capability });
    }, 100);

    switch (config.platform) {
      case "cloudflare":
        document.getElementById(`${config.platform}_api_key`).value =
          config.config_json.api_key || "";
        document.getElementById(`${config.platform}_zone_id`).value =
          config.config_json.zone_id || "";
        break;
      case "tencent_dns":
        document.getElementById(`${config.platform}_secret_id`).value =
          config.config_json.secret_id || "";
        document.getElementById(`${config.platform}_secret_key`).value =
          config.config_json.secret_key || "";
        break;
      case "tencent_eo":
        document.getElementById(`${config.platform}_secret_id`).value =
          config.config_json.secret_id || "";
        document.getElementById(`${config.platform}_secret_key`).value =
          config.config_json.secret_key || "";
        document.getElementById(`${config.platform}_zone_id`).value =
          config.config_json.zone_id || "";
        break;
      case "aliyun_dns":
        document.getElementById(`${config.platform}_access_key_id`).value =
          config.config_json.access_key_id || "";
        document.getElementById(`${config.platform}_access_key_secret`).value =
          config.config_json.access_key_secret || "";
        break;
      case "aliyun_esa":
        document.getElementById(`${config.platform}_access_key_id`).value =
          config.config_json.access_key_id || "";
        document.getElementById(`${config.platform}_access_key_secret`).value =
          config.config_json.access_key_secret || "";
        document.getElementById(`${config.platform}_site_id`).value =
          config.config_json.site_id || 0;
        break;
      case "baidu_dns":
        document.getElementById(`${config.platform}_access_key_id`).value =
          config.config_json.access_key_id || "";
        document.getElementById(`${config.platform}_secret_access_key`).value =
          config.config_json.secret_access_key || "";
        break;
      case "huawei_dns":
        document.getElementById(`${config.platform}_access_key_id`).value =
          config.config_json.access_key_id || "";
        document.getElementById(`${config.platform}_secret_access_key`).value =
          config.config_json.secret_access_key || "";
        document.getElementById(`${config.platform}_zone_id`).value =
          config.config_json.zone_id || "";
      case "volcengine_dns":
        document.getElementById(`${config.platform}_access_key`).value =
          config.config_json.access_key || "";
        document.getElementById(`${config.platform}_secret_key`).value =
          config.config_json.secret_key || "";
        document.getElementById(`${config.platform}_zone_id`).value =
          config.config_json.zone_id || 0;
        break;
    }

    // 设置 domain_name 或 email 字段值
    setDomainAndEmailField(
      config.platform,
      config.capability,
      config.config_json.domain_name || "",
      config.config_json.email || ""
    );

    showModal("formModal", "modalTitle", "编辑 平台配置");
  }

  // 创建
  if (!id) {
    document.getElementById("id").value = "";
    document.getElementById("id").disabled = true;
    document
      .querySelectorAll(".platform-fields")
      .forEach((el) => el.classList.add("hidden"));
    document.getElementById("platform").disabled = false;

    // 重置所有字段显示
    document.querySelectorAll(".domain-or-email-group").forEach((group) => {
      const domainField = group.querySelector(".domain-field");
      const emailField = group.querySelector(".email-field");
      const domainLabel = group.querySelector('[data-type="domain"]');
      const emailLabel = group.querySelector('[data-type="email"]');

      domainField.classList.remove("hidden");
      emailField.classList.add("hidden");
      domainLabel.classList.remove("hidden");
      emailLabel.classList.add("hidden");
    });

    showModal("formModal", "modalTitle", "添加 平台配置");
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 组装JSON
    buildConfigPlatform();

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      capability: document.getElementById("capability").value.trim(),
      name: document.getElementById("name").value.trim(),
      platform: document.getElementById("platform").value.trim(),
      config_json: document.getElementById("config_json").value.trim(),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.capability) {
      return alert("请选择服务能力");
    }
    if (!form_data.name) {
      return alert("请输入组名");
    }
    if (!form_data.platform) {
      return alert("请选择平台");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/platform-configs", form_data);
      } else {
        await API.patch(`/api/platform-configs/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("配置操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 删除
async function onDelete(id) {
  if (confirm("确定删除此配置？相关记录也会被删除")) {
    try {
      await API.delete(`/api/platform-configs/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除配置失败", error);
    }
  }
}
