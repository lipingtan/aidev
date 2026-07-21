# 任务：V2-CR3 数据权限增强 + 三级配置 + 配额管理

## 执行进度

| 指标 | 值 |
|------|-----|
| 总任务数 | 18 |
| 已完成 | 0 |
| 进行中 | 0 |
| 未开始 | 18 |
| 完成率 | 0/18 (0%) |
| 当前阶段 | Phase 1: DDL + 模型层 |

---

## Phase 1: DDL + 模型层

### Task 1: admin_org_unit 表 + OrgUnit Model ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/auth/model/org_unit.go`（新建）
- 不触碰: 其他 model 文件

**Constraints（约束）:**
- 雪花 ID，BeforeCreate 钩子
- JSON tag 含 `json:",string"` 的 ID 字段
- tenant_id + code 唯一约束
- 含 Version int 字段（乐观锁）

**Acceptance（验证标准）:**
- AC: OrgUnit struct 含 ID/TenantID/ParentID/NodeType/Name/Code/SortOrder/Status/Version/CreatedAt/UpdatedAt
- AC: TableName() 返回 "admin_org_unit"
- AC: go build ./... 零错误

### Task 2: admin_user_org 表 + UserOrg Model ⬜

**复杂度**: 低

**Scope（边界）:**
- 涉及文件: `backend/common/auth/model/user_org.go`（新建）

**Acceptance:**
- AC: UserOrg struct 含 ID/UserID/OrgUnitID/TenantID/IsPrimary/CreatedAt
- AC: TableName() 返回 "admin_user_org"
- AC: go build ./... 零错误

### Task 3: admin_data_scope 增加 scope_type + admin_data_scope_config 增加 supported_scope_types ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/auth/model/data_scope.go`, `backend/common/auth/model/data_scope_config.go`
- 不触碰: repository/service 层

**Constraints（约束）:**
- DataScope 增加 ScopeType string 字段，GORM tag `gorm:"type:varchar(16);not null;default:'CUSTOM'"`
- DataScopeConfig 增加 SupportedScopeTypes datatypes.JSON 字段

**Acceptance（验证标准）:**
- AC: Model 编译通过，新字段正确标注 GORM tag 和 JSON tag
- AC: 旧数据兼容（default 'CUSTOM'）
- AC: go build ./... 零错误

### Task 4: admin_config 表 + AdminConfig Model ⬜

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: `backend/common/auth/model/admin_config.go`（新建）
- 不触碰: 现有 config.go（SysConfig 保留）

**Constraints（约束）:**
- uk_scope_key 唯一约束在 GORM tag 中标注
- deleted_at 使用 gorm.DeletedAt 支持软删除

**Acceptance（验证标准）:**
- AC: AdminConfig struct 含 ID/ConfigKey/ConfigValue/ConfigType/Scope/ScopeID/TenantID/DisplayName/Description/IsFeatureFlag/Status/DeletedAt/CreatedAt/UpdatedAt
- AC: TableName() 返回 "admin_config"
- AC: go build ./... 零错误

---

## Phase 2: 组织架构后端（Repository + Service + Handler）

### Task 5: OrgUnit Repository ⬜

**复杂度**: 中

**依赖**: Task 1

**Scope（边界）:**
- 涉及文件: `backend/common/auth/repository/org_unit_repo.go`（新建）
- 不触碰: 其他 repository 文件

**Acceptance（验证标准）:**
- AC: 接口含 Create/Update/Delete/FindByID/ListByTenant/HasChildren/FindByTenantAndCode
- AC: go build ./... 零错误

### Task 6: OrgUnit Service + UserOrg Service ⬜

**复杂度**: 高

