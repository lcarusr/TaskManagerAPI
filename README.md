# Task Manager API

[![CI](https://github.com/lcarusr/TaskManagerAPI/actions/workflows/ci.yml/badge.svg)](https://github.com/lcarusr/TaskManagerAPI/actions/workflows/ci.yml)

基于 **Go + Kratos 微服务框架** 的任务管理 RESTful API，完整 DevOps 工程化交付：容器化 + 云服务器 Kubernetes 部署 + GitHub Actions CI/CD 自动化流水线 + 域名访问。

## 作业交付对照表

| 作业要求 | 实现 | 证据 |
|---------|------|------|
| RESTful CRUD API（/tasks 接口） | GET/POST/PUT/DELETE + 健康检查，标准 RESTful 路径与状态码 | [API](#api) + [交付截图](#交付截图) |
| 内存存储（可扩展） | Repository 接口抽象，内存实现可切换 PostgreSQL/Redis | `internal/data/` |
| 单元测试覆盖率 > 60% | 业务四包实测 77.8%（biz 96.3% / data 88.9% / service 72.9% / server 64.7%） | CI Test job |
| 容器化 Docker | 多阶段构建，非 root，镜像 ~15-40MB | [Docker 构建](#docker-构建与运行) |
| Kubernetes 部署 | 云服务器集群：namespace/Deployment(2副本)/Service/ConfigMap/Ingress | [k8s/README.md](k8s/README.md) |
| 域名访问 | Ingress + ingress-nginx，`task-manager.local` 实测 200 | [域名访问](#kubernetes-部署云服务器集群) |
| CI/CD 流水线 | GitHub Actions：lint→test→build→Trivy 安全扫描→推镜像→自动部署 | [CI/CD](#cicd) |
| 安全漏洞扫描 | Trivy（HIGH/CRITICAL 门禁），依赖漏洞已修复为 0 | CI Security Scan job |
| 限流（加分项） | 令牌桶 100 QPS 中间件，超限 429，含单元测试 | `internal/server/ratelimit.go` |

## 功能特性

- RESTful CRUD：任务增删改查 + 健康检查（`GET /health`），标准状态码 201/204/400/404
- 内存存储，通过 Repository 接口抽象，可无缝切换 PostgreSQL / Redis
- **令牌桶限流（100 QPS，超限 429）**、结构化日志、统一错误码、输入校验
- 一份 proto 同时生成 HTTP + gRPC + OpenAPI/Swagger 文档
- 单元测试（覆盖率 > 60%）+ 并发安全（`-race`）CI 门禁
- Docker 多阶段构建（镜像 < 150MB，非 root 用户运行）
- Kubernetes 部署 + Ingress **域名访问**（namespace / Deployment / Service / ConfigMap / Ingress）

## API

| 方法 | 路径 | 功能 | 成功 | 失败 |
|------|------|------|------|------|
| GET | /health | 健康检查 | 200 | - |
| GET | /tasks | 获取所有任务 | 200 | - |
| GET | /tasks/{id} | 按 ID 获取任务 | 200 | 404 |
| POST | /tasks | 创建任务 | 201 | 400 |
| PUT | /tasks/{id} | 更新任务（部分更新） | 200 | 404 / 400 |
| DELETE | /tasks/{id} | 删除任务 | 204 | 404 |

### 数据模型

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "任务标题",
  "description": "任务描述",
  "status": "todo | in_progress | done",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### 示例

```bash
# 创建任务
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"写报告","description":"Q3 总结","status":"todo"}'

# 更新任务（部分更新：仅覆盖请求中出现的字段）
curl -X PUT http://localhost:8080/tasks/<id> \
  -H "Content-Type: application/json" \
  -d '{"status":"done"}'

# 查询 / 删除
curl http://localhost:8080/tasks
curl http://localhost:8080/tasks/<id>
curl -X DELETE http://localhost:8080/tasks/<id>   # 204
```

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.25+（稳定版本，不追最新） |
| 微服务底座 | Kratos v3.0 |
| API 定义 | proto3 + google.api.http（buf 生成） |
| 存储 | 内存（sync.RWMutex + map，Repository 接口可插拔） |
| 日志 / 限流 | Kratos slog / 令牌桶中间件（100 QPS） |
| 容器 | Docker 多阶段构建 |
| 部署 | 云服务器 Kubernetes 集群 + ingress-nginx |
| CI/CD | GitHub Actions + golangci-lint + Trivy |

## 目录结构

```
task-manager-api/
├── .github/workflows/ci.yml   # CI/CD 流水线
├── api/                       # proto 接口定义
├── cmd/task-manager-api/      # 程序入口
├── configs/                   # 配置文件
├── docs/screenshots/          # 交付截图
├── internal/
│   ├── biz/                   # 业务逻辑层
│   ├── data/                  # 数据访问层
│   ├── server/                # HTTP/gRPC Server（含限流中间件）
│   ├── service/               # 服务层
│   └── conf/                  # 配置解析
├── k8s/                       # Kubernetes 资源清单
├── buf.yaml / buf.gen.yaml    # proto 代码生成（buf）
├── Dockerfile
├── Makefile
└── CONTRIBUTING.md            # Git 工作流规范
```

## 本地开发

### 前置要求

- Go 1.25+（`go version` 验证）
- buf（可选，修改 proto 后重新生成：`go install github.com/bufbuild/buf/cmd/buf@latest`）
- Docker（可选，本地容器运行）
- kubectl 可访问云服务器集群（可选，部署验证）

### 运行

```bash
# 1. 安装依赖
go mod tidy

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
# 部署全部资源（先 namespace，其余资源依赖它）
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/

# 验证
kubectl get all -n task-manager
kubectl rollout status deployment/task-service -n task-manager

# 域名访问（本机 hosts 添加：<云服务器公网IP> task-manager.local）
curl http://task-manager.local:31673/health
curl http://task-manager.local:31673/tasks
```

> 端口链路：`外部 :31673(NodePort) → ingress-nginx :80 → Ingress 按域名路由 → task-service :8080`。
> 完整部署步骤（含 Ingress Controller 安装、国内镜像源、安全组放行）见 [k8s/README.md](k8s/README.md)。

## 交付截图

| 截图 | 内容 | 文件 |
|------|------|------|
| CI 流水线 | GitHub Actions 全绿（Lint/Test/Build/Security Scan） | `docs/screenshots/ci-pipeline.png` |
| K8s 部署 | `kubectl get all -n task-manager`（2 副本 Running） | `docs/screenshots/k8s-deployment.png` |
| 域名访问 | `curl http://task-manager.local:31673/health` 返回 200 | `docs/screenshots/domain-access.png` |

将截图放入 `docs/screenshots/` 后即可在交付时展示。

## CI/CD

`GitHub Actions`，`.github/workflows/ci.yml`。**PR 验证 + main 发布**双阶段：

| Job | 触发（PR） | 触发（push main） | 说明 |
|-----|-----------|-------------------|------|
| Lint | ✅ | ✅ | golangci-lint（go1.26 现场编译，规避版本不匹配） |
| Test | ✅ | ✅ | `go test -race -cover`，覆盖率 ≥ 60% 门禁 |
| Build | ✅ | ✅ | Go 编译 + Docker 镜像构建 |
| Security Scan | ✅ | ✅ | Trivy，HIGH/CRITICAL 漏洞 exit 1（部署前强制扫描） |
| Push image | 跳过 | ✅ | 推 `ghcr.io/lcarusr/task-manager-api:<sha>` + `latest` |
| Deploy to cluster | 跳过 | ✅ | kubectl 滚动更新（需 `KUBE_CONFIG` secret，未配置自动跳过） |

```
Lint → Test → Build → Security Scan(Trivy) → Push image → Deploy to cluster
```

- PR 阶段只做质量门禁（不推镜像、不部署）
- 合并到 main 后自动发布：推镜像 + 滚动部署到云集群
- 失败即阻断（`needs` 依赖），扫描不过不发布

## Git 工作流

Trunk-based + feature 分支，`main` 只通过 PR 合并，Conventional Commits 提交规范，详见 [CONTRIBUTING.md](CONTRIBUTING.md)。
