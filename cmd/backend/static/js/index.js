// 统一给 select 插「请选择」占位
function addPlaceholder(select, text = "请选择") {
  const ph = document.createElement("option");
  ph.value = "";
  ph.text = text;
  ph.textContent = text;
  ph.disabled = true;
  ph.selected = true;
  select.insertBefore(ph, select.firstChild);
}

// 辅助：ISO → MM-DD HH:mm
const fmtDate = (iso) => {
  const d = new Date(iso);
  const MM = String(d.getMonth() + 1).padStart(2, "0");
  const DD = String(d.getDate()).padStart(2, "0");
  const HH = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${MM}-${DD} ${HH}:${mm}`;
};

// 打开配置模态框
function showModal(id, titleID, titleName) {
  const modal = document.getElementById(id);
  if (!modal) {
    return;
  }

  const title = document.getElementById(titleID);
  if (title) {
    title.textContent = titleName;
  }

  // 显示弹窗
  modal.classList.add("modal-open");

  // 在 DOM 元素上挂状态
  modal.dataset.manuallyClosed = "false";
}

// 关闭模态框
function closeModal(id) {
  const modal = document.getElementById(id);
  if (!modal) {
    return;
  }

  // 隐藏弹窗
  modal.classList.remove("modal-open");

  // 在 DOM 元素上挂状态
  modal.dataset.manuallyClosed = "true";

  // 如果有 form 才 reset
  const form = modal.querySelector("form");
  if (form) {
    form.reset();
  }

  // 如果有平台字段才隐藏
  modal
    .querySelectorAll(".platform-fields")
    .forEach((el) => el.classList.add("hidden"));

  // 销毁旧实例（防止 Select2 冲突）
  const webhookSelect = modal.querySelector("#webhook_ids");
  if (webhookSelect) {
    const $webhook = $(webhookSelect);
    if ($webhook.data("select2")) {
      $webhook.select2("destroy");
    }
  }
}

// 格式化健康检查结果
function formatHealthCheckResult(healthData) {
  return healthData
    .map((group) =>
      group.map((item) => `${item.key}: ${item.value || "-"}`).join("\n"),
    )
    .join("\n--------\n");
}

// 显示 loading 弹窗
function showLoadingPopup(idEl = "popup") {
  // const modal = document.getElementById(idEl);
  const titleEl = document.getElementById(`${idEl}Title`);
  const countEl = document.getElementById(`${idEl}Count`);
  const preEl = document.getElementById(`${idEl}Pre`);
  const copyBtn = document.getElementById(`${idEl}CopyBtn`);

  if (titleEl) {
    titleEl.textContent = "检测中...";
  }
  if (preEl) {
    preEl.textContent = "正在检测中，请稍候...（预计 5~15 秒）";
    preEl.style.color = ""; // 清除之前可能存在的颜色
    preEl.classList.add("opacity-50");
  }
  if (countEl) {
    countEl.textContent = "...";
  }
  if (copyBtn) {
    copyBtn.disabled = true;
  }

  showModal(idEl, titleEl, `检测中...`);
}

// 显示弹窗内容
function showPopupContent(
  idEl = "popup",
  title = "",
  content = "",
  count = 0,
  options = {},
) {
  const modal = document.getElementById(idEl);
  const titleEl = document.getElementById(`${idEl}Title`);
  const countEl = document.getElementById(`${idEl}Count`);
  const preEl = document.getElementById(`${idEl}Pre`);
  const copyBtn = document.getElementById(`${idEl}CopyBtn`);

  if (modal) {
    if (options.resetClosedState) {
      if (modal.dataset.manuallyClosed === "true") {
        console.log(`用户已关闭弹窗 ${idEl}，跳过自动显示`);
        return;
      }
    }
  }

  if (titleEl) {
    titleEl.textContent = title;
  }
  if (preEl) {
    preEl.textContent = content;
    preEl.style.color = ""; // 清除之前可能存在的颜色
    preEl.classList.remove("opacity-50");
  }
  if (countEl) {
    countEl.textContent = count;
  }
  if (copyBtn) {
    copyBtn.disabled = false;
  }

  showModal(idEl, titleEl, title);
}

// 自定义 url 请求 post 接口
async function apiPost(url, data, options = {}) {
  console.log("apiPost", url, data, options);

  const { errorMsg = "操作失败", reload = false } = options;

  if (!url) {
    return;
  }

  try {
    const result = await API.post(url, data);
    console.info("apiPost", result);
    // 刷新页面
    if (reload) {
      location.reload();
    }
  } catch (error) {
    alert(`${errorMsg}: ${error.message}`);
    console.error("apiPost", error);
  }
}
// 自定义 url 请求 delete 接口
async function apiDelete(url, options = {}) {
  console.log("apiDelete", url, options);

  const {
    tips = "确定删除此记录？",
    errorMsg = "删除失败",
    reload = false,
  } = options;

  if (!url) {
    return;
  }

  // 提示是否确认删除
  const confirmed = confirm(tips);
  if (!confirmed) {
    return;
  }

  try {
    const result = await API.delete(url);
    console.info("apiDelete", result);
    // 刷新页面
    if (reload) {
      location.reload();
    }
  } catch (error) {
    alert(`${errorMsg}: ${error.message}`);
    console.error("apiDelete", error);
  }
}

document.querySelectorAll(".simple-tooltip").forEach((item) => {
  item.addEventListener("mouseenter", function (e) {
    // 创建提示框
    const tooltip = document.createElement("div");
    tooltip.className =
      "tooltip-content tooltip-" + (this.dataset.position || "top");
    tooltip.textContent = this.dataset.tooltip;

    // 添加到页面
    this.appendChild(tooltip);

    // 显示
    tooltip.style.display = "block";

    // 保存引用，用于移除
    this._tooltip = tooltip;
  });

  item.addEventListener("mouseleave", function (e) {
    // 隐藏并移除
    if (this._tooltip) {
      this._tooltip.style.display = "none";
      this.removeChild(this._tooltip);
      this._tooltip = null;
    }
  });
});

// 通用切换密码显示
function togglePasswordVisibilityByPrev(btn) {
  const input = btn.previousElementSibling;
  const icon = btn.querySelector("span");
  if (input) {
    const isPwd = input.type === "password";
    input.type = isPwd ? "text" : "password";
    // 动态切换图标类名
    if (icon) {
      icon.className = isPwd ? "i-lucide-x w-4 h-4" : "i-lucide-eye w-4 h-4";
    }
  }
}

/**
 * 批量更新 Form 字段状态与值
 * @param {Array} configs - 结构: [{ id: "id名", value: "值", disabled: true/false }]
 */
function updateFormFields(configs = []) {
  if (!Array.isArray(configs)) return;

  configs.forEach((item) => {
    const el = document.getElementById(item.id);
    if (!el) return;

    // 更新值：判断 undefined 是为了允许传 0 或空字符串
    if (item.value !== undefined) {
      el.value = item.value;
    }

    // 更新禁用状态
    if (item.disabled !== undefined) {
      el.disabled = item.disabled;
    }
  });
}

/**
 * 表格全选逻辑
 */
function tableSelectAll(master) {
  const isChecked = master.checked;
  document.querySelectorAll(".row-checkbox").forEach((cb) => {
    cb.checked = isChecked;
  });
  updateBatchBar(); // 同步 UI
}

/**
 * 更新批量操作 UI 状态
 * 只要点击了复选框（不管是全选还是单选），都要跑这个方法
 */
function updateBatchBar() {
  const allRows = document.querySelectorAll(".row-checkbox");
  const checkedRows = document.querySelectorAll(".row-checkbox:checked");
  const checkAll = document.getElementById("check_all");
  const batchBtn = document.getElementById("batch_delete_btn");
  const countSpan = document.getElementById("selected_count");

  const count = checkedRows.length;

  // 1. 同步全选框的状态 (半选、全选、不选)
  if (checkAll) {
    checkAll.checked = allRows.length === count && allRows.length > 0;
    checkAll.indeterminate = count > 0 && count < allRows.length;
  }

  // 2. 控制批量按钮显示/隐藏
  if (batchBtn) {
    if (count > 0) {
      countSpan.innerText = count;
      batchBtn.classList.remove("hidden"); // 显示
    } else {
      batchBtn.classList.add("hidden"); // 隐藏
    }
  }
}
