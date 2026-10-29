#!/usr/bin/env bash
#
# 版本号归一化 + 预发布判定。
#
# 事实来源：GitHub Actions 的 release.yml / build-image.yml 都 source 这一份实现，
# 避免两处各写一套 grep 正则、判断口径慢慢跑偏。规则与 CNB 的 .cnb/workflows/*.yml 一致：
#
#   v1.2.3 / 1.2.3        -> TAG=v1.2.3        PRERELEASE=false   正式版
#   v1.2.3-rc.1           -> TAG=v1.2.3-rc.1   PRERELEASE=true    预发布
#   v1.2.3-beta.1         -> TAG=v1.2.3-beta.1 PRERELEASE=true    预发布
#   v1.2.3-alpha.1        -> TAG=v1.2.3-alpha.1 PRERELEASE=true   预发布
#
# 只用版本号后缀判定，不接受「是不是预发布」这种布尔输入：布尔输入会和版本号打架。
#
# 用法（被 source）：
#   . scripts/ci/resolve-prerelease.sh
#   resolve_version "${VERSION}"
#   printf '%s %s\n' "${TAG}" "${PRERELEASE}"
#
# 直接执行可以跑内置用例自检（改了正则以后先跑它）：
#   bash scripts/ci/resolve-prerelease.sh --self-test

# resolve_version <版本号>
#   归一化版本号（补 v 前缀）并判定是否预发布；
#   结果写入全局变量 TAG / PRERELEASE / VERSION_TYPE / PRE_NOTE 并导出给后续步骤。
#   版本号为空时返回 1，由调用方决定是 `set -e` 直接失败还是自行处理。
resolve_version() {
  local input="${1:-}"

  # 去掉所有空白字符：workflow_dispatch 输入框粘贴时经常带上首尾空格/换行
  input="$(printf '%s' "${input}" | tr -d '[:space:]')"

  if [ -z "${input}" ]; then
    printf '错误: 版本号不能为空。示例: v1.2.3 / v1.2.3-rc.1 / v1.2.3-beta.1 / v1.2.3-alpha.1\n' >&2
    return 1
  fi

  # 统一补 v 前缀：输入 1.2.3 与 v1.2.3 等价；大写 V 前缀也算已经有了，
  # 否则会拼出 vV1.2.3 这种既不像 tag 又不像版本号的东西
  case "${input}" in
    v* | V*) TAG="${input}" ;;
    *) TAG="v${input}" ;;
  esac

  # 后缀判定：-rc / -beta / -alpha（大小写不敏感，rc1 / rc.1 / 裸 rc 都算预发布）
  # 匹配要求后缀出现在版本号末尾，避免 v1.2.3-source、v1.2.3-hotfix 这类被误判
  if printf '%s' "${TAG}" | grep -qiE -- '-(rc|beta|alpha)([.-]?[0-9A-Za-z.-]*)$'; then
    PRERELEASE="true"
    VERSION_TYPE="预发布"
  else
    PRERELEASE="false"
    VERSION_TYPE="正式版"
  fi

  PRE_NOTE="版本号: ${TAG}（${VERSION_TYPE}，prerelease=${PRERELEASE}）"
  export TAG PRERELEASE VERSION_TYPE PRE_NOTE

  return 0
}

# 直接执行（而不是被 source）时跑内置用例，方便本地确认规则没被改坏。
if [ "${BASH_SOURCE[0]:-}" = "${0}" ]; then
  if [ "${1:-}" != "--self-test" ]; then
    printf '用法: %s --self-test\n通常不需要单独执行，CI 里是 source 使用：%s\n' "${0}" '. scripts/ci/resolve-prerelease.sh' >&2
    exit 2
  fi

  failures=0

  check_version() {
    local input="$1" want_tag="$2" want_pre="$3"

    if ! resolve_version "${input}" 2>/dev/null; then
      printf '  ✗ %-20s 解析失败（期望 %s）\n' "${input:-<空>}" "${want_tag}"
      failures=$((failures + 1))
      return 0
    fi

    if [ "${TAG}" = "${want_tag}" ] && [ "${PRERELEASE}" = "${want_pre}" ]; then
      printf '  ✓ %-20s -> %-20s prerelease=%s\n' "${input}" "${TAG}" "${PRERELEASE}"
    else
      printf '  ✗ %-20s -> %-20s prerelease=%s（期望 %s / %s）\n' \
        "${input}" "${TAG}" "${PRERELEASE}" "${want_tag}" "${want_pre}"
      failures=$((failures + 1))
    fi
  }

  check_version 'v1.2.3' 'v1.2.3' 'false'
  check_version '1.2.3' 'v1.2.3' 'false'
  check_version 'v1.2.3-rc.1' 'v1.2.3-rc.1' 'true'
  check_version 'v1.2.3-beta.1' 'v1.2.3-beta.1' 'true'
  check_version 'v1.2.3-alpha.1' 'v1.2.3-alpha.1' 'true'
  check_version 'v1.2.3-rc' 'v1.2.3-rc' 'true'
  check_version 'V1.2.3-RC.1' 'V1.2.3-RC.1' 'true'
  check_version 'v1.2.3-alpha.1.2' 'v1.2.3-alpha.1.2' 'true'
  check_version ' v1.2.3 ' 'v1.2.3' 'false'
  # 不带 rc/beta/alpha 后缀的一律是正式版
  check_version 'v1.2.3-hotfix' 'v1.2.3-hotfix' 'false'
  check_version 'v2.0.0-core' 'v2.0.0-core' 'false'

  # 空版本号必须被拒绝：历史上空输入会打出 :v 这种非法标签
  if resolve_version '' 2>/dev/null; then
    printf '  ✗ %-20s 应当报错，实际通过\n' '<空>'
    failures=$((failures + 1))
  else
    printf '  ✓ %-20s 被拒绝（返回非 0）\n' '<空>'
  fi

  if [ "${failures}" -ne 0 ]; then
    printf '\n自检失败: %d 项\n' "${failures}" >&2
    exit 1
  fi

  printf '\n自检通过\n'
fi
