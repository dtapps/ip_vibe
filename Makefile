# 设置默认目标
.DEFAULT_GOAL := setup

# 一键设置项目（默认执行）
setup: install-tools verify-tools mod pnpm_update
	@echo "🎉 Project setup completed!"

# 安装
# 	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
# 	go install github.com/cloudwego/thriftgo@latest
# 	go install github.com/zxmfke/swagger2openapi3/cmd/swag2op@latest
# 	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest 
install-tools:
	go install entgo.io/ent/cmd/ent@latest && \
	go install github.com/swaggo/swag/cmd/swag@latest && mv $(shell go env GOPATH)/bin/swag $(shell go env GOPATH)/bin/swag_v1 && \
	go install github.com/swaggo/swag/v2/cmd/swag@latest && mv $(shell go env GOPATH)/bin/swag $(shell go env GOPATH)/bin/swag_v2 && \
	go install github.com/bufbuild/buf/cmd/buf@latest && \
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
	go install github.com/envoyproxy/protoc-gen-validate@latest && \
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest && \
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	@echo "🎉 All tools installed successfully!"

# 验证安装
# 	@test -f $(shell go env GOPATH)/bin/protoc-gen-connect-go && echo "✅ protoc-gen-connect-go found" || echo "❌ protoc-gen-connect-go not found"
# 	@test -f $(shell go env GOPATH)/bin/thriftgo && echo "✅ thriftgo found" || echo "❌ thriftgo not found"
# 	@test -f $(shell go env GOPATH)/bin/protoc-gen-go-triple && echo "✅ protoc-gen-go-triple found" || echo "❌ protoc-gen-go-triple not found"
# 	@test -f $(shell go env GOPATH)/bin/swag2op && echo "✅ swag2op found" || echo "❌ swag2op not found"
# 	@test -f $(shell go env GOPATH)/bin/protoc-gen-openapiv2 && echo "✅ protoc-gen-openapiv2 found" || echo "❌ protoc-gen-openapiv2 not found"
verify-tools:
	@test -f $(shell go env GOPATH)/bin/ent && echo "✅ ent found" || echo "❌ ent not found"
	@test -f $(shell go env GOPATH)/bin/swag_v1 && echo "✅ swag_v1 found" || echo "❌ swag_v1 not found"
	@test -f $(shell go env GOPATH)/bin/swag_v2 && echo "✅ swag_v2 found" || echo "❌ swag_v2 not found"
	@test -f $(shell go env GOPATH)/bin/buf && echo "✅ buf found" || echo "❌ buf not found"
	@test -f $(shell go env GOPATH)/bin/protoc-gen-go && echo "✅ protoc-gen-go found" || echo "❌ protoc-gen-go not found"
	@test -f $(shell go env GOPATH)/bin/protoc-gen-go-grpc && echo "✅ protoc-gen-go-grpc found" || echo "❌ protoc-gen-go-grpc not found"
	@test -f $(shell go env GOPATH)/bin/protoc-gen-grpc-gateway && echo "✅ protoc-gen-grpc-gateway found" || echo "❌ protoc-gen-grpc-gateway not found"
	@test -f $(shell go env GOPATH)/bin/protoc-gen-validate && echo "✅ protoc-gen-validate found" || echo "❌ protoc-gen-validate not found"
	@test -f $(shell go env GOPATH)/bin/grpcurl && echo "✅ grpcurl found" || echo "❌ grpcurl not found"
	@test -f $(shell go env GOPATH)/bin/protoc-gen-doc && echo "✅ protoc-gen-doc found" || echo "❌ protoc-gen-doc not found"
	@echo "🎉 Verification complete!"

# 项目目录
PROJECT_ROOT := $(CURDIR)

version:
	go version

fix:
	go fix ./...

download:
	go get -u ./...

tidy:
	go mod tidy

vendor:
	go mod vendor

deps: 
	go version
	go get -u ./...
	go mod tidy
	go mod vendor
	go fix ./...

# generate: swag_v1_all swag_v2_all swag2op_all grpc_generate_all thrift_generate
generate: swag_v1_all swag_v2_all grpc_generate_all

generate_ent:
# 	ent generate --feature schema/snapshot,privacy,entql,namedges,bidiedges,sql/schemaconfig,sql/lock,sql/modifier,sql/execquery,sql/upsert ./ent/schema
	ent generate --feature namedges,bidiedges,sql/modifier,sql/execquery,sql/upsert ./ent/schema

ent_new:
	ent new CronJob --target ./internal/ent/schema

ent_new_log:
	ent new CronJobLog --target ./internal/ent_log/schema

