# Flick 后端服务

这是Flick应用的后端微服务架构，由多个独立的服务组成，每个服务负责不同的业务域。

## 架构概述

Flick采用微服务架构，将不同的业务功能拆分为独立的服务。每个服务都有自己的数据存储、业务逻辑和API端点。服务之间通过HTTP API和消息队列进行通信。

### 技术栈

- **语言**: Go
- **Web框架**: Gin + Gorilla Mux
- **ORM**: GORM
- **数据库**: PostgreSQL
- **缓存**: Redis
- **消息队列**: Kafka
- **日志**: Zerolog
- **监控**: Prometheus + grafana
- **链路追踪**: Jaeger

## 核心服务

### 1. 用户服务 (User Service)

用户服务负责用户账户管理、身份验证和授权。

**主要功能**:
- 用户注册和登录
- 用户资料管理
- 身份验证(JWT)
- 权限控制

**API端点**:
- `GET /api/v1/users` - 获取用户列表
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users/{id}` - 获取用户详情
- `PUT /api/v1/users/{id}` - 更新用户
- `DELETE /api/v1/users/{id}` - 删除用户
- `GET /api/v1/users/me` - 获取当前用户
- `PUT /api/v1/users/me` - 更新当前用户
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/logout` - 登出
- `POST /api/v1/auth/refresh` - 刷新令牌

### 2. 内容服务 (Content Service)

内容服务负责管理用户创建的各种内容，如帖子、文章等。

**主要功能**:
- 内容创建、读取、更新、删除(CRUD)
- 内容分类和标签管理
- 内容搜索
- 内容审核

**API端点**:
- `GET /api/v1/content` - 获取内容列表
- `POST /api/v1/content` - 创建内容
- `GET /api/v1/content/{id}` - 获取内容详情
- `PUT /api/v1/content/{id}` - 更新内容
- `DELETE /api/v1/content/{id}` - 删除内容
- `GET /api/v1/content/categories` - 获取分类列表
- `GET /api/v1/content/tags` - 获取标签列表
- `GET /api/v1/content/search` - 搜索内容
- `GET /api/v1/content/trending` - 获取热门内容
- `POST /api/v1/content/upload` - 上传内容

### 3. 交互服务 (Interaction Service)

交互服务处理用户与内容的交互，如点赞、评论等。

**主要功能**:
- 点赞和评论管理
- 用户关注
- 内容分享
- 用户活动追踪

**API端点**:
- `POST /api/v1/interactions` - 创建交互
- `GET /api/v1/interactions` - 获取交互列表
- `GET /api/v1/interactions/{id}` - 获取交互详情
- `PUT /api/v1/interactions/{id}` - 更新交互
- `DELETE /api/v1/interactions/{id}` - 删除交互

### 4. 推荐服务 (Recommendation Service)

推荐服务为用户提供个性化的内容推荐。

**主要功能**:
- 基于用户行为的推荐
- 热门内容推荐
- A/B测试
- 推荐模型管理

**API端点**:
- `GET /api/v1/recommendations` - 获取推荐
- `POST /api/v1/recommendations/feedback` - 记录反馈
- `POST /api/v1/recommendations/viewed` - 标记为已查看
- `POST /api/v1/recommendations/clicked` - 标记为已点击
- `GET /api/v1/recommendation-models` - 获取推荐模型列表
- `POST /api/v1/recommendation-models` - 创建推荐模型
- `PUT /api/v1/recommendation-models/{id}` - 更新推荐模型
- `DELETE /api/v1/recommendation-models/{id}` - 删除推荐模型
- `GET /api/v1/recommendation-abtests` - 获取A/B测试列表
- `POST /api/v1/recommendation-abtests` - 创建A/B测试
- `GET /api/v1/recommendation-abtests/{id}/metrics` - 获取A/B测试指标
- `PUT /api/v1/recommendation-abtests/{id}` - 更新A/B测试
- `DELETE /api/v1/recommendation-abtests/{id}` - 删除A/B测试

### 5. 通知服务 (Notification Service)

通知服务处理各种用户通知，如新关注、新评论等。

**主要功能**:
- 通知生成和发送
- 通知管理(标记已读等)
- 通知设置
- 推送通知

