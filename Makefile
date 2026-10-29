# 设置默认目标
.DEFAULT_GOAL := setup

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

# 安装依赖
deps:
	@echo "[Deps] 安装 Go 模块依赖..."
	go mod tidy
	go mod download
	@echo "[Deps] 安装前端依赖..."
	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) install --frozen-lockfile --prefer-offline
	@echo "[Deps] 完成。"

# 更新依赖
update-deps:
	@echo "[Deps] 更新 Go 模块依赖..."
	go get -u ./...
	go mod tidy
	@echo "[Deps] 更新前端依赖..."
	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) self-update
# 	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) update
	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) update --latest
	@echo "[Deps] 完成。"

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
# BINARY_NAME_CLI=cli
BINARY_NAME_FRONTEND = frontend
BINARY_NAME_BACKEND = backend
BINARY_NAME_REPORTER = reporter
BINARY_NAME_FORWARDER = forwarder

# 编译输出目录
BUILD_DIR = bin

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

#####################
## 检查相关
#####################
#
# 统一定义构建命令
# CGO_ENABLED=0: 禁用 CGO，实现完全静态链接
# CGO_ENABLED=1: 启用 CGO，实现动态链接
#
# Lint 用的 CGO_ENABLED 由 scripts/check.sh 通过环境变量接收（默认 0）。

# 定义 Lint 过程中需要激活的 Go Build Tags
LINT_TAGS = goolm,stdjson

# Lint 扫描的目标平台，默认当前平台；CI 通过 LINT_PLATFORMS=... 覆盖为全平台
LINT_PLATFORMS ?= $(shell go env GOOS)/$(shell go env GOARCH)

# 实现放在 scripts/check.sh：Make 的每个 recipe 行是独立 shell，
# 原来的多平台 for 循环所在行没有 set -e，只有最后一个平台的退出码决定成败，
# 中间平台 lint 失败会被静默吞掉；放进脚本后 set -euo pipefail 生效，失败即报错。
.PHONY: check-backend
check-backend:
	@LINT_TAGS="$(LINT_TAGS)" LINT_PLATFORMS="$(LINT_PLATFORMS)" \
		CGO_ENABLED=0 bash scripts/check.sh backend

.PHONY: check-frontend
check-frontend:
	@FRONTEND_DIR="cmd/$(BINARY_NAME_FRONTEND)" bash scripts/check.sh frontend

# 检查项目
.PHONY: check
check:
	@LINT_TAGS="$(LINT_TAGS)" LINT_PLATFORMS="$(LINT_PLATFORMS)" \
		CGO_ENABLED=0 FRONTEND_DIR="cmd/$(BINARY_NAME_FRONTEND)" bash scripts/check.sh all

#####################
## 测试覆盖率
#####################

# 覆盖率口径的唯一事实来源：CNB（.cnb/workflows/*-coverage.yml）与 GitHub Actions
# （.github/workflows/coverage.yml、coveralls.yml）都调用本节目标，命令不再在多处漂移。
#
# Go：
#   - coverpkg 范围为 ./cmd/... ./internal/... ./pkg/...，排除 4 处 ent 生成目录
#     （internal/ent、internal/ent_log、internal/ent_stat、cmd/ent；
#      ent 占全树语句约 73%，计入会把业务覆盖率严重稀释）；
#   - 产出 coverage.out（仓库根）。
# 前端：
#   - node:test + c8（CJS 插桩靠 require hook；vitest 对 CJS 经转译会丢失插桩），
#     c8 命令唯一定义在 cmd/frontend/package.json 的 test:coverage 脚本里，
#     本目标只做转发——该脚本必须在仓库根执行，lcov 的 SF 路径才是仓库根相对，
#     Codecov / Coveralls 靠它映射到源码文件；
#   - 产出 cmd/frontend/coverage/lcov.info。
GO_COVER_PKGS = ./cmd/... ./internal/... ./pkg/...

.PHONY: cover-go
cover-go:
	@echo "[Coverage] 生成 Go 测试覆盖率..."
	go test -vet=off -coverpkg=$$(go list $(GO_COVER_PKGS) | grep -vE '/(ent|ent_log|ent_stat)(/|$$)' | paste -sd, -) -coverprofile=coverage.out $(GO_COVER_PKGS)
	go tool cover -func=coverage.out
	@echo "[Coverage] 完成：coverage.out"

# Coveralls 专用：把 coverage.out 转成 lcov。
#   - coverage.out 里的路径是 Go 包导入路径（vibe_ip/internal/core/device.go）；
#   - Coveralls 的 golang parser（coverage-reporter 的 golang_parser.cr）会用正则
#     硬吃 3 段包前缀，只把 basename（device.go）当文件名 → 243 个文件全坍缩到同一级；
#   - 而 lcov parser 的 SF: 是原样保留完整路径的，所以走 lcov 格式可避开这个正则。
# gcov2lcov 转 lcov 时把包导入路径解析回仓库根相对路径（SF:internal/core/device.go），
# 与前端 c8 产出的 lcov 形态一致，Coveralls 侧就能正确按目录层级展示。
.PHONY: cover-go-lcov
cover-go-lcov: cover-go
	@echo "[Coverage] 转换 Go 覆盖率为 lcov (供 Coveralls)..."
	gcov2lcov -infile coverage.out -outfile coverage-lcov.info 2>/dev/null
	@echo "[Coverage] 完成：coverage-lcov.info"

