# 需求计划：V2-CR6 域名-租户映射管理模块

## 需求理解

- **目标**：在后端建立域名与租户的映射关系，C端前端通过公开 API 查询当前访问域名对应的 tenant_code，实现多租户 C端按域名自动路由，用户无需感知租户信息
- **范围**：
  - 后端：新增 `tenant_domain` 表 + CRUD 管理接口（超管操作）+ 公开查询接口（按域名查 tenant_code）
  - 前端 dev-web-user：登录页启动时调用公开接口自动获取 tenant_code，替换原来的静态配置
  - 前端 dev-web-admin：新增域名管理页面（含列表、新增、编辑、删除）
- **预期效果**：管理员在后台配置"域名 → 租户"绑定，C端用户访问不同域名时自动对应不同租户；未匹配的域名使用默认租户兜底

## 假设列表

- [假设-1] 一个域名只能绑定一个租户，但一个租户可以绑定多个域名（一对多）
- [假设-2] 默认租户（tenant_code=default）作为全局兜底，无需显式配置域名绑定
- [假设-3] 域名查询接口为公开接口（无需认证），C端在未登录状态就需要调用
- [假设-4] 域名管理属于超级管理员/平台管理操作，普通租户管理员不能操作
- [假设-5] 域名不含协议前缀（存 `abc.example.com`，不存 `https://abc.example.com`）
- [假设-6] 域名查询接口查不到时，返回默认租户 tenant_code，不返回错误

## 澄清问题

- [Question-1] 域名查询接口查不到时，是返回错误还是返回默认租户？
  **行业实践**：大多数 SaaS 平台选择返回默认值而非错误，避免前端处理复杂异常分支，提升容错能力。推荐返回默认租户。
  [Answer-1]
同意
- [Question-2] 域名管理的权限控制：仅超级管理员可操作，还是租户管理员也可以管理自己租户的域名绑定？
  **行业实践**：平台型 SaaS 通常由平台运营统一管理（防止租户随意绑定他人域名），租户侧自助绑定通常需配合域名所有权验证。推荐先实现超级管理员模式，自助绑定后续扩展。
  [Answer-2]
同意
- [Question-3] 是否需要支持通配符域名匹配？如 `*.example.com` 匹配所有子域名到同一租户？
  **行业实践**：通配符实现复杂度较高（需正则或前缀匹配），如果每个租户只有固定几个域名，精确匹配足够用。
  [Answer-3]
同意，先实现精确匹配，但结构上需要预留扩展通配符匹配
- [Question-4] 是否需要在管理端前端新增域名管理页面，还是仅提供后端接口（通过 API 工具管理）？
  [Answer-4]
在管理端增加管理页面
- [Question-5] 域名查询是否需要缓存？域名-租户关系变化频率低，内存缓存可以显著减少 DB 查询。推荐 TTL 5 分钟，变更时主动失效。
  [Answer-5]
增加缓存
## 非功能需求建议

- **性能**：域名查询接口 < 50ms（P99），命中缓存后接近本地内存速度
- **安全**：公开查询接口仅返回 tenant_code，不暴露 tenant_id 等内部字段；CRUD 接口需 JWT 认证
- **多租户**：域名管理本身是平台级，不走 TenantIsolationCallback

## 影响范围预判

- **新建文件**：
  - `common/auth/model/tenant_domain.go`
  - `common/auth/repository/tenant_domain_repo.go`
  - `common/auth/service/tenant_domain_service.go`
  - `common/auth/handler/tenant_domain_handler.go`
  - `dev-web-admin/src/views/system/tenant-domain/index.vue`
  - `dev-web-admin/src/api/tenant-domain.ts`
- **改动文件**：
  - `common/auth/auth.go`（AutoMigrate + 路由注册）
  - `common/auth/router.go`（注册 CRUD 路由 + 公开路由）
  - `common/auth/seed.go`（补充菜单 seed）
  - `dev-web-admin/src/router/static-routes.ts`（添加路由）
  - `dev-web-user/src/utils/tenant.ts`（改为 API 调用）
  - `dev-web-user/src/views/login/index.vue`（异步获取 tenant_code）
- **可能副作用**：
  - 公开路由注册需在认证中间件之前（不走 AuthMiddleware）
  - dev-web-user 登录页需处理 tenant 查询失败的兜底逻辑
