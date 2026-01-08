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
function changeCapability(el) {
  console.log("changeCapability", el);

  const platform = document.getElementById("platform").value;
  const capability = el.value;

  // 重新渲染，自动处理显隐
  renderPlatformFields(platform, capability);
}

// 请选择平台 platform
function changePlatform(el) {
  console.log("changePlatform", el);

  const platform = el.value;
  const capability = document.getElementById("capability").value;

  // 重新渲染，自动处理显隐
  renderPlatformFields(platform, capability);
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

    // 回填数据
    renderPlatformFields(
      config.platform,
      config.capability,
      config.config_json
    );

    showModal("formModal", "modalTitle", "编辑 平台配置");
  }

  // 创建
  if (!id) {
    document.getElementById("id").value = "";
    document.getElementById("id").disabled = true;
    document.getElementById("platform").disabled = false;

    // 默认渲染第一个
    renderPlatformFields(document.getElementById("platform").value, "all");

    showModal("formModal", "modalTitle", "添加 平台配置");
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 自动从动态字段中组装 JSON
    buildConfigFromDynamicFields();

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
