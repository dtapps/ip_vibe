# 设置默认目标
.DEFAULT_GOAL := setup

# 环境变量优先级高于 .env：include 的赋值会覆盖环境值，先暂存一份
ifeq ($(origin GIT_REMOTE),environment)
    GIT_REMOTE_ENV := $(GIT_REMOTE)
endif

# 加载 .env 文件（如果存在）
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

ifneq ($(GIT_REMOTE_ENV),)
    GIT_REMOTE := $(GIT_REMOTE_ENV)
endif

# 同步/推送目标远程，逗号分隔：GIT_REMOTE=origin,github,gitlab
# 可通过 .env / 环境变量 / 命令行覆盖，未配置时默认 origin。
GIT_REMOTE ?= origin
export GIT_REMOTE

# 一键设置项目（默认执行）
setup: install-tools verify-tools deps
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

# 安装全部依赖
.PHONY: deps
deps: deps-go deps-node
	@echo "[Deps] 完成。"

# 安装全部Go依赖
.PHONY: deps-go
deps-go:
	@echo "[Deps] 安装 Go 模块依赖..."
	go mod tidy
	go mod download

# 安装全部前端依赖
# --frozen-lockfile (冻结锁文件)
# --prefer-offline (优先使用离线缓存)
.PHONY: deps-node
deps-node:
	@echo "[Deps] 安装前端依赖..."
	pnpm --dir $(FRONTEND_DIR) install --frozen-lockfile --prefer-offline

# 更新全部依赖
# --latest (无视版本限制，强制升最新)
.PHONY: update-deps
update-deps: update-deps-go update-deps-node
	@echo "[Deps] 完成。"

# 更新全部Go依赖
.PHONY: update-deps-go
update-deps-go:
	@echo "[Deps] 更新 Go 模块依赖..."
	go get -u ./...
	go mod tidy

# 更新全部前端依赖
# --latest (无视版本范围，强制升最新，并改写 package.json 的版本范围)
.PHONY: update-deps-node
update-deps-node:
	@echo "[Deps] 更新前端依赖..."
	pnpm --dir $(FRONTEND_DIR) update --latest
	pnpm --dir $(FRONTEND_DIR) prune

# 更新 pnpm 自身的版本锁定（package.json 的 packageManager 字段）
.PHONY: update-deps-pnpm
update-deps-pnpm:
	@echo "[Deps] 更新 pnpm 版本锁定..."
	@latest="$$(npm view pnpm version)"; \
	echo "[Deps] packageManager → pnpm@$${latest}"; \
	pnpm --dir $(FRONTEND_DIR) pkg set packageManager=pnpm@$${latest}

# 删除未使用的依赖：knip --fix 把它们从 package.json 摘掉，
.PHONY: deps-clean
deps-clean:
	@echo "[Deps] 删除未使用的依赖..."
	pnpm --dir $(FRONTEND_DIR) run deps:clean
	pnpm --dir $(FRONTEND_DIR) install --lockfile-only
	pnpm --dir $(FRONTEND_DIR) prune

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
	swag_v1 init -generalInfo $(BACKEND_DIR)/main.go --parseDependency --parseInternal --generatedTime -output internal/docs_2

swag_v1_fmt:
	swag_v1 fmt -generalInfo $(BACKEND_DIR)/main.go

swag_v2_all: swag_v2_version swag_v2_fmt swag_v2_init

swag_v2_version:
	swag_v2 -version

swag_v2_init:
	swag_v2 init -generalInfo $(BACKEND_DIR)/main.go --parseDependency --parseInternal --generatedTime --v3.1 -output internal/docs_3

swag_v2_fmt:
	swag_v2 fmt -generalInfo $(BACKEND_DIR)/main.go

swag2op_all: swag2op_version swag2op_init

swag2op_version:
	swag2op -version

swag2op_init:
	swag2op init -g $(BACKEND_DIR)/main.go --parseDependency --parseInternal -o internal/docs_2 -openapiOutputDir internal/docs_2_to_3

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
BINARY_NAME = ip_vibe
BINARY_NAME_CLI=cli
BINARY_NAME_FRONTEND = frontend
BINARY_NAME_BACKEND = backend
BINARY_NAME_REPORTER = reporter
BINARY_NAME_FORWARDER = forwarder

# 前端项目目录
FRONTEND_DIR = cmd/frontend

