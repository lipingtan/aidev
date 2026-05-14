# 代码异味规范

代码异味（Code Smells）是代码中可能存在问题但不直接导致Bug的迹象，长期积累会影响可维护性。

## 冗余代码

### 【Major】规则

1. **删除未使用的代码**

   ```java
   // 删除未使用的导入
   import java.util.List;  // 未使用
   import java.util.Map;

   // 删除未使用的私有方法
   private void unusedMethod() { }

   // 删除未使用的字段
   private int unusedField;

   // 删除未使用的参数
   public void process(String data, int unused) { }
   ```

2. **删除空方法**

   ```java
   // 反例：空实现
   @Override
   public void close() {
       // 空方法
   }

   // 正例：如果不需要，不实现或不声明
   ```

3. **合并条件**

   ```java
   // 反例：重复条件
   if (x != null) {
       if (x.isValid()) {
           process(x);
       }
   }

   // 正例：合并
   if (x != null && x.isValid()) {
       process(x);
   }
   ```

## 设计问题

### 【Major】规则

1. **避免上帝类（God Class）**

   ```java
   // 反例：类承担过多职责
   class UserManager {
       public void createUser() { }
       public void sendEmail() { }
       public void logActivity() { }
       public void validateInput() { }
       public void generateReport() { }
       // ... 太多方法
   }

   // 正例：职责分离
   class UserService { public void createUser() { } }
   class EmailService { public void sendEmail() { } }
   class ActivityLogger { public void logActivity() { } }
   ```

2. **避免上帝方法（God Method）**

   ```java
   // 反例：方法做太多事情
   public void processOrder(Order order) {
       // 验证订单
       if (order == null) throw new IllegalArgumentException();
       if (order.getItems().isEmpty()) throw new IllegalArgumentException();
       // 计算金额
       double total = 0;
       for (Item item : order.getItems()) {
           total += item.getPrice() * item.getQuantity();
       }
       // 检查库存
       for (Item item : order.getItems()) {
           if (!inventory.check(item.getId(), item.getQuantity())) {
               throw new OutOfStockException();
           }
       }
       // 扣减库存
       for (Item item : order.getItems()) {
           inventory.reduce(item.getId(), item.getQuantity());
       }
       // 保存订单
       orderRepository.save(order);
       // 发送邮件
       emailService.send(order.getEmail(), "订单确认");
       // 记录日志
       log.info("Order processed: {}", order.getId());
   }

   // 正例：拆分为小方法
   public void processOrder(Order order) {
       validateOrder(order);
       double total = calculateTotal(order);
       checkInventory(order);
       reduceInventory(order);
       saveOrder(order);
       sendConfirmation(order);
       logOrder(order);
   }
   ```

3. **避免特性依恋（Feature Envy）**

   ```java
   // 反例：过度使用其他对象
   public void report() {
       System.out.println(user.getName());
       System.out.println(user.getEmail());
       System.out.println(user.getAddress().getStreet());
       System.out.println(user.getAddress().getCity());
   }

   // 正例：将方法移到被使用的对象
   // 在User类中
   public void report() {
       System.out.println(getName());
       System.out.println(getEmail());
       address.report();
   }
   ```

## 可维护性问题

### 【Major】规则

1. **避免魔法数字**

   ```java
   // 反例：魔法数字
   if (status == 1) {
       return "Active";
   } else if (status == 2) {
       return "Pending";
   }

   // 正例：使用常量
   private static final int STATUS_ACTIVE = 1;
   private static final int STATUS_PENDING = 2;

   // 更好：使用枚举
   public enum Status { ACTIVE, PENDING }
   ```

2. **避免过长参数列表**

   ```java
   // 反例：参数太多
   public void createReport(String title, String author, Date date,
                           String format, boolean includeCharts,
                           boolean includeTables, String template) { }

   // 正例：使用参数对象
   public void createReport(ReportRequest request) { }

   class ReportRequest {
       String title, author, format, template;
       Date date;
       boolean includeCharts, includeTables;
   }
   ```

3. **避免链式调用过长**

   ```java
   // 反例：链式调用过长，难以调试
   user.getOrder().getItems().get(0).getProduct().getCategory().getName();

   // 正例：拆分并检查null
   Category category = user.getFirstItemCategory();
   if (category != null) {
       return category.getName();
   }
   return "Unknown";
   ```

