#!/bin/sh
#
# 生成 Release 正文（Markdown）。
#
# 事实来源：CNB 的 .cnb/workflows/release.yml 与 GitHub 的 .github/workflows/release.yml
# 都调用这一份，避免两边各写一遍表格/标题/下载链接，改一处忘一处。
#
# 只用环境变量传参（不解析位置参数），这样两条流水线都能直接 `sh scripts/ci/render-release-body.sh`：
#
#   VERSION       必填，归一化后的版本号（带 v 前缀），如 v1.2.3-rc.1
#   PRERELEASE    必填，true / false，决定是否加「预发布」警示块
#   DL_BASE       必填，附件下载基址，如 https://github.com/o/r/releases/download/v1.2.3
#   STABLE_URL    选填，正式版列表页 URL，预发布时提示用户去哪找稳定版
#   ARTIFACT_DIR  选填，产物目录，默认 bin
#   ARTIFACT_LAYOUT 选填，产物命名布局，默认 single：
#                 single —— 单二进制项目（ikuai-{os}-{arch}），下载表 4 列；
#                 triplet —— 多服务项目（{服务}-{os}-{arch}），下载表多一列「组件」，
#                            并按「组件 → 平台 → 架构」排序，方便用户按服务找包。
#   OUT_FILE      选填，输出文件，默认 release_body.md
#   CHANGELOG     选填，更新内容（可多行）；为空时输出「暂无更新内容」
#   COMMIT_SHA    选填，提交 SHA；有则加一行提交链接
#   COMMIT_URL    选填，提交链接前缀，拼成 ${COMMIT_URL}/<sha>
#                 GitHub 是 ${REPO_URL}/commit，CNB 是 ${REPO_URL}/-/commit
#   REPO_URL      选填，仓库地址（仅作 COMMIT_URL 缺省时的回退）
#   BUILD_TIME    选填，构建时间，默认取当前 UTC 时间
#
# 自检（改了格式先跑它）：
#   sh scripts/ci/render-release-body.sh --self-test
#
# ⚠️ 必须保持 POSIX sh 语法：CNB 的 stage 默认用 sh（dash/ash）执行。

set -eu

# 产物名解析。输出三个全局变量：COMP / PLAT / ARCH
#
# 两种命名布局由 ARTIFACT_LAYOUT 决定（默认 single，与 ikuai 一致）：
#   single  —— 单二进制：{name}-{os}-{arch}[.exe].{tar.gz,zip}，如 ikuai-linux-amd64.tar.gz
#              首段是项目名，不是组件，「组件」列没有意义，所以下载表不输出该列
#   triplet —— 多服务：{comp}-{os}-{arch}[.exe].{tar.gz,zip}，如 backend-linux-amd64.tar.gz
#              首段是服务名（backend/reporter/forwarder），需要单独一列区分
parse_artifact_name() {
  fname="$1"
  COMP=''
  PLAT='其它'
  ARCH='-'

  # 先去掉压缩后缀，拿到 xxx-linux-amd64 / xxx-windows-amd64.exe 这种主干
  case "${fname}" in
    *.tar.gz) stem="${fname%.tar.gz}" ;;
    *.zip) stem="${fname%.zip}" ;;
    *) return 0 ;;
  esac
  # Windows 产物主干还带 .exe，一并去掉再切分
  case "${stem}" in
    *.exe) stem="${stem%.exe}" ;;
  esac

  # 主干按 - 切成三段：首段-os-arch
  comp="${stem%%-*}"
  rest="${stem#*-}"
  os="${rest%%-*}"
  arch="${rest#*-}"

  # single 布局下首段是项目名（ikuai / ip_vibe），不是组件，清空免得表格里出现无意义的一列
  case "${layout}" in
    single) COMP='' ;;
    *) COMP="${comp}" ;;
  esac

  case "${os}" in
    linux) PLAT='Linux' ;;
    windows) PLAT='Windows' ;;
    darwin) PLAT='macOS' ;;
    *) PLAT="${os}" ;;
  esac

  case "${arch}" in
    amd64) ARCH='x86_64' ;;
    arm64) ARCH='ARM64' ;;
    *) ARCH="${arch}" ;;
  esac
}

