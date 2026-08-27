# 混合项目开发工作流（通用规范）

> 适用于同时包含以下三类工程的项目：
> - **Go 后端服务**（Web API / 管理端）
> - **Godot App Shell**（游戏盒子 / 应用外壳）
> - **GameModule / 游戏内部工程**
>
> 核心策略：**一套流程框架 + 按 CR 类型激活对应验收层**。
> 项目专属约束在各项目的 `AIDOC/project_doc/{项目名}/dev-workflow-override.md` 中 override。

---

## 上游规范引用

本文件**不重复**以下规范的内容，只定义路由规则和激活逻辑：

| 规范文件 | 激活场景 |
|---|---|
| `steering/software/go/go-development-workflow.md` | Go 后端 CR 的门控、文档格式、三角色 Review |
| `steering/software/go/go-task-representation.md` | Go 任务四要素 |
| `steering/software/go/go-testing.md` | Go 自测真实性规范 |
| `steering/game/feature-development-flow.md` | Godot CR 的流程框架与任务三要素 |
| `steering/game/execution-protocol.md` | Godot PRE/POST-CHECK 清单 |
| `steering/game/godot/code-generation.md` | GDScript 编码规范 |

---

## 0. AI 行为基本原则（对所有 AI 实体强制）

- **先读文件，再行动**：用户提到某个文档（需求计划、设计文档、任务列表等），必须先读取该文件获取答案，**禁止**向用户询问文件中已有的信息。
- **不问已有答案的问题**：凡能通过读取工作区文件得到的信息（字段内容、Answer、决策记录等），必须自行读取，不得让用户重复填写。
- **推断优先**：文件中有明确 Answer/决策时直接采用；文件中有推荐方案但无 Answer 时，默认采纳推荐方案并标注来源，无需再次询问用户确认。
- **每阶段必须等待用户明确确认才能进入下一阶段**：需求计划 → **[等确认]** → 需求 → **[等确认]** → 设计计划 → **[等确认]** → 设计 → **[等确认]** → 任务拆解 → **[多角色Review直到无问题]** → **[等确认]** → 执行。每产出一个阶段文档后必须停下来等用户回复，**禁止连续跨阶段自动生成**。例外：用户明确说"全部按业界最佳实践处理"时可跳过逐步确认。

---

## 1. CR 类型判定

**按顺序匹配，命中第一个即确定类型，不确定时按跨层集成 CR 处理（最严）：**

| 优先级 | 判定条件 | CR 类型 |
|---|---|---|
| 1 | 同时涉及 Godot 客户端与 Go 后端的联动（接口对接/协议变更） | **跨层集成 CR** |
| 2 | 涉及 Go 工程，或 API/DB 变更 | **Go 后端 CR** |
| 3 | 涉及 Shell 主工程（页面/Autoload/主题/Launcher/服务层） | **Shell CR** |
| 4 | 涉及游戏接入 Shell 协议（GameModule/Runner 协议实现） | **GameModule CR** |
| 5 | 涉及游戏内部玩法逻辑（不改 Shell 接口，不改后端） | **游戏内部 CR** |
| 6 | 批量数据填充/资产导入/配置调整，无逻辑变更 | **内容填充 CR** |

---

## 2. 通用流程框架

所有 CR 类型共用同一个阶段序列：

```
需求计划 → 用户确认
    ↓
需求文档 → 用户确认
    ↓
设计计划 → 用户确认
    ↓
设计文档 → 用户确认（激活层特有 Review 在此执行）
    ↓
任务拆解 → 用户确认（一致性自检在此执行）
    ↓
任务执行（激活层特有验收规则）
    ↓
集成验证
```

**简化规则：**

| 场景 | 可跳过 |
|---|---|
| 已有明确设计文档 | 需求计划 + 需求文档，直接从设计计划开始 |
| 简单数据填充 / 内容填充 CR | 设计计划，直接拆任务执行 |
| Bug 修复 | 需求 + 设计，走 §6 Bugfix 流程 |
| 用户明确说"直接做" | 跳过确认等待，但仍需产出需求卡和设计摘要 |

---

## 3. 各 CR 类型的激活层

### 3.1 Go 后端 CR

**激活内容：**
- 门控：`go-development-workflow.md` 完整阶段门控（5 个确认点）
- 设计 Review：三角色 Review（业务专家 / 产品经理 / 架构师）
- 任务格式：**四要素**（Scope / Constraints / Acceptance / 自测）— Go 特有，需编写测试用例
- 任务一致性自检：`go-development-workflow.md` §3.4 六项检查
- 执行后：`go build ./...` 零错误 + `go vet ./...` + 自测全部通过
- 产品验收：100% 完成后强制触发，含前后端接口对齐 17 项检查

