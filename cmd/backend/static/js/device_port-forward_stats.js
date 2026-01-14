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
      }
    );
    render(stats || []);
  } catch (err) {
    console.error("Fetch stats error:", err);
  }
}

function render(rawData) {
  if (!rawData || rawData.length === 0) {
    console.warn("数据量不足，无法计算速率");
    return;
  }

  // 1. 数据聚合：按 created_at 分组累加所有规则的数据
  const grouped = {};
  rawData.forEach((item) => {
    const time = item.created_at;
    if (!grouped[time]) {
      grouped[time] = {
        created_at: time,
        active_conn: 0,
        bytes_in: 0,
        bytes_out: 0,
        errors: 0,
      };
    }
    grouped[time].active_conn += item.active_conn;
    grouped[time].bytes_in += item.bytes_in;
    grouped[time].bytes_out += item.bytes_out;
    grouped[time].errors += item.errors;
  });

  // 2. 转换为有序数组并按时间正序排列
  const data = Object.values(grouped).sort(
    (a, b) => new Date(a.created_at) - new Date(b.created_at)
  );

  const xAxisData = data.map((i) =>
    new Date(i.created_at).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    })
  );

  // 3. 计算速率
  const inSpeed = [];
  const outSpeed = [];
  data.forEach((curr, i) => {
    if (i === 0) {
      inSpeed.push(0);
      outSpeed.push(0);
    } else {
      const prev = data[i - 1];
      const sec =
        (new Date(curr.created_at) - new Date(prev.created_at)) / 1000;
      if (sec <= 0) {
        inSpeed.push(0);
        outSpeed.push(0);
      } else {
        let dIn = curr.bytes_in - prev.bytes_in;
        let dOut = curr.bytes_out - prev.bytes_out;
        inSpeed.push(((dIn < 0 ? 0 : dIn) / 1024 / sec).toFixed(2));
        outSpeed.push(((dOut < 0 ? 0 : dOut) / 1024 / sec).toFixed(2));
      }
    }
  });

  // 4. 更新卡片 (取聚合后的最后一条数据)
  const last = data[data.length - 1];
  document.getElementById("stat-conn").innerText = last.active_conn;
  document.getElementById("stat-bytes-in").innerText = formatBytes(
    last.bytes_in
  );
  document.getElementById("stat-bytes-out").innerText = formatBytes(
    last.bytes_out
  );
  document.getElementById("stat-errors").innerText = last.errors;

  // 5. 渲染图表
  tChart.setOption({
    tooltip: { trigger: "axis" },
    legend: { data: ["入站总速率", "出站总速率"], bottom: 0 },
    xAxis: { type: "category", data: xAxisData, boundaryGap: false },
    yAxis: { type: "value", name: "KB/s" },
    series: [
      {
        name: "入站总速率",
        data: inSpeed,
        type: "line",
        smooth: true,
        itemStyle: { color: "#10b981" },
        areaStyle: { opacity: 0.1 },
      },
      {
        name: "出站总速率",
        data: outSpeed,
        type: "line",
        smooth: true,
        itemStyle: { color: "#3b82f6" },
        areaStyle: { opacity: 0.1 },
      },
    ],
  });

  cChart.setOption({
    xAxis: { type: "category", data: xAxisData },
    yAxis: { type: "value" },
    series: [
      {
        name: "并发汇总",
        data: data.map((i) => i.active_conn),
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
