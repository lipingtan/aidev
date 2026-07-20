# 测试用例：CR-1 Step A — 纯迁移验收

**关联文档**: design.md / tasks.md
**编写日期**: 2026-07-20
**测试类型**: 功能测试 / 接口测试 / 回归测试
**测试方法**: 等价类 + 边界值 + 状态转换 + 错误推测 + 场景法
**测试范围**: Task 1（Model 层）+ Task 2（路由前缀）+ Task 3（AutoDiscover）+ Task 4（Seed）+ Task 5（前端）

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | 接口 | 数据 | 合计 |
|------|------|------|------|------|------|------|------|
| Model 字段扩展 | 2 | 0 | 0 | 0 | 0 | 3 | 5 |
| 路由前缀迁移 | 4 | 2 | 0 | 8 | 6 | 0 | 20 |
| AutoDiscover | 3 | 0 | 1 | 0 | 0 | 3 | 7 |
| Seed 数据 | 2 | 0 | 0 | 0 | 0 | 5 | 7 |
| 前端路径迁移 | 3 | 1 | 0 | 0 | 0 | 0 | 4 |
| **总计** | **14** | **3** | **1** | **8** | **6** | **11** | **43** |

---

## 一、正向测试（Happy Path）

### TC-001: 后端编译通过

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 代码已合入本地 |
| **测试步骤** | 1. `cd backend` <br> 2. `go build ./...` |
| **预期结果** | 零错误零警告退出 |
| **关联需求** | Task 1-4 共同 AC |

### TC-002: 空库启动 AutoMigrate 成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 删除所有 admin_* 表（或新建空库） |
| **测试步骤** | 1. 启动后端服务 <br> 2. 观察日志输出 |
| **预期结果** | 日志显示 `[auth-rbac] 模块初始化完成`，数据库中 admin_* 表全部创建成功 |
| **关联需求** | Task 1 AC |

### TC-003: 新字段有正确默认值

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | TC-002 通过后 |
| **测试步骤** | 查询表结构确认默认值 |
| **预期结果** | admin_application.app_type DEFAULT 'BUILTIN'；admin_resource.platform DEFAULT 'admin'；admin_tenant.timezone DEFAULT 'Asia/Shanghai' |
| **关联需求** | Task 1 AC |

### TC-004: 登录获取 platform_token

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | Seed 完成，admin/admin123 可用 |
| **测试步骤** | POST /auth/login `{"username":"admin","password":"admin123"}` |
| **预期结果** | HTTP 200, code=0, data.token 非空, data.tenants 含 default 租户 |
| **关联需求** | RG-1 |

### TC-005: 选择租户获取 access_token

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | TC-004 获取 platform_token |
| **测试步骤** | POST /auth/tenant/select `{"tenant_id":"{default_tenant_id}"}` + Bearer platform_token |
| **预期结果** | HTTP 200, code=0, data.access_token 非空 |
| **关联需求** | RG-2 |

### TC-006: 新路径访问租户列表

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | TC-005 获取 access_token |
| **测试步骤** | GET /api/v1/admin/tenants + Bearer access_token |
| **预期结果** | HTTP 200, code=0, data.list 含 default 租户 |
| **关联需求** | Task 2 AC |

### TC-007: common 路径获取用户菜单

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | TC-005 获取 access_token |
| **测试步骤** | GET /api/v1/common/user-menu?platform=admin + Bearer access_token |
| **预期结果** | HTTP 200, code=0, data 为菜单树数组（含首页/系统管理等） |
| **关联需求** | Task 2 AC, RG-5 |

### TC-008: AutoDiscover 注册 API 权限

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 服务启动完成 |
| **测试步骤** | `SELECT count(*) FROM admin_api_permission WHERE type='ENDPOINT' AND app_code='platform_admin'` |
| **预期结果** | 记录数 > 0（至少 30+ 条，覆盖所有已注册路由） |
| **关联需求** | Task 3 AC |

### TC-009: AutoDiscover 写入 module_code

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | TC-008 通过 |
| **测试步骤** | `SELECT DISTINCT module_code FROM admin_api_permission WHERE app_code='platform_admin' AND module_code != ''` |
| **预期结果** | 包含 user-mgmt, role-mgmt, tenant-mgmt, resource-mgmt, api-perm-mgmt 等 |
| **关联需求** | Task 3 AC |