**API端点**:
- `GET /api/v1/notifications` - 获取通知列表
- `GET /api/v1/notifications/{id}` - 获取通知详情
- `PUT /api/v1/notifications/{id}/read` - 标记通知为已读
- `GET /api/v1/notifications/settings` - 获取通知设置
- `PUT /api/v1/notifications/settings` - 更新通知设置

### 6. 网关服务 (Gateway Service)

API网关服务作为所有客户端应用的入口点，负责请求路由、认证和限流。

**主要功能**:
- 请求路由
- 认证和授权
- 速率限制
- 请求/响应转换
- API文档

## 项目结构

后端项目使用以下目录结构:

```
backend/
├── deploy/            # 部署相关配置和脚本
├── pkg/               # 共享包
│   ├── config/        # 配置管理
│   ├── database/      # 数据库连接和管理
│   ├── models/        # 共享数据模型
│   └── middleware/    # 共享中间件
├── services/          # 微服务
│   ├── user/          # 用户服务
│   ├── content/       # 内容服务
│   ├── interaction/   # 交互服务
│   ├── recommendation/# 推荐服务
│   ├── notification/  # 通知服务
│   └── gateway/       # API网关
└── vendor/            # 依赖包
```

每个服务目录都包含以下结构:

```
service/
├── main.go           # 服务入口点
├── repository/       # 数据访问层
├── service/          # 业务逻辑层
└── models/           # 服务特定模型（如果有）
```

## 开发指南

### 设置开发环境

1. 安装Go 1.21或更高版本
2. 安装Docker和Docker Compose
3. 克隆仓库
4. 执行以下命令启动依赖服务:

```bash
docker-compose up -d postgres redis kafka
```

### 构建和运行服务

构建和运行所有服务:

```bash
go build -o services/user/user.exe services/user/main.go
go build -o services/content/content.exe services/content/main.go
go build -o services/interaction/interaction.exe services/interaction/main.go
go build -o services/recommendation/recommendation.exe services/recommendation/main.go
go build -o services/notification/notification.exe services/notification/main.go
go build -o services/gateway/gateway.exe services/gateway/main.go
```

也可以使用提供的脚本一次构建所有服务:

```bash
./build.sh
```

## 部署

服务可以作为Docker容器部署到Kubernetes集群或任何支持容器的平台。

### 容器构建

每个服务都有自己的Dockerfile，可以使用以下命令构建:

```bash
docker build -t flick/user-service -f services/user/Dockerfile .
docker build -t flick/content-service -f services/content/Dockerfile .
docker build -t flick/interaction-service -f services/interaction/Dockerfile .
docker build -t flick/recommendation-service -f services/recommendation/Dockerfile .
docker build -t flick/notification-service -f services/notification/Dockerfile .
docker build -t flick/gateway-service -f services/gateway/Dockerfile .
```

## 监控和日志

- 每个服务都暴露Prometheus指标在`/metrics`端点
- 健康检查可通过每个服务的`/health`端点进行
- 所有服务使用结构化日志记录，便于日志聚合和分析

## 服务通信

服务间通信主要通过以下方式:

1. HTTP API调用 - 用于同步请求
2. Kafka消息 - 用于异步事件和数据变更通知

## 贡献指南

如果你想为项目做贡献，请遵循以下步骤:

1. Fork仓库
2. 创建你的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交你的变更 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开一个Pull Request

## 中间件

系统提供了多种中间件用于处理常见的横切关注点：

### 认证中间件

- **AuthMiddleware**: 基本认证中间件，验证JWT令牌并将用户信息添加到上下文。对于没有认证头的请求，会作为游客继续处理。
- **JWTAuthMiddleware**: 严格认证中间件，要求请求必须携带有效的JWT令牌，否则返回401未授权错误。
- **RequireAuth**: 要求请求必须经过认证的中间件，确保只有登录用户才能访问。
- **RequireRoles**: 检查用户是否具有特定角色的中间件。
- **RequirePermission**: 检查用户是否具有特定权限的中间件。

### 其他中间件

- **LoggerMiddleware**: 记录请求日志，包括请求路径、方法、状态码和处理时间。
- **MetricsMiddleware**: 收集请求指标，如请求数量和处理时间。
- **CORSMiddleware**: 处理跨域资源共享，允许前端应用访问API。
- **RecoveryMiddleware**: 从panic中恢复，防止服务因未处理的错误而崩溃。
- **ErrorHandlerMiddleware**: 统一处理和格式化错误响应。 