# 后端项目目录
BACKEND_DIR = cmd/backend

# 上报项目目录
REPORTER_DIR = cmd/reporter

# 转发项目目录
FORWARDER_DIR = cmd/forwarder

# 编译输出目录
BUILD_DIR = bin

#####################
## 开发 相关
#####################

# 环境参数
ENV_DEBUG = true

ENV_SERVER_ADDR=0.0.0.0:9999

ENV_BACKEND_ENDPOINT=localhost:9999

ENV_BASIC_AUTH_USERNAME = admin
ENV_BASIC_AUTH_PASSWORD = 123456

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
	pnpm --dir $(FRONTEND_DIR) dev

# 启动后端服务
dev_backend:
	export DEBUG=$(ENV_DEBUG) && \
	export SERVER_ADDR=$(ENV_SERVER_ADDR) && \
	go run $(BACKEND_DIR)/main.go

# 启动后端服务（启用 Basic Auth）
dev_backend_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export SERVER_ADDR=$(ENV_SERVER_ADDR) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run $(BACKEND_DIR)/main.go

# 启动IP上报服务
dev_reporter:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	go run $(REPORTER_DIR)/main.go

# 启动IP上报服务（启用 Basic Auth）
dev_reporter_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run $(REPORTER_DIR)/main.go

# 启动端口转发服务
dev_forwarder:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	go run $(FORWARDER_DIR)/main.go

# 启动端口转发服务（启用 Basic Auth）
dev_forwarder_basic_auth:
	export DEBUG=$(ENV_DEBUG) && \
	export BACKEND_ENDPOINT=$(ENV_BACKEND_ENDPOINT) && \
	export BASIC_AUTH_USERNAME=$(ENV_BASIC_AUTH_USERNAME) && \
	export BASIC_AUTH_PASSWORD=$(ENV_BASIC_AUTH_PASSWORD) && \
	go run $(FORWARDER_DIR)/main.go

#####################
## 构建 相关
#####################

# 功能开关（设置为 1 开启，0 关闭）
ENABLE_UPX ?= 1
ENABLE_ARCHIVE ?= 1

# 检测是否存在 upx 压缩工具
UPX = $(shell command -v upx 2> /dev/null)
# 检测是否存在 tar 和 zip
TAR = $(shell command -v tar 2> /dev/null)
ZIP = $(shell command -v zip 2> /dev/null)

# 定义压缩逻辑：参数 $(1) 为文件路径
define compress_file
	@if [ "$(ENABLE_UPX)" = "1" ] && [ "$(UPX)" ]; then \
		FILE_SIZE=$$(stat -c%s "$(1)" 2>/dev/null || stat -f%z "$(1)"); \
		ORIG_SIZE_H=$$(du -h "$(1)" | cut -f1); \
		\
		if [ $$FILE_SIZE -lt 10485760 ]; then \
			LEVEL="--best"; \
			DESC="小文件-极限压缩"; \
		elif [ $$FILE_SIZE -lt 52428800 ]; then \
			LEVEL="-6"; \
			DESC="中等文件-平衡模式"; \
		else \
			LEVEL="-1"; \
			DESC="大文件-极速模式"; \
		fi; \
 \
		echo "  📦 正在以 [$$DESC ($$LEVEL)] 压缩..."; \
		ERR_MSG=$$($(UPX) $$LEVEL "$(1)" 2>&1); \
 \
		if [ $$? -eq 0 ]; then \
			NEW_SIZE_H=$$(du -h "$(1)" | cut -f1); \
			echo "  ⚡ UPX 压缩完成: $(1) [ $${ORIG_SIZE_H} -> $${NEW_SIZE_H} ]"; \
		else \
			echo "  ⚠️  UPX 压缩跳过: $(1) ($$ERR_MSG)"; \
		fi \
	elif [ "$(ENABLE_UPX)" = "1" ] && [ ! "$(UPX)" ]; then \
		echo "  ⏭️  UPX 未安装，跳过压缩"; \
	elif [ "$(ENABLE_UPX)" = "0" ]; then \
		echo "  ⏭️  UPX 已禁用"; \
	fi
endef

