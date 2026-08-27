# 执行协议（通用）

> 每个 Phase/步骤执行时必须遵循的标准化流程。
> 确保：1）执行前依赖就绪 2）执行中告知产出 3）执行后验证完整。
> 本文件为引擎无关的通用版本，具体引擎的补充检查项请参见对应引擎目录下的文件。

---

## 零、规范触发时机速查

> 以下规范不是"写完再查"，而是在特定节点**主动触发**。

| 规范 | 触发时机 | 触发条件 |
|------|----------|----------|
| `performance-budget.md` | Phase 0 定稿 + Phase 2 实测 + Phase 4 每新增关卡 + Phase 5 最终验证 | 始终 |
| `experience-benchmarks.md` | 设计计划（定默认值）+ Phase 2 验收 + Phase 5 调优 | 涉及操作手感的功能 |
| `patterns/README.md` | feature-development-flow 2.3 设计计划 | 每个新功能 |
| `templates/README.md` | feature-development-flow 2.3 设计计划 | 每个新功能 |
| `compliance/platform-policies.md` | Phase 5 发布前 | 始终 |

**超标处理流程**（performance-budget）：
1. 发现指标超出预算 → 记录到 `iterations/performance.md`
2. 按 `iteration-workflow.md` 第三节"性能优化流程"执行
3. 优化后重新对照预算表验证
4. 仍超标 → 评估是否调整预算（需用户确认）或降级功能

---

## 一、协议总览

```
┌─────────────────────────────────────┐
│ PRE-CHECK（执行前）                  │
│ - 检查所有输入依赖文件是否存在       │
│ - 验证前置阶段 POST-CHECK 已通过     │
│ - 告知本步骤将生成哪些文件           │
└──────────────────┬──────────────────┘
                   │ 通过
                   ▼
┌─────────────────────────────────────┐
│ EXECUTE（执行）                      │
│ - 按规范生成文件/代码                │
│ - 遵循代码生成规范                   │
│ - 每生成一个文件，标记完成           │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│ POST-CHECK（执行后）                 │
│ - 验证所有预期文件是否已生成         │
│ - 代码级验证（语法/类型/规范）       │
│ - 报告结果                          │
└─────────────────────────────────────┘
```

---

## 二、PRE-CHECK 模板

```markdown
## 🔍 PRE-CHECK: [Phase X: 步骤名称]

### 输入依赖检查

| # | 依赖文件 | 状态 |
|---|----------|------|
| 1 | `path/to/file` | ✅ 存在 / ❌ 缺失 |

### 本步骤将生成

| # | 文件 | 路径 |
|---|------|------|
| 1 | 文件名 | `完整路径` |

### 判定

- ✅ 依赖完整，可以执行
- ❌ 依赖缺失 [列出缺失项]，需要先完成 [对应步骤]
```

---

## 三、POST-CHECK 模板

```markdown
## ✅ POST-CHECK: [Phase X: 步骤名称]

### 产出验证

| # | 预期文件 | 状态 | 备注 |
|---|----------|------|------|
| 1 | `path/to/file` | ✅ / ❌ | |

### 代码规范检查（代码类产出）

| # | 检查项 | 状态 |
|---|--------|------|
| 1 | 脚本语法无错误 | ✅ / ❌ |
| 2 | 类型标注完整 | ✅ / ❌ |
| 3 | 文件头注释规范 | ✅ / ❌ |
| 4 | 单文件 ≤200 行 | ✅ / ❌ |
| 5 | 信号声明有中文注释 | ✅ / ❌ |

### 判定

- ✅ 本步骤完成，可进入下一步
- ❌ 以下文件未生成/不合规：[列出]
```

---

## 四、各 Phase 检查清单

### Phase 0: 项目初始化

**PRE-CHECK**：无（首个步骤）

**POST-CHECK**：
- [ ] 引擎工程配置文件存在
- [ ] `AIDOC/game_doc/{游戏名}/README.md` 存在
- [ ] `AIDOC/game_doc/{游戏名}/design/` 目录存在
- [ ] `AIDOC/game_doc/{游戏名}/narrative/` 目录存在（若澄清不需要叙事系统，可仅保留 `README.md` 占位说明）
- [ ] `AIDOC/game_doc/{游戏名}/iterations/` 目录存在
- [ ] `AIDOC/game_doc/{游戏名}/tracker.md` 存在
- [ ] 引擎工程目录骨架完整（场景/脚本/资源/资产/插件/全局单例）
- [ ] **[双平台]** `clarification.md` 已填写目标平台（桌面 / Android / iOS 等）与 **最低机型档位**
- [ ] **[双平台]** `design/performance-budget.md` 或 `architecture.md` 已包含 **QualityProfile 档位表**（参见 `asset-pipeline.md` 第四节 4.5）

### Phase 1: 核心设计

**PRE-CHECK**：
- [ ] Phase 0 POST-CHECK 通过
- [ ] `AIDOC/game_doc/{游戏名}/README.md` 存在

