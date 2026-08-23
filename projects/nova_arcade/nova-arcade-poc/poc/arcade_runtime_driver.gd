extends Node
## Arcade runtime PoC — validates the full shell-side orchestration of the
## arcade runner with a MOCK core (libretro API shape, replaces with MyOsdHost
## on real MAME binding):
##   1. RunnerRegistry: meta.runtime dispatches to the right runner
##   2. ROM package: zip + sha256 + romset member validation
##   3. MockLibretroCore: load/run/video/audio/serialize API surface
##   4. ArcadeRunner lifecycle: launch -> frames -> combo quit ->
##      savestate/nvram to save_dir -> hi-score extract -> quit_requested -> tmp clean
##   5. Second launch resumes from savestate; storage isolation per gid
## Real MAME core binding (GDExtension, myosd C ABI,参照 mjarch3) is milestone M5;
## this proves the shell orchestration (dispatch/validation/lifecycle/isolation).

var failures: Array = []

func ok(cond: bool, label: String) -> void:
    print((("PASS - " if cond else "FAIL - ") + label))
    if not cond:
        failures.append(label)

func _ready() -> void:
    await get_tree().process_frame
    await _run()
    await get_tree().process_frame
    print("")
    if failures.is_empty():
        print("ALL ARCADE RUNTIME POC DONE")
        get_tree().quit(0)
    else:
        printerr("FAILURES: " + ", ".join(failures))
        get_tree().quit(1)

# ============================================================ 1. registry
func _run() -> void:
    var registry := RunnerRegistry.new()
    ok(registry.pick("pck") is PckRunnerShim, "registry: pck -> PckRunner")
    ok(registry.pick("html") is HtmlRunnerShim, "registry: html -> HtmlRunner")
    ok(registry.pick("arcade") is ArcadeRunner, "registry: arcade -> ArcadeRunner")
    ok(registry.pick("unknown") == null, "registry: unknown runtime -> null")

    # ============================================================ 2. rom package
    var rom_zip := "user://downloads/pacman.zip"
    DirAccess.make_dir_recursive_absolute("user://downloads")
    var zp := ZIPPacker.new()
    zp.open(rom_zip)
    zp.start_file("pacman/pacman.5e")   # fake romset members
    zp.write_file(PackedByteArray([0xE3, 0x88, 0x81, 0x01]))
    zp.close_file()
    zp.start_file("pacman/pacman.5f")
    zp.write_file(PackedByteArray([0x11, 0x22, 0x33]))
    zp.close_file()
    zp.start_file("readme.txt")          # not in manifest -> must be allowed (extra ok)
    zp.write_file("homebrew info".to_utf8_buffer())
    zp.close_file()
    zp.close()

    var sha := FileAccess.get_sha256(rom_zip)
    var meta := {
        "id": "pacman", "runtime": "arcade", "core": "mame",
        "rom": "pacman.zip", "sha256": sha, "romset_version": "0.275",
        "members": ["pacman/pacman.5e", "pacman/pacman.5f"],
        "hiscore_offset": 0x03B0,
    }
    ok(sha.length() == 64, "rom sha256 computed (%d chars)" % sha.length())

    var v := RomValidator.new()
    ok(v.validate(rom_zip, meta) == RomValidator.OK, "rom validation passes (sha+members)")
    # corrupt: wrong sha
    var m2 := meta.duplicate()
    m2.sha256 = "deadbeef"
    ok(v.validate(rom_zip, m2) == RomValidator.ERR_SHA, "rom validation catches bad sha256")
    # corrupt: missing member
    var m3 := meta.duplicate()
    m3.members = ["pacman/pacman.5e", "pacman/MISSING.rom"]
    ok(v.validate(rom_zip, m3) == RomValidator.ERR_MEMBERS, "rom validation catches missing romset member")

    # ============================================================ 3+4. runner lifecycle
    var ctx := _mk_ctx("pacman")
    var runner := ArcadeRunner.new()
    runner.mock_core_mode = true
    add_child(runner)
    var got_result := {}   # dict: lambda mutates contents (captures are by-value)
    runner.quit_requested.connect(func(r: Dictionary) -> void: got_result["res"] = r)

    runner.launch(ctx, rom_zip, meta)
    ok(runner.core_loaded, "mock core retro_load_game succeeded")
    ok(runner.frames_run == 0, "core idle before first process")

    # drive 180 frames (~3s at 60fps)
    for i in 180:
        runner._process(1.0 / 60.0)
    ok(runner.frames_run == 180, "retro_run called per frame (%d)" % runner.frames_run)
    ok(runner.video_updates >= 170, "video callback fired per frame (%d)" % runner.video_updates)
    ok(runner.audio_batches >= 170, "audio sample_batch fired per frame (%d)" % runner.audio_batches)

    # player "scores" via mocked input state writer
    runner.mock_set_input(3, 1)   # coins+start, drives mock core state
    for i in 60:
        runner._process(1.0 / 60.0)

    # combo quit: Start+Select+L
    runner.mock_combo_quit()
    var res: Dictionary = got_result.get("res", {})
    ok(not res.is_empty(), "quit_requested emitted on combo")
    ok(res.get("score", -2) is int and res.get("score", -2) >= 0,
        "result.score extracted from hi-score memory (%s)" % str(res.get("score")))
    ok(absf(float(res.get("playtime", 0.0)) - 4.0) < 0.2,
        "result.playtime ~= 4.0s (%.2f)" % float(res.get("playtime", 0.0)))
    ok(not res.get("achievements", []).is_empty(), "achievements reported")

    # savestate persisted in the game's save_dir
    var state_path: String = ctx.save_dir + "states/slot0.state"
    ok(FileAccess.file_exists(state_path), "savestate persisted to save_dir")
    ok(FileAccess.file_exists(ctx.save_dir + "nvram/pacman.nv"), "nvram persisted to save_dir")
    # storage isolation: another game's dir must NOT receive files
    var ctx2 := _mk_ctx("other_game")
    ok(not FileAccess.file_exists(ctx2.save_dir + "states/slot0.state"), "storage isolated per gid")
    # tmp cleaned on quit
    ok(not FileAccess.file_exists(ctx.tmp_dir + "pacman.zip"), "tmp rom copy cleaned on quit")

    # ============================================================ 5. resume from savestate
    var ctx3 := _mk_ctx("pacman")
    var runner2 := ArcadeRunner.new()
    runner2.mock_core_mode = true
    add_child(runner2)
    runner2.launch(ctx3, rom_zip, meta)
    ok(runner2.resumed_from_state, "second launch resumed from savestate")
    ok(runner2.mock_credits == runner.mock_credits, "game state restored (credits %d)" % runner2.mock_credits)
    runner2.mock_combo_quit()

