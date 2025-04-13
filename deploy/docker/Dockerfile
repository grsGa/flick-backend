# 多阶段构建 - 基础构建阶段
FROM golang:1.24-alpine3.21 AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

# 先复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建服务 - 根据服务类型选择不同的构建方式
ARG SERVICE=gateway
RUN mkdir -p /app/bin && \
    case $SERVICE in \
        "gateway") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/gateway ;; \
        "user") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/user ;; \
        "content") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/content ;; \
        "recommendation") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/recommendation ;; \
        "notification") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/notification ;; \
        "interaction") \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./services/interaction ;; \
        *) \
            echo "Unknown service: $SERVICE" && exit 1 ;; \
    esac

# 生产阶段镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl \
    && update-ca-certificates

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 复制编译好的应用和其他需要的文件
COPY --from=builder /app/bin/service /app/service
COPY deploy/docker/health-check.sh /app/health-check.sh
RUN chmod +x /app/health-check.sh

# 设置用户
USER appuser

# 设置环境变量
ENV GIN_MODE=release

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["/app/service"]