render_body() {
  : "${VERSION:?VERSION 为空}"
  : "${PRERELEASE:?PRERELEASE 为空}"
  : "${DL_BASE:?DL_BASE 为空}"

  artifact_dir="${ARTIFACT_DIR:-bin}"
  out_file="${OUT_FILE:-release_body.md}"
  # 默认 single：保持 ikuai 的既有行为（4 列，无「组件」列）
  layout="${ARTIFACT_LAYOUT:-single}"
  case "${layout}" in
    single | triplet) ;;
    *)
      printf '❌ ARTIFACT_LAYOUT 只能是 single 或 triplet，收到: %s\n' "${layout}" >&2
      return 1
      ;;
  esac

  if [ ! -d "${artifact_dir}" ]; then
    printf '❌ 产物目录不存在: %s\n' "${artifact_dir}" >&2
    return 1
  fi

  if [ "${PRERELEASE}" = 'true' ]; then
    release_type='🚧 预发布版本 (Pre-release)'
  else
    release_type='✅ 正式发布版本 (Stable)'
  fi
  build_time="${BUILD_TIME:-$(date -u '+%Y-%m-%d %H:%M:%S UTC')}"

  # 提交号保留完整 SHA 作为**链接目标**（缩写长度随仓库规模变化，链接里用缩写
  # 部分平台会 404）；另算一个短号只用于**显示**，和 GitHub 页面习惯一致。
  if [ -n "${COMMIT_SHA:-}" ]; then
    COMMIT_SHA="$(printf '%s' "${COMMIT_SHA}" | tr -d '[:space:]')"
    COMMIT_SHORT="$(printf '%s' "${COMMIT_SHA}" | cut -c1-7)"
  fi

  # 统计可发布的附件：排除 SHA256SUMS 本身，避免表格里出现一行校验文件
  artifact_count=0
  # 校验节要给 Windows 用户写出具体文件名，这里顺手把 Windows 产物收集起来，
  # 避免在正文里硬编码某个架构的文件名（架构变了正文就会写错）。
  windows_artifacts=''
  for f in "${artifact_dir}"/*; do
    [ -e "${f}" ] || continue
    [ -d "${f}" ] && continue
    # 与下方下载表保持同一套过滤规则，否则「有附件」的判断会和表格内容打架
    case "$(basename "${f}")" in
      *.tar.gz | *.zip) ;;
      *) continue ;;
    esac
    artifact_count=$((artifact_count + 1))
  done

  {
    # 预发布在最前面单独警示：正文往下滚才发现是 rc，人已经点了下载
    if [ "${PRERELEASE}" = 'true' ]; then
      echo '## ⚠️ 预发布版本'
      echo ''
      echo '> 这是预发布版本，可能包含未充分测试的功能，不建议用于生产环境。'
      if [ -n "${STABLE_URL:-}" ]; then
        printf '> 如需稳定版本，请使用 [最新正式版](%s)。\n' "${STABLE_URL}"
      fi
      echo ''
    fi

    echo '## 发布信息'
    echo ''
    echo '| 项目 | 内容 |'
    echo '| --- | --- |'
    printf '| 版本号 | %s |\n' "${VERSION}"
    printf '| 发布类型 | %s |\n' "${release_type}"
    # 提交链接前缀：COMMIT_URL 优先（CNB 的 Web 路径带 /-/，和 GitHub 不同），
    # 没传时退回 GitHub 风格的 ${REPO_URL}/commit。
    _commit_url="${COMMIT_URL:-}"
    if [ -z "${_commit_url}" ] && [ -n "${REPO_URL:-}" ]; then
      _commit_url="${REPO_URL}/commit"
    fi
    if [ -n "${COMMIT_SHA:-}" ] && [ -n "${_commit_url}" ]; then
      printf '| 提交 | [%s](%s/%s) |\n' "${COMMIT_SHORT}" "${_commit_url}" "${COMMIT_SHA}"
    fi
    printf '| 构建时间 | %s |\n' "${build_time}"
    echo ''

    echo '## 下载'
    echo ''
    if [ "${artifact_count}" -eq 0 ]; then
      echo '> 本次未产出可下载的附件。'
    else
      # 先算一遍用于排序的键，再依次输出：
      # 直接遍历目录会按文件系统顺序排（ext4 上接近随机），同一份产物在 CI 和本地
      # 生成的正文行序会不一样。single 布局按 平台→架构，triplet 布局按 组件→平台→架构。
      rows_file="$(mktemp "${TMPDIR:-/tmp}/relbody-rows.XXXXXX")"
      for f in "${artifact_dir}"/*; do
        [ -e "${f}" ] || continue
        [ -d "${f}" ] && continue
        fname="$(basename "${f}")"
        # 只收发布产物：产物目录里可能混进正文 md、日志、临时文件（GitHub 侧的
        # OUT_FILE 就常常和附件同放一个目录），不筛会把它们当成可下载附件列出去。
        case "${fname}" in
          *.tar.gz | *.zip) ;;
          *) continue ;;
        esac
        # 用 wc -c 取字节数再换算，避免 ls 在非 ASCII 文件名下的解析问题（SC2012）
        fbytes="$(wc -c <"${f}" | tr -d ' ')"
        if [ "${fbytes}" -ge 1048576 ]; then
          fsize="$((fbytes / 1048576))M"
        elif [ "${fbytes}" -ge 1024 ]; then
          fsize="$((fbytes / 1024))K"
        else
          fsize="${fbytes}B"
        fi
        parse_artifact_name "${fname}"
        if [ "${PLAT}" = 'Windows' ]; then
          windows_artifacts="${windows_artifacts}${windows_artifacts:+
}${fname}"
        fi
        # 排序键放在前两段，第三段开始是真正的表格行：
        # 组件用字典序、平台按 linux→windows→darwin 的固定次序（不能直接字典序，
        # 否则 darwin 会排在 linux 前面，和惯例相反），架构 amd64 在 arm64 前。
        case "${PLAT}" in
          Linux) _p=1 ;;
          Windows) _p=2 ;;
          macOS) _p=3 ;;
          *) _p=9 ;;
        esac
        case "${ARCH}" in
          x86_64) _a=1 ;;
          ARM64) _a=2 ;;
          *) _a=9 ;;
        esac
        # 排序键放前 3 段，第 4 段起是完整的表格行（含行首竖线），输出时 cut -f4- 取回。
        # ⚠️ 组件列必须拼在「第 4 段」里，而不是当成排序键之外的单列：
        #    排序键 f1 就是组件，若再指望 cut -f4- 把组件带出来，组件会被整列切掉。
        if [ "${layout}" = 'single' ]; then
          row="| ${PLAT} | ${ARCH} | [${fname}](${DL_BASE}/${fname}) | ${fsize} |"
        else
          row="| ${COMP} | ${PLAT} | ${ARCH} | [${fname}](${DL_BASE}/${fname}) | ${fsize} |"
        fi
        printf '%s\t%s\t%s\t%s\n' "${COMP}" "${_p}" "${_a}" "${row}" >>"${rows_file}"
      done

      if [ "${layout}" = 'single' ]; then
        echo '| 平台 | 架构 | 文件 | 大小 |'
        echo '| --- | --- | --- | --- |'
        LC_ALL=C sort -t "$(printf '\t')" -k2,2n -k3,3n "${rows_file}" \
          | cut -f4-
      else
        echo '| 组件 | 平台 | 架构 | 文件 | 大小 |'
        echo '| --- | --- | --- | --- | --- |'
        LC_ALL=C sort -t "$(printf '\t')" -k1,1 -k2,2n -k3,3n "${rows_file}" \
          | cut -f4-
      fi
      rm -f "${rows_file}"
    fi
    echo ''

    # 有 SHA256SUMS 才写校验说明，避免正文承诺一个不存在的文件
    if [ -f "${artifact_dir}/SHA256SUMS" ]; then
      # 这里一律用 printf '%s\n' 而不是 echo：
      # dash 的 echo 会解释反斜杠转义，Windows 路径的 `.\xxx` 里若出现 \b 会被吞成退格符
      # （`.\backend-...` 显示成 `.ackend-...`），printf 不做任何转义解释。
      # 反引号是 Markdown 代码标记，不是命令替换，单引号正是要它原样输出（SC2016 属误报）
      # shellcheck disable=SC2016
      printf '%s\n' '## 校验'
      printf '\n'
      # shellcheck disable=SC2016
      printf '%s\n' '附件含 `SHA256SUMS`（`sha256sum` 标准格式，每行「哈希 + 两个空格 + 文件名」）。'
      # shellcheck disable=SC2016
      printf '%s\n' '把 `SHA256SUMS` 和压缩包下载到同一目录，然后按平台执行：'
      printf '\n'
      # shellcheck disable=SC2016
      printf '%s\n' '- **Linux**：`sha256sum -c SHA256SUMS`'
      # shellcheck disable=SC2016
      printf '%s\n' '- **macOS**：`shasum -a 256 -c SHA256SUMS`'
      if [ -n "${windows_artifacts}" ]; then
        # 逐个架构给出命令：PowerShell 一次只校验一个文件，只写一个会漏掉另一个架构
        while IFS= read -r wf; do
          [ -n "${wf}" ] || continue
          printf '%s\n' "- **Windows（PowerShell）**：\`Get-FileHash .\\${wf} -Algorithm SHA256\`"
        done <<EOF
${windows_artifacts}
EOF
        # shellcheck disable=SC2016
        printf '%s\n' '  PowerShell 输出为全大写，与 `SHA256SUMS` 里的小写哈希等价，逐位比对即可。'
        # shellcheck disable=SC2016
        printf '%s\n' '  也可用 `certutil -hashfile <文件> SHA256`。'
      fi
      printf '\n'
    fi

    echo '## 更新内容'
    echo ''
    if [ -n "${CHANGELOG:-}" ]; then
      printf '%s\n' "${CHANGELOG}"
    else
      echo '暂无更新内容。'
    fi
  } >"${out_file}"

  cat "${out_file}"
}

# 直接执行才渲染/自检；被 source 时只导出函数。
# 判定用 $0 的 basename（不能用 BASH_SOURCE，dash 会 Bad substitution）。
if [ "${0##*/}" != 'render-release-body.sh' ] && [ "${1:-}" != '--self-test' ]; then
  # shellcheck disable=SC2317
  return 0 2>/dev/null || exit 0
