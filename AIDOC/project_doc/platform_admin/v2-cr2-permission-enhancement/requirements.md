# 需求：V2-CR2 权限体系增强

## 背景

Platform Admin V2 的 CR-1 建立了统一应用模型和中间件链。CR-2 在此基础上增强 RBAC 权限体系，引入角色继承约束、权限集叠加、字段级权限、记录共享四大能力，满足企业级多租户 SaaS 场景下复杂的权限管控需求。

## 用户故事

- 作为**租户管理员**，我希望创建子角色时其权限自动限定在父角色范围内，以便实现权限的层级管控
- 作为**租户管理员**，我希望通过"权限集"为个别用户叠加额外权限，以便无需为每种组合创建独立角色
- 作为**租户管理员**，我希望配置字段级权限，以便不同角色看到同一页面但隐藏敏感字段
- 作为**业务操作员**，我希望将某条记录共享给特定同事/角色/部门，以便跨部门协作

## 功能需求

### Phase 1: 角色继承 + 级联裁剪

#### FR-1: 父角色子集校验（写时校验）

**描述：** AssignResources 和 AssignApis 操作时，如果目标角色有 parent_id，校验分配的权限集是否为父角色已有权限的子集。

**验收标准：**
- WHEN 子角色（parent_id 非空）被分配 resourceIDs 超出父角色已有范围 THEN 系统 SHALL 返回错误 `ErrExceedsParentPermission`
- WHEN 子角色被分配 apiPermissionIDs 超出父角色已有范围 THEN 系统 SHALL 返回错误 `ErrExceedsParentPermission`
- WHEN 顶级角色（parent_id=NULL）被分配权限 THEN 系统 SHALL 不做子集校验，正常执行
- WHEN PERMISSION_SET 类型角色被分配权限 THEN 系统 SHALL 不做子集校验（权限集不受继承约束）

#### FR-2: 级联裁剪（cascadeTrimChildren）

**描述：** 父角色权限缩减时，递归裁剪所有子角色中超出新范围的权限。

**验收标准：**
- WHEN 父角色的 resourceIDs 被缩减 THEN 系统 SHALL 递归裁剪所有子角色中超出范围的 resource 绑定
- WHEN 父角色的 apiPermissionIDs 被缩减 THEN 系统 SHALL 递归裁剪所有子角色中超出范围的 api 绑定
- WHEN 裁剪涉及多级子角色（孙角色） THEN 系统 SHALL 递归处理直至叶子角色
- WHEN 子角色权限未超出父角色新范围 THEN 系统 SHALL 不修改该子角色

#### FR-3: 可分配权限查询接口

**描述：** 提供接口返回某角色可分配给子角色的权限范围（即自身已有权限）。

**验收标准：**
- WHEN 请求 `GET /roles/:id/assignable-resources` THEN 系统 SHALL 返回该角色已绑定的资源 ID 列表
- WHEN 请求 `GET /roles/:id/assignable-apis` THEN 系统 SHALL 返回该角色已绑定的 API 权限 ID 列表
- WHEN 角色为顶级（parent_id=NULL）THEN 系统 SHALL 返回租户已订阅应用范围内的全部资源/API

### Phase 2: 权限集（Permission Set）

#### FR-4: PERMISSION_SET 角色类型支持

**描述：** 新增 `PERMISSION_SET` role_type，创建/管理方式与普通角色一致，但不受 parent_id 约束。

**验收标准：**
- WHEN 创建角色时 role_type=PERMISSION_SET THEN 系统 SHALL 强制 parent_id=NULL（即使传入也忽略），不做继承校验
- WHEN 分配权限集角色给用户 THEN 系统 SHALL 在用户-角色分配界面同时展示普通角色和权限集
- WHEN 计算用户权限 THEN 系统 SHALL 合并所有普通角色权限 ∪ 所有权限集权限（去重）

#### FR-5: 权限计算合并

**描述：** 用户最终权限 = 普通角色权限集合 ∪ PERMISSION_SET 权限集合。

