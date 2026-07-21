# 测试用例：V2-CR2 权限体系增强

**关联文档**: requirements.md / design.md
**编写日期**: 2025-01-20
**测试类型**: 功能测试 / 接口测试 / 回归测试 / 数据验证
**测试方法**: 等价类划分 + 边界值分析 + 状态转换测试 + 错误推测 + 场景法

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | 接口 | 数据验证 | 合计 |
|------|------|------|------|------|------|----------|------|
| Phase 1: 角色继承（子集校验） | 4 | 4 | 3 | 2 | 8 | 3 | 24 |
| Phase 2: 权限集 | 3 | 2 | 2 | 2 | 4 | 2 | 15 |
| Phase 3: 字段权限 | 5 | 4 | 3 | 1 | 12 | 3 | 28 |
| Phase 4: 记录共享 | 4 | 3 | 3 | 2 | 6 | 3 | 21 |
| **总计** | **16** | **13** | **11** | **7** | **30** | **11** | **88** |

---

## 一、正向测试（Happy Path）

### Phase 1: 角色继承

#### TC-001: 子角色分配父角色子集权限 — 资源

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 父角色 A 拥有资源 [R1, R2, R3]；子角色 B（parent_id=A） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:B/resources` body: `{resource_ids: [R1, R2]}` |
| **预期结果** | HTTP 200, 分配成功，子角色 B 拥有 [R1, R2] |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（有效子集） |

#### TC-002: 子角色分配父角色子集权限 — API

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 父角色 A 拥有 API 权限 [AP1, AP2, AP3]；子角色 B（parent_id=A） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:B/apis` body: `{api_permission_ids: [AP1, AP2]}` |
| **预期结果** | HTTP 200, 分配成功 |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（有效子集） |

#### TC-003: 父角色缩减权限触发级联裁剪

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 父角色 A 拥有 [R1,R2,R3,R4]；子角色 B 拥有 [R1,R2,R3]；孙角色 C 拥有 [R1,R2] |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:A/resources` body: `{resource_ids: [R1, R2]}` |
| **预期结果** | HTTP 200, 响应含 `affected_children` 字段，子角色 B 被裁剪为 [R1,R2]，孙角色 C 不变（已在范围内） |
| **关联需求** | FR-2 |
| **设计方法** | 场景法（端到端级联） |

#### TC-004: 查询可分配资源/API 接口

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 角色 A 已绑定资源 [R1, R2, R3] 和 API [AP1, AP2] |
| **测试步骤** | 1. 调用 `GET /api/v1/admin/roles/:A/assignable-resources` <br> 2. 调用 `GET /api/v1/admin/roles/:A/assignable-apis` |
| **预期结果** | 步骤1: HTTP 200, data 包含 [R1, R2, R3] <br> 步骤2: HTTP 200, data 包含 [AP1, AP2] |
| **关联需求** | FR-3 |
| **设计方法** | 等价类（正向查询） |

### Phase 2: 权限集

#### TC-005: 创建 PERMISSION_SET 类型角色

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 管理员已登录，租户上下文有效 |
| **测试步骤** | 1. 调用 `POST /api/v1/admin/roles` body: `{name: "额外导出权限", role_type: "PERMISSION_SET", parent_id: 123}` |
| **预期结果** | HTTP 200, 创建成功，返回角色信息中 parent_id=NULL（传入的 parent_id 被忽略），role_type=PERMISSION_SET |
| **关联需求** | FR-4 |
| **设计方法** | 等价类（有效输入） |

#### TC-006: 用户权限合并计算（普通角色 + 权限集）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 用户 U 拥有普通角色 A（含 user:list, user:view）和权限集 B（含 user:delete, report:export） |
| **测试步骤** | 1. 以用户 U 身份获取权限列表/菜单 |
| **预期结果** | 用户权限包含 [user:list, user:view, user:delete, report:export]（并集去重） |
| **关联需求** | FR-5 |
| **设计方法** | 场景法（权限叠加） |

#### TC-007: 按 role_type 过滤角色列表

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 租户下存在普通角色 3 个、PERMISSION_SET 角色 2 个 |
| **测试步骤** | 1. 调用 `GET /api/v1/admin/roles?role_type=PERMISSION_SET` |
| **预期结果** | HTTP 200, 返回 2 个权限集角色 |
| **关联需求** | FR-4 |
| **设计方法** | 等价类（有效过滤） |

### Phase 3: 字段权限

#### TC-008: 配置字段权限为 HIDDEN

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 角色 A 已存在；字段对象 invoice 已注册，含字段 cost_price |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/field-permissions` body: `{role_id: A, object_code: "invoice", field_name: "cost_price", access: "HIDDEN"}` |
| **预期结果** | HTTP 200, 配置保存成功 |
| **关联需求** | FR-6 |
| **设计方法** | 等价类（有效配置） |

#### TC-009: HIDDEN 字段在 API 响应中被过滤

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 角色 A 对 invoice.cost_price 配置为 HIDDEN；用户 U 仅拥有角色 A |
| **测试步骤** | 1. 以用户 U 身份调用发票列表接口 |
| **预期结果** | 响应 JSON 中不包含 cost_price 字段 |
| **关联需求** | FR-7 |
| **设计方法** | 场景法（字段过滤端到端） |

