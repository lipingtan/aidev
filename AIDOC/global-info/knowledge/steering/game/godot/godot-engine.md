# Godot 引擎特定规范

> 本文件包含与 Godot 引擎绑定的特定配置和约束。
> 通用游戏开发规范请参见其他 steering 文件。

---

## 一、引擎版本与配置

| 配置项 | 默认值 |
|--------|--------|
| 引擎版本 | **Godot 4.7+**（最低要求 4.7，`@abstract` 注解需 4.5+） |
| 主要语言 | **GDScript** |
| 辅助语言 | Shader Language（视觉效果） |
| 渲染管线 | **桌面** `forward_plus`；**移动默认** `mobile`（Forward Mobile）；**兜底** `gl_compatibility`（仅用于 WebGL/极老设备导出预设，不作为 mobile_high 默认） |
| 物理引擎 | Godot Physics（默认）/ Jolt（高精度需求时） |
| 架构模式 | ECS 混合架构（Node + Component + System） |
| 插件框架 | gd_ecs（ECS）+ dlc_manager（DLC 动态挂接） |

> **最低版本要求**：Godot 4.7。框架使用了 `@abstract` 注解（4.5 新增），工具链基于 4.7.2 验证。

## 一-A. 工具路径（本地开发环境）

| 用途 | 路径 |
|------|------|
| headless 测试（console exe） | `C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64_console.exe` |
| GUI 截图 / 编辑器验证（win64 exe） | `C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64.exe` |
| 官方文档本地副本（完整，GFW 下以此为准） | `C:\data\developer\devtool\godot\godot-docs-html-stable\` |
| 类参考（class reference，1079 个类） | `C:\data\developer\devtool\godot\godot-docs-html-stable\classes\` |
| 官方最佳实践（13 篇） | `C:\data\developer\devtool\godot\godot-docs-html-stable\tutorials\best_practices\` |

**本地文档查阅规范：**
- 类参考按类名索引：文件名 = `class_` + **全小写、去下划线**的类名（`TextureRect`→`class_texturerect.html`，`ClassDB`→`class_classdb.html`，`StyleBoxFlat`→`class_styleboxflat.html`）
- 写代码前查 API：先 `grep -oE "方法名\([^)]{0,60}" classes/class_<类>.html` 确认签名/参数；枚举常量值、属性类型同理可 grep
- 设计节点结构 / 数据流 / 项目组织时，先读 `tutorials/best_practices/`（scene_organization、project_organization、scenes_versus_scripts、data_preferences、logic_preferences 等）
- 文档与实机行为冲突时以实机为准（用 `tools/api_probe.tscn` 探针核实），并把结论记入 `kb/godot-4.7-api-facts.md`

**使用规范：**
- headless 运行前必须设置 `$env:APPDATA` 指向临时隔离目录（防止污染 `user://`）
- GUI 截图前清空 `tmp_gui\Godot`（防止 trial 计数耗尽触发 Mock 弹窗）
- 每次运行前先杀孤儿 Godot 进程（脚本错误在 quit() 前会留活进程）
- 新增 `class_name` 文件后需先跑一次 `--headless --import` 重建全局类缓存

---

## 二、GDScript 编码规范

### 2.1 命名规范

| 类型 | 风格 | 示例 |
|------|------|------|
| 类名 | PascalCase | `CharacterController` |
| 函数名 | snake_case | `get_health()` |
| 变量名 | snake_case | `max_speed` |
| 常量 | UPPER_SNAKE_CASE | `MAX_LEVEL` |
| 信号名 | snake_case（过去时态） | `health_changed` |
| 枚举名 | PascalCase | `DamageType` |
| 枚举值 | UPPER_SNAKE_CASE | `PHYSICAL` |
| 私有成员 | 前缀下划线 | `_internal_state` |
| 文件名 | snake_case.gd | `character_controller.gd` |

### 2.2 类型标注

所有函数参数、返回值、成员变量必须有类型标注：

```gdscript
# 正确
var max_hp: float = 100.0
func calculate_damage(base: float, multiplier: float) -> float:
    return base * multiplier

# 错误
var max_hp = 100.0
func calculate_damage(base, multiplier):
    return base * multiplier
```

