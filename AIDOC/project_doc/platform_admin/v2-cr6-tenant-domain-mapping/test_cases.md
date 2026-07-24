# 测试用例：V2-CR6 域名-租户映射管理模块

**关联文档**: requirements.md / design.md
**编写日期**: 2026-07-24
**测试类型**: 接口测试 / 功能测试 / 回归测试 / UI端面测试
**测试方法**: 等价类划分、边界值分析、错误推测、场景法

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | UI | 合计 |
|------|------|------|------|------|-----|------|
| FR-1 CRUD 管理端 | 5 | 6 | 3 | 0 | 4 | 18 |
| FR-2 公开查询接口 | 3 | 2 | 1 | 0 | 0 | 6 |
| FR-3 缓存 | 3 | 0 | 1 | 0 | 0 | 4 |
| FR-5 C端登录页 | 2 | 1 | 1 | 0 | 3 | 7 |
| 回归 | 0 | 0 | 0 | 5 | 0 | 5 |
| **总计** | **13** | **9** | **6** | **5** | **7** | **40** |

---

## 一、正向测试（Happy Path）

### TC-001: 创建域名映射成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 超级管理员已登录，租户 tenant_id 对应的 admin_tenant 存在 |
| **测试步骤** | POST /api/v1/admin/tenant-domains<br>Body: `{"domain":"app.example.com","tenant_id":"{tid}","remark":"测试"}` |
| **预期结果** | HTTP 200，返回 code=0，data 包含 id（string）、domain、tenant_id、version=1 |
| **关联需求** | FR-1 |

### TC-002: 查询域名映射列表（分页+搜索）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已创建多条域名映射记录 |
| **测试步骤** | GET /api/v1/admin/tenant-domains?page=1&page_size=10&domain=example |
| **预期结果** | HTTP 200，返回 list 数组（每项含 tenant_code/tenant_name）、total、page、page_size |
| **关联需求** | FR-1 |

### TC-003: 更新域名映射成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已创建域名映射记录（version=1） |
| **测试步骤** | PUT /api/v1/admin/tenant-domains/:id<br>Body: `{"domain":"new.example.com","version":1}` |
| **预期结果** | HTTP 200，返回更新后的记录，domain 已变更，version=2 |
| **关联需求** | FR-1 |

### TC-004: 删除域名映射成功

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已创建域名映射记录 |
| **测试步骤** | DELETE /api/v1/admin/tenant-domains/:id |
| **预期结果** | HTTP 200，再次查询该 ID 返回 404 或列表中不包含 |
| **关联需求** | FR-1 |

### TC-005: 公开查询接口 — 已配置域名

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 已配置 domain=app.example.com → tenant_code=test_corp |
| **测试步骤** | GET /api/v1/public/tenant-domain?domain=app.example.com |
| **预期结果** | HTTP 200，`{"code":0,"data":{"tenant_code":"test_corp","tenant_name":"...","matched":true}}` |
| **关联需求** | FR-2 |

### TC-006: 公开查询接口 — 未配置域名（fallback default）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 域名 unknown.com 未配置任何映射 |
| **测试步骤** | GET /api/v1/public/tenant-domain?domain=unknown.com |
| **预期结果** | HTTP 200，`{"code":0,"data":{"tenant_code":"default","tenant_name":"默认租户","matched":false}}` |
| **关联需求** | FR-2 |

### TC-007: 公开查询接口 — 无需认证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 不携带任何 Authorization header |
| **测试步骤** | GET /api/v1/public/tenant-domain?domain=anything.com |
| **预期结果** | HTTP 200（不是 401），返回有效响应 |
| **关联需求** | FR-2 |

### TC-008: 缓存命中 — 二次查询不查库

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 已查询过 domain=cached.com |
| **测试步骤** | 再次 GET /api/v1/public/tenant-domain?domain=cached.com |
| **预期结果** | HTTP 200，响应时间显著低于首次查询（< 10ms），结果一致 |
| **关联需求** | FR-3 |

### TC-009: 缓存失效 — 变更后下次查询更新

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | domain=change.com 已缓存为 tenant_a |
| **测试步骤** | 1. 更新 change.com → tenant_b<br>2. GET /api/v1/public/tenant-domain?domain=change.com |
| **预期结果** | 返回 tenant_b（缓存已失效并重新填充） |
| **关联需求** | FR-3 |