---

### 3.2 Shell CR

**适用范围：** App Shell 层——页面、Autoload 单例、服务层、主题系统、背景层。

**激活内容：**
- 门控：游戏流程门控（需求确认 → 设计确认 → 执行）
- 设计确认清单（设计文档完成后执行）：
  - [ ] Autoload 单例职责无重叠，无循环依赖
  - [ ] 信号契约完整（EventBus 信号表覆盖所有跨模块通知）
  - [ ] 页面基类 on_enter/on_exit/on_resume/on_back 接口齐全
  - [ ] 服务层与页面层分层清晰，页面不含业务规则
- 任务格式：**三要素**（Scope / Constraints / Acceptance）— Godot 标准，无强制自测要求
- 执行后验证：Godot 代码级 POST-CHECK（见 §4.1）
- 集成验证（Shell CR 完成后必跑，细节由项目 override 定义）：
  - Shell 主场景可在编辑器中启动无报错
  - 核心导航流程可用（Tab 切换 / 页面 push-pop）
  - Launcher 完整生命周期可走通

**通用 Constraints（所有 Shell CR 必须遵守）：**
- 页面间禁止互相引用，跳转走 `Nav`，通知走 `EventBus`
- GameModule 只依赖 `ctx` 注入，不感知 Shell 存在

---

### 3.3 GameModule CR

**适用范围：** 将某个游戏接入 Shell 协议，实现标准接口。

**激活内容：**
- 门控：游戏流程门控
- 任务格式：**三要素**（Scope / Constraints / Acceptance）
- 执行后验证：Godot 代码级 POST-CHECK + GameModule 协议合规检查（见 §4.2）
- 集成验证：走完整生命周期（launch → boot → quit_requested → 结算）

**通用 Constraints：**
- `boot(ctx)` 必须接收并尊重 `ctx.save_dir`，存档只写自己的 save_dir
- `quit_requested(result)` 的 result 必须包含 `{score, playtime, achievements}`
- 游戏不得依赖 Shell 的任何具体类型，只依赖 ctx 注入

---

### 3.4 游戏内部 CR

**适用范围：** 游戏玩法逻辑迭代，不改 Shell 接口，不改 meta 结构。

**激活内容：**
- 门控：游戏流程门控（可用蓝图推荐 + 人类批准模式简化）
- 任务格式：**三要素**（Scope / Constraints / Acceptance）
- 执行后验证：Godot 代码级 POST-CHECK
- 集成验证：游戏可独立运行，核心玩法循环可完整走通

---

### 3.5 跨层集成 CR

**拆分规则（满足任一条件时，拆为独立 Go 后端 CR + Shell CR 分开走）：**
- 后端任务数 ≥ 3 且 Shell 任务数 ≥ 3（规模过大，并行执行降低风险）
- 后端接口尚未定义，Shell 侧有明确先后依赖（后端先定义接口，Shell 再对接）
- 两侧变更完全独立，无实时交叉验证需求

**执行顺序约定（不拆分时）：**
1. 先完成 Go 后端 CR（接口定义 + 自测通过）
2. 再执行 Shell CR（对接已稳定的接口）
3. 最后跑跨层联调检查

**激活内容 = Go 后端 CR 全套 + Shell CR 全套，并额外执行跨层联调检查：**

| # | 检查项 | 说明 |
|---|---|---|
| 1 | 协议字段名对齐 | 客户端解析字段名 === 服务端 JSON tag |
| 2 | 分页参数一致 | 客户端传参名 === 服务端接收参数名 |
| 3 | 鉴权 header 一致 | Bearer token 格式双端一致 |
| 4 | 错误可降级 | 服务端异常时客户端有降级到本地逻辑的路径 |
| 5 | 状态枚举对齐 | 客户端状态字面值 === 服务端枚举值 |
| 6 | 时间格式一致 | 日期时间字段格式双端一致（RFC3339 或 Unix timestamp） |
| 7 | 配置开关覆盖 | 所有后端功能有对应客户端配置开关，切 Mock 后客户端零改动 |

---

### 3.6 内容填充 CR

**激活内容：**
- 门控：简化（产出样本 → 用户确认格式 → 批量执行）
- 任务格式：只需 Acceptance
- 执行后批量 POST-CHECK：
  - 语法正确，无解析错误
  - 无重复 ID / gid
  - 引用的资产文件（icon / screenshots / scene）路径存在
  - meta.json 必填字段完整（id / title / category / runtime / price_model）
  - Godot 资产：`.import` 文件已生成（导入过一次），无粉红材质占位
  - 数值在设计文档定义的合理范围内（如有 balance.md）

