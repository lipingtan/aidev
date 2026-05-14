# 需求文档：用户权限管理模块（user）

| 字段 | 内容 |
|------|------|
| 文档名称 | 用户权限管理模块需求文档 |
| CR 编号 | CR02 |
| 版本 | 1.0 |
| 创建日期 | 2026-04-18 |
| 更新日期 | 2026-04-18 |
| 负责人 | 技术团队 |
| 状态 | 已确认 |

---

## 简介

本文档定义 SPMP 智慧物业管理平台用户权限管理模块（user）的需求。user 模块是所有业务模块的基础依赖，负责系统的身份认证和权限控制，为工单、缴费、公告、门禁等模块提供用户身份和权限支持。该模块采用 Spring Boot 2.4.4 + Java 1.8 技术栈，位于 `com.spmp.user` 包下，数据库表前缀为 `sys_`，API 路径前缀为 `/api/v1/user/`。

### 已有基础设施（common 模块 CR01 已完成）

以下能力已在 common 模块中实现，user 模块可直接使用：

| 组件 | 说明 |
|------|------|
| SecurityConfig | Spring Security 配置，JWT 无状态认证，白名单路径 |
| JwtTokenProvider | JWT 令牌生成/解析/验证，支持 admin（8h）/owner（7d）分端过期 |
| JwtAuthenticationFilter | JWT 认证过滤器 |
| DataPermissionInterceptor | 数据权限 SQL 拦截器（四级：全部→片区→小区→楼栋，**本 CR 需扩展为五级**） |
| BCryptPasswordEncoder | 密码编码器 |
| Result / PageResult | 统一响应包装 |
| ErrorCode / BusinessException | 错误码和业务异常 |
| GlobalExceptionHandler | 全局异常处理 |
| BaseEntity | 公共字段 + 自动填充 |
| RedisUtils | Redis 缓存工具 |
| EncryptUtils | 敏感数据加解密（AES） |

### 与参考系统的差异说明

domain-share 目录下的参考文档来自雨虹/YH 内部框架，以下概念**不适用于 SPMP**：

| 参考系统概念 | SPMP 处理方式 |
|-------------|--------------|
| tenant_id（多租户） | ❌ 不需要，SPMP 为单租户系统 |
| 内部员工/外部员工身份区分 | ❌ 不需要，SPMP 统一用户模型 |
| 365/SAP/OA 系统集成 | ❌ 不需要，SPMP 无外部用户数据同步 |
| 机构管理（organ） | ❌ 不需要，SPMP 通过数据权限关联片区/小区/楼栋 |
| 岗位管理（position） | ❌ 不需要 |
| 用户组管理（user_group） | ❌ 不需要 |
| Feign/HSF 远程调用 | ❌ 不需要，SPMP 为单体应用，模块间通过 api 包接口调用 |

---

## 术语表

- **User_Module**：用户权限管理模块，位于 `com.spmp.user` 包下，负责身份认证和权限控制
- **sys_user**：用户表，单表用户模型，包含账号信息和基本个人信息
- **sys_role**：角色表，定义系统角色及其数据权限级别
- **sys_menu**：菜单表，三级结构（目录→菜单→按钮），定义功能权限
- **sys_user_role**：用户-角色关联表，多对多关系
- **sys_role_menu**：角色-菜单关联表，多对多关系
- **sys_role_data_permission**：角色-数据权限关联表，记录角色关联的具体片区/小区/楼栋
- **sys_login_log**：登录日志表，记录登录时间、IP、登录结果
- **sys_operation_log**：操作日志表，记录关键业务操作
- **RBAC**：基于角色的访问控制（Role-Based Access Control），用户通过角色获得权限
- **数据权限五级**：全部数据→片区级数据→小区级数据→楼栋级数据→仅本人数据，五个层级的数据可见范围控制
- **权限标识**：按钮级权限的唯一标识，格式为 `模块:资源:操作`（如 `workorder:order:create`）
- **多角色并集**：一个用户拥有多个角色时，菜单权限和数据权限均取所有角色的并集
- **Access Token**：JWT 访问令牌，admin 端有效期 8 小时，owner 端有效期 7 天
- **Refresh Token**：JWT 刷新令牌，有效期为 Access Token 的 2 倍
- **客户端类型**：admin（PC 管理端）和 owner（H5 业主端）两种客户端类型
- **@PreAuthorize**：Spring Security 注解，用于后端接口层的权限标识校验
- **@DataPermission**：数据权限注解，标记需要数据权限控制的 Mapper 方法

---

## SPMP 用户角色

| 角色 | 角色编码 | 客户端 | 数据权限级别 | 说明 |
|------|---------|--------|-------------|------|
| 超级管理员 | super_admin | PC（admin） | 全部 | 系统全局管理 |
| 物业管理员 | property_admin | PC（admin） | 小区 | 所管辖小区的日常管理 |
| 片区经理 | area_manager | PC（admin） | 片区 | 所辖片区多个小区的监管 |
| 楼栋管家 | building_steward | PC（admin） | 楼栋 | 具体楼栋的日常管理 |
| 维修人员 | repairman | PC（admin） | 仅本人 | 处理分配的报修工单 |
| 业主 | owner | H5（owner） | 仅本人 | 移动端物业交互 |

