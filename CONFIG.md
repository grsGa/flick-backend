# Backend Configuration Management

## 概述

本项目采用标准化的环境变量配置管理方式，使用 `.env` 文件进行配置管理。

## 配置文件说明

### `.env.example`
- 配置模板文件，包含所有可配置的环境变量
- 提交到版本控制系统，供团队成员参考
- 不包含敏感信息，使用占位符

### `.env` (本地开发)
- 实际的环境变量配置文件
- **不提交到版本控制系统** (已在 .gitignore 中忽略)
- 包含真实的配置值和敏感信息

### `.env.development`
- 开发环境的默认配置
- 可以提交到版本控制系统
- 包含开发环境的默认值

## 使用方式

### 1. 初始化配置
```bash
# 复制模板文件
cp .env.example .env

# 编辑配置文件，填入实际值
vim .env
```

### 2. 配置优先级
配置加载优先级（从高到低）：
1. 系统环境变量
2. `.env` 文件
3. `.env.development` 文件

### 3. 在代码中使用
```go
import "github.com/flick/backend/pkg/config"

// 加载配置
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// 使用配置
fmt.Printf("Gateway Port: %s\n", cfg.GatewayPort)
```

## 配置项说明

### 服务端口配置
- `GATEWAY_PORT`: API Gateway 端口
- `AUTH_SERVICE_PORT`: 认证服务端口
- `USER_SERVICE_PORT`: 用户服务端口
- `CONTENT_SERVICE_PORT`: 内容服务端口
- `MEDIA_SERVICE_PORT`: 媒体服务端口
- `MESSAGES_SERVICE_PORT`: 消息服务端口
- `NOTIFICATION_SERVICE_PORT`: 通知服务端口
- `INTERACTION_SERVICE_PORT`: 互动服务端口
- `RECOMMENDATION_SERVICE_PORT`: 推荐服务端口
- `SEARCH_SERVICE_PORT`: 搜索服务端口

### 数据库配置
- `POSTGRES_HOST`: PostgreSQL 主机地址
- `POSTGRES_PORT`: PostgreSQL 端口
- `POSTGRES_USER`: 数据库用户名
- `POSTGRES_PASSWORD`: 数据库密码
- `POSTGRES_DB`: 数据库名称

### OAuth 配置
- `GITHUB_CLIENT_ID`: GitHub OAuth 客户端 ID
- `GITHUB_CLIENT_SECRET`: GitHub OAuth 客户端密钥
- `GITHUB_REDIRECT_URL`: GitHub OAuth 回调地址
- `GOOGLE_CLIENT_ID`: Google OAuth 客户端 ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth 客户端密钥
- `GOOGLE_REDIRECT_URL`: Google OAuth 回调地址

### MinIO 配置
- `MINIO_ENDPOINT`: MinIO 服务端点
- `MINIO_ACCESS_KEY`: MinIO 访问密钥
- `MINIO_SECRET_KEY`: MinIO 秘密密钥
- `MINIO_USE_SSL`: 是否使用 SSL
- `MINIO_BUCKET_NAME`: 存储桶名称

### 其他配置
- `CONSUL_AGENT_ADDR`: Consul 代理地址
- `JWT_SECRET`: JWT 签名密钥
- `CUSTOM_CA_CERT_PATH`: 自定义 CA 证书路径

## 安全注意事项

1. **永远不要提交 `.env` 文件到版本控制系统**
2. **敏感信息（如密钥、密码）只能存在于 `.env` 文件中**
3. **生产环境建议使用环境变量而不是 `.env` 文件**
4. **定期更换敏感配置信息**

## 部署环境配置

### Docker 部署
在 docker-compose.yml 中引用环境变量：
```yaml
services:
  gateway:
    environment:
      - GATEWAY_PORT=${GATEWAY_PORT}
      - JWT_SECRET=${JWT_SECRET}
```

### Kubernetes 部署
使用 ConfigMap 和 Secret：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: flick-config
data:
  GATEWAY_PORT: "8080"
  POSTGRES_HOST: "postgres"
---
apiVersion: v1
kind: Secret
metadata:
  name: flick-secrets
data:
  JWT_SECRET: <base64-encoded-secret>
```