ent_new_stat:
	ent new CloudflareZoneAnalysisData --target ./internal/ent_stat/schema

swag_v1_all: swag_v1_version swag_v1_fmt swag_v1_init

swag_v1_version:
	swag_v1 -version

swag_v1_init:
	swag_v1 init -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal --generatedTime -output internal/docs_2

swag_v1_fmt:
	swag_v1 fmt -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go

swag_v2_all: swag_v2_version swag_v2_fmt swag_v2_init

swag_v2_version:
	swag_v2 -version

swag_v2_init:
	swag_v2 init -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal --generatedTime --v3.1 -output internal/docs_3

swag_v2_fmt:
	swag_v2 fmt -generalInfo cmd/$(BINARY_NAME_BACKEND)/main.go

swag2op_all: swag2op_version swag2op_init

swag2op_version:
	swag2op -version

swag2op_init:
	swag2op init -g cmd/$(BINARY_NAME_BACKEND)/main.go --parseDependency --parseInternal -o internal/docs_2 -openapiOutputDir internal/docs_2_to_3

pnpm_outdated:
	cd cmd/$(BINARY_NAME_FRONTEND) && \
	pnpm outdated

pnpm_update:
	cd cmd/$(BINARY_NAME_FRONTEND) && \
	pnpm update --latest

# 	make grpc_generate_connect
# 	make grpc_generate_openapi
grpc_generate_all:
	make grpc_generate_rm
	make grpc_generate_pb
	make grpc_generate_gateway

# 生成 gRPC 核心代码（包括 protobuf、gRPC、Validate）
grpc_generate_rm:
	@echo "rm internal/grpc_generated catalogue"
	rm -rf ./internal/grpc_generated/*
	@echo "rm internal/grpc_generated catalogue done"

# 生成 gRPC 核心代码（包括 protobuf、gRPC、Validate）
grpc_generate_pb:
	@echo "Generating gRPC core code..."
	protoc --proto_path=./internal/grpc_proto \
	       --proto_path=$$(go env GOPATH)/src/googleapis \
		   --proto_path=$$(go env GOPATH)/src/protoc-gen-validate \
	       --go_out=./internal/grpc_generated \
	       --go_opt=paths=source_relative \
	       --go-grpc_out=./internal/grpc_generated \
	       --go-grpc_opt=paths=source_relative \
	       --validate_out="lang=go:./internal/grpc_generated" \
	       --validate_opt=paths=source_relative \
	       ./internal/grpc_proto/*.proto
	@echo "✅ gRPC core code (incl. Validate) generated successfully!"

# 生成 Connect HTTP 处理函数
grpc_generate_connect:
	@echo "Generating Connect HTTP handlers..."
	protoc --proto_path=./internal/grpc_proto \
	       --proto_path=$$(go env GOPATH)/src/googleapis \
		   --proto_path=$$(go env GOPATH)/src/protoc-gen-validate \
	       --connect-go_out=./internal \
	       --connect-go_opt=paths=source_relative \
	       ./internal/grpc_proto/*.proto
	@echo "✅ Connect handlers generated"

# 生成 gRPC-Gateway 代码（RESTful HTTP/JSON）
grpc_generate_gateway:
	@echo "Generating gRPC-Gateway code..."
	protoc --proto_path=./internal/grpc_proto \
	       --proto_path=$$(go env GOPATH)/src/googleapis \
		   --proto_path=$$(go env GOPATH)/src/protoc-gen-validate \
	       --grpc-gateway_out=./internal/grpc_generated \
	       --grpc-gateway_opt=paths=source_relative \
	       --grpc-gateway_opt=logtostderr=true \
	       ./internal/grpc_proto/*.proto
	@echo "✅ gRPC-Gateway code generated successfully!"

# 生成 OpenAPI v2 (Swagger) 规范
grpc_generate_openapi:
	@echo "Generating OpenAPI v2 (Swagger)..."
	protoc --proto_path=./internal/grpc_proto \
	       --proto_path=$$(go env GOPATH)/src/googleapis \
		   --proto_path=$$(go env GOPATH)/src/protoc-gen-validate \
	       --openapiv2_out=./internal/grpc_openapi \
	       --openapiv2_opt=logtostderr=true \
	       ./internal/grpc_proto/*.proto
	@echo "✅ OpenAPI v2 generated at ./internal/grpc_openapi"

# https://fuchsia.googlesource.com/third_party/googleapis/
grpc_generate_old:
	@echo "Generating gRPC code..."
	protoc -I . \
	       -I $$(go env GOPATH)/src/googleapis \
	       --go_out=paths=source_relative:./internal/grpc_generated \
	       --go-grpc_out=paths=source_relative:./internal/grpc_generated \
	       ./internal/grpc_proto/*.proto
	echo "✅ gRPC code generated successfully!"

# https://thrift.apache.org/lib/go.html
# thrift_generate:
# 	thrift \
# 		   --gen go:package_prefix=vibe_ip/internal/thrift_generated/ \
# 	       -out ./internal/thrift_generated \
# 	       ./internal/thrift_idl/*.thrift 
# 	rm -rf ./internal/thrift_generated/*/*-remote

