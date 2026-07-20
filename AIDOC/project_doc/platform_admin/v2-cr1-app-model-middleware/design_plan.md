# 设计计划：CR-1 应用模型升级 + 中间件链重构

## 设计方向

基于 design_v2.md 已定稿的架构设计，CR-1 从 V1 代码渐进式重构，分 Step A（纯迁移）和 Step B（新能力）两步执行。

### 核心概念关系（本 CR 建立的模型基础）

```
请求进入 → 1.认证 → 2.应用订阅 → 3.模块启用 → 4.功能权限 → Handler

1. 认证：Token 有效？租户状态正常？（AuthMiddleware）
2. 应用订阅：API 属于哪个 app_code？租户订阅了吗？（AppResolveMiddleware — 本 CR 新增）
3. 模块启用：API 的 module_code 在租户的 enabled_modules 中？（AppResolveMiddleware 中检查）
4. 功能权限：用户角色有这个 API 的 permission_code？（DynamicPermissionMiddleware — 已有）
```

**字段用途速查**：
- `app_code` = 标识"这东西属于哪个应用" → 控制租户订阅级可见性
- `platform` = 标识"菜单在哪个前端展示"(admin/user) → 控制前端分离
- `module_code` = 标识"属于应用的哪个功能模块" → 控制模块级开关
- `enabled_modules` = 租户对某应用开通了哪些模块 → NULL=全部

**路由分组**：
- `/api/v1/admin/...` → app_code=platform_admin（管理端应用 API，走 AppResolveMiddleware）
- `/api/v1/{plugin}/...` → app_code={plugin}（插件应用 API，走 AppResolveMiddleware）
- `/api/v1/common/...` → 公共接口，**不走** AppResolveMiddleware（管理端和用户端共用，如 user-menu、user-info）
- `/auth/...` → 认证接口，**不走** AppResolveMiddleware

### Step A：纯迁移

1. **Model 层加字段**：通过 GORM tag 添加新字段，AutoMigrate 自动 ALTER TABLE
2. **路由前缀替换**：后端 router.go 中路由注册从 `/api/v1/` 改为 `/api/v1/admin/`，新增 `/api/v1/common/` 公共组
3. **前端批量替换**：所有 `src/api/*.ts` 文件中的 API 路径替换前缀
4. **Seed 更新**：种子数据填充新字段（app_code、platform、module_code 等）
5. **AutoDiscover 适配**：扫描新前缀路由，注册时写入 app_code=platform_admin

### Step B：新能力

1. **AppResolveMiddleware**：启动时加载 route_prefix → app_code 映射，请求时最长前缀匹配 + 租户订阅校验 + enabled_modules 校验
2. **DynamicPermissionMiddleware 增强**：module_code 运行时检查
3. **AuthMiddleware 增强**：识别租户四态（ACTIVE/DISABLED/READ_ONLY/CANCELLING）
4. **租户到期自动降级**：请求级触发（AuthMiddleware 中检查 expired_at）
5. **GetUserMenu 增强**：增加 platform 参数 + enabled_modules 过滤
6. **SUPER_ADMIN 保护**：Service 层校验（删除用户/删除角色/移除角色时检查）

## 技术选型

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 前缀匹配算法 | 有序 slice 倒序匹配（最长优先） | 应用数 < 100，O(n) 足够，实现简单 |
| 到期降级触发 | 请求级触发（AuthMiddleware 中） | 无需额外定时任务，实时性强（首次请求触发） |
| 公共接口路由 | `/api/v1/common/` 独立 group，不走 AppResolveMiddleware | 管理端和用户端共用，不绑定特定应用 |
| enabled_modules 检查位置 | AppResolveMiddleware 中（在权限检查之前） | 早期拦截，减少无效查询 |
| 前端路径替换 | 全局 find & replace + axios baseURL 配置 | 一次性完成，无技术负债 |

## 风险点

- [Risk-1] 路由前缀断裂迁移后，前端遗漏替换某个 API 路径 → 404。缓解：全局搜索 `/api/v1/` 确保无遗漏
- [Risk-2] AutoDiscover 扫描新前缀时，旧的 api_permission 数据残留。缓解：Seed 时先清空 admin_api_permission 表再重新发现
- [Risk-3] expired_at 请求级触发可能在高并发时多次 UPDATE。缓解：使用 CAS 条件更新 `WHERE status=1 AND expired_at < NOW()`

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 登录流程正常（用户名密码 → token） | /auth/login 接口正常返回 |
| RG-2 | 租户选择/切换正常 | /auth/tenant/select 正常签发 access_token |
| RG-3 | 角色 CRUD 正常 | 创建/更新/删除/查询角色无报错 |
| RG-4 | 用户 CRUD + 角色分配正常 | 创建用户 + 关联租户 + 分配角色流程完整 |
| RG-5 | 菜单权限分配和获取正常 | 角色分配菜单后 GetUserMenu 返回正确 |
| RG-6 | API 权限分配和动态检查正常 | 有权限的接口可访问，无权限返回 403 |
| RG-7 | SUPER_ADMIN 可访问所有接口 | SUPER_ADMIN 角色用户不被权限中间件拦截 |
| RG-8 | 操作日志正常记录 | 关键操作后 admin_operation_log 有新记录 |