#### TC-010: 自动注册带 fieldperm tag 的字段对象

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | model struct User 含 `fieldperm:"手机号"` 标签的 phone 字段 |
| **测试步骤** | 1. 系统启动 <br> 2. 调用 `GET /api/v1/admin/field-objects` <br> 3. 调用 `GET /api/v1/admin/field-objects/user/fields` |
| **预期结果** | 步骤2: 返回含 object_code=user 的对象 <br> 步骤3: 返回含 {field_name: "phone", description: "手机号"} 的字段 |
| **关联需求** | FR-8 |
| **设计方法** | 场景法（自动注册流程） |

#### TC-011: 手动注册字段对象和字段

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 管理员已登录 |
| **测试步骤** | 1. 调用 `POST /api/v1/admin/field-objects` body: `{object_code: "customer", object_name: "客户"}` <br> 2. 调用 `POST /api/v1/admin/field-objects/customer/fields` body: `{field_name: "contract_amount", description: "合同金额"}` <br> 3. 调用 `GET /api/v1/admin/field-objects/customer/fields` |
| **预期结果** | 步骤3: 返回含 {field_name: "contract_amount", description: "合同金额", source: "MANUAL"} |
| **关联需求** | FR-8 |
| **设计方法** | 场景法（手动注册流程） |

#### TC-012: 多角色字段权限冲突取最高权限

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 用户 U 拥有角色 A（invoice.cost_price=HIDDEN）和角色 B（invoice.cost_price=VISIBLE） |
| **测试步骤** | 1. 以用户 U 身份调用发票接口 |
| **预期结果** | 响应 JSON 中包含 cost_price 字段（VISIBLE > HIDDEN，取最高权限） |
| **关联需求** | FR-7 |
| **设计方法** | 判定表（多条件组合） |

### Phase 4: 记录共享

#### TC-013: 创建记录共享规则

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 用户 U 拥有发票记录 INV-001 |
| **测试步骤** | 1. 调用 `POST /api/v1/admin/record-shares` body: `{object_code: "invoice", record_id: 1001, share_to_type: "USER", share_to_id: 200, access_level: "READ", expire_at: "2025-12-31T23:59:59Z"}` |
| **预期结果** | HTTP 200, 共享规则创建成功，返回规则 ID |
| **关联需求** | FR-9 |
| **设计方法** | 等价类（有效创建） |

#### TC-014: 共享规则生效 — 用户可访问被共享记录

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-001 被共享给用户 B（share_to_type=USER）；用户 B 原本无权访问该记录 |
| **测试步骤** | 1. 以用户 B 身份查询发票列表 |
| **预期结果** | 查询结果中包含 INV-001 |
| **关联需求** | FR-10 |
| **设计方法** | 场景法（共享生效端到端） |

#### TC-015: 共享给角色 — 拥有该角色的用户可访问

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-002 被共享给角色 R（share_to_type=ROLE）；用户 C 拥有角色 R |
| **测试步骤** | 1. 以用户 C 身份查询发票列表 |
| **预期结果** | 查询结果中包含 INV-002 |
| **关联需求** | FR-10 |
| **设计方法** | 等价类（角色共享） |

#### TC-016: 共享给部门 — 部门成员可访问

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-003 被共享给部门 D（share_to_type=DEPT）；用户 D1 属于部门 D |
| **测试步骤** | 1. 以用户 D1 身份查询发票列表 |
| **预期结果** | 查询结果中包含 INV-003 |
| **关联需求** | FR-10 |
| **设计方法** | 等价类（部门共享） |

---

## 二、反向测试（Negative）

### Phase 1: 角色继承

#### TC-N01: 子角色分配超出父角色范围的资源 — 被拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 父角色 A 拥有资源 [R1, R2]；子角色 B（parent_id=A） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:B/resources` body: `{resource_ids: [R1, R2, R99]}` （R99 不在父角色范围） |
| **预期结果** | HTTP 400, code=ErrExceedsParentPermission |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（无效超集） |

#### TC-N02: 子角色分配超出父角色范围的 API — 被拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 父角色 A 拥有 API [AP1, AP2]；子角色 B（parent_id=A） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:B/apis` body: `{api_permission_ids: [AP1, AP99]}` |
| **预期结果** | HTTP 400, code=ErrExceedsParentPermission |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（无效超集） |

#### TC-N03: 顶级角色分配权限不做子集校验

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 顶级角色 T（parent_id=NULL） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:T/resources` body: `{resource_ids: [R1, R2, R3, R4, R5]}` |
| **预期结果** | HTTP 200, 分配成功（不做子集校验） |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（免校验场景） |

#### TC-N04: PERMISSION_SET 分配权限不做子集校验

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 权限集角色 PS（role_type=PERMISSION_SET） |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/roles/:PS/resources` body: `{resource_ids: [R1, R99, R100]}` |
| **预期结果** | HTTP 200, 分配成功（不做子集校验） |
| **关联需求** | FR-1 |
| **设计方法** | 等价类（免校验场景） |

### Phase 2: 权限集

