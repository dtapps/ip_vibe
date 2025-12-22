function changePlatform(platform) {
  switch (platform) {
    case "wechat_work":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = true;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = false;
      document.getElementById("secret_group").style.display = "none";
      break;
    case "dingtalk":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = true;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = false;
      document.getElementById("secret_group").style.display = "block";
      break;
    case "feishu":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = true;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = false;
      document.getElementById("secret_group").style.display = "block";
      break;
    case "bark":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = false;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = false;
      document.getElementById("secret_group").style.display = "block";
      break;
    case "ft07":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = false;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = true;
      document.getElementById("secret_group").style.display = "block";
      break;
    case "push_deer":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = false;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = false;
      document.getElementById("secret").required = true;
      document.getElementById("secret_group").style.display = "block";
      break;
    case "custom_http":
      document.getElementById("url").disabled = false;
      document.getElementById("url").required = true;
      document.getElementById("url_group").style.display = "block";

      document.getElementById("secret").disabled = true;
      document.getElementById("secret").required = false;
      document.getElementById("secret_group").style.display = "none";
      break;
  }
}

// 选择平台类型 platform
function onPlatformSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onPlatformSelect", opt);
  const id = document.getElementById("id").value.trim();
  console.log("onPlatformSelect id", id);

  changePlatform(opt.value);

  if (!id) {
    document.getElementById("template_title").value =
      "[IPVibe] [[.DeviceName]] IP 更新";
    switch (opt.value) {
      case "wechat_work":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "dingtalk":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "feishu":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "bark":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "ft07":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "push_deer":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
      case "custom_http":
        document.getElementById(
          "template_content"
        ).value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`;
        break;
    }
  }
}

// 打开 表单 弹窗
async function openFormModal(id) {
  document.getElementById("form").reset();

  try {
    // 编辑
    if (id) {
      const { data: webhook } = await API.get(`/api/webhooks/${id}`);
      console.log("webhook", webhook);
      if (!webhook) {
        alert("通知数据异常");
        return;
      }

      document.getElementById("id").value = webhook.id;
      document.getElementById("id").disabled = true;
      document.getElementById("name").value = webhook.name;
      document.getElementById("platform").value = webhook.platform;
      document.getElementById("platform").disabled = true;
      document.getElementById("url").value = webhook.url;
      document.getElementById("secret").value = webhook.secret;
      document.getElementById("template_title").value = webhook.template_title;
      document.getElementById("template_content").value =
        webhook.template_content;

      changePlatform(webhook.platform);

      showModal("modal", "modalTitle", "编辑 通知");
    }

    // 创建
    if (!id) {
      document.getElementById("id").value = "";
      document.getElementById("id").disabled = true;

      showModal("modal", "modalTitle", "添加 通知");
    }
  } catch (error) {
    alert(`加载数据失败: ${error.message}`);
    console.error("打开通知模态框失败", error);
  }
}

// 表单提交
document.getElementById("form").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("id").value.trim());
  const form_data = {
    name: document.getElementById("name").value.trim(),
    platform: document.getElementById("platform").value.trim(),
    url: document.getElementById("url").value.trim(),
    secret: document.getElementById("secret").value.trim(),
    template_title: document.getElementById("template_title").value.trim(),
    template_content: document.getElementById("template_content").value.trim(),
  };
  console.log("form_id, form_data", form_id, form_data);

  // 校验
  if (!form_data.name) {
    return alert("请输入组名");
  }
  if (!form_data.platform) {
    return alert("请选择平台类型");
  }
  if (!form_data.template_title) {
    return alert("请输入通知标题模板");
  }
  if (!form_data.template_content) {
    return alert("请输入通知内容模板");
  }

  try {
    // 提交
    if (isNaN(form_id) || form_id <= 0) {
      await API.post("/api/webhooks", form_data);
    } else {
      await API.patch(`/api/webhooks/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
  } catch (error) {
    alert(`操作失败: ${error.message}`);
    console.error("通知操作失败", error);
  }
};

// 测试
async function onTest(id) {
  if (confirm("确定测试此通知？")) {
    try {
      await API.post(`/api/webhooks/${id}/test`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`测试失败: ${error.message}`);
      console.error("测试通知失败", error);
    }
  }
}

// 删除
async function onDelete(id) {
  if (confirm("确定删除此通知？")) {
    try {
      await API.delete(`/api/webhooks/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`删除失败: ${error.message}`);
      console.error("删除通知失败", error);
    }
  }
}
