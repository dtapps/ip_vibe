// 同步操作
function syncConfig(configId) {
  // 这里的 handleAction 是你项目全局封装好的通用请求函数
  if (typeof handleAction === "function") {
    handleAction({
      dataset: {
        url: `/api/cloudflare-tunnels/sync/${configId}`,
        method: "POST",
      },
    });
  }
}

function syncAllTunnels() {
  if (typeof handleAction === "function") {
    handleAction({
      dataset: {
        url: `/api/cloudflare-tunnels/sync-all`,
        method: "POST",
      },
    });
  }
}
