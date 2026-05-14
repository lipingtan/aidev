---
inclusion: always
---

# 任务可执行表示规范（Task Representation）

> 核心理念：任务不是"描述"，而是"可执行的表示"。系统不会理解"差不多"，它只会执行"被明确表达的部分"。

---

## 一、为什么需要这个规范

当 AI Agent 执行任务时，模糊的任务描述会导致：
- 同一任务多次执行结果不一致
- 边界逻辑遗漏或处理不完整
- 需要反复补充 prompt 才能接近预期

根本原因不是 AI 能力不够，而是**任务没有被完整表达**。

本规范要求每个 Task 必须包含三要素（Task IR）：
- **Scope（边界）**：做什么、不做什么
- **Constraints（约束）**：必须遵循什么规则
- **Acceptance（验证标准）**：怎么算完成

---

## 二、复杂度分级标准

### 判定维度：影响范围 × 逻辑复杂度

| 复杂度 | 影响范围 | 逻辑复杂度 | 典型场景 | 三要素要求 |
|--------|----------|------------|----------|-----------|
| **低** | 单文件或单方法 | 无条件分支、无状态变化 | 改字段长度、加注释、调文案、加常量、加枚举值 | 只需 Acceptance |
| **中** | 单模块（≤5 个文件） | 有条件分支或状态变化，但逻辑自包含 | 新增筛选条件、增加校验规则、新增单表 CRUD、新增单个接口 | Scope + Acceptance |
| **高** | 跨模块（>5 个文件）或跨服务 | 有复杂状态流转、跨模块依赖、并发/事务 | 跨模块改造、安全加固、认证流程、级联操作、Excel 导入 | 完整 Scope + Constraints + Acceptance |

### 判定规则（按顺序匹配，命中即停）

1. 涉及跨模块调用（Feign/Api 接口变更）或跨服务通信 → **高**
2. 涉及数据库 DDL 变更（新增表、修改表结构） → **高**
3. 涉及 ≥3 个文件的联动修改 → **中**
4. 有条件分支或业务规则判断 → **中**
5. 其余 → **低**

### 判定时机与责任人

- **判定者**：AI（Task 阶段，技术项目经理角色）
- **判定时机**：生成 tasks.md 时，为每个任务标注复杂度
- **调整权**：用户在确认 tasks.md 时可调整任务复杂度等级

---

## 三、三要素定义

### Scope（边界）

明确任务的操作范围和禁区：

| 字段 | 说明 | 示例 |
|------|------|------|
| 涉及文件 | 本任务需要新增或修改的文件列表 | `XxxServiceImpl.java`, `XxxController.java` |
| 涉及模块 | 本任务影响的业务模块 | `base-center`, `common` |
| 不触碰 | 明确不应修改的文件或逻辑 | 现有 Redis 配置、其他模块的缓存逻辑 |

### Constraints（约束）

本任务执行时必须遵循的规则：

| 类型 | 说明 | 示例 |
|------|------|------|
| 技术约束 | 必须使用的技术方案或组件 | 必须使用现有 Spring Data Redis / Lettuce |
| 规范约束 | 必须遵循的命名/格式/编码规范 | 缓存 Key 命名遵循 `base:{entity}:{id}` |
| 行为约束 | 不能破坏的已有逻辑 | 缓存失效时降级查询 DB，不抛异常 |
| 兼容约束 | 接口向后兼容要求 | 现有接口签名不变，只新增参数（可选） |

### Acceptance（验证标准）

任务完成的判定条件，必须是可验证的：

| 类型 | 说明 | 示例 |
|------|------|------|
| 功能验证 | 核心功能是否正确 | 缓存预热 Runner 启动时加载数据到 Redis |
| 编译验证 | 代码是否通过编译 | getDiagnostics 零错误 |
| 回归验证 | 现有功能是否不受影响 | 现有 Redis 配置和其他模块缓存不受影响 |
| 测试验证 | 测试是否通过（如有） | mvn test -Dtest=XxxServiceTest 通过 |

---

## 四、tasks.md 中的格式规范

### 高复杂度任务

```markdown
### Task 6: 缓存服务实现

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: BaseCacheService.java, BaseCacheServiceImpl.java, BaseCacheWarmupRunner.java
- 涉及模块: base-center
- 不触碰: 现有 Redis 配置（application.yml）、user-center 的缓存逻辑

**Constraints（约束）:**
- 必须使用现有 Spring Data Redis / Lettuce 客户端，不引入新依赖
- 缓存 Key 命名遵循 `base:{entity}:{id}` 格式（参考 design.md 缓存方案章节）
- 缓存失效时降级查询数据库，不抛异常
- 预热失败不阻塞应用启动

**Acceptance（验证标准）:**
- [ ] BaseCacheService 接口定义完整（含片区/小区/楼栋三类缓存方法）
- [ ] BaseCacheServiceImpl 实现 Redis 优先 + DB 降级逻辑
- [ ] BaseCacheWarmupRunner 启动时加载数据到 Redis
- [ ] getDiagnostics 零错误
- [ ] 现有 Redis 配置和其他模块缓存不受影响（回归）
```

### 中复杂度任务

```markdown
### Task 4.1: 实现 DistrictService

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: DistrictService.java, DistrictServiceImpl.java
- 涉及模块: base-center
- 不触碰: 其他 Service 实现

**Acceptance（验证标准）:**
- [ ] CRUD 方法完整（create/update/delete/getById/page）
- [ ] 编码自动生成逻辑正确（数据库 MAX+1）
- [ ] 唯一性校验（名称+编码不重复）
- [ ] 停用时级联检查下级小区
- [ ] getDiagnostics 零错误
```

### 低复杂度任务

```markdown
### Task 3.1: 新增 BaseErrorCode 错误码枚举

**复杂度**: 低

**Acceptance:**
- [ ] 错误码范围 3000-3999，与 requirements.md 一致
- [ ] 枚举值覆盖 design.md 错误处理章节定义的所有错误码
- [ ] getDiagnostics 零错误
```

---

## 五、执行阶段使用规则

### 执行前（必须）

1. 读取该任务的三要素
2. 确认 Scope 中的文件范围和禁区
3. 确认 Constraints 中的技术和行为约束
4. 如果三要素不完整或有歧义，先向用户澄清再执行

### 执行中（必须）

1. 严格在 Scope 范围内操作，不越界修改其他文件
2. 遵循 Constraints 中的所有约束条件
3. 如果发现需要修改 Scope 之外的文件，先报告用户确认

### 执行后（必须）

1. 逐条验证 Acceptance 中的每个标准
2. 全部通过才标记任务为已完成
3. 如有未通过项，修复后重新验证
4. 回归验证项如果无法自动验证，明确标注"需人工确认"

---

## 六、信息来源映射

三要素的信息从已有文档中提取：

| 三要素 | 主要来源 | 补充来源 |
|--------|----------|----------|
| Scope | design.md（组件接口设计、代码分层） | 现有代码结构 |
| Constraints | design.md（技术方案、正确性属性） + requirements.md（技术约束） | 项目规范（java-conventions、spring-boot-patterns） |
| Acceptance | requirements.md（验收标准） + design.md（正确性属性） | 非功能性需求 |

---

## 七、与现有流程的关系

本规范不改变现有流程顺序，只增强 Task 阶段的输出质量：

```
requirements_plan.md → requirements.md → design_plan.md → design.md → tasks.md（含三要素）→ 执行
```

- Task 拆解时：AI 从 requirements + design 中提取三要素，写入 tasks.md
- Task 执行时：AI 按三要素执行并验证
- 用户确认时：可调整复杂度等级和三要素内容
