# Godot 项目启动配置

> 本文件是 `project-bootstrap.md`（通用版）的 Godot 引擎特定补充。
> 包含 Godot 工程初始化配置、Autoload 注册、输入映射等引擎特定内容。

---

## 一、project.godot 基础配置

### 1.1 配置模板

```ini
[application]
config/name="{游戏显示名}"
config/description="{核心体验一句话}"
run/main_scene="res://scenes/ui/main_menu.tscn"
config/features=PackedStringArray("4.7")

[display]
window/size/viewport_width=1920
window/size/viewport_height=1080
window/stretch/mode="canvas_items"
window/stretch/aspect="expand"

[rendering]
renderer/rendering_method="forward_plus"

[physics]
3d/default_gravity=9.8
3d/default_linear_damp=0.1

[autoload]
EventBus="*res://addons/gd_ecs/core/event_bus.gd"
QualitySettings="*res://autoload/quality_settings.gd"
EcsWorld="*res://addons/gd_ecs/core/ecs_world.gd"
ObjectPool="*res://addons/gd_ecs/core/object_pool.gd"
DataManager="*res://addons/gd_ecs/core/data_manager.gd"
SaveManager="*res://addons/gd_ecs/systems/save/save_manager.gd"
DlcManager="*res://addons/dlc_manager/core/dlc_manager.gd"
```

### 1.2 渲染管线选择（Godot 4.x 三套后端）

| 选项 | rendering_method 值 | 适用场景 | 工作台档位 |
|------|---------------------|----------|------------|
| Forward+ | `"forward_plus"` | 3D 默认、桌面高质量 | `desktop_high` |
| Forward Mobile | `"mobile"` | 移动端首选（保留 PBR / 后效） | `mobile_high`（默认） |
| Compatibility | `"gl_compatibility"` | 兜底/Web/极老机型 | 不在工作台默认覆盖内 |

> **澄清**：`"mobile"` 是 **Forward Mobile** 渲染器，**不是** Compatibility。`performance-budget.md` 表格中"移动档渲染方法"以此为准。

### 1.3 物理引擎选择

| 选项 | 适用场景 | 配置方式 |
|------|----------|----------|
| Godot Physics | 默认，满足大多数需求 | 无需额外配置 |
| Jolt | 高精度物理需求（复杂碰撞、关节） | 启用 Jolt 插件 |

---

## 二、Autoload 注册顺序

### 2.1 标准注册顺序（与 `godot-engine.md` §七 / `quality-settings-spec.md` §1 / demo_game 一致）

| 顺序 | 名称 | 职责 | 文件路径 |
|------|------|------|----------|
| 1 | EventBus | 全局事件总线 | `res://addons/gd_ecs/core/event_bus.gd` |
| 2 | **QualitySettings** | **多平台画质档、纹理路径解析、画质标量** | `res://autoload/quality_settings.gd` |
| 3 | EcsWorld | ECS 框架核心 | `res://addons/gd_ecs/core/ecs_world.gd` |
| 4 | ObjectPool | 对象池 | `res://addons/gd_ecs/core/object_pool.gd` |
| 5 | DataManager | 数据表管理 | `res://addons/gd_ecs/core/data_manager.gd` |
| 6 | SaveManager | 存档管理（条件注册） | `res://addons/gd_ecs/systems/save/save_manager.gd` |
| 7 | DlcManager | DLC 管理 | `res://addons/dlc_manager/core/dlc_manager.gd` |

### 2.2 注册顺序原则

- 被依赖的先注册（EventBus 最先）
- **QualitySettings 必须早于 DataManager / DlcManager**：避免数据表与 DLC 在档位未知前去 `preload` 含 `tier_*` 路径的资源
- DlcManager 在 EcsWorld 之后（DLC 需要向 ECS 注册内容）
- **SaveManager 在 DlcManager 之前**：存档头常含已启用 DLC / 内容版本；若改为先 DLC 再 Save，须同步改加载顺序并在 `architecture.md` 写明理由
- 条件注册：无存档系统时可省略 SaveManager；其余顺序不变

