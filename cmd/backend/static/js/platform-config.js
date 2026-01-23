// 请选择服务能力 capability
function changeCapability(el) {
  console.log("changeCapability", el);

  const platform = document.getElementById("platform").value;
  const capability = el.value;

  // 重新渲染字段，自动处理显隐
  renderPlatformFields(platform, capability);
}

// 请选择平台 platform
function changePlatform(el) {
  console.log("changePlatform", el);

  const platform = el.value;
  const capability = document.getElementById("capability").value;

  // 重新渲染字段，自动处理显隐
  renderPlatformFields(platform, capability);
}

/**
 * 打开 表单 弹窗
 * @param {*} id 平台配置ID
 * @param {string} action 操作类型，add / edit / copy
 * @returns
 */
async function openFormModal(id, action = "") {
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

    // 渲染字段，回填数据，自动处理显隐
    renderPlatformFields(
      config.platform,
      config.capability,
      config.config_json || {},
    );

    // 如果是复制操作，清空 id 字段
    if (action === "copy") {
      document.getElementById("id").value = "";
      document.getElementById("id").disabled = false;

      showModal("formModal", "modalTitle", "复制 平台配置");
    } else {
      showModal("formModal", "modalTitle", "编辑 平台配置");
    }
  }

  // 创建
  if (!id) {
    updateFormFields([
      {
        id: "id",
        value: 0,
        disabled: true,
      },
      {
        id: "platform",
        value: "",
        disabled: false,
      },
    ]);

    // 默认渲染字段
    renderPlatformFields(
      document.getElementById("platform").value.trim(),
      "all",
    );

    showModal("formModal", "modalTitle", "添加 平台配置");
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 自动从动态字段中组装 JSON
    buildConfigFromPlatformFields(
      document.getElementById("platform").value.trim(),
    );

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
