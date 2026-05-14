# 异常处理与安全规范

## 异常处理规范

### 【强制】可预检查的RuntimeException

Java类库可通过预检查规避的RuntimeException不应该用catch处理

### 错误示例
```java
// 错误: NPE不应该catch处理
try {
    String name = user.getName().toString();
} catch (NullPointerException e) {
    logger.error("NPE异常", e);
}
```

### 正确示例
```java
// 正确: 预检查
if (user != null && user.getName() != null) {
    String name = user.getName().toString();
}

// 正确: 使用Objects.requireNonNull
Objects.requireNonNull(user, "user不能为null");
Objects.requireNonNull(user.getName(), "name不能为null");
```

### 【强制】异常不用作流程控制

异常处理效率低,避免将异常用于条件控制

### 错误示例
```java
// 错误: 使用异常做流程控制
try {
    user = userDao.selectById(userId);
} catch (EmptyResultDataAccessException e) {
    user = new UserDO();  // 不存在就创建新用户
}
```

### 正确示例
```java
// 正确: 先查询判断
user = userDao.selectById(userId);
if (user == null) {
    user = new UserDO();
}
```

### 【强制】异常类型区分

catch时分清稳定代码和非稳定代码,尽可能区分异常类型

### 错误示例
```java
// 错误: 大段代码统一catch
try {
    // 大量业务代码
    userDao.insert(user);
    orderDao.insert(order);
    sendSms(user.getPhone());
    sendEmail(user.getEmail());
} catch (Exception e) {
    logger.error("操作失败", e);  // 无法区分具体异常类型
}
```

### 正确示例
```java
// 正确: 分门别类处理
try {
    userDao.insert(user);
} catch (DuplicateKeyException e) {
    throw new UserException("用户已存在:" + user.getName());
}

try {
    orderDao.insert(order);
} catch (DataAccessException e) {
    throw new OrderException("订单创建失败", e);
}

try {
    sendSms(user.getPhone());
    sendEmail(user.getEmail());
} catch (SmsException | EmailException e) {
    throw new NotificationException("通知发送失败", e);
}
```

### 【强制】异常必须处理或抛出

捕获异常不能什么都不处理就抛弃

### 错误示例
```java
// 错误: 捕获但不处理
try {
    orderService.createOrder(orderDTO);
} catch (Exception e) {
    // 空catch,异常丢失
}
```

### 正确示例
```java
// 正确: 处理或抛出
try {
    orderService.createOrder(orderDTO);
} catch (ServiceException e) {
    // 转化为用户可理解的内容
    return "订单创建失败:" + e.getMessage();
}

// 或抛出给调用者
try {
    orderService.createOrder(orderDTO);
} catch (ServiceException e) {
    throw new WebException("订单创建失败", e);
}
```

### 【强制】事务回滚

try块在事务代码中,catch异常后如需回滚必须手动回滚

### 正确示例
```java
@Transactional(rollbackFor = Exception.class)
public void processOrder(OrderDTO orderDTO) {
    try {
        // 业务逻辑
        orderDao.insert(order);
        stockDao.updateStock(order.getProductId(), -order.getQuantity());
    } catch (StockInsufficientException e) {
        logger.error("库存不足", e);
        // 手动回滚事务
        TransactionAspectSupport.currentTransactionStatus().setRollbackOnly();
        throw new OrderException("库存不足", e);
    }
}
```

### 【强制】资源关闭

finally块必须对资源对象/流对象进行关闭,异常也要try-catch

### JDK7之前正确示例
```java
FileInputStream fis = null;
try {
    fis = new FileInputStream(file);
    // 文件操作
} catch (IOException e) {
    logger.error("文件读取异常", e);
} finally {
    if (fis != null) {
        try {
            fis.close();  // 关闭也可能抛异常,所以要try-catch
        } catch (IOException e) {
            logger.error("关闭文件流异常", e);
        }
    }
}
```

### JDK7+正确示例 (推荐)
```java
// 使用try-with-resources自动关闭
try (FileInputStream fis = new FileInputStream(file);
     BufferedInputStream bis = new BufferedInputStream(fis)) {
    // 文件操作
} catch (IOException e) {
    logger.error("文件读取异常", e);
}
// 自动关闭所有资源,无需手动close
```

### 【强制】禁止finally中return

finally块中return会覆盖try块中return

### 错误示例
```java
public int getValue() {
    try {
        return 10;  // 不会执行
    } finally {
        return 20;  // 永远返回20
    }
}
```

