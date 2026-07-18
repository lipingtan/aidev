# 需求：auth-rbac（RBAC 权限管理框架 Go 版）

## 背景

企业级项目频繁重复实现 RBAC 权限管理，导致安全模型不一致、维护成本高。基于 Java 版 auth-pivot 的成熟设计，全新实现 Go 版 RBAC 权限管理框架，作为平台管理端基础能力模块。采用 Go 1.24 + Gin + GORM + MySQL + Redis 技术栈，代码位于 `backend/common/auth/` 模块内。

## 用户故事

- 作为平台管理员，我希望通过统一的权限管理后台管理用户/角色/菜单/接口权限，以便集中控制系统访问
- 作为开发者，我希望引入 auth 模块后零配置即可运行基础认证+授权，以便快速接入
- 作为开发者，我希望通过 SPI 接口扩展用户来源/缓存/认证方式，以便适配不同业务场景
- 作为运维人员，我希望对异常用户执行强制下线，以便保障平台安全
- 作为业务系统负责人，我希望同一角色按应用隔离权限，以便多应用共存不互相干扰
- 作为管理员，我希望对敏感数据配置行级数据权限，以便不同角色只能看到自己范围内的数据

## 功能需求

### FR-1: 用户管理

**描述：** 用户为全局实体（不直接绑定租户），代表一个可登录身份。一个用户可关联多个租户，通过 admin_user_tenant 建立用户-租户 M:N 关系。提供 UserProvider SPI 支持外部用户源对接。

**验收标准：**
- WHEN 调用用户创建接口 THEN 系统 SHALL 在 admin_user 创建全局用户记录（无 tenant_id），密码使用 bcrypt 加密存储
- WHEN 将用户关联到租户 THEN 系统 SHALL 在 admin_user_tenant 创建关联记录
- WHEN 调用用户查询接口（租户上下文） THEN 系统 SHALL 返回该租户下已关联的用户列表，支持分页
- WHEN 调用用户查询接口（平台管理员） THEN 系统 SHALL 返回全局所有用户
- WHEN 调用用户更新接口 THEN 系统 SHALL 更新用户信息并递增 version（乐观锁）
- WHEN 调用用户删除接口 THEN 系统 SHALL 执行软删除（设置 deleted_at）
- WHEN 业务方注册自定义 UserProvider THEN 系统 SHALL 使用该 Provider 替代默认实现

### FR-2: 租户管理

**描述：** 租户生命周期管理，包括租户的创建、启用/禁用、配置，以及租户初始化（预置角色、默认管理员）。用户必须先关联租户才能在该租户下获得身份和权限。

**验收标准：**
- WHEN 创建租户 THEN 系统 SHALL 在 admin_tenant 创建记录，并自动初始化预置角色集和租户管理员账号
- WHEN 禁用租户 THEN 系统 SHALL 将该租户 status 设为禁用，该租户下所有用户登录时返回"租户已禁用"
- WHEN 启用租户 THEN 系统 SHALL 恢复该租户正常访问
- WHEN 删除租户 THEN 系统 SHALL 执行软删除，关联的用户/角色/权限数据保留但不可访问
- WHEN 查询租户列表 THEN 系统 SHALL 仅平台管理员可操作，支持分页和状态筛选
- WHEN 用户未关联某租户 THEN 系统 SHALL 拒绝该用户在此租户下的任何权限操作

### FR-3: 角色管理

**描述：** 角色 CRUD，角色归属于租户（tenant_id 隔离），支持层级继承（parent_id）、临时授权时间窗口。用户在某个租户下可被分配该租户的角色。

**验收标准：**
- WHEN 创建角色 THEN 系统 SHALL 在 admin_role 表创建记录，关联 tenant_id
- WHEN 设置 parent_id THEN 系统 SHALL 验证不产生循环引用，构建层级继承树
- WHEN 为用户在某租户下分配角色 THEN 系统 SHALL 验证用户已关联该租户，并在 admin_user_role 创建记录（含 tenant_id）
- WHEN 为用户分配角色并设置 effective_start/effective_end THEN 系统 SHALL 仅在时间窗口内该角色对用户生效
- WHEN 临时授权过期 THEN 系统 SHALL 自动使角色不再生效（权限判断时实时检查）
- WHEN 查询角色列表 THEN 系统 SHALL 按 tenant_id 隔离，平台管理员角色全局可见
- WHEN 删除角色 THEN 系统 SHALL 检查无用户绑定后执行软删除，有绑定时拒绝并返回提示

### FR-4: 菜单/按钮资源管理

**描述：** 树形菜单/按钮资源管理，控制前端页面/按钮的显示隐藏，带 app_code 应用归属。资源定义为租户级（tenant_id 隔离）。

