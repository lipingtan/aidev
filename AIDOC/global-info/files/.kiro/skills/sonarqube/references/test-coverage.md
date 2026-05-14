# 测试覆盖率规范

测试覆盖率是衡量代码质量的重要指标，包括行覆盖率、分支覆盖率等。

## 覆盖率目标

### 【Critical】规则

1. **新代码覆盖率 >= 80%**

   ```yaml
   # sonar-project.properties
   sonar.coverage.new.lines=80
   sonar.coverage.new.branches=70
   ```

2. **整体代码覆盖率 >= 70%**

   ```yaml
   sonar.coverage.lines=70
   sonar.coverage.branches=60
   ```

3. **关键业务代码覆盖率 >= 90%**

   - 支付/交易相关代码
   - 安全相关代码
   - 数据处理代码

## 覆盖率类型

### 1. 行覆盖率（Line Coverage）

代码被执行的行数比例。

```java
public class Calculator {
    public int add(int a, int b) {
        return a + b;  // 需要测试覆盖此行
    }

    public int divide(int a, int b) {
        if (b == 0) {  // 需要测试
            throw new IllegalArgumentException("Cannot divide by zero");
        }
        return a / b;  // 需要测试
    }
}

// 测试
@Test
void testDivide() {
    calculator.divide(10, 2);  // 覆盖return语句
    assertThrows(IllegalArgumentException.class,
        () -> calculator.divide(10, 0));  // 覆盖异常分支
}
```

### 2. 分支覆盖率（Branch Coverage）

条件语句各分支的执行情况。

```java
public class UserValidator {
    public boolean isValid(User user) {
        if (user == null) return false;        // 分支1
        if (user.getName() == null) return false;  // 分支2
        if (user.getAge() < 18) return false;  // 分支3
        return true;                           // 分支4
    }
}

// 测试：需要覆盖所有分支
@Test
void testIsValid() {
    assertFalse(validator.isValid(null));        // null
    assertFalse(validator.isValid(new User(null, 20)));  // name null
    assertFalse(validator.isValid(new User("Tom", 15)));  // age < 18
    assertTrue(validator.isValid(new User("Tom", 20)));   // valid
}
```

### 3. 路径覆盖率（Path Coverage）

不同执行路径的覆盖情况。

## 测试编写规范

### 【Major】规则

1. **测试命名清晰**

   ```java
   // 推荐命名模式：method_scenario_expectedResult
   @Test
    void divide_byZero_throwsException() { }

   @Test
   void divide_validNumbers_returnsCorrectResult() { }

   @Test
   void divide_negativeNumbers_returnsNegativeResult() { }
   ```

2. **遵循AAA模式**

   ```java
   @Test
   void transfer_sufficientBalance_updatesBalances() {
       // Arrange - 准备测试数据
       Account from = new Account(100);
       Account to = new Account(50);
       TransferService service = new TransferService();

       // Act - 执行被测方法
       service.transfer(from, to, 30);

       // Assert - 验证结果
       assertEquals(70, from.getBalance());
       assertEquals(80, to.getBalance());
   }
   ```

3. **一个测试只验证一个行为**

   ```java
   // 反例：测试多个行为
   @Test
   void userTest() {
       User user = new User("Tom");
       assertEquals("Tom", user.getName());
       user.setAge(20);
       assertEquals(20, user.getAge());
       user.addRole("ADMIN");
       assertTrue(user.hasRole("ADMIN"));
   }

   // 正例：每个测试一个关注点
   @Test
   void constructor_withName_setsName() {
       User user = new User("Tom");
       assertEquals("Tom", user.getName());
   }

   @Test
   void setAge_updatesAge() {
       User user = new User("Tom");
       user.setAge(20);
       assertEquals(20, user.getAge());
   }

   @Test
   void addRole_makesUserHaveRole() {
       User user = new User("Tom");
       user.addRole("ADMIN");
       assertTrue(user.hasRole("ADMIN"));
   }
   ```

4. **使用有意义的断言消息**

   ```java
   // 反例：无消息
   assertEquals(expected, actual);

   // 正例：有消息
   assertEquals("User name should be Tom", expected, actual);

   // 正例：使用Supplier（仅在失败时计算）
   assertEquals(() -> "User name should be Tom", expected, actual);
   ```

## 测试隔离

### 【Critical】规则

1. **测试之间相互独立**

   ```java
   // 反例：测试之间有依赖
   @Test
   void test1() {
       sharedData.add("item");  // 污染后续测试
   }

   // 正例：每个测试独立
   @BeforeEach
   void setUp() {
       sharedData = new ArrayList<>();  // 每次测试前重置
   }

   @Test
   void test1() {
       List<String> data = new ArrayList<>();
       data.add("item");
       assertEquals(1, data.size());
   }
   ```

