---
name: java-standard
description: Java后端开发规范编码标准,基于GDC开发规范V1.0.用于Java后端开发时的代码审查、规范检查、代码生成等场景。包含命名规范、OOP规约、集合处理、并发处理、异常处理、单元测试、安全规约、工程结构等全方面规范。
---

# Java后端编码规范

本skill提供GDC Java后端开发规范的完整编码标准,涵盖命名、OOP、集合、并发、异常、测试、安全、工程结构等各个方面。

## 命名规范

### 【强制】基础命名规则

- 严禁使用下划线或美元符号开头/结尾: `_name`, `__name`, `$name`, `name$`, `name__`
- 严禁使用拼音与英文混合,或直接使用中文
- 包名统一使用小写,点分隔符之间仅一个英语单词,使用单数形式
- POJO类布尔变量不加is前缀(避免序列化错误)

### 【强制】命名风格

- **类名**: UpperCamelCase风格,例外: DO/BO/DTO/VO/AO/PO/UID
- **方法/参数/成员变量/局部变量**: lowerCamelCase,遵循驼峰形式
- **常量**: 全大写,下划线分隔,语义完整(如: MAX_STOCK_COUNT, CACHE_EXPIRED_TIME)
- **抽象类**: Abstract/Base开头
- **异常类**: Exception结尾
- **测试类**: 被测类名开头 + Test结尾
- **枚举类**: Enum后缀,成员全大写下划线分隔

### 【强制】数组定义

类型与中括号紧挨: `int[] arrayDemo` 而非 `String args[]`

### 【强制】Service/DAO层命名

- 获取单个对象: `get`前缀
- 获取多个对象: `list`前缀 + 复数形式结尾(如: listObjects)
- 获取统计值: `count`前缀
- 插入: `save/insert`前缀
- 删除: `remove/delete`前缀
- 修改: `update`前缀

### 【强制】领域模型命名

- 数据对象: `xxxDO`, xxx为数据表名
- 数据传输对象: `xxxDTO`, xxx为业务领域相关名称
- 展示对象: `xxxVO`, xxx一般为网页名称
- 禁止命名成xxxPOJO

## 常量定义

### 【强制】

- 不允许魔法值直接出现在代码中
- long/Long赋值必须使用大写L: `long a = 2L` (非2l)

### 【推荐】

- 按功能归类分开维护常量(如: CacheConsts, ConfigConsts)
- 常量复用层次: 跨应用 > 应用内 > 子工程 > 包内 > 类内
- 值固定范围内使用enum类型定义

## 代码格式

### 【强制】

- 大括号: 左大括号前不换行,后换行;右大括号前换行,后有else不换行,终止则必须换行
- if/for/while/switch/do保留字与括号间必须加空格
- 运算符左右两边必须加空格(包括=, &&, +-*/等)
- 采用4个空格缩进,禁止使用tab
- 单行字符数不超过120,超限换行规则:
  - 第二行缩进4空格,第三行起不继续缩进
  - 运算符与下文一起换行
  - 方法调用点号与下文一起换行
  - 多参数换行在逗号后
  - 括号前不要换行
- 方法参数逗号后必须加空格
- IDE编码设为UTF-8,换行符使用Unix格式

### 【推荐】

- 单个方法总行数不超过80行(含签名、注释、空行等)
- 不同逻辑/语义/业务代码间插入一个空行分隔

## OOP规约

### 【强制】

- 通过类名访问静态变量/方法,禁止通过对象引用访问
- 覆写方法必须加`@Override`注解
- 相同参数类型和业务含义才使用可变参数,避免使用Object;可变参数必须在最后
- 外部调用/二方库依赖的接口禁止修改方法签名,过时加`@Deprecated`并说明新接口
- 禁止使用过时的类或方法
- Object的equals用常量或确定有值对象调用: `"test".equals(object)` 而非 `object.equals("test")`
- 所有包装类对象值比较使用equals方法(不使用==)
- POJO类属性必须使用包装数据类型
- RPC方法返回值和参数必须使用包装数据类型
- 局部变量使用基本数据类型
- DO/DTO/VO等POJO类不设定任何属性默认值
- 序列化类新增属性不修改serialVersionUID
- 构造方法禁止加入业务逻辑,初始化放init方法
- POJO类必须写toString方法
- POJO类禁止同时存在isXxx()和getXxx()方法

### 【推荐】

- 循环体字符串连接使用StringBuilder.append()
- final用于:不可继承类/不可修改域对象/不可重写方法/运行不可重赋值局部变量
- 慎用Object.clone(),默认浅拷贝,深拷贝需重写
- 访问控制从严:private > protected > default > public
- 类内方法顺序:公有/保护方法 > 私有方法 > getter/setter
- setter方法参数名与成员变量名一致:`this.成员名 = 参数名`,不增加业务逻辑

