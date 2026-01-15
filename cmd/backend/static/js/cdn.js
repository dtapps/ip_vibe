// 选择配置 config_id
function onConfigSelect(el) {
  console.log("onConfigSelect", el);

  const opt = el.options[el.selectedIndex];

  const platform = opt.dataset.platform || "";
  const domain_name = opt.dataset.domain_name || "";

  if (platform) {
    document.getElementById("platform").value = platform;
  }
  if (domain_name) {
    document.getElementById("domain_name").innerHTML = "." + domain_name;
  }

  // 重新渲染字段
  renderCdnFields(platform);
}

// 选择类型 type
function onTypeSelect(el) {
  console.log("onTypeSelect", el);

  const opt = el.options[el.selectedIndex];
}

// 选择IP组 ip_group_id
function onIpGroupSelect(el) {
  console.log("onIpGroupSelect", el);

  const opt = el.options[el.selectedIndex];
}

// 打开 表单 弹窗
async function openFormModal(id) {
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
        if (d.capability === "cdn" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json_domain_name || "";
          opt.value = d.id;
          opt.text = `${d.name}（${opt.dataset.domain_name}）`;
          configSelect.appendChild(opt);
        }
      });
    }

    // IP 组列表
    const { data: ipGroups } = await API.get(`/api/ip-groups`);
    console.log("ipGroups", ipGroups);
    if (ipGroups) {
      const ipGroupSelect = document.getElementById("ip_group_id");
      ipGroupSelect.innerHTML = "";
      ipGroups.forEach((d) => {
        const opt = document.createElement("option");
        opt.value = d.id;
        opt.text = `${d.name}`;
        ipGroupSelect.appendChild(opt);
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
      const { data: cdnRecord } = await API.get(`/api/cdn-records/${id}`);
      console.log("cdnRecord", cdnRecord);
      if (!cdnRecord) {
        alert("CDN记录数据异常");
        return;
      }

      document.getElementById("id").value = cdnRecord.id;
      document.getElementById("id").disabled = true;
      document.getElementById("platform").value = cdnRecord.platform;
      document.getElementById("platform").disabled = true;
      document.getElementById("config_id").value = cdnRecord.config_id;
      document.getElementById("config_id").disabled = true;
      document.getElementById("sub_domain").value = cdnRecord.sub_domain;
      document.getElementById("sub_domain").disabled = true;
      document.getElementById("domain_name").innerHTML =
        "." + cdnRecord.domain_name;
      document.getElementById("type").value = cdnRecord.type;
      document.getElementById("ip_group_id").value = cdnRecord.ip_group_id;

      if (cdnRecord.webhooks && cdnRecord.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", cdnRecord.webhooks);
        selectId2.val(cdnRecord.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", cdnRecord.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 渲染字段，回填数据
      renderCdnFields(cdnRecord.platform, cdnRecord.config_json);

      showModal("formModal", "modalTitle", "编辑 CDN记录");
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
          disabled: true,
        },
        {
          id: "domain_name",
          value: "",
          disabled: true,
        },
        {
          id: "config_id",
          disabled: false,
        },
        {
          id: "sub_domain",
          value: "",
          disabled: false,
        },
      ]);

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      // 默认渲染字段
      renderCdnFields(document.getElementById("platform").value.trim());

      showModal("formModal", "modalTitle", "添加 CDN记录");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开CDN记录模态框失败", error);
    return;
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 自动从动态字段中组装 JSON
    buildConfigFromCdnFields(document.getElementById("platform").value.trim());

    const selectedWebhookIDs = $("#webhook_ids").val() || [];

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      config_id: Number(document.getElementById("config_id").value.trim()),
      sub_domain: document.getElementById("sub_domain").value.trim(),
      type: document.getElementById("type").value.trim(),
      ip_group_id: Number(document.getElementById("ip_group_id").value.trim()),
      webhook_ids: selectedWebhookIDs.map(Number),
      config_json: document.getElementById("config_json").value.trim(),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
      return alert("请选择配置");
    }
    if (!form_data.sub_domain) {
      return alert("请输入子域名");
    }
    if (!form_data.type) {
      return alert("请选择记录类型");
    }
    if (isNaN(form_data.ip_group_id) || form_data.ip_group_id <= 0) {
      return alert("请选择IP组");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/cdn-records", form_data);
      } else {
        await API.patch(`/api/cdn-records/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("CDN记录操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 删除
async function onDelete(id) {
  if (confirm("确定删除此记录？")) {
    try {
      await API.delete(`/api/cdn-records/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除CDN记录失败", error);
    }
  }
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: cdnRecordHealthcheck } = await API.get(
      `/api/cdn-records/${id}/health-check`
    );
    console.log("cdnRecordHealthcheck", cdnRecordHealthcheck);
    if (!cdnRecordHealthcheck) {
      alert("CDN记录检测数据异常");
      return;
    }
    if (cdnRecordHealthcheck.length === 0) {
      alert("当前记录无检测数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 检测 域名/IP`,
      formatHealthCheckResult(cdnRecordHealthcheck),
      cdnRecordHealthcheck.length,
      { resetClosedState: true }
    );
  } catch (error) {
    alert(`加载CDN记录检测信息失败: ${error.message}`);
    console.error("查看CDN记录检测信息失败", error);
  }
}

// 打开 实时IP 弹窗
async function onLiveIPPopup(id, name) {
  try {
    const { data: ipGroupDevices } = await API.get(
      `/api/ip-groups/${id}/devices`
    );
    console.log("ipGroupDevices", ipGroupDevices);
    if (!ipGroupDevices) {
      alert("IP组设备数据异常");
      return;
    }
    if (ipGroupDevices.length === 0) {
      alert("当前组内设备无数据");
      return;
    }

    // 格式化 IP 列表
    const preList = ipGroupDevices.map(
      (d) => `${d.name} - ${d.ip_v4 || "-"} / ${d.ip_v6 || "-"}`
    );

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 设备 IP 地址 - ${name}`,
      preList.join("\n"),
      ipGroupDevices.length
    );
  } catch (error) {
    alert(`加载IP组设备信息失败: ${error.message}`);
    console.error("查看设备IP地址失败", error);
  }
}
