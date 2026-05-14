# 代码复杂度规范

## 圈复杂度（Cyclomatic Complexity）

圈复杂度衡量代码中独立线性路径的数量，值越高表示代码越难以测试和维护。

### 【Blocker】规则

1. **单方法圈复杂度不超过10**
   - 圈复杂度 = 判断语句数量 + 1
   - 包括：if、switch、for、while、catch、条件运算符

   ```java
   // 反例：圈复杂度 = 6（Blocker）
   public int calculate(int a, int b, int c) {
       if (a > 0) {                    // +1
           if (b > 0) {                // +1
               return a + b;
           } else if (c > 0) {         // +1
               return a + c;
           }
       } else if (a < 0) {             // +1
           if (b < 0) {                // +1
               return a - b;
           }
       }
       return 0;
   }

   // 正例：拆分方法，降低复杂度
   public int calculate(int a, int b, int c) {
       if (a > 0) {
           return calculatePositive(a, b, c);
       }
       return calculateNegative(a, b);
   }
   ```

### 【Critical】规则

2. **单文件圈复杂度不超过100**
   - 将复杂类拆分为多个职责单一的类

3. **Lambda/匿名函数圈复杂度不超过5**
   ```java
   // 反例：Lambda过于复杂
   list.stream()
       .filter(x -> x != null && x.length() > 0 && x.startsWith("A") || x.endsWith("Z"))
       .collect(Collectors.toList());

   // 正例：抽取方法
   list.stream()
       .filter(this::isValidName)
       .collect(Collectors.toList());
   ```

## 认知复杂度（Cognitive Complexity）

认知复杂度衡量代码对人类理解难度的影响，比圈复杂度更贴近实际阅读感受。

### 【Critical】规则

1. **单方法认知复杂度不超过15**
   - 嵌套层次乘以权重累加
   - 避免深层嵌套、break/continue、多重条件

   ```java
   // 反例：认知复杂度高
   public void process(List<Item> items) {
       for (Item item : items) {                        // +1 嵌套
           if (item != null) {                          // +2 嵌套+if
               if (item.isActive()) {                   // +3 嵌套+if
                   for (Tag tag : item.getTags()) {     // +4 嵌套
                       if (tag.isValid()) {             // +5 嵌套+if
                           processTag(tag);
                       }
                   }
               }
           }
       }
   }

   // 正例：早返回，减少嵌套
   public void process(List<Item> items) {
       for (Item item : items) {
           if (item == null || !item.isActive()) {
               continue;
           }
           processItemTags(item.getTags());
       }
   }
   ```

2. **使用卫语句（Guard Clauses）减少嵌套**
   ```java
   // 正例：卫语句模式
   public void process(Order order) {
       if (order == null) return;
       if (!order.isValid()) return;
       if (!order.hasPermission()) return;

       // 主要逻辑
       doProcess(order);
   }
   ```

## 方法规模

### 【Critical】规则

1. **单个方法不超过50行**
   - 不包括注释、空行、方法签名
   - 超过时应拆分为多个小方法

2. **方法参数不超过5个**
   ```java
   // 反例：参数过多
   void createUser(String name, String email, int age,
                   String address, String phone, String role) { }

   // 正例：使用参数对象
   class UserRequest {
       String name, email, address, phone, role;
       int age;
   }
   void createUser(UserRequest request) { }
   ```

3. **方法返回值类型应明确**
   ```java
   // 推荐：使用Optional表示可能为空的返回值
   public Optional<User> findById(Long id) {
       return Optional.ofNullable(repository.findById(id));
   }
   ```

## 类规模

### 【Major】规则

1. **单个类不超过500行**
   - 超过时考虑拆分为多个类

2. **类方法数量不超过30个**
   - 公开方法不超过15个

3. **类字段数量不超过20个**
   - 超过时考虑拆分或使用数据结构封装

4. **类嵌套层次不超过5层**
   - 避免过深的继承链

## 复杂度降低技巧

### 1. 提取方法（Extract Method）

```java
// 重构前
public String generateReport(User user) {
    StringBuilder sb = new StringBuilder();
    if (user != null) {
        if (user.getName() != null) {
            sb.append("Name: ").append(user.getName());
        }
        if (user.getEmail() != null) {
            sb.append("Email: ").append(user.getEmail());
        }
        // ... 更多字段
    }
    return sb.toString();
}

// 重构后
public String generateReport(User user) {
    if (user == null) return "";
    StringBuilder sb = new StringBuilder();
    appendField(sb, "Name", user.getName());
    appendField(sb, "Email", user.getEmail());
    return sb.toString();
}

private void appendField(StringBuilder sb, String label, String value) {
    if (value != null) {
        sb.append(label).append(": ").append(value).append("\n");
    }
}
```

### 2. 策略模式替代复杂条件

```java
// 重构前：复杂switch/if-else
public double calculate(String type, double amount) {
    switch (type) {
        case "VIP": return amount * 0.8;
        case "MEMBER": return amount * 0.9;
        case "GUEST": return amount * 0.95;
        default: return amount;
    }
}

// 重构后：策略模式
interface DiscountStrategy {
    double apply(double amount);
}

Map<String, DiscountStrategy> strategies = Map.of(
    "VIP", amount -> amount * 0.8,
    "MEMBER", amount -> amount * 0.9,
    "GUEST", amount -> amount * 0.95
);

public double calculate(String type, double amount) {
    return strategies.getOrDefault(type, a -> a).apply(amount);
}
```

### 3. 早返回（Early Return）

```java
// 重构前
public Result process(Input input) {
    Result result = new Result();
    if (input != null) {
        if (input.isValid()) {
            if (input.hasPermission()) {
                // 处理逻辑
                result.setSuccess(true);
            } else {
                result.setError("No permission");
            }
        } else {
            result.setError("Invalid input");
        }
    } else {
        result.setError("Input is null");
    }
    return result;
}

// 重构后
public Result process(Input input) {
    if (input == null) return Result.error("Input is null");
    if (!input.isValid()) return Result.error("Invalid input");
    if (!input.hasPermission()) return Result.error("No permission");

    // 处理逻辑
    return Result.success();
}
```

## SonarQube规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| 圈复杂度 | `S00101` (squid:S00101) | Critical |
| 认知复杂度 | `S00107` (squid:S00107) | Critical |
| 方法行数 | `S00101` | Major |
| 方法参数数量 | `S00107` | Major |
| 类行数 | `S00101` | Major |
| 类方法数量 | `S00101` | Major |