### 2.3 文件组织顺序

```gdscript
class_name ClassName extends BaseClass
## 文件用途说明

# 1. 信号声明
# 2. 枚举定义
# 3. 常量
# 4. 导出变量（@export / @export_category / @export_group）
# 5. @onready 变量
# 6. 公开变量
# 7. 私有变量
# 8. 生命周期函数（_ready, _process, _physics_process）
# 9. 公开函数
# 10. 私有函数
```

> **`@onready` 规范**：必须放在导出变量之后、公开变量之前。`@onready` 变量在 `_ready` 前初始化，用于获取子节点引用。

### 2.4 文件头注释模板

```gdscript
class_name ClassName extends BaseClass
## 简要说明该脚本的用途（一行）
##
## 依赖：
## - ComponentA: 用途说明
## - SystemB: 用途说明
```

### 2.5 信号声明规范

```gdscript
## 当生命值发生变化时触发
signal health_changed(old_value: float, new_value: float)

## 当角色死亡时触发
signal died()
```

---

## 三、QualityProfile（多平台画质档）

### 3.1 职责与实现位置

- **默认定稿**：`performance-budget.md`（档位 ID、`tier_*` 目录名、`get_scalar` 键与数值）。
- **API、Autoload、检测优先级**：`godot/quality-settings-spec.md`。
- **参考实现**：`projects/demo_game/autoload/quality_settings.gd`（新工程复制后改 `TEXTURES_ROOT` 若路径不同）。
- 提供 `resolve_texture(relative_path: String)`（相对 tier 目录的路径），**禁止**业务散落 `OS.get_name()` 拼纹理全路径。
- 提供 `get_scalar(key: StringName)`：阴影、后效、同屏角色、粒子倍率等（键表见 `performance-budget.md` 第五节）。
- **初始化顺序**：任意大批量加载纹理 / `DataManager` / `DlcManager` **之前**完成档位检测。

### 3.2 目录约定

与 `asset-pipeline.md` **第五节**一致，工程内示例：

```
assets/textures/
├── tier_desktop/
└── tier_mobile/
```

或使用同源双导入变体；代码仍通过本单例解析 **最终** `res://` 路径。

### 3.3 Shader / 材质

- 复杂 Shader 须标注 **GPU 开销等级**（见「Shader 注释规范」）；移动端档可切换 `ShaderMaterial` 或 `shader` 变体。
- **Forward Mobile（`mobile`）**：默认可用与桌面相近的 PBR；若导出预设切 **`gl_compatibility`**，须按 `asset-pipeline.md` §5.4 做 Shader/后效降级说明，且**禁止**依赖仅 Forward+ 可用的效果（除非该档明确不支持并有 UI/画质说明）。

### 3.4 与 Adult DLC

- 高清纹理 PCK 仅在 `desktop_high`（或项目定义的桌面档）下挂载；**移动端基座**不得依赖桌面专属路径。

---

## 四、碰撞层标准分配

| Layer | 名称 | 用途 |
|-------|------|------|
| 1 | PlayerHurtbox | 玩家受击区域 |
| 2 | EnemyHurtbox | 敌人受击区域 |
| 3 | PlayerHitbox | 玩家攻击区域（Mask: Layer 2） |
| 4 | EnemyHitbox | 敌人攻击区域（Mask: Layer 1） |
| 5 | Environment | 环境碰撞（地面、墙壁） |
| 6 | Interactable | 可交互物体（NPC、宝箱、门） |
| 7 | Projectile | 投射物 |
| 8 | Trigger | 触发区域（剧情触发、区域切换） |

---

## 五、节点选择规则

| 需求 | 3D 节点 | 2D 节点 |
|------|---------|---------|
| 可控角色 | CharacterBody3D | CharacterBody2D |
| 静态物体 | StaticBody3D | StaticBody2D |
| 物理物体 | RigidBody3D | RigidBody2D |
| 区域检测 | Area3D | Area2D |
| 相机 | Camera3D | Camera2D |
| 光照(点) | OmniLight3D | PointLight2D |
| 光照(方向) | DirectionalLight3D | DirectionalLight2D |
| 导航代理 | NavigationAgent3D | NavigationAgent2D |
| 粒子效果 | GPUParticles3D | GPUParticles2D |
| 音频(空间) | AudioStreamPlayer3D | AudioStreamPlayer2D |

