extends Node
## Demo 场景控制器
##
## 验证 ECS 框架和 DLC 系统的主要功能：
## 1. EcsEntity3D — 3D 物理角色 Entity
## 2. EcsEntityNode — 组合模式（宝箱）
## 3. Component 挂载/读写
## 4. System 批量处理（属性计算、伤害）
## 5. 状态机 + ECS 通信
## 6. EventBus 事件订阅
## 7. DLC 动态加载/卸载

## 场景中的 Entity 节点引用
@onready var player: EcsEntity3D = $Player
@onready var enemy: EcsEntity3D = $Enemy
@onready var chest: EcsEntityNode = $Chest/EcsEntityNode

## UI 标签
@onready var status_label: Label = $UI/StatusLabel
@onready var log_label: RichTextLabel = $UI/LogLabel

## 日志缓冲
var _log_lines: Array[String] = []
const MAX_LOG_LINES: int = 20

## DLC 是否已加载
var _dlc_loaded: bool = false


func _ready() -> void:
	_setup_entities()
	_setup_systems()
	_subscribe_events()
	_log("=== ECS + DLC Demo 启动 ===")
	_log("按键说明：")
	_log("  [A] 对敌人造成伤害")
	_log("  [S] 拾取宝箱物品")
	_log("  [D] 加载 Demo DLC（毒素扩展）")
	_log("  [F] 卸载 Demo DLC")
	_log("  [G] 对玩家施加毒素（需先加载DLC）")
	_log("  [H] 玩家升级")


func _unhandled_input(event: InputEvent) -> void:
	if not event is InputEventKey or not event.pressed:
		return
	match event.keycode:
		KEY_A: _test_damage()
		KEY_S: _test_chest_pickup()
		KEY_D: _test_dlc_load()
		KEY_F: _test_dlc_unload()
		KEY_G: _test_poison()
		KEY_H: _test_level_up()


# ─────────────────────────────────────────────
# 初始化
# ─────────────────────────────────────────────

func _setup_entities() -> void:
	## 验证点 1：EcsEntity3D — 玩家
	_setup_player()

	## 验证点 1：EcsEntity3D — 敌人
	_setup_enemy()

	## 验证点 2：EcsEntityNode — 宝箱（组合模式）
	_setup_chest()

	_log("✅ 所有 Entity 初始化完成，共 %d 个" % EcsWorld.get_entity_count())


func _setup_player() -> void:
	## 验证点 3：Component 挂载
	var base := BaseStatsComponent.new()
	base.max_hp = 100.0
	base.attack = 15.0
	base.defense = 8.0
	base.speed = 5.0
	player.add_component(base)

	var final_stats := FinalStatsComponent.new()
	player.add_component(final_stats)

	var runtime := RuntimeStatsComponent.new()
	runtime.current_hp = 100.0
	player.add_component(runtime)

	var tag := DemoTagComponent.new()
	tag.tag = DemoTagComponent.Tag.PLAYER
	tag.display_name = "玩家"
	player.add_component(tag)

	var exp_comp := ExperienceComponent.new()
	exp_comp.level = 1
	exp_comp.current_exp = 0.0
	var growth := GrowthProfile.new()
	growth.base_exp = 50.0
	exp_comp.growth_profile = growth
	player.add_component(exp_comp)

	var inventory := ItemContainerComponent.new()
	inventory.container_id = &"player_backpack"
	inventory.capacity = 10
	player.add_component(inventory)

	_log("  玩家 Entity ID: %d" % player.entity_id)


func _setup_enemy() -> void:
	var base := BaseStatsComponent.new()
	base.max_hp = 50.0
	base.attack = 8.0
	base.defense = 3.0
	enemy.add_component(base)

	var final_stats := FinalStatsComponent.new()
	enemy.add_component(final_stats)

	var runtime := RuntimeStatsComponent.new()
	runtime.current_hp = 50.0
	enemy.add_component(runtime)

	var tag := DemoTagComponent.new()
	tag.tag = DemoTagComponent.Tag.ENEMY
	tag.display_name = "史莱姆"
	enemy.add_component(tag)

	_log("  敌人 Entity ID: %d" % enemy.entity_id)


