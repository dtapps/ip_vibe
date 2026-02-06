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

// 打开 下次执行时间 弹窗
async function onNextRunPopup(id, name) {
  try {
    const { data: task } = await API.get(`/api/gocrons/${id}`);
    console.log("task", task);
    if (!task) {
      alert("任务数据异常");
      return;
    }
    if (task.length === 0) {
      alert("当前任务执行时间无数据");
      return;
    }

    // 显示弹窗内容
    showPopupContent(
      "popup",
      `实时 执行时间 - ${name}`,
      task.next_runs.join("\n"),
      task.next_runs.length,
    );
  } catch (error) {
    alert(`加载任务执行时间失败: ${error.message}`);
    console.error("查看任务执行时间失败", error);
  }
}