**验收标准：**
- WHEN 创建资源 THEN 系统 SHALL 在 admin_resource 创建记录，type 为 MENU 或 BUTTON，关联 tenant_id
- WHEN 设置 parent_id THEN 系统 SHALL 验证不产生循环引用
- WHEN 查询用户菜单树 THEN 系统 SHALL 根据用户角色合并权限，返回有权限的资源树
- WHEN 资源带 app_code THEN 系统 SHALL 按应用维度过滤资源
- WHEN 为角色分配菜单权限 THEN 系统 SHALL 在 admin_role_resource 表建立关联

### FR-5: 接口权限管理

**描述：** 独立于菜单的接口权限树，按业务对象分组，控制后端 API 访问。

**验收标准：**
- WHEN 创建接口权限节点 THEN 系统 SHALL 区分 GROUP（分组）和 ENDPOINT（叶子）类型
- WHEN 为角色分配接口权限 THEN 系统 SHALL 仅允许绑定 ENDPOINT 叶子节点
- WHEN 勾选 GROUP 节点分配 THEN 系统 SHALL 展开并绑定该 GROUP 下所有 ENDPOINT 叶子
- WHEN ENDPOINT 节点携带 url_pattern + http_method + permission_code THEN 系统 SHALL 在请求匹配时执行权限校验
- WHEN 接口权限带 app_code THEN 系统 SHALL 按应用维度隔离

### FR-6: API 自动发现

**描述：** 三层策略自动注册接口权限——自动扫描 gin.Routes() 兜底 + 代码声明 RegisterAPIs() + 管理后台手动补充。

**验收标准：**
- WHEN 应用启动 THEN 系统 SHALL 遍历 gin.Engine.Routes()，将未注册的 endpoint 插入 admin_api_permission（status=UNASSIGNED）
- WHEN 开发者调用 RegisterAPIs(metadata) THEN 系统 SHALL 带完整元数据注册为 ACTIVE 状态
- WHEN 管理员在后台编辑 UNASSIGNED endpoint THEN 系统 SHALL 允许补充标题/分组/拖入 GROUP
- WHEN endpoint 被分配到 GROUP 下 THEN 系统 SHALL 将其状态从 UNASSIGNED 变更为 ACTIVE

### FR-7: 数据权限

**描述：** 维度注册 + GORM Scope 拦截 + target_entity 业务对象绑定，实现行级数据过滤。

**验收标准：**
- WHEN 注册数据权限维度 THEN 系统 SHALL 在 admin_data_scope_config 创建维度定义（dimension_name + table_column + value_source）
- WHEN 为角色配置数据权限 THEN 系统 SHALL 在 admin_data_scope 绑定角色+维度+target_entity+dimension_values
- WHEN 执行带 @DataScope 标记的查询 THEN 系统 SHALL 通过 GORM Scope 自动拼接 WHERE 条件
- WHEN 用户持有多角色 THEN 系统 SHALL 对同一维度取并集
- WHEN 无数据权限配置 THEN 系统 SHALL 不添加额外过滤条件

### FR-8: 认证体系

**描述：** JWT 为主认证方式（可扩展 Session/OAuth2），含 Token 签发、刷新、黑名单。

**验收标准：**
- WHEN 用户登录成功 THEN 系统 SHALL 签发 JWT Token（含 userId、可用租户列表）
- WHEN 用户选择/切换租户上下文 THEN 系统 SHALL 验证用户已关联该租户，签发含 tenantId 的访问 Token
- WHEN 请求携带有效 Token 且含 tenantId THEN 系统 SHALL 解析并注入用户上下文（含当前租户下的角色和权限）
- WHEN 用户未关联任何租户 THEN 系统 SHALL 登录成功但无法进入任何租户工作空间
- WHEN Token 过期 THEN 系统 SHALL 返回 401 错误
- WHEN Token 在黑名单中 THEN 系统 SHALL 拒绝请求返回 401
- WHEN 业务方注册自定义 AuthProvider THEN 系统 SHALL 使用该 Provider 替代默认 JWT 实现

### FR-9: 两级缓存

**描述：** L1 角色权限缓存 + L2 用户合并权限缓存，支持本地（sync.Map/go-cache）和 Redis。

**验收标准：**
- WHEN 权限判断请求到达 THEN 系统 SHALL 优先查 L2 用户缓存（key: user:{userId}:tenant:{tenantId}:app:{appCode}），命中即返回
- WHEN L2 未命中 THEN 系统 SHALL 查 L1 角色缓存合并后写入 L2
- WHEN 角色权限变更 THEN 系统 SHALL 清除该角色 L1 缓存 + 所有持有该角色用户的 L2 缓存
- WHEN 用户角色变更 THEN 系统 SHALL 仅清除该用户 L2 缓存
- WHEN 配置 cache-type=redis THEN 系统 SHALL 使用 Redis 作为缓存后端

