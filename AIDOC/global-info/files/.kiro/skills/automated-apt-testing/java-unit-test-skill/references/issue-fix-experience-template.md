# Java单元测试问题修复经验模板

> **模板用途**：规范化单元测试问题修复经验的积累格式，确保AI能够快速识别、理解和应用历史经验
> 
> **更新时机**：每次单元测试问题修复后必须立即记录
> 
> **适用范围**：所有Java单元测试中遇到的可复用问题和解决方案

---

## 模板结构说明

每条经验记录必须包含以下7个核心要素：

### 1. 问题标识（Issue-ID）
- **格式**：`[类型]-[日期]-[序号]`
- **示例**：`MOCK-20260201-001`、`NPE-20260201-002`
- **用途**：唯一标识，便于引用和追溯

### 2. 问题类型（Issue-Type）
- **必选值**：
  - `MOCK` - Mock相关问题
  - `NPE` - 空指针异常
  - `EXCEPTION` - 其他异常
  - `COVERAGE` - 覆盖率问题
  - `FRAMEWORK` - 测试框架问题
  - `DEPENDENCY` - 依赖配置问题
  - `ASSERTION` - 断言失败问题
  - `REPORT` - 单元测试报告生成问题
  

### 3. 错误特征（Error-Pattern）
- **关键错误信息**：完整的异常类名或错误消息（必填）
- **触发场景**：导致错误的操作或代码模式（必填）
- **相关技术**：涉及的技术栈或API（必填）

### 4. 根本原因（Root-Cause）
- **一句话概述**：用一句话说明问题本质（必填）
- **详细分析**：深层次原因分析（选填）

### 5. 解决方案（Solution）
- **前置条件**：需要满足的条件（如依赖版本）
- **操作步骤**：清晰的步骤列表
- **代码示例**：包含错误示例❌和正确示例✅
- **关键要点**：需要特别注意的事项

### 6. 适用场景（Applicable-Scenarios）
- 明确列出该解决方案适用的所有场景
- 使用项目符号列表

### 7. 经验标签（Tags）
- 3-5个关键词标签
- 用于快速检索和分类
- **格式**：`#标签1 #标签2 #标签3`

---

## 完整模板格式

```markdown
### 【问题标识】[类型]-[日期]-[序号]

**问题类型**：[MOCK/NPE/EXCEPTION/COVERAGE/FRAMEWORK/DEPENDENCY/ASSERTION/REPORT]

#### 错误特征
- **错误信息**：
  ```
  [完整的异常类名和错误消息]
  ```
- **触发场景**：[描述导致错误的操作]
- **相关技术**：[涉及的框架/类/API]

#### 根本原因
[一句话概述问题本质]

[可选：详细原因分析]

#### 解决方案

**前置条件**：
- [需要的依赖版本]
- [需要的配置项]

**操作步骤**：
1. [步骤1]
2. [步骤2]
3. [步骤3]

**代码示例**：
```java
// ❌ 错误方式
[错误代码示例]

// ✅ 正确方式
[正确代码示例]
```

**关键要点**：
- [要点1]
- [要点2]
- [要点3]

#### 适用场景
- [场景1]
- [场景2]
- [场景3]

#### 经验标签
`#标签1 #标签2 #标签3 #标签4 #标签5`

---
```

---

## 示例1：Mockito静态方法Mock问题

### 【问题标识】MOCK-20260201-001

**问题类型**：MOCK

#### 错误特征
- **错误信息**：
  ```
  org.mockito.exceptions.base.MockitoException:
  The used MockMaker SubclassByteBuddyMockMaker does not support 
  the creation of static mocks
  ```
- **触发场景**：尝试使用Mockito Mock静态方法时
- **相关技术**：Mockito、静态方法Mock、MockedStatic

#### 根本原因
项目缺少`mockito-inline`依赖，默认的MockMaker不支持静态方法Mock

#### 解决方案

**前置条件**：
- Mockito版本 ≥ 3.4.0

**操作步骤**：
1. 在pom.xml中添加`mockito-inline`依赖
2. 使用`MockedStatic`包装静态方法调用
3. 确保测试逻辑在try-with-resources块内执行

**代码示例**：
```java
// ❌ 错误方式 - 缺少依赖会报错
@Test
void testStaticMethod() {
    mockStatic(UtilClass.class); // 报错：MockMaker不支持
}

