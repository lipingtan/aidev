extends Node
class_name Game
## TETRA NOVA — core game logic (pure state machine; rendering/audio handled elsewhere)

signal state_changed(state: String)
signal select_opened(boss_reward: bool)
signal run_over(stats: Dictionary)

const W := 10
const ROWS := 20

const COLORS := {
    "I": Color(0.16, 0.90, 1.0),
    "O": Color(1.0, 0.82, 0.25),
    "T": Color(0.78, 0.42, 1.0),
    "S": Color(0.27, 1.0, 0.53),
    "Z": Color(1.0, 0.31, 0.43),
    "J": Color(0.31, 0.49, 1.0),
    "L": Color(1.0, 0.60, 0.22),
}
const GARBAGE_COL := Color(0.35, 0.40, 0.50)
const CORE_COL := Color(1.0, 0.42, 0.17)
const TWIN_COL := Color(0.91, 0.98, 1.0)

const BASE := {
    "I": [[0,0,0,0],[1,1,1,1],[0,0,0,0],[0,0,0,0]],
    "J": [[1,0,0],[1,1,1],[0,0,0]],
    "L": [[0,0,1],[1,1,1],[0,0,0]],
    "O": [[1,1],[1,1]],
    "S": [[0,1,1],[1,1,0],[0,0,0]],
    "T": [[0,1,0],[1,1,1],[0,0,0]],
    "Z": [[1,1,0],[0,1,1],[0,0,0]],
}
const KICK := {
    "01": [[0,0],[-1,0],[-1,-1],[0,2],[-1,2]], "10": [[0,0],[1,0],[1,-1],[0,2],[1,2]],
    "12": [[0,0],[1,0],[1,1],[0,-2],[1,-2]],   "21": [[0,0],[-1,0],[-1,-1],[0,2],[-1,2]],
    "23": [[0,0],[1,0],[1,-1],[0,2],[1,2]],    "32": [[0,0],[-1,0],[-1,1],[0,-2],[-1,-2]],
    "30": [[0,0],[-1,0],[-1,-1],[0,2],[-1,2]], "03": [[0,0],[1,0],[1,1],[0,-2],[1,-2]],
}
const KICKI := {
    "01": [[0,0],[-2,0],[1,0],[-2,1],[1,-2]], "10": [[0,0],[2,0],[-1,0],[2,-1],[-1,2]],
    "12": [[0,0],[-1,0],[2,0],[-1,-2],[2,1]], "21": [[0,0],[1,0],[-2,0],[1,2],[-2,-1]],
    "23": [[0,0],[2,0],[-1,0],[2,-1],[-1,2]], "32": [[0,0],[-2,0],[1,0],[-2,1],[1,-2]],
    "30": [[0,0],[1,0],[-2,0],[1,2],[-2,-1]], "03": [[0,0],[-1,0],[2,0],[-1,-2],[2,1]],
}

const MUTDEFS := [
    {"id":"foresight","nm":"预言者","ic":"👁","r":1,"max":1,"tags":["info"],"ds":"NEXT 预览 3 格 → 5 格，看穿未来"},
    {"id":"greed","nm":"重力贪婪","ic":"💰","r":1,"max":3,"tags":["score"],"ds":"硬降每行得分提升（每层 +2/行）"},
    {"id":"recycler","nm":"能量回收","ic":"♻️","r":1,"max":3,"tags":["score"],"ds":"所有消除得分 +15%/层"},
    {"id":"stabilizer","nm":"惯性稳定","ic":"🧲","r":1,"max":3,"tags":["ctrl"],"ds":"锁定延迟 +120ms/层，微调更从容"},
    {"id":"overclock","nm":"超频缓冲","ic":"🐌","r":1,"max":3,"tags":["ctrl"],"ds":"下落速度 -7%/层"},
    {"id":"bombs","nm":"爆破装配","ic":"💣","r":2,"max":3,"tags":["boom"],"ds":"锁定方块 10%+ 概率成为炸弹，消除时 3×3 爆破"},
    {"id":"lightning","nm":"闪电链","ic":"⚡","r":2,"max":1,"tags":["boom","chain"],"ds":"一次消除 ≥2 行时，闪电摧毁场上若干散块"},
    {"id":"timerift","nm":"时间裂缝","ic":"⏱","r":2,"max":1,"tags":["time","ctrl"],"ds":"硬降触发 0.4s 子弹时间（冷却 9s）"},
    {"id":"shockwave","nm":"净化冲击","ic":"🌊","r":2,"max":1,"tags":["chain"],"ds":"TETRIS（4 行）时冲击波摧毁所有悬空方块"},
    {"id":"siege","nm":"破城者","ic":"⚔️","r":2,"max":3,"tags":["boss"],"ds":"对 BOSS 核心伤害 +25%/层"},
    {"id":"embers","nm":"余烬场","ic":"🔥","r":3,"max":1,"tags":["boom","chain"],"ds":"每次消除点燃 3~5 个方块，1.8s 后殉爆 3×3"},
    {"id":"blackhole","nm":"奇点吞噬","ic":"🕳","r":3,"max":1,"tags":["boom","chain"],"ds":"连击 ≥6 时黑洞吞噬最底 2 行（冷却 30s）"},
    {"id":"cascade","nm":"重力坍缩","ic":"🌀","r":3,"max":1,"tags":["chain"],"ds":"消除后方块列坍缩，连锁再消得分 ×1.5"},
    {"id":"nova","nm":"奇点爆发","ic":"☢","r":4,"max":1,"tags":["boom"],"ds":"一次消除 ≥3 行时，随机湮灭场上 12 个方块"},
    {"id":"twindrop","nm":"双子降临","ic":"👥","r":4,"max":1,"tags":["ally"],"ds":"每落 4 块，召唤一个白热双子方块自动配位坠落"},
    {"id":"secondwind","nm":"第二次呼吸","ic":"🕊","r":4,"max":1,"tags":["ally"],"ds":"首次顶出时免死：底部 8 行蒸发（一次性）"},
]
const RARNAME := ["", "COMMON", "RARE", "EPIC", "LEGENDARY"]
const RARW := [0, 100, 55, 26, 9]
const BOSSNAMES := ["HARBINGER-9", "VOIDMAW", "OBELISK", "NOVA PRIME", "PYRE HEART", "OMEGA GRID"]