### FR-10: 应用隔离

**描述：** admin_application 实体管理，角色通过 admin_role_app M:N 关联应用，资源/接口权限按 app_code 归属。

**验收标准：**
- WHEN 创建应用 THEN 系统 SHALL 在 admin_application 创建记录（app_code 唯一）
- WHEN 为角色绑定应用 THEN 系统 SHALL 在 admin_role_app 创建 M:N 关联
- WHEN 查询用户权限 THEN 系统 SHALL 按当前 app_code 过滤，仅返回该应用下已关联角色的权限
- WHEN 未指定 app_code THEN 系统 SHALL 使用默认应用标识（default）

### FR-11: 组织结构 SPI

**描述：** 框架不管理组织结构，提供 OrganizationProvider 接口供业务方对接自有组织/部门体系。组织结构与租户关联（每个租户有独立的组织树）。

**验收标准：**
- WHEN 业务方注册 OrganizationProvider THEN 系统 SHALL 在数据权限等场景调用该 Provider 获取组织信息
- WHEN 调用 OrganizationProvider THEN 系统 SHALL 传入当前 tenant_id，Provider 返回该租户下的组织数据
- WHEN 未注册 OrganizationProvider THEN 系统 SHALL 使用 NoOp 默认实现（不注入组织过滤条件）

### FR-12: 权限表达式引擎

**描述：** module:object:action 格式权限码，支持 `*`（单层通配）和 `**`（多层通配）。

**验收标准：**
- WHEN 判断 permission="sys:user:add" 且用户持有 "sys:user:*" THEN 系统 SHALL 返回有权限
- WHEN 判断 permission="sys:user:detail:view" 且用户持有 "sys:user:*" THEN 系统 SHALL 返回无权限（* 仅匹配一层）
- WHEN 用户持有 "sys:**" THEN 系统 SHALL 匹配 sys 下所有层级
- WHEN 用户持有 "**" THEN 系统 SHALL 匹配一切权限码（超级管理员）
- WHEN 精确权限码匹配 THEN 系统 SHALL O(1) 查找（HashSet）

### FR-13: JWT 黑名单（强制下线）

**描述：** 独立存储的 Token 黑名单，支持管理员强制下线指定用户。

**验收标准：**
- WHEN 管理员执行强制下线 THEN 系统 SHALL 将该用户当前 Token 加入黑名单
- WHEN 黑名单 Token 过期 THEN 系统 SHALL 自动清理（TTL = Token 剩余有效期）
- WHEN 配置 cache-type=local THEN 系统 SHALL 使用 ConcurrentMap + 定时清理
- WHEN 配置 cache-type=redis THEN 系统 SHALL 使用 Redis SET + EXPIRE

### FR-14: 操作日志

**描述：** 所有管理操作记录带业务含义的操作日志，包括操作人、操作对象、操作类型、变更内容摘要、IP 等上下文信息。

**验收标准：**
- WHEN 执行任何增删改操作（用户/角色/租户/资源/权限等） THEN 系统 SHALL 记录操作日志到 admin_operation_log
- WHEN 记录日志 THEN 系统 SHALL 包含：操作人ID、租户ID、操作模块、操作类型（CREATE/UPDATE/DELETE/ASSIGN/REVOKE）、目标对象类型+ID、业务摘要（中文描述）、原参数（变更前 JSON）、新参数（变更后 JSON）、客户端 IP、时间戳
- WHEN 业务摘要生成 THEN 系统 SHALL 使用可读的中文描述（如"为用户张三分配角色[编辑者]"、"禁用租户[测试租户]"）
- WHEN 查询操作日志 THEN 系统 SHALL 支持按模块/操作类型/操作人/时间范围/目标对象筛选，支持分页
- WHEN 日志写入失败 THEN 系统 SHALL 不影响主业务流程（异步写入，失败降级到本地日志文件）

### FR-15: 前端管理页面

**描述：** 租户/用户/角色/菜单/接口权限/数据权限/应用/操作日志管理的完整前端管理页面。

**验收标准：**
- WHEN 访问租户管理页面 THEN 系统 SHALL 展示租户列表，支持搜索/分页/新增/编辑/启用/禁用（仅平台管理员）
- WHEN 访问用户管理页面 THEN 系统 SHALL 展示用户列表，支持搜索/分页/新增/编辑/删除/角色分配
- WHEN 访问角色管理页面 THEN 系统 SHALL 展示角色树，支持新增/编辑/删除/菜单权限分配/接口权限分配/数据权限配置
- WHEN 访问菜单管理页面 THEN 系统 SHALL 展示菜单树，支持拖拽排序/新增/编辑/删除
- WHEN 访问接口权限页面 THEN 系统 SHALL 展示按业务对象分组的接口权限树，支持 UNASSIGNED endpoint 拖入分组
- WHEN 访问数据权限页面 THEN 系统 SHALL 展示维度配置列表和角色维度值绑定
- WHEN 访问应用管理页面 THEN 系统 SHALL 展示应用列表，支持 CRUD 和角色关联管理
- WHEN 访问操作日志页面 THEN 系统 SHALL 展示操作日志列表，支持按模块/操作人/时间范围筛选

