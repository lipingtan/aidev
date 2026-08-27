extends Node2D
## All juice: particles, shockwave rings, lightning bolts, floating popups,
## screen shake/hitstop/flash state. Pure _draw() based, no nodes per effect.

const MAX_PARTS := 700

var board_view: Node2D = null

# state read by main (shake offset, flash overlay)
var shake_t := 0.0
var shake_m := 0.0
var flash_amt := 0.0
var flash_col := Color.WHITE
var hitstop := 0.0

var _parts: Array = []
var _rings: Array = []
var _bolts: Array = []
var _pops: Array = []

func shake(m: float) -> void:
    shake_m = maxf(shake_m, m)
    shake_t = 1.0

func stop(t: float) -> void:
    hitstop = maxf(hitstop, t)

func flash(col: Color, a: float) -> void:
    flash_amt = maxf(flash_amt, a)
    flash_col = col

func _cell_center(gx: float, gy: float) -> Vector2:
    if board_view == null:
        return Vector2.ZERO
    return board_view.cell_to_local(gx, gy)

# ============ public API used by game.gd ============
func burst_cell(x: int, y: int, col: Color, n: int, pow: float) -> void:
    burst(_cell_center(x + 0.5, y + 0.5), col, n, pow)

func burst_board(gx: float, gy: float, col: Color, n: int, pow: float) -> void:
    burst(_cell_center(gx, gy), col, n, pow)

func burst(pos: Vector2, col: Color, n: int, pow: float) -> void:
    for i in n:
        if _parts.size() >= MAX_PARTS:
            return
        var a := randf() * TAU
        var sp := randf_range(40.0, 260.0) * pow
        _parts.append({
            "x": pos.x, "y": pos.y,
            "vx": cos(a) * sp, "vy": sin(a) * sp - 60.0,
            "g": randf_range(300.0, 700.0),
            "s": randf_range(2.0, board_view.cell_size * 0.32) if board_view else 4.0,
            "c": col, "t": 0.0, "life": randf_range(0.4, 0.9),
            "rot": randf() * TAU, "vr": randf_range(-8.0, 8.0),
        })

func ring_cell(x: int, y: int, col: Color, r: float, v: float, w: float, life: float) -> void:
    ring(_cell_center(x + 0.5, y + 0.5), col, r, v, w, life)

func ring_board(gx: float, gy: float, col: Color, r: float, v: float, w: float, life: float) -> void:
    ring(_cell_center(gx, gy), col, r, v, w, life)

func ring(pos: Vector2, col: Color, r: float, v: float, w: float, life: float) -> void:
    _rings.append({"x": pos.x, "y": pos.y, "r": r * (board_view.cell_size / 30.0) if board_view else r,
        "v": v, "c": col, "w": w, "t": 0.0, "life": life})

func ring_center(col: Color, r: float, v: float, w: float, life: float) -> void:
    var c: Vector2 = board_view.board_center_screen() if board_view else get_viewport_rect().size / 2.0
    ring(c, col, r, v, w, life)

func bolt(tx: int, ty: int) -> void:
    var cs: float = board_view.cell_size if board_view else 30.0
    var top := _cell_center(tx + 0.5, -2.0)
    var endp := _cell_center(tx + 0.5, ty + 0.5)
    var pts: Array = [top]
    var p := top
    while p.y < endp.y:
        p = Vector2(p.x + randf_range(-34.0, 34.0) * cs / 30.0, p.y + randf_range(18.0, 42.0) * cs / 30.0)
        pts.append(p)
    pts.append(endp)
    _bolts.append({"pts": pts, "t": 0.0, "life": 0.22})
    flash(Color(0.62, 0.91, 1.0), 0.12)

func pop_center(txt: String, col: Color, size: int, dx: int, dy: int) -> void:
    var c: Vector2 = board_view.board_center_screen() if board_view else get_viewport_rect().size / 2.0
    _pops.append({"x": c.x + dx, "y": c.y + dy, "txt": txt, "c": col,
        "size": size, "t": 0.0, "life": 1.3, "vy": -24.0, "big": true})

