extends Node2D
## Board renderer: grid, cells, ghost piece, active piece, twin, bombs/cores/embers,
## boss sigil above the board. Neon look via HDR colors + WorldEnvironment glow.

const GameClass := preload("res://games/tetra_nova/src/scripts/game.gd")

var game: Node = null
var cell_size := 30.0
var board_origin := Vector2.ZERO
var board_size := Vector2.ZERO

func layout(view: Vector2) -> void:
    # portrait mobile layout: board slightly above center to leave room for touch controls
    var cs := minf((view.y * 0.72) / 20.0, (view.x * 0.86) / 10.0)
    cell_size = clampf(cs, 12.0, 64.0)
    board_size = Vector2(cell_size * 10.0, cell_size * 20.0)
    board_origin = Vector2((view.x - board_size.x) / 2.0, view.y * 0.5 - board_size.y * 0.52)
    queue_redraw()

func cell_to_local(gx: float, gy: float) -> Vector2:
    return board_origin + Vector2(gx, gy) * cell_size

func board_center_screen() -> Vector2:
    return board_origin + board_size / 2.0

func _process(_d: float) -> void:
    queue_redraw()

func _draw() -> void:
    if game == null or game.grid.is_empty():
        return
    draw_board_bg()
    draw_cells()
    draw_ghost_and_piece()
    draw_twin()
    draw_boss()

func draw_board_bg() -> void:
    # backdrop
    draw_rect(Rect2(board_origin - Vector2(6, 6), board_size + Vector2(12, 12)),
        Color(0.02, 0.025, 0.07, 0.72))
    draw_rect(Rect2(board_origin - Vector2(6, 6), board_size + Vector2(12, 12)),
        Color(0.17, 0.91, 1.0, 0.30), false, 2.0)
    draw_rect(Rect2(board_origin, board_size), Color(0.03, 0.05, 0.11, 0.6))
    # grid lines
    var line := Color(0.17, 0.91, 1.0, 0.06)
    for x in 11:
        var p0 := board_origin + Vector2(x * cell_size, 0)
        draw_line(p0, p0 + Vector2(0, board_size.y), line, 1.0)
    for y in 21:
        var p1 := board_origin + Vector2(0, y * cell_size)
        draw_line(p1, p1 + Vector2(board_size.x, 0), line, 1.0)

func draw_block(gx: int, gy: int, col: Color, alpha := 1.0) -> void:
    var pos := board_origin + Vector2(gx, gy) * cell_size
    var inset := cell_size * 0.05
    var r := Rect2(pos.x + inset, pos.y + inset, cell_size - inset * 2, cell_size - inset * 2)
    var a := clampf(alpha, 0.0, 1.0)
    # fake neon glow: layered translucent halos (no postfx, deterministic)
    draw_rect(r.grow(cell_size * 0.18), Color(col.r, col.g, col.b, 0.14 * a))
    draw_rect(r.grow(cell_size * 0.09), Color(col.r, col.g, col.b, 0.26 * a))
    # solid readable body
    var body := Color(col.r * 0.62 + 0.05, col.g * 0.62 + 0.05, col.b * 0.62 + 0.05, a)
    draw_rect(r, body)
    # bright rim
    var rim := Color(clampf(col.r * 1.2 + 0.35, 0.0, 1.0),
        clampf(col.g * 1.2 + 0.35, 0.0, 1.0), clampf(col.b * 1.2 + 0.35, 0.0, 1.0), a)
    draw_rect(r, rim, false, 2.0)
    # top sheen + bottom shade for chunky arcade look
    draw_rect(Rect2(r.position + Vector2(2, 2), Vector2(r.size.x - 4, r.size.y * 0.26)),
        Color(1, 1, 1, 0.30 * a))
    draw_rect(Rect2(r.position + Vector2(2, r.size.y * 0.78), Vector2(r.size.x - 4, r.size.y * 0.2)),
        Color(0, 0, 0.1, 0.28 * a))

