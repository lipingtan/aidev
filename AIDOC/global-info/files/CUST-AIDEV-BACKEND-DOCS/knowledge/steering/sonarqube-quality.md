---
inclusion: manual
description: SonarQube 代码质量管理规范。涵盖代码复杂度、Bug检测、安全漏洞、代码异味、代码重复、测试覆盖率和代码风格。
keywords: sonarqube, code quality, complexity, bugs, security, code smells, coverage, static analysis
---

# SonarQube 代码质量管理规范

## 概述

本规范基于 SonarQube 代码质量管理平台,涵盖多语言代码质量标准和最佳实践。用于编写代码、代码审查、重构和提升代码质量。

## 严重程度分级

- **Blocker（阻断）**: 必须修复,严重影响代码质量
- **Critical（严重）**: 应该修复,存在潜在风险
- **Major（重要）**: 建议修复,提升代码质量
- **Minor（次要）**: 可选修复,代码改进建议

## 核心质量门禁标准

### 新代码质量门禁（推荐配置）

```yaml
quality_gate:
  # 覆盖率要求
  coverage:
    line_coverage: ">= 80%"      # 行覆盖率
    branch_coverage: ">= 70%"    # 分支覆盖率
    new_code_coverage: ">= 85%"  # 新代码覆盖率

  # Bug/漏洞限制
  issues:
    blocker_bugs: 0              # 阻断Bug必须为0
    critical_bugs: 0             # 严重Bug必须为0
    security_hotspots: 0         # 安全热点必须为0
    new_vulnerabilities: 0       # 新增漏洞必须为0

  # 代码重复
  duplication:
    duplicated_lines_density: "< 3%"  # 重复行密度
    new_duplicated_lines: "< 5%"      # 新代码重复率

  # 可靠性
  reliability:
    reliability_rating: "A"      # 可靠性评级 A-D

  # 安全性
  security:
    security_rating: "A"         # 安全性评级 A-D

  # 可维护性
  maintainability:
    sqale_rating: "A"            # 可维护性评级 A-D
    code_smells: 0               # 新增代码异味
```

## 核心原则（Blocker/Critical）

### 1. 空指针安全（Blocker）
- 对可能为null的对象调用方法前必须检查
- 使用`Optional`代替可能为null的返回值
- 遵循"Fail Fast"原则

```java
// ❌ 错误
public String getUserName(User user) {
    return user.getName(); // 可能NPE
}

// ✅ 正确
public String getUserName(User user) {
    if (user == null) {
        throw new IllegalArgumentException("User cannot be null");
    }
    return user.getName();
}

// ✅ 更好
public Optional<String> getUserName(User user) {
    return Optional.ofNullable(user)
                   .map(User::getName);
}
```

### 2. 资源管理（Blocker）
- IO流、数据库连接等资源必须在finally块中关闭
- 优先使用try-with-resources
- 不要忽略异常

```java
// ❌ 错误
FileInputStream fis = new FileInputStream("file.txt");
// ... 使用 fis
fis.close(); // 异常时不会执行

// ✅ 正确
try (FileInputStream fis = new FileInputStream("file.txt")) {
    // ... 使用 fis
} // 自动关闭
```

### 3. 并发安全（Critical）
- 共享可变状态必须正确同步
- 不要在同步块中调用可变方法
- 避免死锁:按固定顺序获取锁

```java
// ❌ 错误
private int counter = 0;
public void increment() {
    counter++; // 非线程安全
}

// ✅ 正确
private final AtomicInteger counter = new AtomicInteger(0);
public void increment() {
    counter.incrementAndGet();
}
```

### 4. 安全漏洞（Blocker）
- 禁止硬编码密码/密钥
- SQL查询必须参数化
- 用户输入必须验证和转义
- 敏感数据必须加密存储

```java
// ❌ 错误
String password = "admin123"; // 硬编码密码
String sql = "SELECT * FROM users WHERE name = '" + userName + "'"; // SQL注入

// ✅ 正确
String password = System.getenv("DB_PASSWORD"); // 从环境变量读取
String sql = "SELECT * FROM users WHERE name = ?"; // 参数化查询
```

### 5. 代码复杂度（Critical）
- 单方法圈复杂度不超过10
- 单方法认知复杂度不超过15
- 单个类不超过500行

