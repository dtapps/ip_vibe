// 启动调度器
async function startTaskScheduler() {
  if (confirm("确定启动调度器？")) {
    try {
      await API.post(`/api/tasks/start-scheduler`);
      // 刷新页面
      location.reload();
    } catch (error) {
      alert(`启动任务失败: ${error.message}`);
      console.error("启动任务失败", error);
    }
  }
}

// 停止调度器
async function stopTaskScheduler() {
  if (confirm("确定停止调度器？")) {
    try {
      await API.post(`/api/tasks/stop-scheduler`);
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

// 选择 任务类型 job_type
function onJobTypeSelect(el) {
  console.log("onJobTypeSelect", el);

  const job_type = el.value.trim();

  switch (job_type) {
    case "cron":
      $("#cron_div").show();
      $("#duration_div").hide();
      $("#random_div").hide();
      break;
    case "duration":
      $("#cron_div").hide();
      $("#duration_div").show();
      $("#random_div").hide();
      break;
    case "random":
      $("#cron_div").hide();
      $("#duration_div").hide();
      $("#random_div").show();
      break;
    default:
      $("#cron_div").hide();
      $("#duration_div").hide();
      $("#random_div").hide();
      break;
  }
}

/**
 * 打开 表单 弹窗
 * @param {HTMLButtonElement} btn 调用者，按钮元素
 * @returns
 */
async function openFormModal(btn) {
  const form = document.getElementById("form");
  if (!form) {
    throw new Error("未找到表单");
  }

  // 重置表单
  form.reset();

  try {
    // 加载 jQuery 库
    await loadJQuery();
    // 表单elements集合
    const els = form.elements;

    // 获取并解析数据
    const detail = btn.dataset.detail;
    const d = JSON.parse(detail) || {};
    console.log("d", d);

    // 编辑

    els.id.value = d.id;
    els.id.disabled = true;
    els.job_type.value = d.job_type;
    els.cron_expr.value = d.cron_expr || "";
    els.interval_value.value = d.interval_value;
    els.interval_unit.value = d.interval_unit;
    els.min_interval.value = d.min_interval;
    els.max_interval.value = d.max_interval;
    els.interval_unit.value = d.interval_unit;
    els.is_singleton.checked = d.is_singleton;
    els.run_on_start.checked = d.run_on_start;

    switch (d.job_type) {
      case "cron":
        $("#cron_div").show();
        $("#duration_div").hide();
        $("#random_div").hide();
        break;
      case "duration":
        $("#cron_div").hide();
        $("#duration_div").show();
        $("#random_div").hide();
        break;
      case "random":
        $("#cron_div").hide();
        $("#duration_div").hide();
        $("#random_div").show();
        break;
      default:
        $("#cron_div").hide();
        $("#duration_div").hide();
        $("#random_div").hide();
        break;
    }

    showModal("formModal", "modalTitle", `编辑任务: ${d.name}`);
  } catch (err) {
    alert(`打开 表单 失败: ${err.message}`);
    console.error(err);
    return;
  }
}

// 表单提交
const form = document.getElementById("form");
if (form) {
  form.onsubmit = async (e) => {
    try {
      // 阻止事件默认行为
      e.preventDefault();

      // 表单elements集合
      const els = form.elements;

      // 表单数据
      const form_id = Number(els.id.value.trim());
      const form_data = {
        job_type: els.job_type.value.trim(),
        cron_expr: els.cron_expr.value.trim(),
        interval_value: Number(els.interval_value.value.trim()),
        interval_unit: els.interval_unit.value.trim(),
        min_interval: Number(els.min_interval.value.trim()),
        max_interval: Number(els.max_interval.value.trim()),
        interval_unit: els.interval_unit.value.trim(),
        is_singleton: els.is_singleton.checked.trim(),
        run_on_start: els.run_on_start.checked.trim(),
      };
      console.log("form_id, form_data", form_id, form_data);

      // 校验
      if (form_data.job_type === "cron") {
        if (form_data.cron_expr === "") {
          return alert("请输入cron表达式");
        }
      }
      if (form_data.job_type === "duration") {
        if (
          isNaN(form_data.interval_value) ||
          form_data.interval_value <= 0 ||
          form_data.interval_unit === ""
        ) {
          return alert("请输入间隔值");
        }
      }
      if (form_data.job_type === "random") {
        if (
          isNaN(form_data.min_interval) ||
          form_data.min_interval <= 0 ||
          isNaN(form_data.max_interval) ||
          form_data.max_interval <= 0 ||
          form_data.interval_unit === ""
        ) {
          return alert("请输入随机间隔值");
        }
      }

      // 提交
      if (isNaN(form_id) || form_id <= 0) {
        await API.post("/api/tasks", form_data);
      } else {
        await API.patch(`/api/tasks/${form_id}`, form_data);
      }

      // 刷新页面
      location.reload();
    } catch (err) {
      alert(`操作失败: ${err.message}`);
      console.error(err);
    }
  };
}
