// 选择 配置 config_id
async function onConfigSelect(el) {
  console.log("onConfigSelect", el);

  const opt = el.options[el.selectedIndex];

  const id = opt.value || 0;
  const platform = opt.dataset.platform || "";

  // 从后端接口获取指定平台的动态 Schema
  if (id > 0 && platform) {
    switch (platform) {
      case "tencent_eo":
        await fetchAndReplaceCdnSchema(platform, id);
        break;
      case "aliyun_esa":
        await fetchAndReplaceCdnSchema(platform, id);
        break;
    }
  }

  // 重新渲染字段
  renderCdnFields(platform);
}

// 选择 记录类型 type
function onTypeSelect(el) {
  console.log("onTypeSelect", el);

  const opt = el.options[el.selectedIndex];
}

// 选择 IP 组 ip_group_id
function onIpGroupSelect(el) {
  console.log("onIpGroupSelect", el);

  const opt = el.options[el.selectedIndex];
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
        if (d.capability === "cdn" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.id = d.id;
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json?.domain_name ?? "";
          opt.dataset.email = d.config_json?.email ?? "";
          opt.dataset.account_id = d.config_json?.account_id ?? "";
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
      const ipGroupSelect = els.ip_group_id;
      ipGroupSelect.innerHTML = "";
      ipGroups.forEach((d) => {
        const opt = document.createElement("option");
        opt.dataset.id = d.id;
        opt.dataset.code = d.code;
        opt.dataset.name = d.name;
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

      els.id.value = cdnRecord.id;
      els.id.disabled = true;
      els.platform.value = cdnRecord.platform;
      els.platform.disabled = true;
      els.config_id.value = cdnRecord.config_id;
      els.config_id.disabled = true;
      els.sub_domain.value = cdnRecord.sub_domain;
      els.sub_domain.disabled = true;
      els.domain_name.value = cdnRecord.domain_name;
      els.domain_name.disabled = true;
      els.domain_name.dataset.mode = "edit"; // 标记为编辑模式
      els.type.value = cdnRecord.type;
      els.ip_group_id.value = cdnRecord.ip_group_id;

      if (cdnRecord.webhooks && cdnRecord.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", cdnRecord.webhooks);
        selectId2.val(cdnRecord.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", cdnRecord.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 从后端接口获取指定平台的动态 Schema
      const disabledKeys = [];
      switch (cdnRecord.platform) {
        case "tencent_eo":
          await fetchAndReplaceCdnSchema(
            cdnRecord.platform,
            cdnRecord.config_id,
          );
          disabledKeys.push("zone");
          break;
        case "aliyun_esa":
          await fetchAndReplaceCdnSchema(
            cdnRecord.platform,
            cdnRecord.config_id,
          );
          disabledKeys.push("site");
          break;
        default:
          els.domain_name.disabled = true;
          break;
      }

      // 渲染字段，回填数据
      els.original_config_json.value = JSON.stringify(
        cdnRecord.config_json || {},
      );
      renderCdnFields(
        cdnRecord.platform,
        cdnRecord.config_json || {},
        disabledKeys,
      );

      // 如果是复制操作，清空 id 字段，取消 sub_domain 禁用
      if (action === "copy") {
        els.id.value = "";
        els.id.disabled = false;
        els.sub_domain.disabled = false;

        showModal("formModal", "modalTitle", "复制 CDN 记录");
      } else {
        showModal("formModal", "modalTitle", "编辑 CDN 记录");
      }
    }

    // 创建
    if (!id) {
      els.id.value = 0;
      els.id.disabled = true;
      els.platform.value = "";
      els.platform.disabled = true;
      els.config_id.disabled = false;
      els.sub_domain.value = "";
      els.sub_domain.disabled = false;
      els.domain_name.value = "";
      els.domain_name.disabled = false;

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      // 默认渲染字段
      renderCdnFields(els.platform.value.trim() || "");

      // 显示域名名称标签
      // const domain_name_label = document.getElementById("domain_name_label");
      // if (domain_name_label) {
      //   domain_name_label.style.display = "block";
      // }

      showModal("formModal", "modalTitle", "添加 CDN 记录");
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
      buildConfigFromCdnFields(els.platform.value.trim());

      const selectedWebhookIDs = $("#webhook_ids").val() || [];

      // 表单数据
      const form_id = Number(els.id.value.trim());
      const form_data = {
        config_id: Number(els.config_id.value.trim()),
        sub_domain: els.sub_domain.value.trim(),
        domain_name: els.domain_name.value.trim(),
        type: els.type.value.trim(),
        ip_group_id: Number(els.ip_group_id.value.trim()),
        webhook_ids: selectedWebhookIDs.map(Number),
        config_json: els.config_json.value.trim(),
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

      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/cdn-records", form_data);
      } else {
        await API.patch(`/api/cdn-records/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error(err);
    }
  };
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: cdnRecordHealthcheck } = await API.get(
      `/api/cdn-records/${id}/health-check`,
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
      { resetClosedState: true },
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
      `/api/ip-groups/${id}/devices`,
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
      (d) => `${d.name} - ${d.ip_v4 || "-"} / ${d.ip_v6 || "-"}`,
    );

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 设备 IP 地址 - ${name}`,
      preList.join("\n"),
      ipGroupDevices.length,
    );
  } catch (error) {
    alert(`加载IP组设备信息失败: ${error.message}`);
    console.error("查看设备IP地址失败", error);
  }
}
