// 状态
async function toggleDnsRecordStatus(id, checkbox) {
  try {
    await API.patch(`/api/dns-records/${id}/active`, {
      is_active: checkbox.checked,
    });
  } catch (error) {
    alert(`状态更新失败: ${error.message}`);
    console.error("切换DNS记录状态失败", error);
    // 恢复 checkbox 状态
    checkbox.checked = !checkbox.checked;
  }
}

// DNS记录管理（模态框）
async function openDnsRecordModal(id) {
  try {
    // 平台配置列表
    const { data: configs } = await API.get(`/api/configs`);
    console.log("configs", configs);
    if (configs) {
      const configSelect = document.getElementById("dns_record_config_id");
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
      const ipGroupSelect = document.getElementById("dns_record_ip_group_id");
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
      const cdnRecordSelect = document.getElementById(
        "dns_record_cdn_record_id"
      );
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

    document.getElementById("dns_record_id").value = dnsRecord.id;
    document.getElementById("dns_record_id").disabled = true;
    document.getElementById("dns_record_config_id").value = dnsRecord.config_id;
    document.getElementById("dns_record_config_id").disabled = true;
    document.getElementById("dns_record_domain").value = dnsRecord.domain;
    document.getElementById("dns_record_domain").disabled = true;
    document.getElementById("dns_record_type").value = dnsRecord.type;
    document.getElementById("dns_record_type").disabled = true;

    if (dnsRecord.type === "CNAME") {
      document.getElementById("dns_record_cdn_record_id").value =
        dnsRecord.cdn_record_id;
      document.getElementById("dns_record_cdn_record_id").disabled = false;
    } else {
      document.getElementById("dns_record_ip_group_id").value =
        dnsRecord.ip_group_id;
      document.getElementById("dns_record_ip_group_id").disabled = false;
    }

    openModal("dnsRecordModal", "dnsRecordModalTitle", "编辑 DNS记录");
  }

  // 创建
  if (!id) {
    document.getElementById("dns_record_id").value = 0;
    document.getElementById("dns_record_id").disabled = true;
    document.getElementById("dns_record_config_id").disabled = false;
    document.getElementById("dns_record_domain").disabled = false;
    document.getElementById("dns_record_type").disabled = false;
    document.getElementById("dns_record_cdn_record_id").disabled = true;
    document.getElementById("dns_record_ip_group_id").disabled = true;

    openModal("dnsRecordModal", "dnsRecordModalTitle", "添加 DNS记录");
  }
}

// 选择配置 config_id
function onDnsRecordConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onCdnRecordIponDnsRecordConfigSelectGroupSelect", opt);
}

// 选择类型 type
function onDnsRecordTypeSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onDnsRecordTypeSelect", opt);
  if (opt.value === "CNAME") {
    document.getElementById("dns_record_ip_group_id").disabled = true;
    document.getElementById("dns_record_cdn_record_id").disabled = false;
    document.getElementById("dns_record_ip_group_id").value = "";
  } else {
    document.getElementById("dns_record_ip_group_id").disabled = false;
    document.getElementById("dns_record_cdn_record_id").disabled = true;
    document.getElementById("dns_record_cdn_record_id").value = "";
  }
}

// 选择IP组 ip_group_id
function onDnsRecordIpGroupSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onDnsRecordIpGroupSelect", opt);
}

// 选择CDN记录 cdn_record_id
function onDnsRecordCdnRecordSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onDnsRecordCdnRecordSelect", opt);
}

// DNS记录管理（模态框） -> 表单提交
document.getElementById("dnsRecordForm").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("dns_record_id").value.trim());
  const form_data = {
    config_id: Number(
      document.getElementById("dns_record_config_id").value.trim()
    ),
    domain: document.getElementById("dns_record_domain").value.trim(),
    type: document.getElementById("dns_record_type").value.trim(),
    ip_group_id: Number(
      document.getElementById("dns_record_ip_group_id").value.trim()
    ),
    cdn_record_id: Number(
      document.getElementById("dns_record_cdn_record_id").value.trim()
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
async function deleteDnsRecord(id) {
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