---

## 4. 通用验收规则

### 4.1 Godot 代码级 POST-CHECK

每个 GDScript 文件生成后必须通过（详见 `execution-protocol.md` 第五节）：

| # | 检查项 |
|---|---|
| 1 | 语法正确，无语法错误 |
| 2 | 类型标注完整（参数/返回值/成员变量） |
| 3 | 文件头注释（class_name / 用途 / 依赖） |
| 4 | 文件行数：目标 ≤300 行，强制上限 ≤800 行；超出按设计模式拆分 |
| 5 | 信号在类顶部集中声明，有中文注释 |
| 6 | 命名规范（类名 PascalCase，函数/变量 snake_case） |
| 7 | 无硬编码魔法数字（用常量或 @export） |
| 8 | 无循环依赖 |
| 9 | 代码注释全部中文 |

### 4.2 GameModule 协议合规检查

| # | 检查项 | 验证方式 |
|---|---|---|
| 1 | `boot(ctx: Dictionary)` 已实现 | 代码检查 |
| 2 | `pause_game()` / `resume_game()` 已实现 | 代码检查 |
| 3 | `quit_requested(result: Dictionary)` 信号已声明并发射 | 代码检查 |
| 4 | result 包含 `score` / `playtime` / `achievements` | 代码检查 |
| 5 | 存档只写 `ctx.save_dir` 下路径，无硬编码 `user://` 绝对路径 | grep 验证 |
| 6 | meta.json 存在，关键字段完整（id/title/category/runtime） | 文件检查 |
| 7 | 可在 Shell 工程中走通完整生命周期 | 运行验证 |

---

## 5. 文档存放规则

```
AIDOC/project_doc/{项目名}/{功能名}/
├── requirements.md
├── design.md
└── tasks.md
```

**命名约定（建议，各项目可在 override 中调整）：**

| 功能类型 | 目录名前缀 | 示例 |
|---|---|---|
| Go 后端功能 | `{功能名}` | `review-api/` |
| Shell 功能 | `shell-{功能名}` | `shell-homepage/` |
| GameModule 接入 | `gamemodule-{游戏id}` | `gamemodule-tetris/` |
| 跨层联动 | `{里程碑}-{功能名}` | `m3-cloud-review/` |
| Bug 修复 | `Fix-{N}-{bug简述}` | `Fix-1-crash-on-quit/` |

---

## 6. Bugfix 流程

```
AIDOC/project_doc/{项目名}/Fix-{N}-{bug简述}/
├── bugfix.md
├── design.md
└── tasks.md
```

Bugfix 文档必须包含：
- **当前行为（缺陷）**：WHEN {条件} THEN 系统 {错误表现}
- **期望行为（正确）**：WHEN {条件} THEN 系统 SHALL {正确行为}
- **不变行为（回归防护）**：WHEN {场景} THEN 系统 SHALL CONTINUE TO {行为}
- **根因分析**：定位到具体文件和代码行
- **影响范围**：涉及的模块、接口、数据

---

## 7. 执行前强制检查（所有 CR）

1. **判定 CR 类型**（§1 路由表，不确定则按跨层集成）
2. **加载激活层规范**：§3 对应小节 + 上游规范文件；**如项目有 `dev-workflow-override.md`，必须同时读取，override 约束优先级更高**
   - **写任何计划文档（requirements_plan/design_plan）前，必须重读 `go-development-workflow.md` 附录 A/B 模板 + 取一个工作区既有 plan 实例对照格式，不得凭记忆或沿用旧 CR 的自定格式；每题结构 = [Question-N] → 业界最佳实践 → 推荐答案及理由（结合本项目现状）→ 空 [Answer-N] 填写位，用户只需确认或修改**
3. **确认文档存放路径**（§5）
4. **无设计文档时**：先走需求 → 设计阶段，不得直接拆任务

> **对所有 AI 实体约束**：本规范对主代理、子代理均具有约束力，子代理不得以"已由上级代理确认"为由跳过用户确认步骤。

---

## 8. 项目 Override 机制

每个混合项目在以下路径维护专属约束文件：

```
AIDOC/project_doc/{项目名}/dev-workflow-override.md
```

Override 文件只需包含与本通用规范**不同或额外**的内容：
- 项目特有的 CR 类型追加判定规则
- 项目特有的 Constraints（如特定渲染限制、存档路径约定）
- 项目特有的 Bugfix 高频问题列表
- 里程碑与 CR 路由速查表

**优先级：** `dev-workflow-override.md` > 本文件 > 上游通用规范
