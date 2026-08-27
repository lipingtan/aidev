# AI 代码生成规范（完整版）

> 本文件定义 AI 为 Godot 4.7+ 项目生成 GDScript/Shader 代码时必须遵循的完整规范。
> Godot 4.5 新增 `@abstract` 注解，基类应使用此注解防止直接实例化。

---

## 一、GDScript 4.5+ 编码规范

### 1.1 命名规范

| 类型 | 风格 | 示例 |
|------|------|------|
| 类名 | PascalCase | `CharacterController`, `DamageSystem` |
| 函数名 | snake_case | `get_health()`, `apply_damage()` |
| 变量名 | snake_case | `max_speed`, `current_hp` |
| 常量 | UPPER_SNAKE_CASE | `MAX_LEVEL`, `GRAVITY` |
| 信号名 | snake_case | `health_changed`, `died` |
| 枚举名 | PascalCase | `DamageType`, `ItemType` |
| 枚举值 | UPPER_SNAKE_CASE | `PHYSICAL`, `FIRE` |
| 私有成员 | 前缀下划线 | `_internal_state`, `_cache` |
| 文件名 | snake_case.gd | `character_controller.gd` |

### 1.2 类型标注

```gdscript
# ✅ 正确：完整类型标注
var max_hp: float = 100.0
var items: Array[ItemData] = []
var stats: Dictionary = {}

func calculate_damage(base: float, multiplier: float) -> float:
    return base * multiplier

# ❌ 错误：缺少类型标注
var max_hp = 100.0
func calculate_damage(base, multiplier):
    return base * multiplier
```

### 1.3 文件组织顺序

```gdscript
class_name ClassName extends BaseClass
## 文件用途说明

# 1. 信号声明
signal health_changed(old_value: float, new_value: float)

# 2. 枚举定义
enum State { IDLE, RUNNING, ATTACKING }

# 3. 常量
const MAX_SPEED: float = 10.0

# 4. 导出变量（@export）
@export var speed: float = 5.0

# 5. 公开变量
var current_state: State = State.IDLE

# 6. 私有变量
var _internal_timer: float = 0.0

# 7. 生命周期函数（_ready, _process, _physics_process）
func _ready() -> void:
    pass

# 8. 公开函数
func take_damage(amount: float) -> void:
    pass

# 9. 私有函数
func _update_state() -> void:
    pass
```

---

## 二、文件头注释模板

### 标准模板

```gdscript
class_name DamageSystem extends EcsSystem
## 伤害计算系统
##
## 负责处理所有收到 DamageEvent 的 Entity 的伤害计算，
## 包括防御减伤、元素克制、暴击判定、格挡检查。
##
## 依赖 Component：
## - CharacterStats: 提供攻防数值
## - CombatState: 提供无敌帧/格挡状态
## - DamageEvent: 伤害事件数据（处理后移除）
##
## 依赖系统：
## - EventBus: 发送 damage_dealt 事件
## - ObjectPool: 获取伤害数字实例
```

### 简化模板（简单工具类）

```gdscript
class_name MathUtils extends RefCounted
## 数学工具函数集合
```

---

## 三、碰撞层分配方案

### 标准分配

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

### 代码中的使用

```gdscript
# 在注释中说明碰撞层含义
## 碰撞配置：
## - Collision Layer: 3 (PlayerHitbox) → bitmask = 4
## - Collision Mask: 2 (EnemyHurtbox) → bitmask = 2
@export var collision_layer: int = 1 << 2  # Layer 3 的 bitmask
@export var collision_mask: int = 1 << 1   # Layer 2 的 bitmask

# 或使用十进制值（注释说明含义）
@export var collision_layer: int = 4   # bitmask for Layer 3 (PlayerHitbox)
@export var collision_mask: int = 2    # bitmask for Layer 2 (EnemyHurtbox)
```

> **Layer 与 Bitmask 关系**：Layer N 的 bitmask = `1 << (N - 1)` = `2^(N-1)`。
> Layer 1 = 1, Layer 2 = 2, Layer 3 = 4, Layer 4 = 8...

