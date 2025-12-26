#!/bin/bash

# 项目流程管理系统 - Docker部署脚本
# 使用方法: ./deploy.sh [命令]
# 命令: start, stop, restart, logs, build, clean

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 项目名称
PROJECT_NAME="project-flow"

# 显示帮助
show_help() {
    echo -e "${GREEN}项目流程管理系统 - Docker部署脚本${NC}"
    echo ""
    echo "使用方法: ./deploy.sh [命令]"
    echo ""
    echo "命令:"
    echo "  start     - 启动服务"
    echo "  stop      - 停止服务"
    echo "  restart   - 重启服务"
    echo "  build     - 重新构建并启动"
    echo "  logs      - 查看日志"
    echo "  status    - 查看服务状态"
    echo "  clean     - 清理容器和镜像"
    echo "  backup    - 备份数据"
    echo "  help      - 显示帮助"
}

# 检查Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker未安装${NC}"
        exit 1
    fi
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}错误: Docker Compose未安装${NC}"
        exit 1
    fi
}

# 获取docker compose命令
get_compose_cmd() {
    if docker compose version &> /dev/null 2>&1; then
        echo "docker compose"
    else
        echo "docker-compose"
    fi
}

# 创建环境文件
setup_env() {
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            echo -e "${YELLOW}已创建 .env 文件，请根据需要修改配置${NC}"
        fi
    fi
}

# 创建数据目录
setup_dirs() {
    mkdir -p data/db data/uploads
    echo -e "${GREEN}数据目录已创建${NC}"
}

# 启动服务
start() {
    check_docker
    setup_env
    setup_dirs
    
    echo -e "${GREEN}正在启动服务...${NC}"
    $(get_compose_cmd) up -d
    
    echo -e "${GREEN}服务已启动!${NC}"
    echo -e "访问地址: http://localhost:${APP_PORT:-80}"
    echo -e "默认账号: admin"
    echo -e "默认密码: Admin@123"
}

# 停止服务
stop() {
    check_docker
    echo -e "${YELLOW}正在停止服务...${NC}"
    $(get_compose_cmd) down
    echo -e "${GREEN}服务已停止${NC}"
}

# 重启服务
restart() {
    stop
    start
}

# 重新构建
build() {
    check_docker
    setup_env
    setup_dirs
    
    echo -e "${GREEN}正在构建镜像...${NC}"
    $(get_compose_cmd) build --no-cache
    
    echo -e "${GREEN}正在启动服务...${NC}"
    $(get_compose_cmd) up -d
    
    echo -e "${GREEN}构建完成，服务已启动!${NC}"
}

# 查看日志
logs() {
    check_docker
    $(get_compose_cmd) logs -f
}

# 查看状态
status() {
    check_docker
    $(get_compose_cmd) ps
}

# 清理
clean() {
    check_docker
    echo -e "${YELLOW}正在清理...${NC}"
    $(get_compose_cmd) down -v --rmi local
    echo -e "${GREEN}清理完成${NC}"
}

# 备份数据
backup() {
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    echo -e "${GREEN}正在备份数据...${NC}"
    
    if [ -d "data/db" ]; then
        cp -r data/db "$BACKUP_DIR/"
    fi
    
    if [ -d "data/uploads" ]; then
        cp -r data/uploads "$BACKUP_DIR/"
    fi
    
    echo -e "${GREEN}备份完成: $BACKUP_DIR${NC}"
}

# 主逻辑
case "${1:-help}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    build)
        build
        ;;
    logs)
        logs
        ;;
    status)
        status
        ;;
    clean)
        clean
        ;;
    backup)
        backup
        ;;
    help|*)
        show_help
        ;;
esac
