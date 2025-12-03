// 选择配置 config_id
function onConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onConfigSelect", opt);
}

// 选择类型 type
function onTypeSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onTypeSelect", opt);
}

// 选择回源协议 origin_protocol
function onOriginProtocolSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onOriginProtocolSelect", opt);
}

// 选择IP组 ip_group_id
function onIpGroupSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onIpGroupSelect", opt);
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
        if (d.capability === "cdn" || d.capability === "all") {
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
      ipGroups.forEach((d) => {
        const opt = document.createElement("option");
        opt.value = d.id;
        opt.text = `${d.name}`;
        ipGroupSelect.appendChild(opt);
      });
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开CDN记录模态框失败", error);
    return;
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
    document.getElementById("config_id").value = cdnRecord.config_id;
    document.getElementById("config_id").disabled = true;
    document.getElementById("domain").value = cdnRecord.domain;
    document.getElementById("domain").disabled = true;
    document.getElementById("type").value = cdnRecord.type;
    document.getElementById("type").disabled = true;
    document.getElementById("origin_protocol").value =
      cdnRecord.origin_protocol;
    document.getElementById("http_port").value = cdnRecord.http_port;
    document.getElementById("https_port").value = cdnRecord.https_port;
    document.getElementById("ip_group_id").value = cdnRecord.ip_group_id;

    showModal("modal", "modalTitle", "编辑 CDN记录");
  }

  // 创建
  if (!id) {
    document.getElementById("id").value = 0;
    document.getElementById("id").disabled = true;
    document.getElementById("config_id").disabled = false;
    document.getElementById("domain").disabled = false;
    document.getElementById("type").disabled = false;
    document.getElementById("http_port").value = 80;
    document.getElementById("https_port").value = 443;

    showModal("modal", "modalTitle", "添加 CDN记录");
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
    origin_protocol: document.getElementById("origin_protocol").value.trim(),
    http_port: Number(document.getElementById("http_port").value.trim()),
    https_port: Number(document.getElementById("https_port").value.trim()),
    ip_group_id: Number(document.getElementById("ip_group_id").value.trim()),
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
  try {
    const { data: cdnRecordHealthcheck } = await API.get(
      `/api/cdn-records/${id}/healthcheck`
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

    // 填充代码块
    const pre = document.getElementById("popupPre");

    // 格式化 检测数据
    pre.textContent = formatHealthCheckResult(cdnRecordHealthcheck);

    // 更新计数
    document.getElementById("popupCount").textContent =
      cdnRecordHealthcheck.length;

    // 打开弹窗（统一走你已有的 showModal 即可）
    showModal("popup", "popupTitle", `实时 检测 域名/IP`);
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

    // 填充代码块
    const pre = document.getElementById("popupPre");

    // 格式化 IP 列表
    const preList = ipGroupDevices.map(
      (d) => `${d.name} - ${d.ip_v4 || "-"} / ${d.ip_v6 || "-"}`
    );

    pre.textContent = preList.join("\n");

    // 更新计数
    document.getElementById("popupCount").textContent = ipGroupDevices.length;

    // 打开弹窗（统一走你已有的 showModal 即可）
    showModal("popup", "popupTitle", `实时 设备 IP 地址 - ${name}`);
  } catch (error) {
    alert(`加载IP组设备信息失败: ${error.message}`);
    console.error("查看设备IP地址失败", error);
  }
}