#### TC-N05: 删除权限集后用户立即失去额外权限

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 用户 U 拥有普通角色 A（含 user:list）和权限集 B（含 report:export） |
| **测试步骤** | 1. 删除权限集 B <br> 2. 以用户 U 身份检查权限 |
| **预期结果** | 用户 U 不再拥有 report:export 权限 |
| **关联需求** | FR-5 |
| **设计方法** | 状态转换（权限集删除） |

#### TC-N06: 创建 PERMISSION_SET 时传入非法 role_type

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **前置条件** | 管理员已登录 |
| **测试步骤** | 1. 调用 `POST /api/v1/admin/roles` body: `{name: "test", role_type: "INVALID_TYPE"}` |
| **预期结果** | HTTP 400, 参数校验失败 |
| **关联需求** | FR-4 |
| **设计方法** | 错误推测（非法枚举值） |

### Phase 3: 字段权限

#### TC-N07: 配置不存在的 object_code 字段权限

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 字段对象 "nonexistent" 未注册 |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/field-permissions` body: `{role_id: 1, object_code: "nonexistent", field_name: "xxx", access: "HIDDEN"}` |
| **预期结果** | HTTP 400, 对象不存在错误 |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测（无效引用） |

#### TC-N08: 配置无效的 access 级别

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 字段对象 invoice 已注册 |
| **测试步骤** | 1. 调用 `PUT /api/v1/admin/field-permissions` body: `{role_id: 1, object_code: "invoice", field_name: "cost_price", access: "INVALID"}` |
| **预期结果** | HTTP 400, access 值无效 |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测（非法枚举值） |

#### TC-N09: 删除字段权限后恢复为默认 EDITABLE

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 角色 A 对 invoice.cost_price 配置为 HIDDEN |
| **测试步骤** | 1. 调用 `DELETE /api/v1/admin/field-permissions/:id` <br> 2. 以角色 A 用户身份调用发票接口 |
| **预期结果** | 步骤2: 响应 JSON 中包含 cost_price 字段（恢复默认 EDITABLE） |
| **关联需求** | FR-6 |
| **设计方法** | 状态转换（删除配置） |

#### TC-N10: 手动注册重复 object_code + field_name 时合并

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 自动注册已存在 user.phone（description="手机号"） |
| **测试步骤** | 1. 调用 `POST /api/v1/admin/field-objects/user/fields` body: `{field_name: "phone", description: "联系电话"}` <br> 2. 查询 user 的字段列表 |
| **预期结果** | phone 字段 description 更新为"联系电话"（手动设置优先） |
| **关联需求** | FR-8 |
| **设计方法** | 错误推测（冲突合并） |

### Phase 4: 记录共享

#### TC-N11: 共享规则过期后用户无法访问

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-001 共享给用户 B，expire_at 设为过去时间 |
| **测试步骤** | 1. 以用户 B 身份查询发票列表 |
| **预期结果** | 查询结果中不包含 INV-001 |
| **关联需求** | FR-10 |
| **设计方法** | 状态转换（过期失效） |

#### TC-N12: 删除共享规则后目标用户立即失去访问权

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-001 已共享给用户 B（有效规则） |
| **测试步骤** | 1. 调用 `DELETE /api/v1/admin/record-shares/:id` <br> 2. 以用户 B 身份查询发票列表 |
| **预期结果** | 步骤2: 查询结果中不包含 INV-001 |
| **关联需求** | FR-9 |
| **设计方法** | 状态转换（删除规则） |

#### TC-N13: access_level=READ 时用户不能编辑

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 记录 INV-001 以 access_level=READ 共享给用户 B |
| **测试步骤** | 1. 以用户 B 身份尝试更新 INV-001 |
| **预期结果** | HTTP 403, 权限不足 |
| **关联需求** | FR-10 |
| **设计方法** | 等价类（只读权限） |

---

## 三、边界测试（Boundary）

### Phase 1: 角色继承

#### TC-B01: 子角色权限为父角色完整子集（全等）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 父角色拥有 [R1, R2, R3]；子角色分配 [R1, R2, R3]（完全相等） |
| **预期结果** | 分配成功（子集包含全等情况） |
| **设计方法** | 边界值分析 |
| **关联需求** | FR-1 |

#### TC-B02: 子角色分配空权限集

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 父角色拥有 [R1, R2, R3]；子角色分配 []（空数组） |
| **预期结果** | 分配成功，子角色权限被清空 |
| **设计方法** | 边界值分析（最小边界） |
| **关联需求** | FR-1 |

#### TC-B03: 多级级联裁剪（3 层以上）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 角色层级：A → B → C → D（4 层），A 缩减权限 |
| **预期结果** | B、C、D 均被递归裁剪，直至叶子角色 D |
| **设计方法** | 边界值分析（最大递归深度） |
| **关联需求** | FR-2 |

### Phase 2: 权限集

#### TC-B04: 用户同时拥有多个权限集（权限去重）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 用户拥有普通角色 A（含 user:list）、权限集 B（含 user:list, report:view）、权限集 C（含 report:view, report:export） |
| **预期结果** | 最终权限为 [user:list, report:view, report:export]（去重后） |
| **设计方法** | 边界值分析（多重叠加） |
| **关联需求** | FR-5 |

#### TC-B05: 用户无任何权限集时权限计算不变

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 用户仅拥有普通角色 A（含 user:list, user:view），无权限集 |
| **预期结果** | 最终权限 = [user:list, user:view]，与 CR-1 行为一致 |
| **设计方法** | 边界值分析（零权限集） |
| **关联需求** | FR-5, RG-7 |

### Phase 3: 字段权限

#### TC-B06: 未配置字段权限的对象不做过滤

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试数据** | 对象 "tenant" 未配置任何字段权限 |
| **预期结果** | 租户接口响应 JSON 包含所有字段（无过滤） |
| **设计方法** | 边界值分析（零配置） |
| **关联需求** | FR-7 |

#### TC-B07: 所有字段均配置为 HIDDEN

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | 角色 A 对 invoice 所有已注册字段均配置为 HIDDEN |
| **预期结果** | 响应 JSON 中仅保留未注册到字段权限系统的字段（如 id、created_at 等） |
| **设计方法** | 边界值分析（全隐藏） |
| **关联需求** | FR-7 |

#### TC-B08: fieldperm:"-" 标记的 struct 不注册

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | model AdminRole struct 标注 `fieldperm:"-"` |
| **预期结果** | 系统启动后 field-objects 列表中不包含 admin_role 对象 |
| **设计方法** | 边界值分析（排除边界） |
| **关联需求** | FR-8 |

### Phase 4: 记录共享

#### TC-B09: 共享规则 expire_at=NULL 永久生效

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 创建共享规则时 expire_at 不传或为 null |
| **预期结果** | 共享规则永久有效，不会自动失效 |
| **设计方法** | 边界值分析（NULL 永久） |
| **关联需求** | FR-9 |

#### TC-B10: 共享规则 expire_at 恰好等于当前时间

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **测试数据** | 共享规则 expire_at = NOW()（精确到秒） |
| **预期结果** | 规则已失效（expire_at < NOW() 或 expire_at <= NOW() 取决于实现，验证边界行为） |
| **设计方法** | 边界值分析（临界时间点） |
| **关联需求** | FR-10 |

#### TC-B11: 同一记录被共享给多种类型（USER + ROLE + DEPT）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | INV-001 同时共享给 USER:100, ROLE:200, DEPT:300 |
| **预期结果** | 三种共享规则均独立生效，符合各自条件的用户均可访问 |
| **设计方法** | 边界值分析（多规则叠加） |
| **关联需求** | FR-10 |

---

## 四、回归测试（Regression）

#### TC-R01: 顶级角色 AssignResources 行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | 1. 对顶级角色（parent_id=NULL）调用 `PUT /roles/:id/resources` 分配任意资源 |
| **预期结果** | 分配成功，无子集校验，无级联裁剪逻辑触发，行为与 CR-1 一致 |

#### TC-R02: 顶级角色 AssignApis 行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | 1. 对顶级角色调用 `PUT /roles/:id/apis` 分配任意 API |
| **预期结果** | 分配成功，无子集校验，行为与 CR-1 一致 |

#### TC-R03: 未配置字段权限时 GetUserMenu 响应不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. 确保无任何 FieldPermission 配置 <br> 2. 调用 GetUserMenu 接口 |
| **预期结果** | 响应 JSON 字段完整，与 CR-1 输出完全一致 |

#### TC-R04: 无记录共享时 DataScopeCallback 行为不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | 1. 确保无 record_share 记录 <br> 2. 查询业务列表 |
| **预期结果** | WHERE 条件中无共享子查询注入，查询结果与 CR-1 一致 |

#### TC-R05: 普通角色 CRUD 不受 PERMISSION_SET 影响

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-5 |
| **验证步骤** | 1. 创建/编辑/删除普通角色（role_type=NORMAL） |
| **预期结果** | 操作正常，不受 PERMISSION_SET 类型存在的影响 |

#### TC-R06: 用户-角色分配流程不变

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-6 |
| **验证步骤** | 1. 为用户分配普通角色 |
| **预期结果** | 分配成功，无额外校验逻辑 |

#### TC-R07: 无权限集时用户权限计算与 CR-1 一致

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-7 |
| **验证步骤** | 1. 用户仅有普通角色，无 PERMISSION_SET 角色 <br> 2. 获取用户权限列表 |
| **预期结果** | 权限集合与 CR-1 计算结果完全一致 |

---

## 五、接口测试（API Level）

### Phase 1: 角色继承接口

#### TC-A01: PUT /roles/:id/resources — 正向（子角色子集分配）

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/roles/100/resources` |
| **Headers** | Authorization: Bearer {valid_token}, X-Tenant-ID: 1 |
| **Body** | ```json {"resource_ids": [1, 2, 3]}``` |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"affected_children": []}}``` |
| **关联需求** | FR-1, FR-2 |

#### TC-A02: PUT /roles/:id/resources — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/roles/100/resources` |
| **Headers** | 无 Authorization header |
| **Body** | ```json {"resource_ids": [1, 2, 3]}``` |
| **预期响应** | HTTP 401, ```json {"code": 40101, "message": "未认证"}``` |