func _setup_chest() -> void:
	## 验证点 2：EcsEntityNode 组合模式
	var tag := DemoTagComponent.new()
	tag.tag = DemoTagComponent.Tag.CHEST
	tag.display_name = "宝箱"
	chest.add_component(tag)

	var inventory := ItemContainerComponent.new()
	inventory.container_id = &"chest_01"
	inventory.capacity = 5
	inventory.is_readonly = false
	chest.add_component(inventory)

	_log("  宝箱 Entity ID: %d（EcsEntityNode 组合模式）" % chest.entity_id)


func _setup_systems() -> void:
	## 验证点 4：注册 System
	EcsWorld.register_system(StatsCalculationSystem.new())
	EcsWorld.register_system(DamageSystem.new())
	EcsWorld.register_system(LevelUpSystem.new())
	EcsWorld.register_system(DemoLogSystem.new())
	_log("✅ 已注册 4 个 System")


func _subscribe_events() -> void:
	## 验证点 6：EventBus 事件订阅
	EventBus.subscribe(&"damage_dealt", _on_damage_dealt)
	EventBus.subscribe(&"player_leveled_up", _on_player_leveled_up)
	EventBus.subscribe(&"poison_tick", _on_poison_tick)
	_log("✅ EventBus 事件订阅完成")


# ─────────────────────────────────────────────
# 测试操作
# ─────────────────────────────────────────────

func _test_damage() -> void:
	## 验证点 4：DamageSystem 批量处理
	_log("\n[A] 玩家攻击敌人...")
	var player_stats: FinalStatsComponent = player.get_component(&"FinalStats")
	var dmg_event := DamageEventComponent.new()
	dmg_event.damage_amount = player_stats.attack * 1.5
	dmg_event.damage_type = DamageEventComponent.DamageType.PHYSICAL
	dmg_event.attacker_id = player.entity_id
	enemy.add_component(dmg_event)
	_log("  → 挂载 DamageEvent，等待 DamageSystem 处理...")


func _test_chest_pickup() -> void:
	## 验证点 3：Component 读写
	_log("\n[S] 检查宝箱...")
	var chest_inv: ItemContainerComponent = chest.get_component(&"ItemContainer")
	if chest_inv:
		_log("  宝箱容量: %d，当前物品数: %d" % [chest_inv.capacity, chest_inv.slots.size()])
		_log("  ✅ EcsEntityNode 的 Component 读取正常")
	else:
		_log("  ❌ 宝箱 Component 读取失败")


func _test_dlc_load() -> void:
	## 验证点 7：DLC 动态加载
	if _dlc_loaded:
		_log("\n[D] Demo DLC 已加载，跳过")
		return
	_log("\n[D] 加载 Demo DLC（毒素扩展）...")
	var success: bool = DlcManager.load_dlc("demo_dlc")
	if success:
		_dlc_loaded = true
		_log("  ✅ DLC 加载成功")
		_log("  已加载 DLC: %s" % str(DlcManager.get_loaded_dlcs()))
		# 验证 DLC Component 是否注册到 EcsWorld
		var poison_script = EcsWorld._component_registry.get(&"Poison")
		if poison_script:
			_log("  ✅ PoisonComponent 已注册到 EcsWorld")
		# 验证玩家是否被注入了 Poison Component
		if player.has_component(&"Poison"):
			_log("  ✅ 玩家 Entity 已注入 PoisonComponent（角色扩展验证通过）")
	else:
		_log("  ❌ DLC 加载失败")


func _test_dlc_unload() -> void:
	## 验证点 7：DLC 动态卸载
	if not _dlc_loaded:
		_log("\n[F] Demo DLC 未加载，跳过")
		return
	_log("\n[F] 卸载 Demo DLC...")
	var success: bool = DlcManager.unload_dlc("demo_dlc")
	if success:
		_dlc_loaded = false
		_log("  ✅ DLC 卸载成功")
		# 验证 Component 已从 EcsWorld 注销
		var poison_script = EcsWorld._component_registry.get(&"Poison")
		if not poison_script:
			_log("  ✅ PoisonComponent 已从 EcsWorld 注销")
		# 验证玩家的 Poison Component 已移除
		if not player.has_component(&"Poison"):
			_log("  ✅ 玩家 Entity 的 PoisonComponent 已回退移除")
	else:
		_log("  ❌ DLC 卸载失败")


