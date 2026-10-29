# 使用 golang 镜像作为基础
FROM golang:1.27.3-trixie

# ==========================================
# 定义版本变量
# https://github.com/upx/upx/releases
# https://github.com/golangci/golangci-lint/releases
# https://github.com/nodejs/node/releases
# https://github.com/pnpm/pnpm/releases
# https://github.com/goreleaser/nfpm/releases
# https://github.com/boumenot/gocover-cobertura/releases
# https://github.com/Kunde21/markdownfmt/releases
# https://github.com/google/yamlfmt/releases
# https://github.com/mvdan/sh/releases
# https://github.com/reteps/dockerfmt/releases
# ==========================================
ARG UPX_VERSION=5.2.1
ARG GOLANGCI_LINT_VERSION=v2.15.0
ARG NODE_VERSION=26.12.1
ARG PNPM_VERSION=12.15.0
ARG NFPM_VERSION=2.48.0
ARG GOCOVER_COBERTURA_VERSION=v1.5.0
ARG MARKDOWNFMT_VERSION=v0.1.6
ARG YAMLFMT_VERSION=v0.21.0
ARG SHFMT_VERSION=v3.14.1
ARG DOCKERFMT_VERSION=v0.5.4

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
    unzip \
    yamllint \
    wixl \
    wixl-data \
    && rm -rf /var/lib/apt/lists/* \
    && echo "--- 基础构建工具安装完成 ---" \
    && wget --version | head -1 \
    && xz --version | head -1 \
    && curl --version | head -1 \
    && gpg --version | head -1 \
    && zip -v | sed -n '2p' \
    && git --version \
    && make --version | head -1 \
    && unzip -v | head -1 \
    && yamllint --version \
    && wixl --version \
    && dpkg-query -W -f='ca-certificates ${Version}\n' ca-certificates \
    && dpkg-query -W -f='wixl-data ${Version}\n' wixl-data

# 3. 安装 UPX 工具
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then UPX_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then UPX_ARCH="arm64"; \
    else echo "不支持的架构: $ARCH" && exit 1; fi && \
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
    echo "--- UPX 安装的 ---" && \
    upx --version

# 4. 安装 nfpm 工具
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then NFPM_ARCH="x86_64"; \
    elif [ "$ARCH" = "aarch64" ]; then NFPM_ARCH="arm64"; \
    else echo "不支持的架构: $ARCH" && exit 1; fi && \
    \
    NFPM_URL="https://github.com/goreleaser/nfpm/releases/download/v${NFPM_VERSION}/nfpm_${NFPM_VERSION}_Linux_${NFPM_ARCH}.tar.gz" && \
    \
    cd /tmp && \
    wget -q -O nfpm.tar.gz "${NFPM_URL}" && \
    tar -xzf nfpm.tar.gz nfpm && \
    mv nfpm /usr/local/bin/nfpm && \
    chmod +x /usr/local/bin/nfpm && \
    rm -rf nfpm.tar.gz && \
    echo "--- nfpm 安装的 ---" && \
    nfpm --version

# 5. 安装 Node.js 工具
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then NODE_ARCH="x64"; \
    elif [ "$ARCH" = "aarch64" ]; then NODE_ARCH="arm64"; \
    else echo "不支持的架构: $ARCH" && exit 1; fi && \
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
    rm -rf node.tar.xz /root/.npm /root/.cache/node && \
    echo "--- Node.js 安装的 ---" && \
    node -v && npm -v && pnpm -v

# 6. 预安装 Go 工具链
# 同一层内 go install + 清缓存，避免 go-build 和 mod 缓存留在独立层里
RUN echo "==> 预安装 Go 工具链..." \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && echo "==> golangci-lint 安装完成 (${GOLANGCI_LINT_VERSION})" \
    && go install github.com/boumenot/gocover-cobertura@${GOCOVER_COBERTURA_VERSION} \
    && echo "==> gocover-cobertura 安装完成 (${GOCOVER_COBERTURA_VERSION})" \
    && go install github.com/Kunde21/markdownfmt@${MARKDOWNFMT_VERSION} \
    && echo "==> markdownfmt 安装完成 (${MARKDOWNFMT_VERSION})" \
    && go install github.com/google/yamlfmt/cmd/yamlfmt@${YAMLFMT_VERSION} \
    && echo "==> yamlfmt 安装完成 (${YAMLFMT_VERSION})" \
    && go install mvdan.cc/sh/v3/cmd/shfmt@${SHFMT_VERSION} \
    && echo "==> shfmt 安装完成 (${SHFMT_VERSION})" \
    && go install github.com/reteps/dockerfmt@${DOCKERFMT_VERSION} \
    && echo "==> dockerfmt 安装完成 (${DOCKERFMT_VERSION})" \
    && echo "--- Go 工具链安装完成 ---" \
    && ls -l $GOPATH/bin

# 7. 卸载不需要的工具，并清空所有 apt 缓存
RUN apt-get purge -y wget xz-utils \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /root/.npm

CMD ["bash"]