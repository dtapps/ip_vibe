// 选择配置 config_id
function onConfigSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onConfigSelect", opt);
}

// 选择证书提供方 provider
function onProviderSelect(sel) {
  const opt = sel.options[sel.selectedIndex];
  console.log("onProviderSelect", opt);
}

/**
 * 打开 表单 弹窗
 * @param {*} id 平台配置ID
 * @param {string} action 操作类型，add / edit / copy
 * @returns
 */
async function openFormModal(id, action = "") {
  document.getElementById("form").reset();

  try {
    showModal("formModal", "modalTitle", "导入 SSL记录");
  } catch (err) {
    alert(`加载数据失败: ${err.message}`);
    console.error("openFormModal failed", err);
    return;
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
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
} else {
  console.warn("未找到表单，跳过 onsubmit 绑定");
}

/**
 * 打开 查看 弹窗
 * @param {HTMLElement} btn 触发按钮对象
 * @returns
 */
async function openViewModal(btn) {
  try {
    // 1. 获取并解析数据
    if (!btn.dataset.detail) throw new Error("未找到详情数据");
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
