// 选择配置 config_id
function onConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onConfigSelect", opt);
}

// 选择类型 type
function onTypeSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onTypeSelect", opt);
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

// 选择IP组 ip_group_id
function onIpGroupSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onIpGroupSelect", opt);
}

// 选择CDN记录 cdn_record_id
function onCdnRecordSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordSelect", opt);
}

// 打开 表单 弹窗
async function openFormModal(id) {
  try {
    // 平台配置列表
    const { data: configs } = await API.get(`/api/configs`);
    console.log("configs", configs);
    if (configs) {
      const configSelect = document.getElementById("config_id");
      configSelect.innerHTML = "";
      addPlaceholder(configSelect, "请选择平台配置");
      configs.forEach((d) => {
        if (d.capability === "dns" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.value = d.id;
          opt.text = `${d.name}`;
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
        opt.value = d.id;
        opt.text = `${d.name}`;
        ipGroupSelect.appendChild(opt);
      });
    }

    // CDN列表
    const { data: cdnRecords } = await API.get(`/api/cdn-records`);
    console.log("cdnRecords", cdnRecords);
    if (cdnRecords) {
      const cdnRecordSelect = document.getElementById("cdn_record_id");
      cdnRecordSelect.innerHTML = "";
      addPlaceholder(cdnRecordSelect, "请选择CDN记录");
      cdnRecords.forEach((d) => {
        const opt = document.createElement("option");
        opt.value = d.id;
        opt.text = `${d.domain}`;
        cdnRecordSelect.appendChild(opt);
      });
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开DNS记录模态框失败", error);
    return;
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
    document.getElementById("config_id").value = dnsRecord.config_id;
    document.getElementById("config_id").disabled = true;
    document.getElementById("domain").value = dnsRecord.domain;
    document.getElementById("domain").disabled = true;
    document.getElementById("type").value = dnsRecord.type;
    document.getElementById("type").disabled = true;

    if (dnsRecord.type === "CNAME") {
      document.getElementById("cdn_record_id").value = dnsRecord.cdn_record_id;
      document.getElementById("cdn_record_id").disabled = false;
    } else {
      document.getElementById("ip_group_id").value = dnsRecord.ip_group_id;
      document.getElementById("ip_group_id").disabled = false;
    }

    showModal("modal", "modalTitle", "编辑 DNS记录");
  }

  // 创建
  if (!id) {
    document.getElementById("id").value = 0;
    document.getElementById("id").disabled = true;
    document.getElementById("config_id").disabled = false;
    document.getElementById("domain").disabled = false;
    document.getElementById("type").disabled = false;
    document.getElementById("cdn_record_id").disabled = true;
    document.getElementById("ip_group_id").disabled = true;

    showModal("modal", "modalTitle", "添加 DNS记录");
  }
}

// 表单提交
document.getElementById("form").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("id").value.trim());
  const form_data = {
    config_id: Number(document.getElementById("config_id").value.trim()),
    domain: document.getElementById("domain").value.trim(),
    type: document.getElementById("type").value.trim(),
    ip_group_id: Number(document.getElementById("ip_group_id").value.trim()),
    cdn_record_id: Number(
      document.getElementById("cdn_record_id").value.trim()
    ),
  };
  console.log("form_id, form_data", form_id, form_data);

  // 校验
  if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
    return alert("请选择配置");
  }
  if (!form_data.domain) {
    return alert("请输入域名");
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

// 删除
async function onDelete(id) {
  if (confirm("确定删除此记录？")) {
    try {
      await API.delete(`/api/dns-records/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除DNS记录失败", error);
    }
  }
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: dnsRecordHealthcheck } = await API.get(
      `/api/dns-records/${id}/healthcheck`
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
      { resetClosedState: true }
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