### 正确示例
```java
public int getValue() {
    int result;
    try {
        result = calculate();
    } finally {
        cleanUp();  // 清理资源
    }
    return result;
}
```

### 【推荐】防止NPE

NPE产生的场景及防护

#### 场景1: 返回基本类型return包装对象
```java
// 错误: 自动拆箱可能NPE
public int getAge() {
    return ageObject;  // ageObject可能null
}

// 正确: 提前检查
public int getAge() {
    return ageObject != null ? ageObject : 0;
}
```

#### 场景2: 数据库查询结果可能为null
```java
// 错误
UserDO user = userDao.selectById(userId);
String name = user.getName();  // user可能null

// 正确
UserDO user = userDao.selectById(userId);
if (user != null) {
    String name = user.getName();
}
```

#### 场景3: 集合元素为null
```java
List<String> names = getNames();  // names不为null但元素可能null
// 错误
for (String name : names) {
    int length = name.length();  // name可能null
}

// 正确
for (String name : names) {
    if (name != null) {
        int length = name.length();
    }
}

// 或使用Optional (JDK8+)
Optional.ofNullable(name)
    .map(String::length)
    .orElse(0);
```

#### 场景4: 远程调用返回对象
```java
// 错误
UserDTO user = userService.getUser(userId);
String email = user.getEmail();  // 远程调用可能失败返回null

// 正确
UserDTO user = userService.getUser(userId);
if (user != null) {
    String email = user.getEmail();
}

// 或使用Optional
Optional.ofNullable(userService.getUser(userId))
    .map(UserDTO::getEmail)
    .ifPresent(email -> sendEmail(email));
```

#### 场景5: Session获取数据
```java
// 错误
UserDO user = (UserDO) session.getAttribute("user");
String name = user.getName();  // user可能null

// 正确
UserDO user = (UserDO) session.getAttribute("user");
if (user != null) {
    String name = user.getName();
}
```

#### 场景6: 级联调用
```java
// 错误: 一连串调用易NPE
String city = user.getAddress().getCity();

// 正确: 逐步判断
String city = null;
if (user != null) {
    Address address = user.getAddress();
    if (address != null) {
        city = address.getCity();
    }
}

// 或使用Optional链式调用
String city = Optional.ofNullable(user)
    .map(User::getAddress)
    .map(Address::getCity)
    .orElse(null);
```

### 【推荐】自定义异常

定义区分unchecked/checked异常,避免直接抛RuntimeException

#### 正确示例
```java
// 自定义业务异常
public class UserException extends RuntimeException {
    private String errorCode;

    public UserException(String message) {
        super(message);
    }

    public UserException(String errorCode, String message) {
        super(message);
        this.errorCode = errorCode;
    }

    public String getErrorCode() {
        return errorCode;
    }
}

// DAO异常
public class DAOException extends RuntimeException {
    public DAOException(String message, Throwable cause) {
        super(message, cause);
    }
}

// Service异常
public class ServiceException extends RuntimeException {
    public ServiceException(String message) {
        super(message);
    }
}

// 使用
try {
    userDao.insert(user);
} catch (DuplicateKeyException e) {
    throw new UserException("USER_EXISTS", "用户已存在:" + user.getName());
}
```

## 安全规范

### 【强制】水平权限控制

用户个人页面/功能必须进行权限控制校验

#### 错误示例
```java
// 错误: 没有水平权限校验
@GetMapping("/order/{orderId}")
public OrderVO getOrder(@PathVariable Long orderId) {
    // 直接返回订单,用户可能访问他人订单
    OrderDO order = orderDao.selectById(orderId);
    return convertToVO(order);
}
```

#### 正确示例
```java
// 正确: 校验订单归属
@GetMapping("/order/{orderId}")
public OrderVO getOrder(@PathVariable Long orderId) {
    Long userId = getCurrentUserId();
    OrderDO order = orderDao.selectById(orderId);

    if (order == null) {
        throw new OrderException("订单不存在");
    }

    // 水平权限校验: 只能查看自己的订单
    if (!order.getUserId().equals(userId)) {
        throw new PermissionException("无权限访问此订单");
    }

    return convertToVO(order);
}
```

### 【强制】敏感数据脱敏

用户敏感数据禁止直接展示,必须脱敏