---

## 四、Shader 注释规范

```glsl
shader_type spatial;
// 溶解效果着色器
// GPU 开销等级：中
// 主要开销来源：noise 纹理采样 + discard 操作
// 适用场景：角色死亡溶解、物体消失

// === Uniform 参数 ===
uniform float dissolve_amount : hint_range(0.0, 1.0) = 0.0;  // 溶解进度（0=完整，1=完全消失）
uniform sampler2D noise_texture;  // 噪声纹理（建议 256x256 Simplex Noise）
uniform vec4 edge_color : source_color = vec4(1.0, 0.5, 0.0, 1.0);  // 溶解边缘颜色
uniform float edge_width : hint_range(0.0, 0.2) = 0.05;  // 边缘发光宽度

void fragment() {
    float noise = texture(noise_texture, UV).r;
    if (noise < dissolve_amount) {
        discard;
    }
    // 边缘发光
    float edge = smoothstep(dissolve_amount, dissolve_amount + edge_width, noise);
    EMISSION = edge_color.rgb * (1.0 - edge) * 2.0;
}
```

---

## 五、场景树结构说明文档模板

当系统涉及 3 个以上脚本文件协作时，必须生成此文档：

```markdown
# [系统名] 场景树结构

## 节点层级

```
RootNode (Node3D)
├── CharacterBody3D [character_controller.gd]
│   ├── CollisionShape3D
│   ├── MeshInstance3D
│   ├── AnimationPlayer
│   ├── StateMachine [state_machine.gd]
│   │   ├── IdleState [idle_state.gd]
│   │   ├── RunState [run_state.gd]
│   │   └── AttackState [attack_state.gd]
│   ├── Hurtbox (Area3D) [hurtbox.gd]
│   └── HitboxPivot (Node3D)
│       └── Hitbox (Area3D) [hitbox.gd]
```

## 脚本职责

| 脚本 | 挂载节点 | 职责 |
|------|----------|------|
| character_controller.gd | CharacterBody3D | 移动、重力、输入响应 |
| state_machine.gd | StateMachine | 状态管理和切换 |
| idle_state.gd | IdleState | 待机状态逻辑 |

## 信号连接

| 发送方 | 信号 | 接收方 | 方法 |
|--------|------|--------|------|
| Hurtbox | area_entered | character_controller | _on_hurtbox_hit |
| AnimationPlayer | animation_finished | state_machine | _on_animation_finished |
```

---

## 六、异步操作错误处理模板

```gdscript
## 异步加载资源（带错误处理）
func _load_resource_async(path: String) -> Resource:
    var loader := ResourceLoader.load_threaded_request(path)
    if loader != OK:
        push_error("资源加载请求失败: %s" % path)
        return null
    
    while ResourceLoader.load_threaded_get_status(path) == ResourceLoader.THREAD_LOAD_IN_PROGRESS:
        await get_tree().process_frame
    
    var status := ResourceLoader.load_threaded_get_status(path)
    if status != ResourceLoader.THREAD_LOAD_LOADED:
        push_error("资源加载失败: %s, 状态: %d" % [path, status])
        return null
    
    return ResourceLoader.load_threaded_get(path)


## 文件 I/O（带错误处理）
func _save_to_file(path: String, data: Dictionary) -> bool:
    var file := FileAccess.open(path, FileAccess.WRITE)
    if not file:
        push_error("无法打开文件: %s, 错误: %s" % [path, FileAccess.get_open_error()])
        return false
    
    file.store_string(JSON.stringify(data, "\t"))
    file.close()
    return true
```

---

## 七、导出变量规范

### 何时使用 @export

| 场景 | 使用 @export | 说明 |
|------|:---:|------|
| 策划可调参数 | ✅ | 速度、伤害倍率、冷却时间 |
| 资源引用 | ✅ | 纹理、音效、场景 |
| 节点路径 | ❌ | 使用 @onready + get_node 代替 |
| 内部状态 | ❌ | 运行时计算的值不导出 |
| 调试开关 | ✅ | 开发期调试用参数 |

