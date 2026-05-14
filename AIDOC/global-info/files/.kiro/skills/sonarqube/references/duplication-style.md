# 代码重复与风格规范

## 代码重复（Duplicated Code）

代码重复是技术债务的主要来源之一，增加维护成本和Bug风险。

### 【Major】规则

1. **重复代码块不超过10行**

   ```java
   // 反例：重复代码
   public void processOrderA(Order order) {
       if (order == null) throw new IllegalArgumentException();
       if (order.getItems().isEmpty()) throw new IllegalArgumentException();
       // ... 更多相同逻辑
   }

   public void processOrderB(Order order) {
       if (order == null) throw new IllegalArgumentException();
       if (order.getItems().isEmpty()) throw new IllegalArgumentException();
       // ... 更多相同逻辑
   }

   // 正例：抽取公共方法
   private void validateOrder(Order order) {
       if (order == null) throw new IllegalArgumentException();
       if (order.getItems().isEmpty()) throw new IllegalArgumentException();
   }

   public void processOrderA(Order order) {
       validateOrder(order);
       // 特定逻辑A
   }

   public void processOrderB(Order order) {
       validateOrder(order);
       // 特定逻辑B
   }
   ```

2. **整体重复率不超过3%**

   - 使用SonarQube的重复检测功能
   - 定期审查重复代码报告
   - 优先处理高价值代码的重复

3. **跨文件重复处理**

   ```java
   // 反例：多个类中相同逻辑
   class UserService {
       public boolean isValidEmail(String email) {
           return email != null
               && email.matches("^[A-Za-z0-9+_.-]+@(.+)$");
       }
   }

   class AdminService {
       public boolean isValidEmail(String email) {
           return email != null
               && email.matches("^[A-Za-z0-9+_.-]+@(.+)$");
       }
   }

   // 正例：提取到工具类
   class ValidationUtils {
       public static boolean isValidEmail(String email) {
           return email != null
               && email.matches("^[A-Za-z0-9+_.-]+@(.+)$");
       }
   }

   class UserService {
       public boolean isValidEmail(String email) {
           return ValidationUtils.isValidEmail(email);
       }
   }
   ```

4. **相似代码用模板方法模式**

   ```java
   // 反例：多个相似方法
   public class PdfExporter {
       public void exportHeader() {
           System.out.println("=== Header ===");
           printMetadata();
           System.out.println("===============");
       }
       public void exportFooter() {
           System.out.println("=== Footer ===");
           printMetadata();
           System.out.println("===============");
       }
       private void printMetadata() {
           System.out.println("Date: " + new Date());
           System.out.println("Author: System");
       }
   }

   // 正例：模板方法
   public abstract class Exporter {
       public final void export(String section) {
           System.out.println("=== " + section + " ===");
           printMetadata();
           System.out.println("=================");
       }
       private void printMetadata() {
           System.out.println("Date: " + new Date());
           System.out.println("Author: System");
       }
   }
   ```

## 命名规范

### 【Minor】规则

1. **类名使用UpperCamelCase（PascalCase）**

   ```java
   // 正例
   public class UserService { }
   public class OrderItemController { }
   public class HttpClient { }

   // 反例
   public class userService { }
   public class order_ItemController { }
   public class HTTPClient { }  // 缩写超过2个字母不全大写
   ```

2. **方法名/变量名使用lowerCamelCase**

   ```java
   // 正例
   private String userName;
   public void getUserById() { }
   private int maxRetryCount;

   // 反例
   private String UserName;
   public void GetUserById() { }
   private int max_retry_count;
   ```

3. **常量全大写下划线分隔**

   ```java
   // 正例
   private static final int MAX_RETRY_COUNT = 3;
   private static final String DEFAULT_ENCODING = "UTF-8";

   // 反例
   private static final int maxRetry = 3;
   private static final String defaultEncoding = "UTF-8";
   ```

4. **包名全小写**

   ```java
   // 正例
   package com.example.service.user;
   package org.apache.commons.lang3;

   // 反例
   package com.example.Service.User;
   package com.example.service_user;
   ```

5. **接口/实现命名**

   ```java
   // 推荐：接口简单命名，实现类加Impl后缀
   public interface UserService { }
   public class UserServiceImpl implements UserService { }

   // 推荐：能力接口用-able结尾
   public interface Closeable { }
   public interface Serializable { }
   public interface Callable<V> { }
   ```

## 代码格式

### 【Minor】规则

1. **缩进：4空格**

   ```java
   // 正例：使用4空格缩进
   public class Example {
       public void method() {
           if (condition) {
               doSomething();
           }
       }
   }

   // 反例：使用tab或2空格
   public class Example {
     public void method() {
         if (condition) {
           doSomething();
         }
       }
   }
   ```

