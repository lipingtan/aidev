# 实施计划：用户权限管理模块（user）

## 概述

本计划将 user 模块的 9 个需求模块（R1-R9）及非功能性需求拆解为可执行的编码任务。采用分层递进策略：先变更 common 模块基础设施，再执行数据库 DDL/DML，然后按依赖关系实现各业务组件。认证模块优先（前后端联调关键路径），最后实现日志、限流等增强功能。

**实现语言：** Java 1.8
**技术栈：** Spring Boot 2.4.4 + MyBatis Plus 3.4.x + Redis 7.2.1
**代码根目录：** `src/spmp-backend/`
**模块包名：** `com.spmp.user`（user 模块）、`com.spmp.common`（common 模块）

## 任务依赖关系

```
任务1 (common 模块变更)
  ├── 任务2 (数据库 DDL + DML)
  │     ├── 任务3 (DO 实体 + Mapper)                      [可并行组A]
  │     ├── 任务4 (错误码 + 常量 + 枚举)                   [可并行组A]
  │     │
  │     ├── 任务5 (权限缓存服务 PermissionCacheService)    [依赖任务3,4]
  │     ├── 任务6 (认证模块 AuthService)                   [依赖任务3,4,5] ← 关键路径
  │     │
  │     ├── 任务7 (用户管理 UserService)                   [可并行组B, 依赖任务5,6]
  │     ├── 任务8 (角色管理 RoleService)                   [可并行组B, 依赖任务5]
  │     ├── 任务9 (菜单管理 MenuService)                   [可并行组B, 依赖任务5]
  │     │
  │     ├── 任务10 (数据权限服务 DataPermissionService)    [依赖任务8]
  │     ├── 任务11 (个人中心 ProfileService)               [依赖任务5,6]
  │     │
  │     ├── 任务12 (对外 API: UserApi + PermissionApi)     [可并行组C, 依赖任务7,10]
  │     ├── 任务13 (操作日志 AOP)                          [可并行组C, 依赖任务3]
  │     ├── 任务14 (限流 AOP)                              [可并行组C, 依赖任务4]
  │     │
  │     ├── 任务15 (登录日志模块)                           [依赖任务6,13]
  │     └── 任务16 (Controller 层 + 安全配置)              [依赖任务6-14]
  │
  任务17 (配置文件汇总)                                     [依赖任务1-16]
  任务18 (最终检查点)
```

## 任务列表

- [x] 1. common 模块变更（DataPermissionContext 通用化、SELF 级别、JwtTokenProvider jti、黑名单校验）
  - [x] 1.1 DataPermissionLevel 枚举增加 SELF 值
    - 在 `com.spmp.common.security.DataPermissionLevel` 枚举中增加 `SELF` 枚举值
    - _需求: R5.1_

  - [x] 1.2 DataPermissionContext 重构为通用化设计
    - 重构 `com.spmp.common.security.DataPermissionContext`，移除 areaId/communityId/buildingId 单值字段
    - 新增 `Map<String, Set<Long>> scopeMap` 字段（key=数据库字段名，value=ID 集合）
    - 新增 `Long userId` 字段（SELF 级别使用）
    - 新增 `String username` 字段（SELF 级别 create_by 过滤使用）
    - _需求: R5.1, R5.2, DQ9.3_

  - [x] 1.3 @DataPermission 注解增加 selfField 参数
    - 在 `com.spmp.common.security.DataPermission` 注解中增加 `selfField()` 参数，默认值 `"create_by"`
    - _需求: R5.1_

  - [x] 1.4 DataPermissionInterceptor 扩展支持 SELF 级别和 scopeMap
    - 支持 SELF 级别：追加 `{selfField} = '{username}'` SQL 条件
    - 支持 scopeMap 遍历：动态拼接 `field IN (...)` 条件，多个条件用 OR 连接
    - 将原有的 `= #{id}` 单值过滤改为 `IN (...)` 集合过滤
    - _需求: R5.1, R5.2, DQ9.2_

  - [x] 1.5 JwtTokenProvider 增加 jti 支持
    - `generateToken()` 方法中增加 `setId(UUID)` 设置 jti
    - `generateRefreshToken()` 方法中增加 `setId(UUID)` 设置 jti
    - _需求: R1.4, R1.5, DQ4.3_

  - [x] 1.6 JwtAuthenticationFilter 增加 Token 黑名单校验
    - 在 `doFilterInternal()` 中解析 Token 后，校验 `token:blacklist:{jti}` 是否存在于 Redis
    - Redis 不可用时降级放行 + WARN 日志（DQ5.2 方案 C）
    - _需求: R1.5, NFR-3_

  - [x] 1.7 SecurityConfig 白名单追加 user 模块认证路径
    - 追加 `/api/v1/user/auth/login`、`/api/v1/user/auth/login/sms`、`/api/v1/user/auth/sms-code`、`/api/v1/user/auth/captcha`、`/api/v1/user/auth/refresh`
    - _需求: R1_

  - [x] 1.8 ErrorCode 枚举增加 user 模块错误码段（2000-2999）
    - 在 `com.spmp.common.exception.ErrorCode` 中预留 user 模块码段注释
    - _需求: NFR-4_

  - [ ]* 1.9 编写 common 模块变更单元测试
    - **Property 17: 数据权限多角色并集合并 — scopeMap 正确构建**
    - **Property 18: SELF 级别数据过滤 — SQL 追加 selfField 条件**
    - **Property 19: DataPermissionInterceptor scopeMap 动态拼接**
    - **Property 25: Token 黑名单降级放行**
    - _验证需求: R5.1, R5.2, NFR-3_


