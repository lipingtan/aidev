# 游戏开发知识库 · 索引与写作原则

## 写作原则（必读）

**不重复 Godot / 通用编程的官方文档**（API、节点列表、语法细则以 [官方文档](https://docs.godotengine.org/) 为准）。

本库只收三类内容：

| 类型 | 说明 | 典型位置 |
|------|------|----------|
| **架构约束** | 与本工作台绑定的边界：Autoload 顺序、ECS+DLC、`QualitySettings`、纹理 `tier_*`、碰撞层编号、单文件行数、禁止旁路读 Base 伤害等 | `steering/godot/`、`performance-budget.md`、`godot-engine.md` |
| **Best practice** | 在 Godot 上仍算「选型」的东西：何时用 FSM+AnimationTree、伤害单边入口、存档不序列化 Node 路径等 | `templates/*`、`patterns/state-machine.md`（短文） |
| **反模式 / 陷阱** | 「看似能用但会破坏管线」的清单 | 各 `templates/*` 末节、`steering/pitfalls.md`（若维护） |

`engines/godot/*.md` 里若为「待填充」**≠ 催人写满教程**；仅当出现 **本项目独有约束**（例如：与 `demo_game` 插件的契约、导出特性标签）时再落笔，并 **链接官方文档** 而非抄书。

## 流程与门控（可依赖）

`steering/`：`pre-development-defaults.md`、`development-workflow.md`、`execution-protocol.md`、`performance-budget.md`、`asset-pipeline.md`、`godot/quality-settings-spec.md`、`godot-engine.md`。

## 系统模板（边界 + 验收 + 陷阱）

| 路径 | 性质 |
|------|------|
| `templates/rpg/*.md` | 模块职责与工程对齐，**不是** GDScript 百科 |
| `templates/action/*.md` | 同上 |
| `templates/nsfw/*.md` | NSFW 管线 + 合规挂钩 |

## 占位文件策略

| 路径 | 策略 |
|------|------|
| `engines/godot/*.md` | **默认不补**成官方教程；有工作台级约束时再写短文 |
| `patterns/*.md` | 同上；与架构冲突的设计模式才值得写 |

## 何时新增一篇知识库条目

1. 新 **门控** 或 **导出/合规** 规则进入 `steering` / `compliance`。  
2. 新 **系统模块** 与 ECS/DLC/Quality 有硬契约 → 用 `templates/` 一节描述边界。  
3. 团队反复踩同一坑 → 记入对应模板「常见陷阱」或 `pitfalls.md`。

不满足以上则 **不写**，直接链官方文档。
