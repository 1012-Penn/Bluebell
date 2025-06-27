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

# 停止本地服务
stop_local_service() {
    log_info "停止本地服务..."
    if [ -f "./bin/bluebell.pid" ]; then
        pid=$(cat ./bin/bluebell.pid)
        if ps -p $pid > /dev/null; then
            kill $pid
            rm ./bin/bluebell.pid
            log_info "服务已停止"
        else
            log_warn "服务未运行"
            rm ./bin/bluebell.pid
        fi
    else
        log_warn "未找到服务进程ID文件"
    fi
}

# 主函数
main() {
    log_info "开始停止 Bluebell 服务..."
    
    stop_local_service
    
    log_info "服务停止完成"
}

# 执行主函数
main "$@" 