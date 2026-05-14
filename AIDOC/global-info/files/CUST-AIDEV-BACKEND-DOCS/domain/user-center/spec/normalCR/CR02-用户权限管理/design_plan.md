# CR02 - 用户权限管理 设计计划

| 项目 | 内容 |
|------|------|
| CR 编号 | CR02 |
| 需求名称 | 用户权限管理 |
| 所属模块 | user（用户中心） |
| 版本 | 1.0 |
| 创建日期 | 2026-04-18 |
| 状态 | 已完成 |
| 需求文档 | requirements.md |
| 依赖 | common 模块（CR01 已完成） |

---

## 一、设计范围概述

基于已确认的 requirements.md（9 个需求模块 R1-R9），本设计计划覆盖以下技术设计领域：

| 设计领域 | 涉及需求 | 说明 |
|---------|---------|------|
| 数据库详细设计 | R1-R9 | 8 张表的 DDL、索引、约束 |
| API 接口详细设计 | R1-R9 | 请求/响应 DTO、校验规则 |
| DDD 分层架构设计 | R1-R9 | entity/dto/vo/service/repository 分包 |
| 安全设计 | R1, NFR-2 | CSRF、@PreAuthorize、密码策略、Token 黑名单 |
| 缓存设计 | R1, R5, NFR-3 | Redis key 规范、缓存策略、降级方案 |
| 短信验证码集成 | R1.2 | 短信服务对接或模拟方案 |
| 图形验证码设计 | R1.3 | 验证码生成库选型 |
| 操作日志 AOP 设计 | R9 | @OperationLog 注解 + 切面 |
| 数据权限扩展设计 | R5 | SELF 级别 + 多角色并集 + DataPermissionInterceptor 扩展 |
| 限流设计 | NFR-1 | 登录接口 + 短信接口限流方案 |
| common 模块变更设计 | R5, NFR-2 | DataPermissionInterceptor 扩展、CSRF 策略、ErrorCode 扩展 |
| 对外 API 设计 | R7 | UserApi / PermissionApi 接口定义 |

---

## 二、设计计划步骤

- [x] 步骤 1：等待用户确认所有澄清问题（第三节）
- [x] 步骤 2：数据库详细设计（DDL、索引、初始化数据）
- [x] 步骤 3：DDD 分层架构与包结构设计
- [x] 步骤 4：安全设计（CSRF 策略变更、@PreAuthorize 实现、Token 黑名单）
- [x] 步骤 5：缓存设计（Redis key 规范、缓存策略、降级方案）
- [x] 步骤 6：API 接口详细设计（DTO/VO 定义、校验规则、接口契约）
- [x] 步骤 7：登录认证流程设计（用户名密码登录、手机号验证码登录、图形验证码、Token 刷新/登出）
- [x] 步骤 8：数据权限扩展设计（SELF 级别、多角色并集、DataPermissionInterceptor 改造）
- [x] 步骤 9：操作日志 AOP 设计（@OperationLog 注解 + 异步切面）
- [x] 步骤 10：限流设计（登录接口 + 短信接口限流）
- [x] 步骤 11：对外 API 设计（UserApi / PermissionApi）
- [x] 步骤 12：common 模块变更设计汇总
- [x] 步骤 13：正确性属性（Correctness Properties）定义
- [x] 步骤 14：整合输出 design.md

---

## 三、澄清问题

### DQ1：数据库详细设计

#### DQ1.1 sys_user 表字段设计

[Question] requirements.md 中 sys_user 表包含 `phone`（加密存储）字段。为支持手机号登录和唯一性校验，需要额外存储手机号的哈希值或密文索引。请确认以下方案：
- A：增加 `phone_hash` 字段（SHA-256 哈希），用于唯一性校验和登录查询，`phone` 字段存储 AES 加密的完整手机号用于展示
- B：直接对 `phone` 的 AES 密文建立唯一索引（同一明文加密后密文固定，因为 EncryptUtils 使用固定 IV）
- C：其他方案

建议采用方案 B，因为 EncryptUtils 使用固定 IV（密钥前 16 字节），同一明文加密结果一致，可直接对密文字段建唯一索引。

[Answer] A方案

#### DQ1.2 sys_user 表 — 头像字段

