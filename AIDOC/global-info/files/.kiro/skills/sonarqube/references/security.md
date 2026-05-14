# 安全漏洞规范

## SQL注入（SQL Injection）

SQL注入是最严重的安全漏洞之一，可导致数据泄露、篡改或删除。

### 【Blocker】规则

1. **禁止字符串拼接SQL**

   ```java
   // 反例：SQL注入漏洞
   public User findByUsername(String username) {
       String sql = "SELECT * FROM users WHERE username = '" + username + "'";
       return jdbcTemplate.queryForObject(sql, User.class);
   }
   // 攻击：username = "' OR '1'='1"

   // 正例：参数化查询
   public User findByUsername(String username) {
       String sql = "SELECT * FROM users WHERE username = ?";
       return jdbcTemplate.queryForObject(sql, User.class, username);
   }

   // 正例：命名参数
   @Query("SELECT u FROM User u WHERE u.username = :username")
   User findByUsername(@Param("username") String username);
   ```

2. **LIKE查询转义**

   ```java
   // 反例：LIKE通配符注入
   String sql = "SELECT * FROM users WHERE name LIKE '%" + input + "%'";

   // 正例：使用ESCAPE或参数位置
   String sql = "SELECT * FROM users WHERE name LIKE ?";
   String pattern = "%" + escapeWildcards(input) + "%";
   ```

3. **IN子句参数化**

   ```java
   // 反例：字符串拼接
   String sql = "SELECT * FROM users WHERE id IN (" + idList + ")";

   // 正例：使用NamedParameterJdbcTemplate
   List<Long> ids = Arrays.asList(1L, 2L, 3L);
   MapSqlParameterSource params = new MapSqlParameterSource("ids", ids);
   String sql = "SELECT * FROM users WHERE id IN (:ids)";
   ```

## 命令注入（Command Injection）

### 【Blocker】规则

1. **禁止Runtime.exec()执行用户输入**

   ```java
   // 反例：命令注入
   String fileName = request.getParameter("file");
   Process process = Runtime.getRuntime().exec("cat " + fileName);
   // 攻击：file = "file.txt; rm -rf /"

   // 正例：使用ProcessBuilder（仍然需要验证）
   List<String> command = Arrays.asList("cat", validatedFileName);
   new ProcessBuilder(command).start();
   ```

2. **验证文件名白名单**

   ```java
   private static final Pattern SAFE_FILENAME = Pattern.compile("[a-zA-Z0-9._-]+");

   public void readFile(String filename) {
       if (!SAFE_FILENAME.matcher(filename).matches()) {
           throw new SecurityException("Invalid filename");
       }
       // 处理文件
   }
   ```

## XSS攻击（Cross-Site Scripting）

### 【Critical】规则

1. **输出到HTML必须转义**

   ```java
   // 反例：XSS漏洞
   <div><%= request.getParameter("comment") %></div>
   // 攻击：comment = "<script>alert(document.cookie)</script>"

   // 正例：JSTL转义
   <c:out value="${param.comment}"/>

   // 正例：HTML转义
   String escaped = HtmlUtils.htmlEscape(userInput);
   ```

2. **输出到JavaScript必须转义**

   ```java
   // 反例：XSS漏洞
   <script>
       var name = "<%= request.getParameter("name") %>";
   </script>

   // 正例：JavaScript转义
   <script>
       var name = <%= JsonUtils.toJsonString(request.getParameter("name")) %>;
   </script>
   ```

3. **设置Content-Type防止反射型XSS**

   ```java
   @GetMapping("/api/data")
   public ResponseEntity getData() {
       return ResponseEntity.ok()
           .contentType(MediaType.APPLICATION_JSON)
           .body(data);
   }
   ```

4. **使用CSP头**

   ```java
   @Configuration
   public class SecurityConfig {
       @Bean
       public SecurityFilterChain filterChain(HttpSecurity http) {
           http.headers()
               .contentSecurityPolicy("default-src 'self'; script-src 'self' 'unsafe-inline'");
           return http.build();
       }
   }
   ```

## 路径遍历（Path Traversal）

### 【Critical】规则

1. **验证文件路径**

   ```java
   // 反例：路径遍历
   String filename = request.getParameter("file");
   File file = new File("/var/files/" + filename);
   // 攻击：file = "../../etc/passwd"

   // 正例：规范化并验证
   String filename = request.getParameter("file");
   Path path = Paths.get("/var/files", filename).normalize();
   if (!path.startsWith("/var/files/")) {
       throw new SecurityException("Invalid path");
   }
   ```

2. **使用文件名白名单**

   ```java
   private static final Set<String> ALLOWED_FILES = Set.of("report.pdf", "summary.xlsx");

   public File getFile(String name) {
       if (!ALLOWED_FILES.contains(name)) {
           throw new SecurityException("File not allowed");
       }
       return new File(BASE_DIR, name);
   }
   ```

## 硬编码敏感信息

### 【Blocker】规则

1. **禁止硬编码密码**

   ```java
   // 反例：硬编码密码
   private static final String DB_PASSWORD = "admin123";
   private static final String API_KEY = "sk-1234567890";

   // 正例：使用环境变量
   private static final String DB_PASSWORD = System.getenv("DB_PASSWORD");

   // 正例：使用配置中心
   @Value("${database.password}")
   private String dbPassword;
   ```