---

## 需求

### 需求 R1：登录认证

**用户故事：** 作为系统用户（管理端或业主端），我希望通过安全的方式登录系统并获取 JWT Token，以便安全地访问系统功能。

#### R1.1 用户名密码登录

**用户故事：** 作为 PC 管理端用户，我希望通过用户名和密码登录系统，以便访问管理功能。

##### 验收标准

1. THE User_Module SHALL 提供用户名密码登录接口 `POST /api/v1/user/auth/login`，接受用户名、密码、验证码、验证码 key 和客户端类型参数
2. WHEN 用户提交正确的用户名、密码和验证码时，THE User_Module SHALL 返回 access_token 和 refresh_token
3. WHEN 用户名不存在时，THE User_Module SHALL 返回统一的"用户名或密码错误"提示，不区分用户名不存在和密码错误
4. WHEN 密码错误时，THE User_Module SHALL 记录错误次数到 Redis，key 格式为 `login:fail:{username}`
5. WHEN 同一用户连续密码错误达到 5 次时，THE User_Module SHALL 锁定该账号 30 分钟，锁定期间拒绝登录并返回"账号已锁定，请 {剩余分钟} 分钟后重试"
6. WHEN 用户状态为禁用时，THE User_Module SHALL 拒绝登录并返回"账号已被禁用"
7. WHEN 登录成功时，THE User_Module SHALL 清除该用户的密码错误计数
8. THE User_Module SHALL 在每次登录尝试（成功或失败）时记录登录日志

#### R1.2 手机号验证码登录

**用户故事：** 作为 H5 业主端用户，我希望通过手机号和短信验证码登录系统，以便快捷地访问业主功能。

##### 验收标准

1. THE User_Module SHALL 提供手机号验证码登录接口 `POST /api/v1/user/auth/login/sms`，接受手机号、短信验证码和客户端类型参数
2. THE User_Module SHALL 提供发送短信验证码接口 `POST /api/v1/user/auth/sms-code`，接受手机号参数
3. WHEN 请求发送短信验证码时，THE User_Module SHALL 生成 6 位数字验证码，存入 Redis 并设置 5 分钟有效期，key 格式为 `sms:code:{phone}`
4. WHEN 同一手机号 60 秒内重复请求发送验证码时，THE User_Module SHALL 拒绝发送并返回"请 {剩余秒数} 秒后重试"
5. WHEN 手机号和验证码匹配且未过期时，THE User_Module SHALL 返回 access_token 和 refresh_token
6. WHEN 验证码错误或已过期时，THE User_Module SHALL 返回"验证码错误或已过期"
7. WHEN 手机号未注册时，THE User_Module SHALL 返回"该手机号未注册"
8. THE User_Module SHALL 在验证码验证成功后立即删除 Redis 中的验证码，防止重复使用

#### R1.3 图形验证码

**用户故事：** 作为系统，我需要在登录前提供图形验证码，以便防止暴力破解和机器人攻击。

##### 验收标准

1. THE User_Module SHALL 提供图形验证码接口 `GET /api/v1/user/auth/captcha`，返回 Base64 编码的验证码图片和验证码 key
2. THE User_Module SHALL 生成 4 位字母数字混合的验证码，存入 Redis 并设置 2 分钟有效期，key 格式为 `captcha:{uuid}`
3. WHEN 用户名密码登录时，THE User_Module SHALL 先校验图形验证码，验证码错误则拒绝登录
4. THE User_Module SHALL 在验证码校验成功后立即删除 Redis 中的验证码，防止重复使用

#### R1.4 Token 刷新

**用户故事：** 作为已登录用户，我希望在 Access Token 过期前能自动刷新，以便不中断使用体验。

##### 验收标准

1. THE User_Module SHALL 提供 Token 刷新接口 `POST /api/v1/user/auth/refresh`，接受 refresh_token 参数
2. WHEN refresh_token 有效时，THE User_Module SHALL 返回新的 access_token 和 refresh_token
3. WHEN refresh_token 已过期或无效时，THE User_Module SHALL 返回 401 状态码，提示用户重新登录
4. THE User_Module SHALL 在签发新 Token 后将旧的 refresh_token 加入 Redis 黑名单，防止重复使用

#### R1.5 登出

**用户故事：** 作为已登录用户，我希望能安全退出系统，以便保护账号安全。

##### 验收标准

1. THE User_Module SHALL 提供登出接口 `POST /api/v1/user/auth/logout`
2. WHEN 用户登出时，THE User_Module SHALL 将当前 access_token 加入 Redis 黑名单，黑名单过期时间与 Token 剩余有效期一致
3. WHEN 用户登出时，THE User_Module SHALL 将对应的 refresh_token 加入 Redis 黑名单
4. THE User_Module SHALL 清除该用户在 Redis 中缓存的权限信息