[Question] requirements.md 中 sys_user 包含 `avatar` 字段。请确认：
1. 头像存储方式：存储 URL 路径（如 `/uploads/avatar/xxx.jpg`）还是 Base64？
2. 本期是否实现头像上传功能，还是仅预留字段？

[Answer] 1

#### DQ1.3 sys_role_data_permission 表设计

[Question] requirements.md 中 sys_role_data_permission 表记录角色关联的具体片区/小区/楼栋 ID。当数据权限级别为 AREA 时，一个角色可能关联多个片区。请确认表结构方案：
- A：每行一条关联记录（role_id + data_type + data_id），一个角色多个片区对应多行
- B：使用 JSON 字段存储 data_id 列表（如 `[1, 2, 3]`）

建议采用方案 A（每行一条记录），便于 SQL 查询和索引优化。

[Answer] A

#### DQ1.4 索引策略

[Question] 请确认以下索引设计是否合理：
1. sys_user：`username`（唯一索引）、`phone`（唯一索引，基于密文）、`status + del_flag`（组合索引）
2. sys_role：`role_code`（唯一索引）、`role_name`（唯一索引）
3. sys_menu：`parent_id`（普通索引）、`permission`（普通索引）
4. sys_user_role：`user_id`（普通索引）、`role_id`（普通索引）、`user_id + role_id`（唯一索引）
5. sys_role_menu：`role_id`（普通索引）、`role_id + menu_id`（唯一索引）
6. sys_role_data_permission：`role_id`（普通索引）、`role_id + data_type + data_id`（唯一索引）
7. sys_login_log：`username`（普通索引）、`login_time`（普通索引）
8. sys_operation_log：`operator_id`（普通索引）、`operation_time`（普通索引）、`module`（普通索引）

是否需要调整或补充？

[Answer] 确认

#### DQ1.5 初始化数据

[Question] 请确认以下初始化数据需求：
1. 超级管理员账号：用户名 `admin`，默认密码 `Spmp@2026`，关联 `super_admin` 角色
2. 预置角色：super_admin、property_admin、area_manager、building_steward、repairman、owner
3. 预置菜单：是否需要在 DML 中初始化完整的菜单树（用户管理、角色管理、菜单管理、日志管理等），还是由管理员手动创建？
4. 测试用片区/小区/楼栋数据：需要初始化几组测试数据？

[Answer] 1 确认，2，确认，3， 需要初始化完整菜单树， 4，需要能展示通用化效果，需要多片区，多小区，多楼栋数据

---

### DQ2：API 接口详细设计

#### DQ2.1 分页查询参数规范

[Question] requirements.md 中分页查询默认 pageNum=1、pageSize=10。请确认：
1. pageSize 最大值限制：是否限制为 100？超出时自动截断还是返回错误？
2. 排序参数：是否需要支持动态排序（如 `sortField=createTime&sortOrder=desc`），还是各接口固定排序规则？

[Answer] 1 限制100 2，动态排序

#### DQ2.2 用户列表查询 — 数据权限下的角色筛选

[Question] R2.1 用户列表支持按角色筛选。当管理员只能看到自己数据权限范围内的用户时，角色筛选的下拉列表是否也需要过滤（只显示当前管理员可管理的角色），还是显示所有角色？

[Answer] 按数据权限范围过滤下拉列表

#### DQ2.3 批量操作

[Question] 当前 requirements.md 中用户管理和角色管理均为单条操作。是否需要支持批量操作？
1. 批量删除用户
2. 批量启用/禁用用户
3. 批量删除角色

如不需要，本期仅实现单条操作。

[Answer] 需要支持

#### DQ2.4 登录响应 DTO 内容

[Question] 登录成功后返回的响应 DTO 除了 access_token 和 refresh_token 外，是否还需要包含以下信息？
1. 用户基本信息（userId、username、realName、avatar）
2. 角色列表
3. 权限标识列表（用于前端 v-permission 指令）
4. 菜单树（用于前端动态路由）

还是前端登录后再单独调用 `GET /api/v1/user/profile` 获取这些信息？

建议：登录响应仅返回 token 信息 + 用户基本信息，权限和菜单由前端单独请求，减少登录接口的复杂度。