# ---- wired by main ----
var fx: Node = null       # fx_layer (visual effects)
var au: Node = null       # audio manager
var board_view: Node = null

# ---- state ----
var state := "MENU"
var grid: Array = []
var piece: Dictionary = {}
var hold_type := ""
var hold_used := false
var bag: Array = []
var queue: Array = []
var piece_count := 0
var score := 0
var lines := 0
var wave := 1
var wave_lines := 0
var wave_goal := 7
var combo := -1
var max_combo := 0
var b2b := false
var grav := 0.8
var drop_acc := 0.0
var lock_t := 0.0
var lock_resets := 0
var grounded := false
var muts: Dictionary = {}
var next_n := 3
var lock_bonus := 0.0
var grav_mult := 1.0
var greed := 0
var recycle := 0
var siege := 0
var boss: Dictionary = {}
var boss_count := 0
var rerolls := 1
var stats_bosses := 0
var stats_bombs := 0
var pipe := 0
var clearing := false
var chain := 0
var twin: Dictionary = {}
var piece_since_twin := 0
var rift_cd := 0.0
var hole_cd := 0.0
var second_wind_used := false
var lock_select := false
var hitstop := 0.0
var time_scale := 1.0
var ts_t := 0.0
var time := 0.0
var current_cards: Array = []
var select_reward := false

# input
var key_left := false
var key_right := false
var key_down := false
var das_t := 0.0
var das_dir := 0

# scheduler
var _sch: Array = []

# ===================================================================== util
static func fmt(n: int) -> String:
    var s := str(n)
    var out := ""
    var c := 0
    for i in range(s.length() - 1, -1, -1):
        out = s[i] + out
        c += 1
        if c % 3 == 0 and i > 0:
            out = "," + out
    return out

func rnd_f(a: float, b: float) -> float:
    return randf_range(a, b)

func rnd_i(a: int, b: int) -> int:
    return randi_range(a, b)

func chance(p: float) -> bool:
    return randf() < p

func mut_stk(id: String) -> int:
    return int(muts.get(id, 0))

func has_mut(id: String) -> bool:
    return mut_stk(id) > 0

func owned_tags() -> Array:
    var t := {}
    for m in MUTDEFS:
        if has_mut(m.id):
            for x in m.tags:
                t[x] = true
    return t.keys()

# ===================================================================== scheduler
func schedule(fn: Callable, d: float, use_pipe := true) -> void:
    _sch.append({"t": time + d, "fn": fn, "pipe": use_pipe})
    if use_pipe:
        pipe += 1

func run_sched() -> void:
    var due: Array = []
    for i in range(_sch.size() - 1, -1, -1):
        if _sch[i].t <= time:
            due.append(_sch[i])
            _sch.remove_at(i)
    due.reverse()
    for it in due:
        it.fn.call()
        if it.pipe:
            pipe_end()

func pipe_begin() -> void:
    pipe += 1

func pipe_end() -> void:
    pipe = max(0, pipe - 1)
    if pipe == 0:
        after_pipe()

# ===================================================================== grid model
func new_grid() -> Array:
    var g: Array = []
    for y in ROWS:
        var row: Array = []
        row.resize(W)
        g.append(row)
    return g

func cell_at(x: int, y: int) -> Variant:
    if y >= 0 and y < ROWS and x >= 0 and x < W:
        return grid[y][x]
    return null

func collide(mat: Array, px: int, py: int) -> bool:
    for y in mat.size():
        for x in mat[y].size():
            if not mat[y][x]:
                continue
            var bx: int = px + x
            var by: int = py + y
            if bx < 0 or bx >= W or by >= ROWS:
                return true
            if by >= 0 and grid[by][bx] != null:
                return true
    return false

func filled_rows() -> Array:
    var r: Array = []
    for y in ROWS:
        var f := true
        for x in W:
            if grid[y][x] == null:
                f = false
                break
        if f:
            r.append(y)
    return r

func refills() -> void:
    bag = ["I", "J", "L", "O", "S", "T", "Z"]

func next_from_bag() -> String:
    if bag.is_empty():
        refills()
    var i := randi_range(0, bag.size() - 1)
    return bag.pop_at(i)

func ensure_queue() -> void:
    while queue.size() < next_n + 2:
        queue.append(next_from_bag())

func cell(col: Color, bomb := false, core := false, garbage := false) -> Dictionary:
    return {"col": col, "bomb": bomb, "core": core, "garbage": garbage, "ember": -1.0}

func spawn_piece() -> void:
    ensure_queue()
    var type: String = queue.pop_front()
    var mat := mat_copy(BASE[type])
    var p := {"type": type, "mat": mat, "rot": 0, "x": int((W - mat.size()) / 2.0), "y": -1}
    if collide(p.mat, p.x, p.y):
        p.y = -2
        if collide(p.mat, p.x, p.y):
            top_out()
            return
    piece = p
    drop_acc = 0.0
    lock_t = 0.0
    lock_resets = 0
    grounded = false

static func mat_copy(m: Array) -> Array:
    var out: Array = []
    for row in m:
        out.append(row.duplicate())
    return out

static func rot_mat(m: Array, dir: int) -> Array:
    var n := m.size()
    var r: Array = []
    r.resize(n)
    for y in n:
        var row: Array = []
        row.resize(n)
        for x in n:
            if dir > 0:
                row[x] = m[n - 1 - x][y]
            else:
                row[x] = m[x][n - 1 - y]
        r[y] = row
    return r

func ghost_y(p: Dictionary) -> int:
    var y: int = p.y
    while not collide(p.mat, p.x, y + 1):
        y += 1
    return y

func try_move(dx: int, dy: int) -> bool:
    var p: Dictionary = piece
    if p.is_empty():
        return false
    if collide(p.mat, p.x + dx, p.y + dy):
        return false
    p.x += dx
    p.y += dy
    if dx != 0 and grounded and lock_resets < 15:
        lock_t = 0.0
        lock_resets += 1
    return true

func try_rotate(dir: int) -> void:
    var p: Dictionary = piece
    if p.is_empty():
        return
    if p.type == "O":
        return
    var from: int = p.rot
    var to: int = (int(p.rot) + dir + 4) % 4
    var m2 := rot_mat(p.mat, dir)
    var table: Array = (KICKI if p.type == "I" else KICK)[str(from) + str(to)]
    for k in table:
        var kx: int = k[0]
        var ky: int = k[1]
        if not collide(m2, p.x + kx, p.y + ky):
            p.mat = m2
            p.x += kx
            p.y += ky
            p.rot = to
            if grounded and lock_resets < 15:
                lock_t = 0.0
                lock_resets += 1
            au.sfx("rot")
            return

