extends CanvasLayer
## HUD + all overlay screens (menu / card select / pause / game over), portrait layout.

signal action(a: String)

var game: Node = null

## Save file path (GameModule adapter injects ctx.save_dir + "save.cfg" at boot; standalone default below)
var SAVE_PATH := "user://save.cfg"

var _score: Label
var _wave: Label
var _lines: Label
var _combo: Label
var _wave_bar: ColorRect
var _heat_bar: ColorRect
var _boss_bar: ColorRect
var _boss_wrap: Control
var _boss_name: Label
var _hold_view: Control
var _next_view: Control
var _mut_row: HBoxContainer
var _cards: HBoxContainer
var _sel_title: Label
var _sel_sub: Label
var _reroll_btn: Button
var _menu: Control
var _select: Control
var _pause: Control
var _over: Control
var _over_stats: Label
var _new_best: Label

const RAR_COL := [Color.TRANSPARENT,
    Color(0.17, 0.91, 1.0), Color(0.78, 0.42, 1.0), Color(1.0, 0.60, 0.22), Color(1.0, 0.82, 0.25)]

func _ready() -> void:
    layer = 10
    _build()
    _refresh_best()

# ===================================================================== build
func _lbl(txt: String, size: int, col := Color.WHITE) -> Label:
    var l := Label.new()
    l.text = txt
    l.add_theme_font_size_override("font_size", size)
    l.add_theme_color_override("font_color", col)
    l.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
    return l

func _build() -> void:
    var root := Control.new()
    root.set_anchors_preset(Control.PRESET_FULL_RECT)
    root.mouse_filter = Control.MOUSE_FILTER_IGNORE
    add_child(root)

    # ---------- top bar: HOLD | score block | NEXT ----------
    var top := HBoxContainer.new()
    top.set_anchors_preset(Control.PRESET_TOP_WIDE)
    top.offset_left = 10
    top.offset_right = -10
    top.offset_top = 8
    top.mouse_filter = Control.MOUSE_FILTER_IGNORE
    root.add_child(top)

    _hold_view = preload("res://games/tetra_nova/src/scripts/mini_piece.gd").new()
    _hold_view.custom_minimum_size = Vector2(120, 84)
    top.add_child(_hold_view)

    var mid := VBoxContainer.new()
    mid.size_flags_horizontal = Control.SIZE_EXPAND_FILL
    mid.mouse_filter = Control.MOUSE_FILTER_IGNORE
    top.add_child(mid)
    _score = _lbl("0", 30, Color.WHITE)
    mid.add_child(_score)
    var row := HBoxContainer.new()
    row.alignment = BoxContainer.ALIGNMENT_CENTER
    mid.add_child(row)
    _wave = _lbl("WAVE 1", 13, Color(0.56, 0.85, 1.0))
    _lines = _lbl("  LINES 0", 13, Color(0.56, 0.85, 1.0))
    row.add_child(_wave)
    row.add_child(_lines)
    _wave_bar = _mk_bar(mid, Vector2(200, 8), Color(0.17, 0.91, 1.0))
    _combo = _lbl("—", 15, Color(1.0, 0.31, 0.55))
    mid.add_child(_combo)
    _heat_bar = _mk_bar(mid, Vector2(200, 6), Color(1.0, 0.31, 0.55))

    _next_view = preload("res://games/tetra_nova/src/scripts/mini_piece.gd").new()
    _next_view.custom_minimum_size = Vector2(180, 84)
    _next_view.mode_next = true
    top.add_child(_next_view)

    # ---------- boss bar ----------
    _boss_wrap = VBoxContainer.new()
    _boss_wrap.set_anchors_preset(Control.PRESET_CENTER_TOP)
    _boss_wrap.offset_left = -260
    _boss_wrap.offset_right = 260
    _boss_wrap.offset_top = 108
    _boss_wrap.visible = false
    _boss_wrap.mouse_filter = Control.MOUSE_FILTER_IGNORE
    root.add_child(_boss_wrap)
    _boss_name = _lbl("BOSS", 14, Color(1.0, 0.55, 0.18))
    _boss_wrap.add_child(_boss_name)
    _boss_bar = _mk_bar(_boss_wrap, Vector2(520, 12), Color(1.0, 0.35, 0.18))

    # ---------- mutations row (bottom) ----------
    var mut_anchor := HBoxContainer.new()
    mut_anchor.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
    mut_anchor.offset_left = 8
    mut_anchor.offset_right = -8
    mut_anchor.offset_bottom = -118
    mut_anchor.offset_top = -158
    mut_anchor.alignment = BoxContainer.ALIGNMENT_CENTER
    mut_anchor.mouse_filter = Control.MOUSE_FILTER_IGNORE
    root.add_child(mut_anchor)
    _mut_row = mut_anchor

    # ---------- overlays ----------
    _menu = _mk_overlay(root)
    _build_menu(_menu)
    _select = _mk_overlay(root)
    _build_select(_select)
    _pause = _mk_overlay(root)
    _build_pause(_pause)
    _over = _mk_overlay(root)
    _build_over(_over)
    _show_overlay("MENU")