#### TC-A03: PUT /roles/:id/resources — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/roles/100/resources` |
| **Headers** | Authorization: Bearer {low_privilege_token} |
| **Body** | ```json {"resource_ids": [1, 2, 3]}``` |
| **预期响应** | HTTP 403, ```json {"code": 40301, "message": "权限不足"}``` |

#### TC-A04: PUT /roles/:id/resources — 超出父角色范围

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/roles/100/resources`（角色 100 为子角色） |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"resource_ids": [1, 2, 999]}```（999 不在父角色范围） |
| **预期响应** | HTTP 400, ```json {"code": 40001, "message": "ErrExceedsParentPermission"}``` |
| **关联需求** | FR-1 |

#### TC-A05: PUT /roles/:id/apis — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/roles/100/apis` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"api_permission_ids": [10, 20]}``` |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"affected_children": []}}``` |
| **关联需求** | FR-1 |

#### TC-A06: GET /roles/:id/assignable-resources — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/roles/100/assignable-resources` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"resource_ids": [1, 2, 3, ...]}}``` |
| **关联需求** | FR-3 |

#### TC-A07: GET /roles/:id/assignable-apis — 正向

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/roles/100/assignable-apis` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"api_permission_ids": [10, 20, ...]}}``` |
| **关联需求** | FR-3 |

