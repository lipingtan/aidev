# Java单元测试规范与最佳实践

本文档定义了 Java 单元测试的标准规范、Mockito 使用最佳实践和测试用例设计指南，作为团队的统一标准。

---

## 目录

1. [Mockito使用规范](#mockito使用规范)
2. [测试用例设计规范](#测试用例设计规范)
3. [测试数据构建模式](#测试数据构建模式)
4. [断言使用指南](#断言使用指南)

---

# Mockito使用规范

## 1. 链式调用模拟标准

### 1.1 正确模拟方式
```java
// 逐层创建Mock对象
HistoricTaskInstanceQuery query = mock(HistoricTaskInstanceQuery.class);
HistoricTaskInstance task = mock(HistoricTaskInstance.class);

// 逐层配置返回值
when(historyService.createHistoricTaskInstanceQuery()).thenReturn(query);
when(query.taskId(anyString())).thenReturn(query);
when(query.singleResult()).thenReturn(task);
when(task.getId()).thenReturn("taskId");
```

### 1.2 错误示例（禁止）
```java
// 错误：直接模拟链式调用
when(historyService.createHistoricTaskInstanceQuery().taskId(taskId).singleResult())
    .thenReturn(historicTask);
```

---

## 2. 工作流引擎（Flowable）模拟模板

```java
@Mock
private HistoryService historyService;

@Test
void testGetDetail() {
    // 模拟历史变量查询
    HistoricVariableInstanceQuery variableQuery = mock(HistoricVariableInstanceQuery.class);
    when(historyService.createHistoricVariableInstanceQuery()).thenReturn(variableQuery);
    when(variableQuery.taskIds(anySet())).thenReturn(variableQuery);
    when(variableQuery.list()).thenReturn(new ArrayList<>());
    
    // 模拟历史流程实例查询
    HistoricProcessInstanceQuery processQuery = mock(HistoricProcessInstanceQuery.class);
    when(historyService.createHistoricProcessInstanceQuery()).thenReturn(processQuery);
    when(processQuery.processInstanceId(anyString())).thenReturn(processQuery);
    when(processQuery.singleResult()).thenReturn(mock(HistoricProcessInstance.class));
}
```

---

## 3. 参数匹配器使用规范

### 3.1 常用匹配器
| 匹配器 | 用途 | 使用场景 |
|-------|------|---------|
| `any()` | 匹配任意对象 | 通用场景 |
| `anyString()` | 匹配任意字符串 | 字符串参数 |
| `anyList()` | 匹配任意List | 集合参数 |
| `anySet()` | 匹配任意Set | Set参数 |
| `any(Class.class)` | 匹配指定类型 | 复杂对象 |
| `eq(value)` | 精确匹配值 | 混合使用时 |
| `isNull()` | 匹配null值 | null参数测试 |

### 3.2 使用原则（强制）
1. **一致性原则**：同一方法调用中，要么全部使用匹配器，要么全部使用具体值
2. **可读性优先**：优先使用具体值，可提高测试可读性
3. **复杂对象处理**：对于复杂对象，使用`any(Class.class)`

**正确示例**：
```java
// ✅ 全部使用匹配器
when(mapper.query(anyList(), any(Date.class))).thenReturn(result);

// ✅ 全部使用具体值
when(mapper.query(specificList, specificDate)).thenReturn(result);

// ❌ 错误：混合使用（未使用eq）
// when(mapper.query(specificList, any(Date.class))).thenReturn(result);

// ✅ 混合使用时必须用eq
when(mapper.query(eq(specificList), any(Date.class))).thenReturn(result);
```

---

## 4. 验证（Verify）使用规范

### 4.1 基本验证
```java
@Test
void testMethodCalls() {
    // 执行被测方法
    service.process(input);
    
    // 验证方法被调用
    verify(mockMapper).insert(any());
    
    // 验证调用次数
    verify(mockMapper, times(2)).update(any());
    
    // 验证从未调用
    verify(mockMapper, never()).delete(any());
}
```

### 4.2 参数捕获验证
```java
@Test
void testParameterCapture() {
    service.process(input);
    
    // 验证调用参数
    ArgumentCaptor<Entity> captor = ArgumentCaptor.forClass(Entity.class);
    verify(mockMapper).insert(captor.capture());
    assertEquals("expectedValue", captor.getValue().getName());
}
```

---

## 5. 异常模拟规范

### 5.1 标准异常模拟
```java
@Test
void testExceptionHandling() {
    // 模拟抛出异常
    when(mockMapper.query(any())).thenThrow(new RuntimeException("数据库异常"));
    
    // 验证异常被正确处理
    assertThrows(BusinessException.class, () -> {
        service.process(input);
    });
}
```

### 5.2 Void方法异常
```java
@Test
void testVoidMethodException() {
    // void方法抛出异常
    doThrow(new RuntimeException()).when(mockMapper).delete(any());
    
    // 验证异常传播
    assertThrows(RuntimeException.class, () -> {
        service.processDelete(id);
    });
}
```

---

## 6. lenient() 使用场景

**适用场景**（仅限以下情况）：
- 在`@BeforeEach`中配置的通用Mock，不是每个测试都会调用
- 测试多个分支时，某些stubbing只在特定分支被调用
- 模拟可选的回调或监听器

```java
@BeforeEach
void setUp() {
    // 通用配置，不是每个测试都会调用
    lenient().when(configService.getConfig(anyString())).thenReturn("defaultValue");
}
```

**注意**：避免滥用 `lenient()`，优先通过调整测试结构消除不必要的 stubbing。

---

# 测试用例设计规范

## 1. 测试方法命名规范

### 1.1 命名格式（强制）
- **简单测试**：`方法名()`
- **场景测试**：`方法名_场景后缀()`

### 1.2 标准场景后缀
| 后缀 | 含义 | 使用场景 |
|-----|------|---------|
| `_Normal` | 正常流程 | 主流程测试 |
| `_EmptyResult` | 空结果 | 返回空集合 |
| `_NullResult` | Null返回值 | 返回null |
| `_NullCount` | Null计数 | 计数为null |
| `_PartitionLogic` | 分区逻辑 | 超过999元素 |
| `_SinglePartition` | 单分区 | 少于999元素 |
| `_EmptyUserSet` | 空用户集 | 空集合输入 |
| `_ZeroValue` | 零值 | 数值为0 |
| `_NegativeValue` | 负值 | 负数输入 |
| `_BigDecimalRounding` | 精度处理 | 四舍五入 |
| `_DivisionByZero` | 除零处理 | 除数为0 |
| `_WithValidTime` | 有效时间 | 时间格式化 |
| `_WithLargeTime` | 大时间值 | 超大时间 |
| `_WithSmallTime` | 小时间值 | 极小时间 |

---

## 2. 必测场景清单

### 2.1 数据查询类方法（三个必测场景）

#### ✅ 场景1：正常流程
```java
/**
 * 测试目的：验证正常情况下能正确返回用户登录次数
 */
@Test
void getLoginCount_Normal() {
    // Arrange
    List<UserCountVo> countList = Arrays.asList(vo1, vo2);
    when(mapper.getLoginCount(anyList(), any(Date.class))).thenReturn(countList);
    
    // Act
    Map<String, String> result = service.getLoginCount(userNoSet, date);
    
    // Assert
    assertEquals("5", result.get("user1"));
}
```

#### ✅ 场景2：空结果
```java
/**
 * 测试目的：验证查询无数据时返回空结果而不是抛出异常
 */
@Test
void getLoginCount_EmptyResult() {
    when(mapper.getLoginCount(anyList(), any(Date.class))).thenReturn(new ArrayList<>());
    
    Map<String, String> result = service.getLoginCount(userNoSet, date);
    assertEquals("0", result.get("user1")); // 验证默认值
}
```

#### ✅ 场景3：空输入
```java
/**
 * 测试目的：验证输入空用户集合时能正常处理并返回空Map
 */
@Test
void getLoginCount_EmptyUserSet() {
    Set<String> userNoSet = new HashSet<>();
    
    Map<String, String> result = service.getLoginCount(userNoSet, date);
    assertTrue(result.isEmpty());
}
```

---

### 2.2 计数类方法（两个必测场景）

#### ✅ 场景1：正常计数
```java
/**
 * 测试目的：验证正常情况下能正确返回统计数量
 */
@Test
void countByCode() {
    when(mapper.countByCode(anyString(), any(Date.class))).thenReturn(10L);
    
    String result = service.countByCode("user1", date);
    assertEquals("10", result);
}
```

#### ✅ 场景2：Null返回
```java
/**
 * 测试目的：验证数据库返回null时能正确返回默认值0
 */
@Test
void countByCode_NullResult() {
    when(mapper.countByCode(anyString(), any(Date.class))).thenReturn(null);
    
    String result = service.countByCode("user1", date);
    assertEquals("0", result);
}
```

---

### 2.3 百分比/精度计算（五个必测场景）

#### ✅ 场景1：正常计算
```java
/**
 * 测试目的：验证正常百分比计算结果的正确性
 */
@Test
void getPercentage_Normal() {
    when(mapper.getTotalCount(any())).thenReturn(100L);
    when(mapper.getMatchCount(any(), anyLong())).thenReturn(20L);

    String result = service.getPercentage(date, time);
    assertEquals("20.00%", result);
}
```

#### ✅ 场景2：四舍五入（1/3）
```java
/**
 * 测试目的：验证1/3这类无限循环小数能正确四舍五入到33.33%
 */
@Test
void getPercentage_BigDecimalRounding() {
    when(mapper.getTotalCount(any())).thenReturn(3L);
    when(mapper.getMatchCount(any(), anyLong())).thenReturn(1L);

    String result = service.getPercentage(date, time);
    assertEquals("33.33%", result);
}
```

#### ✅ 场景3：四舍五入（1/7）
```java
/**
 * 测试目的：验证1/7这类无限循环小数能正确四舍五入到14.29%
 */
@Test
void getPercentage_DifferentDecimalRounding() {
    when(mapper.getTotalCount(any())).thenReturn(7L);
    when(mapper.getMatchCount(any(), anyLong())).thenReturn(1L);

    String result = service.getPercentage(date, time);
    assertEquals("14.29%", result);
}
```

#### ✅ 场景4：除零处理
```java
/**
 * 测试目的：验证总数为0时能正确处理除零情况，返回0.00%
 */
@Test
void getPercentage_ZeroTotalTasks() {
    when(mapper.getTotalCount(any())).thenReturn(0L);

    String result = service.getPercentage(date, time);
    assertEquals("0.00%", result);
}
```

#### ✅ 场景5：Null参数
```java
/**
 * 测试目的：验证输入参数为null时能正确处理并返回默认值
 */
@Test
void getPercentage_NullFastestTime() {
    String result = service.getPercentage(date, null);
    assertEquals("0.00%", result);
}
```

---

### 2.4 时间格式化方法（五个必测场景）

#### ✅ 场景1：正常时间
```java
/**
 * 测试目的：验证正常时间值能正确格式化为可读字符串
 */
@Test
void getFastestProcessingTimeStr_WithValidTime() {
    Long time = 3600000L; // 1小时
    String result = service.getFastestProcessingTimeStr(time);
    assertNotNull(result);
    assertTrue(result.contains("小时") || result.contains(":"));
}
```

#### ✅ 场景2：大时间值
```java
/**
 * 测试目的：验证超过1天的大时间值能正确格式化并包含"天"
 */
@Test
void getFastestProcessingTimeStr_WithLargeTime() {
    Long time = 90061000L; // 1天1小时1分1秒
    String result = service.getFastestProcessingTimeStr(time);
    assertTrue(result.contains("天") || result.contains(":"));
}
```

#### ✅ 场景3：Null值
```java
/**
 * 测试目的：验证输入为null时能正确处理并返回null
 */
@Test
void getFastestProcessingTimeStr_NullValue() {
    String result = service.getFastestProcessingTimeStr(null);
    assertNull(result);
}
```

#### ✅ 场景4：零值
```java
/**
 * 测试目的：验证输入为0时能正确处理并返回null
 */
@Test
void getFastestProcessingTimeStr_ZeroValue() {
    String result = service.getFastestProcessingTimeStr(0L);
    assertNull(result);
}
```

#### ✅ 场景5：负值
```java
/**
 * 测试目的：验证输入负数时能正确处理并返回null
 */
@Test
void getFastestProcessingTimeStr_NegativeTime() {
    String result = service.getFastestProcessingTimeStr(-1000L);
    assertNull(result);
}
```

---

### 2.5 分区逻辑测试（两个必测场景）

#### ✅ 场景1：超过999个元素（多分区）
```java
/**
 * 测试目的：验证超过999个用户时的分区查询逻辑，确保数据正确合并
 */
@Test
void getLoginCount_PartitionLogic() {
    // Arrange - 构造超过999个用户
    Set<String> userNoSet = new HashSet<>();
    for (int i = 0; i < 1000; i++) {
        userNoSet.add("user" + i);
    }
    Date currYearDate = new Date();
    
    // Mock返回部分用户的数据
    List<UserCountVo> countList = new ArrayList<>();
    for (int i = 0; i < 500; i++) {
        UserCountVo vo = new UserCountVo();
        vo.setUserNo("user" + i);
        vo.setCount(i + 1);
        countList.add(vo);
    }
    when(mapper.getLoginCount(anyList(), any(Date.class))).thenReturn(countList);

    // Act
    Map<String, String> result = service.getLoginCount(userNoSet, currYearDate);
    
    // Assert
    assertEquals(userNoSet.size(), result.size());
    
    // 验证有数据的用户
    for (int i = 0; i < 500; i++) {
        assertEquals(String.valueOf(i + 1), result.get("user" + i));
    }
    
    // 验证无数据的用户默认值为0
    for (int i = 500; i < 1000; i++) {
        assertEquals("0", result.get("user" + i));
    }
}
```

#### ✅ 场景2：少于999个元素（单分区）
```java
/**
 * 测试目的：验证少于999个用户时不进行分区，直接查询
 */
@Test
void getLoginCount_SinglePartition() {
    Set<String> userNoSet = new HashSet<>();
    for (int i = 0; i < 500; i++) {
        userNoSet.add("user" + i);
    }
    // ... 验证逻辑与多分区类似
}
```

---

# 测试数据构建模式

## 1. 简单对象构建（推荐）

### 使用setter方法（强制）
```java
// ✅ 正确方式
UserCountVo vo = new UserCountVo();
vo.setUserNo("user1");
vo.setCount(5);

// ❌ 错误方式（构造函数可能不存在）
// UserCountVo vo = new UserCountVo("user1", 5);
```

---

## 2. 辅助方法模式

### 单个对象构建
```java
private UserCountVo createUserCountVo(String userNo, int count) {
    UserCountVo vo = new UserCountVo();
    vo.setUserNo(userNo);
    vo.setCount(count);
    return vo;
}
```

### 批量对象构建
```java
private List<UserCountVo> createUserCountList(int size) {
    List<UserCountVo> list = new ArrayList<>();
    for (int i = 0; i < size; i++) {
        list.add(createUserCountVo("user" + i, i + 1));
    }
    return list;
}
```

---

# 断言使用指南

## 1. 基本断言

```java
// 非空断言
assertNotNull(result);

// 相等断言
assertEquals("expected", result);
assertEquals(expected, actual);

// 空断言
assertNull(result);

// 布尔断言
assertTrue(result.isEmpty());
assertTrue(result.contains("天"));
assertFalse(result.isEmpty());

// 集合大小断言
assertEquals(10, result.size());
```

---

## 2. 异常断言（JUnit 5）

```java
// 验证抛出指定异常
assertThrows(NullPointerException.class, () -> {
    service.method(null);
});

// 验证异常消息
IllegalArgumentException exception = assertThrows(
    IllegalArgumentException.class,
    () -> service.processData(invalidInput)
);
assertTrue(exception.getMessage().contains("参数不能为空"));

// 验证不抛出异常
assertDoesNotThrow(() -> {
    service.processData(validInput);
});
```

---

## 3. 可选的描述性断言

```java
// 添加失败消息（推荐用于复杂断言）
assertEquals("5", result.get("user1"), "用户1的登录次数应为5");
assertTrue(result.contains("天"), "时间字符串应包含'天'");
```

---

## 附录：规范更新说明

本文档定义的是**稳定的测试规范**，如无重大变更，不应频繁修改。

- 如需添加新的测试模式或最佳实践，应经过团队评审
- 已废弃的规范应标记为 `[已废弃]` 而不是直接删除
- 每次更新应记录变更日期和原因
