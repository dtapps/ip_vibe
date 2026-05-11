# 使用 golang 镜像作为基础
FROM golang:1.26.3-trixie

# 1. 替换 Debian APT 加速源
RUN sed -i 's|deb.debian.org|mirrors.cloud.tencent.com|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's|deb.debian.org|mirrors.cloud.tencent.com|g' /etc/apt/sources.list

# 2. 设置 Go 环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://mirrors.cloud.tencent.com/go/,direct

# 3. 安装 Node.js 及工具
RUN apt-get update && apt-get install -y curl gnupg && \
    # 获取 Node.js 最新版
    export NODEJS_ORG_MIRROR=https://mirrors.cloud.tencent.com/nodejs-release/ && \
    curl -fsSL https://deb.nodesource.com/setup_current.x | bash - && \
    apt-get install -y nodejs && \
    \
    # 4. 安装 pnpm 最新版并配置加速
    npm config set registry https://mirrors.cloud.tencent.com/npm/ && \
    npm install -g pnpm && \
    pnpm config set registry https://mirrors.cloud.tencent.com/npm/ && \
    \
    # 5. 解决 pnpm v10 脚本拦截问题 (在构建镜像时就允许脚本)
    pnpm config set ignore-scripts false && \
    \
    # 清理
    apt-get clean && rm -rf /var/lib/apt/lists/*

# 确保路径全局可用
ENV PATH="/usr/local/bin:$PATH"

CMD ["bash"]