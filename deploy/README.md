# Flick 高可用部署架构

本目录包含 Flick 后端服务的高可用部署架构配置，涵盖容器化部署、Kubernetes编排、多区域部署、数据库高可用和缓存高可用等方面。

## 架构概览

Flick 采用微服务架构，通过以下技术实现高可用：

- **API网关层**: 实现负载均衡、安全控制和服务路由
- **服务层**: 各微服务均可水平扩展和自愈
- **数据层**: 实现数据存储的高可用和容灾
- **监控层**: 全方位监控系统健康状态

## 目录结构

```
deploy/
├── docker/           # Docker镜像定义和本地开发环境
├── kubernetes/       # Kubernetes部署配置
│   └── multi-region/ # 多区域部署配置
├── database/         # 数据库高可用配置
└── redis/            # Redis高可用配置
```

## 功能模块

### 1. 容器化部署

- **Docker镜像构建优化**: 使用多阶段构建减小镜像体积
- **容器安全最佳实践**: 非root用户运行、最小化基础镜像
- **容器监控配置**: 内置健康检查和Prometheus指标暴露

相关文件:
- `docker/Dockerfile` - 主要服务的Dockerfile
- `docker/docker-compose.yml` - 本地开发环境配置

### 2. Kubernetes编排

- **资源请求与限制**: 为每个服务设置资源上下限
- **水平自动伸缩**: 基于CPU/内存使用率自动扩缩容
- **健康检查与自动恢复**: 存活检查、就绪检查和启动检查
- **滚动更新策略**: 零停机时间部署

相关文件:
- `kubernetes/namespace.yaml` - 命名空间定义
- `kubernetes/gateway-deployment.yaml` - 网关服务部署配置
- `kubernetes/hpa.yaml` - 水平自动伸缩配置

### 3. 多区域部署

- **区域故障隔离**: 将服务部署在多个地理区域
- **数据同步策略**: 跨区域数据复制和同步
- **跨区域负载均衡**: 全球流量分发
- **就近访问路由**: 用户请求路由到最近的区域

相关文件:
- `kubernetes/multi-region/global-lb.yaml` - 全球负载均衡配置
- `kubernetes/multi-region/gateway-us-east.yaml` - US-East区域部署示例

### 4. 数据库高可用

- **PostgreSQL主从复制**: 自动数据复制
- **自动故障转移**: 使用Patroni实现自动主从切换
- **数据备份与恢复**: 定时备份和灾难恢复流程
- **读写分离配置**: 主节点写入，从节点读取

相关文件:
- `database/postgres-ha.yaml` - PostgreSQL高可用配置
- `database/backup-cronjob.yaml` - 自动备份任务

### 5. 缓存高可用

- **Redis多主集群配置**: 3主3从的Redis集群
- **哨兵模式部署**: 实现自动故障检测和转移
- **缓存雪崩防护**: 过期时间随机化和热点数据永不过期策略
- **永久存储保障**: RDB和AOF双重持久化

相关文件:
- `redis/redis-cluster.yaml` - Redis集群配置
- `redis/redis.conf` - Redis缓存雪崩防护配置

## 部署指南

### 前置条件

- Docker 20.10+
- Kubernetes 1.22+
- Helm 3.8+

### 本地开发环境启动

```bash
cd deploy/docker
docker-compose up -d
```

### Kubernetes部署

1. 创建命名空间和配置:

```bash
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/config.yaml
```

2. 部署数据库和缓存:

```bash
kubectl apply -f database/postgres-ha.yaml
kubectl apply -f redis/redis-cluster.yaml
```

3. 部署服务:

```bash
kubectl apply -f kubernetes/gateway-deployment.yaml
kubectl apply -f kubernetes/gateway-service.yaml
```

4. 配置HPA和Ingress:

```bash
kubectl apply -f kubernetes/hpa.yaml
kubectl apply -f kubernetes/ingress.yaml
```

### 多区域部署

在每个区域的Kubernetes集群上执行:

```bash
kubectl apply -f kubernetes/multi-region/gateway-us-east.yaml  # 替换为对应区域配置
```

然后配置全球负载均衡:

```bash
kubectl apply -f kubernetes/multi-region/global-lb.yaml
```

## 监控和维护

- 监控面板: http://grafana.flick.example.com
- 日志查看: `kubectl logs -f -n flick deployment/gateway`
- 应用健康检查: http://api.flick.example.com/health

## 故障恢复流程

1. **数据库故障转移**: 自动通过Patroni执行
2. **手动恢复备份**:

```bash
kubectl create job --from=cronjob/postgres-restore postgres-restore-manual -n flick
```

3. **缓存重建**:

```bash
kubectl exec -it -n flick deployment/redis-cluster-init -- /bin/sh
``` 