[Answer] 同意，分开接口

---

### DQ3：DDD 分层设计

#### DQ3.1 user 模块是否使用 DDD 聚合根模式

[Question] common 模块提供了 AggregateRoot、Repository、AbstractRepository 等 DDD 持久化组件。user 模块的实体关系相对简单（用户-角色多对多、角色-菜单多对多），请确认：
- A：使用完整的 DDD 聚合根模式（UserAggregate 包含角色列表，通过 UserRepository 统一管理）
- B：使用简化的 Service + Mapper 模式（不使用聚合根，Service 直接操作多个 Mapper）

建议采用方案 B（简化模式），原因：
1. user 模块的实体关系以多对多关联为主，不适合聚合根模式
2. 角色和菜单是独立管理的实体，不属于用户聚合的一部分
3. 简化模式更直观，降低开发复杂度

[Answer] 用户管理操作业务比较简单，可以使用B方案

#### DQ3.2 DTO/VO 命名规范

[Question] 请确认 DTO/VO 的命名规范：
- 请求 DTO：`XxxCreateDTO`、`XxxUpdateDTO`、`XxxQueryDTO`（放在 `domain/dto/` 包下）
- 响应 VO：`XxxVO`、`XxxDetailVO`（放在 `domain/vo/` 包下）
- 跨模块 DTO：`XxxBriefDTO`（放在 `api/dto/` 包下）

是否同意以上命名规范？

[Answer] 数据传输入参出参使用DTO结尾命名，比如BillDTO， 领域实体对象则无后缀如Bill, 值对象使用VO结尾命名，表对象则使用DO结尾命名，此规范更新到领域设计实现的技术规范中；

---

### DQ4：安全设计

#### DQ4.1 CSRF 策略变更

[Question] requirements.md NFR-2.4 要求对非 GET 请求启用 CSRF Token 校验。当前 common 模块 SecurityConfig 中 CSRF 已完全禁用（`csrf().disable()`）。启用 CSRF 需要：
1. 前端在每次非 GET 请求中携带 CSRF Token（通过 Cookie 或 Header）
2. 后端配置 CsrfTokenRepository（CookieCsrfTokenRepository 或 HttpSessionCsrfTokenRepository）

但 SPMP 使用 JWT 无状态认证，CSRF 防护的必要性存在争议：
- JWT 存储在 localStorage 中时，天然免疫 CSRF 攻击（因为浏览器不会自动携带 localStorage 中的 Token）
- JWT 存储在 HttpOnly Cookie 中时，才需要 CSRF 防护

请确认：
- A：保持 CSRF 禁用（JWT 存储在 localStorage，前端通过 Authorization Header 携带 Token，天然免疫 CSRF）
- B：启用 CSRF 防护（使用 CookieCsrfTokenRepository，前端需要额外处理 CSRF Token）

建议采用方案 A，因为 SPMP 使用 JWT + Authorization Header 方式，不存在 CSRF 风险。

[Answer] A

#### DQ4.2 @PreAuthorize 实现方式

[Question] requirements.md 要求使用 `@PreAuthorize` 注解进行后端权限标识校验。请确认实现方式：
- A：使用 Spring Security 内置的 `@PreAuthorize("hasAuthority('user:user:list')")`，需要在 JWT 认证时将用户的权限标识列表加载到 SecurityContext 的 GrantedAuthority 中
- B：自定义权限校验方法 `@PreAuthorize("@perm.check('user:user:list')")`，通过自定义 Bean 从 Redis 缓存中查询用户权限

建议采用方案 B，原因：
1. 权限标识列表可能较长，放入 JWT Claims 会增大 Token 体积
2. 从 Redis 缓存查询更灵活，权限变更后无需等待 Token 过期即可生效

[Answer] B

#### DQ4.3 Token 黑名单存储策略

[Question] 登出、禁用用户、修改密码等场景需要将 Token 加入黑名单。请确认黑名单的 Redis key 设计：
- A：`token:blacklist:{tokenHash}`（对 Token 做 SHA-256 哈希后作为 key，避免 key 过长）
- B：`token:blacklist:{jti}`（在 JWT Claims 中增加 jti 唯一标识，用 jti 作为 key）

