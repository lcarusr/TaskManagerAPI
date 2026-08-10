# Task Manager API

[![CI](https://github.com/<username>/task-manager-api/actions/workflows/ci.yml/badge.svg)](https://github.com/<username>/task-manager-api/actions/workflows/ci.yml)

基于 **Go + Kratos 微服务框架** 的任务管理 RESTful API，完整 DevOps 工程化交付：容器化 + 云服务器 Kubernetes 部署 + GitHub Actions CI/CD 自动化流水线。

## 功能特性

- RESTful CRUD：任务增删改查 + 健康检查（`GET /health`）
- 内存存储，通过 Repository 接口抽象，可无缝切换 PostgreSQL / Redis
- 令牌桶限流、结构化日志、统一错误码、输入校验
- 一份 proto 同时生成 HTTP + gRPC + OpenAPI/Swagger 文档
- 单元测试（覆盖率 > 60%），CI 门禁
- Docker 多阶段构建（镜像 < 150MB，非 root 用户运行）
- Kubernetes 部署（namespace / Deployment / Service / ConfigMap / Ingress）

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.21+（稳定版本，不追最新） |
| 微服务底座 | Kratos v2.8 |
| API 定义 | proto3 + google.api.http |
| 存储 | 内存 sync.Map（可插拔） |
| 日志 / 限流 | Kratos log / ratelimit middleware |
| 容器 | Docker 多阶段构建 |
| 部署 | 云服务器 Kubernetes 集群（kubectl 直接管理） |
| CI/CD | GitHub Actions + golangci-lint + Trivy |

## 目录结构

```
task-manager-api/
├── .github/workflows/ci.yml   # CI/CD 流水线
├── api/                       # proto 接口定义
├── cmd/task-manager-api/      # 程序入口
├── configs/                   # 配置文件
├── internal/
│   ├── biz/                   # 业务逻辑层
│   ├── data/                  # 数据访问层
│   ├── server/                # HTTP/gRPC Server
│   ├── service/               # 服务层
│   └── conf/                  # 配置解析
├── k8s/                       # Kubernetes 资源清单
├── third_party/               # proto 依赖
├── Dockerfile
├── Makefile
└── CONTRIBUTING.md            # Git 工作流规范
```

## 本地开发

### 前置要求

- Go 1.21+（`go version` 验证）
- Docker（可选，本地容器运行）
- kubectl 可访问云服务器集群（可选，部署验证）

### 运行

```bash
# 1. 安装依赖并生成代码
go mod tidy
go generate ./...

# 2. 构建并启动
go build -o ./bin/ ./...
./bin/task-manager-api -conf ./configs

# 3. 验证
curl http://localhost:8080/health
curl http://localhost:8080/tasks
```

### 单元测试

```bash
go test -race -cover ./...
```

## Docker 构建与运行

```bash
# 多阶段构建（最终镜像约 15~40MB）
docker build -t task-manager-api .

# 运行（端口可通过环境变量覆盖，默认 8080）
docker run -p 8080:8080 -e PORT=8080 task-manager-api

curl http://localhost:8080/health
```

## Kubernetes 部署（云服务器集群）

```powershell
# 部署全部资源
kubectl apply -f k8s/

# 验证
kubectl get all -n task-manager
kubectl rollout status deployment/task-service -n task-manager

# 访问（本地 hosts 添加：<云服务器公网IP> task-manager.local）
curl http://task-manager.local/health
curl http://task-manager.local/tasks
```

详细部署步骤与验证命令见 [k8s/README.md](k8s/README.md)。

## CI/CD

`Push main` / `Pull Request` 触发，阶段依赖 `needs`，失败即阻断：

```
Lint(golangci-lint) → Test(go test -race -cover ≥60%) → Build(docker image)
    → Security Scan(Trivy HIGH/CRITICAL) → Push(ghcr.io/<username>/task-manager-api:<sha>)
    → Deploy(kubectl 滚动更新云集群，需配置 KUBE_CONFIG secret)
```

## Git 工作流

Trunk-based + feature 分支，Conventional Commits 提交规范，详见 [CONTRIBUTING.md](CONTRIBUTING.md)。