func pop_board(gx: float, gy: float, txt: String, col: Color, size: int) -> void:
    var p := _cell_center(gx, gy)
    _pops.append({"x": p.x, "y": p.y, "txt": txt, "c": col,
        "size": size, "t": 0.0, "life": 1.0, "vy": -70.0, "big": false})

# ============ update & draw ============
func tick(delta: float, frozen: bool) -> void:
    if shake_t > 0.0:
        shake_t = maxf(0.0, shake_t - delta * 3.0)
    if flash_amt > 0.0:
        flash_amt = maxf(0.0, flash_amt - delta * 2.4)
    if hitstop > 0.0:
        hitstop -= delta
    var fdt := 0.0 if frozen else delta
    for i in range(_parts.size() - 1, -1, -1):
        var q: Dictionary = _parts[i]
        q.t += fdt
        q.x += q.vx * fdt
        q.y += q.vy * fdt
        q.vy += q.g * fdt
        q.rot += q.vr * fdt
        if q.t >= q.life:
            _parts.remove_at(i)
    for i in range(_rings.size() - 1, -1, -1):
        var r: Dictionary = _rings[i]
        r.t += fdt
        r.r += r.v * fdt
        if r.t >= r.life:
            _rings.remove_at(i)
    for i in range(_bolts.size() - 1, -1, -1):
        var b: Dictionary = _bolts[i]
        b.t += fdt
        if b.t >= b.life:
            _bolts.remove_at(i)
    for i in range(_pops.size() - 1, -1, -1):
        var pop: Dictionary = _pops[i]
        pop.t += fdt
        pop.y += pop.vy * fdt
        if pop.t >= pop.life:
            _pops.remove_at(i)
    queue_redraw()

func shake_offset() -> Vector2:
    if shake_t <= 0.0 or shake_m <= 0.0:
        return Vector2.ZERO
    return Vector2(randf_range(-1, 1), randf_range(-1, 1)) * shake_m * shake_t

func _draw() -> void:
    # bolts
    for b in _bolts:
        var a: float = 1.0 - b.t / b.life
        var pts: Array = b.pts
        var n := pts.size()
        var line := PackedVector2Array()
        for i in n:
            line.append(pts[i])
        draw_polyline(line, Color(0.62, 0.91, 1.0, a), 2.5 * a + 1.0)
    # rings
    for r in _rings:
        var a2: float = 1.0 - r.t / r.life
        draw_arc(Vector2(r.x, r.y), maxf(1.0, r.r), 0.0, TAU, 48,
            Color(r.c.r, r.c.g, r.c.b, a2 * 0.9), r.w * a2 + 0.5)
    # particles
    for q in _parts:
        var a3: float = 1.0 - q.t / q.life
        var s: float = q.s
        var col: Color = q.c
        col.a = a3
        draw_set_transform(Vector2(q.x, q.y), q.rot, Vector2.ONE)
        draw_rect(Rect2(-s / 2.0, -s * 0.35, s, s * 0.7), col)
    draw_set_transform(Vector2.ZERO, 0.0, Vector2.ONE)
    # popups
    for pop in _pops:
        var a4: float = minf(1.0, (1.0 - pop.t / pop.life) * 1.6)
        var scale := 1.0
        if pop.big:
            scale = 0.7 + 0.5 * minf(1.0, pop.t * 6.0) * (1.0 + 0.08 * sin(pop.t * 20.0))
        var font := ThemeDB.fallback_font
        var fs: int = int(pop.size * scale)
        var txt: String = pop.txt
        var col2: Color = pop.c
        col2.a = a4
        draw_string(font, Vector2(pop.x, pop.y), txt, HORIZONTAL_ALIGNMENT_CENTER, -1, fs, col2)
    # full-screen flash overlay
    if flash_amt > 0.0:
        var vp := get_viewport_rect().size
        var fc := flash_col
        fc.a = clampf(flash_amt, 0.0, 0.75)
        draw_rect(Rect2(Vector2.ZERO, vp), fc)
