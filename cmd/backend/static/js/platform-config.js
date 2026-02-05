// 选择 服务能力 capability
function changeCapability(el) {
  console.log("changeCapability", el);

  // 获取表单元素
  const form = document.getElementById("form");
  if (!form) {
    throw new Error("未找到表单");
  }

  // 表单elements集合
  const els = form.elements;

  const platform = els.platform.value.trim();
  const capability = el.value.trim();
  const original_config_json = els.original_config_json.value.trim() || "{}";

  // 重新渲染字段
  renderPlatformFields(platform, capability, JSON.parse(original_config_json));
}

// 选择 平台类型 platform
function changePlatform(el) {
  console.log("changePlatform", el);

  // 获取表单元素
  const form = document.getElementById("form");
  if (!form) {
    throw new Error("未找到表单");
  }

  // 表单elements集合
  const els = form.elements;

  const platform = els.platform.value.trim();
  const capability = els.capability.value.trim();

  // 重新渲染字段
  renderPlatformFields(platform, capability);
}

/**
 * 打开 表单 弹窗
 * @param {*} id 平台配置ID
 * @param {string} action 操作类型，add / edit / copy
 * @returns
 */
async function openFormModal(id, action = "") {
  const form = document.getElementById("form");
  if (!form) {
    throw new Error("未找到表单");
  }

  // 重置表单
  form.reset();

  try {
    // 表单elements集合
    const els = form.elements;

    // 编辑
    if (id) {
      const { data: config } = await API.get(`/api/platform-configs/${id}`);
      console.log("config", config);
      if (!config) {
        alert("配置数据异常");
        return;
      }

      els.id.value = config.id;
      els.id.disabled = true;
      els.name.value = config.name;
      els.description.value = config.description || "";
      els.capability.value = config.capability;
      els.platform.value = config.platform;
      els.platform.disabled = true;

      // 渲染字段，回填数据
      els.original_config_json.value = JSON.stringify(config.config_json || {});
      renderPlatformFields(
        config.platform,
        config.capability,
        config.config_json || {},
      );

      // 如果是复制操作，清空 id 字段
      if (action === "copy") {
        els.id.value = "";
        els.id.disabled = false;

        showModal("formModal", "modalTitle", "复制 平台配置");
      } else {
        showModal("formModal", "modalTitle", "编辑 平台配置");
      }
    }

    // 创建
    if (!id) {
      els.id.value = 0;
      els.id.disabled = true;
      els.platform.value = "";
      els.platform.disabled = false;

      // 默认渲染字段
      renderPlatformFields(
        els.platform.value.trim() || "",
        els.capability.value.trim() || "all",
      );

      showModal("formModal", "modalTitle", "添加 平台配置");
    }
  } catch (err) {
    alert(`打开 表单 失败: ${err.message}`);
    console.error(err);
    return;
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    try {
      // 阻止事件默认行为
      e.preventDefault();

      // 表单elements集合
      const els = form.elements;

      // 自动从动态字段中组装 JSON
      buildConfigFromPlatformFields(els.platform.value.trim());

      // 表单数据
      const form_id = Number(els.id.value.trim());
      const form_data = {
        name: els.name.value.trim(),
        description: els.description.value.trim(),
        capability: els.capability.value.trim(),
        platform: els.platform.value.trim(),
        config_json: els.config_json.value.trim(),
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

      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/platform-configs", form_data);
      } else {
        await API.patch(`/api/platform-configs/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error(err);
    }
  };
}
