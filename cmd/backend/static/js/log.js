/**
 * 从按钮绑定的 dataset 中解析并展示详情
 * @param {HTMLElement} btn 触发详情的按钮对象
 */
function viewDetailFromData(btn) {
  try {
    // 1. 获取并解析数据
    if (!btn.dataset.detail) throw new Error("未找到详情数据");
    const d = JSON.parse(btn.dataset.detail);

    // 2. 填充顶部元数据
    document.getElementById("detailId").innerText = d.id;
    document.getElementById("detailSource").innerText = d.source;
    document.getElementById("detailOpType").innerText = d.op_type;
    document.getElementById("detailObjType").innerText = d.obj_type;
    document.getElementById("detailObjID").innerText = d.obj_id;

    // 3. 状态码样式映射
    const statusEl = document.getElementById("detailStatus");
    statusEl.innerText = d.status;
    statusEl.className = `badge badge-sm font-bold ${
      d.status === "success" ? "badge-success" : "badge-error"
    }`;

    // 4. 环境与版本
    document.getElementById("verGo").innerText = `GO: ${d.go_version || "-"}`;

    // 5. 填充 Body Textarea
    document.getElementById("detailOldValue").value = formatJSON(d.old_value);
    document.getElementById("detailNewValue").value = formatJSON(d.new_value);

    // 6. 打开 Modal (调用通用方法)
    showModal("viewModal");
  } catch (err) {
    console.error("View detail failed:", err);
    alert("详情解析失败: " + err.message);
  }
}