fi

if [ "${1:-}" = '--self-test' ]; then
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "${tmp_dir}"' EXIT

  mkdir -p "${tmp_dir}/bin"
  : >"${tmp_dir}/bin/ikuai-linux-amd64.tar.gz"
  : >"${tmp_dir}/bin/ikuai-linux-arm64.tar.gz"
  : >"${tmp_dir}/bin/ikuai-windows-amd64.exe.zip"
  : >"${tmp_dir}/bin/ikuai-darwin-arm64.tar.gz"
  : >"${tmp_dir}/bin/SHA256SUMS"

  failures=0
  expect() {
    # expect <文件> <期望出现的行> <说明>
    if grep -qF -- "$2" "$1"; then
      printf '  ✓ %s\n' "$3"
    else
      printf '  ✗ %s（未找到: %s）\n' "$3" "$2"
      failures=$((failures + 1))
    fi
  }
  reject() {
    if grep -qF -- "$2" "$1"; then
      printf '  ✗ %s（不应出现: %s）\n' "$3" "$2"
      failures=$((failures + 1))
    else
      printf '  ✓ %s\n' "$3"
    fi
  }

  echo '── 正式版 ──'
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl/v1.2.3' \
    ARTIFACT_DIR="${tmp_dir}/bin" OUT_FILE="${tmp_dir}/stable.md" CHANGELOG='- 修复 A
