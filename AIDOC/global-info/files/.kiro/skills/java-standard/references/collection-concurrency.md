# 集合与并发处理详细规范

## 集合处理规范

### 【强制】hashCode和equals

**规则1**: 只要重写equals,就必须重写hashCode
**规则2**: Set存储的是不重复对象,依据hashCode和equals判断,Set存储对象必须重写这两个方法
**规则3**: 自定义对象作为Map的键,必须重写hashCode和equals

### 正确示例
```java
public class UserDO {
    private Long id;
    private String name;

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        UserDO userDO = (UserDO) o;
        return Objects.equals(id, userDO.id) &&
               Objects.equals(name, userDO.name);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id, name);
    }
}

// Set使用
Set<UserDO> userSet = new HashSet<>();
userSet.add(user1);
userSet.add(user2);  // 自动根据equals+hashCode去重

// Map键使用
Map<UserDO, OrderDO> userOrderMap = new HashMap<>();
userOrderMap.put(user, order);  // 自动根据equals+hashCode定位
```

### 【强制】subList注意事项

**规则1**: ArrayList的subList结果不可强转成ArrayList
**规则2**: 高度注意对原集合增删会导致子列表ConcurrentModificationException

### 反例
```java
List<String> list = new ArrayList<>(Arrays.asList("a", "b", "c"));
List<String> subList = list.subList(0, 2);

// 错误: ClassCastException
ArrayList<String> castedList = (ArrayList<String>) subList;

// 错误: 修改原集合导致子列表异常
list.add("d");
for (String s : subList) {  // ConcurrentModificationException
    System.out.println(s);
}
```

### 正确示例
```java
List<String> list = new ArrayList<>(Arrays.asList("a", "b", "c"));
List<String> subList = list.subList(0, 2);

// 方式1: 创建新的ArrayList
List<String> newSubList = new ArrayList<>(subList);

// 方式2: 不修改原集合
for (String s : subList) {
    System.out.println(s);
}
```

### 【强制】集合转数组

必须使用`toArray(T[] array)`,传入类型完全一样数组,大小=list.size()

### 正确示例
```java
List<String> list = new ArrayList<>(2);
list.add("guan");
list.add("bao");

// 正确: 使用toArray带参方法
String[] array = new String[list.size()];
array = list.toArray(array);  // array = ["guan", "bao"]
```

### 反例
```java
// 错误: 直接使用toArray无参方法
Object[] objArray = list.toArray();
String[] stringArray = (String[]) objArray;  // ClassCastException

// 错误: 数组大小不对
String[] array = new String[10];
array = list.toArray(array);  // array = ["guan", "bao", null, null, ...]
```

### 【强制】Arrays.asList

Arrays.asList转换的集合不能使用add/remove/clear方法

### 错误示例
```java
String[] str = new String[]{"you", "wu"};
List<String> list = Arrays.asList(str);

// 运行时异常: UnsupportedOperationException
list.add("yangguanbao");

// 原数组修改会影响列表
str[0] = "gujin";
System.out.println(list.get(0));  // "gujin"
```

### 正确示例
```java
String[] str = new String[]{"you", "wu"};
// 使用ArrayList包装
List<String> list = new ArrayList<>(Arrays.asList(str));
list.add("yangguanbao");  // 正常工作
```

### 【强制】泛型通配符PECS原则

**Producer Extends Consumer Super**

- `<? extends T>`: 频繁往外读取内容(生产者)
- `<? super T>`: 经常往里插入(消费者)

### 错误示例
```java
// <? extends T>不能add
List<? extends Number> producerList = new ArrayList<Integer>();
producerList.add(1);  // 编译错误

// <? super T>不能get具体类型
List<? super Integer> consumerList = new ArrayList<Number>();
Integer i = consumerList.get(0);  // 编译错误
```

### 正确示例
```java
// Producer Extends - 读取
List<? extends Number> producerList = new ArrayList<Integer>();
for (Number n : producerList) {  // 可以get
    System.out.println(n);
}

// Consumer Super - 插入
List<? super Integer> consumerList = new ArrayList<Number>();
consumerList.add(1);  // 可以add
Object obj = consumerList.get(0);  // get为Object
```

### 【强制】foreach循环删除元素

不要在foreach循环里进行remove/add操作

### 错误示例
```java
List<String> list = new ArrayList<>();
list.add("1");
list.add("2");

// ConcurrentModificationException或结果不符合预期
for (String item : list) {
    if ("1".equals(item)) {
        list.remove(item);
    }
}
```

### 正确示例
```java
List<String> list = new ArrayList<>();
list.add("1");
list.add("2");

// 使用Iterator方式
Iterator<String> iterator = list.iterator();
while (iterator.hasNext()) {
    String item = iterator.next();
    if (删除元素的条件) {
        iterator.remove();
    }
}

// 或使用removeIf (JDK8+)
list.removeIf(item -> "1".equals(item));
```

### 【强制】Comparator三条件

