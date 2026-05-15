# 关系数值系统模板

> 参考 `templates/rpg/attribute-system.md` 的架构，扩展 NSFW 育成/调教类游戏的核心关系数值。

---

## 一、关系数值轴定义

以下数值与普通 RPG 属性（attack/defense）平级，使用相同的 `BaseStats / FinalStats / Modifier` 抽象：

| 数值 | 标识符 | 范围 | 说明 |
|------|--------|------|------|
| 好感度 | `affection` | 0~100 | 角色对玩家的主动情感倾向 |
| 信任 | `trust` | 0~100 | 角色对玩家行为的信赖程度 |
| 堕落度 | `corruption` | 0~100 | 角色接受 NSFW 行为的开放程度 |
| 支配 | `dominance` | -100~100 | 负值=服从倾向，正值=支配倾向 |
| 羞耻 | `shame` | 0~100 | 当前状态下的羞耻反应强度 |
| 欲望 | `lust` | 0~100 | 当前性欲水平（随时间衰减） |
| 忍耐 | `endurance` | 0~100 | 抵抗/持久能力（影响 H-Scene 阶段进展） |
| 贞操观 | `chastity` | 0~100 | 对性行为的保守程度（影响触发条件） |

## 二、数据结构（GDScript）

```gdscript
# relationship_stats.gd
class_name RelationshipStats
extends Resource

## 关系数值组件 - 挂载到 NPC/角色节点

# 基础数值（设计期配置）
@export var base_affection: float = 0.0
@export var base_trust: float = 20.0
@export var base_corruption: float = 0.0
@export var base_dominance: float = 0.0      # 负=服从，正=支配
@export var base_shame: float = 80.0
@export var base_lust: float = 0.0
@export var base_endurance: float = 50.0
@export var base_chastity: float = 80.0

# 最终数值（运行时计算，含 Modifier）
var final_affection: float
var final_trust: float
var final_corruption: float
var final_dominance: float
var final_shame: float
var final_lust: float
var final_endurance: float
var final_chastity: float

# 变更信号
signal affection_changed(old_val: float, new_val: float)
signal corruption_threshold_reached(threshold: int)
signal scene_unlock_condition_met(scene_id: String)

func calculate_finals(modifiers: Array[StatModifier]) -> void:
    # 与主属性系统共用 Modifier 计算逻辑
    pass
```

## 三、偏好标签系统（Fetish Tags）

角色和玩家双方各持一组偏好标签，用于场景准入判定：

```gdscript
# preference_flags.gd
class_name PreferenceFlags
extends Resource

## 角色偏好标签 - 用于 H-Scene 场景准入过滤

# 偏好标签（角色喜欢/接受的内容类型）
@export var liked_tags: Array[String] = []      # 例: ["vanilla", "consensual"]
@export var tolerated_tags: Array[String] = []  # 低偏好但可接受
@export var disliked_tags: Array[String] = []   # 会导致好感/信任下降

# 解锁条件（某些标签需要条件解锁）
@export var locked_tags: Dictionary = {}        # tag -> 解锁条件

func can_accept_scene(scene_tags: Array[String]) -> bool:
    for tag in scene_tags:
        if tag in disliked_tags:
            return false
        if tag in locked_tags and not _check_unlock(tag):
            return false
    return true
```

## 四、触发条件集成（与叙事系统对接）

对话/场景触发条件扩展，将关系数值作为一等公民：

```gdscript
# 场景触发条件示例（叙事系统 condition 格式）
{
    "type": "relationship_check",
    "conditions": [
        {"stat": "affection", "op": ">=", "value": 60},
        {"stat": "corruption", "op": ">=", "value": 30},
        {"stat": "chastity", "op": "<=", "value": 40},
        {"flag": "first_intimate_scene", "value": false},  # 是否第一次
        {"clothing_state": "inner_only"},                   # 服装状态
        {"player_pref_tag": "gentle"}                       # 玩家选择的偏好
    ]
}
```

## 五、数值变化规则

| 行为 | affection | trust | corruption | shame |
|------|-----------|-------|------------|-------|
| 日常关怀对话 | +2~5 | +1~3 | 0 | 0 |
| 赠送礼物（匹配偏好） | +5~15 | +2 | 0 | 0 |
| 强制/非合意行为 | -20~-50 | -30 | +10~20 | +30 |
| 温柔 H-Scene（高好感触发） | +5 | +5 | +5~15 | -5~-10 |
| 完成情感支线任务 | +10~20 | +10 | 0 | -5 |

**重要**：强制行为触发好感/信任大幅下降，系统不应将非合意内容作为数值提升路径。

## 六、数值阈值与内容解锁对照

| 关系阶段 | 解锁条件示例 | 可触发内容 |
|---------|------------|----------|
| 普通认识 | affection < 30 | 日常对话、任务合作 |
| 友好 | affection ≥ 30 | 特殊对话、部分亲密场景（全年龄） |
| 亲密 | affection ≥ 60 + trust ≥ 50 | 告白、轻度亲密场景 |
| 恋人 | affection ≥ 80 + trust ≥ 70 | 完整 H-Scene（需 DLC 启用） |
| 深度羁绊 | 全满分 | 特殊结局、隐藏 CG |
