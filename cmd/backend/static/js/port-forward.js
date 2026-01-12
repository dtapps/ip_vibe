// 选择协议 protocols
function onProtocolsSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onProtocolsSelect", opt);
}

// 选择设备 device_id
function onDeviceIdSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onDeviceIdSelect", opt);
}

// 打开 表单 弹窗
async function openFormModal(id) {
  document.getElementById("form").reset();

  try {
    // 协议列表
    const protocols = ["tcp4", "tcp6", "udp4", "udp6"];
    const protocolsSelect = document.getElementById("protocols");
    protocolsSelect.innerHTML = "";
    protocols.forEach((d) => {
      const opt = document.createElement("option");
      opt.value = d.toLowerCase(); // 小写
      opt.text = `${d.toUpperCase()}`; // 大写
      protocolsSelect.appendChild(opt);
    });

    // 设备列表
    const { data: devices } = await API.get(`/api/devices`);
    console.log("devices", devices);
    if (devices) {
      const deviceSelect = document.getElementById("device_id");
      deviceSelect.innerHTML = "";
      addPlaceholder(deviceSelect, "请选择设备");
      devices.forEach((d) => {
        const opt = document.createElement("option");
        opt.value = d.id;
        opt.text = `${d.name}`;
        deviceSelect.appendChild(opt);
      });
    }

    // 编辑
    if (id) {
      const { data: portForward } = await API.get(`/api/port-forwards/${id}`);
      console.log("portForward", portForward);
      if (!portForward) {
        alert("端口转发数据异常");
        return;
      }

      document.getElementById("id").value = portForward.id;
      document.getElementById("id").disabled = true;
      document.getElementById("name").value = portForward.name;
      document.getElementById("description").value =
        portForward.description ?? "";
      document.getElementById("local_port").value = portForward.local_port;
      document.getElementById("remote_host").value = portForward.remote_host;
      document.getElementById("remote_port").value = portForward.remote_port;
      document.getElementById("device_id").value = portForward.device_id;

      // 高亮已选
      if (portForward.protocols && portForward.protocols.length > 0) {
        const protocolsSelect = document.getElementById("protocols");
        Array.from(protocolsSelect.options).forEach(
          (opt) => (opt.selected = portForward.protocols.includes(opt.value))
        );
      }

      showModal("formModal", "modalTitle", "编辑 端口转发");
    }

    // 创建
    if (!id) {
      updateFormFields([
        {
          id: "id",
          value: 0,
          disabled: true,
        },
      ]);

      showModal("formModal", "modalTitle", "添加 端口转发");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开IP组模态框失败", error);
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      name: document.getElementById("name").value.trim(),
      description: document.getElementById("description").value.trim(),
      local_port: Number(document.getElementById("local_port").value.trim()),
      remote_host: document.getElementById("remote_host").value.trim(),
      remote_port: Number(document.getElementById("remote_port").value.trim()),
      protocols: Array.from(
        document.getElementById("protocols").selectedOptions
      ).map((opt) => opt.value.trim()),
      device_id: Number(document.getElementById("device_id").value.trim()),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
      return alert("请输入端口转发名");
    }
    if (!form_data.local_port) {
      return alert("请输入本地端口");
    }
    if (!form_data.remote_host) {
      return alert("请输入远程主机");
    }
    if (!form_data.remote_port) {
      return alert("请输入远程端口");
    }
    if (form_data.protocols.length === 0) {
      return alert("请选择协议");
    }
    if (!form_data.device_id) {
      return alert("请选择设备");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/port-forwards", form_data);
      } else {
        await API.patch(`/api/port-forwards/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("端口转发操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 删除
async function onDelete(id) {
  if (confirm("确定删除此 端口转发？")) {
    try {
      await API.delete(`/api/port-forwards/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除端口转发失败", error);
    }
  }
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: portForwardHealthcheck } = await API.get(
      `/api/port-forwards/${id}/health-check`
    );
    console.log("portForwardHealthcheck", portForwardHealthcheck);
    if (!portForwardHealthcheck) {
      alert("端口转发检测数据异常");
      return;
    }
    if (portForwardHealthcheck.length === 0) {
      alert("当前记录无检测数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 检测 IP`,
      formatHealthCheckResult(portForwardHealthcheck),
      portForwardHealthcheck.length,
      { resetClosedState: true }
    );
  } catch (error) {
    alert(`加载端口转发检测信息失败: ${error.message}`);
    console.error("查看端口转发检测信息失败", error);
  }
}
