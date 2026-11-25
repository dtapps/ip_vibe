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

# 安装全部 Go 依赖
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

# 更新全部 Go 依赖
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
BINARY_NAME_CLI = cli
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
## 构建 / 打包 相关
#####################

# goreleaser 版本（所有流水线与本地共用，升级时只改这里）
GORELEASER_VERSION ?= v2.20.2

# goreleaser 可执行文件（不在 PATH 时回退到 $HOME/go/bin）
GORELEASER ?= $(shell command -v goreleaser 2>/dev/null || echo $(HOME)/go/bin/goreleaser)

# 同时编译/链接的目标数。默认按 CPU 数并发，CI 容器内存有限时容易在链接阶段被
# OOM Killer 杀掉（`link: signal: killed`）。本地默认 1，CI 环境（CI=true）默认 2，
# 两者都可显式覆盖：make goreleaser-snapshot GORELEASER_PARALLELISM=4
GORELEASER_PARALLELISM ?= $(if $(filter true,$(CI)),2,1)

# 非 CI（本地）快照只编当前平台：`--single-target` 是 `goreleaser build` 的参数，
# `goreleaser release` 没有（传了会直接报 unknown flag），所以本地走 build，只产出
# 当前平台的三个二进制，不生成压缩包 / Linux 系统包；CI 走 release 出全平台产物。
# 本地要全平台产物（含系统包）：CI=true make goreleaser-snapshot
GORELEASER_SNAPSHOT_CMD ?= $(if $(filter true,$(CI)),release --snapshot,build --snapshot --single-target)

build-frontend:
	@echo "[构建] 打包前端"
	pnpm --dir $(FRONTEND_DIR) build
	@echo "[构建] 打包完成。"

# 安装固定版本的 goreleaser（已存在则跳过）
.PHONY: goreleaser-install
goreleaser-install:
	@if ! $(GORELEASER) --version >/dev/null 2>&1; then \
		echo "安装 goreleaser $(GORELEASER_VERSION) → $(dir $(GORELEASER))"; \
		GOBIN=$(dir $(GORELEASER)) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION); \
	fi

# 校验 .goreleaser.yaml
.PHONY: goreleaser-check
goreleaser-check: goreleaser-install
	$(GORELEASER) check

# 快照构建（无需 tag）：产物在 dist/。本地非 CI 只编当前平台。
.PHONY: goreleaser-snapshot
goreleaser-snapshot: goreleaser-install
	$(GORELEASER) $(GORELEASER_SNAPSHOT_CMD) --clean --parallelism $(GORELEASER_PARALLELISM)

# 发布构建：产物在 dist/。
.PHONY: goreleaser-release
goreleaser-release: goreleaser-install
	$(GORELEASER) release --clean --skip=validate --parallelism $(GORELEASER_PARALLELISM)

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

# 格式化严格 JSON 文件（gojq）
.PHONY: format-json
format-json:
	@bash scripts/format.sh json

# 格式化 JSON5 文件（jsonfmt，保留注释 / 无引号 key / 尾逗号）
.PHONY: format-json5
format-json5:
	@bash scripts/format.sh json5

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

# 格式化/修复项目
.PHONY: format
format: format-node format-go format-yaml format-json format-json5 format-sql format-markdown format-docker format-shell

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

# 运行 Go 静态检查
# -printf=false 关闭 printf 分析，避免 i18n.T() 作日志 format 时触发错误
.PHONY: check-go-vet
check-go-vet:
	@echo "[Check] -> 运行 Go vet 静态检查 (Tags: $(LINT_TAGS))..."
	$(LINT_ENV) go vet -tags=$(LINT_TAGS) -printf=false ./...

