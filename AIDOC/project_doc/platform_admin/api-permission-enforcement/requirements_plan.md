# 需求计划：API 接口权限检查实施

## 需求理解

- 目标：将 PermissionMiddleware 挂载到路由链上，使 API 接口权限配置实际生效
- 当前状态：数据模型已就绪（admin_role_api），PermissionMiddleware 代码已有，但未注册到路由
- 预期效果：非 SUPER_ADMIN 用户访问未授权的 API 时返回 403

## 当前所有 API 对象及注解状态

### 说明

- **注解码（permission_code）**：存储在 admin_api_permission 表中，AutoDiscover 注册时未设置（为空）
- **当前状态**：所有接口均未注解 permission_code，PermissionMiddleware 未挂载
- **设计策略**：`annotation-first` — 无注解的接口默认放行，有注解的才校验

### 对象及接口清单

| 对象 | 方法 | 路径 | 建议 permission_code | 是否需要权限控制 | 当前状态 |
|------|------|------|---------------------|-----------------|----------|
| **users** | GET | /api/v1/users | user:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/users | user:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/users/:id | user:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/users/:id | user:delete | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/users/:id/tenants | user:tenant:assign | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/users/:id/tenants/:tenantId | user:tenant:remove | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/users/:id/tenants | user:tenant:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/users/:id/roles | user:role:assign | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/users/:id/roles | user:role:replace | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/users/:id/force-offline | user:force-offline | ✅ 需要 | ❌ 未注解 |
| **roles** | GET | /api/v1/roles | role:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/roles | role:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/roles/:id | role:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/roles/:id | role:delete | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/roles/:id/resources | role:resource:list | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/roles/:id/resources | role:resource:assign | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/roles/:id/apis | role:api:list | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/roles/:id/apis | role:api:assign | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/roles/:id/apps | role:app:list | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/roles/:id/apps | role:app:assign | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/roles/:id/permission-summary | role:permission:summary | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/roles/:id/data-scopes | role:data-scope:list | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/roles/:id/data-scopes | role:data-scope:assign | ✅ 需要 | ❌ 未注解 |
| **tenants** | GET | /api/v1/tenants | tenant:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/tenants | tenant:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/tenants/:id | tenant:update | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/tenants/:id/status | tenant:status | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/tenants/:id | tenant:delete | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/tenants/:id/apps | tenant:app:assign | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/tenants/:id/apps | tenant:app:list | ✅ 需要 | ❌ 未注解 |
| **resources** | GET | /api/v1/resources/tree | resource:tree | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/resources | resource:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/resources/:id | resource:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/resources/:id | resource:delete | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/resources/sort | resource:sort | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/resources/user-menu | — | ❌ 免检（所有登录用户可用） | — |
| **api-permissions** | GET | /api/v1/api-permissions/tree | api-perm:tree | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/api-permissions | api-perm:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/api-permissions/:id | api-perm:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/api-permissions/:id | api-perm:delete | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/api-permissions/:id/move | api-perm:move | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/api-permissions/:id/visible | api-perm:visible | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/api-permissions/unassigned | api-perm:unassigned | ✅ 需要 | ❌ 未注解 |
| **applications** | GET | /api/v1/applications | app:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/applications | app:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/applications/:id | app:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/applications/:id | app:delete | ✅ 需要 | ❌ 未注解 |
| **configs** | GET | /api/v1/configs | config:list | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/configs/:id | config:detail | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/configs/key/:key | config:detail | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/configs | config:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/configs/:id | config:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/configs/:id | config:delete | ✅ 需要 | ❌ 未注解 |
| **login-logs** | GET | /api/v1/login-logs | log:login:list | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/login-logs/:id | log:login:delete | ✅ 需要 | ❌ 未注解 |
| **operation-logs** | GET | /api/v1/operation-logs | log:operation:list | ✅ 需要 | ❌ 未注解 |
| **data-scope-configs** | GET | /api/v1/data-scope-configs | data-scope:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/data-scope-configs | data-scope:create | ✅ 需要 | ❌ 未注解 |
| | PUT | /api/v1/data-scope-configs/:id | data-scope:update | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/data-scope-configs/:id | data-scope:delete | ✅ 需要 | ❌ 未注解 |
| **plugins** | GET | /api/v1/plugins | plugin:list | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/plugins/install | plugin:install | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/plugins/:name/start | plugin:start | ✅ 需要 | ❌ 未注解 |
| | POST | /api/v1/plugins/:name/stop | plugin:stop | ✅ 需要 | ❌ 未注解 |
| | DELETE | /api/v1/plugins/:name | plugin:uninstall | ✅ 需要 | ❌ 未注解 |
| | GET | /api/v1/plugins/:name/health | plugin:health | ✅ 需要 | ❌ 未注解 |
| **免检接口** | | | | | |
| auth | POST | /auth/login | — | ❌ 公开 | — |
| | POST | /auth/tenant/select | — | ❌ 仅需 platform_token | — |
| | POST | /auth/refresh | — | ❌ 仅需 platform_token | — |
| | POST | /auth/logout | — | ❌ 仅需 access_token | — |
| captcha | GET | /captcha | — | ❌ 公开 | — |
| setup | * | /setup/* | — | ❌ 安装向导 | — |
| dashboard | GET | /api/v1/dashboard/* | — | ❌ 所有登录用户可看 | — |
| monitor | GET | /api/v1/monitor/server | — | ❌ stub | — |
| sys-apis | GET | /api/v1/sys-apis | — | ❌ stub | — |
| resources | GET | /api/v1/resources/user-menu | — | ❌ 所有登录用户可调 | — |

## 澄清问题

- [Question-1] PermissionMiddleware 策略确认：无 permission_code 注解的接口默认放行，还是默认拒绝？
  [Answer-1]
无注解的，默认拒绝，所有要放行的接口，显式增加一个忽略注解,请仔细检查现有实现确保，已有需要免检的都增加了忽略注解
- [Question-2] permission_code 的来源：是在 handler 代码中硬编码（路由元数据），还是从数据库 admin_api_permission 表中按 url_pattern+method 动态匹配？
  [Answer-2]
按url_pattern+method 动态匹配 是否一定能满足要求？是的话就这样
- [Question-3] 上述建议的 permission_code 命名格式（如 `user:list`、`role:create`）是否可接受？
  [Answer-3]
可以
## 实施方案预判

1. 给所有需要权限控制的接口设置 permission_code（AutoDiscover 写入时从映射表填充）
2. 在 router.go 的 api group 中添加 PermissionMiddleware
3. PermissionMiddleware 按 request path + method 从缓存匹配 permission_code，再比对用户权限集
4. 免检接口（auth/setup/dashboard/user-menu）不经过 PermissionMiddleware 或在白名单中
