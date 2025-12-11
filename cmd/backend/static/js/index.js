// 滚动超过 300 px 出现
window.addEventListener("scroll", () => {
  document
    .querySelector(".back-top")
    .classList.toggle("visible", window.scrollY > 300);
});

// API请求封装 - 带错误处理
const API = {
  async get(url, data) {
    // 把 data 对象转成查询字符串
    const queryString = new URLSearchParams(data).toString();
    // 拼到 URL 后面
    const fullUrl = queryString ? `${url}?${queryString}` : url;
    try {
      const response = await fetch(fullUrl, {
        method: "GET",
        headers: { "Content-Type": "application/json" },
      });
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async post(url, data) {
    try {
      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async patch(url, data) {
    try {
      const response = await fetch(url, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async delete(url) {
    try {
      const response = await fetch(url, { method: "DELETE" });
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
};

// 响应处理函数
async function handleResponse(response) {
  let data;
  try {
    data = await response.json();
  } catch (error) {
    throw new Error(`JSON解析错误: ${error.message}`);
  }

  if (!response.ok) {
    const errorMsg =
      data.message ||
      data.error ||
      `请求失败: ${response.status} ${response.statusText}`;
    throw new Error(errorMsg);
  }

  return data;
}

// 错误处理函数
function handleError(error) {
  if (error.name === "TypeError" && error.message.includes("Failed to fetch")) {
    return new Error("网络错误，请检查您的网络连接");
  }
  return error;
}

// 通用错误处理包装函数
async function withErrorHandling(asyncFn, options = {}) {
  try {
    const result = await asyncFn();
    if (options.successAction) {
      options.successAction(result);
    }
    return result;
  } catch (error) {
    const errorMessage = options.errorMessage || "操作失败";
    alert(`${errorMessage}: ${error.message}`);
    console.error(errorMessage, error);

    if (options.revertAction) {
      options.revertAction();
    }
    return null;
  }
}

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
  modal.classList.add("show");

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
  modal.classList.remove("show");

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
    '<svg class="ip-i"><use href="/static/img/icons.svg#ip-check"/></use></svg> 已复制';
  btn.style.color = "var(--success)";
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
    preEl.style.color = "var(--text-secondary)";
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
    preEl.style.color = "#f8f8f2";
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