# 定义打包压缩逻辑：参数 $(1) 为源文件路径
# 源文件命名规范: {binary-name}-{os}-{arch}[.exe]，例如: backend-linux-amd64, reporter-windows-amd64.exe
# Linux/macOS 使用 tar.gz, Windows 使用 zip
define archive_binary
	@if [ "$(ENABLE_ARCHIVE)" = "1" ]; then \
		SOURCE_FILE=$(1); \
		BIN_NAME=$$(basename "$${SOURCE_FILE}"); \
		ARCHIVE_NAME="$(BUILD_DIR)/$${BIN_NAME}"; \
		\
		if echo "$${BIN_NAME}" | grep -q "\.exe$$"; then \
			if [ "$(ZIP)" ]; then \
				ARCHIVE_NAME="$${ARCHIVE_NAME}.zip"; \
				echo "  📦 正在打包 Windows 压缩包: $${ARCHIVE_NAME}..."; \
				cd $(BUILD_DIR) && zip -q "$$(basename $${ARCHIVE_NAME})" "$$(basename $${SOURCE_FILE})"; \
				cd ..; \
				ARCHIVE_SIZE=$$(du -h "$${ARCHIVE_NAME}" | cut -f1); \
				BIN_SIZE=$$(du -h "$${SOURCE_FILE}" | cut -f1); \
				echo "  ✓ 打包完成: $${ARCHIVE_NAME} [ 二进制: $${BIN_SIZE} -> 压缩包: $${ARCHIVE_SIZE} ]"; \
			else \
				echo "  ⚠️  zip 未安装，无法打包 Windows 压缩包"; \
			fi \
		else \
			if [ "$(TAR)" ]; then \
				ARCHIVE_NAME="$${ARCHIVE_NAME}.tar.gz"; \
				echo "  📦 正在打包 Unix 压缩包: $${ARCHIVE_NAME}..."; \
				tar -czf "$${ARCHIVE_NAME}" -C $(BUILD_DIR) "$$(basename $${SOURCE_FILE})"; \
				ARCHIVE_SIZE=$$(du -h "$${ARCHIVE_NAME}" | cut -f1); \
				BIN_SIZE=$$(du -h "$${SOURCE_FILE}" | cut -f1); \
				echo "  ✓ 打包完成: $${ARCHIVE_NAME} [ 二进制: $${BIN_SIZE} -> 压缩包: $${ARCHIVE_SIZE} ]"; \
			else \
				echo "  ⚠️  tar 未安装，无法打包 Unix 压缩包"; \
			fi \
		fi; \
	elif [ "$(ENABLE_ARCHIVE)" = "0" ]; then \
		echo "  ⏭️  压缩包已禁用"; \
	fi
endef

# 定义清理二进制文件逻辑：如果存在对应压缩包则删除二进制
# 参数 $(1) 为二进制文件路径
# 源文件命名规范: {binary-name}-{os}-{arch}[.exe]，例如: backend-linux-amd64, reporter-windows-amd64.exe
define clean_binary
	@SOURCE_FILE=$(1); \
	BIN_NAME=$$(basename "$${SOURCE_FILE}"); \
	if echo "$${BIN_NAME}" | grep -q "\.exe$$"; then \
		ARCHIVE_NAME="$(BUILD_DIR)/$${BIN_NAME}.zip"; \
	else \
		ARCHIVE_NAME="$(BUILD_DIR)/$${BIN_NAME}.tar.gz"; \
	fi; \
	if [ -f "$${ARCHIVE_NAME}" ] && [ -f "$${SOURCE_FILE}" ]; then \
		rm -f "$${SOURCE_FILE}"; \
		echo "  🗑️  已清理二进制: $${SOURCE_FILE} (存在 $${ARCHIVE_NAME})"; \
	else \
		echo "  ⏭️  未找到压缩包或二进制，跳过: $${SOURCE_FILE}"; \
	fi
endef

# 版本信息（支持外部传入）
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILDTIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# 路径
PKG_PATH = vibe_ip/internal/global

# ldflags 模板
# -s -w: 移除符号表和调试信息，减小体积
# -buildid=: 移除构建指纹
LDFLAGS = -s -w \
	-X '$(PKG_PATH).Version=$(VERSION)' \
	-X '$(PKG_PATH).GitCommit=$(COMMIT)' \
	-X '$(PKG_PATH).BuildTime=$(BUILDTIME)'