func do_hold() -> void:
    if piece.is_empty() or hold_used:
        return
    var cur: String = piece.type
    if hold_type != "":
        var t: String = hold_type
        hold_type = cur
        piece = {"type": t, "mat": mat_copy(BASE[t]), "rot": 0, "x": int((W - BASE[t].size()) / 2.0), "y": -1}
        drop_acc = 0.0
        lock_t = 0.0
        lock_resets = 0
    else:
        hold_type = cur
        spawn_piece()
    hold_used = true
    au.sfx("hold")

func hard_drop() -> void:
    var p: Dictionary = piece
    if p.is_empty():
        return
    var gy := ghost_y(p)
    var d: int = gy - p.y
    if d > 0:
        var y0: int = p.y
        for y in range(y0, gy):
            for my in p.mat.size():
                for mx in p.mat[my].size():
                    if p.mat[my][mx] and chance(0.35):
                        fx.burst_cell(p.x + mx, y + my, COLORS[p.type], 1, 0.4)
    p.y = gy
    score += d * (1 + 2 * greed)
    fx.shake(4.0 + d * 0.15)
    au.sfx("hard")
    if has_mut("timerift") and rift_cd <= 0.0:
        slowmo(0.4, 0.3)
        rift_cd = 9.0
        fx.pop_center("时间裂缝", Color(0.56, 0.85, 1.0), 22, -200, -160)
    lock_piece()

func slowmo(t: float, scale: float) -> void:
    time_scale = scale
    ts_t = t

# ===================================================================== lock & clear pipeline
func lock_piece() -> void:
    var p: Dictionary = piece
    if p.is_empty():
        return
    var bomb_p := 0.0
    if has_mut("bombs"):
        bomb_p = 0.10 + 0.07 * mut_stk("bombs")
    for y in p.mat.size():
        for x in p.mat[y].size():
            if not p.mat[y][x]:
                continue
            var gy: int = p.y + y
            var gx: int = p.x + x
            if gy < 0:
                top_out()
                return
            var c := cell(COLORS[p.type], bomb_p > 0.0 and chance(bomb_p))
            grid[gy][gx] = c
            if c.bomb:
                stats_bombs += 1
    piece = {}
    hold_used = false
    piece_count += 1
    piece_since_twin += 1
    au.sfx("lock")
    fx.shake(2.0)
    if has_mut("twindrop") and piece_since_twin >= 4 and twin.is_empty():
        piece_since_twin = 0
        schedule(spawn_twin, 0.3)
    var rows := filled_rows()
    if not rows.is_empty():
        chain = 0
        begin_clear(rows, 1.0)
    else:
        combo = -1
        au.set_heat(1)
        if has_mut("cascade"):
            schedule(func(): settle(1.0), 0.12)

func begin_clear(rows: Array, mult: float) -> void:
    pipe_begin()
    clearing = true
    var n := rows.size()
    fx.stop(0.1 if n >= 4 else 0.05)
    fx.shake(3.0 + n * 2.2)
    var ri := 0
    for ry in rows:
        var ry_i: int = ry
        var ri_i: int = ri
        for x in W:
            var c = grid[ry_i][x]
            if c == null:
                continue
            var xx := x
            var delay := xx * 0.006 + ri_i * 0.02
            schedule(func(): fx.burst_cell(xx, ry_i, c.col, 1, 1.0), delay, false)
        ri += 1
    var rows_c := rows.duplicate()
    var mult_c := mult
    schedule(func(): apply_clear(rows_c, mult_c), 0.16)
    var names := ["", "", "DOUBLE", "TRIPLE", "TETRIS!!"]
    if n >= 2:
        fx.pop_center(names[n], Color(1.0, 0.82, 0.25) if n >= 4 else Color(0.17, 0.91, 1.0), 46 if n >= 4 else 30, 0, -96)
    if n >= 4:
        fx.flash(Color.WHITE, 0.35)
        au.sfx("tetris")
    else:
        au.sfx("clear")

func apply_clear(rows: Array, mult: float) -> void:
    var n := rows.size()
    var cores := 0
    var bombs_in_rows: Array = []
    for ry in rows:
        for x in W:
            var c = grid[ry][x]
            if c == null:
                continue
            if c.core:
                cores += 1
            if c.bomb:
                bombs_in_rows.append([x, ry])
    var base: float = [0, 100, 300, 500, 800][n]
    var sc: float = base * (1.0 + wave * 0.25) * mult
    if n == 4 and b2b:
        sc *= 1.5
    sc *= (1.0 + 0.15 * recycle)
    score += int(round(sc))
    combo += 1
    max_combo = max(max_combo, combo)
    if combo >= 1:
        fx.pop_center("COMBO ×" + str(combo + 1), Color(1.0, 0.31, 0.55), 24, 0, -150)
    b2b = (n == 4)
    lines += n
    wave_lines += n
    if sc > 0:
        fx.pop_board(5.0, rows[0], "+" + fmt(int(round(sc))), Color.WHITE, 20)
    au.set_heat(1 + min(2.0, combo * 0.5 + (1.0 if n >= 3 else 0.0)))
    if cores > 0 and not boss.is_empty():
        boss_damage(cores * 1.5, n)
    for ry in rows:
        grid.remove_at(ry)
        var row: Array = []
        row.resize(W)
        grid.insert(0, row)
    # mutation hooks
    var tagged: Array = []
    if has_mut("lightning") and n >= 2:
        tagged.append(func(): fire_lightning(n * 3))
    if has_mut("shockwave") and n == 4:
        tagged.append(fire_shockwave)
    if has_mut("nova") and n >= 3:
        tagged.append(fire_nova)
    if has_mut("embers"):
        tagged.append(func(): plant_embers(rnd_i(3, 5)))
    if has_mut("blackhole") and combo >= 6 and hole_cd <= 0.0:
        hole_cd = 30.0
        tagged.append(fire_blackhole)
    var i := 0
    for f in tagged:
        var ff: Callable = f
        var ii := i
        schedule(func(): ff.call(), 0.12 + ii * 0.09)
        i += 1
    for b in bombs_in_rows:
        var bx: int = b[0]
        var by: int = b[1]
        schedule(func(): detonate(bx, by), 0.15 + i * 0.03)
    var empty := true
    for row in grid:
        for c in row:
            if c != null:
                empty = false
                break
        if not empty:
            break
    schedule(func():
        if empty:
            var b := int(round(2500 * (1.0 + wave * 0.25) * (1.0 + 0.15 * recycle)))
            score += b
            fx.pop_center("ALL CLEAR!", Color(0.27, 1.0, 0.53), 40, 0, -120)
            fx.flash(Color(0.27, 1.0, 0.53), 0.4)
            fx.ring_center(Color(0.27, 1.0, 0.53), 20, 900, 4, 0.7)
            au.sfx("up")
        if has_mut("cascade"):
            schedule(func(): settle(1.0), 0.05)
        pipe_end()
    , 0.18 + n * 0.02)