#### TC-A08: GET /roles/:id/assignable-resources — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/roles/100/assignable-resources` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |

### Phase 2: 权限集接口

#### TC-A09: POST /roles — 创建 PERMISSION_SET

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/roles` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"name": "导出权限集", "role_type": "PERMISSION_SET", "description": "额外导出权限"}``` |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"id": "...", "role_type": "PERMISSION_SET", "parent_id": null}}``` |
| **关联需求** | FR-4 |

#### TC-A10: GET /roles?role_type=PERMISSION_SET — 过滤

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/roles?role_type=PERMISSION_SET` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, 返回列表中所有角色 role_type 均为 PERMISSION_SET |
| **关联需求** | FR-4 |

#### TC-A11: POST /roles — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/roles` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |

#### TC-A12: GET /roles?role_type=PERMISSION_SET — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/roles?role_type=PERMISSION_SET` |
| **Headers** | Authorization: Bearer {low_privilege_token} |
| **预期响应** | HTTP 403 |

### Phase 3: 字段权限接口

#### TC-A13: GET /field-objects — 获取字段对象列表

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/field-objects` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": [{"object_code": "user", "object_name": "用户", "source": "AUTO"}, ...]}``` |
| **关联需求** | FR-8 |

#### TC-A14: GET /field-objects/:objectCode/fields — 获取字段列表

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/field-objects/user/fields` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": [{"field_name": "phone", "description": "手机号", "source": "AUTO"}, ...]}``` |
| **关联需求** | FR-8 |

#### TC-A15: PUT /field-objects/:objectCode/fields/:fieldName — 修改字段描述

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/field-objects/user/fields/phone` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"description": "联系电话"}``` |
| **预期响应** | HTTP 200, 修改成功 |
| **关联需求** | FR-8 |

#### TC-A16: POST /field-objects — 手动注册对象

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/field-objects` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"object_code": "customer", "object_name": "客户", "app_code": "crm"}``` |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"id": "...", "source": "MANUAL"}}``` |
| **关联需求** | FR-8 |

#### TC-A17: POST /field-objects/:objectCode/fields — 手动注册字段

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/field-objects/customer/fields` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"field_name": "contract_amount", "description": "合同金额"}``` |
| **预期响应** | HTTP 200 |
| **关联需求** | FR-8 |

#### TC-A18: GET /field-permissions — 查询角色字段权限

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/field-permissions?role_id=100&object_code=invoice` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": [{"field_name": "cost_price", "access": "HIDDEN"}, ...]}``` |
| **关联需求** | FR-6 |

#### TC-A19: PUT /field-permissions — 批量设置字段权限

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/field-permissions` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"role_id": 100, "object_code": "invoice", "permissions": [{"field_name": "cost_price", "access": "HIDDEN"}, {"field_name": "profit", "access": "VISIBLE"}]}``` |
| **预期响应** | HTTP 200 |
| **关联需求** | FR-6 |

#### TC-A20: DELETE /field-permissions/:id — 删除字段权限

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/field-permissions/500` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, 删除成功 |
| **关联需求** | FR-6 |

#### TC-A21: GET /field-objects — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/field-objects` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |

#### TC-A22: PUT /field-permissions — 无权限

| 字段 | 内容 |
|------|------|
| **请求** | `PUT /api/v1/admin/field-permissions` |
| **Headers** | Authorization: Bearer {low_privilege_token} |
| **预期响应** | HTTP 403 |

#### TC-A23: POST /field-objects — 重复 object_code

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/field-objects` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"object_code": "user", "object_name": "用户重复"}```（已存在） |
| **预期响应** | HTTP 400 或 409, object_code 已存在 |
| **关联需求** | FR-8 |