**验收标准：**
- WHEN 用户同时拥有普通角色 A（含 user:list）和权限集 B（含 user:delete）THEN 系统 SHALL 判定用户拥有 user:list + user:delete
- WHEN 用户菜单计算时 THEN 系统 SHALL 合并普通角色 + 权限集关联的 resourceIDs
- WHEN 权限集被删除 THEN 系统 SHALL 用户实时失去该权限集带来的额外权限

### Phase 3: 字段级权限

#### FR-6: 字段权限 CRUD

**描述：** 管理员可为角色配置对特定业务对象的字段访问级别（VISIBLE/EDITABLE/HIDDEN）。

**验收标准：**
- WHEN 管理员配置角色 A 对 object_code=invoice 的 cost_price 字段为 HIDDEN THEN 系统 SHALL 保存配置
- WHEN 管理员查询某角色的字段权限配置 THEN 系统 SHALL 返回该角色所有已配置的字段权限列表
- WHEN 管理员删除字段权限配置 THEN 系统 SHALL 恢复为默认（EDITABLE）

#### FR-7: 字段权限过滤（FieldFilter）

**描述：** API 响应时按角色的字段权限配置过滤字段，HIDDEN 字段从 JSON 中移除。

**验收标准：**
- WHEN 角色配置 object_code=invoice 的 cost_price 为 HIDDEN THEN 该角色用户调用发票接口时响应 JSON 中 SHALL 不包含 cost_price 字段
- WHEN 字段配置为 VISIBLE THEN 响应中 SHALL 包含该字段（前端只读展示）
- WHEN 未配置字段权限的对象 THEN 系统 SHALL 不做任何过滤（全部 EDITABLE）
- WHEN 用户有多个角色且配置冲突 THEN 系统 SHALL 取最高权限（EDITABLE > VISIBLE > HIDDEN）

#### FR-8: 字段对象注册（自动发现 + 描述注解 + 跳过控制）

**描述：** 通过自定义 struct tag（`fieldperm:"描述"`）自动注册可配置字段的业务对象和字段列表。系统启动时扫描注册的 model，提取 json tag 作为 field_name（key）、fieldperm tag 作为字段描述（label）。不需要字段权限控制的对象通过 `fieldperm:"-"` 跳过注册。前端配置界面以 `描述(key)` 格式展示。

**对象注册策略：**

| 对象 | 是否注册 | 理由 |
|------|---------|------|
| admin_user | ✅ 注册 | 手机号/邮箱/真实姓名等个人信息需差异化展示 |
| admin_tenant | ✅ 注册 | 联系方式/合同/配额等可对不同角色隐藏 |
| admin_application | ✅ 注册 | app_config 等内部配置信息可隐藏 |
| 插件业务表（如 invoice、customer） | ✅ 注册 | 典型字段权限场景（成本价、合同金额等） |
| admin_role | ❌ 跳过 | 权限配置对象，隐藏字段无业务意义 |
| admin_resource | ❌ 跳过 | 菜单树结构定义，非业务数据 |
| admin_api_permission | ❌ 跳过 | API 权限定义，非业务数据 |
| admin_role_resource / role_api / role_app | ❌ 跳过 | 关联表，无独立展示场景 |
| admin_user_role / user_tenant | ❌ 跳过 | 关联表 |
| admin_tenant_app | ❌ 跳过 | 关联表 |
| admin_data_scope / data_scope_config | ❌ 跳过 | 权限配置对象 |
| sys_config | ❌ 跳过 | 系统配置，非业务数据 |
| admin_login_log / operation_log | ❌ 跳过 | 日志表，审计用途不应隐藏字段 |

**跳过机制：** 在 model struct 上添加 `fieldperm:"-"` 注解（struct 级别），自动注册时跳过整个对象。未标注 `fieldperm:"-"` 且含有 `fieldperm` tag 字段的 model 才会被注册。

**注册方式：** 支持自动注册 + 手动注册两种方式并存：
- 自动注册：通过 struct tag 扫描，适合内置 model
- 手动注册：通过管理页面手动添加对象和字段定义，适合插件业务表或无法加 tag 的外部对象
- 两种方式注册的对象在字段权限配置页统一展示，无差异