# thriftgo_generate:
# 	thriftgo \
# 		   --gen go:package_prefix=vibe_ip/internal/thriftgo_generated/,gen_setter,gen_db_tag,omitempty_for_optional,use_type_alias,validate_set,scan_value_for_enum,reserve_comments \
# 	       --out ./internal/thriftgo_generated \
# 	       ./internal/thrift_idl/*.thrift

# 二进制文件名
BINARY_NAME=ip_vibe
# BINARY_NAME_CLI=cli
BINARY_NAME_FRONTEND=frontend
BINARY_NAME_BACKEND=backend
BINARY_NAME_REPORTER=reporter
BINARY_NAME_FORWARDER=forwarder

# 编译输出目录
BUILD_DIR=bin

# 环境参数
ENV_DEBUG=true

ENV_SERVER_ADDR=0.0.0.0:9999

ENV_BACKEND_ENDPOINT=localhost:9999

ENV_BASIC_AUTH_USERNAME=admin
ENV_BASIC_AUTH_PASSWORD=123456

# ENV_BACKEND_ENDPOINT=192.168.101.4:9998
# ENV_BACKEND_ENDPOINT=100.113.15.21:9998
# ENV_BACKEND_ENDPOINT=asw-ddns.dtapp.net

# ENV_BASIC_AUTH_USERNAME=asw
# ENV_BASIC_AUTH_PASSWORD=235235235

# dev_cli:
# 	export DEBUG=$(ENV_DEBUG) && \
# 	go run cmd/$(BINARY_NAME_CLI)/main.go

# pnpm add -D tailwindcss @tailwindcss/cli daisyui@latest
# pnpm add -D select2 jquery
# 启动前端服务（css 热更新）
dev_frontend:
	cd cmd/$(BINARY_NAME_FRONTEND) && \
	pnpm dev

# 启动后端服务
dev_backend:
	export DEBUG=$(ENV_DEBUG) && \
	export SERVER_ADDR=$(ENV_SERVER_ADDR) && \
	go run cmd/$(BINARY_NAME_BACKEND)/main.go

# 启动后端服务（启用 Basic Auth）
dev_backend_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export SERVER_ADDR=$(ENV_SERVER_ADDR) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_BACKEND)/main.go

# 启动IP上报服务
dev_reporter:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	go run cmd/$(BINARY_NAME_REPORTER)/main.go

# 启动IP上报服务（启用 Basic Auth）
dev_reporter_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_REPORTER)/main.go

# 启动端口转发服务
dev_forwarder:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	go run cmd/$(BINARY_NAME_FORWARDER)/main.go

# 启动端口转发服务（启用 Basic Auth）
dev_forwarder_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run cmd/$(BINARY_NAME_FORWARDER)/main.go

# 版本信息（支持外部传入）
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDTIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# 路径
PKG_PATH = vibe_ip/internal/global

# ldflags 模板（注意：单引号在 Make 中需转义或使用双引号包裹）
# -ldflags="-s -w" - 移除符号表和调试信息
# -gcflags="-m" - 启用内联优化分析
LDFLAGS = -s -w \
	-X '$(PKG_PATH).Version=$(VERSION)' \
	-X '$(PKG_PATH).GitCommit=$(COMMIT)' \
	-X '$(PKG_PATH).BuildTime=$(BUILDTIME)'

# 禁用CGO，实现完全静态链接
CGO_ENABLED=0

# 构建所有平台
build-all: clean \
	build-linux-amd64 \
	build-linux-arm64 \
	build-linux-arm-v7 \
	build-linux-mipsle \
	build-linux-riscv64 \
	build-windows-amd64 \
	build-windows-arm64 \
	build-darwin-amd64 \
	build-darwin-arm64

clean:
	@echo "[Clean] 清理构建目录..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

build-frontend:
	@echo "[构建] 前端静态文件..."
	@cd cmd/$(BINARY_NAME_FRONTEND) && pnpm build

build-linux-amd64:
	@echo "[构建] Linux AMD64 (通用服务器/PC x86_64)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-amd64 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-amd64 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-amd64 .

