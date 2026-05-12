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
config/features=PackedStringArray("4.x")

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
DataManager="*res://addons/gd_ecs/core/data_manager.gd"
EcsWorld="*res://addons/gd_ecs/core/ecs_world.gd"
ObjectPool="*res://addons/gd_ecs/core/object_pool.gd"
DlcManager="*res://addons/dlc_manager/core/dlc_manager.gd"
```

### 1.2 渲染管线选择

| 选项 | 适用场景 | 配置值 |
|------|----------|--------|
| Forward+ | 3D 默认，高质量渲染 | `"forward_plus"` |
| Mobile | 移动端优化 | `"mobile"` |
| Compatibility | 低端设备/Web | `"gl_compatibility"` |

### 1.3 物理引擎选择

| 选项 | 适用场景 | 配置方式 |
|------|----------|----------|
| Godot Physics | 默认，满足大多数需求 | 无需额外配置 |
| Jolt | 高精度物理需求（复杂碰撞、关节） | 启用 Jolt 插件 |

---

## 二、Autoload 注册顺序

### 2.1 标准注册顺序

| 顺序 | 名称 | 职责 | 文件路径 |
|------|------|------|----------|
| 1 | EventBus | 全局事件总线 | `res://addons/gd_ecs/core/event_bus.gd` |
| 2 | DataManager | 数据表管理 | `res://addons/gd_ecs/core/data_manager.gd` |
| 3 | EcsWorld | ECS 框架核心 | `res://addons/gd_ecs/core/ecs_world.gd` |
| 4 | ObjectPool | 对象池 | `res://addons/gd_ecs/core/object_pool.gd` |
| 5 | DlcManager | DLC 管理 | `res://addons/dlc_manager/core/dlc_manager.gd` |

### 2.2 注册顺序原则

- 被依赖的先注册（EventBus 最先，因为其他系统都可能用到）
- DlcManager 在 EcsWorld 之后（DLC 需要向 ECS 注册内容）
- 条件注册：SaveManager（需要存档时）

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

---

## 四、Godot 工程目录详细结构

```
projects/{游戏名}/
├── project.godot
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
- [ ] renderer/rendering_method 与选择一致
- [ ] Autoload 脚本已注册且顺序正确
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
- [ ] 全局单例在 addons/gd_ecs/core/ 下
- [ ] 插件在 addons/ 下
- [ ] 碰撞层按标准方案分配（参见 godot-engine.md）
```