**依赖**: Task 5, Task 2

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/org_unit_service.go`（新建）
- 涉及模块: auth/service
- 不触碰: 其他 service 文件

**Constraints（约束）:**
- 删除节点前检查子节点（有子节点禁止删除）
- 树形查询返回 OrgUnitNode 递归结构
- SetNodeUsers 全量替换（先删后插，保留 is_primary）
- UpdateOrgUnit 必须校验 version 字段，不匹配时返回"数据已被其他操作修改，请刷新后重试"

**Acceptance（验证标准）:**
- AC: CreateOrgUnit/UpdateOrgUnit/DeleteOrgUnit/GetTree/GetNodeUsers/SetNodeUsers 方法
- AC: 删除有子节点的组织报错
- AC: go build ./... 零错误

### Task 7: OrgUnit Handler + 路由注册 ⬜

**复杂度**: 高

**依赖**: Task 6

**Scope（边界）:**
- 涉及文件: `backend/common/auth/handler/org_unit_handler.go`（新建）, `backend/common/auth/router.go`
- 不触碰: 其他 handler

**Constraints（约束）:**
- 路由注册到 admin group（/api/v1/admin/org-units）
- permission_code 按 design.md 定义
- AutoDiscover endpointPermissionCodes/endpointDisplayNames 需同步更新

**Acceptance（验证标准）:**
- AC: 6 个接口全部注册并可访问
- AC: go build ./... 零错误
- AC:【回归】现有路由不受影响（RG-5）

---

## Phase 3: DefaultOrganizationProvider

### Task 8: DefaultOrganizationProvider 实现 ⬜

**复杂度**: 高

**依赖**: Task 1, Task 2

**Scope（边界）:**
- 涉及文件: `backend/common/auth/spi/default_org_provider.go`（新建）, `backend/common/auth/auth.go`（注册 provider）
- 不触碰: spi.go 接口定义

**Constraints（约束）:**
- GetOrgIds: 查 admin_user_org（is_primary=1）
- GetSubOrgIds: 单次查询该租户全量节点 + 内存构建树 BFS（避免 N+1）
- GetOrgPath: 从叶子向根遍历
- 注册到 auth.go 的依赖注入链中（替换 NoOpOrganizationProvider）

**Acceptance（验证标准）:**
- AC: 三个方法实现完整，空数据返回空切片不报错
- AC: auth.go 中 provider 注册正确
- AC: go build ./... 零错误

---

## Phase 4: DataScopeCallback 增强

### Task 9: DataScopeDimension 扩展 + DataScopeCallback scope_type 分支 ⬜

**复杂度**: 高

**依赖**: Task 3, Task 8

**Scope（边界）:**
- 涉及文件: `backend/common/auth/middleware/data_scope_middleware.go`, `backend/common/auth/middleware/data_scope_callback.go`
- 不触碰: data_scope_callback_test.go（后续 Task 更新测试）

**Constraints（约束）:**
- DataScopeDimension 增加 ScopeType/UserID/TenantID 字段
- scope_type=ALL 时短路返回（任一角色 ALL → 整个维度不注入条件）
- scope_type=SELF 注入 create_by = UserID
- scope_type=DEPT/DEPT_TREE 调用 OrganizationProvider
- scope_type=CUSTOM 保持现有逻辑
- 多角色同维度取并集
- 现有 injectDimensionScope 逻辑重构为 injectScopeTypeWhere
- orgProvider 需通过 RegisterDataScopeCallback 的参数传入，存为闭包变量

**Acceptance（验证标准）:**
- AC: 5 种 scope_type 正确注入 WHERE
- AC:【回归】CUSTOM 行为不变（RG-1）
- AC:【回归】记录共享逻辑不变（RG-2）
- AC: go build ./... 零错误

### Task 10: DataScopeMiddleware 加载 scope_type 到 context ⬜

**复杂度**: 中

**依赖**: Task 9, Task 3

**Scope（边界）:**
- 涉及文件: `backend/common/auth/middleware/data_scope_load.go`（或现有加载逻辑文件）
- 不触碰: data_scope_callback.go

**Constraints（约束）:**
- 从 admin_data_scope 表加载时读取 scope_type 字段
- 填充 DataScopeDimension 的 ScopeType/UserID/TenantID

**Acceptance（验证标准）:**
- AC: context 中 DataScopeContext.Dimensions 含正确的 ScopeType
- AC: go build ./... 零错误

---

## Phase 5: 三级配置后端

### Task 11: AdminConfigService 实现（三级合并 + 缓存） ⬜

**复杂度**: 高

**依赖**: Task 4

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/admin_config_service.go`（新建）
- 不触碰: 现有 config_service.go

**Constraints（约束）:**
- Resolve/ResolveInt/ResolveBool/ResolveString 方法
- IsFeatureEnabled 方法
- 本地缓存 sync.Map + TTL（60s 默认）
- InvalidateCache 按 key 前缀清除
- TENANT scope 写入校验 tenant_id 归属

**Acceptance（验证标准）:**
- AC: USER > TENANT > SYSTEM 优先级正确
- AC: 缓存命中时不查库
- AC: InvalidateCache 后下次查询走库
- AC: go build ./... 零错误

### Task 12: AdminConfig Handler + 路由注册 ⬜

**复杂度**: 高

**依赖**: Task 11

**Scope（边界）:**
- 涉及文件: `backend/common/auth/handler/admin_config_handler.go`（新建）, `backend/common/auth/router.go`
- 不触碰: 现有 config_handler.go（保留兼容）

**Constraints（约束）:**
- List 支持 scope/key/is_feature_flag 过滤
- Create/Update 时调用 InvalidateCache
- resolve/:key 接口调用 Resolve 方法
- feature-flags 接口返回 is_feature_flag=1 且当前租户有效值

**Acceptance（验证标准）:**
- AC: 6 个接口全部注册并可访问
- AC: resolve/:key 返回正确的三级合并值
- AC:【回归】现有 /api/v1/admin/configs 接口保持可用（RG-3）
- AC: go build ./... 零错误

### Task 13: sys_config 数据迁移 ⬜

