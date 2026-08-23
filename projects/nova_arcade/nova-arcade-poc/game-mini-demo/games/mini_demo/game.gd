extends Node
## PoC DLC game: implements the GameModule contract.
## boot(ctx) -> play ~1s -> write save to ctx.save_dir, tmp to ctx.tmp_dir
## -> quit_requested(result)

signal quit_requested(result: Dictionary)

var _ctx: Dictionary = {}
var _t := 0.0
var _clicks := 0
var _done := false

func boot(ctx: Dictionary) -> void:
    _ctx = ctx
    print("[mini_demo] boot: id=", ctx.game_id, " best=", ctx.best,
        " save_dir=", ctx.save_dir)

func _process(delta: float) -> void:
    if _done:
        return
    _t += delta
    _clicks += 1
    if _t >= 1.0:
        _quit()

func _quit() -> void:
    _done = true
    var score := _clicks * 10
    # persist into the game's OWN save dir (injected by shell)
    var f := FileAccess.open(_ctx.save_dir + "save.json", FileAccess.WRITE)
    if f:
        f.store_string(JSON.stringify({"best": score, "sessions": 1}))
        f.close()
    # scratch file in the shell-managed tmp dir
    var f2 := FileAccess.open(_ctx.tmp_dir + "cache.bin", FileAccess.WRITE)
    if f2:
        f2.store_64(score)
        f2.close()
    print("[mini_demo] quit: score=", score, " playtime=", _t)
    quit_requested.emit({
        "score": score,
        "playtime": _t,
        "achievements": ["demo_first_run"],
    })