- [x] 2. 数据库 DDL + DML（建表、索引、初始化数据）
  - [x] 2.1 创建 DDL 脚本
    - 创建 `src/spmp-backend/src/main/resources/db/migration/` 目录
    - 编写 8 张表的 DDL：sys_user、sys_role、sys_menu、sys_user_role、sys_role_menu、sys_role_data_permission、sys_login_log、sys_operation_log
    - 包含所有索引（唯一索引、组合索引、普通索引）
    - _需求: R1-R9 数据模型_

  - [x] 2.2 创建 DML 初始化脚本
    - 超级管理员账号（admin / Spmp@2026 BCrypt 加密）
    - 6 个预置角色（super_admin、property_admin、area_manager、building_steward、repairman、owner）
    - 完整菜单树（系统管理、日志管理 + 二级菜单 + 三级按钮权限）
    - 管理员-角色关联、超级管理员角色-菜单关联
    - 测试用基础数据（3 片区、5 小区、8 楼栋）+ 角色数据权限示例
    - _需求: DQ1.5_

- [x] 3. DO 实体类 + Mapper 接口（可并行组A-1）
  - [x] 3.1 实现 DO 实体类
    - 创建 `com.spmp.user.domain.entity` 包
    - 实现 UserDO（继承 BaseEntity）：username、password、realName、phone、phoneHash、avatar、status
    - 实现 RoleDO（继承 BaseEntity）：roleName、roleCode、dataPermissionLevel、status、sort、remark
    - 实现 MenuDO（继承 BaseEntity）：menuName、parentId、menuType、path、component、permission、icon、sort、status
    - 实现 UserRoleDO：id、userId、roleId（不继承 BaseEntity）
    - 实现 RoleMenuDO：id、roleId、menuId
    - 实现 RoleDataPermissionDO：id、roleId、dataType、dataId
    - 实现 LoginLogDO：id、username、loginIp、loginLocation、browser、os、loginTime、loginResult、failReason
    - 实现 OperationLogDO：id、operatorId、operatorName、module、operationType、description、requestMethod、requestUrl、requestParams、responseResult、operationIp、operationTime、costTime
    - _需求: R1-R9 数据模型_

  - [x] 3.2 实现 Mapper 接口
    - 创建 `com.spmp.user.repository` 包
    - UserMapper extends BaseMapper<UserDO>：自定义 selectByUsername、selectByPhoneHash、selectPageWithFilter
    - RoleMapper extends BaseMapper<RoleDO>：自定义 selectByRoleCode
    - MenuMapper extends BaseMapper<MenuDO>：自定义 selectByParentId、selectPermissionsByUserId
    - UserRoleMapper extends BaseMapper<UserRoleDO>：自定义 selectRoleIdsByUserId、selectUserIdsByRoleId、deleteByUserId
    - RoleMenuMapper extends BaseMapper<RoleMenuDO>：自定义 selectMenuIdsByRoleId、deleteByRoleId
    - RoleDataPermissionMapper extends BaseMapper<RoleDataPermissionDO>：自定义 selectByRoleId、deleteByRoleId
    - LoginLogMapper extends BaseMapper<LoginLogDO>
    - OperationLogMapper extends BaseMapper<OperationLogDO>
    - 创建对应的 XML 映射文件（`resources/mapper/`）
    - _需求: R1-R9_


- [x] 4. 错误码 + 常量 + 枚举定义（可并行组A-2）
  - [x] 4.1 实现 UserErrorCode 错误码枚举
    - 创建 `com.spmp.user.constant.UserErrorCode`
    - 认证相关（2000-2099）：USER_NOT_FOUND、PASSWORD_ERROR、ACCOUNT_LOCKED、ACCOUNT_DISABLED、CAPTCHA_ERROR、CAPTCHA_EXPIRED、SMS_CODE_ERROR、PHONE_NOT_REGISTERED、SMS_SEND_TOO_FREQUENT、SMS_DAILY_LIMIT、TOKEN_INVALID、TOKEN_EXPIRED
    - 用户管理（2100-2199）：USERNAME_EXISTS、PHONE_EXISTS、CANNOT_DELETE_ADMIN、CANNOT_DISABLE_SELF、CANNOT_DELETE_SELF、CANNOT_DISABLE_ADMIN
    - 角色管理（2200-2299）：ROLE_CODE_EXISTS、ROLE_NAME_EXISTS、ROLE_HAS_USERS、CANNOT_MODIFY_ADMIN_ROLE、CANNOT_DELETE_ADMIN_ROLE
    - 菜单管理（2300-2399）：MENU_HAS_CHILDREN、MENU_USED_BY_ROLE、MENU_PARENT_INVALID、MENU_NAME_DUPLICATE
    - 个人中心（2400-2499）：OLD_PASSWORD_ERROR、NEW_PASSWORD_SAME、PASSWORD_TOO_WEAK
    - 限流（2900-2999）：RATE_LIMIT_EXCEEDED
    - _需求: NFR-4, design.md UserErrorCode 定义_

  - [x] 4.2 实现 UserConstants 常量类
    - 创建 `com.spmp.user.constant.UserConstants`
    - 默认密码、锁定时间、最大失败次数、Redis key 前缀、超级管理员角色编码等
    - _需求: R1.1, R2.5_

  - [x] 4.3 实现 OperationType 操作类型枚举
    - 创建 `com.spmp.user.constant.OperationType`：CREATE、UPDATE、DELETE、EXPORT、OTHER
    - _需求: R9.1_