### TC-010: AutoDiscover 幂等（重启不重复）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 服务已启动过一次 |
| **测试步骤** | 1. 记录当前记录数 <br> 2. 重启服务 <br> 3. 再次查询记录数 |
| **预期结果** | 两次记录数相同 |
| **关联需求** | Task 3 AC |

### TC-011: Seed 创建完整种子数据

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 空库启动 |
| **测试步骤** | 查询各种子表 |
| **预期结果** | admin_user(1条) + admin_tenant(1条) + admin_role(1条,SUPER_ADMIN) + admin_application(1条,platform_admin) + admin_tenant_app(1条) + admin_resource(≥12条) |
| **关联需求** | Task 4 AC |

### TC-012: Seed 应用字段完整

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | TC-011 通过 |
| **测试步骤** | `SELECT app_code, app_type, route_prefix, platforms FROM admin_application WHERE app_code='platform_admin'` |
| **预期结果** | app_type='BUILTIN', route_prefix='/api/v1/admin', platforms='["admin:pc","admin:h5"]' |
| **关联需求** | Task 4 AC |

### TC-013: 前端编译通过

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 前端代码已更新 |
| **测试步骤** | `cd dev-web-admin && npx vue-tsc --noEmit` |
| **预期结果** | 零 TS 错误 |
| **关联需求** | Task 5 AC |

### TC-014: 前端登录+菜单加载

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 后端运行 + 前端 dev 服务运行 |
| **测试步骤** | 1. 打开浏览器访问前端 <br> 2. 输入 admin/admin123 登录 <br> 3. 选择默认租户 |
| **预期结果** | 侧边栏菜单正确渲染（首页/系统管理/日志管理/监控） |
| **关联需求** | Task 5 AC, RG-5 |

---

## 二、反向测试（Negative）

### TC-N01: 旧路径返回 404

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 服务运行中 |
| **测试步骤** | GET /api/v1/tenants + Bearer access_token |
| **预期结果** | HTTP 404 |
| **关联需求** | Task 2 AC |

### TC-N02: 旧 user-menu 路径 404

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 服务运行中 |
| **测试步骤** | GET /api/v1/resources/user-menu + Bearer access_token |
| **预期结果** | HTTP 404 |
| **关联需求** | Task 2 AC |

### TC-N03: 前端无旧路径调用

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 前端代码 |
| **测试步骤** | 全局搜索 `request.get('/api/v1/tenants` 或 `request.post('/api/v1/users` 等旧路径 |
| **预期结果** | 无匹配结果（全部已迁移到 /api/v1/admin/ 或 /api/v1/common/） |
| **关联需求** | Task 5 AC |

---

## 三、边界测试（Boundary）

### TC-B01: AutoDiscover 异步阈值边界

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | asyncThreshold 设为 1（路由数必然超过） |
| **预期结果** | 启动后立即查询记录为 0，等 5s 后查询有记录（异步执行） |
| **设计方法** | 边界值分析 |

---

## 四、回归测试（Regression）

### TC-R01: 登录流程不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | POST /auth/login `{"username":"admin","password":"admin123"}` |
| **预期结果** | HTTP 200, code=0, 返回 token + tenants 列表 |

### TC-R02: 租户选择不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | POST /auth/tenant/select + Bearer platform_token |
| **预期结果** | HTTP 200, 返回 access_token |

### TC-R03: 角色 CRUD（新路径）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. POST /api/v1/admin/roles 创建 <br> 2. GET /api/v1/admin/roles 查询 <br> 3. PUT /api/v1/admin/roles/:id 更新 <br> 4. DELETE /api/v1/admin/roles/:id 删除 |
| **预期结果** | 全部 HTTP 200, CRUD 完整 |

### TC-R04: 用户 CRUD + 角色分配

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | 1. 创建用户 <br> 2. 关联租户 <br> 3. 分配角色 <br> 4. 查询确认 |
| **预期结果** | 全流程无报错，数据正确写入 |