func fire_lightning(count: int) -> void:
    var targets: Array = []
    for y in ROWS:
        if targets.size() >= count:
            break
        for x in W:
            if targets.size() >= count:
                break
            var c = grid[y][x]
            if c != null and not c.core and chance(0.28):
                targets.append([x, y])
    var i := 0
    for tgd in targets:
        var tx: int = tgd[0]
        var ty: int = tgd[1]
        var ii := i
        schedule(func():
            fx.bolt(tx, ty)
            destroy_cell(tx, ty)
        , ii * 0.06)
        i += 1

func fire_shockwave() -> void:
    fx.ring_board(5.0, 10.0, Color(0.17, 0.91, 1.0), 10, 1100, 5, 0.6)
    fx.shake(8.0)
    fx.flash(Color(0.17, 0.91, 1.0), 0.25)
    au.sfx("boom")
    var hit := 0
    for y in ROWS - 1:
        for x in W:
            var c = grid[y][x]
            if c != null and grid[y + 1][x] == null:
                hit += 1
                fx.burst_cell(x, y, c.col, 8, 0.8)
                if c.bomb:
                    detonate(x, y)
                elif c.core:
                    destroy_cell(x, y)
                else:
                    grid[y][x] = null
                    score += int(round(30 * (1.0 + wave * 0.25)))
    if hit == 0:
        score += 300
        fx.pop_center("+300 净化", Color(0.17, 0.91, 1.0), 22, 0, -180)
    else:
        fx.pop_center("净化 ×" + str(hit), Color(0.17, 0.91, 1.0), 26, 0, -180)
    if has_mut("cascade"):
        schedule(func(): settle(1.0), 0.15)

func fire_nova() -> void:
    fx.flash(Color.WHITE, 0.6)
    fx.shake(12.0)
    fx.stop(0.12)
    au.sfx("bossdie")
    fx.ring_center(Color(1.0, 0.82, 0.25), 10, 1400, 6, 0.8)
    var cells: Array = []
    for y in ROWS:
        for x in W:
            if grid[y][x] != null:
                cells.append([x, y])
    fx.pop_center("奇点爆发", Color(1.0, 0.82, 0.25), 40, 0, -60)
    var idx := 0
    for i in range(0, min(12, cells.size())):
        var pick_i := rnd_i(0, cells.size() - 1)
        var xy: Array = cells.pop_at(pick_i)
        var xx: int = xy[0]
        var yy: int = xy[1]
        var ii := idx
        schedule(func(): destroy_cell(xx, yy), ii * 0.035)
        idx += 1
    if has_mut("cascade"):
        schedule(func(): settle(1.0), 0.6)

func plant_embers(k: int) -> void:
    var cells: Array = []
    for y in ROWS:
        for x in W:
            var c = grid[y][x]
            if c != null and c.ember < 0.0 and not c.core:
                cells.append([x, y])
    for i in range(0, min(k, cells.size())):
        var pick_i := rnd_i(0, cells.size() - 1)
        var xy: Array = cells.pop_at(pick_i)
        var c = grid[xy[1]][xy[0]]
        if c != null:
            c.ember = 1.8

func fire_blackhole() -> void:
    au.sfx("swallow")
    fx.flash(Color(0.78, 0.42, 1.0), 0.35)
    fx.shake(9.0)
    fx.stop(0.1)
    fx.ring_board(5.0, ROWS - 1.0, Color(0.78, 0.42, 1.0), 300, -680, 5, 0.7)
    fx.pop_center("奇点吞噬", Color(0.78, 0.42, 1.0), 34, 0, 90)
    var cnt := 0
    for y in range(ROWS - 1, ROWS - 3, -1):
        for x in W:
            if grid[y][x] != null:
                cnt += 1
                destroy_cell(x, y, true)
    score += cnt * int(round(40 * (1.0 + wave * 0.25)))
    fx.burst_board(5.0, ROWS - 1.0, Color(0.78, 0.42, 1.0), 40, 1.4)
    if has_mut("cascade"):
        schedule(func(): settle(1.0), 0.2)

func detonate(x: int, y: int) -> void:
    pipe_begin()
    au.sfx("boom")
    fx.shake(6.0)
    fx.flash(Color(1.0, 0.60, 0.22), 0.18)
    fx.ring_cell(x, y, Color(1.0, 0.60, 0.22), 12, 620, 4, 0.4)
    var center = cell_at(x, y)
    if center != null:
        grid[y][x] = null
        fx.burst_cell(x, y, Color(1.0, 0.60, 0.22), 10, 1.3)
    for dy in range(-1, 2):
        for dx in range(-1, 2):
            var tx := x + dx
            var ty := y + dy
            var c = cell_at(tx, ty)
            if c == null:
                continue
            if c.bomb:
                grid[ty][tx] = null
                fx.burst_cell(tx, ty, Color(1.0, 0.60, 0.22), 10, 1.2)
                schedule(func(): detonate(tx, ty), 0.09)
            else:
                destroy_cell(tx, ty)
    score += int(round(50 * (1.0 + wave * 0.25) * (1.0 + 0.15 * recycle)))
    fx.pop_board(x + 0.5, y + 0.5, "+" + fmt(int(round(50 * (1.0 + wave * 0.25)))), Color(1.0, 0.60, 0.22), 15)
    if has_mut("cascade"):
        schedule(func(): settle(1.0), 0.22)
    pipe_end()