**验收标准：**
- WHEN 系统启动时 THEN SHALL 自动扫描带有 `fieldperm` tag 的 model struct，提取 object_code + 字段列表（field_name + description）
- WHEN model struct 标注 `fieldperm:"-"`（struct 级别）THEN 系统 SHALL 跳过该对象，不注册到字段权限系统
- WHEN 管理员进入字段权限配置页 THEN 系统 SHALL 返回所有已注册的 object_code 及其字段列表，每个字段包含 key（json tag）和 description（fieldperm tag）
- WHEN 前端展示字段列表 THEN SHALL 以 `{description}({key})` 格式展示，如 `成本价(cost_price)`
- WHEN model struct 新增带 `fieldperm` tag 的字段 THEN 下次启动后自动出现在可配置列表中
- WHEN 字段无 `fieldperm` tag THEN 该字段 SHALL 不出现在字段权限配置列表中（仅标注的字段可配置）
- WHEN 管理员修改某字段的描述名 THEN 系统 SHALL 持久化自定义描述，后续展示以管理员修改后的描述为准
- WHEN 系统启动时发现 struct tag 描述与数据库已存储的自定义描述不同 THEN 系统 SHALL 保留管理员自定义描述（数据库优先于 tag 默认值）
- WHEN 管理员通过管理页面手动添加对象和字段 THEN 系统 SHALL 保存到数据库，与自动注册的对象统一展示
- WHEN 手动注册的对象/字段与自动注册冲突（同 object_code + field_name）THEN 系统 SHALL 合并，手动设置的描述优先

### Phase 4: 记录共享

#### FR-9: 记录共享 CRUD

**描述：** 提供记录共享规则的创建/查询/删除接口。

**验收标准：**
- WHEN 用户创建共享规则（object_code + record_id + share_to_type + share_to_id + access_level）THEN 系统 SHALL 保存记录并支持设置过期时间
- WHEN 查询某记录的共享规则列表 THEN 系统 SHALL 返回所有有效的共享规则
- WHEN 删除共享规则 THEN 系统 SHALL 目标用户立即失去访问权

#### FR-10: 记录共享生效（DataScopeCallback 扩展）

**描述：** DataScopeCallback 的 WHERE 条件中加入 OR 共享规则命中的逻辑。

**验收标准：**
- WHEN 用户原本无权访问记录 X，但记录 X 被共享给该用户 THEN 系统 SHALL 在查询结果中包含记录 X
- WHEN 共享规则的 share_to_type=ROLE THEN 拥有该角色的用户 SHALL 能看到记录
- WHEN 共享规则的 share_to_type=DEPT THEN 该部门的用户 SHALL 能看到记录
- WHEN 共享规则已过期（expire_at < NOW()）THEN 系统 SHALL 不再包含该记录
- WHEN 共享规则的 access_level=READ THEN 用户 SHALL 能查看但不能编辑；access_level=EDIT THEN 可编辑

#### FR-11: 通用共享对话框组件（前端）

**描述：** 实现一个可复用的共享对话框组件，业务页面可通过传入 object_code + record_id 调起。

**验收标准：**
- WHEN 用户点击"共享"按钮 THEN 系统 SHALL 弹出对话框，支持选择共享目标（用户/角色/部门）
- WHEN 选择共享目标后 THEN 系统 SHALL 支持设置 access_level（只读/可编辑）和过期时间
- WHEN 对话框中 THEN 系统 SHALL 展示已有的共享规则列表，支持删除

## 非功能需求

- 性能：权限计算（含权限集合并）缓存后 P99 < 10ms
- 安全：字段权限过滤在服务端执行，前端仅做展示优化
- 多租户：admin_field_permission 按 role_id 隔离（role 已含 tenant_id）；admin_record_share 含 tenant_id 字段
- 兼容性：未配置字段权限/记录共享时，系统行为与 CR-1 完全一致
- 数据迁移：如果现有子角色权限已超出父角色范围，CR-2 上线时需执行一次数据修复（裁剪超出部分）