func draw_cells() -> void:
    var t := Time.get_ticks_msec() / 1000.0
    for y in game.ROWS:
        for x in game.W:
            var c = game.grid[y][x]
            if c == null:
                continue
            draw_block(x, y, c.col)
            if c.bomb:
                var p2 := 0.5 + 0.5 * sin(t * 7.0 + x + y)
                var ctr := board_origin + Vector2(x + 0.5, y + 0.5) * cell_size
                draw_circle(ctr, cell_size * 0.16 + p2 * cell_size * 0.05,
                    Color(1.0, 0.55, 0.16, 0.6 + 0.4 * p2))
                draw_arc(ctr, cell_size * 0.2 + p2 * cell_size * 0.05, 0, TAU, 16,
                    Color(1.0, 0.86, 0.6, 0.9), 1.5)
            if c.core:
                var p3 := 0.5 + 0.5 * sin(t * 5.0)
                var ctr2 := board_origin + Vector2(x + 0.5, y + 0.5) * cell_size
                var pts := PackedVector2Array()
                for k in 6:
                    var a := k / 6.0 * TAU + t * 0.8
                    pts.append(ctr2 + Vector2(cos(a), sin(a)) * cell_size * 0.33)
                draw_colored_polygon(pts, Color(1.0, 0.42, 0.17, 0.35 + 0.3 * p3))
                draw_polyline(pts + PackedVector2Array([pts[0]]),
                    Color(1.0, 0.5, 0.24, 0.7 + 0.3 * p3), 2.0)
                draw_circle(ctr2, cell_size * 0.1, Color(1.0, 0.86, 0.7, 0.5 + 0.5 * p3))
            if c.ember > 0.0:
                var p4 := 0.5 + 0.5 * sin(t * 14.0)
                var pos := board_origin + Vector2(x, y) * cell_size + Vector2(2, 2)
                draw_rect(Rect2(pos, Vector2(cell_size - 4, cell_size - 4)),
                    Color(1.0, 0.47 + 0.31 * p4, 0.16, 0.6 + 0.4 * p4), false, 2.0)

func draw_ghost_and_piece() -> void:
    var p: Dictionary = game.piece
    if p.is_empty() or game.clearing:
        return
    # ghost
    var gy: int = game.ghost_y(p)
    for y in p.mat.size():
        for x in p.mat[y].size():
            if not p.mat[y][x]:
                continue
            var gyy: int = gy + y
            if gyy < 0:
                continue
            var pos := board_origin + Vector2(p.x + x, gyy) * cell_size + Vector2(2, 2)
            draw_rect(Rect2(pos, Vector2(cell_size - 4, cell_size - 4)),
                Color(1, 1, 1, 0.22), false, 1.5)
    # piece
    var col: Color = game.COLORS[p.type]
    for y in p.mat.size():
        for x in p.mat[y].size():
            if not p.mat[y][x]:
                continue
            var pyy: int = p.y + y
            if pyy < 0:
                continue
            draw_block(p.x + x, pyy, col)

func draw_twin() -> void:
    var tw: Dictionary = game.twin
    if tw.is_empty():
        return
    for y in tw.mat.size():
        for x in tw.mat[y].size():
            if not tw.mat[y][x]:
                continue
            var gyy: int = tw.y + y
            if gyy < 0:
                continue
            draw_block(tw.x + x, gyy, game.TWIN_COL, 0.85)

func draw_boss() -> void:
    var b: Dictionary = game.boss
    if b.is_empty():
        return
    var t := Time.get_ticks_msec() / 1000.0
    var dying: bool = b.die_t > 0.0
    var flashing: bool = b.flash_t > 0.0
    var col := Color(0.8, 0.8, 0.8) if dying else (Color.WHITE if flashing else Color(1.0, 0.35, 0.18))
    var ctr := board_origin + Vector2(5.0, -2.6) * cell_size
    var pts := PackedVector2Array()
    for k in 6:
        var a := k / 6.0 * TAU + t * 0.7
        pts.append(ctr + Vector2(cos(a), sin(a)) * cell_size * 1.35)
    draw_polyline(pts + PackedVector2Array([pts[0]]), col, 3.0)
    var tri := PackedVector2Array()
    for k in 3:
        var a2 := k / 3.0 * TAU - t * 2.1
        tri.append(ctr + Vector2(cos(a2), sin(a2)) * cell_size * 0.8)
    draw_polyline(tri + PackedVector2Array([tri[0]]), Color(0.8, 0.8, 0.8) if dying else Color(1.0, 0.6, 0.22), 2.0)
    var p2 := 0.5 + 0.5 * sin(t * 6.0)
    draw_circle(ctr, cell_size * 0.22 + p2 * 3.0, col)
    # garbage telegraph
    if b.atk_t <= 1.5 and b.die_t <= 0.0:
        var warn := 0.4 + 0.4 * sin(t * 18.0)
        draw_rect(Rect2(board_origin - Vector2(6, 6), board_size + Vector2(12, 12)),
            Color(1.0, 0.24, 0.16, warn), false, 3.0)