### 分组规范

```gdscript
@export_group("移动参数")
@export var move_speed: float = 5.0
@export var jump_force: float = 8.0
@export var gravity_scale: float = 1.0

@export_group("战斗参数")
@export var attack_damage: float = 10.0
@export var attack_range: float = 2.0

@export_group("调试", "debug_")
@export var debug_invincible: bool = false
@export var debug_infinite_mp: bool = false
```

---

## 八、信号使用规范

### 何时用信号 vs 直接调用

| 场景 | 选择 | 理由 |
|------|------|------|
| 通知多个监听者 | 信号 | 一对多，解耦 |
| 跨系统通信 | EventBus | 完全解耦，不需要引用 |
| 父节点通知子节点 | 直接调用 | 父知道子的存在 |
| 子节点通知父节点 | 信号 | 子不应依赖父的具体类型 |
| 请求返回值 | 直接调用 | 信号无法返回值 |
| UI 响应数据变化 | 信号 | UI 不应被游戏逻辑直接引用 |

### 信号命名规范

```gdscript
# ✅ 过去时态（表示已发生的事件）
signal damage_taken(amount: float)
signal level_up_completed(new_level: int)
signal item_picked_up(item: ItemData)

# ❌ 命令式（信号不是命令）
signal take_damage(amount: float)
signal level_up(level: int)
```

---

## 九、禁止事项清单

| # | 禁止项 | 原因 | 替代方案 |
|---|--------|------|----------|
| 1 | `var x = value` 无类型标注 | 降低可读性和 IDE 支持 | `var x: Type = value` |
| 2 | 单文件超过 300 行（小文件目标上限） | 违反单一职责 | 拆分为多个组件脚本 |
| 2a | **单文件行数超过 800 行（强制上限）** | 可维护性极差 | 必须按设计模式拆分子类/组件，复杂逻辑才允许到 800 行 |
| 2b | **单文件大小超过 30KB（强制上限）** | 可读性/维护性差 | 常规文件目标 ≤10KB，复杂文件上限 30KB，超出必须拆分 |
| 3 | 英文注释 | 违反语言规范 | 使用中文注释 |
| 4 | 硬编码魔法数字 | 难以维护和调整 | 定义为常量或 @export |
| 5 | _process 中执行事件驱动逻辑 | 浪费性能 | 使用信号或 EventBus |
| 6 | 循环依赖 | 导致加载失败 | 通过信号或中间层解耦 |
| 7 | 全局变量（非 Autoload） | 难以追踪和测试 | 使用 Autoload 单例 |
| 8 | 深层继承（>3层） | 难以理解和维护 | 使用组合模式 |
| 9 | 在 _ready 中硬编码节点路径字符串 | 重构时容易遗漏 | 使用 @onready + 常量路径 |
| 10 | 直接修改其他节点的私有变量 | 破坏封装 | 通过公开方法或信号 |

---

## 十、代码审查检查清单

生成代码后的自检清单：

- [ ] 文件头注释完整（class_name、用途、依赖）
- [ ] 所有类型标注完整
- [ ] 所有注释为中文
- [ ] 信号在类顶部集中声明
- [ ] @export 变量有分组
- [ ] 无硬编码数字（使用常量或 @export）
- [ ] 单文件不超过 300 行（小文件上限），极端复杂逻辑不超过 800 行；单文件不超过 10KB（小文件上限），极端复杂逻辑不超过 30KB；超限必须按设计模式拆分子类/组件
- [ ] 函数不超过 30 行（建议）
- [ ] 无循环依赖
- [ ] 错误处理完整（异步操作、文件 I/O）
- [ ] 碰撞层在注释中说明含义
- [ ] 复杂系统有场景树结构文档

---

## 十一、抽象类规范（Godot 4.5+）

### 11.1 何时使用 @abstract

