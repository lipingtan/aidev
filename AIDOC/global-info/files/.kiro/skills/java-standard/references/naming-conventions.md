# Java命名规范详细参考

## 包名规范

### 【强制】基本规则
- 统一使用小写
- 点分隔符之间有且仅有一个自然语义的英语单词
- 统一使用单数形式
- 类名如有复数含义,类名可使用复数形式

### 正例
```
com.alibaba.ai.util
com.alibaba.service.user
com.alibaba.dao.product
```

### 反例
```
com.alibaba.AiUtil          // 大写开头
com.alibaba.ai.utils         // 复数形式
com.alibaba.service.userService // 子模块使用复数
```

## 类名规范

### 【强制】UpperCamelCase风格
例外情况: DO/BO/DTO/VO/AO/PO/UID

### 正例
```
JavaServerlessPlatform
UserDO
XmlService
TcpUdpDeal
TaPromotion
```

### 反例
```
javaserverlessplatform  // 全小写
UserDo               // 只有首字母大写
XMLService            // 全大写缩写
TCPUDPDeal           // 缩写全大写
TAPromotion          // 缩写全大写
```

## 方法/变量命名

### 【强制】lowerCamelCase驼峰形式

### 正例
```
localValue
getHttpMessage()
inputUserId
```

### 反例
```
LocalValue           // 首字母大写
gethttpmessage()     // 无驼峰
inputuserid          // 全小写
```

## 常量命名

### 【强制】规则
- 全部大写
- 单词间下划线分隔
- 力求语义表达完整清楚
- 不嫌名字长

### 正例
```
MAX_STOCK_COUNT
CACHE_EXPIRED_TIME
DEFAULT_PAGE_SIZE
```

### 反例
```
MAX_COUNT           // 语义不清
EXPIRED_TIME       // 缺少CACHE前缀
maxStockCount       // 驼峰形式
```

## Service/DAO方法前缀

### 【强制】方法命名规约

| 操作类型 | 前缀 | 示例 |
|---------|-------|------|
| 获取单个对象 | get | getUserById() |
| 获取多个对象 | list + 复数形式 | listUsers(), listOrders() |
| 获取统计值 | count | countUsers(), countOrders() |
| 插入 | save/insert | saveUser(), insertOrder() |
| 删除 | remove/delete | removeUser(), deleteOrder() |
| 修改 | update | updateUser(), updateOrder() |

### 正例
```
public UserDO getUserById(Long userId);
public List<OrderDO> listOrders(Long userId);
public int countUsersByStatus(Integer status);
public void saveUser(UserDO user);
public void deleteUser(Long userId);
public void updateUser(UserDO user);
```

### 反例
```
public UserDO findUser(Long userId);           // 应该用get
public List<OrderDO> getOrders(Long userId);   // 应该用list
public int userCount();                      // 应该用count
public void add(UserDO user);                // 应该用save
public OrderDO updateOrder(Long orderId);       // 返回类型应该是void/影响行数
```

## 领域对象命名

### 【强制】领域模型命名

| 类型 | 命名规则 | 说明 | 示例 |
|-----|----------|------|------|
| DO | xxxDO | xxx为数据表名 | UserDO, OrderDO |
| DTO | xxxDTO | xxx为业务领域相关名称 | UserDTO, OrderDTO |
| VO | xxxVO | xxx一般为网页名称 | UserVO, OrderVO |
| BO | xxxBO | 业务对象 | UserBO, OrderBO |
| AO | xxxAO | 应用对象 | UserAO, OrderAO |
| POJO | xxxPOJO | 禁止使用 | - |

### 正例
```
// DO对象 - 数据表user_user_info
public class UserInfoDO {
    private Long id;
    private String userName;
}

// DTO对象 - 用户服务传输
public class UserDTO {
    private Long userId;
    private String name;
}

// VO对象 - 页面展示
public class UserVO {
    private Long id;
    private String displayName;
}
```

### 反例
```
public class UserPOJO {  // 禁止使用POJO作为后缀
    private Long id;
}
```

## 特殊命名规范

### 【强制】布尔类型变量