# 运行 Go lint 扫描
.PHONY: check-go-lint
check-go-lint:
	@echo "[Check] -> 运行 Go lint 扫描 (平台: $(LINT_PLATFORMS), Tags: $(LINT_TAGS))..."
	@for os_arch in $(LINT_PLATFORMS); do \
		os=$$(echo $$os_arch | cut -d'/' -f1); \
		arch=$$(echo $$os_arch | cut -d'/' -f2); \
		echo "  -> $$os 平台 ($$arch)"; \
		$(LINT_ENV) GOOS=$$os GOARCH=$$arch golangci-lint fmt; \
		$(LINT_ENV) GOOS=$$os GOARCH=$$arch golangci-lint run --build-tags $(LINT_TAGS); \
	done

# 运行 go fix 问题
.PHONY: check-go-fix
check-go-fix:
	@echo "[Check] -> 运行 Go fix 问题 (Tags: $(LINT_TAGS))..."
	$(LINT_ENV) golangci-lint run --fix --build-tags $(LINT_TAGS)

# 运行 Go 单元测试
.PHONY: check-go-test
check-go-test:
	@echo "[Check] -> 运行 Go 单元测试 (Tags: $(LINT_TAGS))..."
	$(LINT_ENV) go test -tags=$(LINT_TAGS) $(GO_COVER_TEST_FLAGS) $(GO_COVER_SCAN)

# 运行 Go 依赖漏洞检查
.PHONY: check-go-vuln
check-go-vuln:
	@echo "[Check] -> 运行 Go 依赖漏洞检查 (Tags: $(LINT_TAGS))..."
	$(LINT_ENV) govulncheck -tags=$(LINT_TAGS) -scan=package -show verbose ./...

# 检查后端项目
.PHONY: check-go
check-go: check-go-vet check-go-lint check-go-fix check-go-test check-go-vuln

# 「过时依赖」是否严格把关：
#   1 = 严格（本地默认）：pnpm outdated 发现过时依赖就让 make 失败
#   0 = 宽松（流水线默认）：只提示，不阻塞流水线
# 流水线（.github/workflows / .cnb/workflows）都会显式传 CI=true，此时默认走宽松；
# 需要在流水线里也严格时显式传参：make check STRICT_CHECKS=1
STRICT_CHECKS ?= $(if $(filter true,$(CI)),0,1)

# 同步前端依赖 (确保环境一致)
.PHONY: check-node-install
check-node-install:
	@echo "[Check] -> 同步前端依赖 (确保环境一致)..."
	pnpm --dir $(FRONTEND_DIR) install --frozen-lockfile --prefer-offline

# 检查前端安全漏洞 (仅高危阻断)
.PHONY: check-node-audit
check-node-audit:
	@echo "[Check] -> 检查前端安全漏洞 (仅高危阻断)..."
	pnpm --dir $(FRONTEND_DIR) audit --production --audit-level high --registry=https://registry.npmjs.org

# 检查前端过时依赖
.PHONY: check-node-outdated
check-node-outdated:
	# pnpm outdated 只要存在过时依赖就返回 1（连 pnpm 自己这个 devDependency 也算），
	# 「有新版可用」不该卡住流水线：流水线里降级为提示，本地仍按 STRICT_CHECKS 失败。
	@echo "[Check] -> 检查前端过时依赖..."
	@if [ "$(STRICT_CHECKS)" = "1" ]; then \
		pnpm --dir $(FRONTEND_DIR) outdated; \
	else \
		pnpm --dir $(FRONTEND_DIR) outdated || echo "  -> (宽松模式) 存在过时依赖，不阻塞流水线"; \
	fi

# 检测前端未使用的依赖 (knip)
.PHONY: check-node-knip
check-node-knip:
	@echo "[Check] -> 检测前端未使用的依赖 (knip)..."
	pnpm --dir $(FRONTEND_DIR) run lint:unused

# 运行前端单元测试
.PHONY: check-node-vitest
check-node-vitest:
	@echo "[Check] -> 运行前端单元测试..."
	pnpm --dir $(FRONTEND_DIR) exec vitest run

# 检查前端项目
.PHONY: check-node
check-node: check-node-install check-node-audit check-node-outdated check-node-lint check-node-knip check-node-vitest

# 检查/测试项目
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