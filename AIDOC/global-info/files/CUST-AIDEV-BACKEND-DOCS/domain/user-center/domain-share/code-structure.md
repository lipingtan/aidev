# SPMP user-center 代码结构

> 包名：`com.spmp.user`  
> 架构：Service + Mapper 分层模式（非 DDD 聚合根）

---

## 一、包结构

```
com.spmp.user/
├── api/                          # 对外 API 接口（供其他模块调用）
│   ├── UserApi.java
│   ├── PermissionApi.java
│   └── dto/
│       ├── UserBriefDTO.java
│       └── DataPermissionDTO.java
├── controller/                   # HTTP 接口层
│   ├── AuthController.java
│   ├── UserController.java
│   ├── RoleController.java
│   ├── MenuController.java
│   ├── ProfileController.java
│   ├── LoginLogController.java
│   └── OperationLogController.java
├── service/                      # 业务逻辑层
│   ├── AuthService.java
│   ├── UserService.java
│   ├── RoleService.java
│   ├── MenuService.java
│   ├── ProfileService.java
│   ├── DataPermissionService.java
│   ├── LoginLogService.java
│   ├── OperationLogService.java
│   ├── SmsService.java           # 短信服务接口（预留）
│   ├── PermissionCacheService.java
│   └── impl/
│       ├── AuthServiceImpl.java
│       ├── UserServiceImpl.java          # 同时实现 UserApi
│       ├── RoleServiceImpl.java
│       ├── MenuServiceImpl.java
│       ├── ProfileServiceImpl.java
│       ├── DataPermissionServiceImpl.java  # 同时实现 PermissionApi
│       ├── LoginLogServiceImpl.java
│       ├── OperationLogServiceImpl.java
│       ├── SmsServiceMockImpl.java       # 短信模拟实现
│       └── PermissionCacheServiceImpl.java
├── domain/                       # 领域模型
│   ├── entity/                   # DO 数据库实体
│   │   ├── UserDO.java
│   │   ├── RoleDO.java
│   │   ├── MenuDO.java
│   │   ├── UserRoleDO.java
│   │   ├── RoleMenuDO.java
│   │   ├── RoleDataPermissionDO.java
│   │   ├── LoginLogDO.java
│   │   └── OperationLogDO.java
│   ├── dto/                      # 数据传输对象
│   │   ├── LoginDTO.java
│   │   ├── SmsLoginDTO.java
│   │   ├── TokenDTO.java
│   │   ├── CaptchaDTO.java
│   │   ├── UserCreateDTO.java
│   │   ├── UserUpdateDTO.java
│   │   ├── UserQueryDTO.java
│   │   ├── UserPageDTO.java
│   │   ├── UserDetailDTO.java
│   │   ├── RoleCreateDTO.java
│   │   ├── RoleUpdateDTO.java
│   │   ├── RoleQueryDTO.java
│   │   ├── RolePageDTO.java
│   │   ├── MenuCreateDTO.java
│   │   ├── MenuUpdateDTO.java
│   │   ├── MenuTreeDTO.java
│   │   ├── DataPermissionConfigDTO.java
│   │   ├── ProfileDTO.java
│   │   ├── ProfileUpdateDTO.java
│   │   ├── PasswordUpdateDTO.java
│   │   ├── LoginLogQueryDTO.java
│   │   └── OperationLogQueryDTO.java
│   └── vo/
│       ├── DataPermissionOptionVO.java
│       └── RoleSimpleVO.java
├── repository/                   # 数据访问层（MyBatis Mapper）
│   ├── UserMapper.java
│   ├── RoleMapper.java
│   ├── MenuMapper.java
│   ├── UserRoleMapper.java
│   ├── RoleMenuMapper.java
│   ├── RoleDataPermissionMapper.java
│   ├── LoginLogMapper.java
│   └── OperationLogMapper.java
├── config/                       # 模块级配置
│   ├── AsyncConfig.java          # 异步线程池
│   └── UserSecurityConfig.java
├── constant/                     # 常量和枚举
│   ├── UserErrorCode.java        # 错误码（2000-2999）
│   ├── UserConstants.java
│   └── OperationType.java
├── annotation/                   # 自定义注解
│   ├── OperationLog.java         # 操作日志注解
│   └── RateLimit.java            # 限流注解
├── aspect/                       # AOP 切面
│   ├── OperationLogAspect.java   # 操作日志切面（@Async）
│   └── RateLimitAspect.java      # 限流切面
└── security/
    └── PermissionService.java    # @perm.check() 权限校验 Bean
```

---

## 二、命名规范

| 类型 | 后缀 | 示例 | 说明 |
|------|------|------|------|
| 数据库实体 | DO | UserDO、RoleDO | 表映射对象 |
| 数据传输对象 | DTO | UserCreateDTO、LoginDTO | 入参/出参 |
| 值对象 | VO | DataPermissionOptionVO | 展示用值对象 |
| 服务接口 | Service | UserService | 业务逻辑接口 |
| 服务实现 | ServiceImpl | UserServiceImpl | 业务逻辑实现 |
| 数据访问 | Mapper | UserMapper | MyBatis Mapper |
| 控制器 | Controller | UserController | HTTP 接口 |
| 对外接口 | Api | UserApi | 跨模块接口 |

---

## 三、分层职责

| 层 | 包 | 职责 |
|---|---|---|
| Controller | `controller` | HTTP 接口、参数校验（@Valid）、调用 Service |
| Service | `service` / `service/impl` | 业务逻辑编排、缓存管理、权限校验 |
| Repository | `repository` | MyBatis Mapper 接口、数据访问 |
| Domain | `domain/entity` | DO 数据库实体 |
| Domain | `domain/dto` | 数据传输对象（入参/出参） |
| Domain | `domain/vo` | 值对象 |
| API | `api` / `api/dto` | 对外接口定义 + 跨模块 DTO |
| Config | `config` | 模块级配置（线程池、安全扩展） |
| Constant | `constant` | 枚举、常量、错误码 |
| Annotation | `annotation` | 自定义注解 |
| Aspect | `aspect` | AOP 切面 |

---

## 四、模块间通信方式

单体应用内，其他模块通过注入 `api` 包下的接口调用 user 模块能力：

```java
// 其他模块中注入
@Autowired
private UserApi userApi;

@Autowired
private PermissionApi permissionApi;
```

- `UserApi` 由 `UserServiceImpl` 实现
- `PermissionApi` 由 `DataPermissionServiceImpl` 实现
- 禁止其他模块直接注入 user 模块的 Service 实现类