- [x] 5. 权限缓存服务 PermissionCacheService（依赖任务3、4）
  - [x] 5.1 实现 PermissionCacheService 接口和实现类
    - 创建 `com.spmp.user.service.PermissionCacheService` 接口
    - 实现 `getUserPermissions(Long userId)` — 获取用户权限标识集合（优先 Redis，降级查 DB）
    - 实现 `getDataPermission(Long userId)` — 获取用户数据权限上下文（优先 Redis，降级查 DB）
    - 实现 `getUserMenuTree(Long userId)` — 获取用户菜单树（优先 Redis，降级查 DB）
    - 实现 `clearUserPermissions(Long userId)` — 清除用户权限缓存
    - 实现 `clearUserDataPermission(Long userId)` — 清除用户数据权限缓存
    - 实现 `clearUserMenus(Long userId)` — 清除用户菜单缓存
    - 实现 `clearRoleUsersCache(Long roleId)` — 清除角色下所有用户的缓存
    - Redis 不可用时降级查数据库 + WARN 日志
    - _需求: R5.3, NFR-3_

  - [x] 5.2 实现 PermissionService 权限校验 Bean
    - 创建 `com.spmp.user.security.PermissionService`，注册为 `@Component("perm")`
    - 实现 `check(String permissionCode)` 方法
    - 从 SecurityContext 获取当前用户 ID → 超级管理员直接返回 true → 从缓存获取权限列表 → 判断是否包含目标权限
    - _需求: R3.5, DQ4.2_

  - [ ]* 5.3 编写 PermissionCacheService 单元测试
    - **Property 26: 权限缓存降级查数据库**
    - **Property 14: 权限变更清除缓存**
    - _验证需求: R5.3, NFR-3_


- [x] 6. 认证模块 AuthService（依赖任务3、4、5）← 关键路径
  - [x] 6.1 实现 DTO 类
    - 创建 `com.spmp.user.domain.dto` 包
    - LoginDTO：username、password、captchaCode、captchaKey、clientType（含 @NotBlank 校验）
    - SmsLoginDTO：phone、smsCode、clientType
    - TokenDTO：accessToken、refreshToken、userId、username、realName、avatar
    - CaptchaDTO：captchaKey、captchaImage
    - _需求: R1.1, R1.2, R1.3_

  - [x] 6.2 实现 SmsService 接口 + SmsServiceMockImpl 模拟实现
    - 创建 `com.spmp.user.service.SmsService` 接口：`sendVerificationCode(String phone, String code)`
    - 创建 `com.spmp.user.service.impl.SmsServiceMockImpl`：日志输出验证码，不实际发送
    - _需求: R1.2, DQ6.1_

  - [x] 6.3 实现 AuthService 接口
    - 创建 `com.spmp.user.service.AuthService` 接口
    - 方法：login、loginBySms、sendSmsCode、refreshToken、logout、getCaptcha
    - _需求: R1_

  - [x] 6.4 实现 AuthServiceImpl — 图形验证码
    - 使用 Hutool CaptchaUtil.createLineCaptcha 生成 4 位字母数字验证码
    - 存入 Redis `captcha:{uuid}`，TTL 2 分钟
    - 返回 Base64 图片 + captchaKey
    - _需求: R1.3_

  - [x] 6.5 实现 AuthServiceImpl — 用户名密码登录
    - 校验图形验证码（Redis 取值比对，成功后删除）
    - 检查账号锁定（`login:lock:{username}`）
    - 查询用户（按 username）
    - 校验密码（BCryptPasswordEncoder.matches）
    - 密码错误：递增 `login:fail:{username}`，达 5 次设置 `login:lock:{username}` TTL 30 分钟
    - 检查用户状态（status=1 禁用）
    - 登录成功：清除错误计数、生成 Token（含 jti）、缓存权限到 Redis、异步记录登录日志
    - _需求: R1.1_

  - [x] 6.6 实现 AuthServiceImpl — 手机号验证码登录
    - 校验短信验证码（Redis `sms:code:{phone_hash}` 比对，成功后删除）
    - 按 phone_hash 查询用户，不存在返回 PHONE_NOT_REGISTERED
    - 生成 Token、缓存权限、异步记录登录日志
    - _需求: R1.2_

  - [x] 6.7 实现 AuthServiceImpl — 发送短信验证码
    - 校验 60 秒发送间隔（`sms:interval:{phone_hash}`）
    - 校验每日发送上限 10 次（`sms:daily:{phone_hash}`）
    - 生成 6 位随机数字验证码
    - 存入 Redis `sms:code:{phone_hash}`，TTL 5 分钟
    - 设置发送间隔标记 TTL 60 秒
    - 递增每日计数
    - 调用 SmsService.sendVerificationCode
    - _需求: R1.2_

  - [x] 6.8 实现 AuthServiceImpl — Token 刷新
    - 校验 refreshToken 有效性（validateToken + 检查 tokenType=refresh）
    - 校验 refreshToken 不在黑名单中
    - 生成新的 accessToken + refreshToken
    - 将旧 refreshToken 的 jti 加入黑名单（TTL = 剩余有效期）
    - _需求: R1.4_

  - [x] 6.9 实现 AuthServiceImpl — 登出
    - 解析当前 accessToken 获取 jti，加入黑名单
    - 解析关联的 refreshToken jti，加入黑名单
    - 清除用户权限缓存
    - _需求: R1.5_

  - [ ]* 6.10 编写 AuthService 单元测试
    - **Property 1: 登录成功返回有效 Token**
    - **Property 2: 密码错误计数递增且达到阈值锁定**
    - **Property 3: 登录成功清除错误计数**
    - **Property 4: 图形验证码一次性使用**
    - **Property 5: 短信验证码一次性使用且有时效**
    - **Property 6: Token 刷新后旧 RefreshToken 失效**
    - **Property 7: 登出后 Token 进入黑名单**
    - _验证需求: R1.1-R1.5_

