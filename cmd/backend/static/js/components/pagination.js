/**
 * 切换分页
 * @param {number} page 分页页码
 * @param {number} size 分页大小
 */
function changePage(page, size) {
  const url = new URL(location.href);
  url.searchParams.set("page", page);
  url.searchParams.set("page_size", size);
  location.href = url.toString();
}

window.components = window.components || {};
if (!window.components.pagination) {
  window.components.pagination = { changePage };
}
