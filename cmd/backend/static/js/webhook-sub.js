// 选择通知 webhook_id
function onWebhookIdSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onWebhookIdSelect", opt);
}

// 选择订阅对象 obj_id
function onObjIdSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  // console.log("onObjIdSelect", opt);
}

// 打开表单
async function openModal(id, obj_type) {
  document.getElementById("form").reset();

  await withErrorHandling(
    async () => {
      // 通知列表
      const { data: webhooks } = await API.get(`/api/webhooks`);
      console.log("webhooks", webhooks);
      if (webhooks) {
        const webhookIdSelect = document.getElementById("webhook_id");
        webhookIdSelect.innerHTML = "";
        addPlaceholder(webhookIdSelect, "请选择通知");
        webhooks.forEach((d) => {
          const opt = document.createElement("option");
          opt.value = d.id;
          opt.text = `${d.name}`;
          webhookIdSelect.appendChild(opt);
        });
      }

      var modalTitle = "";
      switch (obj_type) {
        case "device":
          modalTitle = "设备订阅";

          // 设备列表
          const { data: devices } = await API.get(`/api/devices`);
          console.log("devices", devices);
          if (devices) {
            const objIdSelect = document.getElementById("obj_id");
            objIdSelect.innerHTML = "";
            devices.forEach((d) => {
              const opt = document.createElement("option");
              opt.value = d.id;
              opt.text = `${d.name}`;
              objIdSelect.appendChild(opt);
            });
          }

          break;
        case "dns_record":
          modalTitle = "DNS订阅";

          // DNS列表
          const { data: dnsRecords } = await API.get(`/api/dns-records`);
          console.log("dnsRecords", dnsRecords);
          if (dnsRecords) {
            const objIdSelect = document.getElementById("obj_id");
            objIdSelect.innerHTML = "";
            dnsRecords.forEach((d) => {
              const opt = document.createElement("option");
              opt.value = d.id;
              opt.text = `${d.domain} - ${d.type}`;
              objIdSelect.appendChild(opt);
            });
          }

          break;
        case "cdn_record":
          modalTitle = "CDN订阅";

          // CDN列表
          const { data: cdnRecords } = await API.get(`/api/cdn-records`);
          console.log("cdnRecords", cdnRecords);
          if (cdnRecords) {
            const objIdSelect = document.getElementById("obj_id");
            objIdSelect.innerHTML = "";
            cdnRecords.forEach((d) => {
              const opt = document.createElement("option");
              opt.value = d.id;
              opt.text = `${d.domain} - ${d.type}`;
              objIdSelect.appendChild(opt);
            });
          }

          break;
        default:
          return alert("请选择订阅对象类型");
      }

      // 编辑
      if (id) {
        const { data: webhookSubscription } = await API.get(
          `/api/webhook-subs/${id}`
        );
        console.log("webhookSubscription", webhookSubscription);
        if (!webhookSubscription) {
          alert("通知订阅数据异常");
          return;
        }

        document.getElementById("id").value = webhookSubscription.id;
        document.getElementById("id").disabled = true;
        document.getElementById("name").value = webhookSubscription.name;
        document.getElementById("webhook_id").value =
          webhookSubscription.webhook_id;
        document.getElementById("webhook_id").disabled = true;
        document.getElementById("obj_id").value = webhookSubscription.obj_id;
        document.getElementById("obj_id").disabled = true;
        document.getElementById("obj_type").value =
          webhookSubscription.obj_type;
        document.getElementById("obj_type").disabled = true;

        showModal("modal", "modalTitle", `编辑 ${modalTitle}`);
      }

      // 创建
      if (!id) {
        document.getElementById("id").value = "";
        document.getElementById("id").disabled = true;
        document.getElementById("webhook_id").disabled = false;
        document.getElementById("obj_id").disabled = false;
        document.getElementById("obj_type").value = obj_type;
        document.getElementById("obj_type").disabled = true;

        showModal("modal", "modalTitle", `添加 ${modalTitle}`);
      }
    },
    {
      errorMessage: "加载通知订阅信息失败",
    }
  );
}

// 表单提交
document.getElementById("form").onsubmit = async (e) => {
  e.preventDefault();

  // 表单数据
  const form_id = Number(document.getElementById("id").value.trim());
  const form_data = {
    name: document.getElementById("name").value.trim(),
    webhook_id: Number(document.getElementById("webhook_id").value.trim()),
    obj_type: document.getElementById("obj_type").value.trim(),
    obj_ids: Array.from(document.getElementById("obj_id").selectedOptions).map(
      (opt) => Number(opt.value.trim())
    ),
  };
  console.log("form_id, form_data", form_id, form_data);

  // 校验
  if (!form_data.name) {
    return alert("请输入名称");
  }
  if (isNaN(form_data.webhook_id) || form_data.webhook_id <= 0) {
    return alert("请选择通知渠道");
  }
  if (!form_data.obj_type) {
    return alert("请选择订阅对象类型");
  }
  if (form_data.obj_ids.length === 0) {
    return alert("请选择订阅对象");
  }

  await withErrorHandling(
    async () => {
      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/webhook-subs", form_data);
      } else {
        await API.patch(`/api/webhook-subs/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    },
    {
      errorMessage: "操作失败",
    }
  );
};

// 删除通知订阅
async function deleteWebhookSubscription(id) {
  if (confirm("确定删除此通知订阅？")) {
    await withErrorHandling(
      async () => {
        await API.delete(`/api/webhook-subs/${id}`);
        // 刷新页面
        location.reload();
      },
      {
        errorMessage: "删除失败",
      }
    );
  }
}