JDK7+ Comparator实现必须满足:
1. x,y比较结果和y,x比较结果相反(自反性)
2. x>y, y>z, 则x>z(传递性)
3. x=y, 则x,z比较结果和y,z比较结果相同(一致性)

### 错误示例
```java
new Comparator<Student>() {
    @Override
    public int compare(Student o1, Student o2) {
        // 错误: 未处理相等情况,可能抛IllegalArgumentException
        return o1.getId() > o2.getId() ? 1 : -1;
    }
};
```

### 正确示例
```java
new Comparator<Student>() {
    @Override
    public int compare(Student o1, Student o2) {
        return Integer.compare(o1.getId(), o2.getId());
    }
};
```

### 【推荐】集合初始化指定大小

```java
// HashMap初始化: initialCapacity = (元素个数 / 负载因子) + 1
// 负载因子默认0.75,暂时无法确定初始值请设置16

// 正确: 初始容量1024
int size = 1024;
Map<String, Object> map = new HashMap<>((int) (size / 0.75) + 1);

// 正确: 使用默认容量
Map<String, Object> defaultMap = new HashMap<>(16);
```

### 【推荐】Map遍历使用entrySet

```java
// 效率更高: entrySet只遍历一次
Map<String, UserDO> userMap = new HashMap<>();
for (Map.Entry<String, UserDO> entry : userMap.entrySet()) {
    String key = entry.getKey();
    UserDO value = entry.getValue();
    System.out.println(key + ": " + value.getName());
}

// JDK8更简洁: Map.forEach
userMap.forEach((key, value) ->
    System.out.println(key + ": " + value.getName())
);
```

### 【推荐】Map null值支持

| Map类型 | Key支持null | Value支持null | 线程安全 |
|---------|------------|--------------|----------|
| HashMap | 是 | 是 | 否 |
| ConcurrentHashMap | 否 | 否 | 是 |
| TreeMap | 否 | 是 | 否 |
| Hashtable | 否 | 是 | 是 |

```java
// 错误: ConcurrentHashMap不支持null值
Map<String, String> map = new ConcurrentHashMap<>();
map.put("key", null);  // NPE

// HashMap支持null
Map<String, String> hashMap = new HashMap<>();
hashMap.put(null, "value");  // 正常
hashMap.put("key", null);    // 正常
```

## 并发处理规范

### 【强制】线程池创建

禁止使用Executors,使用ThreadPoolExecutor

### 错误示例及原因
```java
// FixedThreadPool - 队列长度Integer.MAX_VALUE,可能堆积大量请求OOM
ExecutorService fixedPool = Executors.newFixedThreadPool(10);

// SingleThreadPool - 队列长度Integer.MAX_VALUE,可能堆积大量请求OOM
ExecutorService singlePool = Executors.newSingleThreadExecutor();

// CachedThreadPool - 创建线程数Integer.MAX_VALUE,可能创建大量线程OOM
ExecutorService cachedPool = Executors.newCachedThreadPool();

// ScheduledThreadPool - 创建线程数Integer.MAX_VALUE,可能创建大量线程OOM
ExecutorService scheduledPool = Executors.newScheduledThreadPool(10);
```

### 正确示例
```java
ThreadPoolExecutor executor = new ThreadPoolExecutor(
    10,                          // corePoolSize
    100,                         // maximumPoolSize
    60L,                         // keepAliveTime
    TimeUnit.SECONDS,               // 时间单位
    new LinkedBlockingQueue<>(1000), // workQueue指定大小
    new ThreadFactoryBuilder().setNameFormat("task-thread-%d").build(),
    new ThreadPoolExecutor.CallerRunsPolicy()  // 拒绝策略
);
```

### 【强制】SimpleDateFormat线程安全

SimpleDateFormat是线程不安全的类

### 错误示例
```java
// 错误: static变量多线程不安全
private static final SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd");

public String format(Date date) {
    return sdf.format(date);  // 多线程可能出错
}
```

### 正确示例
```java
// 方式1: 每次创建新实例
public String format(Date date) {
    return new SimpleDateFormat("yyyy-MM-dd").format(date);
}

// 方式2: ThreadLocal
private static final ThreadLocal<DateFormat> df = new ThreadLocal<DateFormat>() {
    @Override
    protected DateFormat initialValue() {
        return new SimpleDateFormat("yyyy-MM-dd");
    }
};

public String format(Date date) {
    return df.get().format(date);
}

// 方式3: 使用工具类
import org.apache.commons.lang3.time.DateFormatUtils;
public String format(Date date) {
    return DateFormatUtils.format(date, "yyyy-MM-dd");
}

// 方式4: JDK8+ DateTimeFormatter (线程安全)
private static final DateTimeFormatter formatter =
    DateTimeFormatter.ofPattern("yyyy-MM-dd");

public String format(LocalDateTime dateTime) {
    return dateTime.format(formatter);
}
```

### 【强制】并发修改乐观锁

并发修改同一记录需加锁,优先乐观锁