func _test_poison() -> void:
	## 验证点 7：DLC System 功能
	if not _dlc_loaded:
		_log("\n[G] 请先按 [D] 加载 DLC")
		return
	if not player.has_component(&"Poison"):
		_log("\n[G] 玩家没有 PoisonComponent，请先加载 DLC")
		return
	_log("\n[G] 对玩家施加毒素（10秒，每秒5伤害）...")
	var poison: PoisonComponent = player.get_component(&"Poison")
	poison.apply_poison(10.0, 5.0, enemy.entity_id)
	_log("  ✅ 毒素已施加，观察 DLC PoisonSystem 的输出")


func _test_level_up() -> void:
	## 验证点 4：LevelUpSystem
	_log("\n[H] 给玩家添加大量经验（触发升级）...")
	var exp_comp: ExperienceComponent = player.get_component(&"Experience")
	if exp_comp:
		var required: float = exp_comp.get_required_exp()
		exp_comp.add_exp(required * 2.5)  # 足够升 2 级
		_log("  添加经验: %.0f，当前等级: %d，当前经验: %.0f" % [
			required * 2.5, exp_comp.level, exp_comp.current_exp
		])
	else:
		_log("  ❌ 玩家没有 ExperienceComponent")


# ─────────────────────────────────────────────
# 事件回调
# ─────────────────────────────────────────────

func _on_damage_dealt(data: Dictionary) -> void:
	var target_id: int = data.get("target_id", -1)
	var damage: float = data.get("damage", 0.0)
	var is_crit: bool = data.get("is_crit", false)
	var crit_str: String = "（暴击）" if is_crit else ""
	_log("  ⚔️ 伤害事件：Entity[%d] 受到 %.1f 伤害%s" % [target_id, damage, crit_str])

	# 检查敌人是否死亡
	if enemy.entity_id == target_id:
		var runtime: RuntimeStatsComponent = enemy.get_component(&"RuntimeStats")
		if runtime and not runtime.is_alive:
			_log("  💀 敌人已死亡！")


func _on_player_leveled_up(data: Dictionary) -> void:
	var new_level: int = data.get("new_level", 0)
	_log("  🎉 升级事件：玩家升到 %d 级！" % new_level)
	# 验证属性已重算
	var final_stats: FinalStatsComponent = player.get_component(&"FinalStats")
	if final_stats:
		_log("     升级后 MaxHP: %.0f，ATK: %.1f" % [final_stats.max_hp, final_stats.attack])


func _on_poison_tick(data: Dictionary) -> void:
	var damage: float = data.get("damage", 0.0)
	_log("  ☠️ 毒素 Tick（来自 DLC EventBus）：%.2f 伤害" % damage)


# ─────────────────────────────────────────────
# UI 更新
# ─────────────────────────────────────────────

func _process(_delta: float) -> void:
	_update_status()


func _update_status() -> void:
	if not status_label:
		return
	var player_runtime: RuntimeStatsComponent = player.get_component(&"RuntimeStats")
	var enemy_runtime: RuntimeStatsComponent = enemy.get_component(&"RuntimeStats")
	var player_final: FinalStatsComponent = player.get_component(&"FinalStats")
	var enemy_final: FinalStatsComponent = enemy.get_component(&"FinalStats")
	var player_exp: ExperienceComponent = player.get_component(&"Experience")

	var text: String = "=== Entity 状态 ===\n"
	if player_runtime and player_final:
		text += "玩家 HP: %.0f/%.0f  ATK: %.1f\n" % [
			player_runtime.current_hp, player_final.max_hp, player_final.attack
		]
	if player_exp:
		text += "等级: %d  经验: %.0f/%.0f\n" % [
			player_exp.level, player_exp.current_exp, player_exp.get_required_exp()
		]
	if enemy_runtime and enemy_final:
		text += "敌人 HP: %.0f/%.0f  %s\n" % [
			enemy_runtime.current_hp, enemy_final.max_hp,
			"（已死亡）" if not enemy_runtime.is_alive else ""
		]
	text += "DLC: %s\n" % ("已加载" if _dlc_loaded else "未加载")
	text += "Entity 总数: %d  System 总数: %d" % [
		EcsWorld.get_entity_count(), EcsWorld.get_all_systems().size()
	]
	status_label.text = text


func _log(msg: String) -> void:
	print(msg)
	_log_lines.append(msg)
	if _log_lines.size() > MAX_LOG_LINES:
		_log_lines.pop_front()
	if log_label:
		log_label.text = "\n".join(_log_lines)
