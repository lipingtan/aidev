# 任务：V2-CR2 权限体系增强

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 16 |
| 已完成 | 16 |
| 进行中 | 0 |
| 未开始 | 0 |
| 完成率 | 16/16 (100%) |
| 当前阶段 | ✅ 全部完成 |

---

## Phase 1: 角色继承（后端）

### Task 1: AssignResources 父角色子集校验 + 级联裁剪 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/role_service.go`
- 涉及模块: auth/service
- 不触碰: handler、middleware、前端

**Constraints（约束）:**
- PERMISSION_SET 类型角色跳过子集校验
- 顶级角色（parent_id=NULL）跳过校验
- 级联裁剪使用同步递归，在同一事务中完成
- 裁剪结果写入操作日志
- 响应增加 `affected_children` 字段
- 级联裁剪后必须批量失效受影响角色关联用户的权限缓存

**Acceptance（验证标准）:**
- AC: 子角色分配超出父角色范围 → 返回 ErrExceedsParentPermission
- AC: 父角色缩减 → 子角色超出部分被裁剪
- AC: 多级子角色递归裁剪正确
- AC: PERMISSION_SET 角色不受校验约束
- AC: 顶级角色分配无额外校验（RG-1）
- AC: 级联裁剪后受影响用户权限缓存已失效
- AC: `go build ./...` 零错误

### Task 2: AssignApis 父角色子集校验 + 级联裁剪 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/role_service.go`
- 涉及模块: auth/service
- 不触碰: handler、middleware、前端

**Constraints（约束）:**
- 逻辑与 Task 1 对称（资源 → API 权限）
- GROUP 展开后的 ENDPOINT ID 列表作为校验基准
- 级联裁剪同样适用于 API 绑定

**依赖**: Task 1（共享 cascadeTrimChildren 基础设施）

**Acceptance（验证标准）:**
- AC: 子角色 API 超出父角色范围 → 报错
- AC: 父角色缩减 API → 子角色级联裁剪
- AC: 顶级角色无额外校验（RG-2）
- AC: `go build ./...` 零错误

### Task 3: assignable-resources / assignable-apis 接口 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/auth/handler/role_handler.go`, `backend/common/auth/service/role_service.go`, `backend/common/auth/router.go`
- 不触碰: middleware、前端

**Acceptance（验证标准）:**
- AC: `GET /roles/:id/assignable-resources` 返回该角色已绑定的资源 ID 列表
- AC: `GET /roles/:id/assignable-apis` 返回该角色已绑定的 API 权限 ID 列表
- AC: 顶级角色返回租户订阅应用范围内全部资源/API
- AC: `go build ./...` 零错误

---

## Phase 1: 角色继承（前端）

### Task 4: 角色权限配置页增加可分配范围约束 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 前端角色权限配置相关 Vue 组件
- 不触碰: 后端代码

**依赖**: Task 3

**Acceptance（验证标准）:**
- AC: 编辑子角色权限时，调用 assignable-resources/apis 接口获取可选范围
- AC: 不可选的项 disabled 展示
- AC: 保存后如有级联裁剪，Toast 提示影响范围

---

## Phase 2: 权限集（后端）

### Task 5: PERMISSION_SET role_type 支持 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/role_service.go`, `backend/common/auth/model/role.go`
- 不触碰: 字段权限、记录共享模块

**Constraints（约束）:**
- PERMISSION_SET 创建时强制 parent_id=NULL（即使传入也忽略）
- 角色列表接口支持 role_type query param 过滤

**Acceptance（验证标准）:**
- AC: 创建 role_type=PERMISSION_SET 的角色成功，不校验 parent_id
- AC: `GET /roles?role_type=PERMISSION_SET` 仅返回权限集类型
- AC: `go build ./...` 零错误

### Task 6: 权限计算合并（含 Permission Set）✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/permission_service.go`, `backend/common/auth/service/resource_service.go`, `backend/common/auth/cache/`
- 不触碰: middleware（缓存层已合并后，middleware 无需改动）

**依赖**: Task 5

**Constraints（约束）:**
- GetUserMenu 合并所有角色（含 PERMISSION_SET）的 resourceIDs
- GetUserPermCodes 合并所有角色的 permission_code
- 权限缓存 key 包含用户所有 roleIDs

**Acceptance（验证标准）:**
- AC: 用户同时拥有普通角色和权限集，菜单为两者并集
- AC: 用户同时拥有普通角色和权限集，API 权限码为两者并集
- AC: 权限集删除后，用户实时失去额外权限（缓存失效）
- AC: 无权限集时计算结果与 CR-1 一致（RG-7）
- AC: `go build ./...` 零错误

---

## Phase 2: 权限集（前端）

### Task 7: 角色管理页 Tab 切换（角色 / 权限集）✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 前端角色管理页 Vue 组件
- 不触碰: 后端代码

**依赖**: Task 5

**Acceptance（验证标准）:**
- AC: 角色管理页顶部展示 [角色] [权限集] Tab
- AC: Tab 切换时按 role_type 参数过滤列表
- AC: 权限集的创建/编辑/删除/权限配置复用角色管理逻辑
- AC: 用户-角色分配界面同时展示普通角色和权限集供勾选

---

## Phase 3: 字段权限（后端）

### Task 8: 字段对象/字段定义 Model + Repository + 自动注册 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/model/field_object.go`, `backend/common/auth/model/field_definition.go`, `backend/common/auth/repository/field_object_repo.go`, `backend/common/auth/service/field_registry.go`
- 不触碰: 现有 model 文件（仅添加 fieldperm tag）

**Constraints（约束）:**
- struct tag `fieldperm:"描述"` 标注可配置字段
- struct 级别 `fieldperm:"-"` 跳过整个对象
- 自动注册 INSERT IGNORE，不覆盖已有自定义描述
- 支持手动注册 API