---

### 需求 R2：用户管理

**用户故事：** 作为超级管理员或物业管理员，我希望能够管理系统用户账号，以便控制谁可以登录和使用系统。

#### R2.1 用户列表查询

##### 验收标准

1. THE User_Module SHALL 提供用户分页查询接口 `GET /api/v1/user/users`，支持按用户名、姓名、手机号、状态、角色进行条件筛选
2. THE User_Module SHALL 对查询结果应用数据权限过滤，不同角色只能看到其权限范围内的用户
3. THE User_Module SHALL 在返回结果中对手机号进行脱敏展示（如 138****1234）
4. THE User_Module SHALL 支持分页参数 pageNum 和 pageSize，默认 pageNum=1、pageSize=10

#### R2.2 新增用户

##### 验收标准

1. THE User_Module SHALL 提供新增用户接口 `POST /api/v1/user/users`，必填字段包括：用户名、姓名、手机号、所属角色 ID 列表
2. THE User_Module SHALL 校验用户名全局唯一（不区分大小写）
3. THE User_Module SHALL 校验手机号全局唯一
4. THE User_Module SHALL 使用 EncryptUtils 对手机号进行 AES 加密后存储
5. WHEN 新增用户时，THE User_Module SHALL 使用 BCryptPasswordEncoder 对默认密码 `Spmp@2026` 进行加密后存储
6. THE User_Module SHALL 在 sys_user_role 表中建立用户与角色的关联关系

#### R2.3 编辑用户

##### 验收标准

1. THE User_Module SHALL 提供编辑用户接口 `PUT /api/v1/user/users/{id}`，支持修改姓名、手机号、角色、状态等字段
2. THE User_Module SHALL 校验修改后的手机号不与其他用户重复
3. WHEN 修改用户角色时，THE User_Module SHALL 先删除原有角色关联，再建立新的角色关联
4. WHEN 修改用户角色时，THE User_Module SHALL 清除该用户在 Redis 中缓存的权限信息

#### R2.4 用户状态管理

##### 验收标准

1. THE User_Module SHALL 提供用户状态切换接口 `PUT /api/v1/user/users/{id}/status`，支持启用（0）和禁用（1）
2. WHEN 禁用用户时，THE User_Module SHALL 将该用户当前的 access_token 加入 Redis 黑名单，强制下线
3. THE User_Module SHALL 禁止禁用超级管理员账号
4. THE User_Module SHALL 禁止用户禁用自己的账号

#### R2.5 密码重置

##### 验收标准

1. THE User_Module SHALL 提供密码重置接口 `PUT /api/v1/user/users/{id}/reset-password`
2. WHEN 管理员重置用户密码时，THE User_Module SHALL 将密码重置为默认密码 `Spmp@2026`（BCrypt 加密存储）
3. THE User_Module SHALL 在密码重置后清除该用户的密码错误计数和账号锁定状态
4. THE User_Module SHALL 记录密码重置操作到操作日志

#### R2.6 删除用户

##### 验收标准

1. THE User_Module SHALL 提供删除用户接口 `DELETE /api/v1/user/users/{id}`，使用逻辑删除（del_flag=1）
2. THE User_Module SHALL 禁止删除超级管理员账号
3. THE User_Module SHALL 禁止用户删除自己的账号
4. WHEN 删除用户时，THE User_Module SHALL 同时删除该用户的角色关联关系，并清除 Redis 中的缓存

---

### 需求 R3：角色管理

**用户故事：** 作为超级管理员，我希望能够创建和配置角色，为角色分配菜单权限和数据权限范围，以便实现不同岗位的差异化权限控制。

#### R3.1 角色列表查询

##### 验收标准

1. THE User_Module SHALL 提供角色分页查询接口 `GET /api/v1/user/roles`，支持按角色名称、角色编码、状态进行条件筛选
2. THE User_Module SHALL 提供角色全量列表接口 `GET /api/v1/user/roles/list`（不分页），用于用户管理中的角色下拉选择
3. THE User_Module SHALL 支持分页参数 pageNum 和 pageSize，默认 pageNum=1、pageSize=10

#### R3.2 新增角色

##### 验收标准

1. THE User_Module SHALL 提供新增角色接口 `POST /api/v1/user/roles`，必填字段包括：角色名称、角色编码、数据权限级别
2. THE User_Module SHALL 校验角色编码全局唯一（不区分大小写）
3. THE User_Module SHALL 校验角色名称全局唯一
4. THE User_Module SHALL 支持的数据权限级别枚举值为：ALL（全部）、AREA（片区）、COMMUNITY（小区）、BUILDING（楼栋）、SELF（仅本人）

#### R3.3 编辑角色

##### 验收标准

1. THE User_Module SHALL 提供编辑角色接口 `PUT /api/v1/user/roles/{id}`，支持修改角色名称、数据权限级别、状态、备注等字段
2. THE User_Module SHALL 禁止修改角色编码（角色编码创建后不可变更）
3. THE User_Module SHALL 禁止修改预置的超级管理员角色（super_admin）
4. WHEN 修改角色的数据权限级别时，THE User_Module SHALL 清除该角色原有的数据权限关联数据