2. **禁止硬编码密钥/盐值**

   ```java
   // 反例：硬编码密钥
   private static final byte[] KEY = "secretkey123".getBytes();

   // 正例：从安全存储获取
   @Bean
   public Cipher cipher() {
       String key = keyVault.getSecret("encryption-key");
       return Cipher.getInstance("AES/GCM/NoPadding");
   }
   ```

3. **禁止硬编码IP地址**

   ```java
   // 反例：硬编码IP
   private static final String DB_URL = "jdbc:mysql://192.168.1.100:3306/db";

   // 正例：配置文件
   private static final String DB_URL = System.getenv("DB_URL");
   ```

## 加密与哈希

### 【Critical】规则

1. **使用强加密算法**

   ```java
   // 反例：弱算法
   MessageDigest md = MessageDigest.getInstance("MD5");    // 已破解
   Cipher cipher = Cipher.getInstance("DES");              // 密钥太短
   Cipher cipher = Cipher.getInstance("RC4");              // 不安全

   // 正例：强算法
   MessageDigest md = MessageDigest.getInstance("SHA-256");
   Cipher cipher = Cipher.getInstance("AES/GCM/NoPadding");
   ```

2. **密码使用bcrypt/argon2**

   ```java
   // 反例：SHA/MD5可被彩虹表攻击
   String hash = DigestUtils.sha256Hex(password);

   // 正例：使用bcrypt
   String hash = BCrypt.hashpw(password, BCrypt.gensalt());

   // 正例：验证
   if (BCrypt.checkpw(password, hash)) {
       // 密码正确
   }
   ```

3. **使用随机IV**

   ```java
   // 反例：固定IV
   byte[] iv = new byte[16];

   // 正例：随机IV
   byte[] iv = new byte[16];
   SecureRandom random = new SecureRandom();
   random.nextBytes(iv);
   ```

4. **使用HTTPS，不HTTP**

   ```java
   // 反例：HTTP明文传输
   String url = "http://api.example.com/user";

   // 正例：HTTPS加密传输
   String url = "https://api.example.com/user";
   ```

## 认证与授权

### 【Critical】规则

1. **防止暴力破解**

   ```java
   @Configuration
   public class SecurityConfig {
       @Bean
       public SecurityFilterChain filterChain(HttpSecurity http) {
           http.authorizeRequests()
               .antMatchers("/admin/**").hasRole("ADMIN")
               .anyRequest().authenticated()
               .and()
               .formLogin()
               .and()
               .sessionManagement()
               .maximumSessions(1)
               .maxSessionsPreventsLogin(true);
           return http.build();
       }
   }
   ```

2. **CSRF防护**

   ```java
   // Spring Security默认启用CSRF防护
   // 自定义请求需要携带CSRF token
   ```

3. **方法级安全**

   ```java
   @PreAuthorize("hasRole('ADMIN')")
   public void deleteUser(Long userId) {
       // 只有ADMIN角色可执行
   }

   @PreAuthorize("#userId == authentication.principal.id")
   public User getUserProfile(Long userId) {
       // 只能访问自己的资料
   }
   ```

4. **敏感操作需二次验证**

   ```java
   @PostMapping("/transfer")
   @RequiresSecondFactor
   public Response transfer(@RequestBody TransferRequest request) {
       // 大额转账需要二次验证
   }
   ```

## 敏感数据处理

### 【Major】规则

1. **日志脱敏**

   ```java
   // 反例：记录敏感信息
   log.info("User login: username={}, password={}",
       user.getUsername(), user.getPassword());

   // 正例：脱敏处理
   log.info("User login: username={}, password=***",
       user.getUsername());

   // 正例：使用脱敏工具
   log.info("Request: {}", DesensitizedUtil.desensitize(request));
   ```

2. **响应中脱敏**

   ```java
   @JsonIgnore
   private String password;

   @JsonProperty(access = JsonProperty.Access.WRITE_ONLY)
   private String creditCard;
   ```

3. **不在URL中传递敏感信息**

   ```java
   // 反例：敏感信息在URL中
   @GetMapping("/user/{token}/verify")
   public Response verify(@PathVariable String token) { }

   // 正例：使用POST
   @PostMapping("/user/verify")
   public Response verify(@RequestBody VerificationRequest request) { }
   ```

## SonarQube安全规则键

| 规则 | 键 | 严重级别 |
|------|-----|----------|
| SQL注入 | `S2077` (java:S2077) | Blocker |
| 命令注入 | `S2083` (java:S2083) | Blocker |
| 硬编码密码 | `S2068` (java:S2068) | Blocker |
| 路径遍历 | `S2083` | Critical |
| 弱加密算法 | `S5542` (java:S5542) | Critical |
| XSS漏洞 | `S5131` (java:S5131) | Critical |
| 日志中敏感信息 | `S2068` | Major |
| HTTP而非HTTPS | `S5332` (java:S5332) | Major |
| 弱随机数 | `S2245` (java:S2245) | Critical |
