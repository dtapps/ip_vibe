// 删除
async function onDelete(id) {
  if (confirm("确定删除此通知订阅？")) {
    await withErrorHandling(
      async () => {
        await API.delete(`/api/webhook-subs/${id}`);
        // 刷新页面
        location.reload();
      },
      {
        errorMessage: "删除失败",
      }
    );
  }
}
