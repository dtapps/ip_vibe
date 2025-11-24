#!/bin/bash

# Docker Compose/Podman Compose 服务管理脚本
# 在当前目录下管理服务的启动、停止、重启和更新

# 定义颜色
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
RED="\033[0;31m"
NC="\033[0m" # No Color

# 检测可用的 compose 工具
COMPOSE_TOOL=""
if command -v docker &> /dev/null && docker compose version &> /dev/null; then
    COMPOSE_TOOL="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_TOOL="docker-compose"
elif command -v podman &> /dev/null && podman compose version &> /dev/null; then
    COMPOSE_TOOL="podman compose"
fi

if [ -z "$COMPOSE_TOOL" ]; then
    echo -e "${RED}错误: 未找到 docker compose、docker-compose 或 podman compose${NC}"
    exit 1
fi

echo -e "${GREEN}使用: ${COMPOSE_TOOL}${NC}"

# 帮助信息
show_help() {
    echo -e "${YELLOW}使用方法: $0 [命令]${NC}"
    echo -e "命令:"
    echo -e "  up        - 在当前目录启动服务"
    echo -e "  down      - 停止当前目录服务"
    echo -e "  restart   - 重启当前目录服务"
    echo -e "  update    - 更新并重启当前目录服务"
    echo -e "  status    - 查看当前目录服务状态"
    echo -e "示例:"
    echo -e "  $0 up      - 在当前目录启动服务"
    echo -e "  $0 restart - 重启当前目录服务"
    echo -e "  $0 status  - 查看当前目录服务状态"
}

# 检查当前目录是否有 docker-compose.yml 文件
check_compose_file() {
    if [ ! -f "docker-compose.yml" ]; then
        echo -e "${RED}错误: 当前目录未找到 docker-compose.yml 文件${NC}"
        return 1
    fi
    return 0
}

# 启动服务
start_service() {
    if ! check_compose_file; then
        return 1
    fi
    
    echo -e "${GREEN}在当前目录启动服务...${NC}"
    $COMPOSE_TOOL up -d
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}服务启动成功!${NC}"
    else
        echo -e "${RED}服务启动失败!${NC}"
    fi
}

# 停止服务
stop_service() {
    if ! check_compose_file; then
        return 1
    fi
    
    echo -e "${GREEN}停止当前目录服务...${NC}"
    $COMPOSE_TOOL down
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}服务停止成功!${NC}"
    else
        echo -e "${RED}服务停止失败!${NC}"
    fi
}

# 重启服务
restart_service() {
    if ! check_compose_file; then
        return 1
    fi
    
    echo -e "${GREEN}重启当前目录服务...${NC}"
    $COMPOSE_TOOL restart
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}服务重启成功!${NC}"
    else
        echo -e "${RED}服务重启失败!${NC}"
    fi
}

# 更新并重启服务
update_service() {
    if ! check_compose_file; then
        return 1
    fi
    
    echo -e "${GREEN}更新当前目录服务...${NC}"
    $COMPOSE_TOOL pull && $COMPOSE_TOOL up -d
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}服务更新成功!${NC}"
    else
        echo -e "${RED}服务更新失败!${NC}"
    fi
}

# 查看服务状态
status_service() {
    if ! check_compose_file; then
        return 1
    fi
    
    echo -e "${GREEN}查看当前目录服务状态...${NC}"
    $COMPOSE_TOOL ps
}

# 处理参数
if [ $# -ne 1 ]; then
    show_help
    exit 1
fi

COMMAND=$1

# 执行命令
case "$COMMAND" in
    "up") start_service ;;
    "down") stop_service ;;
    "restart") restart_service ;;
    "update") update_service ;;
    "status") status_service ;;
    *) 
        echo -e "${RED}未知命令: $COMMAND${NC}"
        show_help 
        exit 1
        ;;
esac