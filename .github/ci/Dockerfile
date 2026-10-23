# 使用 golang 镜像作为基础
FROM golang:1.27.3-trixie

# ==========================================
# 定义版本变量
# https://github.com/upx/upx/releases
# https://github.com/golangci/golangci-lint/releases
# https://github.com/nodejs/node/releases
# https://github.com/pnpm/pnpm/releases
# https://github.com/goreleaser/nfpm/releases
# https://github.com/gotestyourself/gotestsum/releases
# https://github.com/Kunde21/markdownfmt/releases
# ==========================================
ARG UPX_VERSION=5.2.1
ARG GOLANGCI_LINT_VERSION=v2.15.0
ARG NODE_VERSION=26.12.1
ARG PNPM_VERSION=12.15.0
ARG NFPM_VERSION=2.48.0
ARG GOTESTSUM_VERSION=v1.13.0
ARG MARKDOWNFMT_VERSION=v0.1.6

# 1. 设置 Go 环境变量
ENV GOINSECURE=go.dtapp.net
ENV GOPATH=/go
ENV PATH="$GOPATH/bin:$PATH"

# 2. 安装基础构建工具
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    xz-utils \
    ca-certificates \
    curl \
    gnupg \
    zip \
    git \
    make \
    unzip

# 4. 安装 Node.js 工具
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then NODE_ARCH="x64"; \
    elif [ "$ARCH" = "aarch64" ]; then NODE_ARCH="arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    \
    # 构建下载地址
    NODE_DIR="node-v${NODE_VERSION}-linux-${NODE_ARCH}" && \
    NODE_URL="https://nodejs.org/dist/v${NODE_VERSION}/${NODE_DIR}.tar.xz" && \
    \
    cd /tmp && \
    wget -q -O node.tar.xz "${NODE_URL}" && \
    tar -xJf node.tar.xz -C /usr/local --strip-components=1 && \
    \
    # 配置 npm/pnpm
    npm install -g pnpm@${PNPM_VERSION} && \
    \
    # 清理
    rm -rf node.tar.xz && \
    echo "--- Node.js installed ---" && \
    node -v && npm -v && pnpm -v

# 5. 安装 UPX 压缩工具
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then UPX_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then UPX_ARCH="arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    \
    # 构建下载地址
    UPX_DIR="upx-${UPX_VERSION}-${UPX_ARCH}_linux" && \
    UPX_URL="https://github.com/upx/upx/releases/download/v${UPX_VERSION}/${UPX_DIR}.tar.xz" && \
    \
    cd /tmp && \
    wget -q -O upx.tar.xz "${UPX_URL}" && \
    tar -xf upx.tar.xz && \
    mv ${UPX_DIR}/upx /usr/local/bin/upx && \
    chmod +x /usr/local/bin/upx && \
    # 清理
    rm -rf upx.tar.xz ${UPX_DIR} && \
    echo "--- UPX installed ---" && \
    upx --version

# 6. 安装 nfpm (用于打包 deb/rpm)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then NFPM_ARCH="x86_64"; \
    elif [ "$ARCH" = "aarch64" ]; then NFPM_ARCH="arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    \
    NFPM_URL="https://github.com/goreleaser/nfpm/releases/download/v${NFPM_VERSION}/nfpm_${NFPM_VERSION}_Linux_${NFPM_ARCH}.tar.gz" && \
    \
    cd /tmp && \
    wget -q -O nfpm.tar.gz "${NFPM_URL}" && \
    tar -xf nfpm.tar.gz && \
    mv nfpm /usr/local/bin/nfpm && \
    chmod +x /usr/local/bin/nfpm && \
    rm -rf nfpm.tar.gz && \
    echo "--- nfpm installed ---" && \
    nfpm --version

# 7. 安装 yamllint (用于 YAML 文件检查)
RUN apt-get update && apt-get install -y --no-install-recommends yamllint \
    && echo "--- yamllint installed ---" \
    && yamllint --version

# 8. 安装 wixl (用于编译 Windows MSI)
RUN apt-get update && apt-get install -y --no-install-recommends wixl wixl-data \
    && echo "--- wixl installed ---" \
    && wixl --version

# 9. 预安装 Go 工具链（版本已锁定）
RUN echo "==> 预安装 Go 工具链（版本已锁定）..." \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && echo "==> golangci-lint 安装完成 (${GOLANGCI_LINT_VERSION})" \
    && go install gotest.tools/gotestsum@${GOTESTSUM_VERSION} \
    && echo "==> gotestsum 安装完成 (${GOTESTSUM_VERSION})" \
    && go install github.com/Kunde21/markdownfmt@${MARKDOWNFMT_VERSION} \
    && echo "==> markdownfmt 安装完成 (${MARKDOWNFMT_VERSION})" \
    && echo "--- Go 工具链安装完成 ---" \
    && ls -l $GOPATH/bin

# 10. 卸载不需要的工具，并清空所有 apt 缓存
RUN apt-get purge -y wget xz-utils \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /root/.npm

CMD ["bash"]