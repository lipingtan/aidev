extends Node2D
## Neon starfield + perspective floor grid background, beat-pulsed.

var au: Node = null
var _stars: Array = []
var _view := Vector2(720, 1280)
var _t := 0.0

const STAR_COLORS := [Color(0.17, 0.91, 1.0), Color(0.78, 0.42, 1.0), Color(1.0, 0.31, 0.55), Color.WHITE]

func _ready() -> void:
    randomize()
    for i in 150:
        _stars.append({
            "x": randf(), "y": randf(), "z": randf_range(0.2, 1.0),
            "c": STAR_COLORS[i % 4],
        })

func layout(v: Vector2) -> void:
    _view = v

func tick(delta: float) -> void:
    _t += delta
    var heat: int = au.heat if au else 0
    var spd := 0.02 + heat * 0.03
    for s in _stars:
        s.y += s.z * spd * delta
        if s.y > 1.0:
            s.y = 0.0
            s.x = randf()
    queue_redraw()

func _draw() -> void:
    # nebula pulses (kept dim so board contrast stays high)
    var p := 0.06 + 0.04 * sin(_t * 2.0)
    draw_circle(Vector2(_view.x * 0.3, _view.y * 0.25), _view.x * 0.55,
        Color(0.17, 0.91, 1.0, p))
    draw_circle(Vector2(_view.x * 0.75, _view.y * 0.7), _view.x * 0.6,
        Color(0.78, 0.42, 1.0, p * 0.8))
    # stars
    for s in _stars:
        var col: Color = s.c
        col.a = 0.25 + 0.6 * s.z
        draw_rect(Rect2(s.x * _view.x, s.y * _view.y, s.z * 2.2, s.z * 2.2), col)
    # perspective grid floor
    var hz := _view.y * 0.78
    var line := Color(0.17, 0.91, 1.0, 0.05)
    var scroll := fmod(_t * 0.6, 1.0)
    for i in 12:
        var yy: float = hz + pow(i + scroll, 2.1) * 3.0
        if yy > _view.y:
            break
        draw_line(Vector2(0, yy), Vector2(_view.x, yy), line, 1.0)
    # vanishing rays
    for k in range(-4, 5):
        var x0 := _view.x / 2.0 + k * 30.0
        draw_line(Vector2(_view.x / 2.0, hz), Vector2(x0 * 2.0 - _view.x / 2.0, _view.y), line * 0.6, 1.0)