- [x] 7. 检查点 — 认证模块验证
  - 确保 common 模块变更编译通过
  - 确保认证模块核心流程可运行（登录/登出/刷新）
  - 如有问题请询问用户


- [x] 8. 用户管理 UserService（可并行组B-1，依赖任务5、6）
  - [x] 8.1 实现用户管理 DTO
    - UserCreateDTO：username、realName、phone、roleIds（@NotBlank/@NotEmpty 校验）
    - UserUpdateDTO：realName、phone、roleIds、status
    - UserQueryDTO：username、realName、phone、status、roleId、pageNum、pageSize、sortField、sortOrder
    - UserPageDTO：id、username、realName、phone（脱敏）、status、roles、createTime
    - UserDetailDTO：完整用户信息 + 角色列表
    - _需求: R2_

  - [x] 8.2 实现 UserService 接口和 UserServiceImpl
    - `listUsers(UserQueryDTO)` — 分页查询，Service 层手动数据权限过滤（DQ12.1 方案 C）
    - `createUser(UserCreateDTO)` — 唯一性校验（username 不区分大小写、phone_hash）、密码 BCrypt 加密、手机号 AES 加密 + SHA-256 哈希、建立角色关联
    - `updateUser(Long id, UserUpdateDTO)` — 手机号唯一性校验、更新角色关联（先删后增）、清除权限缓存
    - `deleteUser(Long id)` — 逻辑删除、禁止删除 admin/自己、删除角色关联、清除缓存
    - `batchDeleteUsers(List<Long> ids)` — 批量逻辑删除
    - `updateStatus(Long id, Integer status)` — 禁用时加入 Token 黑名单、禁止禁用 admin/自己
    - `batchUpdateStatus(List<Long> ids, Integer status)` — 批量状态切换
    - `resetPassword(Long id)` — 重置为默认密码、清除错误计数和锁定
    - _需求: R2.1-R2.6_

  - [ ]* 8.3 编写 UserService 单元测试
    - **Property 8: 用户名和手机号全局唯一**
    - **Property 9: 手机号加密存储 round-trip**
    - **Property 10: 禁用用户强制下线**
    - **Property 11: 超级管理员保护**
    - **Property 27: pageSize 自动截断**
    - _验证需求: R2.1-R2.6, NFR-2_

- [x] 9. 角色管理 RoleService（可并行组B-2，依赖任务5）
  - [x] 9.1 实现角色管理 DTO
    - RoleCreateDTO：roleName、roleCode、dataPermissionLevel、sort、remark
    - RoleUpdateDTO：roleName、dataPermissionLevel、status、sort、remark
    - RoleQueryDTO：roleName、roleCode、status、pageNum、pageSize、sortField、sortOrder
    - RolePageDTO：id、roleName、roleCode、dataPermissionLevel、status、sort、createTime
    - DataPermissionConfigDTO：dataPermissionLevel、dataIds
    - _需求: R3_

  - [x] 9.2 实现 RoleService 接口和 RoleServiceImpl
    - `listRoles(RoleQueryDTO)` — 分页查询
    - `listAllRoles()` — 全量列表（按数据权限范围过滤下拉）
    - `createRole(RoleCreateDTO)` — 唯一性校验（roleCode 不区分大小写、roleName）
    - `updateRole(Long id, RoleUpdateDTO)` — 禁止修改 roleCode、禁止修改 super_admin 角色
    - `deleteRole(Long id)` — 禁止删除 super_admin、检查关联用户、删除菜单关联和数据权限关联
    - `batchDeleteRoles(List<Long> ids)` — 批量删除
    - `assignMenus(Long roleId, List<Long> menuIds)` — 先删后增、清除角色下用户权限缓存
    - `getRoleMenuIds(Long roleId)` — 查询角色已分配菜单 ID 列表
    - `configDataPermission(Long roleId, DataPermissionConfigDTO)` — 先删后增、清除数据权限缓存
    - _需求: R3.1-R3.6_

  - [ ]* 9.3 编写 RoleService 单元测试
    - **Property 12: 角色删除前置校验**
    - **Property 13: 角色菜单权限先删后增**
    - **Property 14: 权限变更清除缓存**
    - _验证需求: R3.4, R3.5, R3.6_