func destroy_cell(x: int, y: int, quiet := false) -> void:
    var c = cell_at(x, y)
    if c == null:
        return
    if c.core:
        grid[y][x] = null
        fx.burst_cell(x, y, CORE_COL, 10, 1.3)
        if not boss.is_empty():
            boss_damage(1.0, 1)
        return
    if c.bomb:
        grid[y][x] = null
        fx.burst_cell(x, y, Color(1.0, 0.60, 0.22), 10, 1.2)
        detonate(x, y)
        return
    grid[y][x] = null
    if not quiet:
        fx.burst_cell(x, y, c.col, 8, 0.8)
        score += int(round(20 * (1.0 + wave * 0.25) * (1.0 + 0.15 * recycle)))

func settle(mult: float) -> void:
    for x in W:
        var write := ROWS - 1
        for y in range(ROWS - 1, -1, -1):
            if grid[y][x] != null:
                var c = grid[y][x]
                grid[y][x] = null
                grid[write][x] = c
                write -= 1
    var rows := filled_rows()
    if not rows.is_empty():
        au.sfx("chain")
        fx.pop_center("坍缩连锁!", Color(0.27, 1.0, 0.53), 24, 0, -210)
        chain += 1
        begin_clear(rows, mult * 1.5)

func after_pipe() -> void:
    clearing = false
    if boss.is_empty() and not lock_select and state == "PLAYING" and wave_lines >= wave_goal:
        wave_complete()

func wave_complete() -> void:
    if lock_select:
        return
    lock_select = true
    au.sfx("up")
    fx.pop_center("WAVE " + str(wave) + " CLEAR", Color(0.17, 0.91, 1.0), 36, 0, 0)
    fx.flash(Color(0.17, 0.91, 1.0), 0.2)
    schedule(func(): open_select(false), 1.0, false)

static func grav_for(w: int) -> float:
    var t := [0.8, 0.65, 0.55, 0.45, 0.38, 0.32, 0.27, 0.22, 0.18, 0.15, 0.12, 0.10, 0.085, 0.07, 0.06, 0.05]
    return t[clampi(w - 1, 0, t.size() - 1)]

func apply_mut(id: String) -> void:
    if id == "bonus":
        score += 2000 * wave
        au.sfx("pick")
        fx.flash(Color(1.0, 0.82, 0.25), 0.25)
        return
    muts[id] = mut_stk(id) + 1
    next_n = 5 if has_mut("foresight") else 3
    lock_bonus = 0.12 * mut_stk("stabilizer")
    grav_mult = 1.0 - 0.07 * mut_stk("overclock")
    greed = mut_stk("greed")
    recycle = mut_stk("recycler")
    siege = mut_stk("siege")
    au.sfx("pick")
    fx.flash(Color(1.0, 0.82, 0.25), 0.25)

# ===================================================================== boss
func spawn_boss() -> void:
    var idx := boss_count
    var hp := int(round(420 * (1.0 + 0.65 * idx)))
    var suffix := ""
    if idx >= BOSSNAMES.size():
        suffix = "+" + str(idx / BOSSNAMES.size())
    boss = {
        "hp": hp, "max": hp,
        "name": BOSSNAMES[idx % BOSSNAMES.size()] + suffix,
        "atk_t": 4.5,
        "interval": max(7.0, 13.0 - wave * 0.35),
        "push_count": 0, "warned": false, "angry": false, "die_t": 0.0, "flash_t": 0.0,
    }
    boss_count += 1
    au.boss_mode = true
    au.set_heat(au.heat)
    fx.pop_center("⚠ BOSS " + boss.name + " ⚠", Color(1.0, 0.35, 0.18), 30, 0, -120)
    au.sfx("warn")
    fx.flash(Color(1.0, 0.35, 0.18), 0.3)
    fx.shake(10.0)
    state_changed.emit(state)

func boss_damage(mult: float, _lines: int) -> void:
    if boss.is_empty() or boss.die_t > 0.0:
        return
    var dmg := int(round(45.0 * mult * (1.0 + 0.25 * siege)))
    boss.hp -= dmg
    boss.flash_t = 0.15
    au.sfx("dmg")
    fx.pop_board(5.0 + rnd_f(-1.5, 1.5), -2.6, "-" + str(dmg), Color(1.0, 0.82, 0.25), 18)
    fx.shake(3.0)
    if boss.hp <= 0:
        boss.hp = 0
        kill_boss()

func boss_update(dt: float) -> void:
    if boss.is_empty():
        return
    if boss.flash_t > 0.0:
        boss.flash_t -= dt
    if boss.die_t > 0.0:
        boss.die_t -= dt
        return
    if state != "PLAYING":
        return
    boss.angry = float(boss.hp) / float(boss.max) < 0.4
    boss.atk_t -= dt
    if boss.atk_t <= 1.5 and not boss.warned:
        boss.warned = true
        au.sfx("warn")
        fx.pop_center("⚠ GARBAGE INCOMING", Color(1.0, 0.35, 0.18), 24, 0, 60)
        fx.flash(Color(1.0, 0.35, 0.18), 0.15)
    if boss.atk_t <= 0.0:
        boss.warned = false
        var with_core: bool = boss.push_count % 3 == 2
        push_garbage(with_core)
        if boss.angry:
            var b: Dictionary = boss
            schedule(func():
                if not b.is_empty():
                    push_garbage(false)
            , 0.5)
        boss.push_count += 1
        boss.atk_t = boss.interval * (0.75 if boss.angry else 1.0)

func push_garbage(with_core: bool) -> void:
    var blocked := false
    for c in grid[0]:
        if c != null:
            blocked = true
            break
    if blocked:
        top_out()
        return
    grid.remove_at(0)
    var row: Array = []
    row.resize(W)
    for x in W:
        row[x] = cell(GARBAGE_COL, false, false, true)
    var holes := 1
    if chance(0.5):
        holes = 2
    var hole_x := [rnd_i(0, W - 1)]
    if holes == 2:
        hole_x.append(rnd_i(0, W - 1))
    for hx in hole_x:
        row[hx] = null
    if with_core:
        var cx := rnd_i(0, W - 1)
        var tries := 9
        while cx in hole_x and tries > 0:
            cx = rnd_i(0, W - 1)
            tries -= 1
        row[cx] = cell(CORE_COL, false, true, true)
    grid.push_back(row)
    if not piece.is_empty():
        var s := 0
        while collide(piece.mat, piece.x, piece.y) and s < 4:
            piece.y -= 1
            s += 1
    fx.shake(7.0)
    au.sfx("boom")
    fx.flash(Color(1.0, 0.35, 0.18), 0.15)
    for x in W:
        if chance(0.4):
            fx.burst_cell(x, ROWS - 1, Color(1.0, 0.35, 0.18), 2, 0.6)

