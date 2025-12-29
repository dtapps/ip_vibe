// 定义全局变量，但不立即初始化
let tChart = null;
let cChart = null;

// 初始化图表实例
function initCharts() {
  const tDom = document.getElementById("trafficChart");
  const cDom = document.getElementById("connChart");

  // 增加判断：如果已经初始化过了就不要重复初始化
  if (tDom && !tChart) {
    tChart = echarts.init(tDom);
  }
  if (cDom && !cChart) {
    cChart = echarts.init(cDom);
  }
}

window.addEventListener("resize", () => {
  // 增加判断：如果图表实例不存在，就不调用resize方法
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

  const pfId = container.dataset.ruleId;
  const range = document.getElementById("timeRange").value;

  try {
    const { data: stats } = await API.get(`/api/port-forwards/${pfId}/stats`, {
      range: range,
    });
    render(stats || []);
  } catch (err) {
    console.error("Fetch stats error:", err);
  }
}

function render(data) {
  if (!data || data.length < 2) {
    console.warn("数据量不足，无法计算速率");
    return;
  }

  // 【关键修复 1】确保数据按时间正序排列（从旧到新）
  // 你的 API 返回的是最新的在前面，我们需要反转它
  const sortedData = [...data].sort(
    (a, b) => new Date(a.created_at) - new Date(b.created_at)
  );

  // 1. 处理横坐标
  const xAxisData = sortedData.map((i) =>
    new Date(i.created_at).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    })
  );

  // 2. 计算速率
  const inSpeedArray = [];
  const outSpeedArray = [];

  for (let i = 0; i < sortedData.length; i++) {
    if (i === 0) {
      inSpeedArray.push(0);
      outSpeedArray.push(0);
      continue;
    }

    const curr = sortedData[i];
    const prev = sortedData[i - 1];

    // 计算时间差（秒）
    const sec = (new Date(curr.created_at) - new Date(prev.created_at)) / 1000;

    if (sec <= 0) {
      inSpeedArray.push(0);
      outSpeedArray.push(0);
      continue;
    }

    // 计算字节差并转为 KB/s
    let diffIn = curr.bytes_in - prev.bytes_in;
    let diffOut = curr.bytes_out - prev.bytes_out;

    // 防重启逻辑
    if (diffIn < 0) diffIn = 0;
    if (diffOut < 0) diffOut = 0;

    // 【优化】如果速率太小，可以保留更多小数位，或者直接显示 B/s
    const sIn = (diffIn / 1024 / sec).toFixed(3);
    const sOut = (diffOut / 1024 / sec).toFixed(3);

    inSpeedArray.push(sIn);
    outSpeedArray.push(sOut);
  }

  // 3. 更新卡片 (使用原始数据的第一项，即最新的一项)
  const last = data[0];
  document.getElementById("stat-conn").innerText = last.active_conn;
  document.getElementById("stat-bytes-in").innerText = formatBytes(
    last.bytes_in
  );
  document.getElementById("stat-bytes-out").innerText = formatBytes(
    last.bytes_out
  );
  document.getElementById("stat-errors").innerText = last.errors;

  // 4. 渲染图表
  tChart.setOption({
    tooltip: {
      trigger: "axis",
      formatter: (params) => {
        let res = `${params[0].name}<br/>`;
        params.forEach((p) => {
          res += `${p.marker}${p.seriesName}: <b>${p.value} KB/s</b><br/>`;
        });
        return res;
      },
    },
    xAxis: { type: "category", data: xAxisData, boundaryGap: false },
    yAxis: {
      type: "value",
      name: "KB/s",
      min: 0,
      scale: true, // 自动伸缩坐标轴
    },
    series: [
      {
        name: "入站速率",
        data: inSpeedArray,
        type: "line",
        smooth: true,
        showSymbol: false,
        areaStyle: { opacity: 0.1 },
        itemStyle: { color: "#10b981" },
      },
      {
        name: "出站速率",
        data: outSpeedArray,
        type: "line",
        smooth: true,
        showSymbol: false,
        areaStyle: { opacity: 0.1 },
        itemStyle: { color: "#3b82f6" },
      },
    ],
  });

  cChart.setOption({
    xAxis: { type: "category", data: xAxisData },
    yAxis: { type: "value" },
    series: [
      {
        name: "活跃连接",
        data: sortedData.map((i) => i.active_conn),
        type: "bar",
        itemStyle: { color: "#f59e0b" },
      },
    ],
  });
}

// 辅助函数
function formatBytes(bytes) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

// 确保 DOM 加载后运行
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    initCharts();
    loadStats();
  });
} else {
  initCharts();
  loadStats();
}

setInterval(loadStats, 30000); // 30s 自动刷新
