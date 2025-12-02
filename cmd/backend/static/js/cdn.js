// 状态
async function toggleCdnRecordStatus(id, checkbox) {
  try {
    await API.patch(`/api/cdn-records/${id}/active`, {
      is_active: checkbox.checked,
    });
  } catch (error) {
    alert(`状态更新失败: ${error.message}`);
    console.error("切换CDN记录状态失败", error);
    // 恢复 checkbox 状态
    checkbox.checked = !checkbox.checked;
  }
}

// CDN记录管理
async function openCdnRecordModal(id) {
  try {
    // 平台配置列表
    const { data: configs } = await API.get(`/api/configs`);
    console.log("configs", configs);
    if (configs) {
      const configSelect = document.getElementById("cdn_record_config_id");
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
      const ipGroupSelect = document.getElementById("cdn_record_ip_group_id");
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

    document.getElementById("cdn_record_id").value = cdnRecord.id;
    document.getElementById("cdn_record_id").disabled = true;
    document.getElementById("cdn_record_config_id").value = cdnRecord.config_id;
    document.getElementById("cdn_record_config_id").disabled = true;
    document.getElementById("cdn_record_domain").value = cdnRecord.domain;
    document.getElementById("cdn_record_domain").disabled = true;
    document.getElementById("cdn_record_type").value = cdnRecord.type;
    document.getElementById("cdn_record_type").disabled = true;
    document.getElementById("cdn_record_origin_protocol").value =
      cdnRecord.origin_protocol;
    document.getElementById("cdn_record_http_port").value = cdnRecord.http_port;
    document.getElementById("cdn_record_https_port").value =
      cdnRecord.https_port;
    document.getElementById("cdn_record_ip_group_id").value =
      cdnRecord.ip_group_id;

    openModal("cdnRecordModal", "cdnRecordModalTitle", "编辑 CDN记录");
  }

  // 创建
  if (!id) {
    document.getElementById("cdn_record_id").value = 0;
    document.getElementById("cdn_record_id").disabled = true;
    document.getElementById("cdn_record_config_id").disabled = false;
    document.getElementById("cdn_record_domain").disabled = false;
    document.getElementById("cdn_record_type").disabled = false;
    document.getElementById("cdn_record_http_port").value = 80;
    document.getElementById("cdn_record_https_port").value = 443;

    openModal("cdnRecordModal", "cdnRecordModalTitle", "添加 CDN记录");
  }
}

// 选择配置 config_id
function onCdnRecordConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordConfigSelect", opt);
}

// 选择类型 type
function onCdnRecordTypeSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordTypeSelect", opt);
}

// 选择回源协议 origin_protocol
function onCdnRecordOriginProtocolSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordOriginProtocolSelect", opt);
}

// 选择IP组 ip_group_id
function onCdnRecordIpGroupSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordIpGroupSelect", opt);
}

// CDN记录管理 -> 表单提交
document.getElementById("cdnRecordForm").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("cdn_record_id").value.trim());
  const form_data = {
    config_id: Number(
      document.getElementById("cdn_record_config_id").value.trim()
    ),
    domain: document.getElementById("cdn_record_domain").value.trim(),
    type: document.getElementById("cdn_record_type").value.trim(),
    origin_protocol: document
      .getElementById("cdn_record_origin_protocol")
      .value.trim(),
    http_port: Number(
      document.getElementById("cdn_record_http_port").value.trim()
    ),
    https_port: Number(
      document.getElementById("cdn_record_https_port").value.trim()
    ),
    ip_group_id: Number(
      document.getElementById("cdn_record_ip_group_id").value.trim()
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
async function deleteCdnRecord(id) {
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

// 实时 设备 IP 地址 弹窗
async function viewLiveDeviceIpAddresse(id, name) {
    try {
        const { data: ipGroupDevices } = await API.get(`/api/ip-groups/${id}/devices`);
        console.log("ipGroupDevices", ipGroupDevices);
        if (!ipGroupDevices) {
            alert('IP组设备数据异常');
            return;
        }
        if (ipGroupDevices.length === 0) {
            alert('当前组内设备无数据');
            return;
        }

        // 填充代码块
        const pre = document.getElementById('liveDeviceIpAddressePre');

        // 格式化 IP 列表
        const ipList = ipGroupDevices.map(d => `${d.name} - ${d.ip_v4 || '-'} / ${d.ip_v6 || '-'}`);

        pre.textContent = ipList.join('\n');

        // 更新计数
        document.getElementById('liveDeviceIpAddresseCount').textContent = ipGroupDevices.length;

        // 打开弹窗（统一走你已有的 openModal 即可）
        openModal('liveDeviceIpAddresseModal', 'liveDeviceIpAddresseTitle', `实时 设备 IP 地址 - ${name}`);
    } catch (error) {
        alert(`加载IP组设备信息失败: ${error.message}`);
        console.error("查看设备IP地址失败", error);
    }
}
