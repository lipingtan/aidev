extends Node
## DLC host shell PoC:
##   mount PCK -> instantiate game scene -> boot(ctx) -> game runs ->
##   quit_requested(result) -> free game -> verify save_dir isolation -> back to shell
##
## Run:  godot --headless --path . -- <abs_path_to_pck>

const GAME_ID := "mini_demo"

var game_host: Node = null
var pck_path := ""
var _saved_best := 0

func _ready() -> void:
    for a in OS.get_cmdline_user_args():
        pck_path = a
    print("[shell] booted; pck=", pck_path)
    await get_tree().process_frame
    launch(GAME_ID)

# ---------------------------------------------------------------- launch
func launch(gid: String) -> void:
    # 1. mount the pack (replace_files=false: never overwrite shell resources)
    var mounted := ProjectSettings.load_resource_pack(pck_path, false)
    print("[shell] pack mounted: ", mounted)

    # 2. entry scene from the pack's reserved prefix dir
    var scene_path := "res://games/%s/game.tscn" % gid
    if not ResourceLoader.exists(scene_path):
        print("[shell] FAIL: entry scene not found: ", scene_path)
        _finish(false)
        return
    var scene: PackedScene = load(scene_path)
    game_host = scene.instantiate()
    add_child(game_host)

    # 3. lifecycle wiring
    game_host.quit_requested.connect(_on_game_quit)

    # 4. storage isolation: each game gets its own save_dir / tmp_dir
    var save_dir := "user://saves/%s/" % gid
    var tmp_dir := "user://tmp/%s/" % gid
    DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path(save_dir))
    DirAccess.make_dir_recursive_absolute(ProjectSettings.globalize_path(tmp_dir))

    # pass previous best through (read from the game's own save file)
    _saved_best = _read_best(save_dir)

    game_host.boot({
        "game_id": gid,
        "save_dir": save_dir,
        "tmp_dir": tmp_dir,
        "trial_mode": true,
        "owned": false,
        "best": _saved_best,
        "shell": self,
    })
    print("[shell] game booted (save_dir=", save_dir, " best=", _saved_best, ")")

func _read_best(save_dir: String) -> int:
    var path := save_dir + "save.json"
    if not FileAccess.file_exists(path):
        return 0
    var f := FileAccess.open(path, FileAccess.READ)
    if f == null:
        return 0
    var data: Dictionary = JSON.parse_string(f.get_as_text())
    f.close()
    return int(data.get("best", 0))

# ---------------------------------------------------------------- quit protocol
func _on_game_quit(result: Dictionary) -> void:
    print("[shell] quit_requested received: ", result)
    var save_path := "user://saves/%s/save.json" % GAME_ID
    print("[shell] game save file exists: ", FileAccess.file_exists(save_path))
    print("[shell] tmp file exists (before cleanup): ",
        FileAccess.file_exists("user://tmp/%s/cache.bin" % GAME_ID))

    # free the game scene tree — textures/audio owned by it are released
    game_host.queue_free()
    game_host = null

    # shell-owned tmp cleanup
    _clear_dir("user://tmp/%s" % GAME_ID)
    print("[shell] tmp cleaned: ",
        not FileAccess.file_exists("user://tmp/%s/cache.bin" % GAME_ID))
    print("[shell] score upload: +", result.get("score", 0),
        " playtime=", result.get("playtime", 0.0),
        " achievements=", result.get("achievements", []))
    print("[shell] BACK TO SHELL — full DLC lifecycle complete")
    _finish(true)

func _clear_dir(path: String) -> void:
    var d := DirAccess.open(path)
    if d == null:
        return
    for fn in d.get_files():
        d.remove(fn)

func _finish(ok: bool) -> void:
    await get_tree().process_frame
    get_tree().quit(0 if ok else 1)
