// 选择设备 devices
function onDevicesSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onDevicesSelect", opt);
}

// 打开 表单 弹窗
async function openFormModal(id) {
  document.getElementById("form").reset();

  // 选择 Select2 控件
  const selectId2 = $("#devices");

  try {
    // 设备列表
    const { data: devices } = await API.get(`/api/devices`);
    console.log("devices", devices);
    if (devices) {
      // 销毁旧实例（防止 Select2 冲突）
      if (selectId2.data("select2")) {
        selectId2.select2("destroy");
      }

      // 准备 Select2 需要的数据格式
      const select2Data = [];
      if (devices && devices.length > 0) {
        devices.forEach((d) => {
          select2Data.push({
            id: d.id,
            text: `${d.name}`,
          });
        });
      }

      // 初始化 Select2 控件
      console.log("select2 初始化", selectId2, select2Data);
      selectId2.select2({
        theme: "tailwindcss-4", // 主题
        width: "100%",
        placeholder: "请选择设备",
        allowClear: true,
        multiple: true,
        data: select2Data, // 数据
        closeOnSelect: false, // 多选时防止关闭下拉列表
        dropdownParent: document.querySelector("#modal .ui-modal-box"), // 绑定 modal 容器，防 z-index/定位错
      });
    }

    // 编辑
    if (id) {
      const { data: ipGroup } = await API.get(`/api/ip-groups/${id}`);
      console.log("ipGroup", ipGroup);
      if (!ipGroup) {
        alert("IP 组数据异常");
        return;
      }

      document.getElementById("id").value = ipGroup.id;
      document.getElementById("id").disabled = true;
      document.getElementById("name").value = ipGroup.name;
      document.getElementById("description").value = ipGroup.description ?? "";

      if (ipGroup.devices && ipGroup.devices.length > 0) {
        // 高亮已选
        console.log("select2 高亮已选", ipGroup.devices);
        selectId2.val(ipGroup.devices).trigger("change");
      } else {
        // 清空已选
        console.log("select2 清空已选", ipGroup.devices);
        selectId2.val(null).trigger("change");
      }

      showModal("formModal", "modalTitle", "编辑 IP 组");
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

      // 清空已选
      console.log("select2 清空已选");
      selectId2.val(null).trigger("change");

      showModal("formModal", "modalTitle", "添加 IP 组");
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

    const selectedDeviceIDs = $("#devices").val() || [];

    // 表单数据
    const form_id = Number(document.getElementById("id").value.trim());
    const form_data = {
      name: document.getElementById("name").value.trim(),
      description: document.getElementById("description").value.trim(),
      devices: selectedDeviceIDs.map(Number),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
      return alert("请输入组名");
    }
    if (form_data.devices.length === 0) {
      return alert("请选择设备");
    }

    try {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/ip-groups", form_data);
      } else {
        await API.patch(`/api/ip-groups/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`操作失败: ${error.message}`);
      console.error("IP组操作失败", error);
    }
  };
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

// 删除
async function onDelete(id) {
  if (confirm("确定删除此 IP 组？")) {
    try {
      await API.delete(`/api/ip-groups/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除IP组失败", error);
    }
  }
}

// 打开 检测 弹窗
async function onHealthcheckPopup(id) {
  // 显示 loading 弹窗
  showLoadingPopup("popup");

  try {
    const { data: ipGroupHealthcheck } = await API.get(
      `/api/ip-groups/${id}/health-check`
    );
    console.log("ipGroupHealthcheck", ipGroupHealthcheck);
    if (!ipGroupHealthcheck) {
      alert("IP组检测数据异常");
      return;
    }
    if (ipGroupHealthcheck.length === 0) {
      alert("当前组内无检测数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 检测 IP`,
      formatHealthCheckResult(ipGroupHealthcheck),
      ipGroupHealthcheck.length,
      { resetClosedState: true }
    );
  } catch (error) {
    alert(`加载IP组检测信息失败: ${error.message}`);
    console.error("查看检测信息失败", error);
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
