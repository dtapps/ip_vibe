// 定义全局变量，但不立即初始化
let tChart = null;
let cChart = null;

// 初始化图表实例
function initCharts() {
  const tDom = document.getElementById("trafficChart");
  const cDom = document.getElementById("connChart");
  if (tDom && !tChart) {
    tChart = echarts.init(tDom);
  }
  if (cDom && !cChart) {
    cChart = echarts.init(cDom);
  }
}

window.addEventListener("resize", () => {
  if (tChart) {
    tChart.resize();
  }
  if (cChart) {
    cChart.resize();
  }
});

async function loadStats() {
  const container = document.getElementById("stats-container");
  if (!container) {
    return;
  }

  const deviceId = container.dataset.deviceId;
  const range = document.getElementById("timeRange").value;

  try {
    const { data: stats } = await API.get(
      `/api/devices/${deviceId}/port-forward/stats`,
      {
        range,
      },
    );
    render(stats || []);
  } catch (err) {
    console.error("Fetch stats error:", err);
  }
}

function render(rawData) {
  // 1. 获取数据源
  const rates = rawData.rates || [];
  const latest = rawData.latest || {};
  const delta = rawData.delta || {};

  // 2. 更新卡片数值
  document.getElementById("stat-conn").innerText = latest.active_conn || 0;
  document.getElementById("stat-bytes-in").innerText = formatBytes(
    latest.bytes_in || 0,
  );
  document.getElementById("stat-bytes-out").innerText = formatBytes(
    latest.bytes_out || 0,
  );
  document.getElementById("stat-errors").innerText = latest.errors || 0;

  // 3. 更新 Delta 增量文字
  const updateDelta = (id, text) => {
    const el = document.getElementById(id);
    if (el) el.innerText = text;
  };
  updateDelta(
    "delta-bytes-in",
    `本时段: ${formatBytes(delta.bytes_in_incr || 0)}`,
  );
  updateDelta(
    "delta-bytes-out",
    `本时段: ${formatBytes(delta.bytes_out_incr || 0)}`,
  );
  updateDelta("delta-errors", `本时段: +${delta.errors_incr || 0} 错误`);

  // 4. 处理图表展示逻辑
  if (rates.length === 0) {
    // 【直接提示方案】：直接在图表中心显示文字
    const noDataOption = {
      title: {
        text: "暂无统计数据",
        left: "center",
        top: "middle",
        textStyle: {
          color: "#9ca3af", // daisyUI 风格的浅灰色
          fontSize: 14,
          fontWeight: "normal",
        },
      },
      xAxis: { show: false }, // 没数据时隐藏轴线
      yAxis: { show: false },
      series: [],
    };

    tChart.setOption(noDataOption, true); // true 表示完全重置
    cChart.setOption(noDataOption, true);
    return;
  }

  // 5. 处理时间轴
  const xAxisData = rates.map((r) => {
    return new Date(r.timestamp).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  });

  // 6. 渲染流量速率图
  tChart.setOption({
    tooltip: {
      trigger: "axis",
      formatter: (params) => {
        let res = `${params[0].name}<br/>`;
        params.forEach((p) => {
          res += `${p.marker}${p.seriesName}: <b>${p.value.toFixed(3)} KB/s</b><br/>`;
        });
        return res;
      },
    },
    xAxis: { type: "category", data: xAxisData, boundaryGap: false },
    yAxis: { type: "value", name: "KB/s", scale: true },
    series: [
      {
        name: "入站速率",
        data: rates.map((r) => r.in_kbps), // 对应后端 InKBps
        type: "line",
        smooth: true,
        showSymbol: false,
        areaStyle: { opacity: 0.1 },
        itemStyle: { color: "#10b981" },
      },
      {
        name: "出站速率",
        data: rates.map((r) => r.out_kbps), // 对应后端 OutKBps
        type: "line",
        smooth: true,
        showSymbol: false,
        areaStyle: { opacity: 0.1 },
        itemStyle: { color: "#3b82f6" },
      },
    ],
  });

  // 7. 渲染连接数图
  cChart.setOption({
    tooltip: { trigger: "axis" },
    xAxis: { type: "category", data: xAxisData },
    yAxis: { type: "value" },
    series: [
      {
        name: "活跃连接",
        data: rates.map((r) => r.active_conn),
        type: "bar",
        itemStyle: { color: "#f59e0b" },
      },
    ],
  });
}

// 辅助函数
function formatBytes(b) {
  if (b === 0) return "0 B";
  const i = Math.floor(Math.log(b) / Math.log(1024));
  return (
    (b / Math.pow(1024, i)).toFixed(2) + " " + ["B", "KB", "MB", "GB", "TB"][i]
  );
}

// 确保 DOM 加载后运行（优化版：统一await）
(async () => {
  // 如果DOM未加载，等待它
  if (document.readyState === "loading") {
    await new Promise((resolve) => {
      document.addEventListener("DOMContentLoaded", resolve, { once: true });
    });
  }

  // 现在DOM已就绪，加载并初始化
  try {
    await loadECharts(); // 等待ECharts加载
    initCharts(); // 初始化图表
    loadStats(); // 加载统计数据
    console.log("ECharts 初始化完成");
  } catch (error) {
    console.error("ECharts 加载/初始化失败:", error);
  }
})();

setInterval(loadStats, 30000); // 30s 自动刷新