```java
// ❌ 错误 - 复杂度过高
public void processOrder(Order order) {
    if (order != null) {
        if (order.getStatus() == Status.PENDING) {
            if (order.getAmount() > 0) {
                if (order.getCustomer() != null) {
                    // ... 嵌套过深
                }
            }
        }
    }
}

// ✅ 正确 - 提前返回,降低复杂度
public void processOrder(Order order) {
    if (order == null) return;
    if (order.getStatus() != Status.PENDING) return;
    if (order.getAmount() <= 0) return;
    if (order.getCustomer() == null) return;
    
    // ... 处理逻辑
}
```

### 6. 代码重复（Major）
- 重复代码块不超过10行
- 整体重复率不超过3%

```java
// ❌ 错误 - 代码重复
public void processUserOrder(User user) {
    if (user == null) throw new IllegalArgumentException();
    if (user.getOrder() == null) throw new IllegalArgumentException();
    // ... 处理逻辑
}

public void processAdminOrder(Admin admin) {
    if (admin == null) throw new IllegalArgumentException();
    if (admin.getOrder() == null) throw new IllegalArgumentException();
    // ... 处理逻辑（重复）
}

// ✅ 正确 - 提取公共方法
private void validateOrderOwner(OrderOwner owner) {
    if (owner == null) throw new IllegalArgumentException();
    if (owner.getOrder() == null) throw new IllegalArgumentException();
}

public void processUserOrder(User user) {
    validateOrderOwner(user);
    // ... 处理逻辑
}

public void processAdminOrder(Admin admin) {
    validateOrderOwner(admin);
    // ... 处理逻辑
}
```

## 使用场景

### 编写新代码时
1. 参考对应领域的规范文件
2. 遵循Blocker和Critical级别规则
3. 控制方法复杂度,保持代码简洁
4. 编写单元测试,确保覆盖率

### 代码审查时
1. 使用规范作为检查清单
2. 重点检查Blocker和Critical级别问题
3. 关注安全漏洞和资源泄漏
4. 评估代码可维护性

### 重构代码时
1. 优先处理高复杂度方法
2. 消除代码重复
3. 修复安全漏洞
4. 提升测试覆盖率

### SonarQube扫描

```bash
# Maven项目
mvn clean verify sonar:sonar \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=your_token \
  -Dsonar.qualitygate.wait=true

# Gradle项目
./gradlew sonarqube \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=your_token

# .NET项目
dotnet sonarscanner begin \
  /k:"project-key" \
  /d:sonar.host.url="http://localhost:9000"
dotnet build
dotnet sonarscanner end
```

## 质量评级说明

| 评级 | 可靠性/安全性/可维护性 | 说明 |
|------|----------------------|------|
| A | 0-5% | 优秀 |
| B | 5-10% | 良好 |
| C | 10-20% | 一般 |
| D | 20-50% | 较差 |
| E | 50-100% | 差 |

## 规范详解

详见各领域参考文件:
- **代码复杂度**: [complexity.md](../skills/sonarqube/references/complexity.md)
- **Bug检测**: [bugs.md](../skills/sonarqube/references/bugs.md)
- **安全漏洞**: [security.md](../skills/sonarqube/references/security.md)
- **代码异味**: [code-smells.md](../skills/sonarqube/references/code-smells.md)
- **重复与风格**: [duplication-style.md](../skills/sonarqube/references/duplication-style.md)
- **测试覆盖率**: [test-coverage.md](../skills/sonarqube/references/test-coverage.md)

## 检查清单

### 提交代码前
- [ ] 无 Blocker 级别问题
- [ ] 无 Critical 级别问题
- [ ] 代码复杂度符合要求
- [ ] 无安全漏洞
- [ ] 测试覆盖率达标
- [ ] 代码重复率 < 3%

### 代码审查时
- [ ] 检查空指针安全
- [ ] 检查资源管理
- [ ] 检查并发安全
- [ ] 检查安全漏洞
- [ ] 评估代码复杂度
- [ ] 识别代码重复

## 相关文档

- [SonarQube SKILL.md](../skills/sonarqube/SKILL.md)
- [Complexity Reference](../skills/sonarqube/references/complexity.md)
- [Bugs Reference](../skills/sonarqube/references/bugs.md)
- [Security Reference](../skills/sonarqube/references/security.md)
- [Code Smells Reference](../skills/sonarqube/references/code-smells.md)
- [Duplication & Style Reference](../skills/sonarqube/references/duplication-style.md)
- [Test Coverage Reference](../skills/sonarqube/references/test-coverage.md)