**POST-CHECK**：
- [ ] `design/gdd.md` 存在且包含：玩法循环、胜负条件、目标平台
- [ ] `design/worldview.md` 存在且包含：世界规则、地理、历史、核心冲突
- [ ] `design/architecture.md` 存在且包含：系统架构图、模块接口定义
- [ ] 核心玩法设计文档存在
- [ ] **[双平台]** `architecture.md` 含 **QualityProfile / QualitySettings** 模块说明（Autoload、解析纹理路径、`get_scalar` 约定），与 `godot-engine.md` 一致
- [ ] **[双平台]** 已定义桌面 vs 移动的 **渲染方法**（与 `performance-budget.md` §1、`godot-bootstrap.md` §1.2 一致：`forward_plus` vs `mobile` Forward Mobile；兜底 `gl_compatibility` 仅写进导出策略时声明）

### Phase 2: 原型验证

**PRE-CHECK**：
- [ ] `design/gdd.md` 存在
- [ ] `design/architecture.md` 存在

**POST-CHECK**：
- [ ] 原型场景可在引擎编辑器中启动运行
- [ ] 核心玩法循环可完成至少 1 次完整流程
- [ ] 操作手感验证报告存在
- [ ] 代码通过代码级 POST-CHECK
- [ ] **[双平台]** 已在 **至少一个移动导出目标**（真机或 CI 构建包）上进行 smoke：启动、加载主线原型场景、**纹理/显存峰值**记录在 `iterations/prototype_feedback.md` 或 `iterations/mobile_smoke.md`
- [ ] **[双平台]** 已验证 **当前档** 下 `QualityProfile` 解析的纹理路径无缺失（无粉红材质）

### Phase 3: 核心系统实现

**PRE-CHECK**：
- [ ] Phase 2 原型通过验证
- [ ] 角色设定文档存在（创意轨道产出）

**POST-CHECK**：
- [ ] 角色系统代码存在且通过代码级检查
- [ ] 战斗系统代码存在且通过代码级检查
- [ ] 叙事系统代码存在且通过代码级检查
- [ ] UI 框架代码存在且通过代码级检查
- [ ] 存档系统代码存在且通过代码级检查
- [ ] 各系统可独立运行
- [ ] 系统间集成后无阻断性错误

### Phase 4: 内容填充

**PRE-CHECK**：
- [ ] Phase 3 所有核心系统通过 POST-CHECK
- [ ] 关卡策划文档存在（创意轨道产出）

**POST-CHECK**：
- [ ] 至少 1 个完整关卡可从开始到结束正常游玩
- [ ] 敌人配置数据存在且可正常加载
- [ ] 道具数据存在且可正常使用
- [ ] 对话可正常触发和播放

### Phase 5: 打磨优化

**PRE-CHECK**：
- [ ] Phase 4 至少 1 个关卡通过验证

**POST-CHECK**：
- [ ] 帧率达到性能预算基准
- [ ] 内存使用不超过预算上限
- [ ] 无崩溃性 Bug
- [ ] 可成功导出为目标平台可执行文件
- [ ] **[双平台]** **桌面** 与 **移动** 导出预设各至少一条已成功出包；安装包体积与 **移动端 VRAM 峰值** 对照 `asset-pipeline.md` 第四节 / 第七节预算记录
- [ ] **[双平台]** **桌面档 + 移动档**（`desktop_high` 与 `mobile_high`）各完成可玩通 smoke，无 OOM、无致命掉帧区（记录于 `iterations/performance.md`）

---

## 五、代码级 POST-CHECK 清单

每个代码文件生成后必须通过：

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | 语法正确 | 脚本无语法错误 |
| 2 | 类型标注完整 | 函数参数、返回值、成员变量有类型 |
| 3 | 文件头注释 | 包含类名、用途、依赖 |
| 4 | 文件行数 | 目标 ≤300 行，强制上限 ≤800 行；超出按设计模式拆分 |
| 5 | 信号声明规范 | 类顶部集中声明，有中文注释 |
| 6 | 命名规范 | 类名 PascalCase、函数/变量 snake_case |
| 7 | 无硬编码魔法数字 | 使用常量或导出变量 |
| 8 | 无循环依赖 | A 不同时引用 B 且 B 引用 A |

---

## 六、使用规则

1. 每个 Phase 开始前必须输出 PRE-CHECK 报告
2. 每个 Phase 完成后必须输出 POST-CHECK 报告
3. PRE-CHECK 失败时中止并报告缺失项
4. POST-CHECK 失败时报告不合规项并修复
5. 用户可明确要求跳过 PRE-CHECK，但 **POST-CHECK 不可跳过**
6. 代码级 POST-CHECK 在每个脚本文件生成后立即执行
7. **POST-CHECK 通过后必须等待用户确认才能进入下一 Phase**（与 core.md「每阶段必须等待用户确认」原则一致）