### TC-010: 缓存失效 — 删除后查询返回 default

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | domain=deleted.com 已配置且已缓存 |
| **测试步骤** | 1. DELETE 该记录<br>2. GET /api/v1/public/tenant-domain?domain=deleted.com |
| **预期结果** | 返回 matched=false，tenant_code=default |
| **关联需求** | FR-3 |

### TC-011: C端登录页 — 自动解析 tenant_code

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-user 正常启动，后端在线 |
| **测试步骤** | 访问 C端登录页 |
| **预期结果** | 页面不显示租户编码输入框；发送验证码时请求体中 tenant_code 为自动解析的值 |
| **关联需求** | FR-5 |

### TC-012: C端登录页 — 后端不可达时 fallback

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 后端关闭，前端独立运行 |
| **测试步骤** | 访问 C端登录页，等待 3 秒超时 |
| **预期结果** | 按钮恢复可用，使用 VITE_TENANT_CODE（默认 default）作为 tenant_code |
| **关联需求** | FR-5 |

### TC-013: 列表中 tenant_code/tenant_name 正确展示

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 已创建域名映射记录 |
| **测试步骤** | GET /api/v1/admin/tenant-domains |
| **预期结果** | 每条记录含 tenant_code 和 tenant_name（通过 Service 层补充，非 JOIN） |
| **关联需求** | FR-1 |

---

## 二、反向测试（Negative）

### TC-N01: 创建重复域名返回 409

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | domain=dup.example.com 已存在 |
| **测试步骤** | POST /api/v1/admin/tenant-domains Body: `{"domain":"dup.example.com","tenant_id":"{tid}"}` |
| **预期结果** | HTTP 409（或 400），message 含"域名已存在" |
| **关联需求** | FR-1 |

### TC-N02: 创建保留域名返回 400

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试数据** | domain=localhost / 127.0.0.1 / 0.0.0.0 |
| **预期结果** | HTTP 400，message 含"保留域名" |
| **关联需求** | FR-1 |

### TC-N03: 创建超长域名返回 400

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | domain = "a" × 256 个字符 |
| **预期结果** | HTTP 400，message 含"超过 255" |
| **关联需求** | FR-1 |

### TC-N04: 非 SUPER_ADMIN 用户调用 CRUD 返回 403

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 使用租户管理员 token（role_type != SUPER_ADMIN） |
| **测试步骤** | POST /api/v1/admin/tenant-domains |
| **预期结果** | HTTP 403，message 含"超级管理员" |
| **关联需求** | FR-1 |

### TC-N05: 更新时 version 不匹配返回冲突

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 version=2 |
| **测试步骤** | PUT /api/v1/admin/tenant-domains/:id Body: `{"domain":"x.com","version":1}` |
| **预期结果** | HTTP 409（或 400），message 含"已被修改" |
| **关联需求** | FR-1 |

### TC-N06: 公开查询接口缺少 domain 参数返回 400

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | GET /api/v1/public/tenant-domain（无 domain 参数） |
| **预期结果** | HTTP 400，message 含"不能为空" |
| **关联需求** | FR-2 |

### TC-N07: 创建时 tenant_id 对应租户不存在返回 404

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | POST /api/v1/admin/tenant-domains Body: `{"domain":"x.com","tenant_id":"9999999999"}` |
| **预期结果** | HTTP 404（或 400），message 含"租户不存在" |
| **关联需求** | FR-1 |

### TC-N08: 未认证访问 CRUD 接口返回 401

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | GET /api/v1/admin/tenant-domains（无 Authorization header） |
| **预期结果** | HTTP 401 |
| **关联需求** | FR-1 |

### TC-N09: 删除不存在的 ID 返回 404

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | DELETE /api/v1/admin/tenant-domains/9999999999 |
| **预期结果** | HTTP 404，message 含"不存在" |
| **关联需求** | FR-1 |

---

## 三、边界测试（Boundary）

### TC-B01: 域名长度恰好 255 字符

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | domain = "a" × 255 |
| **预期结果** | 创建成功（恰好在边界内） |
| **关联需求** | FR-1 |

