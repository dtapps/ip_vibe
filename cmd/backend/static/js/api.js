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
 * 统一处理批量操作
 * @param {string} url - 操作接口URL
 * @param {string} key - 操作键名
 * @param {boolean} value - 操作值
 * @returns
 */
async function handleBatchAction(url, key, value) {
  try {
    const ids = Array.from(
      document.querySelectorAll(".row-checkbox:checked"),
    ).map((cb) => parseInt(cb.value));
    if (!ids.length) {
      alert("请选择要操作的项");
      return;
    }

    const msg = value ? "开启" : "关闭";
    if (!confirm(`确定要批量${msg}这 ${ids.length} 项吗？`)) {
      return;
    }

    // 调用 API
    await API.post(url, {
      ids: ids,
      [key]: value,
    });

    // 刷新页面
    window.location.reload();
  } catch (err) {
    alert(`操作失败: ${err.message}`);
  }
}

/**
 * 批量删除
 * @param {string} url - 删除接口URL
 * @returns
 */
async function handleBatchDelete(url) {
  try {
    const ids = Array.from(
      document.querySelectorAll(".row-checkbox:checked"),
    ).map((cb) => parseInt(cb.value));
    if (!ids.length) {
      alert("请选择要删除的项");
      return;
    }

    if (!confirm(`确定删除这 ${ids.length} 项吗？`)) {
      return;
    }

    // 调用 API
    await API.delete(url, { ids: ids });

    // 刷新页面
    window.location.reload();
  } catch (error) {
    alert(`批量删除失败: ${error.message}`);
    console.error("批量删除失败", error);
  }
}

/**
 * 删除
 * @param {string} url - 删除接口URL
 * @returns
 */
async function handleDelete(url, id) {
  try {
    if (!id) {
      alert("请选择要删除的项");
      return;
    }

    if (!confirm(`确定删除 ${id} 吗？`)) {
      return;
    }

    // 调用 API
    await API.delete(`${url}/${id}`);

    // 刷新页面
    window.location.reload();
  } catch (error) {
    alert(`删除失败: ${error.message}`);
    console.error("删除失败", error);
  }
}
