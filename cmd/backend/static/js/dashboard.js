// 打开 表单 弹窗
async function openSettingFormModal() {
  document.getElementById("settingForm").reset();

  try {
    // 编辑
    const { data: setting } = await API.get(`/api/settings/`);
    console.log("setting", setting);
    if (!setting) {
      alert("设置数据异常");
      return;
    }

    document.getElementById("device_offline_invalid_time").value =
      setting.device_offline_invalid_time;
    document.getElementById("ip_report_interval_time").value =
      setting.ip_report_interval_time;
    document.getElementById("port_forward_reload_time").value =
      setting.port_forward_reload_time;
    document.getElementById("port_forward_report_time").value =
      setting.port_forward_report_time;

    showModal("settingFormModal", "modalTitle", "编辑 设置");
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开模态框失败", error);
  }
}

// 表单提交
const form = document.getElementById("settingFormModal");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_data = {
      device_offline_invalid_time: Number(
        document.getElementById("device_offline_invalid_time").value.trim(),
      ),
      ip_report_interval_time: document
        .getElementById("ip_report_interval_time")
        .value.trim(),
      port_forward_reload_time: document
        .getElementById("port_forward_reload_time")
        .value.trim(),
      port_forward_report_time: document
        .getElementById("port_forward_report_time")
        .value.trim(),
    };
    console.log(" form_data", form_data);

    // 校验
    if (!form_data.ip_report_interval_time) {
      return alert("请输入IP上报间隔时间");
    }

    try {
      // 提交
      await API.patch(`/api/settings/`, form_data);

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("设置操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}