#### TC-A24: DELETE /field-permissions/:id — 不存在的 ID

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/field-permissions/999999` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 404, 记录不存在 |

### Phase 4: 记录共享接口

#### TC-A25: POST /record-shares — 创建共享规则

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/record-shares` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"object_code": "invoice", "record_id": 1001, "share_to_type": "USER", "share_to_id": 200, "access_level": "READ", "expire_at": "2025-12-31T23:59:59Z"}``` |
| **预期响应** | HTTP 200, ```json {"code": 0, "data": {"id": "..."}}``` |
| **关联需求** | FR-9 |

#### TC-A26: GET /record-shares — 查询共享规则列表

| 字段 | 内容 |
|------|------|
| **请求** | `GET /api/v1/admin/record-shares?object_code=invoice&record_id=1001` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, 返回该记录的所有有效共享规则列表 |
| **关联需求** | FR-9 |

#### TC-A27: DELETE /record-shares/:id — 删除共享规则

| 字段 | 内容 |
|------|------|
| **请求** | `DELETE /api/v1/admin/record-shares/300` |
| **Headers** | Authorization: Bearer {valid_token} |
| **预期响应** | HTTP 200, 删除成功 |
| **关联需求** | FR-9 |

#### TC-A28: POST /record-shares — 未认证

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/record-shares` |
| **Headers** | 无 Authorization |
| **预期响应** | HTTP 401 |

#### TC-A29: POST /record-shares — 无效 share_to_type

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/record-shares` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"object_code": "invoice", "record_id": 1001, "share_to_type": "INVALID", "share_to_id": 200, "access_level": "READ"}``` |
| **预期响应** | HTTP 400, share_to_type 无效 |
| **关联需求** | FR-9 |

#### TC-A30: POST /record-shares — 无效 access_level

| 字段 | 内容 |
|------|------|
| **请求** | `POST /api/v1/admin/record-shares` |
| **Headers** | Authorization: Bearer {valid_token} |
| **Body** | ```json {"object_code": "invoice", "record_id": 1001, "share_to_type": "USER", "share_to_id": 200, "access_level": "ADMIN"}``` |
| **预期响应** | HTTP 400, access_level 无效（仅支持 READ/EDIT） |
| **关联需求** | FR-9 |

---

## 六、数据验证

### Phase 1: 级联裁剪数据库状态

#### TC-D01: 级联裁剪后子角色数据库状态验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 父角色 A 权限从 [R1,R2,R3,R4] 缩减为 [R1,R2]；子角色 B 原有 [R1,R2,R3] |
| **验证 SQL** | ```sql SELECT resource_id FROM admin_role_resource WHERE role_id = :B_id``` |
| **预期结果** | 仅返回 [R1, R2]，R3 已被裁剪删除 |
| **关联需求** | FR-2 |

#### TC-D02: 级联裁剪不影响未超出范围的子角色

| 字段 | 内容 |
|------|------|
| **触发操作** | 父角色 A 权限从 [R1,R2,R3,R4] 缩减为 [R1,R2,R3]；子角色 C 原有 [R1,R2] |
| **验证 SQL** | ```sql SELECT resource_id FROM admin_role_resource WHERE role_id = :C_id``` |
| **预期结果** | 仍返回 [R1, R2]，未被修改 |
| **关联需求** | FR-2 |

#### TC-D03: API 权限级联裁剪数据库状态

| 字段 | 内容 |
|------|------|
| **触发操作** | 父角色 A API 权限从 [AP1,AP2,AP3] 缩减为 [AP1]；子角色 B 原有 [AP1,AP2] |
| **验证 SQL** | ```sql SELECT api_permission_id FROM admin_role_api WHERE role_id = :B_id``` |
| **预期结果** | 仅返回 [AP1]，AP2 已被裁剪 |
| **关联需求** | FR-2 |

### Phase 2: 权限集数据验证

#### TC-D04: PERMISSION_SET 创建后 parent_id 为 NULL

| 字段 | 内容 |
|------|------|
| **触发操作** | 创建 PERMISSION_SET 角色（即使请求中传入 parent_id） |
| **验证 SQL** | ```sql SELECT parent_id, role_type FROM admin_role WHERE id = :new_role_id``` |
| **预期结果** | parent_id=NULL, role_type='PERMISSION_SET' |
| **关联需求** | FR-4 |

#### TC-D05: 权限集删除后用户角色关联清除

