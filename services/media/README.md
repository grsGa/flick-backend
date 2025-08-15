# Media Service

Media Service是Flick社交媒体平台的媒体文件管理微服务，负责处理用户头像、横幅和帖子媒体文件的上传、存储、访问和删除。

## 功能特性

### 核心功能
- **文件上传**: 支持头像、横幅、帖子媒体文件上传
- **文件存储**: 使用MinIO对象存储，支持分布式存储
- **文件访问**: 生成预签名URL，支持安全的文件访问
- **文件删除**: 支持文件的安全删除
- **文件列表**: 支持用户文件列表查询

### 媒体分类和限制
| 分类 | 最大文件大小 | 支持格式 | 用途 |
|------|-------------|----------|------|
| `avatars` | 5MB | JPG, PNG, WebP, GIF | 用户头像 |
| `banners` | 10MB | JPG, PNG, WebP, GIF | 用户横幅 |
| `posts` | 100MB | JPG, PNG, WebP, GIF, MP4, WebM, MOV, AVI | 帖子媒体 |

### 技术架构
- **存储后端**: MinIO对象存储
- **通信协议**: gRPC
- **服务发现**: Consul注册中心
- **健康检查**: gRPC健康检查协议
- **配置管理**: Viper + 环境变量

## 环境变量配置

```bash
# 服务端口
MEDIA_SERVICE_PORT=50054

# Consul配置
CONSUL_AGENT_ADDR=consul:8500

# MinIO配置
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=flick-media
```

## gRPC接口

### UploadFile - 上传文件
```protobuf
rpc UploadFile(UploadFileRequest) returns (UploadFileResponse);

message UploadFileRequest {
  string user_id = 1;
  bytes file_data = 2;
  string filename = 3;
  string content_type = 4;
  string category = 5; // avatars, banners, posts
  string alt_text = 6;
}
```

### GetFile - 获取文件信息
```protobuf
rpc GetFile(GetFileRequest) returns (GetFileResponse);

message GetFileRequest {
  string file_id = 1;
}
```

### DeleteFile - 删除文件
```protobuf
rpc DeleteFile(DeleteFileRequest) returns (DeleteFileResponse);

message DeleteFileRequest {
  string file_id = 1;
}
```

### ListFiles - 获取文件列表
```protobuf
rpc ListFiles(ListFilesRequest) returns (ListFilesResponse);

message ListFilesRequest {
  string user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}
```

## 部署说明

### Docker Compose部署
服务已集成到主项目的docker-compose.yml中：

```yaml
media-service:
  build:
    context: ../../
    dockerfile: deploy/docker/Dockerfile
  container_name: media-service
  environment:
    - MEDIA_SERVICE_PORT=50054
    - CONSUL_AGENT_ADDR=consul:8500
    - MINIO_ENDPOINT=minio:9000
    - MINIO_ACCESS_KEY=minioadmin
    - MINIO_SECRET_KEY=minioadmin123
    - MINIO_USE_SSL=false
    - MINIO_BUCKET_NAME=flick-media
  depends_on:
    - consul
    - minio
  healthcheck:
    test: ["CMD", "grpc_health_probe", "-addr=:50054", "-service=media-service"]
    interval: 10s
    timeout: 5s
    retries: 5
    start_period: 60s
  networks:
    - flick_network
  entrypoint: ["/usr/local/bin/wait-for-it.sh", "minio:9000", "-t", "0", "--", "/app/media"]
  restart: always
```

### 服务启动流程
1. 加载配置文件和环境变量
2. 初始化MinIO客户端并确保bucket存在
3. 连接Consul注册中心（可选）
4. 初始化仓储层和服务层
5. 启动gRPC服务器
6. 注册到Consul并设置健康检查
7. 设置优雅关闭处理

## 项目结构

```
media/
├── main.go                     # 服务入口
├── go.mod                      # Go模块定义
├── proto/                      # Protocol Buffers定义
│   ├── media.proto
│   ├── media.pb.go
│   └── media_grpc.pb.go
└── internal/                   # 内部实现
    ├── consul/                 # Consul客户端
    │   └── consul_client.go
    ├── repository/             # 仓储层
    │   ├── interfaces.go
    │   ├── media_repository.go
    │   ├── minio_storage_repository.go
    │   └── factory.go
    ├── service/                # 服务层
    │   ├── interfaces.go
    │   ├── media_service.go
    │   └── factory.go
    ├── server/                 # 服务器层
    │   ├── interfaces.go
    │   ├── grpc_server.go
    │   └── factory.go
    ├── storage/                # 存储层
    │   ├── interfaces.go
    │   └── minio.go
    └── validator/              # 验证器
        └── media_validator.go
```

## 安全特性

- **文件类型验证**: 严格验证文件MIME类型和扩展名
- **文件大小限制**: 按分类设置不同的文件大小限制
- **预签名URL**: 使用预签名URL提供安全的文件访问
- **内容类型一致性**: 验证文件扩展名与内容类型的一致性

## 监控和日志

- **健康检查**: 支持gRPC健康检查协议
- **结构化日志**: 使用标准日志格式
- **Consul集成**: 自动服务注册和发现
- **优雅关闭**: 支持优雅关闭和资源清理

## 开发说明

### 重新生成Proto文件
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/media.proto
```

### 本地开发
```bash
# 启动依赖服务
docker-compose up -d minio consul

# 设置环境变量
export MEDIA_SERVICE_PORT=50054
export MINIO_ENDPOINT=localhost:9000
export CONSUL_AGENT_ADDR=localhost:8500

# 运行服务
go run main.go
```

## 后续优化

- [ ] 添加单元测试和集成测试
- [ ] 实现图片缩略图生成
- [ ] 添加视频转码功能
- [ ] 实现CDN集成
- [ ] 添加文件去重功能
- [ ] 实现批量上传接口