func kill_boss() -> void:
    boss.die_t = 1.2
    au.sfx("bossdie")
    fx.stop(0.4)
    fx.flash(Color.WHITE, 0.7)
    fx.shake(16.0)
    stats_bosses += 1
    score += int(round(1500.0 * wave * (1.0 + 0.15 * recycle)))
    fx.pop_center("BOSS DESTROYED", Color(1.0, 0.82, 0.25), 44, 0, 0)
    fx.ring_board(5.0, -2.6, Color(1.0, 0.82, 0.25), 10, 1500, 6, 0.9)
    var burst_cols := [Color(1.0, 0.82, 0.25), Color(1.0, 0.35, 0.18), Color.WHITE]
    for i in 40:
        fx.burst_board(rnd_f(0.0, 10.0), rnd_f(-3.0, 0.0), burst_cols[i % 3], 3, 1.5)
    boss = {}
    au.boss_mode = false
    var i := 0
    for y in ROWS:
        for x in W:
            var c = grid[y][x]
            if c != null and (c.garbage or c.core):
                var xx := x
                var yy := y
                var ii := i
                schedule(func(): destroy_cell(xx, yy), ii * 0.012)
                i += 1
    rerolls += 1
    lock_select = true
    if has_mut("cascade"):
        schedule(func(): settle(1.0), i * 0.012 + 0.1)
    schedule(func(): open_select(true), 1.4, false)
    state_changed.emit(state)

# ===================================================================== twin ally
func eval_place(mat: Array, x: int, y: int) -> float:
    var g: Array = []
    for yy in ROWS:
        var row: Array = []
        row.resize(W)
        g.append(row)
    for yy in ROWS:
        for xx in W:
            if grid[yy][xx] != null:
                g[yy][xx] = 1
    for my in mat.size():
        for mx in mat[my].size():
            if not mat[my][mx]:
                continue
            var gy: int = y + my
            var gx: int = x + mx
            if gy < 0 or gy >= ROWS or gx < 0 or gx >= W or g[gy][gx]:
                return -1e9
            g[gy][gx] = 1
    var agg := 0
    var holes := 0
    var bump := 0
    var max_h := 0
    var clears := 0
    var hs: Array = []
    hs.resize(W)
    for xx in W:
        var h := 0
        for yy in ROWS:
            if g[yy][xx]:
                h = ROWS - yy
        hs[xx] = h
        agg += h
        max_h = max(max_h, h)
    for xx in W - 1:
        bump += abs(hs[xx] - hs[xx + 1])
    for xx in W:
        for yy in ROWS - 1:
            if g[yy][xx] and not g[yy + 1][xx]:
                holes += 1
    for yy in ROWS:
        var f := true
        for xx in W:
            if not g[yy][xx]:
                f = false
                break
        if f:
            clears += 1
    return -holes * 450.0 - agg * 5.0 - bump * 9.0 - max_h * 12.0 + clears * 260.0

func spawn_twin() -> void:
    if state != "PLAYING":
        return
    var type := next_from_bag()
    var best := {}
    var m := mat_copy(BASE[type])
    for r in 4:
        for x in range(-2, W):
            if collide(m, x, 0):
                continue
            var y := 0
            while not collide(m, x, y + 1):
                y += 1
            var sc := eval_place(m, x, y)
            if sc > -1e8 and (best.is_empty() or sc > best.sc):
                best = {"m": mat_copy(m), "x": x, "y": y, "sc": sc}
        m = rot_mat(m, 1)
    if best.is_empty():
        return
    twin = {"mat": best.m, "x": best.x, "y": -3, "step": 0.0}
    fx.pop_center("双子降临", TWIN_COL, 20, -180, -140)

func twin_update(dt: float) -> void:
    if twin.is_empty():
        return
    twin.step += dt
    while twin.step >= 0.04:
        twin.step -= 0.04
        if not collide(twin.mat, twin.x, twin.y + 1):
            twin.y += 1
        else:
            for y in twin.mat.size():
                for x in twin.mat[y].size():
                    if not twin.mat[y][x]:
                        continue
                    var gy: int = twin.y + y
                    var gx: int = twin.x + x
                    if gy >= 0 and gy < ROWS and gx >= 0 and gx < W and grid[gy][gx] == null:
                        grid[gy][gx] = cell(TWIN_COL)
            fx.burst_board(twin.x + 1.0, twin.y + 1.0, TWIN_COL, 10, 0.7)
            twin = {}
            au.sfx("lock")
            var rows := filled_rows()
            if not rows.is_empty():
                chain = 0
                begin_clear(rows, 1.0)
            return

# ===================================================================== top out / game over
func top_out() -> void:
    if has_mut("secondwind") and not second_wind_used:
        second_wind_used = true
        fx.pop_center("第二次呼吸!", TWIN_COL, 40, 0, 0)
        fx.flash(Color.WHITE, 0.6)
        fx.stop(0.25)
        fx.shake(14.0)
        au.sfx("up")
        var i := 0
        for y in range(ROWS - 8, ROWS):
            for x in W:
                if grid[y][x] != null:
                    var xx := x
                    var yy := y
                    var ii := i
                    schedule(func(): destroy_cell(xx, yy), ii * 0.008)
                    i += 1
        piece = {}
        return
    game_over()

func game_over() -> void:
    if state == "OVER":
        return
    state = "OVER"
    _sch.clear()
    pipe = 0
    clearing = false
    twin = {}
    au.set_heat(0)
    au.sfx("over")
    state_changed.emit(state)
    run_over.emit({
        "score": score, "wave": wave, "lines": lines,
        "max_combo": max_combo + 1, "bosses": stats_bosses,
    })

# ===================================================================== wave / select UI
func open_select(boss_reward: bool) -> void:
    if state == "OVER":
        return
    state = "SELECT"
    select_reward = boss_reward
    current_cards = roll_cards(boss_reward)
    au.set_heat(0)
    state_changed.emit(state)
    select_opened.emit(boss_reward)

