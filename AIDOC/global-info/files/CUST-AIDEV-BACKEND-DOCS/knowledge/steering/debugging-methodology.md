---
inclusion: always
---

# 排查方法论

> 本规则基于实际排查经验总结，新会话时自动加载，避免重复犯同样的错误。

---

## 核心原则

**先看异常堆栈/日志，再读业务代码。** 如果无法获取日志，则按下面的优先级顺序排查。

---

## "系统繁忙/系统内部错误" 排查清单

当用户反馈接口返回"系统繁忙"或类似兜底错误信息时，说明抛出了**非 BusinessException 的未知异常**。排查优先级：

### 第一优先级：基础设施层（Controller 之前）

这些问题发生在请求到达业务代码之前，最容易被忽略：

1. **参数反序列化问题**（最常见）
   - DTO 中有 `LocalDateTime` / `LocalDate` / `LocalTime` 字段，前端传字符串
   - 检查 `WebMvcConfig` / `ObjectMapper` 配置中是否注册了 `JavaTimeModule`
   - 检查反序列化格式是否与前端传来的格式一致（如 `yyyy-MM-dd HH:mm:ss`）
   - DTO 中有 `BigDecimal` / `Enum` 等特殊类型，前端传了不匹配的值

2. **参数校验失败但未被正确捕获**
   - `@Valid` / `@NotNull` / `@Size` 等注解触发的异常类型

3. **过滤器/拦截器链异常**
   - `JwtAuthenticationFilter`、数据权限拦截器等

### 第二优先级：AOP 切面层

4. **切面逻辑异常**
   - `@OperationLog` 切面中的 SpEL 解析、序列化失败
   - `@RateLimit` 限流切面
   - `@PreAuthorize` 权限校验中的 NPE

### 第三优先级：业务逻辑层

5. **业务代码 NPE 或运行时异常**
   - 枚举 `fromCode()` 返回 null 后继续调用方法
   - 集合操作未判空
   - 类型转换错误

---

## 高频陷阱速查表

| 场景 | 典型异常 | 根因 |
|------|---------|------|
| 前端传日期字符串，后端 DTO 是 `LocalDateTime` | `InvalidFormatException` | 未注册 `JavaTimeModule` 或格式不匹配 |
| 前端传数字，后端 DTO 是 `String` | 正常 | Spring 会自动转换 |
| 前端传 null，后端 DTO 是基本类型 `int`/`long` | `InvalidDefinitionException` | 用包装类 `Integer`/`Long` 替代 |
| 枚举字段前端传了不存在的值 | `InvalidFormatException` | 加 `@JsonCreator` 或全局异常处理 |
| `@RequestBody` 缺少 Content-Type 头 | `HttpMediaTypeNotSupportedException` | 前端未设置 `Content-Type: application/json` |

---

## 排查步骤模板

当接到"接口报错"类问题时，按以下步骤执行：

```
1. 确认错误现象：具体错误信息、HTTP 状态码、请求参数
2. 定位接口入口：Controller 方法和 URL
3. 读取 DTO 定义：重点关注特殊类型字段（LocalDateTime、Enum、BigDecimal）
4. 检查基础设施配置：WebMvcConfig、ObjectMapper、Jackson 模块
5. 检查全局异常处理器：确认错误消息对应哪个 catch 分支
6. 只有在前 5 步都没问题时，才深入业务逻辑
```

---

## 经验教训记录

### 2026-04-19：派单"系统繁忙"问题
- **现象**：点击派单确认按钮返回"系统繁忙，请稍后重试"
- **根因**：`WorkMvcConfig` 自定义了 `ObjectMapper` 但未注册 `JavaTimeModule`，导致 `WorkOrderDispatchDTO.expectedCompleteTime`（`LocalDateTime` 类型）无法从前端传来的 `"YYYY-MM-DD HH:mm:ss"` 字符串反序列化
- **修复**：在 ObjectMapper 中注册 `JavaTimeModule`，配置 `LocalDateTimeSerializer` 和 `LocalDateTimeDeserializer`，格式统一为 `yyyy-MM-dd HH:mm:ss`
- **教训**：看到 DTO 中的 `LocalDateTime` 字段应立即警觉反序列化问题，而不是跳过去查业务逻辑