### TC-R05: 菜单权限分配和获取

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-5 |
| **验证步骤** | 1. PUT /api/v1/admin/roles/:id/resources 分配菜单 <br> 2. GET /api/v1/common/user-menu?platform=admin 获取菜单 |
| **预期结果** | 分配后 GetUserMenu 返回对应菜单项 |

### TC-R06: API 权限动态检查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | RG-6 |
| **验证步骤** | 1. 创建普通角色（无任何 API 权限） <br> 2. 用该角色用户访问 /api/v1/admin/tenants |
| **预期结果** | HTTP 403, code=40301 |

### TC-R07: SUPER_ADMIN 全通

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-7 |
| **验证步骤** | admin 用户（SUPER_ADMIN）访问全部业务接口 |
| **预期结果** | 不被 DynamicPermissionMiddleware 拦截，全部返回正常数据 |

### TC-R08: 操作日志记录

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | RG-8 |
| **验证步骤** | 1. 创建租户 <br> 2. GET /api/v1/admin/operation-logs?module=tenant |
| **预期结果** | 日志列表含该创建操作记录 |

---

## 五、接口测试（API Level）

### TC-A01: POST /auth/login — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `POST /auth/login` |
| **Body** | `{"username":"admin","password":"admin123"}` |
| **预期响应** | HTTP 200, `{"code":0,"data":{"token":"...","tenants":[...]}}` |

### TC-A02: POST /auth/login — 密码错误

| 字段 | 内容 |
|------|------|
| **请求** | `POST /auth/login` |
| **Body** | `{"username":"admin","password":"wrong"}` |
| **预期响应** | HTTP 401, code=40100 |

### TC-A03: GET /api/v1/admin/tenants — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/tenants`（无 Authorization header） |
| **预期响应** | HTTP 401, code=40101 |

### TC-A04: GET /api/v1/admin/tenants — 正常认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/tenants` + Bearer access_token |
| **预期响应** | HTTP 200, `{"code":0,"data":{"list":[...],"total":N}}` |

### TC-A05: GET /api/v1/common/user-menu — 缺少 platform 参数

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/common/user-menu` + Bearer access_token |
| **预期响应** | HTTP 200, 返回菜单（当前实现不强制 platform，返回全部） |

### TC-A06: GET /api/v1/common/user-menu?platform=admin — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/common/user-menu?platform=admin` + Bearer access_token |
| **预期响应** | HTTP 200, data 为菜单树数组，每项含 id/name/path/children |

---

## 六、数据验证

### TC-D01: admin_application 记录完整

| 字段 | 内容 |
|------|------|
| **触发操作** | 空库启动 |
| **验证 SQL** | `SELECT app_code, app_type, route_prefix, platforms FROM admin_application WHERE app_code='platform_admin'` |
| **预期结果** | 1 条记录: BUILTIN / /api/v1/admin / ["admin:pc","admin:h5"] |

### TC-D02: admin_resource 新字段正确

| 字段 | 内容 |
|------|------|
| **触发操作** | Seed 执行后 |
| **验证 SQL** | `SELECT platform, app_code, module_code FROM admin_resource WHERE app_code='platform_admin' LIMIT 5` |
| **预期结果** | platform='admin', app_code='platform_admin', module_code 非空 |

### TC-D03: admin_tenant 新字段正确

| 字段 | 内容 |
|------|------|
| **触发操作** | Seed 执行后 |
| **验证 SQL** | `SELECT timezone, locale, currency FROM admin_tenant WHERE tenant_code='default'` |
| **预期结果** | Asia/Shanghai / zh-CN / CNY |

### TC-D04: admin_api_permission app_code 全部为 platform_admin

| 字段 | 内容 |
|------|------|
| **触发操作** | AutoDiscover 执行后 |
| **验证 SQL** | `SELECT count(*) FROM admin_api_permission WHERE app_code='platform_admin' AND type='ENDPOINT'` |
| **预期结果** | ≥ 30 |

### TC-D05: admin_api_permission module_code 非空覆盖

| 字段 | 内容 |
|------|------|
| **触发操作** | AutoDiscover 执行后 |
| **验证 SQL** | `SELECT count(*) FROM admin_api_permission WHERE app_code='platform_admin' AND type='ENDPOINT' AND module_code != ''` |
| **预期结果** | 等于 TC-D04 的总数（全部有 module_code） |

