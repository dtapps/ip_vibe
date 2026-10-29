#!/usr/bin/env bash
#
# 统一格式化入口：YAML / Markdown / Dockerfile。
#
# 为什么要独立成脚本，而不是写在 Makefile 里：
#   Make 的每个 recipe 行都是独立的 shell，`|| { ...; exit 0; }` 里的 exit 0
#   只结束当前那一行的 shell，make 仍会继续执行后续 recipe 行。于是「工具未安装」
#   的提示打完之后照样往下调用工具，报 "not found" 并让 make 以非 0 退出。
#   放进单个 shell 脚本后，if/else 的分支天然正确：缺失即跳过，不再误报失败。
#
# 用法：
#   scripts/format.sh              # 全部
#   scripts/format.sh yaml         # 仅 YAML
#   scripts/format.sh shell        # 仅 Shell 脚本
#   scripts/format.sh markdown     # 仅 Markdown
#   scripts/format.sh docker       # 仅 Dockerfile
#
# 环境变量：
#   FORMAT_STRICT=1   工具缺失时以非 0 退出（CI 想强制要求时用）

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# shellcheck source=scripts/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

LOG_PREFIX="[Format] "

STRICT="${FORMAT_STRICT:-0}"

log() {
  printf '%s%s\n' "$LOG_PREFIX" "$*"
}

# skip <工具名> <安装方法> <描述>
# 工具缺失时的处理：默认打印提示后跳过（退出码 0）；FORMAT_STRICT=1 时直接失败。
#
# 注意：本脚本不会自动安装任何工具，第二个参数只是「安装方法」的文字说明，
# 仅供打印，不会被执行。
skip() {
  local tool="$1" install="$2" desc="$3"
  if [ "$STRICT" = "1" ]; then
    printf '%s❌ %s 未安装（FORMAT_STRICT=1，无法跳过）\n' "$LOG_PREFIX" "$tool" >&2
    printf '   本脚本不会自动安装工具，请自行安装后再重试。\n' >&2
    printf '   安装方式: %s\n' "$install" >&2
    return 1
  fi
  printf '%s⚠️  %s 未安装，跳过 %s 格式化。\n' "$LOG_PREFIX" "$tool" "$desc"
  printf '   本脚本不会自动安装工具，如需格式化请自行安装。\n'
  printf '   安装方式: %s\n' "$install"
  return 0
}

fmt_yaml() {
  if command -v yamlfmt >/dev/null 2>&1; then
    log "使用 yamlfmt 格式化 YAML..."
    yamlfmt -conf .yamlfmt \
      ".cnb" \
      ".github" \
      ".gitlab" \
      ".cnb.yml" \
      ".nfpm.yaml" \
      ".golangci.yml" \
      ".gitlab-ci.yml" || {
      printf '❌ yamlfmt 格式化失败\n' >&2
      return 1
    }
    log "YAML 格式化完成（yamlfmt）。"
  else
    skip "yamlfmt" \
      "go install github.com/google/yamlfmt/cmd/yamlfmt@latest" \
      "YAML"
  fi
}

fmt_markdown() {
  if command -v markdownfmt >/dev/null 2>&1; then
    log "使用 markdownfmt 格式化 Markdown..."
    while IFS= read -r -d '' f; do
      markdownfmt -w "$f" || {
        printf '❌ markdownfmt 格式化失败: %s\n' "$f" >&2
        return 1
      }
    done < <(find . -type f -name '*.md' \
      -not -path '*/node_modules/*' \
      -not -path './.git/*' \
      -not -path './.vscode/*' \
      -not -path './.idea/*' \
      -not -path './.codebuddy/*' \
      -not -path './.trae/*' \
      -not -path './.mimocode/*' -print0)
    log "Markdown 格式化完成（markdownfmt）。"
  else
    skip "markdownfmt" \
      "go install github.com/Kunde21/markdownfmt@latest" \
      "Markdown"
  fi
}

fmt_shell() {
  if command -v shfmt >/dev/null 2>&1; then
    log "使用 shfmt 格式化 Shell..."
    shfmt -w -i 2 -ci -bn -s -ln bash scripts/ || {
      printf '❌ shfmt 格式化失败\n' >&2
      return 1
    }
    log "Shell 格式化完成（shfmt）。"
  else
    skip "shfmt" \
      "go install mvdan.cc/sh/v3/cmd/shfmt@latest" \
      "Shell"
  fi
}

fmt_docker() {
  if command -v dockerfmt >/dev/null 2>&1; then
    log "使用 dockerfmt 格式化 Dockerfile..."
    local dir
    for dir in .cnb .github .gitlab; do
      [ -d "$dir" ] || continue
      find "$dir" -type f -name 'Dockerfile*' -exec dockerfmt -w {} + || {
        printf '❌ dockerfmt 格式化失败\n' >&2
        return 1
      }
    done
    log "Dockerfile 格式化完成（dockerfmt）。"
  else
    skip "dockerfmt" \
      "go install github.com/reteps/dockerfmt@latest" \
      "Dockerfile"
  fi
}

main() {
  case "${1:-all}" in
    yaml)
      run_step "YAML 格式化" fmt_yaml
      ;;
    shell)
      run_step "Shell 格式化" fmt_shell
      ;;
    markdown)
      run_step "Markdown 格式化" fmt_markdown
      ;;
    docker)
      run_step "Dockerfile 格式化" fmt_docker
      ;;
    all)
      run_step "YAML 格式化" fmt_yaml
      run_step "Shell 格式化" fmt_shell
      run_step "Markdown 格式化" fmt_markdown
      run_step "Dockerfile 格式化" fmt_docker
      ;;
    -h | --help)
      printf '用法: %s [yaml|shell|markdown|docker|all]\n' "$0"
      ;;
    *)
      printf '用法: %s [yaml|shell|markdown|docker|all]\n' "$0" >&2
      exit 2
      ;;
  esac
}

main "$@"
