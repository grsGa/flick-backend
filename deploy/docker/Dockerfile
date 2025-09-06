# Stage 1: Build the Go applications
FROM golang:1.24 AS builder

# 安装构建依赖和媒体处理工具
RUN apt-get update && apt-get install -y \
    git \
    imagemagick \
    libmagickwand-dev \
    ffmpeg

WORKDIR /app

# Copy the entire backend source code with unified go.mod
COPY . .

# Set environment variables
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOPRIVATE=github.com/flick/backend
ENV GO111MODULE=on

# Download dependencies using the unified go.mod
RUN go mod download

# Build each service using the unified go.mod from root directory
RUN go build -o /app/gateway ./services/gateway
RUN go build -o /app/auth ./services/auth
RUN go build -o /app/user ./services/user
RUN go build -o /app/media ./services/media
RUN go build -o /app/content ./services/content
RUN go build -o /app/interaction ./services/interaction
RUN go build -o /app/messages ./services/messages
RUN go build -o /app/notification ./services/notification
RUN go build -o /app/recommendation ./services/recommendation
RUN go build -o /app/search ./services/search

WORKDIR /app

# Download wait-for-it.sh
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh

# Install grpc_health_probe
RUN go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.11

# Stage 2: Create the final image
FROM debian:stable-slim

# 安装运行时依赖，包括媒体处理工具
RUN apt-get update && apt-get install -y \
    bash \
    ca-certificates \
    imagemagick \
    ffmpeg \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建非root用户
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -m appuser

# 创建媒体处理临时目录
RUN mkdir -p /tmp/media-processing && \
    chown -R appuser:appgroup /tmp/media-processing

WORKDIR /app

# Copy the built binaries and wait-for-it.sh from the builder stage
COPY --from=builder /app/gateway /app/gateway
COPY --from=builder /app/auth /app/auth
COPY --from=builder /app/user /app/user
COPY --from=builder /app/media /app/media
COPY --from=builder /app/content /app/content
COPY --from=builder /app/interaction /app/interaction
COPY --from=builder /app/messages /app/messages
COPY --from=builder /app/notification /app/notification
COPY --from=builder /app/recommendation /app/recommendation
COPY --from=builder /app/search /app/search
COPY --from=builder /usr/local/bin/wait-for-it.sh /usr/local/bin/wait-for-it.sh
COPY --from=builder /go/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

# The command to run will be specified in the docker-compose.yml
# Example: CMD ["/app/gateway"]
