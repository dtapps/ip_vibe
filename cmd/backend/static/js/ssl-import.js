// 选择配置 config_id
function onConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onConfigSelect", opt);
}

/**
 * 打开 添加表单 弹窗
 * @returns
 */
async function openAddFormModal() {
  document.getElementById("addForm").reset();

  try {
    showModal("addFormModal", "modalTitle", "添加 SSL 证书");
  } catch (err) {
    alert(`加载数据失败: ${err.message}`);
    console.error("openAddFormModal failed", err);
    return;
  }
}

// 表单提交
const addForm = document.getElementById("addForm");
if (addForm) {
  addForm.onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_data = {
      domain: document.getElementById("domain").value.trim(),
      certificate: document.getElementById("certificate").value.trim(),
      private_key: document.getElementById("private_key").value.trim(),
    };
    console.log("form_data", form_data);

    // 校验
    if (!form_data.domain) {
      return alert("请输入域名");
    }

    try {
      // 提交
      await API.post("/api/ssl-imports", form_data);

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error("onsubmit failed", err);
      return;
    }
  };
}

/**
 * 打开 导入表单 弹窗
 * @param {HTMLElement} btn 触发按钮对象
 * @returns
 */
async function openImportFormModal(btn) {
  document.getElementById("importForm").reset();

  try {
    // 获取并解析数据
    if (!btn.dataset.detail) {
      throw new Error("未找到详情数据");
    }
    const d = JSON.parse(btn.dataset.detail);

    document.getElementById("domain").value = d.domain;
    document.getElementById("cert_private_key").value = d.private_key;
    document.getElementById("cert_certificate").value = d.certificate;

    // 平台配置列表
    const { data: configs } = await API.get(`/api/platform-configs`);
    console.log("configs", configs);
    if (configs) {
      const configSelect = document.getElementById("config_id");
      configSelect.innerHTML = "";
      addPlaceholder(configSelect, "请选择平台配置");
      configs.forEach((d) => {
        if (d.capability === "ssl" || d.capability === "all") {
          const opt = document.createElement("option");
          opt.dataset.id = d.id;
          opt.dataset.capability = d.capability;
          opt.dataset.platform = d.platform;
          opt.dataset.domain_name = d.config_json?.domain_name ?? "";
          opt.dataset.email = d.config_json?.email ?? "";
          opt.value = d.id;
          opt.text = `${d.name}（${opt.dataset.email}）`;
          configSelect.appendChild(opt);
        }
      });
    }

    showModal("importFormModal", "modalTitle", "导入 SSL 记录");
  } catch (err) {
    alert(`加载数据失败: ${err.message}`);
    console.error("openImportFormModal failed", err);
    return;
  }
}

const importForm = document.getElementById("importForm");
if (importForm) {
  importForm.onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_data = {
      config_id: Number(document.getElementById("config_id").value.trim()),
      domain: document.getElementById("domain").value.trim(),
      cert_private_key: document
        .getElementById("cert_private_key")
        .value.trim(),
      cert_certificate: document
        .getElementById("cert_certificate")
        .value.trim(),
    };
    console.log("form_data", form_data);

    // 校验
    if (!form_data.domain) {
      return alert("请输入域名");
    }

    try {
      // 提交
      await API.post("/api/ssl-records/import", form_data);

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error("onsubmit failed", err);
      return;
    }
  };
}

/**
 * 打开 查看 弹窗
 * @param {HTMLElement} btn 触发按钮对象
 * @returns
 */
async function openViewModal(btn) {
  try {
    // 获取并解析数据
    if (!btn.dataset.detail) {
      throw new Error("未找到详情数据");
    }
    const d = JSON.parse(btn.dataset.detail);

    const crtEl = document.getElementById("sslCrt");
    const keyEl = document.getElementById("sslKey");

    if (crtEl) crtEl.value = d.certificate || "暂无证书内容";
    if (keyEl) keyEl.value = d.private_key || "暂无私钥内容";

    showModal("viewModal", "modalTitle", "证书内容详情");
  } catch (err) {
    alert(`加载数据失败: ${err.message}`);
    console.error("openViewModal failed", err);
    return;
  }
}