**POJO类布尔类型变量不要加is前缀**

原因: 部分框架解析会引起序列化错误

### 错误示例
```java
public class UserDO {
    private Boolean isDeleted;  // 错误
    public Boolean isDeleted() {
        return isDeleted;
    }
}
// RPC框架反向解析时"误以为"属性名是deleted,导致属性获取不到
```

### 正确示例
```java
public class UserDO {
    private Boolean deleted;  // 正确
    public Boolean getDeleted() {
        return deleted;
    }
}
```

### 【强制】接口与实现类命名

#### 1) Service和DAO类
```
接口: CacheService
实现类: CacheServiceImpl
接口: UserDao
实现类: UserDaoImpl
```

#### 2) 形容能力的接口(通常-able形容词)
```
接口: Translatable
实现类: AbstractTranslator
接口: Serializable
实现类: UserSerializable
```

## 抽象类/异常类/测试类/枚举类命名

### 【强制】命名模式

| 类型 | 命名模式 | 示例 |
|-----|----------|------|
| 抽象类 | Abstract/Base开头 | AbstractUserService, BaseDao |
| 异常类 | Exception结尾 | UserException, ServiceException |
| 测试类 | 被测类名 + Test结尾 | UserServiceTest, UserDaoTest |
| 枚举类 | xxxEnum | SeasonEnum, OrderStatusEnum |

### 正例
```
// 抽象类
public abstract class AbstractUserService {
    protected abstract void validateUser(User user);
}

// 异常类
public class UserException extends RuntimeException {
    public UserException(String message) {
        super(message);
    }
}

// 测试类
public class UserServiceTest {
    @Test
    public void testGetUserById() {
        // 测试代码
    }
}

// 枚举类
public enum SeasonEnum {
    SPRING(1), SUMMER(2), AUTUMN(3), WINTER(4);
    private int seq;
    SeasonEnum(int seq) {
        this.seq = seq;
    }
}
```

## 禁止的命名方式

### 【强制】避免完全相同的命名

1. 子父类成员变量之间禁止完全相同命名
2. 不同代码块的局部变量之间禁止完全相同命名
3. 非setter/getter的参数名称禁止与成员变量名称相同

### 反例
```java
public class ConfusingName {
    public int age;
    // 非setter/getter参数名称,不允许与本类成员变量同名
    public void getData(String alibaba) {
        if (condition) {
            final int money = 531;
        }
        for (int i = 0; i < 10; i++) {
            // 在同一方法体中,不允许与其它代码块中的money命名相同
            final int money = 615;
        }
    }
}

class Son extends ConfusingName {
    // 不允许与父类的成员变量名称相同
    public int age;
}
```

### 【强制】杜绝不规范缩写

禁止望文不知义的随意缩写

### 反例
```
AbstractClass → AbsClass    // 随意缩写降低可读性
condition → condi        // 无意义缩写
```

### 正例
```
AtomicReferenceFieldUpdater  // 完整单词表达意图
UserManager             // 清晰命名
```

## 推荐的命名实践

### 【推荐】类型名词放在词尾

提升辨识度

### 正例
```
startTime        // 开始时间
workQueue        // 工作队列
nameList         // 名称列表
TERMINATED_THREAD_COUNT  // 终止线程数
```

### 反例
```
startedAt       // 时间后缀在前
QueueOfWork     // 队列后缀在前
listName        // 列表后缀在前
COUNT_TERMINATED_THREAD  // 计数前缀在前
```

### 【推荐】体现设计模式

接口/类使用设计模式,命名时体现具体模式

### 正例
```
public class OrderFactory;      // 工厂模式
public class LoginProxy;       // 代理模式
public class ResourceObserver;   // 观察者模式
public class UserBuilder;       // 建造者模式
public class UserAdapter;       // 适配器模式
```

### 【推荐】完整单词组合

自定义编程元素命名使用尽量完整的单词组合

### 正例
```
AtomicReferenceFieldUpdater  // JDK中的完整命名
ThreadPoolExecutor        // 完整单词组合
```

### 反例
```
int a;  // 随意命名,无含义
```