# 统一定义构建命令
# CGO_ENABLED=0: 禁用 CGO，实现完全静态链接
# CGO_ENABLED=1: 启用 CGO，实现动态链接
# -trimpath: 移除本地路径，保护隐私
# -buildvcs=false: 不把 git 状态盖章进二进制
#   版本/提交号本来就走上面的 LDFLAGS 注入，VCS 盖章信息用不上；而它会调用
#   `git status --porcelain`，git 一旦不可用就报 `error obtaining VCS status: exit status 128`
#   把构建打断（go test 不做盖章，所以覆盖率流水线一直是绿的）。
BUILD_CMD = CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="$(LDFLAGS)"

# 构建所有平台
build-all: clean \
	build-linux-amd64 \
	build-linux-arm64 \
	build-windows-amd64 \
	build-windows-arm64 \
	build-darwin-amd64 \
	build-darwin-arm64

clean:
	@echo "[Clean] 清理构建目录..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

build-frontend:
	@echo ""
	@echo "┌────────────────────────────────────────────────────────────"
	@echo "│ [构建] 前端静态文件..."
	@echo "└────────────────────────────────────────────────────────────"
	pnpm --dir $(FRONTEND_DIR) build

# 定义构建单个二进制文件的通用逻辑
# 参数: $(1)=GOOS, $(2)=GOARCH, $(3)=描述信息, $(4)=扩展名(.exe 或空)
define build_binary
	@echo ""
	@echo "┌────────────────────────────────────────────────────────────"
	@echo "│ [构建] $(3)"
	@echo "└────────────────────────────────────────────────────────────"
	@cd $(BACKEND_DIR) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-$(1)-$(2)$(4) .
	$(call archive_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-$(1)-$(2)$(4))
	@cd $(REPORTER_DIR) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-$(1)-$(2)$(4) .
	$(call archive_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-$(1)-$(2)$(4))
	@cd $(FORWARDER_DIR) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-$(1)-$(2)$(4) .
	$(call archive_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-$(1)-$(2)$(4))
endef

build-linux-amd64:
	$(call build_binary,linux,amd64,Linux x86_64,)

build-linux-arm64:
	$(call build_binary,linux,arm64,Linux ARM64,)

build-windows-amd64:
	$(call build_binary,windows,amd64,Windows x86_64,.exe)

build-windows-arm64:
	$(call build_binary,windows,arm64,Windows ARM64,.exe)

build-darwin-amd64:
	$(call build_binary,darwin,amd64,macOS x86_64,)

build-darwin-arm64:
	$(call build_binary,darwin,arm64,macOS Apple Silicon,)

# 清理所有已压缩的二进制文件（上传前执行）
.PHONY: clean-binaries
clean-binaries:
	@echo "[清理] 删除已压缩的二进制文件..."
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-linux-arm64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-linux-arm64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-linux-arm64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-amd64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-amd64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-amd64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-windows-arm64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-windows-arm64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-windows-arm64.exe)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-amd64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-darwin-arm64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-darwin-arm64)
	$(call clean_binary,./$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-darwin-arm64)
	@echo "✓ 清理完成"

#####################
## 格式化 相关
#####################

# 格式化后端文件
.PHONY: format-go
format-go:
	@echo "[Format] 格式化后端文件 (GO)..."
	gofmt -w -s $(shell find . -type f -name '*.go' -not -path './internal/ent/*' -not -path './internal/ent_log/*' -not -path './internal/ent_stat/*')
	go fmt ./...
	go fix ./...

# 格式前端文件
.PHONY: format-node
format-node:
	@echo "[Format] 格式化前端文件 (PNPM)..."
	pnpm --dir $(FRONTEND_DIR) format

# 格式化 YAML 文件
.PHONY: format-yaml
format-yaml:
	@bash scripts/format.sh yaml

# 格式化 JSON 文件
.PHONY: format-json
format-json:
	@bash scripts/format.sh json

# 格式化 SQL 文件
.PHONY: format-sql
format-sql:
	@bash scripts/format.sh sql

# 格式化 Markdown 文件
.PHONY: format-markdown
format-markdown:
	@bash scripts/format.sh markdown

# 格式化 Dockerfile 文件
.PHONY: format-docker
format-docker:
	@bash scripts/format.sh docker

# 格式化 Shell 脚本
.PHONY: format-shell
format-shell:
	@bash scripts/format.sh shell

# 格式化
.PHONY: format
format: format-node format-go format-yaml format-json format-sql format-markdown format-docker format-shell

#####################
## 检查 / 测试 相关
#####################