func _mk_ctx(gid: String) -> Dictionary:
    var save_dir := "user://saves/%s/" % gid
    var tmp_dir := "user://tmp/%s/" % gid
    DirAccess.make_dir_recursive_absolute(save_dir + "states")
    DirAccess.make_dir_recursive_absolute(save_dir + "nvram")
    DirAccess.make_dir_recursive_absolute(tmp_dir)
    return {"game_id": gid, "save_dir": save_dir, "tmp_dir": tmp_dir}

# ============================================================ registry + runner shims
class RunnerRegistry:
    func pick(runtime: String) -> Node:
        match runtime:
            "pck": return PckRunnerShim.new()
            "html": return HtmlRunnerShim.new()
            "arcade": return ArcadeRunner.new()
            _: return null

class PckRunnerShim extends Node:
    pass

class HtmlRunnerShim extends Node:
    pass

# ============================================================ rom validator
class RomValidator:
    const OK := 0
    const ERR_SHA := 1
    const ERR_MEMBERS := 2
    const ERR_IO := 3

    func validate(zip_path: String, meta: Dictionary) -> int:
        var sha := FileAccess.get_sha256(zip_path)
        if sha != String(meta.get("sha256", "")):
            return ERR_SHA
        var zr := ZIPReader.new()
        if zr.open(zip_path) != OK:
            return ERR_IO
        var names: PackedStringArray = zr.get_files()
        zr.close()
        for m in meta.get("members", []):
            if not names.has(m):
                return ERR_MEMBERS
        return OK