- [x] 10. 菜单管理 MenuService（可并行组B-3，依赖任务5）
  - [x] 10.1 实现菜单管理 DTO
    - MenuCreateDTO：menuName、parentId、menuType、path、component、permission、icon、sort
    - MenuUpdateDTO：menuName、path、component、permission、icon、sort、status
    - MenuTreeDTO：id、menuName、parentId、menuType、path、component、permission、icon、sort、status、children
    - _需求: R4_

  - [x] 10.2 实现 MenuService 接口和 MenuServiceImpl
    - `getMenuTree(String menuName, Integer status)` — 查询全部菜单，构建树结构，按 sort 升序
    - `getUserMenuTree(Long userId)` — 从缓存获取用户菜单树（仅有权限的菜单）
    - `createMenu(MenuCreateDTO)` — 校验同一父级下名称不重复、按钮权限标识格式校验
    - `updateMenu(Long id, MenuUpdateDTO)` — 禁止修改 menuType、防止循环引用
    - `deleteMenu(Long id)` — 检查子菜单、检查角色关联
    - 树构建工具方法：递归构建、parentId=0 为根节点
    - _需求: R4.1-R4.4_

  - [ ]* 10.3 编写 MenuService 单元测试
    - **Property 15: 菜单树结构正确性**
    - **Property 16: 菜单删除前置校验**
    - _验证需求: R4.1, R4.4_


- [x] 11. 数据权限服务 DataPermissionService（依赖任务9）
  - [x] 11.1 实现 DataPermissionService 接口和实现类
    - `getOptions()` — 返回数据权限可选范围（本期硬编码：3 片区、5 小区、8 楼栋）
    - `loadUserDataPermission(Long userId)` — 加载用户数据权限上下文（多角色并集合并算法）
    - `configRoleDataPermission(Long roleId, DataPermissionConfigDTO)` — 配置角色数据权限（先删后增）
    - 实现多角色合并算法：ALL 优先 → scopeMap 合并 → SELF 兜底
    - _需求: R5.1-R5.4, R3.6_

  - [x] 11.2 实现 DataPermissionOptionVO 值对象
    - 创建 `com.spmp.user.domain.vo.DataPermissionOptionVO`
    - 包含 areas、communities、buildings 三个列表（id + name）
    - _需求: R5.4_

  - [ ]* 11.3 编写 DataPermissionService 单元测试
    - **Property 17: 数据权限多角色并集合并**
    - _验证需求: R5.2_

- [x] 12. 个人中心 ProfileService（依赖任务5、6）
  - [x] 12.1 实现个人中心 DTO
    - ProfileDTO：userId、username、realName、phone（脱敏）、avatar、roles、permissions、dataPermissionLevel
    - ProfileUpdateDTO：realName、phone
    - PasswordUpdateDTO：oldPassword、newPassword（@NotBlank 校验）
    - _需求: R6_

  - [x] 12.2 实现 ProfileService 接口和 ProfileServiceImpl
    - `getProfile()` — 获取当前用户信息 + 角色列表 + 权限标识列表 + 数据权限级别
    - `updateProfile(ProfileUpdateDTO)` — 修改姓名/手机号，手机号唯一性校验 + AES 加密
    - `updatePassword(PasswordUpdateDTO)` — 旧密码校验、新密码强度校验（正则）、新旧不同校验、BCrypt 加密存储、Token 加入黑名单强制重新登录
    - _需求: R6.1-R6.3_

  - [ ]* 12.3 编写 ProfileService 单元测试
    - **Property 20: 修改密码后强制重新登录**
    - **Property 21: 密码强度校验**
    - _验证需求: R6.2_

- [x] 13. 对外 API: UserApi + PermissionApi（可并行组C-1，依赖任务8、11）
  - [x] 13.1 实现 UserApi 接口和 DTO
    - 创建 `com.spmp.user.api.UserApi` 接口
    - 创建 `com.spmp.user.api.dto.UserBriefDTO`：id、username、realName、phone（脱敏）
    - 方法：getUserById、getUsersByRoleCode、getUsersByIds
    - _需求: R7.1_

  - [x] 13.2 实现 PermissionApi 接口和 DTO
    - 创建 `com.spmp.user.api.PermissionApi` 接口
    - 创建 `com.spmp.user.api.dto.DataPermissionDTO`：level、scopeMap
    - 方法：getDataPermission、checkPermission、getUserRoles
    - _需求: R7.2_

  - [x] 13.3 在 UserServiceImpl 中实现 UserApi 接口
    - UserServiceImpl 同时 implements UserApi
    - _需求: R7.3_

  - [x] 13.4 在 DataPermissionServiceImpl 中实现 PermissionApi 接口
    - DataPermissionServiceImpl 同时 implements PermissionApi
    - _需求: R7.3_

  - [ ]* 13.5 编写对外 API 单元测试
    - **Property 28: UserApi round-trip**
    - **Property 29: PermissionApi.checkPermission 一致性**
    - _验证需求: R7.1, R7.2_