# 统一定义构建命令
# CGO_ENABLED=0: 禁用 CGO，实现完全静态链接
# CGO_ENABLED=1: 启用 CGO，实现动态链接
LINT_ENV = CGO_ENABLED=0

# 定义 Lint 过程中需要激活的 Go Build Tags
LINT_TAGS = goolm,stdjson

# Lint 扫描的目标平台，默认当前平台；CI 通过 LINT_PLATFORMS=... 覆盖为全平台
LINT_PLATFORMS ?= $(shell go env GOOS)/$(shell go env GOARCH)

# 检查后端代码
.PHONY: check-go
check-go:
	@echo "[Check] 运行 go vet 静态检查..."
	go vet ./...

	@echo "[Check] 运行 golangci-lint 扫描 (平台: $(LINT_PLATFORMS), Tags: $(LINT_TAGS))..."
	@for os_arch in $(LINT_PLATFORMS); do \
		os=$$(echo $$os_arch | cut -d'/' -f1); \
		arch=$$(echo $$os_arch | cut -d'/' -f2); \
		echo "  -> $$os 平台 ($$arch)"; \
		$(LINT_ENV) GOOS=$$os GOARCH=$$arch golangci-lint fmt; \
		$(LINT_ENV) GOOS=$$os GOARCH=$$arch golangci-lint run --build-tags $(LINT_TAGS); \
	done

	@echo "[Check] 自动修复代码风格问题..."
	$(LINT_ENV) golangci-lint run --fix --build-tags $(LINT_TAGS)

	@echo "[Check] 运行 Go 单元测试..."
	go test $(GO_COVER_TEST_FLAGS) $(GO_COVER_SCAN)

	@echo "[Check] 运行 Go 赖漏洞检查..."
	govulncheck -show verbose ./...

	@echo "[Check] 完成。"

# 检查前端代码
.PHONY: check-node
check-node:
	@echo "[Check] 检查前端依赖与质量..."
	@echo "  -> 同步依赖 (确保环境一致)..."
	pnpm --dir $(FRONTEND_DIR) install --frozen-lockfile --prefer-offline

	@echo "  -> 检查安全漏洞 (仅高危阻断)..."
	pnpm --dir $(FRONTEND_DIR) audit --production --audit-level high --registry=https://registry.npmjs.org

	@echo "  -> 检查过时依赖..."
	pnpm --dir $(FRONTEND_DIR) outdated

	@echo "  -> 检查代码风格 (Lint)..."
	pnpm --dir $(FRONTEND_DIR) run lint

	@echo "  -> 检测未使用的依赖 (knip)..."
	pnpm --dir $(FRONTEND_DIR) run lint:unused

	@echo "  -> 运行前端单元测试..."
	pnpm --dir $(FRONTEND_DIR) exec vitest run

	@echo "[Check] 前端检查完成。"

# 检查代码
.PHONY: check
check: check-node check-go

#####################
## 更新 / 拉取 / 推送 相关
#####################

# 按 GIT_REMOTE 顺序先全部拉取合并，再统一推送（保留本地未提交改动）
.PHONY: sync
sync:
	@IFS=,; for r in $$GIT_REMOTE; do \
		git fetch "$$r" && git merge --ff-only "$$r/master" || exit 1; \
	done
	@IFS=,; for r in $$GIT_REMOTE; do git push "$$r" HEAD || exit 1; done
	@echo "同步完成！"

# 推送到所有远程仓库
.PHONY: push
push:
	@IFS=,; for r in $$GIT_REMOTE; do git push "$$r" HEAD || exit 1; done
	@echo "推送完成！"

# 强制推送到所有远程仓库（忽略冲突）
.PHONY: push-force
push-force:
	@IFS=,; for r in $$GIT_REMOTE; do git push --force "$$r" HEAD || exit 1; done
	@echo "强制推送完成！"

#####################
## 测试覆盖率 相关
#####################

# ── 清单配置：新增目录 / 生成代码只改这里，下面的命令不散落路径 ──────────────
# Go 扫描范围：业务代码全在 internal/ 下，cmd/ 只是可执行入口，不计入分母。
GO_COVER_SCAN = ./cmd/... ./internal/... ./pkg/...
# Go 排除目录（包路径前缀，空格分隔）：ent 生成代码占总语句数 94.6%
# （实测 225382 条中占 213232 条），计入会把指标压到接近 0。
# 以前只写 /internal/ent 靠前缀顺带排掉 internal/ent_log，现在显式列全。
GO_COVER_EXCLUDE = /internal/ent /internal/ent_log /internal/ent_stat /cmd/ent
# 关闭 vet 检查，加快测试与覆盖率统计速度
GO_COVER_TEST_FLAGS = -vet=off

