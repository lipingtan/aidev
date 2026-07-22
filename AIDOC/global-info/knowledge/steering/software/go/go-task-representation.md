# Go 任务可执行表示规范（Task Representation）

> 核心理念：任务不是"描述"，而是"可执行的表示"。AI 只会执行"被明确表达的部分"。

---

## 一、为什么需要这个规范

模糊的任务描述会导致：
- 同一任务多次执行结果不一致
- 边界逻辑遗漏或处理不完整
- 修改了不该修改的文件
- 破坏了已有功能

每个 Task 必须包含四要素（Task IR）：
- **Scope（边界）**：做什么、不做什么
- **Constraints（约束）**：必须遵循什么规则
- **Acceptance（验证标准）**：怎么算完成
- **自测（SelfTest）**：必须编写哪些测试用例来证明功能正确

---

## 二、复杂度分级标准

### 判定维度：影响范围 × 逻辑复杂度

| 复杂度 | 影响范围 | 逻辑复杂度 | 典型场景 | 四要素要求 |
|--------|----------|------------|----------|-----------|
| **低** | 单文件或单方法 | 无条件分支、无状态变化 | 改字段、加常量、加注释、改配置 | 只需 Acceptance |
| **中** | 单模块（≤5 个文件） | 有条件分支或状态变化，逻辑自包含 | 新增单表 CRUD、新增筛选条件、增加校验 | Scope + Acceptance + 自测 |
| **高** | 跨模块（>5 个文件）或跨服务 | 复杂状态流转、跨模块依赖、事务 | 多租户改造、认证流程、级联操作 | 完整 Scope + Constraints + Acceptance + 自测 |

### 判定规则（按顺序匹配，命中即停）

1. 涉及跨模块调用或新增公共中间件 → **高**
2. 涉及数据库 DDL 变更（新增表、修改表结构） → **高**
3. 涉及 ≥3 个文件的联动修改 → **中**
4. 有条件分支或业务规则判断 → **中**
5. 其余 → **低**

---

## 三、四要素定义

### Scope（边界）

| 字段 | 说明 | 示例 |
|------|------|------|
| 涉及文件 | 本任务需要新增或修改的文件列表 | `app/game/apis/game.go`, `app/game/service/game.go` |
| 涉及模块 | 本任务影响的业务模块 | `game`, `common/middleware` |
| 不触碰 | 明确不应修改的文件或逻辑 | 现有 auth 中间件、其他域的 service |

### Constraints（约束）

| 类型 | 说明 | 示例 |
|------|------|------|
| 技术约束 | 必须使用的技术方案 | 必须使用 GORM 的 Scopes 做数据过滤 |
| 规范约束 | 必须遵循的编码规范 | 错误必须用 `fmt.Errorf` 包装并返回 |
| 行为约束 | 不能破坏的已有逻辑 | tenant_id=0 时不过滤（超级管理员） |
| 兼容约束 | 接口向后兼容要求 | 现有接口签名不变，只新增可选参数 |

### Acceptance（验证标准）

| 类型 | 说明 | 示例 |
|------|------|------|
| 功能验证 | 核心功能是否正确 | 游戏列表按 tenant_id 正确过滤 |
| 编译验证 | 代码是否通过编译 | `go build ./...` 零错误 |
| 静态分析 | 是否通过 vet 检查 | `go vet ./...` 无警告 |
| 回归验证 | 现有功能是否不受影响 | 【回归】现有登录接口正常（RG-1） |

### 自测（SelfTest）

> 自测是四要素中的**强制字段**（中/高复杂度任务必须包含，低复杂度任务可省略）。
> 执行阶段必须逐条编写对应的 Test 函数，且测试必须符合 `go-testing.md` 自测真实性强制规范。

| 字段 | 说明 | 示例 |
|------|------|------|
| 测试文件 | 测试代码所在文件 | `common/plugin/manager_test.go` |
| 测试点 | 需要验证的具体场景（每行一个） | 启动插件后 ActionRegistry 包含该插件 actions |

**格式：**
```markdown
**自测:**
- 测试文件: `xxx_test.go`
- ST: {场景描述} → {预期结果}
- ST: {场景描述} → {预期结果}
```

**规则：**
- 每个 Acceptance 中的功能验证项（AC）至少有一个对应的自测点（ST）
- 回归验证项（RG-N）如涉及可自动化验证的逻辑，也需要对应自测点
- 自测点必须可映射为一个具体的 `Test{Xxx}` 函数
- 禁止写空测试/只检查 nil error/Mock 自循环验证（详见 `go-testing.md`）
- `go test ./涉及的包/...` 全部通过才算完成

---

## 四、格式规范

### 高复杂度任务

