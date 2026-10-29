#!/usr/bin/env bash
#
# 项目检查入口：后端 golangci-lint + 前端 pnpm 依赖/审计/Lint。
#
# 为什么要独立成脚本，而不是写在 Makefile 里：
#   1. Make 的每个 recipe 行都是独立的 shell，跨行的错误处理（set -e、条件分支）
#      写起来极易出错——原来的 check-backend 用 `for` 循环遍历平台，但该行 shell 没有
#      set -e，只有最后一个平台的退出码决定成败，中间平台 lint 失败会被静默吞掉；
#   2. 放进单个 shell 脚本后 set -euo pipefail 生效，任何一个平台失败立即报错，
#      且各子命令的失败都能带上清晰的上下文。
#
# 用法：
#   scripts/check.sh            # 全部（后端 + 前端）
#   scripts/check.sh backend    # 仅后端
#   scripts/check.sh frontend   # 仅前端
#
# 环境变量（均可覆盖，默认值与 Makefile 保持一致）：
#   LINT_TAGS        Go build tags，默认 goolm,stdjson
#   LINT_PLATFORMS   lint 扫描平台，默认当前平台（仅在执行后端检查时才求值）；
#                    CI 用 LINT_PLATFORMS=... 覆盖为全平台
#   CGO_ENABLED      默认 0（完全静态链接）
#   FRONTEND_DIR     前端目录，默认 cmd/frontend
#
# 工具缺失时的行为：直接报错并以非 0 退出，**不会自动安装任何工具**。
# 唯一的安装行为是前端的 `pnpm install --frozen-lockfile`——它只同步 node_modules
# （--frozen-lockfile 保证 lockfile 不被改写），用于让后续 audit / lint 有依赖可用。

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# shellcheck source=scripts/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

LOG_PREFIX="[Check] "

LINT_TAGS="${LINT_TAGS:-goolm,stdjson}"
CGO_ENABLED="${CGO_ENABLED:-0}"
FRONTEND_DIR="${FRONTEND_DIR:-cmd/frontend}"

log() {
  printf '%s%s\n' "$LOG_PREFIX" "$*"
}

# missing <工具名> <安装方法>：check 是质量门禁，工具缺失必须显式失败，
# 不能像 format 那样静默跳过——跳过会让 CI 假装通过。
#
# 注意：这里只做「检测 + 报错 + 退出」，本脚本不会自动安装任何工具。
# 第二个参数仅作为「安装方法」的文字说明打印出来，不会被执行。
missing() {
  printf '%s❌ %s 未安装，无法执行检查。\n' "$LOG_PREFIX" "$1" >&2
  printf '   本脚本不会自动安装工具，请自行安装后再重试。\n' >&2
  printf '   安装方式: %s\n' "$2" >&2
  return 1
}

check_backend() {
  # 显式 `|| return 1`：本函数会在 run_step 的条件上下文中被调用，
  # 此时 bash 会抑制函数体内的 set -e，错误必须靠 return 显式往外传。
  if ! command -v go >/dev/null 2>&1; then
    missing "go" "https://go.dev/dl/" || return 1
  fi
  if ! command -v golangci-lint >/dev/null 2>&1; then
    missing "golangci-lint" \
      "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest" || return 1
  fi

  # 默认平台延迟到此处计算：只有在真正要做后端检查时才需要 go，
  # 避免脚本启动就因 go 缺失而失败（前端检查不依赖 go）。
  local platforms="${LINT_PLATFORMS:-$(go env GOOS)/$(go env GOARCH)}"

  log "运行 golangci-lint 扫描 (平台: ${platforms}, Tags: ${LINT_TAGS})..."
  local os_arch os arch
  for os_arch in ${platforms}; do
    os="${os_arch%%/*}"
    arch="${os_arch##*/}"
    echo "  -> ${os} 平台 (${arch})"
    env CGO_ENABLED="${CGO_ENABLED}" GOOS="${os}" GOARCH="${arch}" \
      golangci-lint fmt
    env CGO_ENABLED="${CGO_ENABLED}" GOOS="${os}" GOARCH="${arch}" \
      golangci-lint run --build-tags "${LINT_TAGS}"
  done

  log "自动修复代码风格问题..."
  env CGO_ENABLED="${CGO_ENABLED}" golangci-lint run --fix --build-tags "${LINT_TAGS}"

  log "完成。"
}

check_frontend() {
  if ! command -v pnpm >/dev/null 2>&1; then
    missing "pnpm" "npm install -g pnpm" || return 1
  fi

  log "检查前端依赖与质量..."
  echo "  -> 同步依赖 (确保环境一致)..."
  pnpm --dir "${FRONTEND_DIR}" install --frozen-lockfile --prefer-offline

  echo "  -> 检查安全漏洞 (仅高危阻断)..."
  pnpm --dir "${FRONTEND_DIR}" audit --production --audit-level high \
    --registry=https://registry.npmjs.org

  echo "  -> 检查过时依赖..."
  # pnpm outdated 在发现可升级依赖时退出码为 1，属于「有更新可用」的正常信号，
  # 不是检查失败——只提示，不阻断（原来的 Makefile 版本在这里会误报失败）。
  if ! pnpm --dir "${FRONTEND_DIR}" outdated; then
    echo "  -> ⚠️  存在可升级的依赖（仅提示，不阻断）"
  fi

  echo "  -> 检查代码风格 (Lint)..."
  pnpm --dir "${FRONTEND_DIR}" run lint

  log "前端检查完成。"
}

main() {
  case "${1:-all}" in
    backend)
      run_step "后端检查 (golangci-lint)" check_backend
      ;;
    frontend)
      run_step "前端检查 (pnpm)" check_frontend
      ;;
    all)
      run_step "后端检查 (golangci-lint)" check_backend
      run_step "前端检查 (pnpm)" check_frontend
      ;;
    -h | --help)
      printf '用法: %s [backend|frontend|all]\n' "$0"
      ;;
    *)
      printf '用法: %s [backend|frontend|all]\n' "$0" >&2
      exit 2
      ;;
  esac
}

main "$@"
