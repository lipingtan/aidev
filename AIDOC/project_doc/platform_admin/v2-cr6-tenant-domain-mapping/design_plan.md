# 设计计划：V2-CR6 域名-租户映射管理模块

## 设计方向

### 整体架构

```
dev-web-user (登录页)
    │  GET /api/v1/public/tenant-domain?domain=xxx
    ▼
后端公开路由（无 AuthMiddleware）
    │
    ├─ 缓存命中 → 直接返回 tenant_code + tenant_name
    │
    └─ 缓存 miss → TenantDomainService.QueryByDomain(domain)
                     │
                     ├─ 精确匹配 tenant_domain 表
                     └─ 未找到 → 返回 default 租户
                                  │
                                  └─ 结果写入 LocalCache（TTL 5min）

dev-web-admin (域名管理页)
    │  CRUD /api/v1/admin/tenant-domains
    ▼
认证 + DynamicPermissionMiddleware
    │
TenantDomainHandler → TenantDomainService
    │  写操作完成后 → 主动失效该域名缓存
    ▼
tenant_domain 表（MySQL）
```

### 后端代码组织

完全在 `common/auth/` 包内，与现有 Tenant 模块平级：

```
common/auth/
├── model/
│   └── tenant_domain.go        # TenantDomain model
├── repository/
│   └── tenant_domain_repo.go   # Repository 接口 + GORM 实现
├── service/
│   └── tenant_domain_service.go # 业务逻辑 + 缓存管理
└── handler/
    └── tenant_domain_handler.go # CRUD + 公开查询
```

### 缓存设计

复用现有 `cache.LocalCache`（已有 TTL + 后台清理）：

```
缓存 key：tenant_domain:{domain}
缓存 value：QueryDomainResult{TenantCode, TenantName}
TTL：5 分钟
失效触发：Create / Update / Delete 操作完成后 Delete(key)
```

`TenantDomainService` 持有一个 `*cache.LocalCache` 实例（在 `buildDependencies` 中注入或新建）。

## 技术选型

| 方案 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 复用现有 `LocalCache` | 无新依赖，统一缓存基础设施 | 单实例内存，多实例部署需各自回填 | ✓ |
| 新建专用 sync.Map | 更轻量 | 没有 TTL/自动清理，需自己实现 | ✗ |
| Redis 缓存 | 多实例共享一致 | 当前未引入 Redis | ✗ |

多实例说明：当前部署为单实例；如未来横向扩展，变更时主动失效本实例缓存 + DB 最终一致仍可保证正确性（最多 5 分钟漂移），可接受。

## 澄清问题

- [Question-1] 公开查询接口路径放在 `/api/v1/public/` 还是 `/api/v1/user/` 下？
  **考量**：`/api/v1/public/` 语义更清晰（表达"任何人可访问"），且与现有 `/api/v1/user/auth/`（需认证）区分明确。`/api/v1/user/` 下也有无需认证的路由（send-code/login），但混在一起不直观。
  **推荐**：使用 `/api/v1/public/tenant-domain?domain={hostname}` ，独立的公开路由组，在 `router.go` 注册时不挂 AuthMiddleware。
  [Answer-1]
按独立公开路由
- [Question-2] `tenant_domain` 表中 `tenant_id` 存数字 ID 还是 `tenant_code` 存字符串？
  **考量**：存 `tenant_id`（int64）+外键关联性能更好，查询时 JOIN 取 tenant_code；存 `tenant_code` 则无需 JOIN 但存在冗余。现有 Tenant model 两个字段都有。
  **推荐**：存 `tenant_id`（与现有模型风格一致），查询时 JOIN admin_tenant 取 tenant_code + name，返回给前端。
  [Answer-2]
同意
- [Question-3] 管理端 CRUD 接口是否需要通过 DynamicPermissionMiddleware 检查 permission_code？如需要，permission_code 建议为 `system:tenant-domain:list` / `system:tenant-domain:create` 等。
  **推荐**：走 DynamicPermissionMiddleware（与其他管理端接口一致），permission_code 前缀为 `system:tenant-domain:`。同时在 Seed 中注册对应 API 权限记录，确保 SUPER_ADMIN 自动获得权限。
  [Answer-3]
同意
## 风险点

- [Risk-1] 公开路由路径需在 `router.go` 中在 AuthMiddleware **之前**注册，否则会被 401 拦截。当前 `RegisterRoutes` 结构中，公开路由通过不加 middleware 的 Group 注册，需确保新的 public 路由组同样如此。
- [Risk-2] 缓存 key 冲突：其他模块也可能用 `LocalCache` 存储域名相关数据，key 命名需加前缀 `tenant_domain:` 以隔离。
- [Risk-3] `localhost`/`127.0.0.1` 保留约束需在 Service 层校验（Create/Update 时拒绝），不能只依赖 DB 唯一索引。
- [Risk-4] dev-web-user 登录页的 tenant_code 解析是异步操作（网络请求），需处理请求进行中时用户点击发送验证码的竞态，建议用 `loading` 状态禁用发送按钮直到解析完成。
