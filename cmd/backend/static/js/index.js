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

// 一键复制
function copyData(idEl = "popup") {
  // const modal = document.getElementById(idEl);
  const preEl = document.getElementById(`${idEl}Pre`);
  if (!preEl || !preEl.textContent?.trim()) {
    console.log("无可复制内容");
    return;
  }
  const text = preEl.textContent;

  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(text)
      .then(() => showCopySuccess(idEl))
      .catch(() => fallbackToExecCommand(text, idEl));
  } else {
    fallbackToExecCommand(text, idEl);
  }
}

// 执行命令复制
function fallbackToExecCommand(text, idEl) {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.cssText =
    "position:fixed; top:0; left:0; opacity:0; pointer-events:none;";
  document.body.appendChild(textarea);
  textarea.select();
  try {
    const success = document.execCommand("copy");
    if (success) {
      showCopySuccess(idEl);
    } else {
      console.log("复制失败，请手动选择文本复制");
    }
  } catch (err) {
    console.log("浏览器限制，请手动复制");
  } finally {
    document.body.removeChild(textarea);
  }
}

// 显示复制成功
function showCopySuccess(idEl = "popup") {
  const btn = document.getElementById(`${idEl}CopyBtn`);
  if (!btn) {
    return;
  }
  const old = btn.innerHTML;
  btn.innerHTML =
    '<svg class="ui-icon"><use href="/static/img/icons.svg#ip-check"/></use></svg> 已复制';
  setTimeout(() => {
    btn.innerHTML = old;
    btn.style.color = "";
  }, 1500);
}

// 格式化健康检查结果
function formatHealthCheckResult(healthData) {
  return healthData
    .map((group) =>
      group.map((item) => `${item.key}: ${item.value || "-"}`).join("\n")
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
  options = {}
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