func _mk_bar(parent: Control, sz: Vector2, col: Color) -> ColorRect:
    var bg := ColorRect.new()
    bg.color = Color(1, 1, 1, 0.08)
    bg.custom_minimum_size = sz
    bg.mouse_filter = Control.MOUSE_FILTER_IGNORE
    parent.add_child(bg)
    var fill := ColorRect.new()
    fill.color = col
    fill.size = Vector2.ZERO
    fill.mouse_filter = Control.MOUSE_FILTER_IGNORE
    bg.add_child(fill)
    return fill

func _mk_overlay(parent: Control) -> Control:
    var ov := Control.new()
    ov.set_anchors_preset(Control.PRESET_FULL_RECT)
    ov.mouse_filter = Control.MOUSE_FILTER_STOP
    var dim := ColorRect.new()
    dim.color = Color(0.01, 0.016, 0.05, 0.86)
    dim.set_anchors_preset(Control.PRESET_FULL_RECT)
    ov.add_child(dim)
    var box := VBoxContainer.new()
    box.set_anchors_preset(Control.PRESET_CENTER)
    box.offset_left = -330
    box.offset_right = 330
    box.offset_top = -320
    box.offset_bottom = 320
    box.alignment = BoxContainer.ALIGNMENT_CENTER
    ov.add_child(box)
    ov.set_meta("box", box)
    return ov

func _mk_button(parent: Control, txt: String, cb: String) -> Button:
    var b := Button.new()
    b.text = txt
    b.add_theme_font_size_override("font_size", 20)
    b.custom_minimum_size = Vector2(240, 56)
    b.pressed.connect(func(): action.emit(cb))
    (parent as VBoxContainer).add_child(b)
    return b

func _build_menu(ov: Control) -> void:
    var box: VBoxContainer = ov.get_meta("box")
    box.add_child(_lbl("TETRA NOVA", 58, Color(0.17, 0.91, 1.0)))
    box.add_child(_lbl("新 星 方 阵 · 变 异 消 除", 16, Color(0.56, 0.85, 1.0)))
    box.add_child(_lbl(" ", 10, Color.TRANSPARENT))
    _mk_button(box, "▸ INITIALIZE", "confirm")
    box.add_child(_lbl(" ", 12, Color.TRANSPARENT))
    box.add_child(_lbl("手势/键盘：← → 移动 · ↓ 软降 · 空格 硬降", 13, Color(0.56, 0.72, 0.85)))
    box.add_child(_lbl("↑/X 旋转 · Z 反旋 · C 暂存 · P 暂停 · M 静音", 13, Color(0.56, 0.72, 0.85)))
    box.add_child(_lbl("每 5 波遭遇 BOSS · 3 选 1 变异无限叠加", 13, Color(0.56, 0.72, 0.85)))
    var best := Label.new()
    best.name = "BestLbl"
    box.add_child(best)

func _build_select(ov: Control) -> void:
    var box: VBoxContainer = ov.get_meta("box")
    _sel_title = _lbl("SYSTEM UPGRADE", 34, Color.WHITE)
    box.add_child(_sel_title)
    _sel_sub = _lbl("选 择 一 项 变 异", 14, Color(0.56, 0.85, 1.0))
    box.add_child(_sel_sub)
    box.add_child(_lbl(" ", 8, Color.TRANSPARENT))
    _cards = HBoxContainer.new()
    _cards.add_theme_constant_override("separation", 18)
    _cards.alignment = BoxContainer.ALIGNMENT_CENTER
    box.add_child(_cards)
    _reroll_btn = Button.new()
    _reroll_btn.text = "↻ 重抽 [R]"
    _reroll_btn.add_theme_font_size_override("font_size", 16)
    _reroll_btn.custom_minimum_size = Vector2(180, 44)
    _reroll_btn.pressed.connect(func(): action.emit("reroll"))
    box.add_child(_reroll_btn)