| 字段 | 内容 |
|------|------|
| **触发操作** | 删除权限集角色 PS |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM admin_user_role WHERE role_id = :PS_id``` |
| **预期结果** | 返回 0（关联关系已清除） |
| **关联需求** | FR-5 |

### Phase 3: 字段过滤后 JSON 结构验证

#### TC-D06: HIDDEN 字段从 JSON 响应中完全移除

| 字段 | 内容 |
|------|------|
| **触发操作** | 角色 A 配置 invoice.cost_price=HIDDEN；以角色 A 用户调用 `GET /invoices` |
| **验证方式** | 解析响应 JSON，遍历 data 数组中每个对象 |
| **预期结果** | 所有对象均不包含 "cost_price" key（非 null，是完全不存在） |
| **关联需求** | FR-7 |

#### TC-D07: VISIBLE 字段在 JSON 响应中保留

| 字段 | 内容 |
|------|------|
| **触发操作** | 角色 A 配置 invoice.amount=VISIBLE；以角色 A 用户调用 `GET /invoices` |
| **验证方式** | 解析响应 JSON |
| **预期结果** | 所有对象均包含 "amount" key 且有值 |
| **关联需求** | FR-7 |

#### TC-D08: 自动注册数据库状态验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 系统启动（含带 fieldperm tag 的 User model） |
| **验证 SQL** | ```sql SELECT * FROM admin_field_object WHERE object_code = 'user'``` <br> ```sql SELECT * FROM admin_field_definition WHERE object_code = 'user'``` |
| **预期结果** | admin_field_object 存在 object_code=user 记录；admin_field_definition 包含 phone/email/real_name 等带 fieldperm tag 的字段 |
| **关联需求** | FR-8 |

### Phase 4: 记录共享数据验证

#### TC-D09: 共享规则过期后查询排除验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 创建共享规则 expire_at 设为 1 小时前的时间 |
| **验证 SQL** | ```sql SELECT record_id FROM admin_record_share WHERE expire_at IS NULL OR expire_at > NOW()``` |
| **预期结果** | 过期规则不在有效规则集中 |
| **关联需求** | FR-10 |

#### TC-D10: DataScope OR 注入后的查询条件验证

| 字段 | 内容 |
|------|------|
| **触发操作** | 用户 B 被共享记录 1001；以用户 B 查询发票列表 |
| **验证方式** | 开启 SQL 日志，检查生成的 WHERE 条件 |
| **预期结果** | SQL 中包含 `OR id IN (SELECT record_id FROM admin_record_share WHERE ...)` 子查询 |
| **关联需求** | FR-10 |

#### TC-D11: 删除共享规则后数据库状态

| 字段 | 内容 |
|------|------|
| **触发操作** | 调用 `DELETE /record-shares/:id` |
| **验证 SQL** | ```sql SELECT COUNT(*) FROM admin_record_share WHERE id = :deleted_id``` |
| **预期结果** | 返回 0（物理删除） |
| **关联需求** | FR-9 |

---

## 执行结果记录（测试执行时填写）

| 用例编号 | 结果 | 执行人 | 日期 | 备注 |
|----------|------|--------|------|------|
| TC-001 | ⬜ | | | 前置数据不足（跳过） |
| TC-002 | ⬜ | | | 前置数据不足（跳过） |
| TC-003 | ⬜ | | | 前置数据不足（跳过） |
| TC-004 | ✅ | Playwright | 2026-07-21 | assignable 接口正常 |
| TC-005 | ✅ | Playwright | 2026-07-21 | PERMISSION_SET 角色创建成功 |
| TC-006 | ⬜ | | | 需要用户权限合并场景，未覆盖 |
| TC-007 | ✅ | Playwright | 2026-07-21 | role_type 过滤正常 |
| TC-008 | ✅ | Playwright | 2026-07-21 | 字段权限 HIDDEN 配置成功 |
| TC-009 | ✅ | Playwright | 2026-07-21 | 查询字段权限配置正常 |
| TC-010 | ✅ | Playwright | 2026-07-21 | user 字段对象已自动注册 |
| TC-011 | ✅ | Playwright | 2026-07-21 | 手动注册对象和字段成功 |
| TC-012 | ⬜ | | | 多角色合并场景，未覆盖 |
| TC-013 | ✅ | Playwright | 2026-07-21 | 共享规则创建成功 |
| TC-014 | ⬜ | | | 需要业务数据，未覆盖 |
| TC-015 | ⬜ | | | 需要业务数据，未覆盖 |
| TC-016 | ⬜ | | | 需要业务数据，未覆盖 |
| TC-N01 | ⬜ | | | 前置数据不足（跳过） |
| TC-N02 | ⬜ | | | 前置数据不足（跳过） |
| TC-N03 | ⬜ | | | 前置数据不足（跳过） |
| TC-N04 | ⬜ | | | 前置数据不足（跳过） |
| TC-N05 | ⬜ | | | 需要用户角色场景，未覆盖 |
| TC-N06 | ❌ | Playwright | 2026-07-21 | BUG: 后端未校验 role_type 枚举，非法值被接受 |
| TC-N07 | ❌ | Playwright | 2026-07-21 | BUG: 后端未校验 object_code 是否存在 |
| TC-N08 | ❌ | Playwright | 2026-07-21 | BUG: 后端未校验 access 枚举值 |
| TC-N09 | ❌ | Playwright | 2026-07-21 | BUG: DELETE /field-permissions/:id 返回 500 |
| TC-N10 | ❌ | Playwright | 2026-07-21 | BUG: 重复字段注册返回 500 而非合并 |
| TC-N11 | ❌ | Playwright | 2026-07-21 | BUG: ListByRecord 未过滤 expire_at |
| TC-N12 | ❌ | Playwright | 2026-07-21 | BUG: DELETE /record-shares/:id 返回 500 |
| TC-N13 | ⬜ | | | 需要业务权限场景，未覆盖 |
| TC-B01 | ⬜ | | | 前置数据不足（跳过） |
| TC-B02 | ❌ | Playwright | 2026-07-21 | BUG: resource_ids binding:"required" 拒绝空数组 |
| TC-B03 | ⬜ | | | 前置数据不足（跳过） |
| TC-B04 | ⬜ | | | 需要多权限集场景，未覆盖 |
| TC-B05 | ⬜ | | | 需要权限集场景，未覆盖 |
| TC-B06 | ✅ | Playwright | 2026-07-21 | 未配置字段权限时接口正常 |
| TC-B07 | ⬜ | | | 需要业务接口数据，未覆盖 |
| TC-B08 | ⬜ | | | 需要 model 标注，未覆盖 |
| TC-B09 | ✅ | Playwright | 2026-07-21 | expire_at=null 共享规则创建成功 |
| TC-B10 | ⬜ | | | 边界时间需要精确控制，未覆盖 |
| TC-B11 | ⬜ | | | 需要多类型共享场景，未覆盖 |
| TC-R01 | ❌ | Playwright | 2026-07-21 | BUG: resource_ids binding:"required" 拒绝空数组 |
| TC-R02 | ❌ | Playwright | 2026-07-21 | BUG: api_permission_ids binding 拒绝空数组 |
| TC-R03 | ✅ | Playwright | 2026-07-21 | GetUserMenu 正常 |
| TC-R04 | ⬜ | | | 需要 SQL 日志验证，未覆盖 |
| TC-R05 | ✅ | Playwright | 2026-07-21 | 普通角色 CRUD 正常 |
| TC-R06 | ⬜ | | | 需要用户角色分配场景，未覆盖 |
| TC-R07 | ✅ | Playwright | 2026-07-21 | 字段对象接口正常 |
| TC-A01 | ✅ | Playwright | 2026-07-21 | PUT /roles/:id/resources 正向通过 |
| TC-A02 | ✅ | Playwright | 2026-07-21 | 未认证返回 401 |
| TC-A03 | ⬜ | | | 需要低权限账号，未覆盖 |
| TC-A04 | ⬜ | | | 前置数据不足（跳过） |
| TC-A05 | ❌ | Playwright | 2026-07-21 | BUG: assignable-resources 对 SUPER_ADMIN 返回 400 |
| TC-A06 | ❌ | Playwright | 2026-07-21 | BUG: 同 A05 |
| TC-A07 | ❌ | Playwright | 2026-07-21 | BUG: 同 A05 |
| TC-A08 | ✅ | Playwright | 2026-07-21 | 未认证返回 401 |
| TC-A09 | ✅ | Playwright | 2026-07-21 | 创建 PERMISSION_SET 成功 |
| TC-A10 | ✅ | Playwright | 2026-07-21 | PERMISSION_SET 过滤正常 |
| TC-A11 | ✅ | Playwright | 2026-07-21 | 未认证返回 401 |
| TC-A12 | ⬜ | | | 需要低权限账号，未覆盖 |
| TC-A13 | ✅ | Playwright | 2026-07-21 | GET /field-objects 正常 |
| TC-A14 | ✅ | Playwright | 2026-07-21 | GET /field-objects/user/fields 正常 |
| TC-A15 | ✅ | Playwright | 2026-07-21 | 修改字段描述成功 |
| TC-A16 | ✅ | Playwright | 2026-07-21 | 手动注册对象成功 |
| TC-A17 | ✅ | Playwright | 2026-07-21 | 手动注册字段成功 |
| TC-A18 | ✅ | Playwright | 2026-07-21 | 查询字段权限正常 |
| TC-A19 | ✅ | Playwright | 2026-07-21 | 批量设置字段权限成功 |
| TC-A20 | ❌ | Playwright | 2026-07-21 | BUG: DELETE /field-permissions/:id 返回 500 |
| TC-A21 | ✅ | Playwright | 2026-07-21 | 未认证返回 401 |
| TC-A22 | ⬜ | | | 需要低权限账号，未覆盖 |
| TC-A23 | ❌ | Playwright | 2026-07-21 | BUG: 重复 object_code 返回 500 而非 400/409 |
| TC-A24 | ❌ | Playwright | 2026-07-21 | BUG: 删除不存在 ID 返回 500 而非 404 |
| TC-A25 | ✅ | Playwright | 2026-07-21 | 创建共享规则成功 |
| TC-A26 | ✅ | Playwright | 2026-07-21 | 查询共享规则列表正常 |
| TC-A27 | ❌ | Playwright | 2026-07-21 | BUG: DELETE /record-shares/:id 返回 500 |
| TC-A28 | ✅ | Playwright | 2026-07-21 | 未认证返回 401 |
| TC-A29 | ✅ | Playwright | 2026-07-21 | 无效 share_to_type 返回 400 |
| TC-A30 | ✅ | Playwright | 2026-07-21 | 无效 access_level 返回 400 |
| TC-D01 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D02 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D03 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D04 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D05 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D06 | ⬜ | | | 需要业务接口数据，未覆盖 |
| TC-D07 | ⬜ | | | 需要业务接口数据，未覆盖 |
| TC-D08 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D09 | ⬜ | | | 需要 SQL 验证，未覆盖 |
| TC-D10 | ⬜ | | | 需要 SQL 日志，未覆盖 |
| TC-D11 | ⬜ | | | 需要 SQL 验证，未覆盖 |
