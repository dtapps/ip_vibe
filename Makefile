# 项目目录
PROJECT_ROOT := $(CURDIR)

mod:
	go get -u -v ./...
	go mod tidy
	go mod vendor

generate_ent:
# 	ent generate --feature schema/snapshot,privacy,entql,namedges,bidiedges,sql/schemaconfig,sql/lock,sql/modifier,sql/execquery,sql/upsert ./ent/schema
	ent generate --feature namedges,bidiedges,sql/modifier,sql/execquery,sql/upsert ./ent/schema

ent_new:
	ent new Device --target ./internal/ent/schema

ent_new_log:
	ent new RestyLog --target ./internal/ent_log/schema
	ent new OperationLog --target ./internal/ent_log/schema

swag_v1_install:
	go install github.com/swaggo/swag/cmd/swag@latest

swag_v1:
	make swag_v1_version
	make swag_v1_fmt
	make swag_v1_init

swag_v1_version:
	swag -version

swag_v1_init:
	swag init -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal --generatedTime -output internal/docs_2

swag_v1_fmt:
	swag fmt -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go

swag_v2_install:
	go install github.com/swaggo/swag/v2/cmd/swag@latest

swag_v2:
	make swag_v2_version
	make swag_v2_fmt
	make swag_v2_init

swag_v2_version:
	swag_v2 -version

swag_v2_init:
	swag_v2 init -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal --generatedTime --v3.1 -output internal/docs_3

swag_v2_fmt:
	swag_v2 fmt -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go

swag2op:
	swag2op init -g cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal -o internal/docs_2 -openapiOutputDir internal/docs_3

# 二进制文件名
BINARY_NAME=ip_vibe
# BINARY_NAME_CLI=cli
BINARY_NAME_BACKEND=backend
BINARY_NAME_REPORTER=reporter
BINARY_NAME_FORWARDER=forwarder

# 编译输出目录
BUILD_DIR=bin

# 环境参数
ENV_DEBUG=true
ENV_HTTP_URL=http://localhost:9999
ENV_HTTP_PORT=9999
ENV_BASIC_AUTH_USERNAME=admin
ENV_BASIC_AUTH_PASSWORD=123456

# dev_cli:
# 	export DEBUG=$(ENV_DEBUG) && \
# 	go run cmd/$(BINARY_NAME_CLI)/main.go

dev_backend:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_PORT=$(ENV_HTTP_PORT) && \
	go run cmd/$(BINARY_NAME_BACKEND)/main.go

dev_backend_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_PORT=$(ENV_HTTP_PORT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_BACKEND)/main.go

dev_reporter:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	go run cmd/$(BINARY_NAME_REPORTER)/main.go

dev_reporter_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_REPORTER)/main.go

dev_forwarder:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	go run cmd/$(BINARY_NAME_FORWARDER)/main.go

dev_forwarder_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export HTTP_URL=$(ENV_HTTP_URL) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_FORWARDER)/main.go

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
# 	@rm -f $(BUILD_DIR)/$(BINARY_NAME_CLI)-*
	@rm -f $(BUILD_DIR)/$(BINARY_NAME_BACKEND)-*
	@rm -f $(BUILD_DIR)/$(BINARY_NAME_REPORTER)-*
	@rm -f $(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-*

build-linux-amd64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 linux/amd64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-linux-amd64 .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 linux/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-amd64 .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 linux/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-amd64 .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 linux/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-amd64 .

build-linux-arm64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 linux/arm64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-linux-arm64 .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 linux/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-arm64 .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 linux/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-arm64 .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 linux/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-arm64 .

build-windows-amd64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 windows/amd64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-windows-amd64.exe .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 windows/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-amd64.exe .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 windows/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-amd64.exe .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 windows/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-amd64.exe .

build-windows-arm64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 windows/arm64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-windows-arm64.exe .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 windows/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-arm64.exe .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 windows/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-arm64.exe .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 windows/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-arm64.exe .

build-darwin-amd64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 darwin/amd64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-darwin-amd64 .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 darwin/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-amd64 .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 darwin/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-amd64 .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 darwin/amd64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-amd64 .

build-darwin-arm64:
# 	@echo "编译 $(BINARY_NAME_CLI) 为 darwin/arm64... (version: $(VERSION))"
# 	@cd cmd/$(BINARY_NAME_CLI) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_CLI)-darwin-arm64 .
	@echo "编译 $(BINARY_NAME_BACKEND) 为 darwin/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-arm64 .
	@echo "编译 $(BINARY_NAME_REPORTER) 为 darwin/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-arm64 .
	@echo "编译 $(BINARY_NAME_FORWARDER) 为 darwin/arm64... (version: $(VERSION))"
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-arm64 .
