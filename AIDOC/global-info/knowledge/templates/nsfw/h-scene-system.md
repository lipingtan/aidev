# H-Scene 系统模板

> H-Scene 系统比普通对话/战斗系统复杂，本文件定义场景触发、阶段状态机、互动控制的架构。

---

## 一、系统概览

H-Scene 系统由以下子系统组成：

```
HSceneManager（单例）
  ├── TriggerEvaluator    # 场景触发条件评估
  ├── HSceneStateMachine  # 场景内阶段状态机
  ├── InteractionLayer    # 玩家互动控制（QTE/节奏条/选项）
  ├── AnimationController # 动画/姿势切换
  ├── AudioDirector       # 语音/音效阶段同步
  └── CGUnlockTracker     # 触发后解锁 CG Gallery
```

## 二、触发条件评估

触发一个 H-Scene 前，`TriggerEvaluator` 检查以下条件：

```gdscript
# h_scene_trigger.gd
class_name HSceneTrigger
extends Resource

@export var scene_id: String
@export var required_relationship_stats: Dictionary  # 关系数值阈值
@export var required_flags: Array[String]             # 剧情 flag
@export var required_clothing_state: String = ""      # 服装状态（可选）
@export var required_player_tags: Array[String] = []  # 玩家偏好标签（可选）
@export var content_tags: Array[String] = []          # 场景内容标签（用于玩家过滤开关）
@export var is_consensual: bool = true                # 是否合意（非合意场景额外校验）

func can_trigger(character: CharacterBase, player: PlayerBase) -> bool:
    # 1. 合规检查：角色必须是成年声明角色
    if not character.compliance.declared_adult:
        push_error("H-Scene 触发拒绝：角色 %s 未通过成年声明" % character.character_id)
        return false
    
    # 2. DLC 检查
    if not DLCManager.is_adult_dlc_active():
        return false
    
    # 3. 玩家内容开关检查
    for tag in content_tags:
        if PlayerPreferences.is_tag_disabled(tag):
            return false
    
    # 4. 关系数值检查
    for stat_name in required_relationship_stats:
        var threshold = required_relationship_stats[stat_name]
        if character.relationship_stats.get_stat(stat_name) < threshold:
            return false
    
    # 5. 偏好兼容性检查
    if not character.preference_flags.can_accept_scene(content_tags):
        return false
    
    return true
```

## 三、场景阶段状态机

H-Scene 内分为标准阶段，每个阶段有独立的动画、语音、互动逻辑：

```
IDLE（等待开始）
  ↓ 触发
APPROACH（接近/前置对话）
  ↓ 通过互动/对话分支
FOREPLAY（前戏阶段）
  ↓ lust 达到阈值 / 玩家推进
MAIN（主要阶段）
  ├── 循环直到 climax 条件满足
  └── 互动控制影响进展速度
CLIMAX（高潮阶段）
  ↓ endurance 耗尽 / 特定条件
AFTERMATH（余韵阶段）
  ↓ 自动或玩家跳过
END（结束，触发 CG 解锁 + 数值更新）
```

```gdscript
# h_scene_state_machine.gd
enum HScenePhase {
    IDLE, APPROACH, FOREPLAY, MAIN, CLIMAX, AFTERMATH, END
}

signal phase_changed(old_phase: HScenePhase, new_phase: HScenePhase)
signal climax_reached(character_id: String)
signal scene_ended(scene_id: String, completion_data: Dictionary)
```

## 四、互动控制层

H-Scene 支持多种互动模式，在场景配置中选择：

| 互动类型 | 说明 | 实现方式 |
|---------|------|----------|
| 纯观看（VN 模式） | 无互动，纯文字/动画播放 | 禁用 InteractionLayer |
| QTE 模式 | 特定时机按键提升体验值 | `QTEHandler` |
| 节奏条 | 维持节奏影响角色反应 | `RhythmBar` |
| 菜单选择 | 选择动作/姿势 | `ActionMenu` |
| 自由移动 | 鼠标/手柄控制互动点 | `FreeInteractionArea` |

```gdscript
# interaction_layer.gd
@export var interaction_mode: InteractionMode = InteractionMode.MENU_CHOICE

# 互动结果影响数值
func on_successful_interaction() -> void:
    active_scene.adjust_lust(+5)
    active_scene.adjust_endurance(-3)
    AudioDirector.intensify_voice()
```

## 五、姿势/动作切换

多角色场景的动画同步：

```gdscript
# animation_controller.gd
# 姿势切换：通知所有参与角色同步动画
func switch_pose(pose_id: String) -> void:
    for character in active_participants:
        var anim_name = "%s_%s" % [pose_id, character.role]  # "standing_cowgirl_receiver"
        character.animation_player.play(anim_name)
    
    # 多角色 IK 同步（骨骼锚点对齐）
    IKSynchronizer.align_participants(active_participants, pose_id)
```

**骨骼锚点规范**：每个姿势定义若干"连接锚点"（`IK anchor`），多角色动画通过锚点对齐实现骨骼同步。

## 六、服装状态机

```gdscript
# clothing_state_machine.gd
enum ClothingLayer {
    FULLY_DRESSED,    # 完整着装
    OUTER_REMOVED,    # 外衣脱去
    INNER_ONLY,       # 仅内衣
    PARTIALLY_EXPOSED,# 部分暴露（扯开/移位）
    FULLY_EXPOSED     # 全裸
}

signal clothing_changed(layer: ClothingLayer)

# 战斗破损（衣服 HP 系统）
@export var clothing_hp: float = 100.0

func on_hit_received(damage: float) -> void:
    clothing_hp -= damage
    if clothing_hp <= 0:
        advance_to_next_layer()
```

## 七、马赛克/审查层

在 H-Scene 渲染层上叠加实时马赛克 Shader，通过 DLC/地区设置控制：

```gdscript
# censorship_manager.gd（单例）
func update_censorship_state() -> void:
    var use_mosaic = _should_use_mosaic()
    $MosaicShaderLayer.visible = use_mosaic

func _should_use_mosaic() -> bool:
    if DLCManager.is_uncensored_patch_active():
        return false
    if OS.get_locale().begins_with("ja"):
        return true  # 日本地区强制打码
    return GameSettings.get("use_mosaic", false)
```

马赛克 Shader 参考：`engines/godot/rendering-pipeline.md` 的 Shader 扩展节。