## 集合处理

### 【强制】

- 重写equals必须重写hashCode(Set存储对象/Map的键必须重写这两个方法)
- ArrayList的subList结果不可强转成ArrayList
- subList场景高度注意原集合增删导致子列表ConcurrentModificationException
- 集合转数组使用`toArray(T[] array)`,数组大小=list.size()
- Arrays.asList()转换的集合不能使用add/remove/clear方法
- `<? extends T>`泛型集合不能使用add方法,`<? super T>`不能使用get方法
- foreach循环里禁止remove/add操作,使用Iterator方式
- JDK7+ Comparator必须满足: 自反/传递/一致性
- 不要在foreach循环里进行元素remove/add操作

### 【推荐】

- JDK7+使用diamond语法`<>`或全省略泛型
- 集合初始化指定初始值大小: `new HashMap<>(16)` (初始容量=元素个数/负载因子+1)
- 使用entrySet遍历Map而非keySet (效率更高,JDK8可用Map.foreach)
- 注意Map能否存储null值(HashMap可以,ConcurrentHashMap不可以)
- 利用Set元素唯一性快速去重,避免List.contains遍历

## 并发处理

### 【强制】

- 单例对象和方法必须保证线程安全
- 创建线程/线程池指定有意义的线程名称
- 线程资源必须通过线程池提供,禁止显式创建线程
- 禁止使用Executors创建线程池,使用ThreadPoolExecutor方式(避免OOM)
- SimpleDateFormat线程不安全,static定义必须加锁,使用ThreadLocal或DateUtils
- 高并发同步调用考量锁性能:无锁数据结构>锁区块>锁整个方法体>对象锁>类锁
- 多资源/数据库/对象加锁保持一致顺序(避免死锁)
- 并发修改同一记录需加锁:应用层/缓存层/数据库乐观锁(version作为更新依据),重试次数≥3
- 多线程并行定时任务使用ScheduledExecutorService而非Timer(Timer异常导致其他任务终止)

### 【推荐】

- 使用CountDownLatch异步转同步,每个线程退出前必须调用countDown,确保异常时也能执行
- 避免Random实例多线程使用,使用ThreadLocalRandom(JDK7+)
- 双重检查锁延迟初始化目标属性声明为volatile

## 控制语句

### 【强制】

- switch块内每个case通过break/return终止或注释继续执行哪个case,必须包含default语句放在最后
- if/else/for/while/do语句必须使用大括号(即使只有一行)
- 高并发场景避免使用"等于"判断作为中断/退出条件,使用大于/小于区间判断

### 【推荐】

- 表达异常分支少用if-else,改写成if后return的卫语句方式
- if-else避免超过3层,超限使用卫语句/策略模式/状态模式
- 复杂逻辑判断赋值给有意义的布尔变量
- 循环体内定义对象/变量/数据库连接/try-catch移至循环体外
- 避免取反逻辑运算符(如:使用`if (x < 628)`而非`if (!(x >= 628))`)

## 注释规约

### 【强制】

- 类/类属性/类方法必须使用Javadoc规范`/**内容*/`,禁止使用`// xxx`
- 所有抽象方法(含接口方法)必须Javadoc注释,说明返回值/参数/异常/方法功能/实现要求
- 所有类必须添加创建者和创建日期
- 方法内部单行注释使用`//`在上方另起一行;多行注释使用`/* */`与代码对齐
- 所有枚举类型字段必须注释说明用途

### 【推荐】

- 使用中文注释清晰说明问题,专有名词与关键字保持英文
- 代码修改同时修改注释(参数/返回值/异常/核心逻辑)
- 注释准确反映设计思想和代码逻辑/业务含义
- 好的命名和代码结构自解释,注释精简准确,避免过多过滥
- 特殊注释标记注明标记人和标记时间:
  - TODO(标记人,标记时间,[预计处理时间])
  - FIXME(标记人,标记时间,[预计处理时间])

## 异常处理

### 【强制】

- 可预检查规避的RuntimeException(NPE,IndexOutOfBoundsException等)不应通过catch处理
- 异常不要用作流程控制
- catch分清稳定和非稳定代码,尽可能区分异常类型进行对应处理
- 捕获异常必须处理或抛给调用者,最外层业务使用者必须处理并转化为用户可理解内容
- try块在事务代码中,catch异常后需回滚事务必须手动回滚
- finally块必须对资源对象/流对象关闭,异常也要try-catch
- 禁止在finally块中使用return
- 捕获异常与抛异常必须完全匹配或父类

### 【推荐】

