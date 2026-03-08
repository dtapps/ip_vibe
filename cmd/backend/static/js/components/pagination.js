/**
 * 切换分页并持久化选择
 * @param {number} page 分页页码
 * @param {number} size 分页大小
 */
function changePage(page, size) {
  // 持久化存储到 localStorage
  localStorage.setItem("pref_page_size", size);

  // 存入 Cookie (让 Go 后端能读到)，设置有效期 30 天，路径全局有效
  document.cookie =
    encodeURIComponent("pref_page_size") +
    "=" +
    encodeURIComponent(size) +
    "; path=/; max-age=2592000";

  const url = new URL(location.href);
  url.searchParams.set("page", page);
  url.searchParams.set("page_size", size);

  // 执行跳转
  location.href = url.toString();
}

window.components = window.components || {};
if (!window.components.pagination) {
  window.components.pagination = { changePage };
}