#### R3.4 删除角色

##### 验收标准

1. THE User_Module SHALL 提供删除角色接口 `DELETE /api/v1/user/roles/{id}`，使用逻辑删除
2. THE User_Module SHALL 禁止删除预置的超级管理员角色（super_admin）
3. WHEN 角色下存在关联用户时，THE User_Module SHALL 拒绝删除并返回"该角色下存在用户，无法删除"
4. WHEN 删除角色时，THE User_Module SHALL 同时删除该角色的菜单关联和数据权限关联

#### R3.5 角色菜单权限分配

##### 验收标准

1. THE User_Module SHALL 提供角色菜单权限分配接口 `PUT /api/v1/user/roles/{id}/menus`，接受菜单 ID 列表
2. THE User_Module SHALL 采用先删后增策略：先删除该角色原有的所有菜单关联，再批量插入新的关联
3. THE User_Module SHALL 提供查询角色已分配菜单 ID 列表接口 `GET /api/v1/user/roles/{id}/menus`
4. WHEN 角色菜单权限变更时，THE User_Module SHALL 清除该角色下所有用户在 Redis 中缓存的权限信息

#### R3.6 角色数据权限配置

##### 验收标准

1. THE User_Module SHALL 提供角色数据权限配置接口 `PUT /api/v1/user/roles/{id}/data-permission`，接受数据权限级别和关联的数据 ID 列表
2. WHEN 数据权限级别为 ALL（全部）或 SELF（仅本人）时，THE User_Module SHALL 不需要关联具体的数据 ID
3. WHEN 数据权限级别为 AREA（片区）时，THE User_Module SHALL 要求关联至少一个片区 ID
4. WHEN 数据权限级别为 COMMUNITY（小区）时，THE User_Module SHALL 要求关联至少一个小区 ID
5. WHEN 数据权限级别为 BUILDING（楼栋）时，THE User_Module SHALL 要求关联至少一个楼栋 ID
6. THE User_Module SHALL 采用先删后增策略更新数据权限关联
7. WHEN 角色数据权限变更时，THE User_Module SHALL 清除该角色下所有用户在 Redis 中缓存的数据权限信息

---

### 需求 R4：菜单管理

**用户故事：** 作为超级管理员，我希望能够管理系统菜单和按钮权限，以便控制不同角色可以看到和操作的功能。

#### R4.1 菜单树查询

##### 验收标准

1. THE User_Module SHALL 提供菜单树查询接口 `GET /api/v1/user/menus/tree`，返回完整的三级菜单树结构
2. THE User_Module SHALL 支持按菜单名称、状态进行条件筛选
3. THE User_Module SHALL 按 sort 字段升序排列同级菜单
4. THE User_Module SHALL 提供当前用户菜单树接口 `GET /api/v1/user/menus/user-tree`，仅返回当前登录用户有权限的菜单（用于前端动态路由）

#### R4.2 新增菜单

##### 验收标准

1. THE User_Module SHALL 提供新增菜单接口 `POST /api/v1/user/menus`，必填字段包括：菜单名称、菜单类型、排序号
2. THE User_Module SHALL 支持三级菜单类型：目录（D）、菜单（M）、按钮（B）
3. WHEN 菜单类型为目录（D）时，THE User_Module SHALL 要求填写菜单名称、路由路径、图标、排序号
4. WHEN 菜单类型为菜单（M）时，THE User_Module SHALL 要求填写菜单名称、路由路径、组件路径、权限标识、图标、排序号
5. WHEN 菜单类型为按钮（B）时，THE User_Module SHALL 要求填写菜单名称、权限标识、排序号
6. THE User_Module SHALL 校验按钮权限标识格式为 `模块:资源:操作`（如 `user:user:create`、`workorder:order:edit`）
7. THE User_Module SHALL 校验同一父级下菜单名称不重复

#### R4.3 编辑菜单

##### 验收标准

1. THE User_Module SHALL 提供编辑菜单接口 `PUT /api/v1/user/menus/{id}`
2. THE User_Module SHALL 禁止修改菜单类型（创建后不可变更）
3. THE User_Module SHALL 禁止将菜单的父级设置为自身或其子级（防止循环引用）

#### R4.4 删除菜单

##### 验收标准

1. THE User_Module SHALL 提供删除菜单接口 `DELETE /api/v1/user/menus/{id}`
2. WHEN 菜单存在子级菜单时，THE User_Module SHALL 拒绝删除并返回"存在子菜单，无法删除"
3. WHEN 菜单已被角色关联时，THE User_Module SHALL 拒绝删除并返回"该菜单已被角色使用，无法删除"

---

### 需求 R5：数据权限

**用户故事：** 作为系统，我需要根据用户的角色和数据权限配置，在 SQL 层面自动过滤查询结果，以便不同角色只能看到自己权限范围内的数据。