# 前端目录与产物（相对仓库根）
NODE_DIR = $(FRONTEND_DIR)
NODE_COVER_DIR = $(NODE_DIR)/coverage
NODE_COVER_LCOV = $(NODE_COVER_DIR)/lcov.info
NODE_JUNIT_XML = $(NODE_COVER_DIR)/junit.xml
# 前端扫描 / 排除清单（glob 相对 vitest root=cmd/，由 cmd/frontend/vitest.config.mts 读取）：
#   NODE_TEST_SCAN      测试用例
#   NODE_COVER_SCAN     计入覆盖率的源码
#   NODE_COVER_EXCLUDE  不计入的第三方目录
# 多个 glob 用空格或逗号分隔；留空则回退到配置文件里的默认值。
NODE_TEST_SCAN = frontend/test/**/*.test.js
NODE_COVER_SCAN = backend/static/js/**/*.js
NODE_COVER_EXCLUDE = backend/static/libs/**
NODE_VITEST_ENV = VITEST_TEST_INCLUDE="$(NODE_TEST_SCAN)" VITEST_COVER_INCLUDE="$(NODE_COVER_SCAN)" VITEST_COVER_EXCLUDE="$(NODE_COVER_EXCLUDE)"

# -coverpkg 只收「扫描范围内且不在排除列表」的包。用 $$(...) 让 go list 在配方执行时才跑，
# 避免任何 make 目标都触发一次包扫描；grep 参数由排除列表生成，加目录不用改命令。
GO_COVER_EXCLUDE_ARGS = $(foreach d,$(GO_COVER_EXCLUDE),-e '$(d)')
GO_COVER_PKGS = $$(go list $(GO_COVER_SCAN) | grep -v $(GO_COVER_EXCLUDE_ARGS) | paste -sd, -)
GO_COVER_ARGS = $(GO_COVER_TEST_FLAGS) -coverpkg=$(GO_COVER_PKGS) -coverprofile=coverage.out $(GO_COVER_SCAN)

# Go 覆盖率（golang 格式）→ coverage.out；CNB 的 testing:coverage 与 Codecov 都读这份
.PHONY: cover-go
cover-go:
	@echo "[Coverage] Go 测试覆盖率 → coverage.out"
	go test $(GO_COVER_ARGS)
	go tool cover -func=coverage.out

# Go 覆盖率 + JUnit（Codecov Test Analytics 用），依赖 CI 镜像预装的 gotestsum
.PHONY: junit-go
junit-go:
	@echo "[Coverage] Go 测试 → coverage.out + go-test-report.xml"
	gotestsum --junitfile go-test-report.xml -- $(GO_COVER_ARGS)
	go tool cover -func=coverage.out

# Coveralls 的 golang parser 会把包前缀压成 basename，改用 Cobertura 保留完整 filename 路径。
# gocover-cobertura 由统一 CI 镜像预装；-strict 让「包/模块信息缺失被静默跳过」直接失败。
.PHONY: cover-go-cobertura
cover-go-cobertura: cover-go
	@echo "[Coverage] coverage.out → coverage.xml（Cobertura）"
	gocover-cobertura -by-files -strict < coverage.out > coverage.xml

# 前端覆盖率（Vitest v8）→ cmd/frontend/coverage/lcov.info
.PHONY: cover-node
cover-node:
	@echo "[Coverage] 前端覆盖率 → $(NODE_COVER_LCOV)"
	cd $(NODE_DIR) && $(NODE_VITEST_ENV) pnpm exec vitest run --coverage

# 前端 JUnit（Codecov Test Analytics 用）；outputFile 用绝对路径避开 vitest root 的相对路径歧义
.PHONY: junit-node
junit-node:
	@echo "[Coverage] 前端 JUnit → $(NODE_JUNIT_XML)"
	@mkdir -p $(NODE_COVER_DIR)
	cd $(NODE_DIR) && $(NODE_VITEST_ENV) pnpm exec vitest run --reporter=junit --outputFile=$(CURDIR)/$(NODE_JUNIT_XML)