| 场景 | 是否使用 | 说明 |
|------|:---:|------|
| 框架基类（EcsComponent、EcsSystem、State） | ✅ | 防止直接实例化，强制子类实现接口 |
| 有未实现方法的基类 | ✅ | 明确标记哪些方法必须被覆盖 |
| 普通工具类 | ❌ | 不需要 |
| 数据容器类 | ❌ | 通常可以直接实例化 |

### 11.2 语法规范

```gdscript
# 抽象类声明
@abstract
class_name EcsSystem extends RefCounted
## ECS System 抽象基类

## 抽象方法：子类必须覆盖，否则编译报错
## 注意：@abstract 方法必须有函数体（通常写 pass），但函数体内容会被编译器忽略
@abstract
func get_query() -> Array[StringName]:
    pass

@abstract
func process(entities: Array, delta: float) -> void:
    pass

## 非抽象方法：提供默认实现，子类可选择覆盖
func on_registered() -> void:
    pass
```

### 11.3 子类实现规范

```gdscript
# 具体子类必须实现所有 @abstract 方法
class_name DamageSystem extends EcsSystem
## 伤害计算系统

func get_query() -> Array[StringName]:
    return [&"DamageEvent", &"FinalStats", &"RuntimeStats"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        # 具体实现
        pass
```

### 11.4 注意事项

- `@abstract` 类不能被直接实例化（`EcsSystem.new()` 会报错）
- `@abstract` 方法**必须有函数体**（通常写 `pass`），但函数体内容会被编译器忽略
- 子类如果没有实现所有 `@abstract` 方法，编译时会报错
- `@abstract` 是 Godot 4.5 新增特性，4.4 及以下版本不支持


---

## 十二、类文件拆分规范

> GDScript 没有 C# 的 `partial class`，无法把一个类拆成多文件。
> Godot 的拆分思路是 **组合 > 继承**——把职责拆成不同对象类型，而非硬切同一个类。
>
> **与 ECS 架构的关系**：本规范与 `gd_ecs` 插件兼容。ECS 是组合模式的一种实现，
> 其 Component 即本规范的 Resource，System 即本规范的 RefCounted 逻辑对象。

---

### 12.1 决策树（先问"这件事是什么"，再选载体）

| 问自己这件事… | 选这个 | ECS 对应 | 典型例子 |
|---------------|--------|----------|----------|
| 要 `_process/_physics_process`、要在编辑器里拖来拖去？ | **子节点组件** | — | Movement、Hitbox、CameraRig |
| 是配置/数据，要 Inspector 调、要 `.tres` 存盘、多角色共用？ | **Resource** | `EcsComponent` | 属性表、技能定义、物品 |
| 需要被 ECS System 批量查询和处理？ | **EcsComponent** | ✅ | BaseStats、RuntimeStats、DamageEvent |
| 有运行时状态，但不需进场景树、也不需存成资源？ | **RefCounted** | — | 冷却器、战斗结算上下文 |
| 无状态、纯计算、到处调用？ | **静态工具类** | — | 伤害公式、向量工具 |
| 全局唯一、跨场景活着？ | **Autoload** | `EcsWorld` | EventBus、SaveManager |
| 只和当前脚本强绑定的小类型？ | **内部类** | — | 解析结果、小型 DTO |
| 行为会随状态大变？ | **独立状态机** | — | 移动态、战斗态 |
| 需要批量处理一类 Entity？ | **EcsSystem** | ✅ | DamageSystem、BuffSystem |

**一句话**：数据用 Resource（或 EcsComponent），行为用组件节点（或 EcsSystem），瞬时状态用 RefCounted，全局用 Autoload，计算用 static。

---

### 12.2 ECS 架构 vs 传统组件模式：何时选哪个？

| 场景 | 推荐模式 | 原因 |
|------|----------|------|
| 属性/Buff/战斗状态等需批量处理的数据 | **ECS** | System 可统一调度，解耦 |
| 移动/碰撞/动画等需 `_physics_process` 的行为 | **传统组件节点** | 需要场景树生命周期 |
| 纯 UI、菜单、编辑器工具 | **传统组件节点** | 不需 ECS 开销 |
| 技能定义、物品模板等静态配置 | **Resource（非 EcsComponent）** | 不需挂载到 Entity |
| 角色运行时的属性实例 | **EcsComponent** | 需要被 System 查询 |
| 相机控制、IK 约束 | **传统组件节点** | 需要直接操作 Godot 节点 |