func _build_pause(ov: Control) -> void:
    var box: VBoxContainer = ov.get_meta("box")
    box.add_child(_lbl("PAUSED", 40, Color(0.78, 0.42, 1.0)))
    box.add_child(_lbl(" ", 10, Color.TRANSPARENT))
    _mk_button(box, "▸ RESUME", "pause")
    _mk_button(box, "RESTART RUN", "rerun")

func _build_over(ov: Control) -> void:
    var box: VBoxContainer = ov.get_meta("box")
    box.add_child(_lbl("SIGNAL LOST", 46, Color(1.0, 0.31, 0.55)))
    _new_best = _lbl("★ NEW RECORD ★", 20, Color(1.0, 0.82, 0.25))
    _new_best.visible = false
    box.add_child(_new_best)
    _over_stats = _lbl("", 17, Color(0.87, 0.94, 1.0))
    box.add_child(_over_stats)
    box.add_child(_lbl(" ", 10, Color.TRANSPARENT))
    _mk_button(box, "▸ REBOOT [R]", "rerun")

# ===================================================================== state
func _show_overlay(which: String) -> void:
    _menu.visible = which == "MENU"
    _select.visible = which == "SELECT"
    _pause.visible = which == "PAUSED"
    _over.visible = which == "OVER"

func on_state_changed(state: String) -> void:
    match state:
        "MENU": _show_overlay("MENU")
        "SELECT": _show_overlay("SELECT")
        "PAUSED": _show_overlay("PAUSED")
        "OVER": _show_overlay("OVER")
        _: _show_overlay("NONE")

func on_select_opened(boss_reward: bool) -> void:
    _sel_title.text = "BOSS SPOILS" if boss_reward else "SYSTEM UPGRADE"
    _sel_sub.text = "战 利 品 · 稀 有 以 上" if boss_reward else "选 择 一 项 变 异 · WAVE " + str(game.wave) + " 完成"
    _reroll_btn.text = "↻ 重抽 ×" + str(game.rerolls) + " [R]"
    _reroll_btn.disabled = game.rerolls <= 0
    build_cards()

func on_run_over(stats: Dictionary) -> void:
    _over_stats.text = "SCORE  %d\nWAVE  %d\nLINES  %d\nMAX COMBO  ×%d\nBOSS SLAIN  %d" % [
        stats.score, stats.wave, stats.lines, stats.max_combo, stats.bosses]
    var best := 0
    var cfg := ConfigFile.new()
    if cfg.load(SAVE_PATH) == OK:
        best = cfg.get_value("nova", "best", 0)
    _new_best.visible = stats.score >= best and stats.score > 0
    _refresh_best()

func _refresh_best() -> void:
    var best := 0
    var wave := 1
    var runs := 0
    var cfg := ConfigFile.new()
    if cfg.load(SAVE_PATH) == OK:
        best = cfg.get_value("nova", "best", 0)
        wave = cfg.get_value("nova", "wave", 1)
        runs = cfg.get_value("nova", "runs", 0)
    var lbl: Label = _menu.get_node_or_null("VBoxContainer/BestLbl")
    # find by traversing
    if lbl == null:
        var box: VBoxContainer = _menu.get_meta("box")
        for c in box.get_children():
            if c is Label and c.name == "BestLbl":
                lbl = c
                break
    if lbl:
        lbl.text = ("BEST %d · WAVE %d · RUNS %d" % [best, wave, runs]) if runs > 0 else "NO DATA YET"
        lbl.add_theme_font_size_override("font_size", 14)
        lbl.add_theme_color_override("font_color", Color(1.0, 0.82, 0.25))

