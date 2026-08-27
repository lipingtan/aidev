# Godot 开发工作流补充规范

> 本文件是 `development-workflow.md`（通用版）的 Godot 引擎特定补充。
> 使用时需同时参考通用工作流规范。

---

## 一、Godot 技术轨道产出物格式

### 1.1 文件格式对应

| 通用概念 | Godot 具体格式 | 说明 |
|----------|---------------|------|
| 脚本代码 | `.gd`（GDScript） | 主要开发语言 |
| 场景文件 | `.tscn` | Godot 场景格式 |
| 数据资源 | `.tres` | Godot 资源格式 |
| 着色器 | `.gdshader` | Godot Shader Language |
| 插件 | `addons/` 目录下的 GDScript 插件 | 遵循 Godot 插件规范 |
| 引擎工程配置文件 | `project.godot` | INI 格式配置 |

### 1.2 Phase 0 技术轨道产出物（Godot 特定）

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| 目录骨架 | 文件系统 | `projects/{解决方案名}/{游戏名}/` 完整创建 |
| Godot 工程 | project.godot | 可在 Godot 编辑器中打开，渲染管线/物理引擎已配置 |
| 插件初始化 | addons/ | gd_ecs、dlc_manager 目录已创建 |
| **QualitySettings** | autoload/ + 注册项 | 多平台项目：`QualitySettings`（或等价）已加入 Autoload，顺序见 `godot-engine.md` |

---

## 二、Godot 特定 POST-CHECK 项

### 2.1 代码级 POST-CHECK（GDScript）

在通用 POST-CHECK 基础上，增加以下 Godot 特定检查：

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | GDScript 语法无错误 | 无解析错误，可被 Godot 编辑器加载 |
| 2 | 类型标注完整 | 所有函数参数、返回值、成员变量有 GDScript 类型标注 |
| 3 | class_name 声明 | 每个脚本文件有 class_name 声明 |
| 4 | @export 分组 | 导出变量使用 @export_group 分组 |
| 5 | 信号使用过去时态 | 信号名使用 snake_case 过去时态命名 |

### 2.1a 运行时 POST-CHECK

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | F5 运行无红色报错 | Output 面板无 `push_error` 输出 |
| 2 | 预期外警告为零 | 无意外 `push_warning`（已知/预期警告需在文档标注） |
| 3 | 场景切换无内存泄漏 | 切换后 Orphan Nodes 数量不增长（Debugger > Monitors） |

### 2.2 场景级 POST-CHECK

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | .tscn 可在 Godot 编辑器打开 | 场景文件格式正确，无损坏 |
| 2 | 节点类型正确 | 使用正确的 Godot 节点类型（参见 godot-engine.md 节点选择规则） |
| 3 | 脚本挂载正确 | 脚本路径正确，extends 匹配节点类型 |
| 4 | 碰撞层配置 | 碰撞层/掩码按标准分配方案配置 |

### 2.3 资源级 POST-CHECK

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | .tres 格式正确 | 资源文件可被 Godot 正确加载 |
| 2 | 资源类型匹配 | Resource 子类与数据结构一致 |
| 3 | 路径引用有效 | 资源中引用的其他资源路径存在 |
| 4 | UID 文件配对 | `.uid` 文件存在且与 `.tres` 配对（Godot 4.x 资源标识） |

### 2.4 多平台 / QualityProfile（含纹理分档）

> 项目含移动端或多画质档时，在资源集成类步骤后执行。

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | `QualitySettings` 已注册 | Autoload **早于** `DataManager` 与 `DlcManager`，顺序见 `godot-engine.md` |
| 2 | 当前档纹理路径有效 | 抽样主角色/主场景材质，无缺失纹理（无洋红） |
| 3 | 导出预设一致 | 目标平台预设与 `asset-pipeline.md` 第五节策略一致（剔除/包含 tier 目录正确） |

---

## 三、Godot 特定验收标准补充

### 3.1 Phase 0 验收补充

- [ ] `project.godot` 存在且配置正确
- [ ] 渲染管线已配置（桌面 Forward+；移动 Forward Mobile 或声明的 Compatibility 兜底预设）
- [ ] 物理引擎已配置（Godot Physics / Jolt）
- [ ] Autoload 脚本已注册且顺序正确
- [ ] 输入映射已根据游戏类型预配置
- [ ] ECS 基类已存在：`addons/gd_ecs/core/ecs_component.gd` 与 `ecs_system.gd`（或项目中等价约定）
- [ ] 工程可在 Godot 编辑器中正常打开（无报错）
- [ ] main_menu.tscn 存在（即使是空场景）
- [ ] .gitignore 包含 Godot 特定忽略项（.godot/、*.import）
- [ ] **多平台项目**：`QualitySettings`（或等价）Autoload 已配置，详见 `godot-engine.md`

### 3.2 Phase 2 验收补充

- [ ] 原型场景可在 Godot 编辑器中按 F5 启动运行
- [ ] GDScript 无运行时错误（Output 面板无红色报错）
- [ ] 场景树结构合理（无过深嵌套）
- [ ] **多平台**：至少一次导出的移动端/等价包能加载原型主场景，`QualitySettings.current_tier` 与导出意图一致（参见 `execution-protocol.md` Phase 2）

### 3.3 Phase 4 验收补充

- [ ] 所有 .tres 数据文件可在 Inspector 中正确显示
- [ ] 所有 .tscn 场景可独立运行
- [ ] Autoload 单例状态在场景切换时正确保持

---

## 四、Godot 工程目录结构

```
projects/{解决方案名}/{游戏名}/
├── project.godot
├── scenes/
│   ├── levels/
│   ├── characters/
│   ├── ui/
│   └── 2d/                     # 按需创建
├── scripts/
│   ├── core/
│   ├── gameplay/
│   ├── ui/
│   ├── narrative/
│   └── ai/
├── resources/                  # 数据资源（.tres）
├── assets/                     # 游戏资产
│   ├── characters/
│   ├── environments/
│   ├── effects/
│   └── shared/
│       ├── audio/
│       ├── ui/
│       └── fonts/
├── addons/                     # 插件
│   ├── gd_ecs/
│   ├── dlc_manager/
│   └── shared/
├── autoload/                   # 全局单例
├── dlc/                        # DLC 包
└── export/                     # 导出配置
```

---

## 五、Godot 特定的生成节奏补充

### 5.1 角色数据生成

每个角色设计完成后，生成以下 Godot 特定文件：
- `resources/characters/{角色名}.tres` — 角色属性数据
- `scripts/gameplay/{角色名}_controller.gd` — 角色专属逻辑（如有）
- `scenes/characters/{角色名}.tscn` — 角色场景

### 5.2 关卡场景生成

每个关卡设计完成后，生成：
- `scenes/levels/{关卡名}.tscn` — 关卡场景骨架
- `resources/levels/{关卡名}_config.tres` — 关卡配置数据

### 5.3 道具/技能批量生成

批量设计完成后，生成：
- `resources/items/{道具名}.tres` — 每个道具的数据资源
- `resources/skills/{技能名}.tres` — 每个技能的数据资源