建议采用方案 B（使用 jti），原因：
1. jti 是 JWT 标准 Claim，语义清晰
2. key 更短，Redis 内存占用更小
3. 需要在 JwtTokenProvider.generateToken 中增加 jti（UUID）

[Answer] B

---

### DQ5：缓存设计

#### DQ5.1 Redis key 命名规范

[Question] 请确认以下 Redis key 命名规范是否合理：

| 用途 | Key 格式 | TTL | 说明 |
|------|---------|-----|------|
| 登录失败计数 | `login:fail:{username}` | 30 分钟 | 密码错误计数 |
| 账号锁定 | `login:lock:{username}` | 30 分钟 | 账号锁定标记 |
| 图形验证码 | `captcha:{uuid}` | 2 分钟 | 验证码文本 |
| 短信验证码 | `sms:code:{phone}` | 5 分钟 | 6 位数字验证码 |
| 短信发送间隔 | `sms:interval:{phone}` | 60 秒 | 防止频繁发送 |
| 短信日发送次数 | `sms:daily:{phone}` | 当日剩余秒数 | 每日限制 10 次 |
| Token 黑名单 | `token:blacklist:{jti}` | Token 剩余有效期 | 已失效的 Token |
| 用户权限缓存 | `user:permissions:{userId}` | 8 小时 | 权限标识列表 |
| 用户数据权限缓存 | `user:data-permission:{userId}` | 8 小时 | 数据权限信息 |
| 用户菜单缓存 | `user:menus:{userId}` | 8 小时 | 用户菜单树 |

是否需要调整 key 格式或 TTL？

[Answer] 确认

#### DQ5.2 缓存降级策略

[Question] NFR-3.3 要求 Redis 不可用时降级为直接查询数据库。请确认降级范围：
1. 权限缓存（user:permissions）：降级为查数据库 ✅
2. 数据权限缓存（user:data-permission）：降级为查数据库 ✅
3. Token 黑名单（token:blacklist）：Redis 不可用时如何处理？
   - A：放行所有 Token（安全风险：已登出的 Token 仍可使用）
   - B：拒绝所有请求（可用性风险：Redis 故障导致全站不可用）
   - C：记录日志告警，放行 Token 但标记为降级状态

建议采用方案 C，在 Redis 恢复后自动恢复黑名单校验。

[Answer] C

---

### DQ6：短信验证码集成方案

#### DQ6.1 短信服务实现方式

[Question] R1.2 手机号验证码登录需要发送短信验证码。请确认本期的实现方式：
- A：集成真实短信服务（如阿里云短信、腾讯云短信），需要提供 AccessKey 和短信模板 ID
- B：使用日志模拟方式（验证码生成后仅打印到日志，不实际发送短信），预留短信服务接口，后续对接真实服务
- C：使用本地模拟 + 固定验证码（如固定返回 `123456`，仅用于开发测试）

建议采用方案 B（日志模拟 + 接口预留），原因：
1. 本期重点是登录认证流程的完整性，短信发送是外部依赖
2. 通过 SmsService 接口 + SmsServiceMockImpl 实现，后续只需替换实现类即可对接真实服务
3. 验证码仍然存入 Redis 并校验，保证流程完整性

[Answer] B

---

### DQ7：图形验证码实现方案

#### DQ7.1 图形验证码库选型

[Question] R1.3 图形验证码需要生成 Base64 编码的验证码图片。请确认使用的库：
- A：Hutool 的 CaptchaUtil（项目已引入 Hutool 依赖，零额外依赖）
- B：Kaptcha（Google 开源验证码库，功能丰富但需额外引入依赖）
- C：EasyCaptcha（轻量级验证码库，支持算术验证码、GIF 验证码等）
- D：自行实现（使用 Java AWT/Graphics2D 绘制）

建议采用方案 A（Hutool CaptchaUtil），原因：
1. 项目已引入 Hutool 依赖，无需额外引入
2. 支持线段干扰、圆圈干扰、扭曲等多种样式
3. 直接输出 Base64 编码，满足需求

[Answer] A

---

### DQ8：操作日志 AOP 设计

#### DQ8.1 操作日志记录粒度

