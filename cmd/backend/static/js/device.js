// 管理设备（模态框）
async function openDeviceModal(id) {
  document.getElementById("deviceForm").reset();

  // 编辑
  if (id) {
    try {
      const { data: device } = await API.get(`/api/devices/${id}`);
      console.log("device", device);
      if (!device) {
        alert("设备数据异常");
        return;
      }

      document.getElementById("device_id").value = device.id;
      document.getElementById("device_id").disabled = true;
      document.getElementById("device_code").value = device.code;
      document.getElementById("device_code").disabled = true;
      document.getElementById("device_name").value = device.name;

      openModal("deviceModal", "deviceModalTitle", "编辑 设备");
    } catch (error) {
      alert(`加载设备信息失败: ${error.message}`);
      console.error("打开设备模态框失败", error);
    }
  }

  // 创建
  if (!id) {
    document.getElementById("device_id").value = "";
    document.getElementById("device_id").disabled = true;

    const deviceCode = uuid.v7();
    if (!deviceCode) {
      alert("生成设备编码失败");
      return;
    }
    document.getElementById("device_code").value = deviceCode;
    document.getElementById("device_code").disabled = false;

    openModal("deviceModal", "deviceModalTitle", "添加 设备");
  }
}

// 管理设备（模态框） -> 表单提交
document.getElementById("deviceForm").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("device_id").value.trim());
  const form_data = {
    code: document.getElementById("device_code").value.trim(),
    name: document.getElementById("device_name").value.trim(),
  };
  console.log("form_id, form_data", form_id, form_data);

  // 校验
  if (!form_data.code) {
    return alert("请输入设备编码");
  }
  if (!form_data.name) {
    return alert("请输入设备名称");
  }

  try {
    // 提交
    if (isNaN(form_id) || form_id <= 0) {
      await API.post("/api/devices", form_data);
    } else {
      await API.patch(`/api/devices/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
  } catch (error) {
    alert(`操作失败: ${error.message}`);
    console.error("设备操作失败", error);
  }
};

// 删除
async function deleteDevice(id) {
  if (confirm("确定删除此设备？")) {
    try {
      await API.delete(`/api/devices/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除设备失败", error);
    }
  }
}

// 实时 设备 IP 历史 弹窗
async function viewLiveDeviceIpHistory(id, name) {
  try {
    const { data: deviceIpHistory } = await API.get(
      `/api/devices/${id}/ip-history`
    );
    console.log("deviceIpHistory", deviceIpHistory);
    if (!deviceIpHistory) {
      alert("设备数据异常");
      return;
    }
    if (deviceIpHistory.length === 0) {
      alert("设备无历史IP数据");
      return;
    }

    // 填充代码块
    const pre = document.getElementById("liveDeviceIpHistoryPre");

    // 格式化 IP 列表
    const ipList = deviceIpHistory.map(
      (d) => `${fmtDate(d.created_at)} - ${d.ip_v4 || "-"} / ${d.ip_v6 || "-"}`
    );

    pre.textContent = ipList.join("\n");

    // 更新计数
    document.getElementById("liveDeviceIpHistoryCount").textContent =
      deviceIpHistory.length;

    // 打开弹窗（统一走你已有的 openModal 即可）
    openModal(
      "liveDeviceIpHistoryModal",
      "liveDeviceIpHistoryTitle",
      `实时 设备 IP 历史 - ${name}`
    );
  } catch (error) {
    alert(`加载设备IP历史失败: ${error.message}`);
    console.error("查看设备IP历史失败", error);
  }
}

document.addEventListener("DOMContentLoaded", function () {
  const now = Math.floor(Date.now() / 1000);
  // console.log("当前时间戳（秒）:", now);
  const tenMinutes = 12 * 60;

  document
    .querySelectorAll("#deviceTable tr[data-last-heartbeat]")
    .forEach((row) => {
      const heartbeatAt = parseInt(row.getAttribute("data-last-heartbeat"));
      // console.log("设备心跳时间戳:", heartbeatAt, "差值:", now - heartbeatAt);

      // 如果是零值（从未心跳），heartbeatAt <= 0
      if (heartbeatAt <= 0) {
        row.classList.add("never-heartbeat");
        return;
      }

      if (now - heartbeatAt > tenMinutes) {
        row.classList.add("offline");
      } else {
        row.classList.remove("offline");
      }
    });
});
