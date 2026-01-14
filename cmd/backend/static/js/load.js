// CDN配置（同之前，保持不变）
const CDN_CONFIG = {
  clipboard: [
    "https://cdn.jsdelivr.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
  ],
  echarts: [
    "https://cdn.jsdelivr.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/echarts/6.0.0/echarts.min.js",
  ],
  jquery: [
    "https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js",
  ],
  "select2-js": [
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
  ],
  "select2-css": [
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
  ],
};

// 通用加载函数（同之前，支持fallback，返回Promise）
function loadLibrary(libName, urls, options = {}) {
  return new Promise((resolve, reject) => {
    let currentIndex = 0;

    function tryLoad() {
      if (currentIndex >= urls.length) {
        reject(new Error(`${libName} 所有CDN加载失败`));
        return;
      }

      const url = urls[currentIndex];
      console.log(`尝试加载 ${libName} 从: ${url}`);

      loadjs(url, {
        ...options,
        success: () => {
          console.log(`${libName} 加载成功 (CDN: ${currentIndex + 1})`);
          resolve();
        },
        error: (pathsNotFound) => {
          console.warn(`${libName} 从 ${url} 加载失败，切换下一个CDN`);
          currentIndex++;
          tryLoad();
        },
      });
    }

    tryLoad();
  });
}

// 每个库的专用加载函数（按需调用）
async function loadClipboard() {
  if (typeof Clipboard !== "undefined") {
    console.log("Clipboard 已加载");
    return Promise.resolve();
  }
  return loadLibrary("clipboard", CDN_CONFIG.clipboard);
}

async function loadECharts() {
  if (typeof echarts !== "undefined") {
    console.log("ECharts 已加载");
    return Promise.resolve();
  }
  return loadLibrary("echarts", CDN_CONFIG.echarts);
}

async function loadJQuery() {
  if (typeof $ !== "undefined") {
    console.log("jQuery 已加载");
    return Promise.resolve();
  }
  return loadLibrary("jquery", CDN_CONFIG.jquery);
}

async function loadSelect2() {
  try {
    // 先确保jQuery就绪
    await loadJQuery();
    if (typeof $.fn.select2 !== "undefined") {
      console.log("Select2 已加载");
      return Promise.resolve();
    }
    // 并行加载JS和CSS
    await Promise.all([
      loadLibrary("select2-js", CDN_CONFIG["select2-js"]),
      loadLibrary("select2-css", CDN_CONFIG["select2-css"]),
    ]);
    console.log("Select2 加载完成");
  } catch (error) {
    console.error("Select2 加载失败:", error);
    throw error;
  }
}