4. **避免过度使用三元运算符**

   ```java
   // 反例：嵌套三元难以阅读
   String result = a > b ? (a > c ? "A" : "C") : (b > c ? "B" : "C");

   // 正例：使用if-else
   String result;
   if (a > b && a > c) {
       result = "A";
   } else if (b > c) {
       result = "B";
   } else {
       result = "C";
   }
   ```

## 代码组织

### 【Minor】规则

1. **避免深层次嵌套**

   ```java
   // 反例：嵌套过深
   public void process(Data data) {
       if (data != null) {
           if (data.isValid()) {
               if (data.hasPermission()) {
                   if (data.isReady()) {
                       execute(data);
                   }
               }
           }
       }
   }

   // 正例：早返回
   public void process(Data data) {
       if (data == null) return;
       if (!data.isValid()) return;
       if (!data.hasPermission()) return;
       if (!data.isReady()) return;
       execute(data);
   }
   ```

2. **相关代码放在一起**

   ```java
   // 推荐：成员变量分组
   public class UserService {
       // 常量
       private static final int MAX_RETRY = 3;

       // 静态变量
       private static final Logger log = LoggerFactory.getLogger(UserService.class);

       // 实例变量
       private final UserRepository userRepository;

       // 构造方法
       public UserService(UserRepository userRepository) {
           this.userRepository = userRepository;
       }

       // 公开方法
       public User create() { }

       // 私有方法
       private void validate() { }
   }
   ```

3. **方法职责单一**

   ```java
   // 反例：一个方法做多件事
   public void saveUser(User user) {
       validateUser(user);
       sanitizeUser(user);
       encryptPassword(user);
       userRepository.save(user);
       sendWelcomeEmail(user);
       logUserCreation(user);
   }

   // 正例：每个方法只做一件事
   public void saveUser(User user) {
       validateUser(user);
       userRepository.save(user);
   }
   ```

## 命名与可读性

### 【Minor】规则

1. **使用有意义的命名**

   ```java
   // 反例：无意义命名
   int d = 5;
   List list = new ArrayList();

   // 正例：语义化命名
   int maxRetryAttempts = 5;
   List<User> activeUsers = new ArrayList<>();
   ```

2. **布尔值命名避免否定**

   ```java
   // 反例：双重否定
   if (!user.isNotInactive()) { }

   // 正例：正面命名
   if (user.isActive()) { }
   ```

3. **避免误导性命名**

   ```java
   // 反例：命名不准确
   List<User> getUserList() {
       return Set.of(user1, user2);  // 返回Set而非List
   }

   // 正例：命名准确
   Collection<User> getUsers() {
       return Set.of(user1, user2);
   }
   ```

## 注释

### 【Minor】规则

1. **注释解释"为什么"而非"是什么"**

   ```java
   // 反例：注释重复代码
   // 设置用户名为John
   user.setName("John");

   // 正例：注释解释原因
   // 使用UTC时间避免时区问题
   ZonedDateTime utcTime = ZonedDateTime.now(ZoneOffset.UTC);
   ```

2. **删除过时注释**

   ```java
   // 反例：注释与代码不一致
   // TODO: 需要处理null情况
   public void process(String data) {
       String result = data.toUpperCase();  // 实际已处理null
   }
   ```

3. **优先用清晰代码代替注释**

   ```java
   // 反例：需要注释解释
   // 检查用户是VIP且已激活
   if (user.getType() == 1 && user.getStatus() == 2) { }

   // 正例：代码自解释
   if (user.isVip() && user.isActivated()) { }
   ```

## SonarQube代码异味规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| 未使用的导入 | `S1128` (java:S1128) | Minor |
| 未使用的私有方法 | `S1144` (java:S1144) | Major |
| 未使用的参数 | `S1172` (java:S1172) | Minor |
| 空方法 | `S1186` (java:S1186) | Minor |
| 魔法数字 | `S109` (java:S109) | Minor |
| 参数过多 | `S00107` | Major |
| 方法过长 | `S138` (java:S138) | Major |
| 类过长 | `S1148` | Major |
| 嵌套过深 | `S134` (java:S134) | Major |
| 上帝类 | `S1200` (java:S1200) | Major |
| 复制粘贴检测 | `S1118` (java:S1118) | Major |