### TC-B02: 域名长度 256 字符

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | domain = "a" × 256 |
| **预期结果** | 返回 400 |
| **关联需求** | FR-1 |

### TC-B03: 同一租户绑定多个域名

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | 为同一 tenant_id 创建 domain1、domain2、domain3 |
| **预期结果** | 三条记录全部创建成功（一对多关系） |
| **关联需求** | FR-1 |

### TC-B04: 分页 page=0

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试步骤** | GET /api/v1/admin/tenant-domains?page=0&page_size=10 |
| **预期结果** | 返回第一页数据（page 修正为 1），不报错 |
| **关联需求** | FR-1 |

### TC-B05: 缓存 TTL 到期后重新查库

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | 等待 5 分钟后再次查询 |
| **预期结果** | 查库获取最新值（验证缓存过期后重新填充） |
| **关联需求** | FR-3 |

### TC-B06: C端解析超时 3 秒后 fallback

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 后端延迟 > 3 秒 |
| **预期结果** | 3 秒后自动 fallback，按钮恢复可用 |
| **关联需求** | FR-5 |

---

## 四、回归测试（Regression）

### TC-R01: 管理端登录不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | POST /auth/login 正常返回 token |
| **预期结果** | 行为与 CR-6 实现前完全一致 |

### TC-R02: C端认证不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | POST /api/v1/user/auth/send-code 正常返回 200 |
| **预期结果** | 验证码发送流程正常 |

### TC-R03: TenantIsolationCallback 不影响 tenant_domain

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 普通租户管理员（tenant_id > 0）查询域名列表 |
| **预期结果** | 查询不被 TenantIsolation 拦截（表无 tenant_id 字段，Callback 自动跳过） |

### TC-R04: 其他公开接口不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | POST /api/v1/user/auth/send-code 不带 token |
| **预期结果** | 正常返回 200 或 400（业务校验），不是 404 |

### TC-R05: 现有 LocalCache 模块不受影响

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **关联** | RG-5 |
| **验证步骤** | 执行权限检查相关请求（触发 PermissionCache） |
| **预期结果** | 权限缓存正常命中/失效，不被域名缓存干扰 |

---

## 五、UI 端面测试

### TC-F01: 域名管理页列表加载

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 超级管理员已登录 |
| **测试步骤** | 导航到 /system/tenant-domain |
| **预期结果** | 表格正确展示（域名、租户名称、备注、创建时间、操作），分页组件可见 |

### TC-F02: 域名管理页新增弹窗

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | 1. 点击"新增映射"<br>2. 填写域名和租户ID<br>3. 点击确定 |
| **预期结果** | 弹窗关闭，ElMessage 成功提示，列表刷新含新记录 |

### TC-F03: 域名管理页删除二次确认

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | 1. 点击"删除"<br>2. 确认弹窗出现（含域名）<br>3. 点击确认 |
| **预期结果** | 确认框文案含具体域名，确认后记录消失 |

### TC-F04: C端登录页加载状态

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | 1. 打开 C端登录页<br>2. 观察按钮状态 |
| **预期结果** | 页面加载期间按钮禁用/显示"加载中…"，解析完成后恢复 |

### TC-F05: C端登录页无租户输入框

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试步骤** | 检查登录页表单 |
| **预期结果** | 只有手机号和验证码两个输入框，无租户编码相关字段 |

### TC-F06: C端登录页错误验证码提示

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | 输入手机号 + 错误验证码，点击登录 |
| **预期结果** | ElMessage.error 展示后端错误信息（如"验证码错误"） |

### TC-F07: 域名管理页编辑回填

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试步骤** | 1. 点击某条记录的"编辑"<br>2. 观察弹窗 |
| **预期结果** | 弹窗标题"编辑域名映射"，域名和租户ID回填当前值 |

---

## 执行命令

```powershell
# API 测试
npx playwright test tests/v2-cr6-tenant-domain.api.spec.ts --project=chromium --reporter=list

# UI 测试
npx playwright test tests/v2-cr6-tenant-domain.ui.spec.ts --project=chromium --reporter=list

# 全部
npx playwright test tests/v2-cr6-tenant-domain.*.spec.ts --reporter=html
```
