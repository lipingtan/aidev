# Bug检测规范

## 空指针异常（NullPointerException）

NPE是最常见的运行时异常，必须通过代码规范避免。

### 【Blocker】规则

1. **解引用前检查null**

   ```java
   // 反例：可能NPE
   public void process(User user) {
       String name = user.getName();  // user可能为null
       System.out.println(name.toUpperCase());
   }

   // 正例1：显式检查
   public void process(User user) {
       if (user != null) {
           String name = user.getName();
           if (name != null) {
               System.out.println(name.toUpperCase());
           }
       }
   }

   // 正例2：使用Optional
   public void process(User user) {
       Optional.ofNullable(user)
           .map(User::getName)
           .map(String::toUpperCase)
           .ifPresent(System.out::println);
   }
   ```

2. **String常量equals比较**

   ```java
   // 反例：可能NPE
   if (user.getStatus().equals("ACTIVE")) { }

   // 正例：常量在前
   if ("ACTIVE".equals(user.getStatus())) { }

   // 正例：Objects.equals
   if (Objects.equals(user.getStatus(), "ACTIVE")) { }
   ```

3. **valueOf()代替toString()**

   ```java
   // 反例：可能NPE
   String value = object.toString();

   // 正例：valueOf自动处理null
   String value = String.valueOf(object);
   ```

4. **Collections.emptyList()而非null**

   ```java
   // 反例：返回null
   public List<Item> getItems() {
       return items.isEmpty() ? null : items;
   }

   // 正例：返回空集合
   public List<Item> getItems() {
       return items.isEmpty() ? Collections.emptyList() : items;
   }
   ```

## 资源泄漏

### 【Blocker】规则

1. **IO/DB/网络资源必须关闭**

   ```java
   // 反例：资源可能泄漏
   public String readFile(String path) throws IOException {
       BufferedReader reader = new BufferedReader(new FileReader(path));
       return reader.readLine();  // 未关闭reader
   }

   // 正例1：try-finally
   public String readFile(String path) throws IOException {
       BufferedReader reader = new BufferedReader(new FileReader(path));
       try {
           return reader.readLine();
       } finally {
           reader.close();
       }
   }

   // 正例2：try-with-resources（推荐）
   public String readFile(String path) throws IOException {
       try (BufferedReader reader = new BufferedReader(new FileReader(path))) {
           return reader.readLine();
       }
   }
   ```

2. **多个资源关闭顺序**

   ```java
   // 正例：按创建相反顺序关闭
   try (FileInputStream fis = new FileInputStream(file);
        BufferedInputStream bis = new BufferedInputStream(fis)) {
       // 使用资源
   }
   ```

3. **不要忽略关闭异常**

   ```java
   // 反例：忽略异常
   try (Resource r = getResource()) {
       // ...
   } catch (Exception e) {
       // 忽略
   }

   // 正例：至少记录日志
   try (Resource r = getResource()) {
       // ...
   } catch (Exception e) {
       log.error("Failed to close resource", e);
   }
   ```

## 并发问题

### 【Critical】规则

1. **共享可变状态必须同步**

   ```java
   // 反例：非线程安全的单例
   public class Singleton {
       private static Singleton instance;
       public static Singleton getInstance() {
           if (instance == null) {           // 竞态条件
               instance = new Singleton();
           }
           return instance;
       }
   }

   // 正例1：同步方法
   public static synchronized Singleton getInstance() {
       if (instance == null) {
           instance = new Singleton();
       }
       return instance;
   }

   // 正例2：双重检查锁定
   private static volatile Singleton instance;
   public static Singleton getInstance() {
       if (instance == null) {
           synchronized (Singleton.class) {
               if (instance == null) {
                   instance = new Singleton();
               }
           }
       }
       return instance;
   }

   // 正例3：枚举单例（推荐）
   public enum Singleton {
       INSTANCE;
   }
   ```

2. **不要在锁内调用外来方法**

   ```java
   // 反例：可能死锁
   private final Object lock = new Object();
   private List<String> items = new ArrayList<>();

   public void process() {
       synchronized (lock) {
           items.forEach(this::externalMethod);  // 外来方法可能获取其他锁
       }
   }

   // 正例：复制后处理
   public void process() {
       List<String> copy;
       synchronized (lock) {
           copy = new ArrayList<>(items);
       }
       copy.forEach(this::externalMethod);
   }
   ```