### TC-D06: admin_tenant_app 订阅正确

| 字段 | 内容 |
|------|------|
| **触发操作** | Seed 执行后 |
| **验证 SQL** | `SELECT * FROM admin_tenant_app WHERE app_code='platform_admin'` |
| **预期结果** | 1 条记录，tenant_id 为默认租户 ID |

### TC-D07: Seed 幂等（重复启动不重复插入）

| 字段 | 内容 |
|------|------|
| **触发操作** | 重启服务 |
| **验证 SQL** | `SELECT count(*) FROM admin_user` |
| **预期结果** | 仍为 1（不增加） |

### TC-D08: AutoDiscover 幂等

| 字段 | 内容 |
|------|------|
| **触发操作** | 重启服务 |
| **验证 SQL** | `SELECT count(*) FROM admin_api_permission WHERE type='ENDPOINT'` |
| **预期结果** | 与首次启动相同 |

### TC-D09: admin_resource 菜单数量

| 字段 | 内容 |
|------|------|
| **触发操作** | Seed 执行后 |
| **验证 SQL** | `SELECT count(*) FROM admin_resource WHERE app_code='platform_admin' AND platform='admin'` |
| **预期结果** | ≥ 12（4 一级 + 8 系统子 + 2 日志子 + 1 监控子 = 15） |

### TC-D10: admin_tenant_app enabled_modules 字段存在

| 字段 | 内容 |
|------|------|
| **触发操作** | 检查表结构 |
| **验证 SQL** | `SHOW COLUMNS FROM admin_tenant_app LIKE 'enabled_modules'` |
| **预期结果** | 字段存在，类型为 json，允许 NULL |

### TC-D11: admin_application 新字段存在

| 字段 | 内容 |
|------|------|
| **触发操作** | 检查表结构 |
| **验证 SQL** | `SHOW COLUMNS FROM admin_application` |
| **预期结果** | 含 app_type / route_prefix / platforms / modules / app_config / icon / sort_order 字段 |

---

## 执行结果记录

| 用例编号 | 结果 | 执行人 | 日期 | 备注 |
|----------|------|--------|------|------|
| TC-001 | ⬜ | | | |
| TC-002 | ⬜ | | | |
| TC-003 | ⬜ | | | |
| TC-004 | ⬜ | | | |
| TC-005 | ⬜ | | | |
| TC-006 | ⬜ | | | |
| TC-007 | ⬜ | | | |
| TC-008 | ⬜ | | | |
| TC-009 | ⬜ | | | |
| TC-010 | ⬜ | | | |
| TC-011 | ⬜ | | | |
| TC-012 | ⬜ | | | |
| TC-013 | ⬜ | | | |
| TC-014 | ⬜ | | | |
| TC-N01 | ⬜ | | | |
| TC-N02 | ⬜ | | | |
| TC-N03 | ⬜ | | | |
| TC-B01 | ⬜ | | | |
| TC-R01 | ⬜ | | | |
| TC-R02 | ⬜ | | | |
| TC-R03 | ⬜ | | | |
| TC-R04 | ⬜ | | | |
| TC-R05 | ⬜ | | | |
| TC-R06 | ⬜ | | | |
| TC-R07 | ⬜ | | | |
| TC-R08 | ⬜ | | | |
| TC-A01 | ⬜ | | | |
| TC-A02 | ⬜ | | | |
| TC-A03 | ⬜ | | | |
| TC-A04 | ⬜ | | | |
| TC-A05 | ⬜ | | | |
| TC-A06 | ⬜ | | | |
| TC-D01 | ⬜ | | | |
| TC-D02 | ⬜ | | | |
| TC-D03 | ⬜ | | | |
| TC-D04 | ⬜ | | | |
| TC-D05 | ⬜ | | | |
| TC-D06 | ⬜ | | | |
| TC-D07 | ⬜ | | | |
| TC-D08 | ⬜ | | | |
| TC-D09 | ⬜ | | | |
| TC-D10 | ⬜ | | | |
| TC-D11 | ⬜ | | | |