```markdown
### Task N: 实现多租户数据隔离

**复杂度**: 高

**Scope（边界）:**
- 涉及文件:
  - `common/models/tenant.go`（新增）
  - `common/middleware/tenant.go`（新增）
  - `app/game/apis/game.go`（修改）
  - `app/game/service/game.go`（修改）
  - `app/game/service/dto/game.go`（修改）
  - `app/game/router/router.go`（修改）
- 涉及模块: common, game
- 不触碰: app/admin/ 下的所有文件、handler/auth.go 的 PayloadFunc

**Constraints（约束）:**
- 必须使用 middleware.GetTenantId(c) 从 context 获取租户 ID
- tenant_id 不从请求参数读取（防止伪造）
- tenant_id=0 时不过滤数据（超级管理员权限）
- 所有业务 Model 必须嵌入 models.TenantBy
- 错误必须用 fmt.Errorf 包装并返回

**Acceptance（验证标准）:**
- AC: TenantBy mixin 定义正确，包含 gorm tag
- AC: WithTenantId 中间件从 JWT claims 正确提取 tenantId
- AC: 游戏列表查询按 tenant_id 过滤
- AC: 创建游戏时自动设置 tenant_id
- AC: go build ./... 零错误
- AC: go vet ./... 无警告
- AC: 【回归】现有登录接口不受影响（RG-1）
- AC: 【回归】admin 角色仍可查看所有数据（RG-2）

**自测:**
- 测试文件: `app/game/service/game_test.go`
- ST: tenant_id=1 查询 → 仅返回该租户数据
- ST: tenant_id=0 查询 → 返回全部数据（超管不过滤）
- ST: 创建游戏未设 tenant_id → 自动从 context 注入
- ST: 【回归】现有登录接口调用 → 响应不变（RG-1）
```

### 中复杂度任务

```markdown
### Task N: 新增游戏管理 CRUD 接口

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: app/game/apis/game.go, app/game/service/game.go, app/game/service/dto/game.go, app/game/router/router.go
- 涉及模块: game
- 不触碰: 其他域的 apis/service

**Acceptance（验证标准）:**
- AC: GET /api/v1/game 返回分页列表
- AC: GET /api/v1/game/:id 返回单条详情
- AC: POST /api/v1/game 创建成功
- AC: PUT /api/v1/game/:id 更新成功
- AC: DELETE /api/v1/game/:id 删除成功
- AC: go build ./... 零错误

**自测:**
- 测试文件: `app/game/apis/game_test.go`
- ST: GET /api/v1/game → 200 + 返回列表含分页字段
- ST: GET /api/v1/game/:id 存在 → 200 + 正确数据
- ST: GET /api/v1/game/:id 不存在 → 404
- ST: POST /api/v1/game 合法参数 → 201 + DB 有记录
- ST: POST /api/v1/game 缺必填字段 → 400
- ST: DELETE /api/v1/game/:id → DB 记录被删除
```

### 低复杂度任务

```markdown
### Task N: 给 Game model 新增 Version 字段

**复杂度**: 低

**Acceptance:**
- AC: Game struct 包含 Version 字段，类型 string，gorm tag 正确
- AC: go build ./... 零错误
```

---

## 五、执行阶段使用规则

### 执行前（必须）

1. 读取该任务的四要素（Scope/Constraints/Acceptance/自测）
2. 确认 Scope 中的文件范围和禁区
3. 确认 Constraints 中的技术和行为约束
4. 确认自测中的测试文件和测试点列表
5. 如果四要素不完整或有歧义，先向用户澄清再执行

### 执行中（必须）

1. 严格在 Scope 范围内操作，不越界修改其他文件
2. 遵循 Constraints 中的所有约束条件
3. 如果发现需要修改 Scope 之外的文件，先报告用户确认

### 执行后（必须）

1. 运行 `go build ./...` 确认零错误
2. 运行 `go vet ./...` 确认无警告
3. 逐条验证 Acceptance 中的每个标准
4. **编写自测**：按自测部分的 ST 列表编写对应的 Test 函数（遵循 `go-testing.md` 自测真实性强制规范）
5. 运行 `go test ./涉及的包/...` 确认全部通过
6. 全部通过才标记任务为已完成
7. 如有未通过项，修复后重新验证
8. 回归验证项如果无法自动验证，明确标注"需人工确认"

---

## 六、信息来源映射

四要素的信息从已有文档中提取：

| 四要素 | 主要来源 | 补充来源 |
|--------|----------|----------|
| Scope | design.md（API 设计、数据库设计） | 现有代码结构 |
| Constraints | design.md（技术方案、不变行为清单） + requirements.md（技术约束） | Go 编码规范 |
| Acceptance | requirements.md（验收标准） + design.md（不变行为清单） | 非功能性需求 |
| 自测 | Acceptance 中的功能验证项 + design.md 正确性属性 | go-testing.md 规范 |

---

## 七、与开发流程的关系

本规范不改变开发流程顺序，只增强 Task 阶段的输出质量：

```
requirements.md → design.md → tasks.md（含四要素）→ 执行
```

- Task 拆解时：从 requirements + design 中提取四要素，写入 tasks.md
- Task 执行时：按四要素执行、编写测试并验证
- 用户确认时：可调整复杂度等级和四要素内容
