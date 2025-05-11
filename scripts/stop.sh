#!/bin/bash
set -e

if [ -f "./bin/bluebell.pid" ]; then
    pid=$(cat ./bin/bluebell.pid)
    kill $pid 2>/dev/null || true
    rm ./bin/bluebell.pid
    echo "服务已停止"
else
    echo "未找到服务进程ID文件"
fi 