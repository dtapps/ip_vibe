// CDN配置（同之前，保持不变）
const CDN_CONFIG = {
  "clipboard-js": [
    "/static/libs/clipboard/clipboard.min.js",
    "https://esa-cdnjs.dtapp.net/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js",
    "https://cdn.jsdelivr.net/npm/clipboard@2.0.11/dist/clipboard.min.js",
    "https://unpkg.com/npm/clipboard@2.0.11/dist/clipboard.min.js",
  ],
  jquery: [
    "/static/libs/jquery/jquery.min.js",
    "https://esa-cdnjs.dtapp.net/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js",
    "https://cdn.jsdelivr.net/npm/jquery@3.7.1/dist/jquery.min.js",
    "https://unpkg.com/npm/jquery@3.7.1/dist/jquery.min.js",
  ],
  echarts: [
    "/static/libs/echarts/echarts.min.js",
    "https://esa-cdnjs.dtapp.net/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/echarts/6.0.0/echarts.min.js",
    "https://cdn.jsdelivr.net/npm/echarts@6.0.0/dist/echarts.min.js",
    "https://unpkg.com/npm/echarts@6.0.0/dist/echarts.min.js",
  ],
  "select2-js": [
    "/static/libs/select2/js/select2.min.js",
    "https://esa-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://esa-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://eo-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://eo-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/js/select2.min.js",
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
    "https://unpkg.com/npm/select2@4.1.0-rc.0/dist/js/select2.min.js",
  ],
  "select2-css": [
    "/static/libs/select2/css/select2.min.css",
    "https://esa-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://esa-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://eo-cdnjs.dtapp.net/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://eo-jsdelivr.dtapp.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://cdnjs.cloudflare.com/ajax/libs/select2/4.1.0-rc.0/css/select2.min.css",
    "https://cdn.jsdelivr.net/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
    "https://unpkg.com/npm/select2@4.1.0-rc.0/dist/css/select2.min.css",
  ],
};

// 通用加载函数
function loadLibrary(libName, urls, options = {}) {
  // 默认超时5秒，可在调用时覆盖，如 { timeout: 3000 }
  const { timeout = 5000 } = options;

  return new Promise((resolve, reject) => {
    let currentIndex = 0;
    let loadPromise; // 保存当前加载的Promise

    function tryLoad() {
      if (currentIndex >= urls.length) {
        reject(new Error(`${libName} 加载所失败`));
        return;
      }

      const url = urls[currentIndex];
      console.log(`${libName} 尝试从 ${url} 加载`);

      // 创建loadjs的Promise（使用returnPromise: true）
      const jsLoadPromise = new Promise((jsResolve, jsReject) => {
        loadjs(url, {
          ...options,
          returnPromise: true, // 启用Promise模式
          success: () => {
            // console.log(`${libName} 加载成功 ${url}`);
            jsResolve();
          },
          error: (pathsNotFound) => {
            // console.warn(
            //   `${libName} 从 ${url} 加载失败（非超时原因），切换下一个`
            // );
            jsReject(new Error(`Load error: ${pathsNotFound}`));
          },
        });
      });

      // 超时Promise
      const timeoutPromise = new Promise((_, timeoutReject) => {
        setTimeout(() => {
          // console.warn(`${libName} 从 ${url} 超时（${timeout}ms），切换下一个`);
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

// 专用加载 ClipboardJS
async function loadClipboardJS() {
  const libName = "clipboard-js";

  console.log(`${libName} 尝试加载`, typeof ClipboardJS !== "undefined");
  try {
    if (typeof ClipboardJS !== "undefined") {
      console.log(`${libName} 加载已完成`);
      return Promise.resolve();
    }
    await loadLibrary(libName, CDN_CONFIG[libName]);
    console.log(`${libName} 加载成功`);
  } catch (error) {
    console.error(`${libName} 加载失败:`, error);
    throw error;
  }
}

// 专用加载 jQuery
async function loadJQuery() {
  const libName = "jquery";

  console.log(`${libName} 尝试加载`, typeof $ !== "undefined");
  try {
    if (typeof $ !== "undefined") {
      console.log(`${libName} 加载已完成`);
      return Promise.resolve();
    }
    await loadLibrary(libName, CDN_CONFIG[libName]);
    console.log(`${libName} 加载成功`);
  } catch (error) {
    console.error(`${libName} 加载失败:`, error);
    throw error;
  }
}

// 专用加载 ECharts
async function loadECharts() {
  const libName = "echarts";

  console.log(`${libName} 尝试加载`, typeof echarts !== "undefined");
  try {
    if (typeof echarts !== "undefined") {
      console.log(`${libName} 加载已完成`);
      return Promise.resolve();
    }
    await loadLibrary(libName, CDN_CONFIG[libName]);
    console.log(`${libName} 加载成功`);
  } catch (error) {
    console.error(`${libName} 加载失败:`, error);
    throw error;
  }
}

// 专用加载 Select2
async function loadSelect2() {
  const libName = "select2";

  console.log(`${libName} 尝试加载`, typeof $.fn.select2 !== "undefined");
  try {
    if (typeof $.fn.select2 !== "undefined") {
      console.log(`${libName} 加载已完成`);
      return Promise.resolve();
    }
    // 并行加载JS和CSS
    await Promise.all([
      loadLibrary(`${libName}-js`, CDN_CONFIG[`${libName}-js`]),
      loadLibrary(`${libName}-css`, CDN_CONFIG[`${libName}-css`]),
    ]);
    console.log(`${libName} 加载成功`);
  } catch (error) {
    console.error(`${libName} 加载失败:`, error);
    throw error;
  }
}