build-linux-arm64:
	@echo "[构建] Linux ARM64 (树莓派4/5, AWS Graviton, 苹果M系Linux虚拟机)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-arm64 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-arm64 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-arm64 .

build-linux-arm-v7:
	@echo "[构建] Linux ARM v7 (32位, 旧款树莓派, 大部分NAS, 嵌入式盒子)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-arm-v7 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-arm-v7 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-arm-v7 .

build-linux-mipsle:
	@echo "[构建] Linux MIPS LE (小端序, 常见于 OpenWrt 路由器/光猫)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-mipsle .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-mipsle .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-mipsle .

build-linux-riscv64:
	@echo "[构建] Linux RISC-V 64位 (新兴开源指令集架构)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=linux GOARCH=riscv64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-riscv64 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=linux GOARCH=riscv64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-riscv64 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=linux GOARCH=riscv64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-riscv64 .

build-windows-amd64:
	@echo "[构建] Windows AMD64 (主流 Windows PC/服务器)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-amd64.exe .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-amd64.exe .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-amd64.exe .

build-windows-arm64:
	@echo "[构建] Windows ARM64 (Surface Pro X, 骁龙笔记本等)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=windows GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-arm64.exe .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=windows GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-arm64.exe .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=windows GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-arm64.exe .

build-darwin-amd64:
	@echo "[构建] macOS Intel (旧款 Macbook/iMac)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-amd64 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-amd64 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-amd64 .

build-darwin-arm64:
	@echo "[构建] macOS Apple Silicon (M1/M2/M3 芯片)"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-arm64 .
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-arm64 .
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-arm64 .

#####################
## 检查相关
#####################

check-backend:
	@echo "[Check]检查后端代码语法转换与格式化..."
	@echo "  -> 校验 YAML 格式..."
	yamllint -c ./.yamllint . 
	@echo "  -> 自动语法转换..."
	find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./internal/ent/*" ! -path "./internal/ent_log/*" ! -path "./internal/ent_stat/*" ! -path "./internal/grpc_generated/*" -exec dirname {} \; | sort -u | xargs -I {} go fix {}
	@echo "  -> 格式化..."
	find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./internal/ent/*" ! -path "./internal/ent_log/*" ! -path "./internal/ent_stat/*" ! -path "./internal/grpc_generated/*" -exec gofmt -w {} +
	@echo "[Check] 运行 golangci-lint 深度扫描 (Tags: $(LINT_TAGS))..."
	@echo "  -> Linux 平台 (amd64)"
	$(LINT_CMD) GOOS=linux GOARCH=amd64 golangci-lint fmt
	$(LINT_CMD) GOOS=linux GOARCH=amd64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=linux GOARCH=amd64 golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "  -> Linux 平台 (arm64)"
	$(LINT_CMD) GOOS=linux GOARCH=arm64 golangci-lint fmt
	$(LINT_CMD) GOOS=linux GOARCH=arm64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=linux GOARCH=arm64 golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "  -> macOS 平台 (amd64)"
	$(LINT_CMD) GOOS=darwin GOARCH=amd64 golangci-lint fmt
	$(LINT_CMD) GOOS=darwin GOARCH=amd64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=darwin GOARCH=amd64 golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "  -> macOS 平台 (arm64)"
	$(LINT_CMD) GOOS=darwin GOARCH=arm64 golangci-lint fmt
	$(LINT_CMD) GOOS=darwin GOARCH=arm64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=darwin GOARCH=arm64 golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "  -> Windows 平台 (amd64)"
	$(LINT_CMD) GOOS=windows GOARCH=amd64 golangci-lint fmt
	$(LINT_CMD) GOOS=windows GOARCH=amd64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=windows GOARCH=amd64 golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "  -> Windows 平台 (arm64)"
	$(LINT_CMD) GOOS=windows GOARCH=arm64 golangci-lint fmt
	$(LINT_CMD) GOOS=windows GOARCH=arm64 golangci-lint run --build-tags $(LINT_TAGS)
	$(LINT_CMD) GOOS=windows GOARCH=arm64 golangci-lint run --fix --build-tags $(LINT_TAGS)
check-frontend:
	@echo "[Check] 检查前端依赖..."
	@echo "  -> 检查安全漏洞 (audit)"
	-cd cmd/$(BINARY_NAME_FRONTEND) && pnpm audit || echo "  ⚠️  audit 检查失败，可能镜像源不支持"
	@echo "  -> 检查过时依赖 (outdated)"
	-cd cmd/$(BINARY_NAME_FRONTEND) && pnpm outdated || true

check-all: check-backend check-frontend
