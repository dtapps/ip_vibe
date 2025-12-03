// 滚动超过 300 px 出现
window.addEventListener("scroll", () => {
  document
    .querySelector(".back-top")
    .classList.toggle("visible", window.scrollY > 300);
});

// API请求封装 - 带错误处理
const API = {
  async get(url) {
    try {
      const response = await fetch(url);
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
  const title = document.getElementById(titleID);
  if (title) title.textContent = titleName;
  if (modal) modal.classList.add("show");
}

// 关闭模态框
function closeModal(id) {
  const modal = document.getElementById(id);
  if (!modal) return;

  // 1. 隐藏弹窗
  modal.classList.remove("show");

  // 2. 如果有 form 才 reset
  const form = modal.querySelector("form");
  if (form) form.reset();

  // 3. 如果有平台字段才隐藏
  modal
    .querySelectorAll(".platform-fields")
    .forEach((el) => el.classList.add("hidden"));
}

// 一键复制
function copyData(id) {
  const pre = document.getElementById(id);
  navigator.clipboard
    .writeText(pre.textContent)
    .then(() => {
      const btn = document.querySelector(".ip-popup-copy");
      const old = btn.innerHTML;
      btn.innerHTML =
        '<svg class="ip-i"><use href="/static/img/icons.svg#ip-check"/></use></svg> 已复制';
      btn.style.color = "var(--success)";
      setTimeout(() => {
        btn.innerHTML = old;
        btn.style.color = "";
      }, 1500);
    })
    .catch(() => alert("复制失败，请手动选择复制"));
}

// 格式化健康检查结果
function formatHealthCheckResult(healthData) {
  return healthData
    .map((group) =>
      group.map((item) => `${item.key}: ${item.value || "-"}`).join("\n")
    )
    .join("\n-----\n");
}

// 显示 loading 弹窗
function showLoadingPopup(idEl = "popup") {
  const titleEl = document.getElementById(`${idEl}Title`);
  const countEl = document.getElementById(`${idEl}Count`);
  const preEl = document.getElementById(`${idEl}Pre`);
  const copyBtn = document.getElementById(`${idEl}CopyBtn`);

  titleEl.textContent = "检测中...";
  preEl.textContent = "正在检测中，请稍候...（预计 5~15 秒）";
  preEl.style.color = "var(--text-secondary)";
  countEl.textContent = "...";
  copyBtn.disabled = true;

  showModal(idEl, titleEl, `检测中...`);
}

// 显示弹窗内容
function showPopupContent(idEl = "popup", title = "", content = "", count = 0) {
  const titleEl = document.getElementById(`${idEl}Title`);
  const countEl = document.getElementById(`${idEl}Count`);
  const preEl = document.getElementById(`${idEl}Pre`);
  const copyBtn = document.getElementById(`${idEl}CopyBtn`);

  titleEl.textContent = title;
  preEl.textContent = content;
  preEl.style.color = "#f8f8f2";
  countEl.textContent = count;
  copyBtn.disabled = false;

  showModal(idEl, titleEl, title);
}
