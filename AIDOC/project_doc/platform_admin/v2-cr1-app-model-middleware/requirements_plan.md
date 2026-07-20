# 需求计划：CR-1 应用模型升级 + 中间件链重构

## 需求理解

- 目标：建立 V2 架构的核心骨架 — 统一应用模型（含 app_type/route_prefix/platforms/modules）和请求级应用识别中间件（AppResolveMiddleware）
- 范围：DDL 字段扩展 + 新增中间件 + API 路由前缀迁移 + GetUserMenu 增强 + 租户四态 + 管理员保护规则 + 前端路径适配
- 预期效果：
  - 后端 API 统一使用 `/api/v1/admin/` 前缀（platform_admin 应用）
  - 每个请求自动识别所属应用并校验租户订阅
  - 菜单支持 platform 和 module_code 维度
  - 租户支持四态生命周期（正常/禁用/只读/注销中）
  - SUPER_ADMIN 保护规则生效

## 假设列表

- [假设-1] 从 V1 渐进式重构，不清空现有代码，通过 GORM AutoMigrate 自动加字段
- [假设-2] 前端路径迁移一次性完成（`/api/v1/xxx` → `/api/v1/admin/xxx`），前后端同步改
- [假设-3] `admin_tenant.config` JSON 列保留（不删除），timezone/locale/currency 作为独立列添加
- [假设-4] 种子数据全部重新生成（drop 旧数据重建），因为未上生产
- [假设-5] 本 CR 不涉及：角色继承校验、权限集、字段权限、记录共享、biz_user、审批流（后续 CR）
- [假设-6] 重新初始化流程（drop 全部 admin_* 表 → AutoMigrate 重建 → Seed 写入新种子数据）必须作为本 CR 的强制验收项，确保全新空库启动时系统完整可用

## 澄清问题

- [Question-1] 路由迁移策略：是否在 CR-1 中同时保留旧路径做兼容（如 `/api/v1/tenants` 也能访问），还是直接断裂迁移到 `/api/v1/admin/tenants`？鉴于未上生产，建议直接断裂。
  [Answer-1]
直接断裂，避免某些未迁移而应该迁移的逻辑访问旧的路由无法发现
- [Question-2] `admin_tenant.status` 四态的实际触发：READ_ONLY(2) 状态是否在 CR-1 中实现自动触发（expired_at 到期自动降级）？还是仅实现手动切换 + AuthMiddleware 识别？
  [Answer-2]
自动触发
- [Question-3] enabled_modules 的运行时过滤：CR-1 中是否实现 DynamicPermissionMiddleware 中的 module_code 检查？还是仅实现 GetUserMenu 过滤，运行时检查留到后续 CR？
  [Answer-3]
需要实现module_code检查
## 非功能需求建议

- 性能：AppResolveMiddleware 前缀匹配使用启动时构建的有序 map，O(1) 或 O(log n)
- 回归：现有登录/角色/用户/资源/API权限等全部功能不中断
- 兼容：AutoMigrate 自动加字段，旧数据在新字段中有合理默认值
- **初始化正确性**：全新空库（drop 全部表后）启动时，AutoMigrate + Seed + AutoDiscover 完整链路必须正常运行，结果包括：
  - 所有表结构正确创建（含新增字段）
  - 超级管理员 + 默认租户 + SUPER_ADMIN 角色 + 应用 + 租户订阅 + 菜单 + API 权限种子数据完整写入
  - AutoDiscover 正确扫描新前缀路由并注册到 admin_api_permission（app_code=platform_admin）
  - 种子菜单包含 platform 和 module_code 字段正确值
  - 系统可直接登录使用（无需任何人工干预）

## 影响范围预判

- 涉及模块：
  - `backend/common/auth/model/` — 4-5 个 Model 加字段
  - `backend/common/auth/middleware/` — 新增 app_resolve.go，修改 auth_middleware.go（四态）
  - `backend/common/auth/handler/` — 路由前缀变更
  - `backend/common/auth/router.go` — 路由注册重构
  - `backend/common/auth/service/resource_service.go` — GetUserMenu 增强
  - `backend/common/auth/service/tenant_service.go` — 四态状态转换
  - `backend/common/auth/discovery/` — AutoDiscover 适配新前缀
  - `backend/common/auth/auth.go` — 初始化流程调整
  - `dev-web-admin/src/api/` — 全部 API 路径修改
  - `dev-web-admin/src/views/` — 应用管理页增加新字段展示
- 涉及文件数预估：后端 ~20 文件，前端 ~15 文件
