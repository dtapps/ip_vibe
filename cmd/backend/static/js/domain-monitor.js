// 选择类型 check_protocol
function onCheckProtocolSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onCheckProtocolSelect", opt);
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
      const { data: domainMonitor } = await API.get(
        `/api/domain-monitors/${id}`,
      );
      console.log("domainMonitor", domainMonitor);
      if (!domainMonitor) {
        alert("域名监控数据异常");
        return;
      }

      document.getElementById("id").value = domainMonitor.id;
      document.getElementById("id").disabled = true;
      document.getElementById("name").value = domainMonitor.name;
      document.getElementById("description").value =
        domainMonitor.description ?? "";
      document.getElementById("domain").value = domainMonitor.domain;
      document.getElementById("domain").disabled = true;
      document.getElementById("request_headers").value =
        domainMonitor.request_headers == ""
          ? domainMonitor.request_headers.join("\n")
          : "";
      document.getElementById("check_interval").value =
        domainMonitor.check_interval;
      document.getElementById("check_protocol").value =
        domainMonitor.check_protocol;
      document.getElementById("check_protocol").disabled = true;

      if (domainMonitor.webhooks && domainMonitor.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", domainMonitor.webhooks);
        selectId2.val(domainMonitor.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", domainMonitor.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 如果是复制操作，清空 id 字段，取消 domain 、check_protocol 禁用
      if (action === "copy") {
        document.getElementById("id").value = "";
        document.getElementById("id").disabled = false;
        document.getElementById("domain").disabled = false;
        document.getElementById("check_protocol").disabled = false;

        showModal("formModal", "modalTitle", "复制 域名监控");
      } else {
        showModal("formModal", "modalTitle", "编辑 域名监控");
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
          id: "domain",
          disabled: false,
        },
        {
          id: "check_interval",
          value: 60,
          disabled: false,
        },
        {
          id: "check_protocol",
          disabled: false,
        },
      ]);

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      showModal("formModal", "modalTitle", "添加 域名监控");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开域名监控模态框失败", error);
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
      name: document.getElementById("name").value.trim(),
      description: document.getElementById("description").value.trim(),
      domain: document.getElementById("domain").value.trim(),
      request_headers: document
        .getElementById("request_headers")
        .value.trim()
        .split("\n"),
      check_interval: Number(
        document.getElementById("check_interval").value.trim(),
      ),
      check_protocol: document.getElementById("check_protocol").value.trim(),
      webhook_ids: selectedWebhookIDs.map(Number),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.domain) {
      return alert("请输入域名");
    }
    if (!form_data.check_protocol) {
      return alert("请选择检查协议");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/domain-monitors", form_data);
      } else {
        await API.patch(`/api/domain-monitors/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("域名监控操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}
