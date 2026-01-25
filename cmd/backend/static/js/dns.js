// 选择 配置 config_id
async function onConfigSelect(el) {
  console.log("onConfigSelect", el);

  const opt = el.options[el.selectedIndex];

  const id = opt.value || 0;
  const platform = opt.dataset.platform || "";
  const domain_name = opt.dataset.domain_name || "";

  if (platform) {
    document.getElementById("platform").value = platform;
  }
  if (domain_name) {
    document.getElementById("domain_name").innerHTML = "." + domain_name;
  }

  // 从后端接口获取指定平台的动态 Schema，并更新 DNS_SCHEMA
  switch (platform) {
    case "tencent_dns":
      await fetchAndReplaceDnsSchema(platform, id);
      break;
    case "aliyun_dns":
      await fetchAndReplaceDnsSchema(platform, id);
      break;
    case "volcengine_dns":
      await fetchAndReplaceDnsSchema(platform, id);
      break;
  }

  // 重新渲染字段
  renderDnsFields(platform);
}

// 选择 记录类型 type
function onTypeSelect(el) {
  console.log("onTypeSelect", el);

  const opt = el.options[el.selectedIndex];
  if (opt.value === "CNAME") {
    document.getElementById("ip_group_id").disabled = true;
    document.getElementById("cdn_record_id").disabled = false;
    document.getElementById("ip_group_id").value = "";
  } else {
    document.getElementById("ip_group_id").disabled = false;
    document.getElementById("cdn_record_id").disabled = true;
    document.getElementById("cdn_record_id").value = "";
  }
}

// 选择 IP 组 ip_group_id
function onIpGroupSelect(el) {
  console.log("onIpGroupSelect", el);

  const opt = el.options[sel.selectedIndex];

  const code = opt.dataset.code || "";
  if (code) {
    document.getElementById("ip_group_code").value = code;
  }
}