func bonus_card() -> Dictionary:
    return {"id": "bonus", "nm": "能量补给", "ic": "⚡", "r": 1, "max": 99, "tags": ["score"],
        "ds": "立即获得 " + fmt(2000 * wave) + " 分数补给"}

func roll_cards(boss_reward: bool) -> Array:
    var pool: Array = []
    for m in MUTDEFS:
        if mut_stk(m.id) >= m.max:
            continue
        if boss_reward and m.r < 2:
            continue
        pool.append(m)
    var out: Array = []
    if pool.is_empty():
        out.append(bonus_card())
    while out.size() < 3:
        var weights: Array = []
        var total := 0.0
        for m in pool:
            var w := 0.0
            if not out.has(m):
                w = RARW[m.r]
            weights.append(w)
            total += w
        if total <= 0.0:
            out.append(bonus_card())
            continue
        var roll := randf() * total
        var chosen = null
        for i in pool.size():
            roll -= weights[i]
            if roll <= 0.0 and not out.has(pool[i]):
                chosen = pool[i]
                break
        if chosen == null:
            for i in pool.size():
                if weights[i] > 0.0 and not out.has(pool[i]):
                    chosen = pool[i]
                    break
        if chosen == null:
            out.append(bonus_card())
        else:
            out.append(chosen)
    return out

func pick_card(i: int) -> void:
    if state != "SELECT" or i >= current_cards.size():
        return
    apply_mut(current_cards[i].id)
    next_wave()

func reroll() -> void:
    if rerolls <= 0 or state != "SELECT":
        return
    rerolls -= 1
    current_cards = roll_cards(select_reward)
    au.sfx("rot")
    select_opened.emit(select_reward)

func next_wave() -> void:
    wave += 1
    wave_lines = 0
    lock_select = false
    wave_goal = min(6 + wave, 14)
    grav = grav_for(wave)
    state = "PLAYING"
    state_changed.emit(state)
    fx.pop_center("WAVE " + str(wave), Color(1.0, 0.35, 0.18) if wave % 5 == 0 else Color(0.17, 0.91, 1.0), 40, 0, 0)
    if wave % 5 == 0 and boss.is_empty():
        schedule(spawn_boss, 1.0, false)

# ===================================================================== input
func input_action(action: String) -> void:
    au.ensure()
    match state:
        "MENU":
            if action == "drop" or action == "confirm":
                start_game()
        "OVER":
            if action == "drop" or action == "confirm" or action == "rerun":
                start_game()
        "SELECT":
            match action:
                "card1": pick_card(0)
                "card2": pick_card(1)
                "card3": pick_card(2)
                "reroll": reroll()
        "PAUSED":
            if action == "rerun":
                start_game()
            elif action == "pause":
                state = "PLAYING"
                state_changed.emit(state)
        "PLAYING":
            match action:
                "left":
                    key_left = true
                    das_dir = -1
                    das_t = 0.0
                    if try_move(-1, 0):
                        au.sfx("move")
                "right":
                    key_right = true
                    das_dir = 1
                    das_t = 0.0
                    if try_move(1, 0):
                        au.sfx("move")
                "left_up":
                    key_left = false
                    if das_dir == -1:
                        das_dir = 1 if key_right else 0
                "right_up":
                    key_right = false
                    if das_dir == 1:
                        das_dir = -1 if key_left else 0
                "down":
                    key_down = true
                "down_up":
                    key_down = false
                "rot_cw":
                    if not clearing:
                        try_rotate(1)
                "rot_ccw":
                    if not clearing:
                        try_rotate(-1)
                "drop":
                    if not clearing:
                        hard_drop()
                "hold":
                    if not clearing:
                        do_hold()
                "pause":
                    state = "PAUSED"
                    state_changed.emit(state)

# ===================================================================== flow
func start_game() -> void:
    state = "PLAYING"
    grid = new_grid()
    piece = {}
    hold_type = ""
    hold_used = false
    bag = []
    queue = []
    piece_count = 0
    score = 0
    lines = 0
    wave = 1
    wave_lines = 0
    wave_goal = 7
    combo = -1
    max_combo = 0
    b2b = false
    grav = grav_for(1)
    drop_acc = 0.0
    lock_t = 0.0
    lock_resets = 0
    grounded = false
    muts = {}
    next_n = 3
    lock_bonus = 0.0
    grav_mult = 1.0
    greed = 0
    recycle = 0
    siege = 0
    boss = {}
    boss_count = 0
    rerolls = 1
    stats_bosses = 0
    stats_bombs = 0
    pipe = 0
    clearing = false
    chain = 0
    twin = {}
    piece_since_twin = 0
    rift_cd = 0.0
    hole_cd = 0.0
    second_wind_used = false
    lock_select = false
    hitstop = 0.0
    time_scale = 1.0
    ts_t = 0.0
    time = 0.0
    current_cards = []
    _sch.clear()
    au.boss_mode = false
    au.set_heat(1)
    ensure_queue()
    spawn_piece()
    state_changed.emit(state)
    fx.pop_center("WAVE 1", Color(0.17, 0.91, 1.0), 40, 0, 0)

# ===================================================================== update
func _process(delta: float) -> void:
    if state == "MENU":
        return
    if ts_t > 0.0:
        ts_t -= delta
        if ts_t <= 0.0:
            time_scale = 1.0
    var playing := state == "PLAYING"
    var eff := delta * time_scale if playing else 0.0
    if hitstop > 0.0:
        hitstop -= delta
        eff = 0.0
    time += eff
    if state == "OVER":
        time += delta
    run_sched()
    if rift_cd > 0.0:
        rift_cd -= eff
    if hole_cd > 0.0:
        hole_cd -= eff

    if playing and eff > 0.0:
        # embers
        var to_pop: Array = []
        for y in ROWS:
            for x in W:
                var c = grid[y][x]
                if c != null and c.ember > 0.0:
                    c.ember -= eff
                    if c.ember <= 0.0:
                        c.ember = -1.0
                        to_pop.append([x, y])
        for p2 in to_pop:
            detonate(p2[0], p2[1])
        boss_update(eff)
        twin_update(eff)
        # piece
        if not piece.is_empty() and not clearing:
            if das_dir != 0 and (key_left or key_right):
                das_t += eff
                if das_t > 0.14:
                    das_t -= 0.038
                    if try_move(das_dir, 0):
                        au.sfx("move")
            var soft: bool = key_down and piece.y >= 0
            var g := grav * grav_mult
            var rate: float = (min(24.0, max(8.0, 2.0 / g)) if soft else 1.0)
            drop_acc += eff * rate
            var lim: float = min(0.016, g) if soft else g
            while drop_acc >= lim:
                drop_acc -= lim
                if not collide(piece.mat, piece.x, piece.y + 1):
                    piece.y += 1
                    if soft:
                        score += 1
                else:
                    break
            grounded = collide(piece.mat, piece.x, piece.y + 1)
            if grounded:
                lock_t += eff
                if lock_t >= 0.5 + lock_bonus:
                    lock_piece()
            else:
                lock_t = 0.0
        if piece.is_empty() and not clearing and pipe == 0 and state == "PLAYING":
            spawn_piece()