**混用规则**：同一个角色可以同时有：
- `EcsEntity3D`（挂载 EcsComponent：BaseStats、RuntimeStats、CombatState）
- 传统子节点组件（MovementComponent、HitboxArea、AnimationController）

```
Player (EcsEntity3D)
├── MovementComponent (Node)      ← 传统组件，处理 _physics_process
├── Hitbox (Area3D)               ← 传统组件，处理碰撞信号
└── [EcsComponents]               ← 数据容器，被 EcsSystem 批量处理
    ├── BaseStatsComponent
    ├── RuntimeStatsComponent
    └── CombatStateComponent
```

---

### 12.3 方法 A：子节点组件（有生命周期/可视化的行为）

**适用**：Movement、Hitbox、CameraRig、表情控制器、IK 约束、动画状态机。

**规范**：
1. 组件做成独立小场景（`health_component.tscn`），可复用、可继承、可在编辑器调参。
2. 组件之间不要互相 `$"../Combat"` 乱抓。由宿主注入，或组件只对外发 signal。
3. 用 `%UniqueName` 或 `@export` 引用，少写脆弱路径。
4. **宿主脚本要薄**：只做装配、转发、生命周期。

```gdscript
# player.gd —— 宿主只协调，不堆业务
class_name Player extends EcsEntity3D

@onready var movement: MovementComponent = %Movement
@onready var hitbox: HitboxComponent = %Hitbox

func _ready() -> void:
    super._ready()  # EcsEntity3D 注册，子类必须调用
    hitbox.hit_received.connect(_on_hit)

func _physics_process(delta: float) -> void:
    movement.tick(self, delta)

func _on_hit(damage: float, source: Node) -> void:
    # 创建 DamageEvent，让 DamageSystem 处理
    var event := DamageEventComponent.new()
    event.amount = damage
    event.source_id = source.entity_id if source is EcsEntity3D else -1
    add_component(event)
```

```gdscript
# components/health_component.gd
class_name HealthComponent extends Node
## 生命值表现与结算入口组件

signal damaged(amount: float, source: Node)
signal died

@export var stats: CharacterStats  # Resource，不是写死在节点里
var current: float

func _ready() -> void:
    current = stats.max_hp

func apply_hit(amount: float, source: Node) -> void:
    current = maxf(current - amount, 0.0)
    damaged.emit(amount, source)
    if current <= 0.0:
        died.emit()
```

**不要用于**：物品定义、一堆冷却计时器、纯数值表 → 用 Resource 或 EcsComponent。

---

### 12.4 方法 B：EcsComponent（ECS 数据容器）

> **与 Resource 的关系**：`EcsComponent extends Resource`，是 Resource 的特化。
> 需要被 EcsSystem 查询时用 EcsComponent，否则用普通 Resource。

**适用**：BaseStats、RuntimeStats、CombatState、BuffList、SkillSet、DamageEvent。

```gdscript
# components/base_stats_component.gd
class_name BaseStatsComponent extends EcsComponent
## 基础属性 Component

@export var max_hp: float = 100.0
@export var attack: float = 10.0
@export var defense: float = 5.0

func get_component_name() -> StringName:
    return &"BaseStats"
```

**注意**：
- EcsComponent 没有 `_process/_ready`，逻辑在 EcsSystem 中。
- 运行时直接改 Component 字段（EcsComponent 是实例，不是共享 .tres）。
- DamageEvent 等一次性事件：处理后由 System 移除。

---

### 12.5 方法 C：普通 Resource（静态配置/模板）

**适用**：技能定义模板、物品模板、对话配置、关卡配置。

```gdscript
# data/skill_definition.gd
class_name SkillDefinition extends Resource
## 技能定义（模板，不是 EcsComponent）

@export var id: StringName
@export var display_name: String
@export var cooldown: float = 1.0
@export var damage_multiplier: float = 1.0
@export var animation_name: StringName
```