**复杂度**: 中

**依赖**: Task 11

**Scope（边界）:**
- 涉及文件: `backend/common/auth/auth.go`（启动时调用迁移）
- 不触碰: sys_config 表结构

**Constraints（约束）:**
- 幂等：已迁移的 key 跳过
- 在 AutoDiscover 之后、路由注册之前执行
- 保留 sys_config 表不删

**Acceptance（验证标准）:**
- AC: 启动后 admin_config 含 sys_config 所有数据（scope=SYSTEM）
- AC: 重复启动不报错不重复插入
- AC: go build ./... 零错误

---

## Phase 6: 功能开关 + 配额

### Task 14: DynamicPermissionMiddleware 集成功能开关 ⬜

**复杂度**: 中

**依赖**: Task 11

**Scope（边界）:**
- 涉及文件: `backend/common/auth/middleware/dynamic_permission_middleware.go`, `backend/common/auth/router.go`
- 不触碰: 权限码校验逻辑

**Constraints（约束）:**
- 在 SUPER_ADMIN 放行之后、权限码校验之前检查
- 按 module_code 查 feature.{module_code}.enabled
- 关闭时返回 403 code=40302 message="该功能未开启"
- AdminConfigService 需通过构造参数传入
- 需同步修改 router.go 中 DynamicPermissionMiddleware 的调用方，传入 configSvc 实例

**Acceptance（验证标准）:**
- AC: 功能开关关闭 → 403/40302
- AC: 功能开关开启或未配置 → 正常走权限码校验
- AC: SUPER_ADMIN 不受功能开关限制
- AC:【回归】现有权限校验逻辑不变（RG-6）
- AC: router.go 调用方签名同步更新
- AC: go build ./... 零错误

### Task 15: 配额校验植入 Service 层 ⬜

**复杂度**: 中

**依赖**: Task 11

**Scope（边界）:**
- 涉及文件: `backend/common/auth/service/user_service.go`, `backend/common/auth/service/role_service.go`, `backend/common/auth/service/tenant_service.go`（订阅应用处）
- 不触碰: 查询/更新/删除方法

**Constraints（约束）:**
- 仅在 Create 相关方法中植入
- SUPER_ADMIN 跳过配额检查
- configSvc.ResolveInt(tenantID, "quota.xxx", 999999)
- 超限返回 errors.ErrQuotaExceeded

**Acceptance（验证标准）:**
- AC: 用户数超限 → 拒绝创建
- AC: 角色数超限 → 拒绝创建
- AC: 应用订阅数超限 → 拒绝订阅
- AC: SUPER_ADMIN 不受限（RG-4）
- AC: 未配置配额 → 默认 999999 不限制
- AC: go build ./... 零错误

### Task 16: Seed 数据（默认配额 + 功能开关示例） ⬜

**复杂度**: 低

**依赖**: Task 13

**Scope（边界）:**
- 涉及文件: 种子数据文件或 auth.go 中的初始化逻辑

**Acceptance:**
- AC: admin_config 含 quota.max_admin_users/quota.max_roles/quota.max_apps（scope=SYSTEM, value=999999）
- AC: 种子维度配置的 supported_scope_types 包含全部 5 种（["ALL","SELF","DEPT","DEPT_TREE","CUSTOM"]）
- AC: go build ./... 零错误

---

## Phase 7: 前端

### Task 17: 前端 — 组织架构管理页 ⬜

**复杂度**: 高

**依赖**: Task 7

**Scope（边界）:**
- 涉及文件: `frontend/src/views/org/` 目录（新建）, `frontend/src/api/org.ts`（新建）
- 不触碰: 其他 views

**Acceptance（验证标准）:**
- AC: 树形展示组织架构（el-tree）
- AC: 新增/编辑/删除节点（node_type 下拉选择）
- AC: 节点下用户管理（穿梭框分配）
- AC: API 路径带 /api/v1/admin/ 前缀
- AC: 配额超限操作时弹窗提示含"请联系管理员升配"引导

### Task 18: 前端 — 三级配置管理页 + 数据权限 scope_type 选择 ⬜

**复杂度**: 高

**依赖**: Task 12, Task 10

**Scope（边界）:**
- 涉及文件: `frontend/src/views/config/`（改造）, `frontend/src/views/role/data-scope/`（改造）, `frontend/src/api/config.ts`（改造）
- 不触碰: 其他 views

**Acceptance（验证标准）:**
- AC: 配置页支持 SYSTEM/TENANT/USER Tab 切换
- AC: 创建配置可选 scope + 标记功能开关
- AC: 配置详情展示各 scope 值覆盖关系
- AC: 角色数据权限配置支持 scope_type 下拉（ALL/SELF/DEPT/DEPT_TREE/CUSTOM）
- AC: 选择 CUSTOM 时显示值列表输入，其他类型无需额外输入
- AC: API 路径带 /api/v1/admin/ 前缀
