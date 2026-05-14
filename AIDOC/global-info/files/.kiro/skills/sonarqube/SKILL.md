---
name: sonarqube
description: SonarQube代码质量管理规范。涵盖代码复杂度、Bug检测、安全漏洞、代码异味、代码重复、测试覆盖率和代码风格。当编写代码、进行代码审查、重构或需要提升代码质量时使用此skill。
---

# SonarQube代码质量管理规范

本规范基于SonarQube代码质量管理平台，涵盖多语言代码质量标准和最佳实践。严格程度分为三级：
- **Blocker（阻断）**：必须修复，严重影响代码质量
- **Critical（严重）**：应该修复，存在潜在风险
- **Major（重要）**：建议修复，提升代码质量
- **Minor（次要）**：可选修复，代码改进建议

## 快速参考

### 规范结构

| 类别 | 文件 | 说明 |
|------|------|------|
| 代码复杂度 | [complexity.md](references/complexity.md) | 圈复杂度、认知复杂度、方法/类复杂度控制 |
| Bug检测 | [bugs.md](references/bugs.md) | 空指针、资源泄漏、并发问题等常见Bug |
| 安全漏洞 | [security.md](references/security.md) | SQL注入、XSS、加密、权限等安全问题 |
| 代码异味 | [code-smells.md](references/code-smells.md) | 冗余代码、设计问题、可维护性问题 |
| 重复与风格 | [duplication-style.md](references/duplication-style.md) | 代码重复率、命名规范、代码格式 |
| 测试覆盖率 | [test-coverage.md](references/test-coverage.md) | 单元测试覆盖率、测试质量 |

### 核心质量门禁标准

```yaml
# 新代码质量门禁（推荐配置）
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

### 核心原则（Blocker/Critical）

#### 1. 空指针安全（Blocker）
- 对可能为null的对象调用方法前必须检查
- 使用`Optional`代替可能为null的返回值
- 遵循"Fail Fast"原则

#### 2. 资源管理（Blocker）
- IO流、数据库连接等资源必须在finally块中关闭
- 优先使用try-with-resources
- 不要忽略异常

#### 3. 并发安全（Critical）
- 共享可变状态必须正确同步
- 不要在同步块中调用可变方法
- 避免死锁：按固定顺序获取锁

#### 4. 安全漏洞（Blocker）
- 禁止硬编码密码/密钥
- SQL查询必须参数化
- 用户输入必须验证和转义
- 敏感数据必须加密存储

#### 5. 代码复杂度（Critical）
- 单方法圈复杂度不超过10
- 单方法认知复杂度不超过15
- 单个类不超过500行

#### 6. 代码重复（Major）
- 重复代码块不超过10行
- 整体重复率不超过3%

## 使用场景

### 编写新代码时

1. 参考对应领域的规范文件
2. 遵循Blocker和Critical级别规则
3. 控制方法复杂度，保持代码简洁
4. 编写单元测试，确保覆盖率

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

### SonarQube扫描时

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

详见各领域参考文件：
- 代码复杂度详见 [complexity.md](references/complexity.md)
- Bug检测详见 [bugs.md](references/bugs.md)
- 安全漏洞详见 [security.md](references/security.md)
- 代码异味详见 [code-smells.md](references/code-smells.md)
- 重复与风格详见 [duplication-style.md](references/duplication-style.md)
- 测试覆盖率详见 [test-coverage.md](references/test-coverage.md)
