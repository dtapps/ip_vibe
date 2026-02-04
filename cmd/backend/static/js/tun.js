// 选择 配置 config_id
async function onConfigSelect(el) {
  console.log("onConfigSelect", el);

  const opt = el.options[el.selectedIndex];

  const id = opt.value || 0;
  const platform = opt.dataset.platform || "";

  // 从后端接口获取指定平台的动态 Schema
  if (id > 0 && platform) {
    switch (platform) {
      case "cloudflare":
        await fetchAndReplaceTunSchema(platform, id);
        break;
    }
  }

  // 重新渲染字段
  renderTunFields(platform);
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

    // 加载 jQuery 库
    await loadJQuery();
    // 加载 Select2 库
    await loadSelect2();

    // 选择 Select2 控件
    const selectId2 = $("#webhook_ids");

    // 平台配置列表
    const { data: configs } = await API.get(`/api/platform-configs`);
    console.log("configs", configs);
    if (configs) {
      const configSelect = els.config_id;
      configSelect.innerHTML = "";
      addPlaceholder(configSelect, "请选择平台配置");
      configs.forEach((d) => {
        if (d.capability === "tun" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.id = d.id;
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json?.domain_name ?? "";
          opt.dataset.email = d.config_json?.email ?? "";
          opt.dataset.account_id = d.config_json?.account_id ?? "";
          opt.value = d.id;
          opt.text = `${d.name}（${opt.dataset.account_id}）`;
          configSelect.appendChild(opt);
        }
      });
    }

    // 通知列表
    const { data: webhooks } = await API.get(`/api/webhooks`, {
      is_active: true,
    });
    console.log("webhooks", webhooks);
    if (webhooks) {
      // 销毁旧实例（防止 Select2 冲突）
      if (selectId2.data("select2")) {
        selectId2.select2("destroy");
      }

      // 准备 Select2 需要的数据格式
      const select2Data = [];
      if (webhooks && webhooks.length > 0) {
        webhooks.forEach((d) => {
          select2Data.push({
            id: d.id,
            text: `${d.name} (${d.platform})`,
          });
        });
      }

      // 初始化 Select2 控件
      console.log("select2 初始化", selectId2, select2Data);
      selectId2.select2({
        theme: "tailwindcss-4", // 主题
        width: "100%",
        placeholder: "请选择通知 Webhook",
        allowClear: true,
        multiple: true,
        data: select2Data, // 数据
        closeOnSelect: false, // 多选时防止关闭下拉列表
        dropdownParent: document.querySelector("#modal .ui-modal-box"), // 绑定 modal 容器，防 z-index/定位错
      });
    }

    // 编辑
    if (id) {
      const { data: tunRecord } = await API.get(`/api/tun-records/${id}`);
      console.log("tunRecord", tunRecord);
      if (!tunRecord) {
        alert("TUN记录数据异常");
        return;
      }

      els.id.value = tunRecord.id;
      els.id.disabled = true;
      els.platform.value = tunRecord.platform;
      els.platform.disabled = true;
      els.config_id.value = tunRecord.config_id;
      els.config_id.disabled = true;
      els.sub_domain.value = tunRecord.sub_domain;
      els.domain_path.value = tunRecord.domain_path;
      els.serve_type.value = tunRecord.serve_type;
      els.serve_uri.value = tunRecord.serve_uri;

      if (tunRecord.webhooks && tunRecord.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", tunRecord.webhooks);
        selectId2.val(tunRecord.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", tunRecord.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 从后端接口获取指定平台的动态 Schema
      const disabledKeys = [];
      switch (tunRecord.platform) {
        case "cloudflare":
          await fetchAndReplaceTunSchema(
            tunRecord.platform,
            tunRecord.config_id,
          );
          disabledKeys.push("tunnel_id");
          disabledKeys.push("zone_id");
          break;
      }

      // 渲染字段，回填数据
      els.original_config_json.value = JSON.stringify(
        tunRecord.config_json || {},
      );
      renderTunFields(
        tunRecord.platform,
        tunRecord.config_json || {},
        disabledKeys,
      );

      // 如果是复制操作，清空 id 字段
      if (action === "copy") {
        els.id.value = "";
        els.id.disabled = false;

        showModal("formModal", "modalTitle", "复制 DNS记录");
      } else {
        showModal("formModal", "modalTitle", "编辑 DNS记录");
      }
    }

    // 创建
    if (!id) {
      els.id.value = 0;
      els.id.disabled = true;
      els.platform.value = "";
      els.platform.disabled = true;
      els.config_id.disabled = false;

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      // 默认渲染字段
      renderTunFields(document.getElementById("platform").value.trim() || "");

      showModal("formModal", "modalTitle", "添加 DNS记录");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开DNS记录模态框失败", error);
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
      buildConfigFromTunFields(els.platform.value.trim());

      // 读取组装 JSON 字符串
      let configObj = {};
      const rawJson = els.config_json.value.trim();
      if (rawJson) {
        try {
          configObj = JSON.parse(rawJson);
        } catch (err) {
          console.warn("config_json 解析失败，使用空对象", err);
        }
      }

      // 处理平台自定义
      switch (els.platform.value.trim()) {
        case "cloudflare":
          // 隧道
          const tunnelSelect = form.querySelector(
            'select[data-key="tunnel_id"]',
          );
          if (!tunnelSelect?.value) {
            return alert("请选择隧道");
          }
          // 注入
          configObj.tunnel_name =
            tunnelSelect.selectedOptions[0]?.dataset.label || "";

          // 域名
          const zoneSelect = form.querySelector('select[data-key="zone_id"]');
          if (!zoneSelect?.value) {
            return alert("请选择域名");
          }
          // 注入
          configObj.zone_name =
            zoneSelect.selectedOptions[0]?.dataset.label || "";
          break;
      }

      const selectedWebhookIDs = $("#webhook_ids").val() || [];

      // 表单数据
      const form_id = Number(els.id.value.trim());
      const form_data = {
        config_id: Number(els.config_id.value.trim()),
        sub_domain: els.sub_domain.value.trim(),
        domain_path: els.domain_path.value.trim(),
        serve_type: els.serve_type.value.trim(),
        serve_uri: els.serve_uri.value.trim(),
        webhook_ids: selectedWebhookIDs.map(Number),
        config_json: JSON.stringify(configObj),
      };
      console.log("form_id, form_data", form_id, form_data);

      // 校验
      if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
        return alert("请选择配置");
      }
      if (!form_data.sub_domain) {
        return alert("请输入子域名");
      }
      if (!form_data.serve_type) {
        return alert("请选择服务类型");
      }
      if (!form_data.serve_uri) {
        return alert("请输入服务 URI");
      }

      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/tun-records", form_data);
      } else {
        await API.patch(`/api/tun-records/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error("onsubmit failed", err);
    }
  };
}
