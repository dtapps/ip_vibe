const consoleBox = document.getElementById("logConsole");
if (consoleBox && document.getElementById("autoScroll").checked) {
  consoleBox.scrollTop = consoleBox.scrollHeight;
}

function viewDetailFromData(btn) {
  try {
    // 1. 从 data-detail 属性中直接读取解析好的 JSON
    const d = JSON.parse(btn.dataset.detail);

    // 2. 填充基础状态条
    document.getElementById("detailId").innerText = d.id;
    document.getElementById("detailMethod").innerText = d.method;
    document.getElementById("detailLatency").innerText =
      (d.elapse_time || 0) + "ms";

    const statusEl = document.getElementById("detailStatus");
    statusEl.innerText = d.status_code;
    statusEl.className = `badge badge-sm font-bold ${
      d.status_code < 400 ? "badge-success" : "badge-error"
    }`;

    // 3. 填充内容
    // 这里的 d.request_body 在后端已经是 json.RawMessage，JS 解析后可能是对象或字符串
    document.getElementById("detailReqBody").value = formatJSON(d.request_body);
    document.getElementById("detailResBody").value = formatJSON(
      d.response_body
    );

    // 4. 打开弹窗
    showModal("viewModal");
  } catch (err) {
    alert(`解析详情数据失败: ${err.message}`);
    console.error("解析详情数据失败:", err);
  }
}

function formatJSON(val) {
  if (!val) return "// 无数据";
  // 如果 val 已经是对象，直接格式化；如果是字符串且是 JSON，解析后再格式化
  try {
    if (typeof val === "object") {
      return JSON.stringify(val, null, 4);
    }
    const obj = JSON.parse(val);
    return JSON.stringify(obj, null, 4);
  } catch (e) {
    return val;
  }
}