- 优化 B' \
    render_body >/dev/null
  expect "${tmp_dir}/stable.md" '| 版本号 | v1.2.3 |' '版本号正确'
  # 预发布与否只看「发布类型」行：原来还有一行 `是否预发布 | true/false`，
  # 但把原始布尔值摆给用户没有意义（发布类型那行已经说清了），按参考格式删掉了。
  reject "${tmp_dir}/stable.md" '是否预发布' '正式版不出现多余的预发布字段行'
  expect "${tmp_dir}/stable.md" '✅ 正式发布版本 (Stable)' '正式版标记'
  expect "${tmp_dir}/stable.md" '| 构建时间 |' '构建时间行'
  expect "${tmp_dir}/stable.md" '| Linux | x86_64 | [ikuai-linux-amd64.tar.gz](https://example.com/dl/v1.2.3/ikuai-linux-amd64.tar.gz) |' 'Linux x86_64 下载行'
  expect "${tmp_dir}/stable.md" '| Windows | x86_64 | [ikuai-windows-amd64.exe.zip]' 'Windows 下载行'
  expect "${tmp_dir}/stable.md" '| macOS | ARM64 | [ikuai-darwin-arm64.tar.gz]' 'macOS 下载行'
  expect "${tmp_dir}/stable.md" '- 修复 A' '更新内容保留'
  expect "${tmp_dir}/stable.md" '## 校验' '有 SHA256SUMS 时写校验节'
  expect "${tmp_dir}/stable.md" 'sha256sum -c SHA256SUMS' 'Linux 校验命令'
  expect "${tmp_dir}/stable.md" 'shasum -a 256 -c SHA256SUMS' 'macOS 校验命令'
  # 反斜杠必须原样出现在正文里：dash 的 echo 会把它当转义，\b 会吞成退格符
  expect "${tmp_dir}/stable.md" 'Get-FileHash .\ikuai-windows-amd64.exe.zip -Algorithm SHA256' 'Windows 校验命令（反斜杠未被吞）'
  expect "${tmp_dir}/stable.md" 'certutil -hashfile' 'certutil 备选命令'
  reject "${tmp_dir}/stable.md" '## ⚠️ 预发布版本' '正式版不出现预发布警示'
  reject "${tmp_dir}/stable.md" 'SHA256SUMS |' '校验文件不进下载表格'
  reject "${tmp_dir}/stable.md" '| 说明 |' '下载表格不出现说明列'
  expect "${tmp_dir}/stable.md" '| 平台 | 架构 | 文件 | 大小 |' '下载表头四列'

  echo '── 预发布 ──'
  # CHANGELOG 显式置空：上一条用例给 render_body 传过 CHANGELOG，
  # 变量赋值会残留在当前 shell，不清掉这一条就走不到「无更新内容」兜底。
  VERSION='v1.2.3-rc.1' PRERELEASE='true' DL_BASE='https://example.com/dl/v1.2.3-rc.1' \
    STABLE_URL='https://example.com/releases/latest' \
    ARTIFACT_DIR="${tmp_dir}/bin" OUT_FILE="${tmp_dir}/pre.md" CHANGELOG='' \
    render_body >/dev/null
  expect "${tmp_dir}/pre.md" '## ⚠️ 预发布版本' '预发布警示块'
  reject "${tmp_dir}/pre.md" '是否预发布' '预发布版不出现多余的预发布字段行'
  expect "${tmp_dir}/pre.md" '🚧 预发布版本 (Pre-release)' '预发布标记'
  expect "${tmp_dir}/pre.md" '[最新正式版](https://example.com/releases/latest)' '稳定版指引链接'
  expect "${tmp_dir}/pre.md" '暂无更新内容。' '无 changelog 时的兜底文案'

  echo '── 提交号完整 SHA ──'
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl/v1.2.3' \
    ARTIFACT_DIR="${tmp_dir}/bin" OUT_FILE="${tmp_dir}/sha.md" \
    COMMIT_SHA='0123456789abcdef0123456789abcdef01234567' REPO_URL='https://example.com/repo' \
    render_body >/dev/null
  expect "${tmp_dir}/sha.md" \
    '[0123456](https://example.com/repo/commit/0123456789abcdef0123456789abcdef01234567)' \
    '显示短号、链接用完整 SHA'

  echo '── triplet 布局（多服务项目）──'
  # 三服务产物名 {服务}-{os}-{arch}：必须多出「组件」列，且按 组件→平台→架构 排序
  mkdir -p "${tmp_dir}/triplet"
  : >"${tmp_dir}/triplet/backend-windows-amd64.exe.zip"
  : >"${tmp_dir}/triplet/backend-linux-amd64.tar.gz"
  : >"${tmp_dir}/triplet/reporter-linux-amd64.tar.gz"
  : >"${tmp_dir}/triplet/forwarder-darwin-arm64.tar.gz"
  : >"${tmp_dir}/triplet/SHA256SUMS"
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl/v1.2.3' \
    ARTIFACT_DIR="${tmp_dir}/triplet" OUT_FILE="${tmp_dir}/triplet.md" CHANGELOG='- x' \
    ARTIFACT_LAYOUT='triplet' \
    render_body >/dev/null
  expect "${tmp_dir}/triplet.md" '| 组件 | 平台 | 架构 | 文件 | 大小 |' 'triplet 表头含组件列'
  expect "${tmp_dir}/triplet.md" '| backend | Linux | x86_64 | [backend-linux-amd64.tar.gz]' 'triplet 组件列=backend'
  expect "${tmp_dir}/triplet.md" '| reporter | Linux | x86_64 | [reporter-linux-amd64.tar.gz]' 'triplet 组件列=reporter'
  expect "${tmp_dir}/triplet.md" '| forwarder | macOS | ARM64 | [forwarder-darwin-arm64.tar.gz]' 'triplet 组件列=forwarder'
  # 排序：组件字典序 backend < forwarder < reporter，且 Linux 排在 Windows 前
  first_row="$(grep -E '^\| [a-z]+ \|' "${tmp_dir}/triplet.md" | head -n 1)"
  case "${first_row}" in
    '| backend | Linux |'*)
      printf '  ✓ %s\n' 'triplet 行序：backend 在最前、Linux 先于 Windows'
      ;;
    *)
      printf '  ✗ %s（首行: %s）\n' 'triplet 行序：backend 在最前、Linux 先于 Windows' "${first_row}"
      failures=$((failures + 1))
      ;;
  esac
  reject "${tmp_dir}/triplet.md" 'SHA256SUMS |' 'triplet 也不把校验文件列进表格'
  # 默认布局（未设置 ARTIFACT_LAYOUT）必须是 single，保证 ikuai 那边行为不变。
  # ⚠️ 必须显式 unset：自检全在同一个 shell 里跑，上一用例的 ARTIFACT_LAYOUT 会残留，
  #    不清理的话这条测的是「上一条的布局」而不是默认值。
  unset ARTIFACT_LAYOUT
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl/v1.2.3' \
    ARTIFACT_DIR="${tmp_dir}/bin" OUT_FILE="${tmp_dir}/default.md" CHANGELOG='- x' \
    render_body >/dev/null
  reject "${tmp_dir}/default.md" '| 组件 |' '默认布局仍是 4 列（不含组件列）'

  echo '── 产物目录混入非产物文件 ──'
  # GitHub 侧的 release_body.md 常和附件同目录；只有 .tar.gz/.zip 才算可下载附件
  mkdir -p "${tmp_dir}/mixed"
  : >"${tmp_dir}/mixed/backend-linux-amd64.tar.gz"
  : >"${tmp_dir}/mixed/release_body.md"
  : >"${tmp_dir}/mixed/pull.log"
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl' \
    ARTIFACT_DIR="${tmp_dir}/mixed" OUT_FILE="${tmp_dir}/mixed.md" CHANGELOG='- x' \
    ARTIFACT_LAYOUT='triplet' \
    render_body >/dev/null
  expect "${tmp_dir}/mixed.md" 'backend-linux-amd64.tar.gz' '产物进表格'
  reject "${tmp_dir}/mixed.md" 'release_body.md' '正文 md 不进表格'
  reject "${tmp_dir}/mixed.md" 'pull.log' '日志文件不进表格'

  echo '── 非法 ARTIFACT_LAYOUT ──'
  if VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl' \
    ARTIFACT_DIR="${tmp_dir}/bin" OUT_FILE="${tmp_dir}/bogus.md" ARTIFACT_LAYOUT='bogus' \
    render_body >/dev/null 2>&1; then
    printf '  ✗ %s\n' '非法布局应当失败'
    failures=$((failures + 1))
  else
    printf '  ✓ %s\n' '非法布局被拒绝'
  fi
  unset ARTIFACT_LAYOUT

  echo '── 空产物目录 ──'
  mkdir -p "${tmp_dir}/empty"
  VERSION='v1.2.3' PRERELEASE='false' DL_BASE='https://example.com/dl/v1.2.3' \
    ARTIFACT_DIR="${tmp_dir}/empty" OUT_FILE="${tmp_dir}/empty.md" \
    render_body >/dev/null
  expect "${tmp_dir}/empty.md" '> 本次未产出可下载的附件。' '空目录提示而不是空表格'
  reject "${tmp_dir}/empty.md" '| 平台 | 架构 |' '空目录不画表头'

  if [ "${failures}" -ne 0 ]; then
    printf '\n自检失败: %d 项\n' "${failures}" >&2
    exit 1
  fi
  printf '\n自检通过\n'
  exit 0
fi

render_body
