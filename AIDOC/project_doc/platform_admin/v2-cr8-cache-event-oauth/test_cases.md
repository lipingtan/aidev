# 测试用例：V2-CR8 缓存事件驱动 + OAuth2 预留 + 扩展能力收尾

## 一、接口测试（TC-A 系列）

### Phase 1: 权限缓存 + 事件驱动

| 用例ID | 描述 | 优先级 | 前置条件 | 验证点 |
|--------|------|--------|---------|--------|
| TC-A01 | 权限接口正常放行（缓存降级仍可用） | P0 | 已登录 admin | 有权限接口返回 200，不因缓存问题中断 |
| TC-A02 | 无权限接口被拦截 | P0 | 普通用户 token | 无权限接口返回 403 |
| TC-A03 | 分配角色后新权限立即生效（缓存失效验证） | P1 | 创建用户+角色+接口权限 | 分配角色前 403，分配后 200 |

### Phase 2: 操作日志风险分级

| 用例ID | 描述 | 优先级 | 前置条件 | 验证点 |
|--------|------|--------|---------|--------|
| TC-A04 | 查询操作日志（SUPER_ADMIN 可查所有） | P0 | admin 登录 | 返回 200，code=0 |
| TC-A05 | 按 risk_level=HIGH 过滤日志 | P1 | 执行过高风险操作 | 返回结果均为 HIGH |
| TC-A06 | 按 risk_level=LOW 过滤日志 | P1 | 存在 LOW 日志 | 返回结果均为 LOW |
| TC-A07 | 操作日志 risk_level 字段存在 | P0 | 任意操作日志记录 | 响应体每条记录含 risk_level 字段 |

### Phase 3: OAuth2/LDAP 骨架

| 用例ID | 描述 | 优先级 | 前置条件 | 验证点 |
|--------|------|--------|---------|--------|
| TC-A08 | grant_type=oauth2 返回 501 | P0 | 无需登录 | HTTP 200，code=50101，message 含"OAuth2" |
| TC-A09 | grant_type=ldap 返回 501 | P0 | 无需登录 | HTTP 200，code=50101，message 含"LDAP" |
| TC-A10 | grant_type=password 不受影响 | P0 | admin 账号 | 正常登录返回 token |

### Phase 4: ext_fields + admin_custom_field

| 用例ID | 描述 | 优先级 | 前置条件 | 验证点 |
|--------|------|--------|---------|--------|
| TC-A11 | 用户创建时不传 ext_fields 正常 | P0 | admin 登录 | 创建用户成功，无报错 |
| TC-A12 | 租户创建时不传 ext_fields 正常 | P0 | admin 登录 | 创建租户成功，无报错 |
| TC-A13 | 用户接口响应包含 ext_fields 字段 | P1 | admin 登录 | 用户列表/详情含 ext_fields（可为 null） |
| TC-A14 | 租户接口响应包含 ext_fields 字段 | P1 | admin 登录 | 租户列表/详情含 ext_fields（可为 null） |
| TC-A15 | admin_custom_field 表存在（系统启动验证） | P1 | 服务已启动 | 后端启动无 AutoMigrate 错误（通过 status 接口验证） |

---

## 二、正向测试（TC-P 系列）

| 用例ID | 描述 | 优先级 | 验证点 |
|--------|------|--------|--------|
| TC-P01 | 高风险操作（删除租户）后日志 risk_level=HIGH | P1 | 删除租户 → 查操作日志 → risk_level=HIGH |
| TC-P02 | 普通写操作日志 risk_level=LOW | P1 | 创建角色 → 查操作日志 → risk_level=LOW |
| TC-P03 | grant_type=password 登录返回有效 token | P0 | 正常登录返回 access_token |
| TC-P04 | 权限检查在无 Redis 时降级正常（内存缓存） | P0 | 登录后访问受保护接口返回 200 |

---

## 三、反向测试（TC-N 系列）

| 用例ID | 描述 | 优先级 | 验证点 |
|--------|------|--------|--------|
| TC-N01 | grant_type=oauth2 不返回 token | P0 | 响应无 token 字段 |
| TC-N02 | grant_type=ldap 不返回 token | P0 | 响应无 token 字段 |
| TC-N03 | grant_type=unknown 返回"不支持"错误 | P1 | code 非 0，message 含"不支持"或类似 |
| TC-N04 | 未认证请求被拒绝（不因缓存问题放行） | P0 | 无 token 请求返回 401 |

---

## 四、回归测试（TC-R 系列）

| 用例ID | 描述 | 优先级 | 验证点 |
|--------|------|--------|--------|
| TC-R01 | 现有 password 登录不受影响（RG-8） | P0 | admin 正常登录 |
| TC-R02 | SUPER_ADMIN 权限放行不变（RG-4） | P0 | admin 访问任意接口不被拦截 |
| TC-R03 | 租户隔离不被破坏（RG-2） | P0 | 不同租户数据互不可见 |
| TC-R04 | sys_opera_log 不受影响（RG-6） | P1 | 原有 sys_opera_log 接口正常（如存在） |
| TC-R05 | 操作日志查询接口向后兼容（RG-1） | P0 | GET /operation-logs 返回 200 |