#### R5.1 五级数据权限模型

##### 验收标准

1. THE User_Module SHALL 支持五级数据权限模型：

| 级别 | 枚举值 | 说明 | SQL 过滤方式 |
|------|--------|------|-------------|
| 全部 | ALL | 可查看所有数据 | 不追加过滤条件 |
| 片区 | AREA | 可查看所辖片区数据 | `WHERE area_id IN (...)` |
| 小区 | COMMUNITY | 可查看所辖小区数据 | `WHERE community_id IN (...)` |
| 楼栋 | BUILDING | 可查看所辖楼栋数据 | `WHERE building_id IN (...)` |
| 仅本人 | SELF | 只能看自己的数据 | `WHERE create_by = {userId}` 或自定义字段 |

2. THE User_Module SHALL 扩展 common 模块的 DataPermissionInterceptor，增加"仅本人"（SELF）级别的支持
3. THE @DataPermission 注解 SHALL 增加 selfField 参数（默认 "create_by"），用于指定"仅本人"级别的过滤字段

#### R5.2 多角色数据权限合并

##### 验收标准

1. WHEN 用户拥有多个角色时，THE User_Module SHALL 将所有角色的数据权限取并集
2. IF 用户的任一角色拥有 ALL（全部）数据权限，THEN THE User_Module SHALL 不追加任何数据过滤条件
3. IF 用户的多个角色分别拥有不同片区/小区/楼栋的数据权限，THEN THE User_Module SHALL 合并所有数据 ID 作为过滤条件
4. IF 用户的任一角色拥有 SELF（仅本人）数据权限且无更高级别权限，THEN THE User_Module SHALL 按用户 ID 过滤

#### R5.3 数据权限缓存

##### 验收标准

1. THE User_Module SHALL 在用户登录时将其数据权限信息缓存到 Redis，key 格式为 `user:data-permission:{userId}`
2. THE User_Module SHALL 在角色数据权限变更时，清除受影响用户的数据权限缓存
3. THE DataPermissionInterceptor SHALL 优先从 Redis 缓存获取数据权限信息，缓存未命中时从数据库加载并写入缓存

#### R5.4 数据权限基础数据

##### 验收标准

1. THE User_Module SHALL 在本期使用硬编码/配置文件方式提供片区、小区、楼栋的基础数据，用于数据权限配置时的选择
2. THE User_Module SHALL 预留与 base 模块（`com.spmp.base`）对接的接口，后续通过 BaseApi 获取真实基础数据
3. THE User_Module SHALL 提供数据权限可选范围查询接口 `GET /api/v1/user/data-permission/options`，返回可选的片区/小区/楼栋列表

---

### 需求 R6：个人中心

**用户故事：** 作为已登录用户，我希望能够查看自己的个人信息和修改密码，以便维护自己的账号安全。

#### R6.1 个人信息查看

##### 验收标准

1. THE User_Module SHALL 提供获取当前用户信息接口 `GET /api/v1/user/profile`，返回用户基本信息、角色列表、菜单权限列表、按钮权限标识列表
2. THE User_Module SHALL 在返回结果中对手机号进行脱敏展示
3. THE User_Module SHALL 返回用户的数据权限级别和关联的数据范围信息

#### R6.2 修改密码

##### 验收标准

1. THE User_Module SHALL 提供修改密码接口 `PUT /api/v1/user/profile/password`，接受旧密码和新密码参数
2. THE User_Module SHALL 验证旧密码正确后才允许修改
3. THE User_Module SHALL 校验新密码强度：至少 8 位，必须包含字母和数字
4. THE User_Module SHALL 校验新密码不能与旧密码相同
5. WHEN 密码修改成功后，THE User_Module SHALL 将该用户当前的 access_token 和 refresh_token 加入 Redis 黑名单，强制重新登录
6. THE User_Module SHALL 记录密码修改操作到操作日志

#### R6.3 修改个人信息

##### 验收标准

1. THE User_Module SHALL 提供修改个人信息接口 `PUT /api/v1/user/profile`，支持修改姓名、手机号等非敏感字段
2. THE User_Module SHALL 校验修改后的手机号不与其他用户重复
3. THE User_Module SHALL 使用 EncryptUtils 对修改后的手机号进行加密存储

---

### 需求 R7：对外 API

**用户故事：** 作为其他业务模块的开发人员，我希望 user 模块提供标准的内部 API 接口，以便在工单、缴费、公告等模块中查询用户信息和校验权限。

#### R7.1 用户查询 API

##### 验收标准

1. THE User_Module SHALL 在 `com.spmp.user.api` 包下提供 `UserApi` 接口
2. THE UserApi SHALL 提供 `getUserById(Long userId)` 方法，返回 `UserBriefDTO`（包含 id、用户名、姓名、手机号脱敏）
3. THE UserApi SHALL 提供 `getUsersByRoleCode(String roleCode)` 方法，返回指定角色下的用户列表（如查询所有维修人员）
4. THE UserApi SHALL 提供 `getUsersByIds(List<Long> userIds)` 方法，批量查询用户信息

