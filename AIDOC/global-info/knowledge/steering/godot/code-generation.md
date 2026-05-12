# AI 代码生成规范（完整版）

> 本文件定义 AI 为 Godot 4.x 项目生成 GDScript/Shader 代码时必须遵循的完整规范。
> `.kiro/steering/core.md` 中的"AI 代码生成规则"是本文件的摘要版。

---

## 一、GDScript 4.x 编码规范

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
## - Collision Layer: 3 (PlayerHitbox)
## - Collision Mask: 2 (EnemyHurtbox)
@export var collision_layer: int = 4  # Layer 3
@export var collision_mask: int = 2   # Layer 2
```

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
| 2 | 单文件超过 200 行 | 违反单一职责 | 拆分为多个组件脚本 |
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
- [ ] 单文件不超过 200 行
- [ ] 函数不超过 30 行（建议）
- [ ] 无循环依赖
- [ ] 错误处理完整（异步操作、文件 I/O）
- [ ] 碰撞层在注释中说明含义
- [ ] 复杂系统有场景树结构文档