- 方法返回可为null,必须注释说明什么情况下返回null
- 防止NPE场景:
  - 返回基本类型return包装对象可能NPE
  - 数据库查询结果可能为null
  - 集合元素即使isNotEmpty,取出数据可能为null
  - 远程调用返回对象必须空指针判断
  - Session获取数据建议NPE检查
  - 级联调用obj.getA().getB().getC()易产生NPE
- 定义区分unchecked/checked异常,避免直接抛new RuntimeException(),更不允许抛Exception/Throwable,使用业务含义自定义异常(DAOException/ServiceException等)
- 避免重复代码(DRY原则),抽取共性方法或抽象公共类

## 日志规约

### 【强制】

- 不可直接使用日志系统(Log4j/Logback)API,使用SLF4J门面模式API
  ```java
  import org.slf4j.Logger;
  import org.slf4j.LoggerFactory;
  private static final Logger logger = LoggerFactory.getLogger(Abc.class);
  ```
- 日志文件至少保存15天
- 扩展日志命名: `appName_logType_logName.log` (如: mppserver_monitor_timeZoneConvert.log)
- trace/debug/info级别使用条件输出或占位符
  ```java
  // 条件方式
  if (logger.isDebugEnabled()) {
    logger.debug("Processing trade with id: " + id + " and symbol: " + symbol);
  }
  // 占位符方式
  logger.debug("Processing trade with id: {} and symbol : {} ", id, symbol);
  ```
- 避免重复打印日志,log4j.xml设置additivity=false
- 异常信息包括案发现场信息和异常堆栈信息,不处理通过throws往上抛
  ```java
  logger.error(参数或对象toString() + "_" + e.getMessage(), e);
  ```

### 【推荐】

- 谨慎记录日志,生产环境禁止输出debug日志,有选择性输出info日志,warn记录用户输入参数错误,避免error频繁报警
- 日志错误信息尽量用英文,国际团队或海外部署强制全英文

## 单元测试

### 【强制】

- 遵守AIR原则(Automatic自动化/Independent独立性/Repeatable可重复)
- 单元测试全自动执行非交互式,禁止使用System.out验证,必须使用assert
- 保持单元测试独立性,测试用例间不能互相调用,不依赖执行先后次序
- 单元测试可重复执行,不受外界环境(网络/服务/中间件)影响
- 测试粒度足够小,至多类级别一般是方法级别,不负责跨类/跨系统交互逻辑
- 核心业务/应用/模块增量代码确保单元测试通过
- 单元测试代码写在src/test/java,禁止写在业务代码目录

### 【推荐】

- 语句覆盖率70%,核心模块语句覆盖率和分支覆盖率100%(DAO/Manager/可重用度高的Service都应测试)
- 遵守BCDE原则(Border边界值/Correct正确输入/Design结合设计/Error强制错误输入)
- 数据库查询/更新/删除操作不能假设数据存在,使用程序插入或导入方式准备数据
- 数据库相关单元测试设定自动回滚机制或明确前后缀标识
- 不可测代码建议重构使代码可测

## 安全规约

### 【强制】

- 用户个人页面/功能必须进行权限控制校验(防止水平权限校验缺失)
- 用户敏感数据禁止直接展示,必须脱敏(如手机号:158****9119)
- 用户输入SQL参数严格使用参数绑定或METADATA字段值限定,禁止字符串拼接SQL
- 用户请求传入任何参数必须做有效性验证(page size过大/恶意order by/任意重定向/SQL注入/反序列化注入/正则ReDoS)
- 禁止向HTML页面输出未经安全过滤或未正确转义的用户数据
- 表单/AJAX提交必须执行CSRF安全验证
- 使用平台资源(短信/邮件/电话/下单/支付)必须实现正确防重放机制(数量限制/疲劳度控制/验证码校验)

### 【推荐】

- 发贴/评论/发送即时消息等用户生成内容场景必须实现防刷/文本内容违禁词过滤等风控策略

## 工程结构

### 应用分层

```
开放接口层: 封装Service暴露RPC接口/Web封装http接口/网关安全控制/流量控制
终端显示层: 各端模板渲染并执行显示(velocity渲染/JS渲染/JSP渲染/移动端展示)
Web层: 访问控制转发/基本参数校验/不复用业务简单处理
Service层: 相对具体业务逻辑服务层
Manager层: 通用业务处理层(第三方平台封装/Service层通用能力下沉/与DAO层交互多DAO组合复用)
DAO层: 数据访问层,与MySQL/Oracle/Hbase等底层交互
外部接口/第三方平台: 其它部门RPC开放接口/基础平台/其它公司HTTP接口
```

### 分层异常处理

- DAO层: catch(Exception e),throw new DAOException(e),不打印日志(日志在Manager/Service层捕获打印)
- Service层: 异常必须记录出错日志到磁盘,尽可能带上参数信息
- Manager层: 与Service同机部署同DAO处理,单独部署同Service处理
- Web层: 不应该继续往上抛异常,直接跳转到友好错误页面加用户易理解提示
- 开放接口层: 异常处理成错误码和错误信息返回