// ✅ 正确方式 - 先添加依赖
<!-- pom.xml -->
<dependency>
    <groupId>org.mockito</groupId>
    <artifactId>mockito-inline</artifactId>
    <scope>test</scope>
</dependency>

// 测试代码
@Test
void testStaticMethod() {
    try (MockedStatic<UtilClass> mockedStatic = mockStatic(UtilClass.class)) {
        mockedStatic.when(() -> UtilClass.staticMethod(anyString()))
            .thenReturn("mocked result");
        
        String result = service.methodThatCallsStaticMethod();
        assertEquals("expected", result);
    }
}
```

**关键要点**：
- 必须使用try-with-resources确保Mock作用域正确
- 静态Mock在try块结束后自动清理
- 测试逻辑必须在try块内执行

#### 适用场景
- Mock静态工具类方法（如日期格式化、字符串处理等）
- Mock静态枚举方法（如`findByCode()`）
- Mock第三方库的静态API

#### 经验标签
`#Mockito #静态方法Mock #mockito-inline #MockedStatic #依赖配置`

---

## 示例2：ResponseVO导致StackOverflow问题

### 【问题标识】EXCEPTION-20260201-002

**问题类型**：EXCEPTION

#### 错误特征
- **错误信息**：
  ```
  java.lang.StackOverflowError
  ```
- **触发场景**：使用`when().thenReturn()`模式Mock返回`ResponseVO.success()`
- **相关技术**：Mockito、Spring工具类、ResponseVO、MessageUtils

#### 根本原因
`ResponseVO.success()`内部调用`MessageUtils.message()`依赖Spring容器，在单元测试环境触发无限递归

#### 解决方案

**前置条件**：
- 无特殊依赖要求

**操作步骤**：
1. 创建Mock的ResponseVO对象
2. Mock对象的具体方法（如`isSuccess()`、`getData()`）
3. 使用`doReturn().when()`模式替代`when().thenReturn()`

**代码示例**：
```java
// ❌ 错误方式 - 会触发StackOverflow
when(miTaskInfoInterface.upd(any())).thenReturn(ResponseVO.success());

// ✅ 正确方式1 - 使用doReturn().when()
ResponseVO<Boolean> mockResponse = mock(ResponseVO.class);
when(mockResponse.isSuccess()).thenReturn(true);
when(mockResponse.getData()).thenReturn(true);
doReturn(mockResponse).when(miTaskInfoInterface).upd(any(MiTaskInfoUpdReq.class));

// ✅ 正确方式2 - 完全避免静态方法
ResponseVO<Boolean> mockResponse = mock(ResponseVO.class);
when(mockResponse.isSuccess()).thenReturn(true);
doReturn(mockResponse).when(miTaskInfoInterface).upd(any());
```

**关键要点**：
- 使用`doReturn().when()`替代`when().thenReturn()`
- 先mock对象，再mock对象的方法
- 避免调用依赖Spring容器的静态方法

#### 适用场景
- Mock返回`ResponseVO`的接口调用
- Mock返回任何依赖Spring容器的对象
- Mock工具类（如`MessageUtils`、`SpringUtils`）

#### 经验标签
`#StackOverflow #ResponseVO #Spring依赖 #doReturn #静态方法依赖`

---

## 示例3：空指针异常-依赖链未完整Mock

### 【问题标识】NPE-20260201-003

**问题类型**：NPE

#### 错误特征
- **错误信息**：
  ```
  java.lang.NullPointerException
  ```
- **触发场景**：被测方法调用链式依赖时，中间对象为null
- **相关技术**：Mockito、依赖注入、链式调用

#### 根本原因
只Mock了顶层依赖，未Mock依赖链上的中间对象和方法返回值

#### 解决方案

**前置条件**：
- 无特殊依赖要求

**操作步骤**：
1. 分析被测方法的完整调用链
2. 为每一层依赖创建Mock对象
3. 逐层配置Mock的返回值
4. 确保调用链上无null返回

**代码示例**：
```java
// ❌ 错误方式 - 只Mock了顶层
@Mock
private FlowCustomExtFactory flowCustomExtFactory;

@Test
void test() {
    // helper为null导致NPE
    service.process(); // NPE: helper is null
}

// ✅ 正确方式 - 完整Mock依赖链
@Mock
private FlowCustomExtFactory flowCustomExtFactory;
@Mock 
private BaseFlowIdentityExtHelper flowIdentityExtHelper;

@BeforeEach
void setUp() {
    // 配置工厂返回Mock对象
    when(flowCustomExtFactory.getFlowIdentityExtHelper())
        .thenReturn(flowIdentityExtHelper);
    // 配置helper的行为
    when(flowIdentityExtHelper.doSomething()).thenReturn(result);
}
```

