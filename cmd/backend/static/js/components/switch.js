/**
 * 切换开关状态
 * @param {string} url 接口地址
 * @param {string} dataKey 数据键名
 * @param {string} isNoReload 是否不刷新页面
 * @param {HTMLInputElement} element 开关元素
 */
async function changeStatus(url, dataKey, isNoReload, element) {
  try {
    if (typeof API !== "undefined" && typeof API.patch === "function") {
      // 通用 API PATCH 动作
      await API.patch(url, { [dataKey]: element.checked });
      // 根据 isNoReload 决定是否刷新，转换为布尔值
      if (isNoReload != "false") {
        location.reload();
      }
    } else {
      alert("API.patch 函数未定义，请确保引入了必要的 JS 库");
    }
  } catch (error) {
    console.error("API失败", error);
    alert(`状态更新失败: ${error.message}`);
    // 恢复 checkbox 状态
    element.checked = !element.checked;
  }
}

window.components = window.components || {};
if (!window.components.switch) window.components.switch = { changeStatus };