#### 正确示例
```java
public class DesensitizedUtil {

    // 手机号脱敏: 保留前3后4位
    public static String desensitizePhone(String phone) {
        if (StringUtils.isBlank(phone) || phone.length() < 7) {
            return phone;
        }
        return phone.substring(0, 3) + "****" + phone.substring(phone.length() - 4);
    }

    // 身份证脱敏: 保留前3后4位
    public static String desensitizeIdCard(String idCard) {
        if (StringUtils.isBlank(idCard) || idCard.length() < 8) {
            return idCard;
        }
        return idCard.substring(0, 3) + "***********" + idCard.substring(idCard.length() - 4);
    }

    // 邮箱脱敏: @前2位, @后3位
    public static String desensitizeEmail(String email) {
        if (StringUtils.isBlank(email) || !email.contains("@")) {
            return email;
        }
        int atIndex = email.indexOf("@");
        String prefix = email.substring(0, Math.min(2, atIndex));
        String suffix = email.substring(atIndex + 1);
        suffix = suffix.substring(0, Math.min(3, suffix.length())) + "***";
        return prefix + "***@" + suffix;
    }
}

// 使用
UserVO userVO = new UserVO();
userVO.setPhone(DesensitizedUtil.desensitizePhone(userDO.getPhone()));
userVO.setIdCard(DesensitizedUtil.desensitizeIdCard(userDO.getIdCard()));
userVO.setEmail(DesensitizedUtil.desensitizeEmail(userDO.getEmail()));
```

### 【强制】SQL注入防护

用户输入SQL参数严格使用参数绑定

#### 错误示例
```java
// 错误: 字符串拼接SQL,存在注入风险
public List<UserDO> searchUsers(String keyword) {
    String sql = "SELECT * FROM user WHERE name LIKE '%" + keyword + "%'";
    return jdbcTemplate.query(sql, userRowMapper);
}
```

#### 正确示例
```java
// 正确: 使用参数绑定
public List<UserDO> searchUsers(String keyword) {
    String sql = "SELECT * FROM user WHERE name LIKE ?";
    return jdbcTemplate.query(sql, userRowMapper, "%" + keyword + "%");
}

// 或使用命名参数
public List<UserDO> searchUsers(String keyword) {
    String sql = "SELECT * FROM user WHERE name LIKE :keyword";
    Map<String, Object> params = new HashMap<>();
    params.put("keyword", "%" + keyword + "%");
    return namedParameterJdbcTemplate.query(sql, params, userRowMapper);
}

// MyBatis正确使用
@Select("SELECT * FROM user WHERE name LIKE CONCAT('%', #{keyword}, '%')")
List<UserDO> searchUsers(@Param("keyword") String keyword);
```

### 【强制】参数有效性验证

用户请求传入任何参数必须做有效性验证

#### 正确示例
```java
@GetMapping("/orders")
public PageResult<OrderVO> listOrders(OrderQuery query) {
    // page size过大导致内存溢出
    if (query.getPageSize() != null && query.getPageSize() > 1000) {
        throw new IllegalArgumentException("分页大小不能超过1000");
    }

    // 恶意order by导致数据库慢查询
    String orderBy = query.getOrderBy();
    if (orderBy != null && !isValidOrderBy(orderBy)) {
        throw new IllegalArgumentException("非法排序字段:" + orderBy);
    }

    // 任意重定向防护
    String redirectUrl = query.getRedirectUrl();
    if (redirectUrl != null && !isValidRedirectUrl(redirectUrl)) {
        throw new SecurityException("非法重定向URL");
    }

    return orderService.listOrders(query);
}

// 验证排序字段白名单
private boolean isValidOrderBy(String orderBy) {
    Set<String> allowedFields = new HashSet<>(Arrays.asList(
        "id", "createTime", "updateTime", "amount"
    ));
    return allowedFields.contains(orderBy);
}

// 验证重定向URL白名单
private boolean isValidRedirectUrl(String url) {
    try {
        URL validUrl = new URL(url);
        String host = validUrl.getHost();
        Set<String> allowedDomains = new HashSet<>(Arrays.asList(
            "example.com", "www.example.com"
        ));
        return allowedDomains.contains(host);
    } catch (Exception e) {
        return false;
    }
}
```

### 【强制】CSRF防护

表单/AJAX提交必须执行CSRF安全验证

