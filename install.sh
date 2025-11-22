#!/bin/sh

# vibe_ip 安装/更新脚本
# 此脚本将自动检测系统架构并安装相应版本的 vibe_ip

reset="\033[0m"
red="\033[31m"
green="\033[32m"
yellow="\033[33m"
cyan="\033[36m"
white="\033[37m"
bold="\e[1m"
dim="\e[2m"

# 项目信息
PROJECT_NAME="vibe_ip"
CNB_REPO="dtapp/vibe_ip"
CNB_API="https://cnb.cool"

# 环境变量
if [ -z "$INSTALL_DIRECTORY" ]; then
    # 使用当前执行目录作为默认安装目录
    INSTALL_DIRECTORY="$(pwd)"
fi

# 输出消息函数
echo_info() {
    echo -e "${cyan}> ${reset}$1"
}

echo_warn() {
    echo -e "${yellow}> ${reset}$1"
}

echo_error() {
    echo -e "${red}> ${reset}$1"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查依赖
echo_info "检查必要的依赖..."
if ! command_exists "curl" && ! command_exists "wget"; then
    echo_error "请先安装 curl 或 wget 命令"
    exit 1
fi

# 检测架构
initArch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv8*) ARCH="arm64" ;;
        arm64) ARCH="arm64" ;;
        *)
            echo_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    echo_info "检测到架构: $ARCH"
}

# 检测操作系统
initOS() {
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        darwin) OS='darwin' ;;
        linux) OS='linux' ;;
        mingw*|msys*) OS='windows' ;;
        *)
            echo_error "不支持的操作系统: $OS"
            exit 1
            ;;
    esac
    echo_info "检测到操作系统: $OS"
    
    # 设置二进制文件后缀
    if [ "$OS" = "windows" ]; then
        BIN_SUFFIX=".exe"
    else
        BIN_SUFFIX=""
    fi
}

# 获取最新版本
get_latest_version() {
    echo_info "从CNB获取最新版本信息..."
    
    # 尝试获取版本信息（基于示例脚本的简化方法）
    # 在实际使用中，我们可能无法直接获取版本号，但可以尝试从页面提取
    if command_exists "curl"; then
        LATEST_VERSION=$(curl -s "$CNB_API/$CNB_REPO" | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
    elif command_exists "wget"; then
        LATEST_VERSION=$(wget -q -O - "$CNB_API/$CNB_REPO" | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
    fi
    
    # 如果无法获取，使用默认版本
    if [ -z "$LATEST_VERSION" ]; then
        echo_warn "无法从CNB获取最新版本信息，使用默认版本"
        LATEST_VERSION="v0.0.0"
    fi
    
    echo_info "最新版本: $LATEST_VERSION"
}

# 下载文件函数
download_file() {
    url="$1"
    destination="$2"

    echo "获取 $url.."
    if command_exists "curl"; then
        code=$(curl -w '%{http_code}' -L "$url" -o "$destination")
    elif command_exists "wget"; then
        code=$(wget --show-progress --progress=bar:force:noscroll -q -O "$destination" --server-response "$url" 2>&1 | awk '/^  HTTP/{print $2}' | tail -1)
    else
        echo_error "无法执行HTTP请求，curl和wget都不可用"
        exit 1
    fi

    if [ "$code" != 200 ]; then
        echo_error "文件下载失败，状态码: $code"
        return 1
    fi
    return 0
}

# 下载二进制文件
download() {
    echo_info "下载二进制文件..."
    
    # 构建下载URL（使用正确的CNB格式，包含版本号）
    if [ "$ARCH" = "arm64" ]; then
        BINARY_NAME="${PROJECT_NAME}-${OS}-arm64${BIN_SUFFIX}"
    else
        BINARY_NAME="${PROJECT_NAME}-${OS}-amd64${BIN_SUFFIX}"
    fi
    
    # 使用正确的CNB API格式，包含版本号
    BINARY_URL="${CNB_API}/${CNB_REPO}/-/releases/download/${LATEST_VERSION}/${BINARY_NAME}"
    echo_info "尝试下载: $BINARY_URL"
    
    # 创建临时文件
    DOWNLOAD_FILE=$(mktemp -t ${PROJECT_NAME}.XXXXXXXXXX)
    
    # 尝试下载
    if download_file "$BINARY_URL" "$DOWNLOAD_FILE"; then
        # 检查下载的文件是否为空
        if [ ! -s "$DOWNLOAD_FILE" ]; then
            echo_error "下载的文件为空"
            return 1
        else
            printf "\033[2A${cyan}> 下载最新版本... ✓${reset}\033[K\n"
            chmod +x "$DOWNLOAD_FILE"
            
            # 确保安装目录存在
            mkdir -p "$INSTALL_DIRECTORY"
            
            # 移动到安装目录（直接放在安装目录下，不创建bin子目录）
            mv "$DOWNLOAD_FILE" "$INSTALL_DIRECTORY/$PROJECT_NAME$BIN_SUFFIX"
            echo_info "二进制文件已下载到: $INSTALL_DIRECTORY/$PROJECT_NAME$BIN_SUFFIX"
            return 0
        fi
    else
        echo_error "从CNB下载失败"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
${PROJECT_NAME} 安装/更新脚本

用法:
  ./install.sh [选项]

环境变量:
  INSTALL_DIRECTORY (可选): 默认为当前执行目录

选项:
  -h, --help      显示此帮助信息
  -v, --version   显示脚本版本

示例:
  ./install.sh                # 安装或更新到最新版本
  INSTALL_DIRECTORY=/opt/${PROJECT_NAME} ./install.sh  # 指定安装目录

EOF
}

# 安装主函数
install() {
    echo_info "安装 $PROJECT_NAME!"
    
    # 识别平台
    initOS
    initArch
    
    # 获取最新版本
    get_latest_version
    
    # 下载二进制文件
    if ! download; then
        echo_error "下载失败，安装中断"
        exit 1
    fi
    
        echo ""
    echo -e "${green}======================================${reset}"
    echo -e "${green}${PROJECT_NAME} 安装/更新完成!${reset}"
    echo ""
    echo -e "${green}📌 配置文件位于: $INSTALL_DIRECTORY/config.yaml${reset}"
    echo ""
    echo -e "${green}🔧 首次使用请按顺序操作：${reset}"
    echo -e "${green}  1. 初始化配置: ${PROJECT_NAME} config init${reset}"
    echo -e "${green}  2. 安装服务:   ${PROJECT_NAME} service install${reset}"
    echo -e "${green}  3. 启动服务:   sudo systemctl start ${PROJECT_NAME}${reset}"
    echo -e "${green}  4. 开机自启（可选）: sudo systemctl enable ${PROJECT_NAME}${reset}"
    echo ""
    echo -e "${yellow}🗑️ 卸载服务时请完整执行：${reset}"
    echo -e "${yellow}  1. 停止服务:   sudo systemctl stop ${PROJECT_NAME}${reset}"
    echo -e "${yellow}  2. 卸载服务:   ${PROJECT_NAME} service uninstall${reset}"
    echo -e "${yellow}  3. 重载 systemd: sudo systemctl daemon-reload${reset}"
    echo ""
    echo -e "${cyan}ℹ️  查看帮助: ${PROJECT_NAME} --help${reset}"
    echo -e "${green}======================================${reset}"
    echo ""
}

# 主函数
main() {
    # 处理命令行参数
    while [ $# -gt 0 ]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                echo "${PROJECT_NAME} 安装脚本 v1.0.0"
                exit 0
                ;;
            *)
                echo_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
        shift
done
    
    # 运行安装
    install
}

# 运行主函数
main "$@"