# Task Manager API — Kubernetes 部署指南

部署到云服务器 K8s 集群（本地 kubectl 已可访问）。

## 资源清单

| 文件 | 资源 | 说明 |
|------|------|------|
| namespace.yaml | Namespace | 独立命名空间 `task-manager` |
| configmap.yaml | ConfigMap | 非敏感配置（日志级别、HTTP 地址） |
| deployment.yaml | Deployment | 2 副本；CPU 100m-200m / Memory 128Mi-256Mi；Liveness + Readiness 探针（/health）；imagePullSecrets 引用 ghcr-secret |
| service.yaml | Service | ClusterIP，端口 8080 |
| ingress.yaml | Ingress | 域名 `task-manager.local`，`ingressClassName: nginx` |

## 部署步骤

```powershell
# 0) 确认本地 kubectl 可访问云服务器集群
kubectl cluster-info
kubectl get nodes

# 1) 安装 Ingress Controller（ingress-nginx，云服务器用 baremetal + NodePort 方式）
#    注意：国内服务器直连 registry.k8s.io 拉镜像会超时，
#    需把 yaml 中的镜像前缀替换为 DaoCloud 加速源 m.daocloud.io
curl.exe -o ingress-nginx.yaml `
  https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.0/deploy/static/provider/baremetal/deploy.yaml
(Get-Content ingress-nginx.yaml) -replace 'registry\.k8s\.io','m.daocloud.io/registry.k8s.io' | Set-Content ingress-nginx-cn.yaml
kubectl apply -f ingress-nginx-cn.yaml
kubectl -n ingress-nginx rollout status deployment/ingress-nginx-controller

#    查入口端口（HTTP 80 映射的 NodePort，示例 31673）
kubectl get svc -n ingress-nginx

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

# 5) 云服务器安全组放行 NodePort 端口（如 31673）
#    腾讯云控制台：云服务器 → 安全组 → 入站规则 → 放行 TCP 31673

# 6) 本机 hosts 添加（管理员权限编辑 C:\Windows\System32\drivers\etc\hosts）
#    <云服务器公网IP> task-manager.local
#    例：124.222.10.156 task-manager.local
```

## 域名访问（已实测验证）

```powershell
# NodePort 方式：域名 + 端口访问（hosts 已配好后，无需 -H 头）
curl http://task-manager.local:31673/health      # 期望 {"status":"ok"}
curl http://task-manager.local:31673/tasks       # 期望 200 JSON
```

## 验证命令汇总

```bash
# 健康检查
curl http://task-manager.local:31673/health      # 期望 {"status":"ok"}

# CRUD
curl http://task-manager.local:31673/tasks       # 期望 200 JSON
curl -X POST http://task-manager.local:31673/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"k8s task"}'                      # 期望 201

# 查看日志
kubectl logs -n task-manager -l app=task-manager
```

## 说明

- 镜像默认 `ghcr.io/lcarusr/task-manager-api:latest`（**全小写**，Docker 镜像名规则）；CI 的 deploy 阶段会用 `<sha>` 标签滚动更新
- **Ingress 必须指定 `ingressClassName: nginx`**，否则 ingress-nginx 不处理该 Ingress，外部访问返回 404
- 无真实域名时用 hosts 指向公网 IP；集群若有真实 DNS，可替换 ingress 的 host
- 想用标准 80 端口（不带 :31673）访问：可将 ingress-nginx controller 改为 `hostNetwork: true`（controller 直接占用节点 80），或服务器防火墙把 80 转发到 NodePort
- 端口链路：`外部 :31673(NodePort) → ingress-nginx :80 → Ingress 按域名路由 → task-service ClusterIP :8080 → Pod :8080`