- [x] 14. 操作日志 AOP（可并行组C-2，依赖任务3）
  - [x] 14.1 实现 @OperationLog 自定义注解
    - 创建 `com.spmp.user.annotation.OperationLog`
    - 参数：module（模块名称）、type（操作类型）、description（操作描述）
    - _需求: R9.1_

  - [x] 14.2 实现 AsyncConfig 异步线程池配置
    - 创建 `com.spmp.user.config.AsyncConfig`，@EnableAsync
    - 注册 `operationLogExecutor` Bean：corePoolSize=2、maxPoolSize=5、queueCapacity=200、CallerRunsPolicy
    - _需求: R9.1_

  - [x] 14.3 实现 OperationLogAspect 切面
    - 创建 `com.spmp.user.aspect.OperationLogAspect`
    - @Around 拦截 @OperationLog 注解方法
    - 记录：操作人信息、请求方法/URL/参数、响应结果、耗时
    - 请求参数 JSON 序列化，password 字段脱敏为 "******"
    - 响应结果超 2000 字符截断
    - 异常时记录异常类名 + 异常消息
    - @Async("operationLogExecutor") 异步写入数据库
    - _需求: R9.1_

  - [x] 14.4 实现 OperationLogService 接口和实现类
    - `saveLog(OperationLogDO)` — 异步保存操作日志
    - `listLogs(OperationLogQueryDTO)` — 分页查询操作日志（按操作时间倒序）
    - _需求: R9.1, R9.2_

  - [ ]* 14.5 编写操作日志 AOP 单元测试
    - **Property 22: 操作日志异步记录完整性**
    - **Property 23: 操作日志密码脱敏**
    - _验证需求: R9.1_

- [x] 15. 限流 AOP（可并行组C-3，依赖任务4）
  - [x] 15.1 实现 @RateLimit 自定义注解
    - 创建 `com.spmp.user.annotation.RateLimit`
    - 参数：key（限流 key 前缀）、window（时间窗口秒）、maxCount（最大请求次数）、dimension（IP/USER/PHONE/CUSTOM）、message（提示消息）
    - _需求: NFR-1_

  - [x] 15.2 实现 RateLimitAspect 切面
    - 创建 `com.spmp.user.aspect.RateLimitAspect`
    - @Before 拦截 @RateLimit 注解方法
    - 根据 dimension 构建 Redis key：`rate-limit:{key}:{dimension_value}`
    - Redis INCR + EXPIRE 实现固定窗口计数器
    - 超过 maxCount 抛出 BusinessException(RATE_LIMIT_EXCEEDED)
    - Redis 不可用时降级放行 + WARN 日志
    - _需求: NFR-1_

  - [ ]* 15.3 编写限流 AOP 单元测试
    - **Property 24: 限流计数器正确性**
    - _验证需求: NFR-1_

- [x] 16. 登录日志模块（依赖任务6、14）
  - [x] 16.1 实现 LoginLogService 接口和实现类
    - `saveLoginLog(LoginLogDO)` — 异步保存登录日志
    - `listLoginLogs(LoginLogQueryDTO)` — 分页查询登录日志（按登录时间倒序）
    - _需求: R8.1, R8.2_

  - [x] 16.2 实现 LoginLogQueryDTO
    - 字段：username、loginIp、loginResult、startTime、endTime、pageNum、pageSize
    - _需求: R8.2_

  - [ ]* 16.3 编写登录日志单元测试
    - **Property 30: 登录日志完整记录**
    - _验证需求: R8.1_


