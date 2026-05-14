# SPMP user-center 系统集成

> 单体应用架构，模块间通过 `api` 包接口直接注入调用

---

## 一、与 common 模块的依赖关系

user 模块依赖 common 模块（CR01 已完成）提供的基础设施：

| common 组件 | 说明 | user 模块使用方式 |
|-------------|------|------------------|
| SecurityConfig | Spring Security 配置、JWT 无状态认证、白名单路径 | 追加 user 模块认证接口到白名单 |
| JwtTokenProvider | JWT 令牌生成/解析/验证（含 jti） | 登录生成 Token、刷新 Token |
| JwtAuthenticationFilter | JWT 认证过滤器 + Token 黑名单校验 | 每次请求自动校验 |
| DataPermissionInterceptor | 数据权限 SQL 拦截器（五级） | 自动追加数据过滤条件 |
| BCryptPasswordEncoder | 密码编码器 | 密码加密/校验 |
| Result / PageResult | 统一响应包装 | 所有接口返回值 |
| ErrorCode / BusinessException | 错误码和业务异常 | 抛出业务异常 |
| GlobalExceptionHandler | 全局异常处理 | 统一异常转换 |
| BaseEntity | 公共字段 + 自动填充 | DO 实体继承 |
| RedisUtils | Redis 缓存工具 | 缓存、限流、验证码、黑名单 |
| EncryptUtils | 敏感数据加解密（AES） | 手机号加密存储 |

### user 模块对 common 的扩展需求

| 扩展项 | 说明 |
|--------|------|
| DataPermissionLevel 增加 SELF | 五级数据权限支持"仅本人" |
| DataPermissionContext 通用化 | `Map<String, Set<Long>> scopeMap` 动态字段 |
| @DataPermission 增加 selfField | 指定"仅本人"过滤字段（默认 create_by） |
| JwtTokenProvider 增加 jti | Token 黑名单机制 |
| JwtAuthenticationFilter 增加黑名单校验 | 登出/禁用强制下线 |
| SecurityConfig 白名单追加 | user 模块认证接口路径 |
| ErrorCode 增加 2000-2999 段 | user 模块错误码 |

---

## 二、对外提供的 API

定义在 `com.spmp.user.api` 包下，其他模块通过注入接口调用。

### UserApi

```java
public interface UserApi {
    UserBriefDTO getUserById(Long userId);
    List<UserBriefDTO> getUsersByRoleCode(String roleCode);
    List<UserBriefDTO> getUsersByIds(List<Long> userIds);
}
```

实现类：`UserServiceImpl`

### PermissionApi

```java
public interface PermissionApi {
    DataPermissionDTO getDataPermission(Long userId);
    boolean checkPermission(Long userId, String permissionCode);
    List<String> getUserRoles(Long userId);
}
```

实现类：`DataPermissionServiceImpl`

---

## 三、被哪些模块调用

| 调用方模块 | 调用接口 | 使用场景 |
|-----------|---------|---------|
| workorder（工单中心） | `UserApi.getUsersByRoleCode("repairman")` | 查询维修人员列表用于派单 |
| workorder（工单中心） | `UserApi.getUserById(userId)` | 查询工单创建人/处理人信息 |
| workorder（工单中心） | `PermissionApi.getDataPermission(userId)` | 工单列表数据权限过滤 |
| billing（缴费中心） | `UserApi.getUserById(userId)` | 查询操作人信息 |
| billing（缴费中心） | `PermissionApi.getDataPermission(userId)` | 账单数据权限过滤 |
| notice（公告中心） | `UserApi.getUsersByIds(userIds)` | 批量查询公告发布人 |
| notice（公告中心） | `PermissionApi.getDataPermission(userId)` | 公告推送范围过滤 |
| access（门禁中心） | `UserApi.getUserById(userId)` | 查询审批人信息 |
| 所有业务模块 | `PermissionApi.checkPermission(userId, permCode)` | 业务层权限校验 |

---

## 四、短信服务集成

### 当前实现（模拟）

```java
@Slf4j
@Service
public class SmsServiceMockImpl implements SmsService {
    @Override
    public boolean sendVerificationCode(String phone, String code) {
        log.info("【短信模拟】向手机号 {} 发送验证码: {}", mask(phone), code);
        return true;
    }
}
```

- 当前阶段仅日志输出，不实际发送短信
- 验证码存入 Redis，与短信发送解耦

### 预留接口

```java
public interface SmsService {
    boolean sendVerificationCode(String phone, String code);
}
```

后续对接真实短信平台时，只需新增实现类替换 Mock 实现，无需修改业务逻辑。

---

## 五、集成约束

| 约束 | 说明 |
|------|------|
| 调用方式 | 单体应用内 Spring 注入，禁止 Feign/HTTP 调用 |
| DTO 位置 | 跨模块 DTO 定义在 `com.spmp.user.api.dto` |
| 禁止事项 | 其他模块禁止直接注入 user 模块的 ServiceImpl |
| 基础数据 | 片区/小区/楼栋数据本期硬编码，后续对接 base 模块 BaseApi |
