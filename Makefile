# 项目目录
PROJECT_ROOT := $(CURDIR)

mod:
	go get -u -v ./...
	go mod tidy

generate_ent:
# 	ent generate --feature schema/snapshot,privacy,entql,namedges,bidiedges,sql/schemaconfig,sql/lock,sql/modifier,sql/execquery,sql/upsert ./ent/schema
	ent generate --feature namedges,bidiedges,sql/modifier,sql/execquery,sql/upsert ./ent/schema

# 二进制文件名
BINARY_NAME=ip_vibe

# 编译输出目录
BUILD_DIR=bin

# 环境参数
ENV_DEBUG=true
ENV_HTTP_URL=http://localhost:9999
ENV_HTTP_PORT=9999
ENV_BASIC_AUTH_USERNAME=admin
ENV_BASIC_AUTH_PASSWORD=123456

dev_cli:
	export DEBUG=$(ENV_DEBUG) && \
	go run cmd/cli/main.go

dev_backend:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_PORT=$(ENV_HTTP_PORT) && \
	go run cmd/backend/main.go

dev_backend_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_PORT=$(ENV_HTTP_PORT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/backend/main.go

dev_reporter:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	go run cmd/reporter/main.go

dev_reporter_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/reporter/main.go

# 已应用的优化：
# CGO_ENABLED=0 - 禁用CGO，实现完全静态链接
# -ldflags="-s -w" - 移除符号表和调试信息
# -gcflags="-m" - 启用内联优化分析

# 版本信息（支持外部传入）
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDTIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# ldflags 模板（注意：单引号在 Make 中需转义或使用双引号包裹）
LDFLAGS = -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.buildTime=$(BUILDTIME)'

# 构建所有平台
build-all: clean build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64 build-darwin-amd64 build-darwin-arm64

clean:
	@echo "清理..."
	@rm -f $(BUILD_DIR)/cli-*
	@rm -f $(BUILD_DIR)/backend-*
	@rm -f $(BUILD_DIR)/reporter-*

build-linux-amd64:
	@echo "编译 cli 为 linux/amd64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-linux-amd64 .
	@echo "编译 backend 为 linux/amd64..."
	@cd cmd/backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-linux-amd64 .
	@echo "编译 reporter 为 linux/amd64..."
	@cd cmd/reporter && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-linux-amd64 .

build-linux-arm64:
	@echo "编译 cli 为 linux/arm64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-linux-arm64 .
	@echo "编译 backend 为 linux/arm64..."
	@cd cmd/backend && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-linux-arm64 .
	@echo "编译 reporter 为 linux/arm64..."
	@cd cmd/reporter && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-linux-arm64 .

build-windows-amd64:
	@echo "编译 cli 为 windows/amd64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-windows-amd64.exe .
	@echo "编译 backend 为 windows/amd64..."
	@cd cmd/backend && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-windows-amd64.exe .
	@echo "编译 reporter 为 windows/amd64..."
	@cd cmd/reporter && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-windows-amd64.exe .

build-windows-arm64:
	@echo "编译 cli 为 windows/arm64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-windows-arm64.exe .
	@echo "编译 backend 为 windows/arm64..."
	@cd cmd/backend && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-windows-arm64.exe .
	@echo "编译 reporter 为 windows/arm64..."
	@cd cmd/reporter && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-windows-arm64.exe .

build-darwin-amd64:
	@echo "编译 cli 为 darwin/amd64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-darwin-amd64 .
	@echo "编译 backend 为 darwin/amd64..."
	@cd cmd/backend && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-darwin-amd64 .
	@echo "编译 reporter 为 darwin/amd64..."
	@cd cmd/reporter && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-darwin-amd64 .

build-darwin-arm64:
	@echo "编译 cli 为 darwin/arm64... (version: $(VERSION))"
	@cd cmd/cli && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/cli-darwin-arm64 .
	@echo "编译 backend 为 darwin/arm64..."
	@cd cmd/backend && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/backend-darwin-arm64 .
	@echo "编译 reporter 为 darwin/arm64..."
	@cd cmd/reporter && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/reporter-darwin-arm64 .
