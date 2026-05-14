# Go 项目 Git 工作流规范

## 分支策略

| 分支 | 用途 | 命名规范 |
|------|------|----------|
| `main` | 生产分支，始终可部署 | 固定 |
| `develop` | 开发集成分支 | 固定 |
| `feature/*` | 功能开发 | `feature/{项目名}-{功能简述}` |
| `fix/*` | Bug 修复 | `fix/{项目名}-{bug简述}` |
| `release/*` | 发布准备 | `release/{版本号}` |
| `hotfix/*` | 生产紧急修复 | `hotfix/{问题简述}` |

**示例：**
- `feature/game-server-tenant-management`
- `fix/game-server-login-401`
- `release/1.0.0`

## Commit 规范

使用 Conventional Commits 格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(game): 新增游戏管理 CRUD 接口` |
| `fix` | Bug 修复 | `fix(auth): 修复登录时 code 字段必填校验` |
| `refactor` | 重构（不改变功能） | `refactor(middleware): 提取租户中间件` |
| `docs` | 文档变更 | `docs: 更新 API 文档` |
| `style` | 代码格式（不影响逻辑） | `style: gofmt 格式化` |
| `test` | 测试相关 | `test(game): 新增游戏服务单元测试` |
| `chore` | 构建/工具变更 | `chore: 更新 build.ps1 脚本` |
| `perf` | 性能优化 | `perf(query): 优化游戏列表查询索引` |

### Scope 范围

使用模块名作为 scope：
- `auth` — 认证授权
- `game` — 游戏管理
- `tenant` — 租户管理
- `admin` — 系统管理
- `setup` — 安装向导
- `frontend` — 前端
- `build` — 构建部署

### 规则

- 第一行不超过 50 个字符
- 使用中文描述（与项目语言规范一致）
- 使用祈使语气（"新增"而非"新增了"）
- Body 用于解释 what 和 why，不解释 how
- Breaking change 在 footer 标注 `BREAKING CHANGE:`

### 示例

```
feat(tenant): 新增租户管理模块

- 创建 Tenant model 和 CRUD 接口
- 业务表统一嵌入 TenantBy mixin
- JWT payload 加入 tenantId 字段
- 游戏查询自动按 tenant_id 过滤
```

```
fix(auth): 修复 /login 路由返回 HTML 的问题

生产模式下 /login 未注册为独立路由，被 NoRoute 的 SPA
处理器拦截返回了 index.html。

在 sysCheckRoleRouterInit 中显式注册 POST /login 路由。
```

## 工作流程

### 功能开发

```
1. 从 develop 创建 feature 分支
   git checkout -b feature/game-server-xxx develop

2. 开发并提交（小步提交，每个逻辑变更一个 commit）
   git add .
   git commit -m "feat(game): 新增游戏列表接口"

3. 开发完成后推送并创建 PR
   git push -u origin feature/game-server-xxx

4. Code Review 通过后合并到 develop
   git checkout develop
   git merge --no-ff feature/game-server-xxx

5. 删除 feature 分支
   git branch -d feature/game-server-xxx
```

### Bug 修复

```
1. 从 develop 创建 fix 分支
   git checkout -b fix/game-server-xxx develop

2. 修复并提交
   git commit -m "fix(auth): 修复登录验证码必填问题"

3. 合并回 develop
```

### 发布

```
1. 从 develop 创建 release 分支
   git checkout -b release/1.0.0 develop

2. 只做 bug 修复和文档更新，不加新功能

3. 合并到 main 并打 tag
   git checkout main
   git merge --no-ff release/1.0.0
   git tag -a v1.0.0 -m "Release 1.0.0"

4. 合并回 develop
   git checkout develop
   git merge --no-ff release/1.0.0
```

## .gitignore 规范

Go 项目必须忽略的文件：

```gitignore
# 构建产物
dist/
*.exe

# 配置文件（含敏感信息）
config/settings.yml

# 前端依赖和构建
frontend/node_modules/
frontend/dist/
backend/web/dist/

# IDE
.idea/
.vscode/
*.swp

# 日志
logs/
temp/

# OS
.DS_Store
Thumbs.db
```

## 安全规则

- **禁止**提交密码、密钥、token 到仓库
- `config/settings.yml` 必须在 `.gitignore` 中
- 提交前检查 `git diff --staged` 确认无敏感信息
- 使用 `config/settings.example.yml` 提供配置模板
