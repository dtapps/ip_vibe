// 选择配置 config_id
function onConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onConfigSelect", opt);
}

// 选择证书提供方 provider
function onProviderSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onProviderSelect", opt);
}

/**
 * 打开 表单 弹窗
 * @param {*} id 平台配置ID
 * @param {string} action 操作类型，add / edit / copy
 * @returns
 */
async function openFormModal(id, action = "") {
  document.getElementById("form").reset();

  try {
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
      const configSelect = document.getElementById("config_id");
      configSelect.innerHTML = "";
      addPlaceholder(configSelect, "请选择平台配置");
      configs.forEach((d) => {
        if (d.capability === "ssl" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.id = d.id;
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json?.domain_name ?? "";
          opt.dataset.email = d.config_json?.email ?? "";
          opt.dataset.account_id = d.config_json?.account_id ?? "";
          opt.value = d.id;
          opt.text = `${d.name}（${opt.dataset.email}）`;
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
        closeOnSelect: false, // // 多选时防止关闭下拉列表
        dropdownParent: document.querySelector("#modal .ui-modal-box"), // 绑定 modal 容器，防 z-index/定位错
      });
    }

    // 编辑
    if (id) {
      const { data: sslRecord } = await API.get(`/api/ssl-records/${id}`);
      console.log("sslRecord", sslRecord);
      if (!sslRecord) {
        alert("SSL记录数据异常");
        return;
      }

      document.getElementById("id").value = sslRecord.id;
      document.getElementById("id").disabled = true;
      document.getElementById("config_id").value = sslRecord.config_id;
      document.getElementById("config_id").disabled = true;
      document.getElementById("domain").value = sslRecord.domain;
      document.getElementById("domain").disabled = true;
      document.getElementById("provider").value = sslRecord.provider;
      document.getElementById("renewal_days").value =
        sslRecord.renewal_days || 30;

      if (sslRecord.webhooks && sslRecord.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", sslRecord.webhooks);
        selectId2.val(sslRecord.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", sslRecord.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 如果是复制操作，清空 id 字段，取消 domain 禁用
      if (action === "copy") {
        document.getElementById("id").value = "";
        document.getElementById("id").disabled = false;
        document.getElementById("domain").disabled = false;

        showModal("formModal", "modalTitle", "复制 SSL记录");
      } else {
        showModal("formModal", "modalTitle", "编辑 SSL记录");
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
          id: "config_id",
          disabled: false,
        },
        {
          id: "domain",
          value: "",
          disabled: false,
        },
        {
          id: "renewal_days",
          value: 30,
          disabled: false,
        },
      ]);

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      showModal("formModal", "modalTitle", "添加 SSL记录");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开SSL记录模态框失败", error);
    return;
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    const selectedWebhookIDs = $("#webhook_ids").val() || [];

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      config_id: Number(document.getElementById("config_id").value.trim()),
      domain: document.getElementById("domain").value.trim(),
      provider: document.getElementById("provider").value.trim(),
      renewal_days: Number(
        document.getElementById("renewal_days").value.trim(),
      ),
      webhook_ids: selectedWebhookIDs.map(Number),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
      return alert("请选择配置");
    }
    if (!form_data.domain) {
      return alert("请输入域名");
    }
    if (!form_data.provider) {
      return alert("请选择提供方");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/ssl-records", form_data);
      } else {
        await API.patch(`/api/ssl-records/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("SSL记录操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 打开 查看 弹窗
async function openViewModal(id) {
  try {
    const { data: sslRecord } = await API.get(`/api/ssl-records/${id}`);
    console.log("sslRecord", sslRecord);
    if (!sslRecord) {
      alert("SSL记录数据异常");
      return;
    }

    const crtEl = document.getElementById("sslCrt");
    const keyEl = document.getElementById("sslKey");

    if (crtEl) crtEl.value = sslRecord.cert_certificate || "暂无证书内容";
    if (keyEl) keyEl.value = sslRecord.cert_private_key || "暂无私钥内容";

    showModal("viewModal", "modalTitle", "证书内容详情");
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开SSL证书详情模态框失败", error);
    return;
  }
}

// 打开 查看 弹窗
async function openHistoryViewModal(id, historyId) {
  try {
    const { data: sslRecord } = await API.get(
      `/api/ssl-records/${id}/history/${historyId}`,
    );
    console.log("sslRecord", sslRecord);
    if (!sslRecord) {
      alert("SSL记录数据异常");
      return;
    }

    const crtEl = document.getElementById("sslCrt");
    const keyEl = document.getElementById("sslKey");

    if (crtEl) crtEl.value = sslRecord.cert_certificate || "暂无证书内容";
    if (keyEl) keyEl.value = sslRecord.cert_private_key || "暂无私钥内容";

    showModal("viewModal", "modalTitle", "证书内容详情");
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开SSL证书详情模态框失败", error);
    return;
  }
}

// 申请新的SSL证书
async function applySSL(id) {
  if (confirm("确定申请新的SSL证书？")) {
    try {
      await API.post(`/api/ssl-records/${id}/apply`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`申请失败: ${error.message}`);
      console.error("申请新的SSL证书失败", error);
    }
  }
}