**与 EcsComponent 的区别**：
| 类型 | 用途 | 挂载到 Entity | 被 System 查询 |
|------|------|:---:|:---:|
| Resource | 静态模板/配置 | ❌ | ❌ |
| EcsComponent | 运行时数据实例 | ✅ | ✅ |

**注意**：运行时改 Resource 数据前先 `duplicate()`，否则所有引用同一份 .tres 的对象会一起变。

---

### 12.6 方法 D：EcsSystem（批量处理逻辑）

**适用**：伤害计算、Buff tick、属性汇总、AI 决策、经验结算。

```gdscript
# systems/damage_system.gd
class_name DamageSystem extends EcsSystem
## 伤害计算系统

func get_query() -> Array[StringName]:
    return [&"DamageEvent", &"RuntimeStats", &"CombatState"]

func process(entities: Array, delta: float) -> void:
    for entity in entities:
        var event: DamageEventComponent = entity.get_component(&"DamageEvent")
        var stats: RuntimeStatsComponent = entity.get_component(&"RuntimeStats")
        var combat: CombatStateComponent = entity.get_component(&"CombatState")
        
        # 无敌帧跳过
        if combat.is_invincible:
            entity.remove_component(&"DamageEvent")
            continue
        
        # 计算最终伤害
        var final_damage: float = maxf(event.amount - stats.defense, 1.0)
        stats.current_hp -= final_damage
        
        # 发送事件
        EventBus.damage_dealt.emit(entity.entity_id, final_damage)
        
        # 移除一次性事件
        entity.remove_component(&"DamageEvent")
```

**规范**：
- System 是无状态的（extends RefCounted），不存储 Entity 引用。
- 每个 System 只关注一个职责（单一职责原则）。
- 通过 `get_query()` 声明依赖的 Component 组合。

---

### 12.7 方法 E：RefCounted 逻辑对象

**适用**：冷却器、一次技能释放上下文、战斗结算上下文（不需要被 EcsSystem 查询时）。

```gdscript
# runtime/cooldown_tracker.gd
class_name CooldownTracker extends RefCounted
## 冷却计时器

var _until: Dictionary = {}  # StringName -> float

func start(id: StringName, duration: float, now: float) -> void:
    _until[id] = now + duration

func ready(id: StringName, now: float) -> bool:
    return now >= float(_until.get(id, 0.0))
```

**何时用 RefCounted vs EcsComponent**：
| 场景 | 选择 |
|------|------|
| 需要被 System 批量查询 | EcsComponent |
| 只被单个组件/节点内部使用 | RefCounted |
| 需要存档恢复 | EcsComponent（自带序列化） |

---

### 12.8 方法 F：静态工具类

只放无状态函数，不要把游戏状态塞进去。

```gdscript
# utils/math_utils.gd
class_name MathUtils
## 数学工具函数

static func damp(current: float, target: float, lambda: float, delta: float) -> float:
    return lerpf(current, target, 1.0 - exp(-lambda * delta))

static func angle_diff(from: float, to: float) -> float:
    return wrapf(to - from, -PI, PI)
```

**注意**：
- Godot 4.x 支持不继承任何类的纯静态脚本。
- `class_name` 后可直接 `MathUtils.damp(...)`，不必 preload。
- 工具类数量要克制，避免全局类型污染。

---

### 12.9 方法 G：内部类

同一文件内的私有类型，不进全局命名空间。适合 DTO、命令、解析结果。

```gdscript
class_name CombatComponent extends Node
## 战斗组件

class HitInfo extends RefCounted:
    var amount: float
    var source: Node
    var tags: PackedStringArray
    
    func _init(p_amount: float, p_source: Node, p_tags: PackedStringArray = []) -> void:
        amount = p_amount
        source = p_source
        tags = p_tags

signal hit_landed(info: HitInfo)
```

**限制**：
- 内部类不能好好 `@export` 到 Inspector，也不能当独立 Resource 类型用。
- 要进编辑器就单独建文件 + `class_name`。