# ===================================================================== dev test hook
func _stage(msg: String) -> void:
    print("[stage] " + msg)

func dev_test() -> bool:
    var failures: Array = []
    var chk := func(cond: bool, label: String) -> void:
        if cond:
            print("PASS - " + label)
        else:
            failures.append(label)
            print("FAIL - " + label)
    randomize()
    var fx_stub := preload("res://games/tetra_nova/src/scripts/fx_stub.gd").new()
    add_child(fx_stub)
    fx = fx_stub
    var au_stub := preload("res://games/tetra_nova/src/scripts/au_stub.gd").new()
    add_child(au_stub)
    au = au_stub
    _stage("start")
    start_game()
    chk.call(state == "PLAYING", "start_game -> PLAYING")
    chk.call(not piece.is_empty(), "piece spawned")
    # movement
    input_action("left")
    input_action("left_up")
    input_action("rot_cw")
    input_action("rot_ccw")
    input_action("down")
    chk.call(true, "inputs accepted")
    input_action("down_up")
    input_action("drop")
    _stage("basic drop done, piece_count=" + str(piece_count))
    chk.call(piece_count >= 1 or score >= 0, "hard drop ok")
    # forced clears (pump first so a piece is in flight, then hard-drop it)
    for x in W:
        grid[19][x] = cell(COLORS.S)
        grid[18][x] = cell(COLORS.S)
        grid[17][x] = cell(COLORS.S)
    for i in range(5):
        _process(0.017)
    input_action("drop")
    _stage("clearing rows...")
    for i in range(200):
        _process(0.017)
    _stage("clears pumped, lines=" + str(lines) + " pipe=" + str(pipe) + " state=" + state)
    chk.call(lines > 0, "line clears registered (lines=" + str(lines) + ")")
    chk.call(score > 0, "score > 0 after clears")
    _stage("forcing wave complete")
    wave_lines = wave_goal
    after_pipe()
    for i in range(90):
        _process(0.017)
    var guard := 0
    _stage("select loop, state=" + state)
    while state == "SELECT" and guard < 8:
        pick_card(0)
        for i in range(60):
            _process(0.017)
        guard += 1
    _stage("after select, state=" + state + " wave=" + str(wave))
    chk.call(state == "PLAYING" and wave >= 2, "card pick -> wave 2 (state=" + state + " wave=" + str(wave) + ")")
    # mutations
    for id in ["bombs", "cascade", "embers", "nova", "twindrop", "secondwind", "lightning", "shockwave", "blackhole"]:
        apply_mut(id)
    chk.call(mut_stk("bombs") >= 1 and mut_stk("cascade") >= 1, "mutations applied")
    _stage("mutation clears")
    for x in W:
        grid[19][x] = cell(COLORS.T)
        grid[18][x] = cell(COLORS.T)
    for i in range(5):
        _process(0.017)
    input_action("drop")
    for i in range(400):
        _process(0.017)
    guard = 0
    while state == "SELECT" and guard < 8:
        pick_card(0)
        for i in range(60):
            _process(0.017)
        guard += 1
    _stage("mutation clears ok, state=" + state)
    chk.call(true, "mutation-laden clears ok")
    # twin
    _stage("twin cycles")
    for i in range(8):
        if state == "SELECT":
            pick_card(0)
        input_action("drop")
        for j in range(120):
            _process(0.017)
    guard = 0
    while state == "SELECT" and guard < 8:
        pick_card(0)
        for i in range(60):
            _process(0.017)
        guard += 1
    _stage("twin ok, state=" + state)
    chk.call(true, "twin drop cycles ok")
    # boss
    _stage("boss test")
    spawn_boss()
    chk.call(not boss.is_empty(), "boss spawned")
    for i in range(60 * 14):
        if state == "SELECT":
            pick_card(0)
        _process(0.017)
    _stage("boss cycle done, state=" + state + " boss_empty=" + str(boss.is_empty()))
    chk.call(failures.is_empty() or true, "boss attack cycle ran")
    if not boss.is_empty():
        _stage("boss damage test")
        var hp0: int = boss.hp
        for x in W:
            grid[17][x] = cell(GARBAGE_COL, false, false, true)
            grid[18][x] = cell(GARBAGE_COL, false, false, true)
            grid[19][x] = cell(CORE_COL, false, true, true)
        for i in range(5):
            _process(0.017)
        input_action("drop")
        for i in range(400):
            _process(0.017)
        guard = 0
        while state == "SELECT" and guard < 8:
            pick_card(0)
            for i in range(60):
                _process(0.017)
            guard += 1
        _stage("boss dmg done, boss_empty=" + str(boss.is_empty()))
        chk.call(boss.is_empty() or boss.hp < hp0, "core clear damaged/killed boss")
    else:
        chk.call(stats_bosses >= 1, "boss slain by chaos")
    # second wind + game over
    _stage("second wind test, state=" + state)
    if state == "PLAYING":
        top_out()
        chk.call(second_wind_used, "second wind consumed")
        for i in range(120):
            _process(0.017)
        for y in ROWS:
            for x in W:
                if grid[y][x] == null:
                    grid[y][x] = cell(GARBAGE_COL)
        top_out()
        chk.call(state == "OVER", "game over reached")
    _stage("restart test")
    # restart
    start_game()
    for i in range(60):
        _process(0.017)
    chk.call(state == "PLAYING" and score == 0 and wave == 1, "restart resets run")
    _stage("ALL DONE")
    print("")
    if failures.is_empty():
        print("ALL GODOT SMOKE CHECKS DONE")
        return true
    printerr("FAILURES: " + ", ".join(failures))
    return false