#### R7.2 权限查询 API

##### 验收标准

1. THE User_Module SHALL 在 `com.spmp.user.api` 包下提供 `PermissionApi` 接口
2. THE PermissionApi SHALL 提供 `getDataPermission(Long userId)` 方法，返回用户的数据权限级别和关联的数据 ID 列表
3. THE PermissionApi SHALL 提供 `checkPermission(Long userId, String permissionCode)` 方法，校验用户是否拥有指定的权限标识
4. THE PermissionApi SHALL 提供 `getUserRoles(Long userId)` 方法，返回用户的角色编码列表

#### R7.3 API 实现约束

##### 验收标准

1. THE UserApi 和 PermissionApi SHALL 由 user 模块的 Service 层实现类同时实现（单体应用内直接注入）
2. THE API 接口的 DTO 类 SHALL 定义在 `com.spmp.user.api.dto` 包下
3. 其他模块 SHALL 通过注入 UserApi 或 PermissionApi 接口调用，禁止直接注入 user 模块的 Service 实现类

---

### 需求 R8：登录日志

**用户故事：** 作为超级管理员，我希望能够查看系统的登录日志，以便监控系统安全状况和排查异常登录行为。

#### R8.1 登录日志记录

##### 验收标准

1. THE User_Module SHALL 在每次登录尝试时自动记录登录日志到 sys_login_log 表
2. THE 登录日志 SHALL 包含以下字段：用户名、登录 IP、登录地点（可选，基于 IP 解析）、浏览器/客户端信息、操作系统、登录时间、登录结果（成功/失败）、失败原因（如密码错误、账号锁定、账号禁用）
3. THE User_Module SHALL 通过异步方式记录登录日志，不影响登录接口的响应时间

#### R8.2 登录日志查询

##### 验收标准

1. THE User_Module SHALL 提供登录日志分页查询接口 `GET /api/v1/user/login-logs`，支持按用户名、登录 IP、登录结果、时间范围进行条件筛选
2. THE User_Module SHALL 支持分页参数 pageNum 和 pageSize，默认 pageNum=1、pageSize=10
3. THE User_Module SHALL 按登录时间倒序排列查询结果
4. THE 登录日志查询接口 SHALL 仅超级管理员可访问，使用 `@PreAuthorize` 注解校验权限

---

### 需求 R9：操作日志

**用户故事：** 作为超级管理员，我希望能够查看系统的关键操作日志，以便审计和追溯重要操作。

#### R9.1 操作日志记录

##### 验收标准

1. THE User_Module SHALL 提供 `@OperationLog` 自定义注解，用于标记需要记录操作日志的方法
2. THE @OperationLog 注解 SHALL 支持以下参数：模块名称（module）、操作类型（type，如新增/修改/删除/导出）、操作描述（description）
3. THE User_Module SHALL 通过 AOP 切面拦截标记了 @OperationLog 的方法，自动记录操作日志到 sys_operation_log 表
4. THE 操作日志 SHALL 包含以下字段：操作人 ID、操作人用户名、模块名称、操作类型、操作描述、请求方法（GET/POST/PUT/DELETE）、请求 URL、请求参数、响应结果、操作 IP、操作时间、耗时（ms）
5. THE User_Module SHALL 通过异步方式记录操作日志，不影响业务接口的响应时间
6. THE User_Module SHALL 对以下关键操作记录操作日志：
   - 用户管理：新增用户、编辑用户、删除用户、重置密码、启用/禁用用户
   - 角色管理：新增角色、编辑角色、删除角色、分配菜单权限、配置数据权限
   - 菜单管理：新增菜单、编辑菜单、删除菜单

#### R9.2 操作日志查询

##### 验收标准

1. THE User_Module SHALL 提供操作日志分页查询接口 `GET /api/v1/user/operation-logs`，支持按操作人、模块名称、操作类型、时间范围进行条件筛选
2. THE User_Module SHALL 支持分页参数 pageNum 和 pageSize，默认 pageNum=1、pageSize=10
3. THE User_Module SHALL 按操作时间倒序排列查询结果
4. THE 操作日志查询接口 SHALL 仅超级管理员可访问，使用 `@PreAuthorize` 注解校验权限

---

## 非功能性需求

### NFR-1：性能要求

#### 验收标准

1. THE 登录接口（R1.1、R1.2）的响应时间 SHALL 小于 500ms（P99）
2. THE 权限校验（JwtAuthenticationFilter + DataPermissionInterceptor）增加的延迟 SHALL 小于 5ms
3. THE 用户列表分页查询（R2.1）的响应时间 SHALL 小于 200ms（P99，1000 条数据量级）
4. THE 登录接口 SHALL 实施限流策略：同一 IP 每分钟最多 10 次请求，超出后返回 429 状态码
5. THE 短信验证码发送接口 SHALL 实施限流策略：同一手机号每天最多 10 次

### NFR-2：安全要求