2. **行宽：不超过120字符**

   ```java
   // 反例：行过长
   List<User> users = userRepository.findAllByStatusAndCreatedDateAfterOrderByCreatedDateDesc(Status.ACTIVE, LocalDate.of(2024, 1, 1));

   // 正例：换行
   List<User> users = userRepository.findAllByStatusAndCreatedDateAfterOrderByCreatedDateDesc(
       Status.ACTIVE,
       LocalDate.of(2024, 1, 1)
   );
   ```

3. **大括号：左前不换行**

   ```java
   // 正例：K&R风格
   if (condition) {
       doSomething();
   } else {
       doOtherThing();
   }

   // 反例：换行风格
   if (condition)
   {
       doSomething();
   }
   else
   {
       doOtherThing();
   }
   ```

4. **空格使用**

   ```java
   // 运算符左右加空格
   int sum = a + b;
   boolean result = a > b && c < d;

   // 逗号后加空格
   method(a, b, c);

   // 保留字后加空格
   if (condition) { }
   while (running) { }
   for (int i = 0; i < n; i++) { }

   // 不要在括号内加空格
   if (a == b) { }  // 正例
   if ( a == b ) { }  // 反例
   ```

5. **空行使用**

   ```java
   // 类成员之间空一行
   public class Example {
       private static final Logger log = LoggerFactory.getLogger(Example.class);

       private String name;
       private int age;

       public Example(String name) {
           this.name = name;
       }

       public String getName() {
           return name;
       }
   }

   // 方法之间空一行
   public void method1() { }

   public void method2() { }
   ```

## 导入规范

### 【Minor】规则

1. **按顺序分组导入**

   ```java
   // 1. Java标准库
   import java.util.List;
   import java.util.Map;

   // 2. 第三方库
   import org.springframework.stereotype.Service;

   // 3. 本地包
   import com.example.model.User;
   import com.example.service.UserService;

   // 静态导入最后
   import static java.util.Collections.emptyList;
   ```

2. **删除未使用的导入**

   ```java
   // 反例：未使用的导入
   import java.util.List;  // 未使用
   import java.util.Set;
   import java.util.Map;   // 未使用

   // 正例：只导入使用的
   import java.util.Set;
   ```

3. **避免使用通配符导入**

   ```java
   // 反例：通配符导入
   import java.util.*;

   // 正例：显式导入
   import java.util.List;
   import java.util.Map;
   import java.util.Set;
   ```

## 注释规范

### 【Minor】规则

1. **Javadoc格式**

   ```java
   /**
    * 用户服务类，负责用户相关业务逻辑处理。
    *
    * <p>主要功能包括用户创建、更新、查询、删除等操作。</p>
    *
    * @author 张三
    * @since 1.0.0
    */
   public class UserService {

       /**
        * 根据用户ID查询用户信息。
        *
        * @param id 用户ID，不能为null
        * @return 用户信息，如果不存在返回null
        * @throws IllegalArgumentException 当id为null时抛出
        */
       public User findById(Long id) {
           // 实现
       }
   }
   ```

2. **行内注释**

   ```java
   // 单行注释
   int maxUsers = 100;  // 最大用户数限制

   /*
    * 多行注释
    * 用于复杂逻辑说明
    */
   ```

3. **TODO注释**

   ```java
   // TODO: 需要添加参数验证
   // FIXME: 修复并发问题
   // XXX: 临时方案，需要重构
   ```

## 集合与数组

### 【Major】规则

1. **返回空集合而非null**

   ```java
   // 反例：返回null
   public List<User> getUsers() {
       return users.isEmpty() ? null : users;
   }

   // 正例：返回空集合
   public List<User> getUsers() {
       return users.isEmpty() ? Collections.emptyList() : new ArrayList<>(users);
   }

   // 正例：使用Optional
   public Optional<List<User>> getUsers() {
       return users.isEmpty() ? Optional.empty() : Optional.of(new ArrayList<>(users));
   }
   ```

2. **集合初始化指定容量**

   ```java
   // 正例：指定初始容量
   List<User> users = new ArrayList<>(100);
   Map<String, User> userMap = new HashMap<>(16);

   // 反例：使用默认容量可能导致扩容
   List<User> users = new ArrayList<>();
   ```

3. **使用diamond运算符**

   ```java
   // 正例：使用diamond
   List<String> list = new ArrayList<>();
   Map<String, Integer> map = new HashMap<>();

   // 反例：重复类型参数
   List<String> list = new ArrayList<String>();
   ```

## SonarQube规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| 重复代码 | `S1192` (java:S1192) | Major |
| 未使用的导入 | `S1128` (java:S1128) | Minor |
| 通配符导入 | `S1214` (java:S1214) | Minor |
| 方法参数过多 | `S00107` | Major |
| 行过长 | `S103` (java:S103) | Minor |
| 类命名 | `S101` (java:S101) | Minor |
| 方法命名 | `S100` (java:S100) | Minor |