# ============================================================ mock libretro core
## Simulates the retro_* API surface for PoC; real core (MAME 0.288 via myosd
## GDExtension,参照 mjarch3) replaces this class 1:1 behind the same calls.
class MockLibretroCore:
    signal video_frame(framebuffer: PackedByteArray, w: int, h: int)
    signal audio_batch(pcms: PackedVector2Array)
    signal input_polled()

    var loaded := false
    var credits := 0
    var hiscore_mem := PackedByteArray()
    var serialized_state := PackedByteArray()
    var _rand := RandomNumberGenerator.new()

    func retro_load_game(rom_path: String, _meta: Dictionary) -> bool:
        hiscore_mem.resize(0x1000)
        hiscore_mem.fill(0)
        loaded = FileAccess.file_exists(rom_path)
        return loaded

    func retro_run() -> void:
        if not loaded:
            return
        # game logic: credits accumulate score in hi-score memory area
        var off := 0x03B0
        var sc := credits * 100 + 555
        hiscore_mem.encode_s32(off, sc)
        video_frame.emit(hiscore_mem.slice(off, off + 256), 16, 16)
        var audio := PackedVector2Array()
        audio.resize(2)
        audio_batch.emit(audio)
        input_polled.emit()

    func retro_input_state() -> int:
        return credits

    func retro_serialize() -> PackedByteArray:
        var buf := PackedByteArray()
        buf.append(credits)
        buf.append_array(hiscore_mem)
        return buf

    func retro_unserialize(buf: PackedByteArray) -> void:
        if buf.size() < 1:
            return
        credits = buf[0]
        hiscore_mem = buf.slice(1)

    func retro_get_hiscore(offset: int) -> int:
        if offset + 4 > hiscore_mem.size():
            return -1
        return hiscore_mem.decode_s32(offset)

# ============================================================ arcade runner
class ArcadeRunner extends Node:
    signal quit_requested(result: Dictionary)

    var mock_core_mode := false
    var core_loaded := false
    var frames_run := 0
    var video_updates := 0
    var audio_batches := 0
    var resumed_from_state := false
    var mock_credits := 0

    var _ctx: Dictionary = {}
    var _meta: Dictionary = {}
    var _core: MockLibretroCore = null
    var _t := 0.0
    var _quit_armed := false

    func launch(ctx: Dictionary, rom_zip: String, meta: Dictionary) -> void:
        _ctx = ctx
        _meta = meta
        # work copy of rom into tmp (real cores mmap-read from disk)
        DirAccess.make_dir_recursive_absolute(ctx.tmp_dir)
        var dst: String = ctx.tmp_dir + String(meta.rom)
        DirAccess.copy_absolute(rom_zip, dst)

        _core = MockLibretroCore.new()
        core_loaded = _core.retro_load_game(dst, meta)
        _core.video_frame.connect(func(_fb, _w, _h) -> void: video_updates += 1)
        _core.audio_batch.connect(func(_p) -> void: audio_batches += 1)

        # resume: real RetroHost calls retro_unserialize with the state file
        var state_path: String = ctx.save_dir + "states/slot0.state"
        if FileAccess.file_exists(state_path):
            var f := FileAccess.open(state_path, FileAccess.READ)
            _core.retro_unserialize(f.get_buffer(f.get_length()))
            f.close()
            resumed_from_state = true
            mock_credits = _core.credits

    func mock_set_input(_port: int, credits: int) -> void:
        _core.credits = credits
        mock_credits = credits

    func mock_combo_quit() -> void:
        if _quit_armed:
            return
        _quit_armed = true
        # pause core -> autosave -> emit result
        var state := _core.retro_serialize()
        var sp: String = _ctx.save_dir + "states/slot0.state"
        var f := FileAccess.open(sp, FileAccess.WRITE)
        f.store_buffer(state)
        f.close()
        # nvram
        var nv: String = _ctx.save_dir + "nvram/%s.nv" % String(_meta.get("rom", "game").get_basename())
        var f2 := FileAccess.open(nv, FileAccess.WRITE)
        f2.store_32(_core.retro_get_hiscore(int(_meta.get("hiscore_offset", 0))))
        f2.close()
        var achievements: Array = []
        if frames_run > 100:
            achievements.append("arc_first_credit")
        quit_requested.emit({
            "score": _core.retro_get_hiscore(int(_meta.get("hiscore_offset", 0))),
            "playtime": _t,
            "achievements": achievements,
        })
        # cleanup tmp rom copy
        var tmp_rom: String = _ctx.tmp_dir + String(_meta.rom)
        if FileAccess.file_exists(tmp_rom):
            DirAccess.remove_absolute(tmp_rom)

    func _process(delta: float) -> void:
        if not core_loaded or _quit_armed:
            return
        _t += delta
        _core.retro_run()
        frames_run += 1
