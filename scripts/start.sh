#!/bin/bash

# 设置错误时退出
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO] $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}[WARN] $1${NC}"
}

log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

# 检查必要的命令是否存在
check_commands() {
    log_info "检查必要的命令..."
    commands=("go")
    for cmd in "${commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "$cmd 未安装，请先安装 $cmd"
            exit 1
        fi
    done
}

# 检查配置文件
check_config() {
    log_info "检查配置文件..."
    if [ ! -f "conf/config.yaml" ]; then
        log_error "配置文件 conf/config.yaml 不存在"
        exit 1
    fi
}

# 编译项目
build_project() {
    log_info "开始编译项目..."
    go build -o ./bin/bluebell
    if [ $? -ne 0 ]; then
        log_error "编译失败"
        exit 1
    fi
    log_info "编译成功"
}

# 启动服务
start_service() {
    log_info "启动服务..."
    log_info "使用本地方式启动服务..."
    ./bin/bluebell conf/config.yaml &
    echo $! > ./bin/bluebell.pid
}

# 主函数
main() {
    log_info "开始启动 Bluebell 服务..."
    
    # 检查命令
    check_commands
    
    # 检查配置
    check_config
    
    # 编译项目
    build_project
    
    # 启动服务
    start_service
    
    log_info "服务启动完成"
    log_info "API 文档地址: http://127.0.0.1:8084/swagger/index.html"
}

# 执行主函数
main "$@" 