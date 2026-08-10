# Git 工作流规范

本仓库采用 **Trunk-based + feature 分支** 工作流，配合 Conventional Commits 提交规范。

## 分支模型

| 分支 | 用途 | 保护 |
|------|------|------|
| `main` | 主分支，始终保持稳定可运行 | 建议开启分支保护（PR 合并 + CI 通过） |
| `feature/*` | 功能开发分支，如 `feature/task-api`、`feature/ci-cd` | 开发完成后合并回 `main` |

### 开发流程

```bash
# 1. 从最新的 main 拉出功能分支
git checkout main && git pull
git checkout -b feature/task-api

# 2. 在功能分支上开发，小步提交（Conventional Commits）
git add . && git commit -m "feat: add task CRUD usecase"

# 3. 合并回 main（建议走 Pull Request，CI 通过后合并）
git checkout main && git merge --no-ff feature/task-api
```

## 提交规范（Conventional Commits）

格式：`<type>(<scope>): <subject>`

| type | 用途 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复缺陷 |
| `docs` | 文档变更 |
| `test` | 测试相关 |
| `ci` | CI/CD 配置变更 |
| `chore` | 构建/工具/杂项 |
| `refactor` | 重构（不改变行为） |
| `style` | 格式调整（不影响逻辑） |
| `perf` | 性能优化 |

### 示例

```
feat: add task CRUD API
test: add unit tests for task usecase
ci: add github actions pipeline
docs: add deployment guide
feat: complete homework submission
```

### 建议

- 每个 commit 只做一件事，保持原子性
- subject 用祈使句，简洁描述"做了什么"
- 需要更多上下文时使用 body 说明原因

## 初始提交历史（演示演进过程）

```
docs: add project readme and git workflow        ← 文档基线
chore: add gitignore for go project               ← 工程配置
feat: bootstrap kratos project skeleton           ← 框架初始化
ci: add github actions pipeline                   ← 流水线
```
