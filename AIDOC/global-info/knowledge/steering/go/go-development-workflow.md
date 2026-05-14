# Go 软件开发工作流程

## 角色定义

| 阶段 | 角色 |
|------|------|
| Requirement | 产品经理，擅长需求分析、用户故事、验收标准定义 |
| Design | Go 后端架构师，精通 Gin/GORM/go-admin 架构设计、API 设计、数据库设计 |
| Task | 技术项目经理，擅长任务拆解、依赖分析、并行规划 |
| 执行 | Go 高级开发工程师，精通 Go 1.21+、Gin、GORM、多租户架构 |
| Bugfix | 软件质量工程师，擅长缺陷分析、根因定位 |

---

## 1. Requirement 阶段

### 1.1 生成需求文档

**流程：**
1. 在 spec 目录下创建 `requirements_plan.md`，制定需求计划
2. 用 `[Question]` 标签提出澄清问题，创建空的 `[Answer]` 标签
3. 同时提出非功能性需求（性能、安全、多租户隔离等）
4. 不自行做出关键决策，完成计划后请求用户审查
5. 用户批准后，输出中文的 `requirements.md`

### 1.2 需求文档格式

```markdown
# 需求：{功能名称}

## 背景
{业务背景和目标}

## 用户故事
- 作为 {角色}，我希望 {功能}，以便 {价值}

## 功能需求
### FR-1: {需求名}
**描述：** ...
**验收标准：**
- WHEN {条件} THEN 系统 SHALL {行为}

## 非功能需求
- 性能：接口响应时间 < 200ms（P99）
- 安全：所有接口需 JWT 认证
- 多租户：数据按 tenant_id 隔离
```

---

## 2. Design 阶段

### 2.1 生成设计文档

**流程：**
1. 创建 `design_plan.md`，制定设计计划
2. 提出设计相关澄清问题
3. 用户批准后，输出 `design.md`

### 2.2 设计文档必须包含

```markdown
# 设计：{功能名称}

## 技术方案
### API 设计
| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | /api/v1/game | 游戏列表 | JWT + 租户 |

### 数据库设计
```sql
-- 新增/修改的表结构
ALTER TABLE game ADD COLUMN xxx ...;
```

### 核心逻辑
{流程图或伪代码}

## 不变行为清单（回归防护）
| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 现有登录流程不受影响 | /login 接口正常返回 token |
| RG-2 | 租户隔离不被破坏 | 不同租户数据不互通 |

## 正确性属性
- 所有写操作必须记录 create_by/update_by
- 业务表查询必须带 tenant_id 过滤
```

---

## 3. Task 阶段

### 3.1 任务复杂度判定

按以下顺序匹配，命中即停：
1. 涉及跨模块调用或新增接口 → **高**
2. 涉及数据库 DDL 变更 → **高**
3. 涉及 ≥3 个文件的联动修改 → **中**
4. 有条件分支或业务规则判断 → **中**
5. 其余 → **低**

### 3.2 任务三要素格式

**高复杂度任务：**
```markdown
### Task N: {任务名}

**复杂度**: 高

**Scope（边界）:**
- 涉及文件: app/game/apis/game.go, app/game/service/game.go, app/game/service/dto/game.go
- 涉及模块: game
- 不触碰: common/middleware/tenant.go、其他域的 service

**Constraints（约束）:**
- 必须使用 middleware.GetTenantId(c) 获取租户 ID，不从请求参数读取
- 查询必须加 WHERE tenant_id = ? 过滤（tenant_id=0 时不过滤）
- 错误必须用 fmt.Errorf 包装并返回，不得忽略

**Acceptance（验证标准）:**
- [ ] 接口返回正确的分页数据
- [ ] 不同租户数据互相隔离
- [ ] go build ./... 零错误
- [ ] 【回归】现有接口不受影响（RG-1）
```

**中复杂度任务：**
```markdown
### Task N: {任务名}

**复杂度**: 中

**Scope（边界）:**
- 涉及文件: app/game/models/game.go, app/game/service/dto/game.go
- 不触碰: 其他 model 文件

**Acceptance（验证标准）:**
- [ ] Model 字段定义正确，包含 GORM tag
- [ ] DTO 与 Model 字段对应
- [ ] go build ./... 零错误
```

**低复杂度任务：**
```markdown
### Task N: {任务名}

**复杂度**: 低

**Acceptance:**
- [ ] 修改内容符合需求描述
- [ ] go build ./... 零错误
```

---

## 4. Task 执行阶段

### 4.1 执行前（必须）

1. 读取任务的三要素（Scope/Constraints/Acceptance）
2. 确认涉及文件范围和禁区
3. 确认技术和行为约束
4. 三要素不完整时先澄清再执行

### 4.2 执行中（必须）

1. 严格在 Scope 范围内操作
2. 遵循 Constraints 中的所有约束
3. 需要修改 Scope 之外的文件时，先报告用户确认

### 4.3 执行后（必须）

1. 运行 `go build ./...` 确认零错误
2. 运行 `go vet ./...` 确认无警告
3. 逐条验证 Acceptance 中的每个标准
4. 全部通过才标记任务为已完成
5. 回归验证项无法自动验证时，标注"需人工确认"

---

## 5. Bugfix 流程

### 5.1 文件存放位置

```
projects/{project}/
└── .kiro/specs/Fix-{N}-{bug简述}/
    ├── bugfix.md       # bug 分析文档
    ├── design.md       # 修复设计
    └── tasks.md        # 修复任务
```

### 5.2 Bugfix 文档格式

```markdown
# Bugfix: {bug 简述}

## Bug 分析
**当前行为（缺陷）：**
WHEN {触发条件} THEN 系统 {错误表现}

**期望行为（正确）：**
WHEN {触发条件} THEN 系统 SHALL {正确行为}

**不变行为（回归防护）：**
WHEN {相关场景} THEN 系统 SHALL CONTINUE TO {保持行为}

## 根因分析
{定位到具体文件和代码行}

## 影响范围
{涉及的模块、接口、数据}
```

### 5.3 Go 常见 Bug 排查清单

**第一优先级：基础设施层**
1. **JSON 反序列化问题**
   - `time.Time` 字段格式不匹配
   - 指针类型 vs 值类型（`*int` vs `int`）
   - 枚举值不在预期范围内

2. **中间件链问题**
   - JWT 验证失败（token 格式、过期）
   - 租户 ID 未正确注入（`WithTenantId` 中间件顺序）
   - 数据库连接未初始化（`WithContextDb` 中间件）

**第二优先级：业务逻辑层**
3. **nil pointer dereference**
   - 数据库查询结果未判空
   - 接口类型断言失败
   - 未初始化的 map/slice

4. **GORM 查询问题**
   - 忘记加 `tenant_id` 过滤
   - 软删除字段未正确配置
   - 事务未正确提交/回滚

---

## 文件结构

每个需求在对应的 project_doc 目录下产出：

```
AIDOC/project_doc/{project-name}/{feature-name}/
├── requirements.md     # 需求文档
├── design.md           # 设计文档
└── tasks.md            # 任务列表
```

Bugfix 文档：
```
AIDOC/project_doc/{project-name}/Fix-{N}-{bug简述}/
├── bugfix.md
├── design.md
└── tasks.md
```

工程代码放在：
```
projects/{project-name}/    # 与游戏工程平级
```