---

### 12.10 方法 H：场景继承 + 组合

角色基场景 `actor.tscn` 带 Health/Movement 空槽；`player.tscn`、`enemy_melee.tscn` 继承它，再加自己的组件。

```
actor.tscn (EcsEntity3D)           ← 基场景
├── %Health (HealthComponent)
├── %Movement (MovementComponent)
└── CollisionShape3D

player.tscn (inherits actor.tscn)  ← 继承场景
├── PlayerInput (Node)             ← 新增组件
└── Camera3D

enemy_melee.tscn (inherits actor.tscn)
└── AiController (Node)            ← 新增组件
```

**规则**：
- 继承用在场景蓝图和真正的 is-a（Player is an Actor）。
- 功能扩展用组合（添加子节点），不要深继承脚本树。

---

### 12.11 方法 I：Autoload + EventBus

**ECS 架构中**：`EcsWorld` 是核心 Autoload。EventBus 用于 System 向外部（UI/音频）通知。

```gdscript
# autoload/event_bus.gd
extends Node
## 全局事件总线

signal damage_dealt(entity_id: int, amount: float)
signal entity_died(entity_id: int)
signal level_up(entity_id: int, new_level: int)
```

**规则**：
- EcsSystem 可以调用 EventBus 发事件。
- UI/音频系统监听 EventBus，不直接查询 EcsWorld。
- **不要把玩法逻辑写进 Autoload**。

---

### 12.12 通信规则

| 方向 | 传统组件模式 | ECS 模式 |
|------|--------------|----------|
| **向下** | 宿主注入依赖给组件 | System 通过 Query 获取 Component |
| **向上** | 组件发 signal | System 调用 EventBus |
| **横向** | 经宿主或 EventBus | 不同 System 通过 EcsWorld 共享 Entity 数据 |
| **跨层** | EventBus | EventBus |
| **查能力** | `has_method` / Group / 自定义标记 | `has_component` |

**Group 机制（官方推荐）**：
```gdscript
# 添加到组
add_to_group("enemies")

# 通过组查找
for enemy in get_tree().get_nodes_in_group("enemies"):
    if enemy.has_method("take_damage"):
        enemy.take_damage(10.0)
```

---

### 12.13 推荐目录结构（ECS + 传统混用）

```
res://
  actors/
    player/
      player.tscn                    # EcsEntity3D + 传统组件
      player.gd
    components/                      # 传统节点组件
      movement_component.tscn
      hitbox_component.gd
  addons/
    gd_ecs/
      core/                          # ECS 核心
      systems/                       # EcsSystem 按领域分
        stats/
        combat/
        ai/
  data/
    stats/                           # EcsComponent
    skills/                          # Resource 模板
    items/                           # Resource 模板
  utils/
  autoload/
```

---

### 12.14 什么时候不该拆

| 情况 | 建议 |
|------|------|
| 还不到 ~200–300 行、职责本来就是一件事 | **先别拆** |
| 拆完每个文件都要来回跳才能看懂一个功能 | 拆错边界了，按用例重切 |
| 为了"干净"给每个变量做一个 Component/节点 | 违反官方建议，Node/Component 有开销 |
| 硬要把 Movement 做成 EcsSystem | Movement 需要 `_physics_process` 和场景树，用传统组件 |

---

### 12.15 方法总结

| 方法 | 适用场景 | ECS 对应 |
|------|----------|----------|
| 子节点组件 | 有生命周期/可视化的行为 | — |
| EcsComponent | 被 System 批量处理的数据 | ✅ |
| Resource | 静态配置/模板 | — |
| EcsSystem | 批量处理逻辑 | ✅ |
| RefCounted | 内部运行时对象 | — |
| 静态工具类 | 无状态函数 | — |
| 内部类 | 文件内私有类型 | — |
| 场景继承 | 角色变体蓝图 | — |
| Autoload/EventBus | 跨场景解耦 | EcsWorld/EventBus |
| FSM | 复杂状态行为 | — |