// 选择 CDN记录 cdn_record_id
function onCdnRecordSelect(el) {
  console.log("onCdnRecordSelect", el);

  const opt = el.options[el.selectedIndex];
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
        if (d.capability === "dns" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.id = d.id;
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json?.domain_name ?? "";
          opt.dataset.email = d.config_json?.email ?? "";
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
      addPlaceholder(ipGroupSelect, "请选择IP组");
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

    // CDN列表
    const { data: cdnRecords } = await API.get(`/api/cdn-records`, {
      is_active: true,
    });
    console.log("cdnRecords", cdnRecords);
    if (cdnRecords) {
      const cdnRecordSelect = document.getElementById("cdn_record_id");
      cdnRecordSelect.innerHTML = "";
      addPlaceholder(cdnRecordSelect, "请选择CDN记录");
      cdnRecords.forEach((d) => {
        const opt = document.createElement("option");
        opt.value = d.id;
        opt.text = `${d.sub_domain}.${d.domain_name}`;
        cdnRecordSelect.appendChild(opt);
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
      const { data: dnsRecord } = await API.get(`/api/dns-records/${id}`);
      console.log("dnsRecord", dnsRecord);
      if (!dnsRecord) {
        alert("DNS记录数据异常");
        return;
      }

      document.getElementById("id").value = dnsRecord.id;
      document.getElementById("id").disabled = true;
      document.getElementById("platform").value = dnsRecord.platform;
      document.getElementById("platform").disabled = true;
      document.getElementById("ip_group_code").value =
        dnsRecord.config_json?.ip_group_code ?? "";
      document.getElementById("ip_group_code").disabled = true;
      document.getElementById("config_id").value = dnsRecord.config_id;
      document.getElementById("config_id").disabled = true;
      document.getElementById("sub_domain").value = dnsRecord.sub_domain;
      document.getElementById("sub_domain").disabled = true;
      document.getElementById("domain_name").innerHTML =
        "." + dnsRecord.domain_name;
      document.getElementById("type").value = dnsRecord.type;

      if (dnsRecord.type === "CNAME") {
        document.getElementById("cdn_record_id").value =
          dnsRecord.cdn_record_id;
        document.getElementById("cdn_record_id").disabled = false;
      } else {
        document.getElementById("ip_group_id").value = dnsRecord.ip_group_id;
        document.getElementById("ip_group_id").disabled = false;
      }

      if (dnsRecord.webhooks && dnsRecord.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", dnsRecord.webhooks);
        selectId2.val(dnsRecord.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", dnsRecord.webhooks);
        selectId2.val(null).trigger("change");
      }

      // 从后端接口获取指定平台的动态 Schema，并更新 DNS_SCHEMA
      switch (dnsRecord.platform) {
        case "tencent_dns":
          await fetchAndReplaceDnsSchema(
            dnsRecord.platform,
            dnsRecord.config_id,
          );
          break;
        case "aliyun_dns":
          await fetchAndReplaceDnsSchema(
            dnsRecord.platform,
            dnsRecord.config_id,
          );
          break;
        case "volcengine_dns":
          await fetchAndReplaceDnsSchema(
            dnsRecord.platform,
            dnsRecord.config_id,
          );
          break;
      }

      // 渲染字段，回填数据
      document.getElementById("original_config_json").value = JSON.stringify(
        dnsRecord.config_json || {},
      );
      renderDnsFields(dnsRecord.platform, dnsRecord.config_json || {});

      // 如果是复制操作，清空 id 字段，取消 sub_domain 禁用
      if (action === "copy") {
        document.getElementById("id").value = "";
        document.getElementById("id").disabled = false;
        document.getElementById("sub_domain").disabled = false;

        showModal("formModal", "modalTitle", "复制 DNS记录");
      } else {
        showModal("formModal", "modalTitle", "编辑 DNS记录");
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
          disabled: true,
        },
        {
          id: "ip_group_code",
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
        {
          id: "cdn_record_id",
          value: 0,
          disabled: true,
        },
        {
          id: "ip_group_id",
          value: 0,
          disabled: true,
        },
      ]);

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      // 默认渲染字段
      renderDnsFields(document.getElementById("platform").value.trim() || "");

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
    e.preventDefault();

    // 自动从动态字段中组装 JSON
    buildConfigFromDnsFields(document.getElementById("platform").value.trim());

    const selectedWebhookIDs = $("#webhook_ids").val() || [];

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      config_id: Number(document.getElementById("config_id").value.trim()),
      sub_domain: document.getElementById("sub_domain").value.trim(),
      type: document.getElementById("type").value.trim(),
      ip_group_id: Number(document.getElementById("ip_group_id").value.trim()),
      cdn_record_id: Number(
        document.getElementById("cdn_record_id").value.trim(),
      ),
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
    if (form_data.type === "CNAME") {
      if (isNaN(form_data.cdn_record_id) || form_data.cdn_record_id <= 0) {
        return alert("请选择CDN记录");
      }
    } else {
      if (isNaN(form_data.ip_group_id) || form_data.ip_group_id <= 0) {
        return alert("请选择IP组");
      }
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/dns-records", form_data);
      } else {
        await API.patch(`/api/dns-records/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("DNS记录操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: dnsRecordHealthcheck } = await API.get(
      `/api/dns-records/${id}/health-check`,
    );
    console.log("dnsRecordHealthcheck", dnsRecordHealthcheck);
    if (!dnsRecordHealthcheck) {
      alert("DNS记录检测数据异常");
      return;
    }
    if (dnsRecordHealthcheck.length === 0) {
      alert("当前记录无检测数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 检测 域名/IP`,
      formatHealthCheckResult(dnsRecordHealthcheck),
      dnsRecordHealthcheck.length,
      { resetClosedState: true },
    );
  } catch (error) {
    alert(`加载DNS记录检测信息失败: ${error.message}`);
    console.error("查看DNS记录检测信息失败", error);
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
