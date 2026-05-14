---
skillName: code-review-backend
description: Java 后端代码审查技能，检查代码质量、安全性和规范符合度
version: 1.0.0
---

# Code Review 技能

## 使用场景

在以下情况触发此 skill：

- **提交 PR 前**：开发完成后，提交 Pull Request 之前进行自查或请 AI 审查
- **完成任务后**：完成某个开发任务节点，进行阶段性代码质量检查
- **代码重构时**：对已有代码进行重构，确保重构后质量不下降
- **新人代码审查**：对新加入团队成员的代码进行规范性检查

## 核心审查规则

### 命名规范

- 类名使用 `UpperCamelCase`，方法名和变量名使用 `lowerCamelCase`
- 常量全大写，单词间用下划线分隔，如 `MAX_RETRY_COUNT`
- 包名全小写，不使用下划线，如 `com.example.service`
- 布尔类型变量/方法不使用 `is` 前缀（POJO 中除外），如用 `enabled` 而非 `isEnabled`
- 方法名应体现动作，如 `getUserById`、`validateToken`，避免 `doSomething`、`handle` 等模糊命名

### 异常处理

- 不允许捕获异常后直接 `e.printStackTrace()` 或空 catch 块
- 业务异常应使用自定义异常类，继承 `RuntimeException`，包含错误码和错误信息
- 不在 Controller 层直接 try-catch 业务逻辑，应通过全局异常处理器统一处理
- 对外接口必须有兜底异常处理，避免将堆栈信息暴露给调用方
- finally 块中不使用 return 语句

### SQL 安全

- 禁止使用字符串拼接构造 SQL，必须使用 MyBatis 的 `#{}` 参数绑定
- 禁止使用 `${}` 传入用户输入（仅允许用于表名、列名等可信的动态 SQL 场景，且需严格校验）
- 批量操作需评估数据量，超过 500 条应分批处理
- 查询必须有 WHERE 条件限制，禁止全表扫描的更新/删除操作
- 涉及敏感数据（手机号、身份证等）的查询结果需脱敏处理

### 事务使用

- `@Transactional` 注解只能加在 public 方法上
- 事务方法内不应有耗时的远程调用（HTTP、RPC），避免长事务
- 事务传播行为需明确指定，不依赖默认值（`REQUIRED`）时需注释说明原因
- 手动管理事务时，确保 rollback 在 finally 块中执行
- 只读查询使用 `@Transactional(readOnly = true)` 优化性能

### 日志规范

- 使用 SLF4J + Logback，通过 `@Slf4j` 注解注入，不直接使用 `System.out.println`
- 日志级别使用规范：`DEBUG` 调试信息、`INFO` 关键业务节点、`WARN` 可恢复异常、`ERROR` 需人工介入的错误
- 日志中不记录密码、Token、完整身份证号等敏感信息
- 使用占位符格式：`log.info("用户 {} 登录成功", userId)`，不使用字符串拼接
- 接口入参、出参、关键业务决策点必须有 INFO 级别日志

## 审查清单

### 代码质量

- [ ] 方法长度不超过 80 行，超过则考虑拆分
- [ ] 方法参数不超过 5 个，超过则封装为对象
- [ ] 无重复代码（DRY 原则），公共逻辑已抽取为工具方法
- [ ] 无注释掉的废弃代码块
- [ ] 复杂逻辑有清晰的注释说明

### 安全性

- [ ] 所有用户输入已做参数校验（`@Valid`、`@NotNull` 等）
- [ ] SQL 查询使用参数绑定，无 SQL 注入风险
- [ ] 敏感数据已脱敏，不在日志中明文输出
- [ ] 接口有权限校验，无越权访问风险
- [ ] 文件上传有类型和大小限制

### 规范符合度

- [ ] 命名符合团队规范
- [ ] 异常处理符合规范，无空 catch 块
- [ ] 日志记录完整，级别使用正确
- [ ] 事务使用合理，无长事务风险
- [ ] 代码已通过 Checkstyle/SonarQube 静态检查

### 测试覆盖

- [ ] 核心业务逻辑有单元测试
- [ ] 边界条件和异常场景有测试覆盖
- [ ] 测试用例命名清晰，体现测试意图

## 示例

### 好的代码示例

```java
@Slf4j
@Service
@RequiredArgsConstructor
public class UserService {

    private final UserMapper userMapper;
    private final PasswordEncoder passwordEncoder;

    /**
     * 根据用户ID查询用户信息
     *
     * @param userId 用户ID，不能为空
     * @return 用户信息
     * @throws BusinessException 用户不存在时抛出
     */
    public UserDTO getUserById(Long userId) {
        Assert.notNull(userId, "userId 不能为空");
        log.info("查询用户信息, userId={}", userId);

        UserDO user = userMapper.selectById(userId);
        if (user == null) {
            log.warn("用户不存在, userId={}", userId);
            throw new BusinessException(ErrorCode.USER_NOT_FOUND, "用户不存在");
        }

        UserDTO dto = UserConverter.toDTO(user);
        // 手机号脱敏
        dto.setPhone(MaskUtils.maskPhone(dto.getPhone()));
        log.info("查询用户信息成功, userId={}", userId);
        return dto;
    }
}
```

### 需要改进的代码示例

```java
// ❌ 问题：命名不规范、SQL 注入风险、异常处理不当、无日志
public class userservice {

    public Object getUser(String id) {
        try {
            // ❌ 字符串拼接 SQL，存在注入风险
            String sql = "SELECT * FROM user WHERE id = " + id;
            Object result = db.query(sql);
            return result;
        } catch (Exception e) {
            // ❌ 空 catch 块，异常被吞掉
            e.printStackTrace();
        }
        return null;
    }
}

// ✅ 改进后
@Slf4j
@Service
public class UserService {

    public UserDTO getUserById(Long userId) {
        log.info("查询用户, userId={}", userId);
        // ✅ 使用 MyBatis #{} 参数绑定，安全
        UserDO user = userMapper.selectById(userId);
        if (user == null) {
            throw new BusinessException(ErrorCode.USER_NOT_FOUND);
        }
        return UserConverter.toDTO(user);
    }
}
```