#### 验收标准

1. THE User_Module SHALL 使用 BCryptPasswordEncoder 对用户密码进行加密存储，禁止明文存储
2. THE User_Module SHALL 使用 EncryptUtils（AES 加密）对手机号进行加密存储
3. THE User_Module SHALL 对所有需要权限控制的接口使用 `@PreAuthorize` 注解进行后端权限标识校验，权限标识格式为 `模块:资源:操作`
4. THE User_Module SHALL 实施 CSRF 防护，对非 GET 请求进行 CSRF Token 校验（需更新 common 模块 SecurityConfig 的 CSRF 配置策略）
5. THE User_Module SHALL 对密码强度进行校验：至少 8 位，必须包含字母和数字
6. THE User_Module SHALL 实施账号锁定策略：连续 5 次密码错误锁定 30 分钟
7. THE 登录接口 SHALL 不区分"用户名不存在"和"密码错误"的错误提示，统一返回"用户名或密码错误"
8. THE User_Module SHALL 对登录接口要求图形验证码校验，防止暴力破解

### NFR-3：可用性要求

#### 验收标准

1. THE User_Module SHALL 将用户权限信息缓存到 Redis，减少数据库查询压力
2. THE User_Module SHALL 在权限数据变更时主动清除相关缓存，保证数据一致性
3. THE User_Module SHALL 在 Redis 不可用时降级为直接查询数据库，不影响核心认证功能

### NFR-4：可维护性要求

#### 验收标准

1. THE User_Module SHALL 遵循 DDD 模块化分层结构：controller → service → repository（mapper），domain 包含 entity/dto/vo
2. THE User_Module SHALL 将错误码定义在 ErrorCode 枚举的 user 模块段（2000-2999）
3. THE User_Module SHALL 对所有公共方法编写 Javadoc 注释
4. THE User_Module 的核心业务逻辑单元测试覆盖率 SHALL 不低于 90%

---

## 数据模型概要

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| sys_user | 用户表 | id, username, password, real_name, phone(加密), avatar, status, del_flag |
| sys_role | 角色表 | id, role_name, role_code, data_permission_level, status, sort, remark, del_flag |
| sys_menu | 菜单表 | id, menu_name, parent_id, menu_type(D/M/B), path, component, permission, icon, sort, status, del_flag |
| sys_user_role | 用户-角色关联表 | id, user_id, role_id |
| sys_role_menu | 角色-菜单关联表 | id, role_id, menu_id |
| sys_role_data_permission | 角色-数据权限关联表 | id, role_id, data_type(AREA/COMMUNITY/BUILDING), data_id |
| sys_login_log | 登录日志表 | id, username, login_ip, login_location, browser, os, login_time, login_result, fail_reason |
| sys_operation_log | 操作日志表 | id, operator_id, operator_name, module, operation_type, description, request_method, request_url, request_params, response_result, operation_ip, operation_time, cost_time |

---

## 接口需求概要

### 认证接口（白名单，无需 Token）

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/user/auth/captcha` | 获取图形验证码 | 公开 |
| POST | `/api/v1/user/auth/login` | 用户名密码登录 | 公开 |
| POST | `/api/v1/user/auth/login/sms` | 手机号验证码登录 | 公开 |
| POST | `/api/v1/user/auth/sms-code` | 发送短信验证码 | 公开 |
| POST | `/api/v1/user/auth/refresh` | 刷新 Token | 公开 |
| POST | `/api/v1/user/auth/logout` | 登出 | 需认证 |

### 用户管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/users` | 用户分页查询 | user:user:list |
| POST | `/api/v1/user/users` | 新增用户 | user:user:create |
| PUT | `/api/v1/user/users/{id}` | 编辑用户 | user:user:edit |
| DELETE | `/api/v1/user/users/{id}` | 删除用户 | user:user:delete |
| PUT | `/api/v1/user/users/{id}/status` | 用户状态切换 | user:user:edit |
| PUT | `/api/v1/user/users/{id}/reset-password` | 重置密码 | user:user:reset-pwd |

### 角色管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/roles` | 角色分页查询 | user:role:list |
| GET | `/api/v1/user/roles/list` | 角色全量列表 | user:role:list |
| POST | `/api/v1/user/roles` | 新增角色 | user:role:create |
| PUT | `/api/v1/user/roles/{id}` | 编辑角色 | user:role:edit |
| DELETE | `/api/v1/user/roles/{id}` | 删除角色 | user:role:delete |
| GET | `/api/v1/user/roles/{id}/menus` | 查询角色菜单 | user:role:list |
| PUT | `/api/v1/user/roles/{id}/menus` | 分配菜单权限 | user:role:assign |
| PUT | `/api/v1/user/roles/{id}/data-permission` | 配置数据权限 | user:role:assign |

