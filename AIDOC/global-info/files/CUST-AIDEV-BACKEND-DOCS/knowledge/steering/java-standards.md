---
inclusion: fileMatch
fileMatchPattern: "**/*.java"
---

# Java 开发规范

当处理 Java 文件时，自动应用以下标准和约定。

## 项目架构约定

### 架构模式
- 使用六边形架构（Hexagonal Architecture）处理复杂领域
- 在需要读写分离时实现 CQRS 模式
- 业务逻辑应用领域驱动设计（DDD）原则

### 测试策略
- 所有业务逻辑必须编写单元测试
- 集成测试使用 TestContainers
- 保持最低 80% 代码覆盖率（核心模块要求更高）
- 遵循 AAA 模式（Arrange, Act, Assert）

### 错误处理
- 业务逻辑错误使用自定义异常
- 使用 @ControllerAdvice 实现全局异常处理器
- 错误日志必须包含 correlation ID 以便追溯
- 返回一致的错误响应格式

### 性能指南
- 数据库访问使用连接池
- 为频繁访问的数据实现缓存策略
- 长时间运行的操作使用异步处理
- 监控并优化数据库查询

## 编码标准

完整的 Java 编码标准请参考：
#[[file:.kiro/skills/java-standard/SKILL.md]]

### 快速参考

#### 命名规范
- 类名：UpperCamelCase（例外：DO/BO/DTO/VO/AO/PO/UID）
- 方法/变量：lowerCamelCase
- 常量：全大写下划线分隔（MAX_STOCK_COUNT）
- 包名：全小写单数形式

#### Service/DAO 层方法命名
- 获取单个对象：`get` 前缀
- 获取多个对象：`list` 前缀 + 复数形式
- 获取统计值：`count` 前缀
- 插入：`save/insert` 前缀
- 删除：`remove/delete` 前缀
- 修改：`update` 前缀

#### OOP 规约
- 通过类名访问静态变量/方法
- 覆写方法必须加 @Override
- equals 比较用常量或确定有值对象调用
- 包装类对象值比较使用 equals（不用 ==）
- POJO 类必须写 toString 方法

#### 集合处理
- 重写 equals 必须重写 hashCode
- 使用 Iterator 方式在循环中 remove/add
- 集合初始化指定初始值大小
- 使用 entrySet 遍历 Map（而非 keySet）

#### 并发处理
- 单例对象和方法必须保证线程安全
- 线程资源必须通过线程池提供
- 使用 ThreadPoolExecutor 创建线程池（禁止 Executors）
- SimpleDateFormat 线程不安全，使用 ThreadLocal 或 DateUtils
- 多资源加锁保持一致顺序（避免死锁）

#### 异常处理
- 可预检查规避的 RuntimeException 不应通过 catch 处理
- 异常不要用作流程控制
- catch 异常必须处理或抛给调用者
- 防止 NPE：级联调用、远程调用、数据库查询结果都要检查

#### 日志规约
- 使用 SLF4J API（不直接使用 Log4j/Logback）
- trace/debug/info 级别使用条件输出或占位符
- 异常信息包括案发现场和异常堆栈
- 生产环境禁止输出 debug 日志

#### 单元测试
- 遵守 AIR 原则（Automatic/Independent/Repeatable）
- 使用 assert 验证（禁止 System.out）
- 保持测试独立性，不依赖执行顺序
- 语句覆盖率 70%，核心模块 100%

#### 安全规约
- 用户页面/功能必须进行权限控制
- 敏感数据必须脱敏展示
- SQL 参数严格使用参数绑定（禁止字符串拼接）
- 用户输入必须做有效性验证
- 表单/AJAX 提交必须执行 CSRF 验证

## 应用分层

```
开放接口层: 封装 Service 暴露 RPC 接口/Web 封装 HTTP 接口
终端显示层: 各端模板渲染并执行显示
Web 层: 访问控制转发/基本参数校验
Service 层: 相对具体业务逻辑服务层
Manager 层: 通用业务处理层
DAO 层: 数据访问层
外部接口层: 第三方平台/其它部门 RPC 接口
```

## 领域模型

- **DO (Data Object)**: 与数据库表结构一一对应
- **DTO (Data Transfer Object)**: Service 或 Manager 向外传输对象
- **BO (Business Object)**: Service 层输出封装业务逻辑对象
- **VO (View Object)**: Web 向模板渲染引擎层传输对象
- **Query**: 各层接收上层查询请求（超过 2 个参数必须封装）

## 设计原则

- 单一职责原则（SRP）
- 开闭原则（对扩展开放，对修改闭合）
- 里氏替换原则（优先聚合/组合而非继承）
- 依赖倒置原则（依赖抽象类与接口）
- 接口隔离原则

## 常见问题速查

| 问题类型 | 参考章节 |
|---------|---------|
| 命名问题 | 命名规范 |
| 类设计问题 | OOP 规约 |
| 集合使用 | 集合处理 |
| 并发问题 | 并发处理 |
| 逻辑控制 | 控制语句 |
| 注释缺失 | 注释规约 |
| 异常处理 | 异常处理 |
| 日志记录 | 日志规约 |
| 测试覆盖 | 单元测试 |
| 安全漏洞 | 安全规约 |
| 架构设计 | 工程结构/设计规约 |

## 相关文档

- [Java Standard SKILL.md](../skills/java-standard/SKILL.md) - 完整编码规范
- [Spring Boot Patterns](./spring-boot-patterns.md) - Spring Boot 开发模式
- [Automated Testing](./automated-api-testing.md) - 自动化测试规范
- [SonarQube Quality](./sonarqube-quality.md) - 代码质量标准
