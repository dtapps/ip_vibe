// 启动调度器
async function startScheduler() {
  if (confirm("确定启动调度器？")) {
    try {
      await API.post(`/api/gocrons/start-scheduler`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`启动任务失败: ${error.message}`);
      console.error("启动任务失败", error);
    }
  }
}

// 停止调度器
async function stopScheduler() {
  if (confirm("确定停止调度器？")) {
    try {
      await API.post(`/api/gocrons/stop-scheduler`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`停止任务失败: ${error.message}`);
      console.error("停止任务失败", error);
    }
  }
}

// 清理全部任务
async function cleanTask() {
  if (confirm("确定清理全部任务？")) {
    try {
      await API.post(`/api/gocrons/clean`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`清理任务失败: ${error.message}`);
      console.error("清理任务失败", error);
    }
  }
}

// 运行某任务
async function runTask(id) {
  if (confirm("确定运行此任务？")) {
    try {
      await API.post(`/api/gocrons/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`运行任务失败: ${error.message}`);
      console.error("运行任务失败", error);
    }
  }
}

/**
 * 打开 执行时间 弹窗
 * @param {HTMLButtonElement} btn 按钮元素
 * @returns
 */
async function showExecutionTimePopup(btn) {
  try {
    // 加载 Day.js 库
    // await loadDayJS();

    // 获取并解析数据
    const detail = btn.dataset.detail;
    const d = JSON.parse(detail);

    // 解析数据
    const { name, execution_time = [] } = d;

    if (execution_time.length === 0) {
      alert("当前任务执行时间无数据");
      return;
    }

    // 转换时间数组为带相对时间的格式
    const formattedRuns = execution_time.map((timeStr) => {
      try {
        // 解析时间字符串
        const time = new Date(timeStr);

        // 检查是否解析成功
        if (isNaN(time.getTime())) {
          return `${timeStr} | 格式错误`;
        }

        // 格式化为 YYYY-MM-DD HH:mm:ss
        const formattedTime = time
          .toLocaleString("zh-CN", {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
            hour12: false,
          })
          .replace(/\//g, "-")
          .replace(",", "");

        // 拼接格式：原始时间 | 相对时间
        // return `${timeStr} | ${diffForHumans(timeStr)}`;
        return `${formattedTime} | ${diffForHumans(timeStr)}`;
      } catch (err) {
        // 解析失败时显示原始时间
        return `${timeStr} | 解析失败`;
      }
    });

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `执行时间 - ${name}`,
      formattedRuns.join("\n"),
      formattedRuns.length,
    );
  } catch (err) {
    alert(`打开 执行时间 弹窗失败: ${err.message}`);
    console.error(err);
  }
}
