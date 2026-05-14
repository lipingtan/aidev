---
inclusion: always
---

# Java 编码规范

## 命名规范

**类名：**
- 使用 UpperCamelCase（大驼峰）命名
- 类名应为名词或名词短语，清晰表达类的职责
- 示例：`UserService`、`OrderController`、`PaymentDTO`

**方法名：**
- 使用 lowerCamelCase（小驼峰）命名
- 方法名应为动词或动词短语
- 布尔类型方法使用 `is`、`has`、`can` 前缀
- 示例：`getUserById()`、`createOrder()`、`isActive()`、`hasPermission()`

**变量名：**
- 使用 lowerCamelCase 命名，名称应有实际含义，禁止使用 `a`、`b`、`tmp` 等无意义名称
- 集合类型变量名使用复数形式，如 `userList`、`orderIds`
- 示例：`userId`、`orderStatus`、`pageSize`

**常量命名：**
- 使用全大写字母 + 下划线分隔（SCREAMING_SNAKE_CASE）
- 必须使用 `static final` 修饰
- 示例：`MAX_RETRY_COUNT`、`DEFAULT_PAGE_SIZE`、`ORDER_STATUS_PENDING`

**包名：**
- 全部小写，使用公司域名倒置 + 项目名 + 模块名
- 示例：`com.example.project.user.service`

## 代码格式规范

**缩进：**
- 使用 4 个空格缩进，禁止使用 Tab 字符
- IDE 统一配置 EditorConfig 或 Checkstyle 规则

**行长度：**
- 单行代码不超过 120 个字符
- 超长行在合适位置换行，换行后缩进 8 个空格（相对于当前代码块）

**空行规则：**
- 类的成员变量、构造方法、普通方法之间用 1 个空行分隔
- 方法内部逻辑分组之间用 1 个空行分隔，不得连续出现 2 个以上空行
- 类定义的最后一行不留空行

**其他格式：**
- 左大括号 `{` 不换行，右大括号 `}` 单独占一行
- 运算符前后各加 1 个空格
- 逗号后加 1 个空格，逗号前不加空格
- 每行只声明一个变量

## 注释规范

**Javadoc 格式：**

类和公共方法必须编写 Javadoc 注释：

```java
/**
 * 用户服务接口，提供用户基本信息的增删改查功能。
 *
 * @author 张三
 * @since 1.0.0
 */
public interface UserService {

    /**
     * 根据用户 ID 查询用户信息。
     *
     * @param userId 用户 ID，不能为空
     * @return 用户信息 VO，用户不存在时返回 null
     * @throws BusinessException 当 userId 格式非法时抛出
     */
    UserVO getUserById(Long userId);
}
```

**类注释要求：**
- 说明类的职责和用途
- 标注 `@author` 和 `@since`
- 如有特殊使用注意事项，在注释中说明

**方法注释要求：**
- 公共方法必须有 Javadoc，私有方法视复杂度决定是否添加
- 参数说明使用 `@param`，返回值使用 `@return`，异常使用 `@throws`
- 禁止无意义注释，如 `// 获取用户` 对应 `getUser()` 方法

**行内注释：**
- 复杂业务逻辑、算法、特殊处理必须添加行内注释说明原因
- 注释使用中文，与代码保持同步，禁止保留过时注释

## 异常处理规范

**不得吞异常：**

```java
// 错误示例（严禁）
try {
    doSomething();
} catch (Exception e) {
    // 什么都不做，吞掉异常
}

// 正确示例
try {
    doSomething();
} catch (Exception e) {
    log.error("执行 doSomething 失败，原因：{}", e.getMessage(), e);
    throw new BusinessException("操作失败，请稍后重试", e);
}
```

**统一异常处理：**
- 使用 `@RestControllerAdvice` + `@ExceptionHandler` 实现全局异常处理
- 所有未捕获异常统一在全局处理器中记录日志并返回标准错误响应
- 不得在 Controller 层使用 try-catch 处理业务异常，应让异常向上传播

**自定义业务异常：**
- 业务异常统一继承自 `BusinessException` 基类
- 异常类包含错误码（`errorCode`）和错误信息（`message`）
- 错误码统一在枚举类中定义，禁止硬编码字符串错误码
- 示例：
  ```java
  throw new BusinessException(ErrorCode.USER_NOT_FOUND, "用户不存在，userId: " + userId);
  ```

## 日志规范

**使用 SLF4J：**

```java
// 正确：使用 SLF4J
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

private static final Logger log = LoggerFactory.getLogger(UserServiceImpl.class);

// 或使用 Lombok @Slf4j 注解（推荐）
@Slf4j
public class UserServiceImpl implements UserService { ... }
```

**禁止使用 `System.out.println`：**
- 严禁在任何环境使用 `System.out.println` 或 `System.err.println` 输出日志
- 调试时使用 `log.debug()`，完成后及时清理

**日志级别选择：**

| 级别 | 使用场景 |
|------|----------|
| `ERROR` | 系统异常、需要立即处理的错误 |
| `WARN` | 潜在问题、降级处理、重试等 |
| `INFO` | 关键业务节点、接口调用记录、状态变更 |
| `DEBUG` | 开发调试信息，生产环境不输出 |

**日志格式要求：**
- 使用占位符 `{}` 而非字符串拼接，避免不必要的字符串构造
- 记录异常时必须传入异常对象，以输出完整堆栈：`log.error("msg", e)`
- 关键操作日志需包含业务标识（如 userId、orderId）便于追踪
