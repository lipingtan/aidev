extends Control
## Draws HOLD or NEXT queue pieces in a small strip.

var game: Node = null
var hold_type := ""
var mode_next := false
var next_n := 3

func _process(_d: float) -> void:
    queue_redraw()

func _draw() -> void:
    if game == null:
        return
    var f := ThemeDB.fallback_font
    if not mode_next:
        draw_string(f, Vector2(4, 12), "HOLD", HORIZONTAL_ALIGNMENT_LEFT, -1, 11, Color(0.17, 0.91, 1.0, 0.75))
        if hold_type != "":
            _draw_piece(hold_type, Vector2(8, 22), 16.0)
    else:
        draw_string(f, Vector2(4, 12), "NEXT", HORIZONTAL_ALIGNMENT_LEFT, -1, 11, Color(0.17, 0.91, 1.0, 0.75))
        var count := mini(next_n, game.queue.size())
        for i in count:
            var s := 14.0 if i > 0 else 16.0
            _draw_piece(game.queue[i], Vector2(52 + i * 42, 18), s)

func _draw_piece(type: String, at: Vector2, s: float) -> void:
    var m: Array = game.BASE[type]
    var col: Color = game.COLORS[type]
    for y in m.size():
        for x in m[y].size():
            if not m[y][x]:
                continue
            var r := Rect2(at + Vector2(x * s, y * s), Vector2(s - 2, s - 2))
            draw_rect(r, col * Color(0.5, 0.5, 0.5))
            draw_rect(r, col, false, 1.5)
