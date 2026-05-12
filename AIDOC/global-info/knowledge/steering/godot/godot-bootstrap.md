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
renderer/rendering_method="forward_plus"  # 或根据选择调整
textures/vram_compression/import_etc2_astc=true  # 移动端需要时

[physics]
3d/default_gravity=9.8
3d/default_linear_damp=0.1

[input]
# 基础输入映射（根据游戏类型预配置）

[autoload]
GameManager="*res://autoload/game_manager.gd"
EventBus="*res://autoload/event_bus.gd"
SaveManager="*res://autoload/save_manager.gd"  # 如需存档
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
| 1 | EventBus | 全局事件总线 | `res://autoload/event_bus.gd` |
| 2 | DataManager | 数据表管理 | `res://autoload/data_manager.gd` |
| 3 | EcsWorld | ECS 框架核心 | `res://autoload/ecs_world.gd` |
| 4 | ObjectPool | 对象池 | `res://autoload/object_pool.gd` |
| 5 | SaveManager | 存档管理 | `res://autoload/save_manager.gd` |
| 6 | DlcManager | DLC 管理 | `res://autoload/dlc_manager.gd` |
| 7 | GameManager | 游戏状态管理 | `res://autoload/game_manager.gd` |

### 2.2 自动生成的基础脚本

| 文件 | 用途 | 内容 |
|------|------|------|
| autoload/game_manager.gd | 全局游戏状态管理 | 游戏状态枚举 + 状态切换接口 |
| autoload/event_bus.gd | 全局事件总线 | 信号集中声明 + 发射辅助函数 |
| autoload/save_manager.gd | 存档管理 | 存/读/删接口骨架 |
| scripts/core/base_component.gd | ECS Component 基类 | 空基类 + 类型标注模板 |
| scripts/core/base_system.gd | ECS System 基类 | _process 调度框架 |

### 2.3 注册顺序原则

- 被依赖的先注册（EventBus 最先，因为其他系统都可能用到）
- GameManager 最后注册（依赖其他所有系统）
- 条件注册：SaveManager（需要存档时）、DlcManager（需要 DLC 时）

---

## 三、输入映射预配置

根据游戏类型自动配置输入映射：

### 3.1 ARPG / 动作类

```
move_forward, move_backward, move_left, move_right
attack, heavy_attack, dodge, block
interact, inventory, pause
camera_rotate (鼠标)
lock_on
```

### 3.2 回合 RPG

```
confirm, cancel, menu
cursor_up, cursor_down, cursor_left, cursor_right
page_up, page_down
fast_forward (加速战斗)
```

### 3.3 平台跳跃

```
move_left, move_right
jump, dash, attack
interact
```

### 3.4 输入映射配置规则

- 键盘 + 手柄同时配置
- 使用 Godot 的 InputMap 系统
- 动作名使用 snake_case
- 在 project.godot 的 [input] 节中定义

---

## 四、Godot 工程目录详细结构

```
godot_projects/{游戏名}/
├── project.godot                 # 工程配置文件
├── scenes/
│   ├── levels/                   # 关卡场景
│   │   └── README.md
│   ├── characters/               # 角色场景
│   │   └── README.md
│   ├── ui/                       # UI 场景
│   │   ├── main_menu.tscn
│   │   └── README.md
│   └── 2d/                       # [如主维度含2D]
├── scripts/
│   ├── core/                     # 核心系统
│   │   ├── base_component.gd
│   │   └── base_system.gd
│   ├── gameplay/                 # 玩法逻辑
│   ├── ui/                       # UI 逻辑
│   ├── narrative/                # [如需叙事系统]
│   └── ai/                       # AI 行为
├── resources/
│   ├── characters/               # 角色数据（.tres）
│   ├── items/                    # 道具数据（.tres）
│   ├── levels/                   # 关卡配置（.tres）
│   └── themes/                   # UI 主题（.tres）
├── assets/
│   ├── characters/               # 角色资产（模型/贴图/动画）
│   ├── environments/             # 环境资产
│   ├── effects/                  # 特效资产
│   └── shared/
│       ├── audio/                # 音频
│       ├── ui/                   # UI 素材
│       └── fonts/                # 字体
├── addons/
│   ├── gd_ecs/                   # ECS 插件
│   ├── dlc_manager/              # [如需DLC]
│   └── shared/                   # 共享工具插件
├── autoload/                     # 全局单例脚本
│   ├── game_manager.gd
│   ├── event_bus.gd
│   └── save_manager.gd
├── dlc/                          # [如需DLC] DLC 包目录
└── export/                       # 导出配置
```

### 4.1 条件目录创建规则（Godot 特定）

| 条件 | 创建的额外目录/文件 |
|------|---------------------|
| 主维度 = 3D | scenes/levels/（3D 结构）、assets/characters/（3D 模型） |
| 主维度 = 2D | scenes/2d/、assets/sprites/ |
| 需要叙事系统 | scripts/narrative/ |
| 需要 DLC | dlc/、addons/dlc_manager/ |
| 联机模式 ≠ 单机 | scripts/network/、autoload/network_manager.gd |
| 有已购素材包 | assets_source/{游戏名}/3d_packages/{包名}/ |

---

## 五、Godot 特定的完成检查项

### 5.1 Phase 0 完成检查（Godot 补充）

在通用 Phase 0 检查清单基础上，增加以下 Godot 特定检查：

```markdown
### Godot 工程检查
- [ ] project.godot 配置正确（分辨率、渲染管线、物理）
- [ ] renderer/rendering_method 与选择一致
- [ ] Autoload 脚本已注册且顺序正确
- [ ] 输入映射已根据游戏类型预配置
- [ ] base_component.gd 和 base_system.gd 已生成
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
- [ ] 全局单例在 autoload/ 下
- [ ] 插件在 addons/ 下
- [ ] 碰撞层按标准方案分配（参见 godot-engine.md）
```