- [x] 17. Controller 层 + 安全配置（依赖任务6-16）
  - [x] 17.1 实现 AuthController
    - `GET /api/v1/user/auth/captcha` — 获取图形验证码
    - `POST /api/v1/user/auth/login` — 用户名密码登录（@RateLimit IP 60s/10次）
    - `POST /api/v1/user/auth/login/sms` — 手机号验证码登录
    - `POST /api/v1/user/auth/sms-code` — 发送短信验证码（@RateLimit PHONE 60s/1次 + IP 60s/10次）
    - `POST /api/v1/user/auth/refresh` — 刷新 Token（@RateLimit IP 60s/20次）
    - `POST /api/v1/user/auth/logout` — 登出
    - _需求: R1_

  - [x] 17.2 实现 UserController
    - `GET /users` — 用户分页查询（@PreAuthorize user:user:list）
    - `POST /users` — 新增用户（@PreAuthorize user:user:create + @OperationLog）
    - `PUT /users/{id}` — 编辑用户（@PreAuthorize user:user:edit + @OperationLog）
    - `DELETE /users/{id}` — 删除用户（@PreAuthorize user:user:delete + @OperationLog）
    - `DELETE /users/batch` — 批量删除（@PreAuthorize user:user:delete + @OperationLog）
    - `PUT /users/{id}/status` — 状态切换（@PreAuthorize user:user:edit + @OperationLog）
    - `PUT /users/batch-status` — 批量状态切换（@PreAuthorize user:user:edit + @OperationLog）
    - `PUT /users/{id}/reset-password` — 重置密码（@PreAuthorize user:user:reset-pwd + @OperationLog）
    - _需求: R2_

  - [x] 17.3 实现 RoleController
    - `GET /roles` — 角色分页查询（@PreAuthorize user:role:list）
    - `GET /roles/list` — 角色全量列表（@PreAuthorize user:role:list）
    - `POST /roles` — 新增角色（@PreAuthorize user:role:create + @OperationLog）
    - `PUT /roles/{id}` — 编辑角色（@PreAuthorize user:role:edit + @OperationLog）
    - `DELETE /roles/{id}` — 删除角色（@PreAuthorize user:role:delete + @OperationLog）
    - `DELETE /roles/batch` — 批量删除（@PreAuthorize user:role:delete + @OperationLog）
    - `GET /roles/{id}/menus` — 查询角色菜单（@PreAuthorize user:role:list）
    - `PUT /roles/{id}/menus` — 分配菜单权限（@PreAuthorize user:role:assign + @OperationLog）
    - `PUT /roles/{id}/data-permission` — 配置数据权限（@PreAuthorize user:role:assign + @OperationLog）
    - _需求: R3_

  - [x] 17.4 实现 MenuController
    - `GET /menus/tree` — 菜单树查询（@PreAuthorize user:menu:list）
    - `GET /menus/user-tree` — 当前用户菜单树（仅需认证）
    - `POST /menus` — 新增菜单（@PreAuthorize user:menu:create + @OperationLog）
    - `PUT /menus/{id}` — 编辑菜单（@PreAuthorize user:menu:edit + @OperationLog）
    - `DELETE /menus/{id}` — 删除菜单（@PreAuthorize user:menu:delete + @OperationLog）
    - _需求: R4_

  - [x] 17.5 实现 ProfileController
    - `GET /profile` — 获取个人信息（仅需认证）
    - `PUT /profile` — 修改个人信息（@OperationLog）
    - `PUT /profile/password` — 修改密码（@OperationLog）
    - _需求: R6_

  - [x] 17.6 实现 LoginLogController + OperationLogController
    - `GET /login-logs` — 登录日志查询（@PreAuthorize user:log:list）
    - `GET /operation-logs` — 操作日志查询（@PreAuthorize user:log:list）
    - _需求: R8.2, R9.2_

  - [x] 17.7 实现数据权限接口
    - `GET /data-permission/options` — 数据权限可选范围（@PreAuthorize user:role:assign）
    - _需求: R5.4_

  - [x] 17.8 实现 UserSecurityConfig 模块安全扩展配置
    - 启用 @EnableGlobalMethodSecurity(prePostEnabled = true)
    - 注册 PermissionService Bean
    - _需求: NFR-2_

- [x] 18. 检查点 — 业务模块验证
  - 确保所有 Controller 编译通过
  - 确保 @PreAuthorize 注解正确配置
  - 确保 @OperationLog 注解正确标注
  - 确保 @RateLimit 注解正确配置
  - 如有问题请询问用户


- [x] 19. 配置文件汇总（依赖任务1-18）
  - [x] 19.1 更新 application.yml 配置
    - user 模块配置：default-password、login.max-fail-count、login.lock-minutes
    - 安全白名单扩展：追加 user 模块认证接口路径
    - 异步线程池配置（如需外部化）
    - _需求: 全局配置_

  - [x] 19.2 创建 Mapper XML 目录和 MyBatis 配置
    - 确保 `resources/mapper/` 目录下所有 XML 文件正确映射
    - mybatis-plus 配置中增加 mapper-locations 扫描路径
    - _需求: 全局配置_

- [x] 20. 最终检查点 — 全模块验证
  - 确保所有组件编译通过，所有测试通过
  - 验证 user 模块包结构完整性（controller、service、repository、domain、api、config、constant、annotation、aspect、security）
  - 确认所有 9 个需求模块（R1-R9）均已覆盖
  - 确认所有 30 个正确性属性均已分配到对应的测试子任务中
  - 确认 common 模块变更向后兼容（除 DataPermissionContext 外）
  - 如有问题请询问用户

## 并行执行说明

| 并行组 | 包含任务 | 前置依赖 | 说明 |
|--------|---------|---------|------|
| 组A（可并行） | 任务3、任务4 | 任务2 完成后 | DO 实体/Mapper 与 错误码/常量 无依赖关系 |
| 组B（可并行） | 任务8、任务9、任务10 | 任务5、6 完成后 | 用户/角色/菜单管理互不依赖 |
| 组C（可并行） | 任务13、任务14、任务15 | 任务3、4 完成后 | 对外 API、操作日志、限流互不依赖 |