### 领域模型

- DO(Data Object): 与数据库表结构一一对应,通过DAO层向上传输数据源对象
- DTO(Data Transfer Object): Service或Manager向外传输对象
- BO(Business Object): Service层输出封装业务逻辑对象
- AO(Application Object): Web层与Service层之间抽象复用对象模型,贴近展示层
- VO(View Object): Web向模板渲染引擎层传输对象
- Query: 各层接收上层查询请求,超过2个参数查询封装禁止使用Map类传输

### 二方库依赖

- GroupID: com.{公司/BU}.业务线[.子业务线],最多4级
- ArtifactID: 产品线名-模块名(如: dubbo-client/fastjson-api/jstorm-tool)
- Version: 主版本号.次版本号.修订号(起始1.0.0非0.0.1)
- 线上应用禁止依赖SNAPSHOT版本
- 二方库新增/升级保持除功能点外jar包仲裁结果不变
- 二方库可定义枚举类型参数,但接口返回值不允许使用枚举类型或包含枚举的POJO对象
- 依赖二方库群定义统一版本变量避免版本号不一致
- 禁止子项目pom依赖相同GroupId相同ArtifactId不同Version

### 服务器配置

- 高并发服务器调小TCP time_wait超时时间(linux: net.ipv4.tcp_fin_timeout = 30)
- 调大服务器支持最大文件句柄数(fd)
- JVM环境参数设置-XX:+HeapDumpOnOutOfMemoryError
- 生产环境JVM Xms和Xmx设置同样大小内存容量

## 设计规约

### 【强制】

- 存储方案和底层数据结构设计获得评审一致通过并沉淀成为文档
- 与系统交互User超过一类且相关User Case超过5个,使用用例图
- 某业务对象状态超过3个,使用状态图并明确状态变化触发条件
- 某功能调用链路涉及对象超过3个,使用时序图并明确各调用环节输入输出
- 系统中模型类超过5个且存在复杂依赖关系,使用类图并明确类之间关系
- 超过2个对象存在协作关系且需要表示复杂处理流程,使用活动图

### 【推荐】

- 需求分析与系统设计考虑主干功能同时充分评估异常流程与业务边界
- 类设计符合单一原则
- 谨慎使用继承扩展,优先聚合/组合方式(必须符合里氏代换原则)
- 依赖倒置原则,尽量依赖抽象类与接口
- 对扩展开放对修改闭合
- 共性业务或公共行为抽取公共模块/配置/类/方法,避免重复代码或重复配置
- 避免敏捷开发=讲故事+编码+发布误解,核心关键点必要设计和文档沉淀需要
- 系统设计主要目的明确需求/理顺逻辑/后期维护,次要目的指导编码

### 系统架构设计目的

- 确定系统边界(技术层面做与不做)
- 确定系统内模块关系(依赖关系及模块宏观输入输出)
- 确定指导后续设计与演化原则
- 确定非功能性需求(安全性/可用性/可扩展性等)

## 专有名词解释

- POJO: 只有setter/getter/toString的简单类(含DO/DTO/BO/VO等)
- GAV: GroupId/ArtifactId/Version, Maven坐标唯一标识jar包
- OOP: 类/对象编程处理方式
- ORM: 对象关系映射(泛指iBATIS/mybatis等框架)
- NPE: 空指针异常
- SOA: 面向服务架构
- IDE: 程序开发环境(泛指IntelliJ IDEA/eclipse)
- OOM: 源于OutOfMemoryError,JVM没有足够内存为对象分配空间
- 一方库: 本工程内部子项目模块依赖库
- 二方库: 公司内部发布到中央仓库可供其它应用依赖库
- 三方库: 公司之外开源库

## 使用场景

使用本skill的场景包括但不限于:

1. **代码审查**: 检查代码是否符合规范要求
2. **代码生成**: 根据规范生成符合标准的Java代码
3. **规范检查**: 识别违反规范的地方并建议修改
4. **代码重构**: 按照规范重构不合规代码
5. **技术评审**: 基于规范评审设计和实现方案
6. **质量保障**: 确保代码质量符合团队标准

## 快速参考

代码问题时参考以下章节:
- 命名问题 → 命名规范
- 类设计问题 → OOP规约
- 集合使用 → 集合处理
- 并发问题 → 并发处理
- 逻辑控制 → 控制语句
- 注释缺失 → 注释规约
- 异常处理 → 异常处理
- 日志记录 → 日志规约
- 测试覆盖 → 单元测试
- 安全漏洞 → 安全规约
- 架构设计 → 工程结构/设计规约