---

## 三、输入映射预配置

根据游戏类型自动配置输入映射：

### 3.1 ARPG / 动作类

```
move_forward, move_backward, move_left, move_right
attack, heavy_attack, dodge, block
interact, inventory, pause
camera_rotate（鼠标）
lock_on
```

### 3.2 回合 RPG

```
confirm, cancel, menu
cursor_up, cursor_down, cursor_left, cursor_right
page_up, page_down
fast_forward（加速战斗）
```

### 3.3 平台跳跃

```
move_left, move_right
jump, dash, attack
interact
```

### 3.4 手柄映射（通用）

```
# 左摇杆
joy_axis_left_x, joy_axis_left_y

# 右摇杆
joy_axis_right_x, joy_axis_right_y

# 按键（Xbox 命名）
joy_button_a (确认), joy_button_b (取消)
joy_button_x (攻击), joy_button_y (道具)
joy_button_lb (格挡), joy_button_rb (闪避)
joy_button_lt (锁定), joy_button_rt (重击)
joy_button_start (暂停), joy_button_back (背包)
```

---

## 四、Godot 工程目录详细结构

```
projects/{解决方案名}/{游戏名}/
├── project.godot
├── autoload/                     # 全局单例脚本（QualitySettings 等）
│   └── quality_settings.gd       # 多平台画质档，参考 demo_game
├── scenes/
│   ├── levels/
│   ├── characters/
│   ├── ui/
│   │   └── main_menu.tscn
│   └── 2d/                       # 按需创建
├── scripts/
│   ├── core/
│   ├── gameplay/
│   ├── ui/
│   ├── narrative/                # 按需创建
│   └── ai/
├── resources/
│   ├── characters/
│   ├── items/
│   ├── skills/
│   ├── levels/
│   └── themes/
├── assets/
│   ├── characters/
│   ├── environments/
│   ├── effects/
│   ├── textures/                 # 含 tier_desktop/、tier_mobile/ 两档纹理
│   │   ├── tier_desktop/
│   │   └── tier_mobile/
│   └── shared/
│       ├── audio/
│       ├── ui/
│       └── fonts/
├── addons/
│   ├── gd_ecs/                   # ECS 框架插件
│   ├── dlc_manager/              # DLC 管理插件
│   └── shared/
├── dlc/                          # DLC 包目录
└── export/                       # 导出配置
```

---

## 五、Godot 特定的完成检查项

### 5.1 Phase 0 完成检查（Godot 补充）

```markdown
### Godot 工程检查
- [ ] project.godot 配置正确（分辨率、渲染管线、物理）
- [ ] renderer/rendering_method 与目标首档一致（`desktop_high` → `forward_plus`；`mobile_high` 见 `performance-budget.md`）
- [ ] Autoload 脚本已注册且顺序正确（**QualitySettings 早于 DataManager / DlcManager**）
- [ ] `autoload/quality_settings.gd` 已就位（可直接复制 `projects/demo_game/autoload/quality_settings.gd`）
- [ ] `assets/textures/tier_desktop/`、`assets/textures/tier_mobile/` 至少建空目录（`.gitkeep`）
- [ ] 输入映射已根据游戏类型预配置
- [ ] 工程可在 Godot 编辑器中正常打开（无报错）
- [ ] main_menu.tscn 存在（即使是空场景）
- [ ] .gitignore 包含 Godot 特定忽略项（.godot/、*.import）
```

### 5.2 结构合规性检查（Godot 特定）

```markdown
## Godot 结构合规性检查
- [ ] 所有 .gd 文件有 class_name 声明
- [ ] 所有 .gd 文件有类型标注
- [ ] 所有 .gd 文件有中文注释
- [ ] 场景按 levels/characters/ui 分类
- [ ] .tres 文件在 resources/ 下
- [ ] 工程级 Autoload 脚本在 autoload/ 下；插件提供的单例在 addons/*/ 下并在 project.godot 注册
- [ ] 插件在 addons/ 下
- [ ] 碰撞层按标准方案分配（参见 godot-engine.md）
```
