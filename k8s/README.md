# Task Manager API — Kubernetes 部署指南

部署到云服务器 K8s 集群（本地 kubectl 已可访问）。

## 资源清单

| 文件 | 资源 | 说明 |
|------|------|------|
| namespace.yaml | Namespace | 独立命名空间 `task-manager` |
| configmap.yaml | ConfigMap | 非敏感配置（日志级别、HTTP 地址） |
| deployment.yaml | Deployment | 2 副本；CPU 100m-200m / Memory 128Mi-256Mi；Liveness + Readiness 探针（/health） |
| service.yaml | Service | ClusterIP，端口 8080 |
| ingress.yaml | Ingress | 域名 `task-manager.local` |

## 部署步骤

```powershell
# 0) 确认本地 kubectl 可访问云服务器集群
kubectl cluster-info
kubectl get nodes

# 1) 确认集群已安装 Ingress Controller（没有则先装 ingress-nginx）
kubectl get pods -n ingress-nginx

# 2) 一次性创建 ghcr.io 镜像拉取凭证（私有镜像必需；公开镜像可跳过）
#    需要 GitHub Personal Access Token（read:packages 权限）
kubectl create secret docker-registry ghcr-secret `
  --docker-server=ghcr.io `
  --docker-username=lcarusr `
  --docker-password=<PAT> `
  -n task-manager

# 3) 部署全部资源（先 apply namespace，其余资源依赖它；
#    kubectl apply -f k8s/ 按字母序，namespace 排在最后会报 not found）
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/

# 4) 验证
kubectl get all -n task-manager
kubectl rollout status deployment/task-service -n task-manager
kubectl get pods -n task-manager        # 期望 2 个 Pod Running/Ready

# 5) 访问（本地 hosts 添加：<云服务器公网IP> task-manager.local）
curl http://task-manager.local/health
curl http://task-manager.local/tasks
```

## 验证命令汇总

```bash
# 健康检查
curl http://task-manager.local/health      # 期望 {"status":"ok"}

# CRUD
curl http://task-manager.local/tasks       # 期望 200 JSON
curl -X POST http://task-manager.local/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"k8s task"}'                # 期望 201

# 查看日志
kubectl logs -n task-manager -l app=task-manager
```

## 说明

- 镜像默认 `ghcr.io/lcarusr/TaskManagerAPI:latest`；CI 的 deploy 阶段会用 `<sha>` 标签滚动更新
- 无真实域名时用 hosts 指向公网 IP；集群若有真实 DNS，可替换 ingress 的 host
- 公网快速验证可临时将 Service 改为 NodePort/LoadBalancer 类型