### 正确示例
```java
// 数据库表增加version字段
public class ProductDO {
    private Long id;
    private Integer stock;
    private Integer version;  // 版本号
}

// 更新操作
public void updateStock(Long productId, int delta) {
    ProductDO product = productDao.selectById(productId);
    while (true) {
        int currentVersion = product.getVersion();
        int newStock = product.getStock() - delta;

        // 更新时检查version是否变化
        int updateCount = productDao.updateStockAndVersion(
            productId,
            newStock,
            currentVersion,
            currentVersion + 1
        );

        if (updateCount > 0) {
            break;  // 更新成功
        }

        // 重试至少3次
        product = productDao.selectById(productId);
    }
}

// SQL: UPDATE product SET stock=?, version=? WHERE id=? AND version=?
```

### 【强制】CountDownLatch使用

每个线程退出前必须调用countDown,确保异常时也能执行

### 正确示例
```java
final CountDownLatch latch = new CountDownLatch(threadCount);
for (int i = 0; i < threadCount; i++) {
    new Thread(() -> {
        try {
            // 业务逻辑
            doWork();
        } catch (Exception e) {
            logger.error("工作线程异常", e);
        } finally {
            // 确保countDown被执行
            latch.countDown();
        }
    }).start();
}

try {
    // 等待所有线程完成,设置超时
    latch.await(30, TimeUnit.SECONDS);
} catch (InterruptedException e) {
    logger.error("等待线程完成被中断", e);
    Thread.currentThread().interrupt();
}
```

### 【推荐】双重检查锁

延迟初始化目标属性声明为volatile

### 正确示例
```java
class LazyInitDemo {
    // 目标属性声明为volatile
    private volatile Helper helper = null;

    public Helper getHelper() {
        if (helper == null) {
            synchronized (this) {
                if (helper == null) {
                    helper = new Helper();
                }
            }
        }
        return helper;
    }
}
```

### 【推荐】volatile使用场景

volatile解决多线程内存不可见问题:
- 一写多读: 可以解决变量同步
- 多写: 无法解决线程安全

### 正确示例
```java
// 一写多读: 可以使用volatile
private volatile boolean running = true;

public void stop() {
    running = false;  // 写操作
}

public void run() {
    while (running) {  // 读操作
        // 工作
    }
}

// 多写: 使用Atomic类
private AtomicInteger count = new AtomicInteger();
public void increment() {
    count.incrementAndGet();
}
```

### 【推荐】ThreadLocal使用

ThreadLocal对象建议使用static修饰

### 正确示例
```java
public class UserContext {
    // static修饰,所有此类实例共享此静态变量
    private static final ThreadLocal<UserDO> userThreadLocal =
        new ThreadLocal<UserDO>();

    public static void setUser(UserDO user) {
        userThreadLocal.set(user);
    }

    public static UserDO getUser() {
        return userThreadLocal.get();
    }

    public static void removeUser() {
        userThreadLocal.remove();  // 防止内存泄漏
    }
}
```

### 【推荐】HashMap并发问题

HashMap在容量不够resize时高并发可能死链,导致CPU飙升

### 解决方案
```java
// 方式1: 使用ConcurrentHashMap (推荐)
ConcurrentHashMap<String, Object> concurrentMap = new ConcurrentHashMap<>();

// 方式2: 使用Collections.synchronizedMap包装
Map<String, Object> synchronizedMap =
    Collections.synchronizedMap(new HashMap<>());
// 使用时仍需手动同步
synchronized (synchronizedMap) {
    synchronizedMap.put(key, value);
}

// 方式3: 业务代码层面加锁
private final Object lock = new Object();
private Map<String, Object> map = new HashMap<>();

public void put(String key, Object value) {
    synchronized (lock) {
        map.put(key, value);
    }
}
```

## 定时任务并发

### 【强制】使用ScheduledExecutorService

Timer运行多个TimeTask时,只要其中之一未捕获异常,其它任务自动终止

### 错误示例
```java
// 错误: 使用Timer
Timer timer = new Timer();
timer.schedule(new TimerTask() {
    @Override
    public void run() {
        // 如果这里抛异常,其他任务也会终止
        doTask1();
    }
}, 0, 1000);

timer.schedule(new TimerTask() {
    @Override
    public void run() {
        doTask2();
    }
}, 0, 2000);
```

### 正确示例
```java
// 正确: 使用ScheduledExecutorService
ScheduledExecutorService executor = Executors.newScheduledThreadPool(10);

executor.scheduleAtFixedRate(() -> {
    try {
        // 捕获异常,不影响其他任务
        doTask1();
    } catch (Exception e) {
        logger.error("Task1执行异常", e);
    }
}, 0, 1, TimeUnit.SECONDS);

executor.scheduleAtFixedRate(() -> {
    try {
        doTask2();
    } catch (Exception e) {
        logger.error("Task2执行异常", e);
    }
}, 0, 2, TimeUnit.SECONDS);
```