// CDN配置（同之前，保持不变）
const CDN_CONFIG = {
  clipboard: [
    "https://esa-cdnjs.dtapp.net/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://cdn.jsdelivr.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://unpkg.com/npm/clipboard@2.0.11/dist/clipboard.min.js",
  ],
  jquery: [
    "https://esa-cdnjs.dtapp.net/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://unpkg.com/npm/jquery@3.7.1/dist/jquery.min.js",
  ],
  echarts: [
    "https://esa-cdnjs.dtapp.net/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://cdn.jsdelivr.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://unpkg.com/npm/echarts@6.0.0/dist/echarts.min.js",
  ],
  "select2-js": [
    "https://esa-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://unpkg.com/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
  ],
  "select2-css": [
    "https://esa-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://esa-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://eo-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://eo-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://unpkg.com/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
  ],
};

// 通用加载函数（添加超时支持）
function loadLibrary(libName, urls, options = {}) {
  const { timeout = 5000 } = options; // 默认超时5秒，可在调用时覆盖，如 { timeout: 3000 }

  return new Promise((resolve, reject) => {
    let currentIndex = 0;
    let loadPromise; // 保存当前加载的Promise

    function tryLoad() {
      if (currentIndex >= urls.length) {
        reject(new Error(`${libName} 所有CDN加载失败`));
        return;
      }

      const url = urls[currentIndex];
      console.log(`尝试加载 ${libName} 从: ${url}`);

      // 创建loadjs的Promise（使用returnPromise: true）
      const jsLoadPromise = new Promise((jsResolve, jsReject) => {
        loadjs(url, {
          ...options,
          returnPromise: true, // 启用Promise模式
          success: () => {
            console.log(`${libName} 加载成功 (CDN: ${currentIndex + 1})`);
            jsResolve();
          },
          error: (pathsNotFound) => {
            console.warn(
              `${libName} 从 ${url} 加载失败（非超时原因），切换下一个CDN`
            );
            jsReject(new Error(`Load error: ${pathsNotFound}`));
          },
        });
      });

      // 超时Promise
      const timeoutPromise = new Promise((_, timeoutReject) => {
        setTimeout(() => {
          console.warn(
            `${libName} 从 ${url} 超时（${timeout}ms），切换下一个CDN`
          );
          timeoutReject(new Error(`Timeout after ${timeout}ms`));
        }, timeout);
      });

      // 竞争：loadjs成功/失败 vs 超时
      loadPromise = Promise.race([jsLoadPromise, timeoutPromise])
        .then(() => {
          // 成功：取消任何pending超时（如果有）
          if (loadPromise && loadPromise.cancel) loadPromise.cancel();
          resolve();
        })
        .catch((error) => {
          // 失败：取消pending loadjs（可选，浏览器级）
          if (loadPromise && loadPromise.cancel) loadPromise.cancel();
          currentIndex++;
          tryLoad(); // 切换下一个
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

async function loadJQuery() {
  if (typeof $ !== "undefined") {
    console.log("jQuery 已加载");
    return Promise.resolve();
  }
  return loadLibrary("jquery", CDN_CONFIG.jquery);
}

async function loadECharts() {
  if (typeof echarts !== "undefined") {
    console.log("ECharts 已加载");
    return Promise.resolve();
  }
  return loadLibrary("echarts", CDN_CONFIG.echarts);
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
