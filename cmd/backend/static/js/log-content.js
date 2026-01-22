const consoleBox = document.getElementById("logConsole");
if (consoleBox && document.getElementById("autoScroll").checked) {
  consoleBox.scrollTop = consoleBox.scrollHeight;
}

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
    document.getElementById("detailHost").innerText = d.hostname || "N/A";
    document.getElementById("detailMethod").innerText = d.method;
    document.getElementById("detailLatency").innerText =
      (d.elapse_time || 0) + "ms";
    document.getElementById("detailTime").innerText = d.created_at
      ? new Date(d.created_at).toLocaleString()
      : "-";

    // 3. 状态码样式映射
    const statusEl = document.getElementById("detailStatus");
    statusEl.innerText = d.status_code;
    statusEl.className = `badge badge-sm font-bold ${
      d.status_code < 400 ? "badge-success" : "badge-error"
    }`;

    // 4. 环境与版本
    document.getElementById("verGo").innerText = `GO: ${d.go_version || "-"}`;
    document.getElementById("verPlugin").innerText = `PLUGIN: ${
      d.plugin_version || "-"
    }`;
    const restyBox = document.getElementById("verResty");
    if (d.resty_version) {
      restyBox.innerText = `RESTY: ${d.resty_version}`;
      restyBox.classList.remove("hidden");
    } else {
      restyBox.classList.add("hidden");
    }

    // 5. 处理 Headers 渲染与复制 JSON 注入
    const reqHeadersJson = JSON.stringify(d.request_headers || {}, null, 4);
    const resHeadersJson = JSON.stringify(d.response_headers || {}, null, 4);

    document.getElementById("detailReqHeaders").innerHTML = renderHeaders(
      d.request_headers,
    );
    document.getElementById("detailResHeaders").innerHTML = renderHeaders(
      d.response_headers,
    );

    // 注入复制属性
    document
      .getElementById("copyReqHeaders")
      .setAttribute("data-clipboard-text", reqHeadersJson);
    document
      .getElementById("copyResHeaders")
      .setAttribute("data-clipboard-text", resHeadersJson);

    // 6. 填充 Body Textarea
    document.getElementById("detailReqBody").value = formatJSON(d.request_body);
    document.getElementById("detailResBody").value = formatJSON(
      d.response_body,
    );

    // 7. 打开 Modal (调用通用方法)
    showModal("viewModal");
  } catch (err) {
    console.error("View detail failed:", err);
    alert("详情解析失败: " + err.message);
  }
}

/**
 * 格式化 Header Map 为 HTML 列表
 * @param {*} headers Header Map 对象
 * @returns 格式化后的 HTML 字符串
 */
function renderHeaders(headers) {
  if (!headers || Object.keys(headers).length === 0)
    return '<span class="opacity-30 italic">No Headers</span>';
  let html = "";
  for (const [key, values] of Object.entries(headers)) {
    html += `<div class="mb-1.5 leading-tight"><span class="text-primary/80 font-bold select-all">${key}:</span> <span class="opacity-70 break-all font-sans">${values.join(
      ", ",
    )}</span></div>`;
  }
  return html;
}