[Question] R9.1 要求通过 AOP 切面记录操作日志。请确认以下设计细节：
1. 请求参数记录：是否记录完整的请求参数 JSON？对于包含密码的请求（如登录、修改密码），是否需要脱敏处理？
2. 响应结果记录：是否记录完整的响应结果 JSON？对于大数据量的分页查询响应，是否需要截断？
3. 异常记录：方法执行异常时，是否记录异常信息到操作日志的响应结果字段？

建议：
- 请求参数：记录 JSON，对 password 字段自动脱敏为 `******`
- 响应结果：记录 JSON，超过 2000 字符时截断
- 异常：记录异常类名 + 异常消息，不记录完整堆栈（堆栈在应用日志中已有）

[Answer] 确认

#### DQ8.2 异步记录方式

[Question] R9.1 要求异步记录操作日志。请确认异步实现方式：
- A：使用 Spring `@Async` 注解 + 自定义线程池
- B：使用 Spring ApplicationEvent（发布事件，异步监听器处理）
- C：使用 BlockingQueue + 单独消费线程

建议采用方案 A（@Async + 线程池），原因：
1. 实现简单，Spring 原生支持
2. 线程池可配置核心线程数、队列大小，便于控制资源
3. 登录日志也可复用同一异步机制

[Answer] A

---

### DQ9：数据权限扩展设计

#### DQ9.1 SELF 级别与 DataPermissionInterceptor 集成

[Question] 需要扩展 common 模块的 DataPermissionInterceptor 增加 SELF（仅本人）级别。请确认实现方案：
1. @DataPermission 注解增加 `selfField` 参数（默认 `"create_by"`），用于指定"仅本人"级别的过滤字段
2. SELF 级别的 SQL 过滤条件为 `WHERE {selfField} = {currentUserId}`
3. 当 selfField 为 `"create_by"` 时，过滤条件为 `WHERE create_by = {currentUserId}`
4. 某些场景可能需要自定义 selfField（如工单表的 `assigned_to` 字段），通过注解参数指定

请确认：
- selfField 的默认值是 `"create_by"` 还是其他字段？
- selfField 存储的是用户 ID（Long）还是用户名（String）？BaseEntity 的 create_by 存储的是用户名，但按用户名过滤可能不够精确。

建议：selfField 默认值为 `"create_by"`，存储用户名（与 BaseEntity.createBy 一致）。如果需要按用户 ID 过滤，可通过 `@DataPermission(selfField = "user_id")` 指定。

[Answer] 同意建议

#### DQ9.2 多角色数据权限合并算法

[Question] R5.2 要求多角色数据权限取并集。请确认以下合并规则的优先级：
1. 任一角色为 ALL → 不追加过滤条件（最高优先级）
2. 混合级别（如一个角色 AREA + 另一个角色 COMMUNITY）→ 如何合并？
   - A：取最高级别（AREA > COMMUNITY > BUILDING > SELF），只使用最高级别的数据范围
   - B：合并所有级别的数据 ID（AREA 的片区 ID + COMMUNITY 的小区 ID 都作为过滤条件）
   - C：将低级别转换为高级别后合并（如 COMMUNITY 的小区 ID 转换为所属片区 ID，与 AREA 合并）

建议采用方案 B（合并所有数据 ID），原因：
1. 语义最清晰：用户能看到角色 A 关联的片区数据 + 角色 B 关联的小区数据
2. 实现相对简单：将所有角色的 area_id、community_id、building_id 分别合并为集合
3. SQL 条件：`WHERE area_id IN (...) OR community_id IN (...) OR building_id IN (...)`

[Answer] B

#### DQ9.3 DataPermissionContext 扩展

[Question] 当前 common 模块的 DataPermissionContext 存储单个 ID（areaId、communityId、buildingId）。多角色并集需要存储 ID 列表。请确认是否将 DataPermissionContext 改为：

```java
public class DataPermissionContext {
    private DataPermissionLevel level;       // 最终合并后的级别
    private Set<Long> areaIds;              // 片区 ID 集合
    private Set<Long> communityIds;         // 小区 ID 集合
    private Set<Long> buildingIds;          // 楼栋 ID 集合
    private Long userId;                    // 当前用户 ID（SELF 级别使用）
    private String username;                // 当前用户名（SELF 级别 create_by 过滤使用）
}
```

