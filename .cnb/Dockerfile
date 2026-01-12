# 使用你指定的 golang 镜像作为基础
FROM golang:1.25.5-trixie

# 1. 替换 Debian APT 加速源
RUN sed -i 's|deb.debian.org|mirrors.cloud.tencent.com|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's|deb.debian.org|mirrors.cloud.tencent.com|g' /etc/apt/sources.list

# 2. 设置 Go 环境变量
export GO111MODULE=on
export GOPROXY=https://mirrors.cloud.tencent.com/go,direct

# 3. 安装 Node.js 和基础工具
RUN apt-get update && apt-get install -y curl gnupg && \
    export NODEJS_ORG_MIRROR=https://mirrors.cloud.tencent.com/nodejs-release/ && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    # 3. 安装 pnpm 并配置加速
    npm config set registry https://mirrors.cloud.tencent.com/npm/ && \
    npm install -g pnpm && \
    pnpm config set registry https://mirrors.cloud.tencent.com/npm/ && \
    # 4. 清理缓存减小镜像体积
    apt-get clean && rm -rf /var/lib/apt/lists/*

# 确保 pnpm 在 PATH 中
ENV PATH="/usr/local/bin:$PATH"

CMD ["bash"]