#!/bin/sh

# 健康检查接口地址
HEALTH_ENDPOINT="http://localhost:8080/health"

# 尝试请求健康检查接口
response=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_ENDPOINT)

# 检查响应状态码
if [ "$response" = "200" ]; then
    exit 0
else
    echo "健康检查失败: 状态码 $response"
    exit 1
fi 