这是对 common 模块的 breaking change，需要同步修改 DataPermissionInterceptor 的 SQL 拼接逻辑（从 `= #{id}` 改为 `IN (...)`）。是否确认？

[Answer] 关于数据权限这块是否可以抽线抽像一下，让这个模块更通用一点，不要直接写死areaIds这种属性，通过外部传入，对应的key和值的方式，这样组合灵活一点？

---

### DQ10：限流实现方案

#### DQ10.1 限流技术选型

[Question] NFR-1.4 和 NFR-1.5 要求对登录接口和短信接口实施限流。请确认限流技术方案：
- A：基于 Redis + Lua 脚本实现滑动窗口限流（自行实现）
- B：使用 Guava RateLimiter（单机限流，不支持分布式）
- C：使用 Bucket4j（支持 Redis 分布式限流）
- D：使用 Spring AOP + Redis 自定义 @RateLimit 注解

建议采用方案 D（自定义 @RateLimit 注解 + Redis），原因：
1. 实现灵活，可针对不同接口配置不同的限流策略
2. 基于 Redis 支持分布式场景（虽然当前是单体，但预留扩展能力）
3. 通过注解声明式使用，对业务代码无侵入
4. 底层使用 Redis INCR + EXPIRE 实现固定窗口计数器，简单可靠

[Answer] D

#### DQ10.2 限流维度

[Question] 请确认限流的维度和阈值：
1. 登录接口（`POST /api/v1/user/auth/login`）：按 IP 限流，每分钟 10 次
2. 短信验证码接口（`POST /api/v1/user/auth/sms-code`）：
   - 按手机号限流：60 秒内 1 次（需求已定义）
   - 按手机号限流：每天 10 次（需求已定义）
   - 按 IP 限流：每分钟 10 次（防止同一 IP 对不同手机号发送）
3. Token 刷新接口：是否需要限流？

[Answer] 同意； 需要

---

### DQ11：common 模块变更确认

#### DQ11.1 变更影响评估

[Question] 本 CR 需要对 common 模块进行以下变更，请确认变更范围和影响：

| 变更项 | 变更内容 | 影响范围 |
|--------|---------|---------|
| DataPermissionLevel | 增加 SELF 枚举值 | 所有使用 DataPermissionLevel 的模块 |
| DataPermissionContext | 单 ID → ID 集合，增加 userId/username | DataPermissionInterceptor、所有设置上下文的代码 |
| DataPermissionInterceptor | 支持 SELF 级别、IN 查询、OR 条件 | 所有标注 @DataPermission 的 Mapper 方法 |
| @DataPermission 注解 | 增加 selfField 参数 | 无影响（新增参数有默认值） |
| SecurityConfig | CSRF 策略（待确认 DQ4.1） | 前端请求方式 |
| ErrorCode | 增加 user 模块错误码段（2000-2999） | 无影响（新增枚举值） |

由于 DataPermissionContext 的变更是 breaking change，是否需要：
1. 在 common 模块中保持向后兼容（旧的单 ID 字段标记 @Deprecated，新增集合字段）
2. 直接替换为新结构（当前只有 common 模块使用，无其他业务模块依赖）

建议采用方案 2（直接替换），因为当前没有其他业务模块已经上线。

[Answer] 这个参考前面说的方案调整优化，需要保持通用性，后续扩展或者移植到其他项目时可以作为更通用组件

---

### DQ12：其他设计问题

#### DQ12.1 用户-数据权限关联查询

[Question] R2.1 用户列表查询需要应用数据权限过滤。用户表（sys_user）本身没有 area_id、community_id、building_id 字段，数据权限是通过角色关联的。请确认用户列表的数据权限过滤方案：
- A：在 sys_user 表增加冗余字段（area_id、community_id、building_id），用户角色变更时同步更新
- B：通过 SQL JOIN 关联 sys_user_role → sys_role → sys_role_data_permission 表进行过滤
- C：用户列表不使用 DataPermissionInterceptor 自动过滤，而是在 Service 层手动实现过滤逻辑

建议采用方案 C（Service 层手动过滤），原因：
1. 用户表的数据权限过滤逻辑较复杂（需要关联多张表）
2. DataPermissionInterceptor 适用于业务表（如工单表有 community_id 字段）的简单过滤
3. 用户管理是低频操作，手动过滤不影响性能

