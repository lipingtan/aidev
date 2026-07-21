# 需求：数据权限增强 + 三级配置 + 配额管理（V2-CR3）

## 背景

V2 架构需要将数据权限从"仅 CUSTOM 固定值列表"升级为 5 种 scope_type（ALL/SELF/DEPT/DEPT_TREE/CUSTOM），同时将单层 sys_config 替换为三级配置链（SYSTEM > TENANT > USER），并在此基础上实现功能开关和配额管理能力。

## 用户故事

- 作为系统管理员，我希望为角色配置不同类型的数据权限（全部/仅自己/本部门/部门及下级/自定义值），以便实现细粒度的数据可见性控制
- 作为租户管理员，我希望管理本租户的组织架构（创建/编辑组织节点、分配人员归属），以便数据权限规则能正确生效
- 作为租户管理员，我希望覆盖系统级配置参数（如日期格式、分页大小），以便适配租户自身业务偏好
- 作为系统管理员，我希望通过配额限制租户的资源使用量（用户数/角色数/应用数），以便进行商业化计费和资源管控
- 作为系统管理员，我希望通过功能开关控制特定功能的可用性，以便按租户进行灰度发布或模块售卖

## 功能需求

### FR-1: DataScopeCallback scope_type 分支注入

**描述：** 增强 DataScopeCallback，根据角色的 scope_type 配置注入不同的 WHERE 条件

**验收标准：**
- WHEN 角色数据权限 scope_type=ALL THEN 系统 SHALL 不注入任何 WHERE 条件
- WHEN 角色数据权限 scope_type=SELF THEN 系统 SHALL 注入 `create_by = currentUserID`
- WHEN 角色数据权限 scope_type=DEPT THEN 系统 SHALL 调用 OrganizationProvider.GetOrgIds 获取用户所属组织节点 ID，注入 `{column} IN (orgIDs)`
- WHEN 角色数据权限 scope_type=DEPT_TREE THEN 系统 SHALL 调用 OrganizationProvider.GetSubOrgIds 获取组织节点及子节点 ID，注入 `{column} IN (allOrgIDs)`
- WHEN 角色数据权限 scope_type=CUSTOM THEN 系统 SHALL 按 dimension_values 注入 `{column} IN (values)`（与现有行为兼容）
- WHEN 用户有多个角色且 scope_type 不同 THEN 系统 SHALL 取并集（最宽松的范围生效）
- WHEN 用户任一角色 scope_type=ALL THEN 系统 SHALL 短路放行（不再检查其他角色的数据权限限制）
- WHEN OrganizationProvider 返回空切片 THEN 系统 SHALL 注入 `1 = 0` 等效于"无可见数据"
- WHEN 现有维度数据权限和记录共享逻辑 THEN 系统 SHALL 保持不变（RG-1）

### FR-2: OrganizationProvider 默认实现

**描述：** 提供基于 admin_org_unit + admin_user_org 表的默认 OrganizationProvider 实现，支持 DEPT/DEPT_TREE 查询

**验收标准：**
- WHEN 系统启动 THEN 系统 SHALL 自动注册 DefaultOrganizationProvider（基于 admin_org_unit + admin_user_org 表）
- WHEN 调用 GetOrgIds(userID, tenantID) THEN 系统 SHALL 返回用户在该租户下的主归属组织节点 ID 列表（is_primary=1）
- WHEN 调用 GetSubOrgIds(orgID, tenantID) THEN 系统 SHALL 返回指定组织节点及所有下级节点的 ID 列表
- WHEN admin_org_unit 表无数据 THEN 系统 SHALL 返回空切片（不报错）
- WHEN 外部应用注册自定义 provider THEN 系统 SHALL 允许覆盖默认实现

### FR-3: admin_data_scope DDL 升级

**描述：** 为 admin_data_scope 和 admin_data_scope_config 增加 scope_type 相关字段

**验收标准：**
- WHEN DDL 执行 THEN admin_data_scope 表 SHALL 新增 scope_type VARCHAR(16) NOT NULL DEFAULT 'CUSTOM'
- WHEN DDL 执行 THEN admin_data_scope_config 表 SHALL 新增 supported_scope_types JSON DEFAULT '["ALL","SELF","CUSTOM"]'
- WHEN 现有 admin_data_scope 数据 THEN scope_type 字段 SHALL 自动填充 'CUSTOM'（兼容旧数据）
- WHEN 角色配置数据权限选择 scope_type THEN 系统 SHALL 校验该 scope_type 在对应维度的 supported_scope_types 中

### FR-4: 三级配置服务（admin_config）

**描述：** 新建 admin_config 表，实现 SYSTEM > TENANT > USER 三级配置链合并查询

**验收标准：**
- WHEN 调用 Resolve(key, tenantID, userID) THEN 系统 SHALL 按 USER > TENANT > SYSTEM 优先级返回第一个匹配值
- WHEN USER scope 有值 THEN 系统 SHALL 返回 USER scope 值（忽略 TENANT/SYSTEM）
- WHEN USER scope 无值但 TENANT scope 有值 THEN 系统 SHALL 返回 TENANT scope 值
- WHEN 仅 SYSTEM scope 有值 THEN 系统 SHALL 返回 SYSTEM scope 值
- WHEN 所有 scope 均无值 THEN 系统 SHALL 返回调用方提供的 defaultValue
- WHEN 配置类型为 number THEN ResolveInt 方法 SHALL 返回 int 类型值
- WHEN 配置类型为 boolean THEN ResolveBool 方法 SHALL 返回 bool 类型值
- WHEN 频繁读取同一 key THEN 系统 SHALL 使用本地缓存减少数据库查询（缓存 TTL 可配）
- WHEN TENANT scope 配置写入 THEN 系统 SHALL 校验 tenant_id 归属合法性

