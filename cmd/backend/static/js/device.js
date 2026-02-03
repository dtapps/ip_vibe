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
      // 设备信息
      const { data: device } = await API.get(`/api/devices/${id}`);
      console.log("device", device);
      if (!device) {
        alert("设备数据异常");
        return;
      }

      document.getElementById("id").value = device.id;
      document.getElementById("id").disabled = true;
      document.getElementById("code").value = device.code;
      document.getElementById("code").disabled = true;
      document.getElementById("name").value = device.name;
      document.getElementById("description").value = device.description ?? "";
      document.getElementById("ip_v4").value = device.ip_v4 ?? "";
      document.getElementById("ip_v6").value = device.ip_v6 ?? "";

      if (device.webhooks && device.webhooks.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", device.webhooks);
        selectId2.val(device.webhooks).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", device.webhooks);
        selectId2.val(null).trigger("change");
      }

      showModal("formModal", "modalTitle", "编辑 设备");
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

      const deviceCode = uuid.v7();
      if (!deviceCode) {
        alert("生成设备编码失败");
        return;
      }
      document.getElementById("code").value = deviceCode;
      document.getElementById("code").disabled = false;

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      showModal("formModal", "modalTitle", "添加 设备");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开设备模态框失败", error);
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    e.preventDefault();

    const selectedWebhookIDs = $("#webhook_ids").val() || [];

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      code: document.getElementById("code").value.trim(),
      name: document.getElementById("name").value.trim(),
      description: document.getElementById("description").value.trim(),
      ip_v4: document.getElementById("ip_v4").value.trim(),
      ip_v6: document.getElementById("ip_v6").value.trim(),
      webhook_ids: selectedWebhookIDs.map(Number),
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
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: deviceHealthcheck } = await API.get(
      `/api/devices/${id}/health-check`,
    );
    console.log("deviceHealthcheck", deviceHealthcheck);
    if (!deviceHealthcheck) {
      alert("接口异常");
      return;
    }
    if (deviceHealthcheck.length === 0) {
      alert("当前设备无检测数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 检测 IP`,
      formatHealthCheckResult(deviceHealthcheck),
      deviceHealthcheck.length,
      { resetClosedState: true },
    );
  } catch (err) {
    alert(`打开 检测 弹窗失败: ${err.message}`);
    console.error(err);
  }
}

// 打开 历史IP 弹窗
async function onHistoryIPPopup(id, name) {
  try {
    const { data: deviceIpHistory } = await API.get(
      `/api/devices/${id}/ip-history`,
    );
    console.log("deviceIpHistory", deviceIpHistory);
    if (!deviceIpHistory) {
      alert("接口异常");
      return;
    }
    if (deviceIpHistory.length === 0) {
      alert("设备无历史IP数据");
      return;
    }

    // 格式化
    const preList = deviceIpHistory.map(
      (d) => `${fmtDate(d.created_at)} - ${d.ip_v4 || "-"} / ${d.ip_v6 || "-"}`,
    );

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 设备 IP 历史 - ${name}`,
      preList.join("\n"),
      deviceIpHistory.length,
    );
  } catch (err) {
    alert(`打开 历史IP 弹窗失败: ${err.message}`);
    console.error(err);
  }
}

// 打开 设备信息 弹窗
async function onDeviceInfoPopup(id, name) {
  try {
    const { data: deviceStatus } = await API.get(`/api/devices/${id}/status`);
    console.log("deviceStatus", deviceStatus);
    if (!deviceStatus) {
      alert("接口异常");
      return;
    }
    if (deviceStatus.length === 0) {
      alert("当前设备无状态数据");
      return;
    }

    // 格式化
    const preList = deviceStatus.map(
      (d) =>
        `${d.role} - ${d.version || "-"} / ${d.go_version || "-"} (${d.os}/${d.arch})`,
    );

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `设备状态 - ${name}`,
      preList.join("\n"),
      deviceStatus.length,
    );
  } catch (err) {
    alert(`打开 设备信息 弹窗失败: ${err.message}`);
    console.error(err);
  }
}
