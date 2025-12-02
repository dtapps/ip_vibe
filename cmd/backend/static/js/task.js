// 启动全部任务
async function startTask() {
  if (confirm("确定启动全部任务？")) {
    try {
      await API.post(`/api/tasks/start`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`启动任务失败: ${error.message}`);
      console.error("启动任务失败", error);
    }
  }
}

// 停止全部任务
async function stopTask() {
  if (confirm("确定停止全部任务？")) {
    try {
      await API.post(`/api/tasks/stop`);
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
      await API.post(`/api/tasks/clean`);
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
      await API.post(`/api/tasks/${id}`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`运行任务失败: ${error.message}`);
      console.error("运行任务失败", error);
    }
  }
}

// 实时 执行时间 弹窗
async function viewLiveExecutionTime(id, name) {
  try {
    const { data: task } = await API.get(`/api/tasks/${id}`);
    console.log("task", task);
    if (!task) {
      alert("任务数据异常");
      return;
    }
    if (task.length === 0) {
      alert("当前任务执行时间无数据");
      return;
    }

    // 填充代码块
    const pre = document.getElementById("liveExecutionTimePre");

    pre.textContent = task.next_runs.join("\n");

    // 更新计数
    document.getElementById("liveExecutionTimeCount").textContent =
      task.next_runs.length;

    // 打开弹窗（统一走你已有的 openModal 即可）
    openModal(
      "liveExecutionTimeModal",
      "liveExecutionTimeTitle",
      `实时 执行时间 - ${name}`
    );
  } catch (error) {
    alert(`加载任务执行时间失败: ${error.message}`);
    console.error("查看任务执行时间失败", error);
  }
}