**关键要点**：
- 完整分析所有调用链
- 为每层依赖创建Mock对象
- 按调用顺序逐层配置返回值
- 测试前验证Mock配置完整性

#### 适用场景
- 方法内有链式调用（如`factory.getHelper().doWork()`）
- 依赖对象返回其他依赖对象
- 复杂的依赖注入场景

#### 经验标签
`#NullPointerException #依赖链 #链式调用 #Mock配置 #完整性检查`

---

## AI识别与应用指南

### AI如何识别经验
1. **通过问题类型（Issue-Type）**：快速定位问题类别
2. **通过错误特征（Error-Pattern）**：匹配具体错误信息
3. **通过经验标签（Tags）**：多维度关键词检索
4. **通过适用场景（Applicable-Scenarios）**：判断是否适用当前情况

### AI如何应用经验
1. **问题诊断阶段**：
   - 提取当前错误信息
   - 在经验库中搜索匹配的`Error-Pattern`
   - 确认`Issue-Type`和`Tags`是否相符

2. **方案选择阶段**：
   - 检查`Applicable-Scenarios`是否匹配
   - 验证`前置条件`是否满足
   - 确认技术栈一致性

3. **修复执行阶段**：
   - 严格按照`操作步骤`执行
   - 参考`代码示例`中的正确方式
   - 注意`关键要点`中的特殊事项

4. **经验积累阶段**：
   - 修复成功后立即记录新经验
   - 遵循模板的7要素结构
   - 补充到troubleshooting-guide.md的"最新经验积累"部分

### AI记录经验的流程
```
修复问题成功
    ↓
生成问题标识（Issue-ID）
    ↓
分类问题类型（Issue-Type）
    ↓
提取错误特征（Error-Pattern）
    ↓
总结根本原因（Root-Cause）
    ↓
整理解决方案（Solution）
    ↓
归纳适用场景（Applicable-Scenarios）
    ↓
生成经验标签（Tags）
    ↓
添加到troubleshooting-guide.md
    ↓
更新常见问题速查表（如果是高频问题）
```

---

## 模板使用约束

### 必须遵守的规范
1. ✅ **完整性**：7个核心要素必须全部填写
2. ✅ **准确性**：错误信息必须完整、准确
3. ✅ **可操作性**：解决方案必须有清晰的步骤和代码示例
4. ✅ **可复用性**：经验必须具有通用性，不能仅适用单一场景
5. ✅ **即时性**：问题修复后必须立即记录，不能延迟

### 禁止的行为
1. ❌ 不完整记录（缺少核心要素）
2. ❌ 模糊描述（错误信息不完整）
3. ❌ 缺少代码示例
4. ❌ 仅记录特定业务场景的经验（无通用性）
5. ❌ 延迟记录导致遗忘关键细节

---

## 模板维护说明

### 更新频率
- **每次问题修复后**：立即添加新经验
- **每周回顾**：整理高频问题，更新速查表
- **每月归档**：将经验分类归档到对应章节

### 质量标准
- **可识别性**：AI能通过错误信息快速定位
- **可理解性**：步骤清晰，无歧义
- **可操作性**：代码示例可直接使用
- **可复用性**：适用于多个类似场景

### 模板演进
- 根据实际使用情况优化模板结构
- 新增高频问题类型到`Issue-Type`枚举
- 补充通用的经验标签库

---

## 附录：常用经验标签库

### Mockito相关
`#Mockito` `#Mock配置` `#lenient` `#doReturn` `#when` `#thenReturn` `#ArgumentMatchers` `#MockedStatic` `#mockito-inline`

### 异常相关
`#NullPointerException` `#StackOverflow` `#UnnecessaryStubbingException` `#ExceptionInInitializerError` `#ClassNotFoundException`

### Spring相关
`#Spring依赖` `#ResponseVO` `#MessageUtils` `#SpringUtils` `#BeanUtil` `#ApplicationContext`

### 测试框架
`#JUnit5` `#JaCoCo` `#maven-surefire-plugin` `#测试执行` `#覆盖率`

### 依赖配置
`#pom.xml` `#依赖版本` `#依赖冲突` `#插件配置`

### 技术模式
`#静态方法Mock` `#依赖链` `#链式调用` `#参数匹配器` `#泛型Mock`
