class_name CtaBar extends Control
## CTA 固定区组件（≤80行）；setup(gid) 初始化，refresh() 重算状态机

@onready var _hint: Label = $VBox/HintLabel
@onready var _btn: Button = $VBox/CtaButton

var _gid: String = ""

func _ready() -> void:
	# 信号在 _ready 中连接一次，避免 setup() 多次调用重复连接
	EventBus.records_updated.connect(_on_records_updated)
	EventBus.trial_consumed.connect(_on_trial_consumed)
	# payment_succeeded / dlc_installed → 占位注释（M3/M4 接线）

func setup(gid: String) -> void:
	_gid = gid
	refresh()

func refresh() -> void:
	if _gid == "": return
	var state: Dictionary = _cta_state()
	_hint.text = state["hint"]
	_btn.text = state["label"]
	# 断开旧连接后重连，防止 push 同一页面时 action 变化
	if _btn.pressed.is_connected(_on_cta_pressed):
		_btn.pressed.disconnect(_on_cta_pressed)
	_btn.pressed.connect(_on_cta_pressed)

func _on_cta_pressed() -> void:
	var state: Dictionary = _cta_state()
	if state["action"] == "launch":
		Launcher.launch(_gid)
	else:
		_show_purchase()

## CTA 状态机（FR-5）：纯读 DB+Registry，无副作用
func _cta_state() -> Dictionary:
	var meta: Variant = Registry.lookup(_gid)
	if meta == null: return {"label": "—", "hint": "", "action": "none"}
	var rec: Variant = DB.get_record(_gid)
	var r: Dictionary = rec if rec is Dictionary else {}
	var model: String = (meta as GameMeta).price_model
	var price: int = (meta as GameMeta).price
	match model:
		"free", "ad", "iap":
			return _open_play_state(r)
		"trial":
			var left: int = TrialGuard.left(_gid)
			if left > 0:
				return {"label": "▶ 试玩 · 剩 %d 次" % left,
						"hint": "试玩结束后 ¥%d 解锁全部" % price, "action": "launch"}
			else:
				return {"label": "¥%d 解锁完整版" % price,
						"hint": "试玩次数已用完", "action": "buy"}
		"paid":
			if _is_owned():
				return _open_play_state(r)
			return {"label": "¥%d 购买" % price,
					"hint": "买断制 · 一次购买永久拥有", "action": "buy"}
		_:
			return {"label": "—", "hint": "", "action": "none"}

func _open_play_state(r: Dictionary) -> Dictionary:
	var best: int = int(r.get("best", 0))
	var hint: String = "首次开玩" if best == 0 else "上次最高 %d 分 · 继续" % best
	return {"label": "▶ 开玩", "hint": hint, "action": "launch"}

func _is_owned() -> bool:
	for order in DB.list_orders().values():
		var o: Dictionary = order as Dictionary
		if o.get("gid", "") == _gid and o.get("status", "") == "paid":
			return true
	return false

func _show_purchase() -> void:
	var meta: Variant = Registry.lookup(_gid)
	if meta == null: return
	var dlg := get_tree().root.find_child("PurchaseDialog", true, false)
	if dlg == null:
		push_error("CtaBar: 未找到 PurchaseDialog 节点")
		return
	(dlg as Control).call("show_mock", (meta as GameMeta).title, (meta as GameMeta).price)

func _on_records_updated(gid: String) -> void:
	if gid == _gid: refresh()

func _on_trial_consumed(gid: String, _left: int) -> void:
	if gid == _gid: refresh()