[Answer] C， 这里如果我们的数据权限模型有通用性，在这种场景处理可以更灵活适配

#### DQ12.2 密码加密存储的 BCrypt 轮次

[Question] BCryptPasswordEncoder 默认使用 10 轮加密。是否需要调整轮次？
- 10 轮（默认）：加密耗时约 100ms，安全性足够
- 12 轮：加密耗时约 400ms，更安全但影响登录性能

建议保持默认 10 轮。

[Answer] 确认

#### DQ12.3 菜单树缓存策略

[Question] R4.1 菜单树查询和 R4.4 当前用户菜单树查询是高频接口（每次页面加载都会调用）。请确认缓存策略：
- A：缓存完整菜单树（所有菜单），用户请求时从缓存中按权限过滤
- B：按用户缓存菜单树（每个用户一份缓存）
- C：不缓存菜单树（菜单数据量小，直接查数据库）

建议采用方案 B（按用户缓存），原因：
1. 菜单树按用户权限不同，结果不同
2. 缓存后减少数据库查询和树构建的开销
3. 菜单变更或权限变更时清除相关用户的缓存

[Answer] B

---

## 四、已有基础设施确认

以下 common 模块组件将在 user 模块中直接使用，无需重新实现：

| 组件 | 用途 | 使用场景 |
|------|------|---------|
| Result / PageResult | 统一响应包装 | 所有 Controller 返回值 |
| ErrorCode / BusinessException | 错误码和业务异常 | 业务校验失败时抛出 |
| GlobalExceptionHandler | 全局异常处理 | 自动捕获并转换异常 |
| BaseEntity | 公共字段 + 自动填充 | 所有 entity 继承 |
| JwtTokenProvider | JWT 令牌生成/解析 | 登录认证、Token 刷新 |
| JwtAuthenticationFilter | JWT 认证过滤器 | 自动认证已登录用户 |
| SecurityConfig | Spring Security 配置 | 白名单、认证规则 |
| BCryptPasswordEncoder | 密码加密 | 密码存储和校验 |
| RedisUtils | Redis 缓存操作 | 缓存、限流、验证码 |
| EncryptUtils | AES 加解密 | 手机号加密存储 |
| DataPermissionInterceptor | 数据权限 SQL 拦截 | 业务数据查询过滤（需扩展） |
| @DataPermission | 数据权限注解 | 标记需要数据权限的 Mapper 方法（需扩展） |
| CommonMetaObjectHandler | 自动填充 | createTime/updateTime/createBy/updateBy |

---

## 五、设计输出物预览

最终 design.md 将包含以下章节：

```
design.md
├── 概述（技术栈、设计目标、模块定位）
├── 架构（模块架构图、内部分层、请求处理流程）
├── 数据库设计
│   ├── ER 图
│   ├── 8 张表的完整 DDL（含索引、约束）
│   └── 初始化 DML（管理员账号、预置角色、预置菜单）
├── 包结构设计（DDD 分层）
├── 组件与接口设计
│   ├── 认证模块（AuthController、AuthService）
│   ├── 用户管理（UserController、UserService）
│   ├── 角色管理（RoleController、RoleService）
│   ├── 菜单管理（MenuController、MenuService）
│   ├── 数据权限（DataPermissionService）
│   ├── 个人中心（ProfileController、ProfileService）
│   ├── 登录日志（LoginLogService）
│   ├── 操作日志（OperationLogService、@OperationLog、OperationLogAspect）
│   └── 对外 API（UserApi、PermissionApi）
├── 安全设计
│   ├── CSRF 策略
│   ├── @PreAuthorize 权限校验
│   ├── Token 黑名单
│   └── 密码策略
├── 缓存设计
│   ├── Redis key 规范
│   ├── 缓存策略
│   └── 降级方案
├── 限流设计
├── common 模块变更清单
├── API 接口契约（请求/响应 DTO 定义）
└── 正确性属性（Correctness Properties）
```

---

**请审阅以上设计计划和澄清问题，在每个 [Answer] 处填写您的回答。确认后我将按步骤逐步生成 design.md。**