### 菜单管理接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/menus/tree` | 菜单树查询 | user:menu:list |
| GET | `/api/v1/user/menus/user-tree` | 当前用户菜单树 | 需认证 |
| POST | `/api/v1/user/menus` | 新增菜单 | user:menu:create |
| PUT | `/api/v1/user/menus/{id}` | 编辑菜单 | user:menu:edit |
| DELETE | `/api/v1/user/menus/{id}` | 删除菜单 | user:menu:delete |

### 个人中心接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/user/profile` | 获取个人信息 | 需认证 |
| PUT | `/api/v1/user/profile` | 修改个人信息 | 需认证 |
| PUT | `/api/v1/user/profile/password` | 修改密码 | 需认证 |

### 数据权限接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/data-permission/options` | 数据权限可选范围 | user:role:assign |

### 日志查询接口

| 方法 | 路径 | 说明 | 权限标识 |
|------|------|------|---------|
| GET | `/api/v1/user/login-logs` | 登录日志查询 | user:log:list |
| GET | `/api/v1/user/operation-logs` | 操作日志查询 | user:log:list |

---

## 对 common 模块的变更需求

本 CR 需要对 common 模块（CR01 已完成）进行以下扩展：

| 变更项 | 说明 |
|--------|------|
| DataPermissionInterceptor | 增加"仅本人"（SELF）数据权限级别支持 |
| @DataPermission 注解 | 增加 selfField 参数，默认值 "create_by" |
| DataPermissionContext | 增加 SELF 级别的上下文数据（当前用户 ID） |
| SecurityConfig | 更新 CSRF 防护策略（从完全禁用改为对非 GET 请求启用 CSRF Token 校验） |
| ErrorCode | 在 user 模块段（2000-2999）增加用户权限相关错误码 |

---

## 对外提供的 API（供其他模块调用）

| 接口 | 方法 | 说明 | 调用方 |
|------|------|------|--------|
| UserApi | `getUserById(Long userId)` | 查询用户基本信息 | workorder（查维修人员）、notice、billing |
| UserApi | `getUsersByRoleCode(String roleCode)` | 按角色查用户列表 | workorder（查维修人员列表） |
| UserApi | `getUsersByIds(List<Long> userIds)` | 批量查询用户信息 | 所有业务模块 |
| PermissionApi | `getDataPermission(Long userId)` | 获取用户数据权限范围 | 所有业务模块 |
| PermissionApi | `checkPermission(Long userId, String permCode)` | 校验用户权限标识 | 所有业务模块 |
| PermissionApi | `getUserRoles(Long userId)` | 获取用户角色列表 | 所有业务模块 |

---

## 前端需求概要

### PC 管理端页面

| 页面 | 路由 | 功能 |
|------|------|------|
| 登录页 | `/login` | 用户名密码登录 + 图形验证码 |
| 用户管理 | `/system/users` | 用户 CRUD、角色分配、状态管理、密码重置 |
| 角色管理 | `/system/roles` | 角色 CRUD、菜单权限树勾选、数据权限配置 |
| 菜单管理 | `/system/menus` | 菜单树管理、按钮权限配置 |
| 登录日志 | `/system/login-logs` | 登录日志查询 |
| 操作日志 | `/system/operation-logs` | 操作日志查询 |
| 个人中心 | `/profile` | 个人信息查看、修改密码 |

交互要求：
- 菜单权限使用树形复选框，支持全选/半选
- 数据权限配置：选择级别后，弹出对应的片区/小区/楼栋选择器
- 侧边栏菜单根据当前用户权限动态渲染（通过 `GET /api/v1/user/menus/user-tree` 获取）
- 按钮权限通过 `v-permission` 指令控制显隐
- 所有 `system/` 前缀路由在 PC 侧边栏中归入"系统管理"分组子菜单（图标：Setting），`log/` 前缀归入"日志管理"分组（图标：Document），分组规范见 `CUST-AIDEV-FRONTEND-DOCS/global-info/frontend-pages-route.md`

### H5 业主端页面

| 页面 | 路由 | 功能 |
|------|------|------|
| 登录页 | `/login` | 同时支持用户名密码登录和手机号验证码登录（Tab 切换） |
| 个人中心 | `/mine` | 查看个人信息、修改密码 |

---

## 技术约束汇总

| 约束项 | 说明 |
|--------|------|
| 包名 | `com.spmp.user` |
| 表前缀 | `sys_` |
| API 前缀 | `/api/v1/user/` |
| 架构 | 单体应用，模块间通过 api 包接口调用 |
| Token 有效期 | admin: access 8h / refresh 16h；owner: access 7d / refresh 14d |
| 密码策略 | 至少 8 位含字母数字，默认密码 Spmp@2026，5 次错误锁定 30 分钟 |
| 数据权限 | 五级（全部/片区/小区/楼栋/仅本人），多角色取并集 |
| 权限标识 | `模块:资源:操作` 格式，前端 v-permission + 后端 @PreAuthorize 双重校验 |
| 敏感数据 | 手机号 AES 加密存储，密码 BCrypt 加密存储 |
| 基础数据 | 本期硬编码/配置文件，后续对接 base 模块 |
| 日志 | 登录日志 + 操作日志，异步记录 |
