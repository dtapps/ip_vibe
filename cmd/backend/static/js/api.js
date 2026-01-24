const ApiContentType = "application/json";

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
        headers: { "Content-Type": ApiContentType },
      });
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async post(url, data) {
    try {
      info = {
        method: "POST",
        headers: { "Content-Type": ApiContentType },
      };
      if (data) {
        info.body = JSON.stringify(data);
      }
      const response = await fetch(url, info);
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async patch(url, data) {
    try {
      info = {
        method: "PATCH",
        headers: { "Content-Type": ApiContentType },
      };
      if (data) {
        info.body = JSON.stringify(data);
      }
      const response = await fetch(url, info);
      return await handleResponse(response);
    } catch (error) {
      throw handleError(error);
    }
  },
  async delete(url, data) {
    try {
      info = {
        method: "DELETE",
        headers: { "Content-Type": ApiContentType },
      };
      if (data) {
        info.body = JSON.stringify(data);
      }
      const response = await fetch(url, info);
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

/**
 * 批量操作执行器
 * @param {HTMLButtonElement} el - 触发批量操作的按钮元素
 */
async function handleBatchAction(el) {
  console.log("[handleBatchAction] el.dataset", el.dataset);
  try {
    // 提取配置
    const {
      url,
      method = "POST",
      idKey = "ids",
      key,
      val,
      confirm: customMsg,
      isReload = true,
    } = el.dataset;

    // 获取选中的 ID 数组
    const selectedIds = Array.from(
      document.querySelectorAll(".row-checkbox:checked"),
    ).map((cb) => parseInt(cb.value));
    if (!selectedIds.length) {
      alert("请先选择要操作的项目");
      return;
    }

    // 构造确认信息
    const count = selectedIds.length;
    let msg = customMsg
      ? customMsg.replace("{n}", count)
      : `确认对选中的 ${count} 项执行该操作吗？`;
    if (!confirm(msg)) {
      return;
    }

    // 动态构造 Payload
    const payload = { [idKey]: selectedIds };
    if (key) {
      // 转换布尔值或保留原始字符串
      payload[key] = val === "true" ? true : val === "false" ? false : val;
    }

    // 根据请求方式调用 API
    const methodFunc = method.toLowerCase();
    if (typeof API[methodFunc] !== "function") {
      throw new Error(`不支持的请求方式: ${method}`);
    }
    await API[methodFunc](url, payload);

    // 刷新页面
    if (isReload) {
      window.location.reload();
    }
  } catch (err) {
    alert(`批量操作失败: ${err.message}`);
    console.error("Batch Action Error:", err);
  }
}

/**
 * 处理操作
 * @param {HTMLButtonElement} el - 删除按钮元素
 * @returns
 */
async function handleAction(el) {
  console.log("[handleAction] el.dataset", el.dataset);
  try {
    // 提取配置
    const {
      id,
      url,
      method = "POST",
      confirm: customMsg,
      isReload = true,
    } = el.dataset;

    if (id && !confirm(customMsg || `确定对 ${id} 执行该操作吗？`)) {
      return;
    }

    // 如果存在 id 就添加到 URL 中
    // if (id) {
    //   url += `/${id}`;
    // }

    // 根据请求方式调用 API
    const methodFunc = method.toLowerCase();
    if (typeof API[methodFunc] !== "function") {
      throw new Error(`不支持的请求方式: ${method}`);
    }
    await API[methodFunc](url);

    // 刷新页面
    if (isReload) {
      window.location.reload();
    }
  } catch (err) {
    alert(`操作失败: ${err.message}`);
    console.error("Action Error:", err);
  }
}