3. **使用线程安全集合**

   ```java
   // 反例：ArrayList非线程安全
   private List<String> shared = new ArrayList<>();

   // 正例：使用并发集合
   private List<String> shared = new CopyOnWriteArrayList<>();
   private Map<String, String> cache = new ConcurrentHashMap<>();
   ```

4. **避免嵌套锁，按固定顺序获取**

   ```java
   // 反例：可能死锁
   synchronized (lockA) {
       synchronized (lockB) {  // 如果其他线程相反顺序获取
           // ...
       }
   }

   // 正例：固定顺序
   // 按锁对象的hashCode排序后获取
   ```

## 逻辑错误

### 【Major】规则

1. **equals()与hashCode()契约**

   ```java
   // 反例：重写equals但未重写hashCode
   public class User {
       private int id;
       private String name;

       @Override
       public boolean equals(Object obj) {
           if (this == obj) return true;
           if (!(obj instanceof User)) return false;
           return this.id == ((User) obj).id;
       }
       // 缺少hashCode()！
   }

   // 正例：同时重写
   @Override
   public boolean equals(Object obj) {
       // ...
   }

   @Override
   public int hashCode() {
       return Objects.hash(id);
   }
   ```

2. **compareTo()与equals()一致性**

   ```java
   // 反例：不一致导致排序问题
   public class User implements Comparable<User> {
       private String name;

       @Override
       public int compareTo(User other) {
           return this.name.compareTo(other.name);
       }

       @Override
       public boolean equals(Object obj) {
           return this.id == ((User) obj).id;  // 不一致！
       }
   }

   // 正例：保持一致
   @Override
   public boolean equals(Object obj) {
       return compareTo((User) obj) == 0;
   }
   ```

3. **不要在循环中修改集合**

   ```java
   // 反例：ConcurrentModificationException
   List<String> items = new ArrayList<>();
   for (String item : items) {
       if (item.length() > 5) {
           items.remove(item);  // 异常！
       }
   }

   // 正例1：使用Iterator.remove()
   Iterator<String> it = items.iterator();
   while (it.hasNext()) {
       String item = it.next();
       if (item.length() > 5) {
           it.remove();
       }
   }

   // 正例2：removeIf
   items.removeIf(item -> item.length() > 5);
   ```

4. **条件判断顺序**

   ```java
   // 反例：可能NPE
   if (user.getName().length() > 0 && user != null) { }

   // 正例：短路逻辑，先判空
   if (user != null && user.getName() != null && user.getName().length() > 0) { }
   ```

5. **避免浮点数相等比较**

   ```java
   // 反例：精度问题
   if (a == 0.1 + 0.2) { }  // false

   // 正例：使用精度范围
   if (Math.abs(a - 0.3) < 0.0001) { }

   // 正例：使用BigDecimal
   BigDecimal result = new BigDecimal("0.1").add(new BigDecimal("0.2"));
   ```

## 异常处理

### 【Major】规则

1. **不要忽略异常**

   ```java
   // 反例：吞掉异常
   try {
       riskyOperation();
   } catch (Exception e) {
       // 空catch块
   }

   // 正例：至少记录日志
   try {
       riskyOperation();
   } catch (Exception e) {
       log.error("Operation failed", e);
   }
   ```

2. **不要捕获通用异常**

   ```java
   // 反例：捕获Throwable
   try {
       // ...
   } catch (Throwable t) {
       // 可能包括Error
   }

   // 正例：捕获具体异常
   try {
       // ...
   } catch (IOException e) {
       // 处理IO异常
   } catch (SQLException e) {
       // 处理SQL异常
   }
   ```

3. **不要捕获Exception后继续执行**

   ```java
   // 反例：异常后继续可能导致不一致状态
   try {
       updateState();
   } catch (Exception e) {
       log.error("Failed");
   }
   // 继续执行，但状态可能不一致

   // 正例：异常后终止或恢复
   try {
       updateState();
   } catch (Exception e) {
       log.error("Failed, rolling back");
       rollback();
       throw e;
   }
   ```

## SonarQube规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| 空指针解引用 | `S2259` (java:S2259) | Blocker |
| 资源未关闭 | `S2095` (java:S2095) | Blocker |
| 双重检查锁定 | `S2168` (java:S2168) | Critical |
| equals/hashCode契约 | `S1206` (java:S1206) | Major |
| 循环中修改集合 | `S2259` | Major |
| 空catch块 | `S1181` (java:S1181) | Major |
| 捕获Throwable | `S1181` | Critical |
