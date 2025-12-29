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