#### 正确示例
```java
// Spring Security配置
@Configuration
@EnableWebSecurity
public class SecurityConfig extends WebSecurityConfigurerAdapter {

    @Override
    protected void configure(HttpSecurity http) throws Exception {
        http
            .csrf()  // 启用CSRF防护
            .csrfTokenRepository(CookieCsrfTokenRepository.withHttpOnlyFalse())
            .and()
            .authorizeRequests()
            // ...
    }
}

// Controller中添加CSRF Token
@PostMapping("/order/create")
public Result createOrder(@RequestBody OrderDTO orderDTO,
                        @CsrfToken String csrfToken) {
    // CSRF Token自动验证
    return orderService.createOrder(orderDTO);
}
```

### 【强制】防重放机制

使用平台资源必须实现正确防重放机制

#### 正确示例: 验证码防重放
```java
@Service
public class VerificationCodeService {

    @Autowired
    private RedisTemplate<String, String> redisTemplate;

    // 发送验证码: 数量限制 + 疲劳度控制
    public void sendSmsCode(String phone) {
        // 1. 每天限制次数
        String dailyKey = "sms:count:" + phone + ":" + LocalDate.now();
        Long dailyCount = redisTemplate.opsForValue().increment(dailyKey);
        if (dailyCount == 1) {
            redisTemplate.expire(dailyKey, 1, TimeUnit.DAYS);
        }
        if (dailyCount > 10) {
            throw new TooManyRequestsException("验证码发送次数过多");
        }

        // 2. 疲劳度控制: 60秒内不能重复发送
        String cooldownKey = "sms:cooldown:" + phone;
        if (redisTemplate.hasKey(cooldownKey)) {
            throw new TooManyRequestsException("验证码发送过于频繁,请60秒后再试");
        }
        redisTemplate.opsForValue().set(cooldownKey, "1", 60, TimeUnit.SECONDS);

        // 3. 生成6位随机验证码
        String code = String.format("%06d", new Random().nextInt(1000000));

        // 4. 存储验证码,5分钟有效
        String codeKey = "sms:code:" + phone;
        redisTemplate.opsForValue().set(codeKey, code, 5, TimeUnit.MINUTES);

        // 5. 发送短信
        smsService.send(phone, "您的验证码是:" + code);
    }

    // 校验验证码: 一次性使用
    public boolean verifyCode(String phone, String code) {
        String codeKey = "sms:code:" + phone;
        String storedCode = redisTemplate.opsForValue().get(codeKey);

        if (storedCode == null) {
            return false;  // 验证码过期或不存在
        }

        if (!storedCode.equals(code)) {
            return false;  // 验证码错误
        }

        // 验证成功后删除,防重放
        redisTemplate.delete(codeKey);
        return true;
    }
}
```

### 【推荐】内容风控

发贴/评论/即时消息等用户生成内容场景必须实现防刷/违禁词过滤

#### 正确示例
```java
@Service
public class ContentFilterService {

    @Autowired
    private AntiSpamService antiSpamService;

    @Autowired
    private WordFilterService wordFilterService;

    @Autowired
    private FrequencyLimitService frequencyLimitService;

    public void checkContent(String userId, String content) {
        // 1. 防刷: 同一用户发布频率限制
        if (!frequencyLimitService.checkAndRecord(userId, "publish", 60, 10)) {
            throw new TooManyRequestsException("发布过于频繁,请稍后再试");
        }

        // 2. 违禁词过滤
        List<String> bannedWords = wordFilterService.check(content);
        if (!bannedWords.isEmpty()) {
            throw new ContentException("内容包含违禁词:" + String.join(",", bannedWords));
        }

        // 3. 内容风控: 检测垃圾广告/恶意链接等
        RiskLevel riskLevel = antiSpamService.check(content);
        if (riskLevel == RiskLevel.HIGH) {
            throw new ContentException("内容违规,已被拦截");
        }
    }
}
```

### 【推荐】XSS防护

禁止向HTML页面输出未经安全过滤或未正确转义的用户数据

#### 正确示例
```java
// 使用Spring的HTML转义
@GetMapping("/comment/{id}")
public String viewComment(@PathVariable Long id, Model model) {
    CommentDO comment = commentDao.selectById(id);
    model.addAttribute("comment", comment);

    // 使用@HtmlEscape自动转义
    model.addAttribute("content", HtmlUtils.htmlEscape(comment.getContent()));
    return "comment/detail";
}

// 或使用模板引擎自动转义
// Thymeleaf默认自动转义
<div th:text="${comment.content}"></div>

// 如需输出原始HTML(需谨慎)
<div th:utext="${comment.content}"></div>
```