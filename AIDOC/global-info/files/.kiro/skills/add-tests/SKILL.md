---
name: add-tests
description: 自动生成单元测试，覆盖核心业务逻辑、边界条件和异常场景
version: 1.0.0
---

# Add Tests 技能

## 使用场景

- 新功能开发完成后补充单元测试
- 修复 Bug 后添加回归测试
- 重构前为现有代码补充测试保护网
- 提升代码覆盖率时

## 测试规范

### 后端测试（JUnit 5 + Mockito）

**命名规范**
- 测试类名：`被测类名 + Test`，如 `UserServiceTest`
- 测试方法名：`方法名_场景描述_期望结果`，如 `getUserById_userNotFound_throwException`

**结构规范（AAA 模式）**
```java
@Test
void methodName_scenario_expectedResult() {
    // Arrange - 准备数据和 Mock
    
    // Act - 执行被测方法
    
    // Assert - 验证结果
}
```

**覆盖要求**
- 正常路径：主流程至少一个测试
- 边界条件：null 入参、空集合、边界值
- 异常场景：业务异常、数据不存在、权限不足
- 每个 if/else 分支都应有对应测试

**Mock 规范**
- 使用 `@ExtendWith(MockitoExtension.class)` + `@Mock` / `@InjectMocks`
- 只 Mock 外部依赖（Mapper、外部服务），不 Mock 被测类本身
- 验证关键交互：`verify(mock).method(args)`

### 前端测试（Vitest + Vue Test Utils）

**命名规范**
- 测试文件：`组件名.test.ts` 或 `composable名.test.ts`
- describe 块：组件/函数名
- it 块：具体行为描述，如 `it('should emit update event when button clicked')`

**覆盖要求**
- 组件渲染：默认 Props 下正常渲染
- Props 变化：不同 Props 值下的渲染差异
- 用户交互：点击、输入等事件触发
- Emit 验证：事件是否正确触发及参数
- Composable：响应式状态变化、异步操作

## 生成测试的步骤

1. 分析被测代码，识别所有分支和边界
2. 列出测试用例清单（场景 + 期望结果）
3. 生成测试代码，确保每个用例独立
4. 检查 Mock 是否完整，断言是否准确

## 输出格式

先输出测试用例清单，再输出完整测试代码：

```
## 测试用例清单
- [ ] 正常场景：...
- [ ] 边界条件：...
- [ ] 异常场景：...

## 测试代码
[完整测试文件内容]
```