## 非功能需求

### NFR-1: 性能
- 权限判断热路径响应 < 1ms（缓存命中）
- API 接口响应 < 200ms（P99）
- 缓存失效采用分批异步 + 随机延迟，防止雪崩

### NFR-2: 安全
- 密码 bcrypt 加密存储
- JWT 含过期时间 + 黑名单机制
- 租户隔离查询：业务表按 tenant_id 过滤；admin_user 为全局表不带 tenant_id
- SQL 参数化查询，防注入

### NFR-3: 可扩展性
- 10+ 个 interface 扩展点（UserProvider、RoleProvider、CacheAdapter、AuthProvider、OrganizationProvider 等）
- 每个子功能独立开关（YAML 配置启停）
- 扩展点通过 interface + 构造函数注入（显式，无框架魔法）

### NFR-6: 模块独立性
- 用户权限模块不依赖框架中其他业务模块，所有功能自包含
- JWT 签发/验证、密码加密、缓存、数据权限拦截等均由本模块自行实现
- 仅依赖基础库（Gin、GORM、Redis client），不依赖项目内其他 common 子模块

### NFR-4: 多租户隔离
- 角色/资源/接口权限/数据权限等表带 tenant_id 字段
- admin_user 为全局表（无 tenant_id），通过 admin_user_tenant 关联租户
- 租户上下文内的查询自动注入 tenant_id 过滤
- 平台管理员（SUPER_ADMIN）跨租户访问

### NFR-5: 零侵入默认
- 引入模块 + 零配置即可运行
- 所有功能提供合理默认值
- GORM AutoMigrate 自动建表

## 数据模型

表名统一使用 `admin_` 前缀。

| 表名 | 职责 | 软删除 | 乐观锁 |
|------|------|--------|--------|
| admin_tenant | 租户（含状态、配置） | ✅ | ✅ |
| admin_user | 用户（全局，不绑定租户，代表登录身份） | ✅ | ✅ |
| admin_user_tenant | 用户-租户关联 M:N | ❌ | ❌ |
| admin_role | 角色（含层级 parent_id，tenant_id 隔离） | ✅ | ✅ |
| admin_application | 应用实体 | ✅ | ✅ |
| admin_resource | 菜单/按钮资源树（type: MENU/BUTTON，含 app_code，tenant_id 隔离） | ✅ | ✅ |
| admin_api_permission | 接口权限树（type: GROUP/ENDPOINT，含 app_code，tenant_id 隔离） | ❌ | ❌ |
| admin_user_role | 用户-角色关联（含 tenant_id、临时授权时间窗口） | ❌ | ❌ |
| admin_role_resource | 角色-菜单/按钮关联 | ❌ | ❌ |
| admin_role_api | 角色-接口权限关联（仅叶子节点） | ❌ | ❌ |
| admin_role_app | 角色-应用关联 M:N | ❌ | ❌ |
| admin_data_scope | 数据权限维度配置（角色绑定） | ❌ | ❌ |
| admin_data_scope_config | 数据权限维度注册表 | ❌ | ❌ |
| admin_operation_log | 操作日志（业务含义） | ❌ | ❌ |

## 实施分阶段

| Phase | 内容 |
|-------|------|
| Phase 1 | 核心模型 + 租户管理 + 用户/角色 CRUD + JWT 认证 |
| Phase 2 | 菜单/按钮资源 + 接口权限树 + API 自动发现 |
| Phase 3 | 权限表达式引擎 + 缓存 + 中间件集成 |
| Phase 4 | 数据权限 + 应用隔离 + 组织 SPI |
| Phase 5 | 前端管理页面 |
| Phase 6 | 临时授权 + JWT 黑名单 + 高级功能 |

## 假设

- [假设-1] 现有 go-admin 的 sys_user/sys_role/sys_menu 表将被废弃，由新表完全替代
- [假设-2] 现有的 JWT 中间件和 Casbin 权限检查逻辑将被新模块替换
- [假设-3] 前端页面完全重做用户/角色/菜单/接口权限管理
- [假设-4] RBAC 模块启动时自动建表（GORM AutoMigrate）
- [假设-5] 平台管理员（SUPER_ADMIN）拥有全部权限，不受 RBAC 限制