---

## 六、Godot 工程目录规范

```
projects/{解决方案名}/{游戏名}/
├── project.godot
├── scenes/                     # 场景文件
│   ├── levels/
│   ├── characters/
│   ├── ui/
│   └── 2d/                     # 按需创建
├── scripts/                    # 脚本
│   ├── core/
│   ├── gameplay/
│   ├── ui/
│   ├── narrative/
│   └── ai/
├── resources/                  # 数据资源（.tres）
├── assets/                     # 游戏资产（按对象自包含 + shared/按类型）
├── addons/                     # 插件
├── autoload/                   # 全局单例（含 quality_settings.gd）
└── dlc/                        # DLC 包
```

---

## 七、Autoload 命名规范

Autoload 注册名**不能**与脚本的 `class_name` 相同，否则 Godot 会将其视为类名而非实例，导致实例方法调用报错。

**规则**：Autoload 脚本要么不声明 `class_name`，要么使用与注册名不同的名称（如加 `Gd` 前缀）。

| Autoload 注册名 | 脚本 class_name | 说明 |
|----------------|----------------|------|
| `EcsWorld` | `GdEcsWorld` | 加前缀区分 |
| `EventBus` | （无 class_name） | Autoload 脚本不需要 class_name |
| `DataManager` | （无 class_name） | 同上 |
| `ObjectPool` | （无 class_name） | 同上 |
| `DlcManager` | （无 class_name） | 同上 |
| `SaveManager` | （无 class_name） | 同上 |

## 八、Autoload 注册顺序

| 顺序 | 名称 | 职责 |
|------|------|------|
| 1 | EventBus | 全局事件总线 |
| 2 | **QualitySettings** | **多平台画质档、纹理路径解析、画质标量**（**必须在 DataManager / DlcManager 之前**） |
| 3 | EcsWorld | ECS 框架核心 |
| 4 | ObjectPool | 对象池 |
| 5 | DataManager | 数据表管理 |
| 6 | SaveManager | 存档管理 |
| 7 | DlcManager | DLC 管理 |
| 8 | GameManager | 游戏状态管理（可选） |

> **GameManager 建议**：默认 **不注册**。仅在 **跨场景流程状态**、**全局暂停 / 时间缩放**、**关卡流** 等与 `EcsWorld` / `DataManager` 边界反复牵扯时，再增加 **薄 Autoload**（只做事件转发与子系统组合，避免上帝类）。

**参考工程 `demo_game` 当前顺序**：EventBus → QualitySettings → EcsWorld → ObjectPool → DataManager → SaveManager → DlcManager。

---

## 九、禁止事项

- `var x = value` 不带类型标注
- 单文件超过 300 行（建议上限），800 行（强制上限）
- 使用英文注释
- 硬编码魔法数字
- 在 `_process` 中执行可用信号驱动的逻辑
- 循环依赖
- 深层继承（>3 层）
- 直接修改其他节点的私有变量
- **业务代码中直接根据 OS 类型拼接高清纹理路径**（须经过 QualitySettings）

---

## 十、Shader 注释规范

```glsl
shader_type spatial;
// 着色器用途说明
// GPU 开销等级：低/中/高
// 主要开销来源：[说明]

uniform float param_name : hint_range(0.0, 1.0) = 0.5;  // 参数用途说明
```

---

## 十一、导出变量规范

```gdscript
@export_group("移动参数")
@export var move_speed: float = 5.0
@export var jump_force: float = 8.0

@export_group("战斗参数")
@export var attack_damage: float = 10.0

@export_group("调试", "debug_")
@export var debug_invincible: bool = false
```

使用规则：
- 策划可调参数 → 使用 @export
- 资源引用 → 使用 @export
- 内部运行时状态 → 不导出
- 使用 `@export_category` 做一级分类（如"战斗"、"移动"）
- 使用 `@export_group` 做二级分组（如"基础参数"、"高级参数"）

> **碰撞层分配**：详见 `code-generation.md` 第三节。
