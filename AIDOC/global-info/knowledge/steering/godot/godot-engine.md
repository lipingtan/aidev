# Godot 引擎特定规范

> 本文件包含与 Godot 引擎绑定的特定配置和约束。
> 通用游戏开发规范请参见其他 steering 文件。

---

## 一、引擎版本与配置

| 配置项 | 默认值 |
|--------|--------|
| 引擎版本 | **Godot 4.5+**（最低要求 4.5） |
| 主要语言 | **GDScript** |
| 辅助语言 | Shader Language（视觉效果） |
| 渲染管线 | Forward+（3D 默认）/ Compatibility（低端适配时） |
| 物理引擎 | Godot Physics（默认）/ Jolt（高精度需求时） |
| 架构模式 | ECS 混合架构（Node + Component + System） |
| 插件框架 | gd_ecs（ECS）+ dlc_manager（DLC 动态挂接） |

> **最低版本要求**：Godot 4.5。框架使用了 `@abstract` 注解（4.5 新增），不兼容 4.4 及以下版本。

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
# 4. 导出变量（@export）
# 5. 公开变量
# 6. 私有变量
# 7. 生命周期函数（_ready, _process, _physics_process）
# 8. 公开函数
# 9. 私有函数
```

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

## 三、碰撞层标准分配

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

## 四、节点选择规则

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

## 五、Godot 工程目录规范

```
projects/{游戏名}/
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
├── autoload/                   # 全局单例
└── dlc/                        # DLC 包
```

---

## 六、Autoload 注册顺序

| 顺序 | 名称 | 职责 |
|------|------|------|
| 1 | EventBus | 全局事件总线 |
| 2 | DataManager | 数据表管理 |
| 3 | EcsWorld | ECS 框架核心 |
| 4 | ObjectPool | 对象池 |
| 5 | SaveManager | 存档管理 |
| 6 | DlcManager | DLC 管理 |
| 7 | GameManager | 游戏状态管理 |

---

## 七、禁止事项

- `var x = value` 不带类型标注
- 单文件超过 200 行
- 使用英文注释
- 硬编码魔法数字
- 在 `_process` 中执行可用信号驱动的逻辑
- 循环依赖
- 深层继承（>3 层）
- 直接修改其他节点的私有变量

---

## 八、Shader 注释规范

```glsl
shader_type spatial;
// 着色器用途说明
// GPU 开销等级：低/中/高
// 主要开销来源：[说明]

uniform float param_name : hint_range(0.0, 1.0) = 0.5;  // 参数用途说明
```

---

## 九、导出变量规范

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
- 使用 @export_group 分组