.PHONY: cover-node
cover-node:
	@echo "[Coverage] 生成前端测试覆盖率..."
	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) test:coverage
	@echo "[Coverage] 完成：cmd/$(BINARY_NAME_FRONTEND)/coverage/lcov.info"

.PHONY: cover
cover: cover-go cover-node

# JUnit XML：供 Codecov 的 Tests 面板消费（codecov/test-results-action）。
#   - 前端：Node 内置 --test-reporter=junit（Node >= 22.8；镜像里是 Node 26），
#     单进程跑聚合入口 test/index.test.js，不用 node --test（那会把多文件塞进
#     worker，CJS 源不被 c8 插桩 → 覆盖率归零）；
#   - Go：gotestsum（镜像第 7 步已预装，v1.13.0 —— v1.12.0 在 Go 1.27 编译失败），
#     与 cover-go 同一套 coverpkg 口径，顺带产出 coverage.out。
.PHONY: junit-node
junit-node:
	@echo "[JUnit] 生成前端测试报告 (JUnit XML)..."
	pnpm --dir cmd/$(BINARY_NAME_FRONTEND) test:junit
	@echo "[JUnit] 完成：cmd/$(BINARY_NAME_FRONTEND)/coverage/junit.xml"

.PHONY: junit-go
junit-go:
	@echo "[JUnit] 生成 Go 测试报告 (JUnit XML)..."
	gotestsum --junitfile go-test-report.xml --format testname --raw-command -- go test -json -vet=off -coverpkg=$$(go list $(GO_COVER_PKGS) | grep -vE '/(ent|ent_log|ent_stat)(/|$$)' | paste -sd, -) -coverprofile=coverage.out $(GO_COVER_PKGS)
	go tool cover -func=coverage.out
	@echo "[JUnit] 完成：go-test-report.xml（同时产出 coverage.out）"

.PHONY: junit
junit: junit-go junit-node

#####################
## 构建相关
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
	@cd cmd/$(BINARY_NAME_FRONTEND) && pnpm build

# 定义构建单个二进制文件的通用逻辑
# 参数: $(1)=GOOS, $(2)=GOARCH, $(3)=描述信息, $(4)=扩展名(.exe 或空)
define build_binary
	@echo ""
	@echo "┌────────────────────────────────────────────────────────────"
	@echo "│ [构建] $(3)"
	@echo "└────────────────────────────────────────────────────────────"
	@cd cmd/$(BINARY_NAME_BACKEND) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-$(1)-$(2)$(4) .
	$(call archive_binary,./$(BUILD_DIR)/$(BINARY_NAME_BACKEND)-$(1)-$(2)$(4))
	@cd cmd/$(BINARY_NAME_REPORTER) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-$(1)-$(2)$(4) .
	$(call archive_binary,./$(BUILD_DIR)/$(BINARY_NAME_REPORTER)-$(1)-$(2)$(4))
	@cd cmd/$(BINARY_NAME_FORWARDER) && GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o ../../$(BUILD_DIR)/$(BINARY_NAME_FORWARDER)-$(1)-$(2)$(4) .
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

# ==================== 工具 ====================

# 格式化（Go + 配置 + 脚本 + 文档 + Docker）
.PHONY: format
format: format-go format-yaml format-shell format-markdown format-docker

# 格式化 Go 代码
.PHONY: format-go
format-go:
	@echo "[Format] 格式化后端文件 (GO)..."
	gofmt -w -s $(shell find . -type f -name '*.go' -not -path './internal/ent/*' -not -path './internal/ent_log/*' -not -path './internal/ent_stat/*')
	go fmt ./...
	go fix ./...

# 格式化 YAML 文件
.PHONY: format-yaml
format-yaml:
	@bash scripts/format.sh yaml

# 格式化 Shell 脚本
# 实现放在 scripts/format.sh：工具缺失即跳过，不误报失败（同 format-yaml 的说明）。
.PHONY: format-shell
format-shell:
	@bash scripts/format.sh shell

# 格式化 Markdown 文件
.PHONY: format-markdown
format-markdown:
	@bash scripts/format.sh markdown

# 格式化 Dockerfile 文件
.PHONY: format-docker
format-docker:
	@bash scripts/format.sh docker

# ==================== 更新 / 拉取 ====================

sync: sync-github sync-cnb push-force ## 拉取所有远程最新（先 GitHub，后 CNB 镜像）然后强制推送

sync-github: ## 拉取 GitHub 最新并以 fast-forward 合并（保留本地未提交改动）
	git fetch github
	git merge --ff-only github/master
	@echo "已同步 github/master 最新代码，本地未提交改动已保留"

sync-cnb: ## 拉取 CNB 最新并以 fast-forward 合并（保留本地未提交改动）
	git fetch origin
	git merge --ff-only origin/master
	@echo "已同步 origin/master 最新代码，本地未提交改动已保留"

# ==================== 推送 ====================

# 推送到所有远程仓库
.PHONY: push
push:
	git push origin HEAD
	git push github HEAD
	@echo "推送完成！"

# 强制推送到所有远程仓库（忽略冲突）
.PHONY: push-force
push-force:
	git push --force origin HEAD
	git push --force github HEAD
	@echo "强制推送完成！"