2. **使用Mock隔离依赖**

   ```java
   @ExtendWith(MockitoExtension.class)
   class UserServiceTest {

       @Mock
       private UserRepository userRepository;

       @InjectMocks
       private UserService userService;

       @Test
       void findById_existingUser_returnsUser() {
           // Arrange
           Long userId = 1L;
           User expectedUser = new User("Tom");
           when(userRepository.findById(userId)).thenReturn(Optional.of(expectedUser));

           // Act
           User actualUser = userService.findById(userId);

           // Assert
           assertEquals(expectedUser, actualUser);
           verify(userRepository).findById(userId);
       }
   }
   ```

## 边界条件测试

### 【Major】规则

1. **测试边界值**

   ```java
   @Test
   void ageValidator_minBoundary() {
       assertFalse(validator.isValid(17));   // 边界下
       assertTrue(validator.isValid(18));    // 边界值
       assertTrue(validator.isValid(19));    // 边界上
   }

   @Test
   void ageValidator_maxBoundary() {
       assertTrue(validator.isValid(99));    // 边界下
       assertTrue(validator.isValid(100));   // 边界值
       assertFalse(validator.isValid(101));  // 边界上
   }
   ```

2. **测试空值/null**

   ```java
   @Test
   void process_withNull_throwsException() {
       assertThrows(IllegalArgumentException.class,
           () -> service.process(null));
   }

   @Test
   void process_withEmptyList_returnsEmpty() {
       List<String> result = service.process(Collections.emptyList());
       assertTrue(result.isEmpty());
   }
   ```

3. **测试异常情况**

   ```java
   @Test
   void divide_byZero_throwsException() {
       IllegalArgumentException ex = assertThrows(
           IllegalArgumentException.class,
           () -> calculator.divide(10, 0)
       );
       assertEquals("Cannot divide by zero", ex.getMessage());
   }
   ```

## TDD最佳实践

### 1. 红绿重构循环

```
Red   → 写一个失败的测试
Green → 写最简单的代码使测试通过
Refactor → 重构代码，保持测试通过
```

### 2. 测试驱动开发示例

```java
// 第1步：写失败的测试（Red）
@Test
void calculateDiscount_vipCustomer_returns20Percent() {
   Customer customer = new Customer(CustomerType.VIP);
   assertEquals(0.2, service.calculateDiscount(customer), 0.001);
}

// 第2步：写最简单的实现（Green）
public double calculateDiscount(Customer customer) {
   return 0.2;  // 硬编码使测试通过
}

// 第3步：重构实现
public double calculateDiscount(Customer customer) {
   return switch (customer.getType()) {
       case VIP -> 0.2;
       case MEMBER -> 0.1;
       case GUEST -> 0.0;
   };
}
```

## 测试配置

### Maven配置

```xml
<plugin>
    <groupId>org.jacoco</groupId>
    <artifactId>jacoco-maven-plugin</artifactId>
    <version>0.8.11</version>
    <executions>
        <execution>
            <goals>
                <goal>prepare-agent</goal>
            </goals>
        </execution>
        <execution>
            <id>report</id>
            <phase>test</phase>
            <goals>
                <goal>report</goal>
            </goals>
        </execution>
        <execution>
            <id>jacoco-check</id>
            <goals>
                <goal>check</goal>
            </goals>
            <configuration>
                <rules>
                    <rule>
                        <element>PACKAGE</element>
                        <limits>
                            <limit>
                                <counter>LINE</counter>
                                <value>COVEREDRATIO</value>
                                <minimum>0.80</minimum>
                            </limit>
                        </limits>
                    </rule>
                </rules>
            </configuration>
        </execution>
    </executions>
</plugin>
```

### Gradle配置

```groovy
jacocoTestCoverageVerification {
    violationRules {
        rule {
            limit {
                minimum = 0.80
            }
        }
    }
}

jacocoTestReport {
    dependsOn test
}
```

## SonarQube配置

```properties
# sonar-project.properties
sonar.core.codeCoveragePlugin=jacoco
sonar.coverage.jacoco.xmlReportPaths=target/site/jacoco/jacoco.xml

# 新代码覆盖率要求
sonar.coverage.new.lines=80
sonar.coverage.new.branches=70

# 整体覆盖率要求
sonar.coverage.lines=70
sonar.coverage.branches=60
```

## SonarQube规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| 缺少测试 | `S2187` (java:S2187) | Major |
| 测试无断言 | `S2699` (java:S2699) | Major |
| 测试覆盖率过低 | 配置 | Critical |