# ===================================================================== cards
func build_cards() -> void:
    for c in _cards.get_children():
        c.queue_free()
    var owned: Dictionary = {}
    for t in game.owned_tags():
        owned[t] = true
    for i in game.current_cards.size():
        var m: Dictionary = game.current_cards[i]
        var card := PanelContainer.new()
        card.custom_minimum_size = Vector2(200, 264)
        var st := StyleBoxFlat.new()
        st.bg_color = Color(0.05, 0.08, 0.17, 0.95)
        st.set_corner_radius_all(12)
        st.set_border_width_all(2)
        var rc: Color = RAR_COL[m.r]
        st.border_color = rc
        st.content_margin_left = 12
        st.content_margin_right = 12
        st.content_margin_top = 10
        st.content_margin_bottom = 10
        card.add_theme_stylebox_override("panel", st)
        var v := VBoxContainer.new()
        v.alignment = BoxContainer.ALIGNMENT_CENTER
        card.add_child(v)
        var rar := _lbl("◆ " + game.RARNAME[m.r], 11, rc)
        v.add_child(rar)
        var stk: int = game.mut_stk(m.id)
        if m.max < 90 and stk > 0:
            v.add_child(_lbl("%d / %d" % [stk, m.max], 10, Color(0.7, 0.8, 0.9)))
        v.add_child(_lbl(m.ic, 44, Color.WHITE))
        v.add_child(_lbl(m.nm, 19, Color.WHITE))
        v.add_child(_lbl(m.ds, 12, Color(0.66, 0.8, 0.91)))
        var syn := 0
        for t in m.tags:
            if owned.has(t):
                syn += 1
        if syn > 0:
            v.add_child(_lbl("SYNERGY ×%d" % syn, 11, Color(1.0, 0.82, 0.25)))
        v.add_child(_lbl("[%d]" % (i + 1), 11, Color(0.4, 0.55, 0.65)))
        var ii: int = i
        card.gui_input.connect(func(ev: InputEvent):
            if ev is InputEventScreenTouch and ev.pressed:
                action.emit("card" + str(ii + 1))
        )
        card.mouse_entered.connect(func(): card.modulate = Color(1.25, 1.25, 1.25))
        card.mouse_exited.connect(func(): card.modulate = Color.WHITE)
        _cards.add_child(card)

# ===================================================================== per-frame
func tick() -> void:
    if game == null:
        return
    _score.text = game.fmt(game.score)
    _wave.text = "WAVE " + str(game.wave)
    _lines.text = "  LINES " + str(game.lines)
    _combo.text = ("×" + str(game.combo + 1)) if game.combo >= 1 else "—"
    if game.boss.is_empty():
        _boss_wrap.visible = false
        _wave_bar.size.x = clampf(float(game.wave_lines) / float(game.wave_goal), 0.0, 1.0) * 200.0
    else:
        _boss_wrap.visible = true
        _boss_name.text = "☠ " + str(game.boss.name)
        _boss_bar.size.x = clampf(float(game.boss.hp) / float(game.boss.max), 0.0, 1.0) * 520.0
    _heat_bar.size.x = clampf(float(game.combo + 1) / 8.0, 0.0, 1.0) * 200.0
    _hold_view.hold_type = game.hold_type
    _hold_view.game = game
    _next_view.game = game
    _next_view.next_n = game.next_n
    _mut_tick()

var _mut_cache := ""
func _mut_tick() -> void:
    var sig := ""
    for k in game.muts:
        sig += k + str(game.muts[k]) + ","
    if sig == _mut_cache:
        return
    _mut_cache = sig
    for c in _mut_row.get_children():
        c.queue_free()
    for m in game.MUTDEFS:
        var s: int = game.mut_stk(m.id)
        if s <= 0:
            continue
        var box := HBoxContainer.new()
        var pnl := PanelContainer.new()
        var st := StyleBoxFlat.new()
        st.bg_color = Color(1, 1, 1, 0.06)
        st.set_corner_radius_all(8)
        st.border_color = RAR_COL[m.r]
        st.set_border_width_all(1)
        st.content_margin_left = 8
        st.content_margin_right = 8
        st.content_margin_top = 3
        st.content_margin_bottom = 3
        pnl.add_theme_stylebox_override("panel", st)
        pnl.add_child(box)
        var ic := _lbl(m.ic, 20, Color.WHITE)
        box.add_child(ic)
        if s > 1:
            box.add_child(_lbl("×" + str(s), 13, Color(1.0, 0.82, 0.25)))
        _mut_row.add_child(pnl)