### FR-5: sys_config 数据迁移

**描述：** 将现有 sys_config 数据一次性迁移到 admin_config（scope=SYSTEM），ConfigService 统一读取 admin_config

**验收标准：**
- WHEN 系统首次启动（迁移未执行） THEN 系统 SHALL 将 sys_config 所有记录导入 admin_config（scope=SYSTEM, scope_id=0, tenant_id=0）
- WHEN 迁移完成后 THEN ConfigHandler 的 List/Get/Create/Update/Delete 接口 SHALL 统一操作 admin_config 表
- WHEN sys_config 表 THEN 系统 SHALL 保留不删除（兼容期）
- WHEN 现有通过 ConfigService 读取配置的代码 THEN 系统 SHALL 继续正常工作（接口兼容）

### FR-6: 配置管理 API（CRUD + Resolve）

**描述：** 新增配置管理接口，支持 SYSTEM/TENANT/USER 三个 scope 的 CRUD 和配置解析

**验收标准：**
- WHEN GET /api/v1/admin/configs THEN 系统 SHALL 返回配置列表（支持按 scope/key/is_feature_flag 过滤）
- WHEN POST /api/v1/admin/configs THEN 系统 SHALL 创建配置（需指定 scope/scope_id/tenant_id）
- WHEN PUT /api/v1/admin/configs/:id THEN 系统 SHALL 更新配置
- WHEN DELETE /api/v1/admin/configs/:id THEN 系统 SHALL 软删除配置
- WHEN GET /api/v1/admin/configs/resolve/:key THEN 系统 SHALL 按三级合并逻辑返回解析后的有效值
- WHEN GET /api/v1/admin/configs/feature-flags THEN 系统 SHALL 返回当前租户的功能开关列表及状态

### FR-7: 功能开关

**描述：** is_feature_flag=1 的配置项作为功能开关，可按租户控制功能模块的启用/禁用

**验收标准：**
- WHEN 配置项 is_feature_flag=1 且该租户的值为 "false"/"0" THEN 系统 SHALL 视为功能关闭
- WHEN 功能关闭且用户访问对应模块 API THEN 系统 SHALL 返回 403（code=40302, message="该功能未开启"）
- WHEN 功能开关未配置或值为 "true"/"1" THEN 系统 SHALL 放行
- WHEN SUPER_ADMIN 访问功能开关管理接口 THEN 系统 SHALL 正常返回（功能开关不拦截管理接口自身）

### FR-8: 配额管理

**描述：** 通过 admin_config（scope=TENANT）存储租户配额，Service 层在关键操作前校验

**验收标准：**
- WHEN 租户创建用户且当前用户数 >= quota.max_admin_users THEN 系统 SHALL 拒绝操作并返回"管理用户数已达上限"
- WHEN 租户创建角色且当前角色数 >= quota.max_roles THEN 系统 SHALL 拒绝操作并返回"角色数已达上限"
- WHEN 租户订阅应用且当前应用数 >= quota.max_apps THEN 系统 SHALL 拒绝操作并返回"可订阅应用数已达上限"
- WHEN 配额未配置 THEN 系统 SHALL 使用 SYSTEM scope 默认值（999999 = 不限制）
- WHEN SUPER_ADMIN 操作 THEN 系统 SHALL 跳过配额校验（超管不受限）

### FR-9: 前端三级配置管理界面

**描述：** 完整实现 SYSTEM/TENANT/USER 三个 scope 的配置 CRUD 管理页面

**验收标准：**
- WHEN 管理员进入系统配置页 THEN 系统 SHALL 展示配置列表，支持 scope 切换 Tab（系统/租户/用户）
- WHEN 管理员创建配置 THEN 系统 SHALL 提供 scope 选择、key/value/type 输入、功能开关标记
- WHEN 管理员查看配置详情 THEN 系统 SHALL 展示该 key 在各 scope 的值（方便理解覆盖关系）；WHEN 该 key 无 SYSTEM scope 默认值 THEN 前端 SHALL 显示"（未设置系统默认）"
- WHEN 管理员编辑租户 scope 配置 THEN 系统 SHALL 显示所覆盖的系统默认值作为参考
- WHEN 管理员创建 USER scope 配置 THEN 系统 SHALL 提供用户选择器（下拉搜索当前租户下的用户）

### FR-10: 数据权限配置界面支持 scope_type

**描述：** 角色数据权限配置页增加 scope_type 选择

**验收标准：**
- WHEN 管理员配置角色数据权限 THEN 系统 SHALL 展示维度的 supported_scope_types 供选择
- WHEN 选择 CUSTOM THEN 系统 SHALL 展示值列表输入
- WHEN 选择 DEPT/DEPT_TREE THEN 系统 SHALL 无需额外输入（运行时动态解析）
- WHEN 选择 ALL THEN 系统 SHALL 无需额外输入
- WHEN 选择 SELF THEN 系统 SHALL 无需额外输入

## 非功能需求

- 性能：配置 Resolve 接口响应时间 < 50ms（P99，含缓存命中）；缓存未命中时 < 200ms
- 安全：TENANT/USER scope 写入需校验 tenant_id 归属；配额配置仅 SUPER_ADMIN 可修改
- 回归：现有 DataScopeCallback 维度+记录共享行为不被破坏（RG-1）
- 回归：现有 ConfigService 对外接口行为兼容（RG-2）
- 兼容：sys_config 表保留不删除