## 关键路径

```
任务1 → 任务2 → 任务3/4(并行) → 任务5 → 任务6(认证) → 任务7(检查点)
                                                          ↓
                                              任务8/9/10(并行) → 任务11/12(并行)
                                                          ↓
                                              任务13/14/15(并行) → 任务16 → 任务17(Controller)
                                                                              ↓
                                                                    任务18(检查点) → 任务19 → 任务20
```

**预估关键路径工期：** 任务1 → 2 → 3 → 5 → 6 → 8 → 11 → 13 → 17 → 19 → 20（约 11 个串行节点）

## 备注

- 标记 `*` 的子任务为可选测试任务，可跳过以加速 MVP 交付
- 每个任务引用了具体的需求编号（R1-R9）和正确性属性（Property 1-30），确保可追溯性
- 单元测试使用 JUnit 5 + Mockito 框架
- 检查点任务确保增量验证，及早发现问题
- 认证模块（任务6）是前后端联调的关键路径，优先完成
- common 模块变更（任务1）是所有后续任务的前置条件，必须最先完成
- 数据库脚本（任务2）必须在 DDL 执行后才能进行 DML 初始化
- Controller 层（任务17）统一放在最后实现，确保所有 Service 层逻辑已就绪


---

## 前端任务（Task 21-27）

> 以下任务为 user 模块前端管理页面开发，对接已完成的后端接口。

### 并行分组

| 分组 | 任务 | 说明 |
|---|---|---|
| D（前端基础） | Task 21 | API 封装和 Store，其他前端任务依赖 |
| E（前端页面，可并行） | Task 22, 23, 24, 25, 26 | 各管理页面互不依赖 |
| F（路由集成） | Task 27 | 所有页面完成后集成路由和菜单 |

- [x] 21. 前端 API 封装和 Store
  - [x] 21.1 创建 `src/api/user.ts` — 用户管理 API（listUsers、createUser、updateUser、deleteUser、batchDelete、updateStatus、resetPassword）
  - [x] 21.2 创建 `src/api/role.ts` — 角色管理 API（listRoles、listAllRoles、createRole、updateRole、deleteRole、getRoleMenuIds、assignMenus、configDataPermission）
  - [x] 21.3 创建 `src/api/menu.ts` — 菜单管理 API（getMenuTree、getUserMenuTree、createMenu、updateMenu、deleteMenu）
  - [x] 21.4 创建 `src/api/auth.ts` — 认证 API（login、logout、refreshToken、getCaptcha、sendSmsCode）
  - [x] 21.5 创建 `src/api/profile.ts` — 个人中心 API（getProfile、updateProfile、updatePassword）
  - [x] 21.6 创建 `src/api/log.ts` — 日志 API（listLoginLogs、listOperationLogs）
  - [x] 21.7 创建 `src/api/data-permission.ts` — 数据权限 API（getOptions）
  - [x] 21.8 创建 `src/store/modules/user.ts` — 用户状态 Store（token、userInfo、permissions、menus）

- [x] 22. 用户管理页面
  - [x] 22.1 创建 `src/views/system/user/UserList.vue` — 用户列表页（表格 + 搜索 + 分页 + 状态切换 + 重置密码）
  - [x] 22.2 创建 `src/views/system/user/UserForm.vue` — 用户新增/编辑表单（对话框，含角色选择）

- [x] 23. 角色管理页面
  - [x] 23.1 创建 `src/views/system/role/RoleList.vue` — 角色列表页（表格 + 搜索 + 分页）
  - [x] 23.2 创建 `src/views/system/role/RoleForm.vue` — 角色新增/编辑表单
  - [x] 23.3 创建 `src/views/system/role/MenuAssign.vue` — 菜单权限分配（树形勾选对话框）
  - [x] 23.4 创建 `src/views/system/role/DataPermissionConfig.vue` — 数据权限配置对话框

- [x] 24. 菜单管理页面
  - [x] 24.1 创建 `src/views/system/menu/MenuList.vue` — 菜单树形列表页（树形表格 + 新增/编辑/删除）
  - [x] 24.2 创建 `src/views/system/menu/MenuForm.vue` — 菜单新增/编辑表单（对话框，含图标选择）

- [x] 25. 个人中心页面
  - [x] 25.1 创建 `src/views/profile/ProfilePage.vue` — 个人信息展示 + 修改 + 密码修改

- [x] 26. 日志管理页面
  - [x] 26.1 创建 `src/views/system/log/LoginLogList.vue` — 登录日志列表（表格 + 搜索 + 分页）
  - [x] 26.2 创建 `src/views/system/log/OperationLogList.vue` — 操作日志列表（表格 + 搜索 + 分页）

- [x] 27. 路由和菜单集成
  - [x] 27.1 更新 `src/router/static-routes.ts` — 添加系统管理子路由（用户、角色、菜单、日志）和个人中心路由
  - [x] 27.2 更新 `src/router/guard.ts` — 完善登录状态检查和动态菜单加载
  - [x] 27.3 更新 `src/layout/AppLayout.vue` — 侧边栏菜单渲染（基于用户菜单树）