**Acceptance（验证标准）:**
- AC: 系统启动时自动扫描注册带 fieldperm tag 的 model
- AC: 跳过标注 `fieldperm:"-"` 的 model
- AC: admin_field_object 和 admin_field_definition 表正确填充
- AC: 手动注册 API 正常工作
- AC: `go build ./...` 零错误

### Task 9: 字段权限 CRUD 接口 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/model/field_permission.go`, `backend/common/auth/repository/field_permission_repo.go`, `backend/common/auth/service/field_permission_service.go`, `backend/common/auth/handler/field_permission_handler.go`, 修改 `backend/common/auth/router.go`
- 不触碰: middleware

**依赖**: Task 8

**Acceptance（验证标准）:**
- AC: GET /field-objects 返回已注册对象列表
- AC: GET /field-objects/:objectCode/fields 返回字段列表（含描述）
- AC: PUT /field-permissions 批量设置角色字段权限
- AC: DELETE /field-permissions/:id 删除配置
- AC: PUT /field-objects/:objectCode/fields/:fieldName 修改字段描述
- AC: `go build ./...` 零错误

### Task 10: FieldFilterMiddleware 实现 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/middleware/field_filter.go`, 修改 `backend/common/auth/router.go`（注册 middleware）
- 不触碰: 现有 handler 逻辑

**依赖**: Task 9

**Constraints（约束）:**
- Gin middleware 自动拦截 JSON 响应
- 路由-对象映射从 admin_field_object 加载
- 多角色冲突取最高权限（EDITABLE > VISIBLE > HIDDEN）
- 支持 Handler 设置 `skip_field_filter` 跳过
- 未配置字段权限时不做任何过滤（RG-3）

**Acceptance（验证标准）:**
- AC: HIDDEN 字段在响应 JSON 中被移除
- AC: VISIBLE 字段正常返回
- AC: 未配置时响应不变（RG-3）
- AC: 设置 skip_field_filter 的接口不被过滤
- AC: 多角色冲突取最高权限
- AC: `go build ./...` 零错误

---

## Phase 3: 字段权限（前端）

### Task 11: 字段权限配置页面 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 前端新增字段权限配置 Vue 组件
- 不触碰: 后端代码

**依赖**: Task 9

**Acceptance（验证标准）:**
- AC: 页面展示已注册对象列表，选择对象后展示字段列表
- AC: 字段以 `描述(key)` 格式展示
- AC: 每个字段可设置 VISIBLE/EDITABLE/HIDDEN
- AC: 支持修改字段描述名
- AC: 支持手动添加对象和字段

---

## Phase 4: 记录共享（后端）

### Task 12: RecordShare Model + Repository + CRUD 接口 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/model/record_share.go`, `backend/common/auth/repository/record_share_repo.go`, `backend/common/auth/service/record_share_service.go`, `backend/common/auth/handler/record_share_handler.go`
- 不触碰: middleware

**Acceptance（验证标准）:**
- AC: POST /record-shares 创建共享规则（含 expire_at）
- AC: GET /record-shares?object_code=X&record_id=Y 返回共享列表
- AC: DELETE /record-shares/:id 删除规则
- AC: tenant_id 隔离正确
- AC: `go build ./...` 零错误

### Task 13: DataScopeCallback 扩展 OR 共享规则 ✅

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: `backend/common/auth/middleware/data_scope.go`（或对应 callback 文件）
- 涉及模块: auth/middleware
- 不触碰: 现有 scope_type 逻辑（仅追加 OR）

**依赖**: Task 12

**Constraints（约束）:**
- 通过 GORM Statement.Context 获取 object_code（方案 B）
- 无 object_code 时不注入共享子查询（RG-4）
- 共享规则 expire_at 过期检查
- OR 关系：不缩小原有数据权限范围

**Acceptance（验证标准）:**
- AC: 用户原本无权访问的记录，被共享后可查到
- AC: share_to_type=USER/ROLE/DEPT 均生效
- AC: 过期规则不返回
- AC: 无 object_code context 时行为与 CR-1 一致（RG-4）
- AC: `go build ./...` 零错误

### Task 14: Service 层 WithObjectCode 上下文注入 ✅

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/middleware/context_keys.go`（WithObjectCode / GetObjectCode 辅助函数）
- 不触碰: 现有 service 文件（本 CR 仅提供框架，具体业务 service 注入留后续 CR）

**Acceptance（验证标准）:**
- AC: WithObjectCode / GetObjectCode 函数可编译使用
- AC: `go build ./...` 零错误

---

## Phase 4: 记录共享（前端）

### Task 15: 通用共享对话框组件 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 前端新增通用共享对话框 Vue 组件
- 不触碰: 后端代码

**依赖**: Task 12

**Acceptance（验证标准）:**
- AC: 组件接收 object_code + record_id props
- AC: 弹窗内支持选择共享目标（用户/角色/部门）
- AC: 支持设置 access_level（只读/可编辑）和过期时间
- AC: 展示已有共享规则列表，支持删除

---

## 数据修复

### Task 16: 现有数据兼容修复脚本 ✅

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: 新增 `backend/common/auth/migration/` 或 seed 脚本
- 不触碰: 运行时业务逻辑

**Constraints（约束）:**
- 扫描所有有 parent_id 的角色，检查其权限是否超出父角色
- 超出部分自动裁剪（等价于对每个父角色执行一次 cascadeTrimChildren）
- 输出修复报告（受影响角色列表 + 被裁剪的权限 ID）
- 幂等可重复执行

**Acceptance（验证标准）:**
- AC: 修复后所有子角色权限 ⊆ 父角色权限
- AC: 输出修复报告
- AC: `go build ./...` 零错误
