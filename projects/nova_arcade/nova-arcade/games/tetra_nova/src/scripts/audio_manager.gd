extends Node
## Procedural audio: all SFX + music synthesized as PCM at load time. No assets.

const SR := 22050

var ctx_ok := true
var muted := false
var boss_mode := false
var heat := 0
var bpm := 108.0

var _sfx: Dictionary = {}
var _inst: Dictionary = {}
var _players: Array = []
var _pi := 0
var _music_players: Array = []
var _mi := 0

# sequencer
var _step := 0
var _next_t := 0.0
var _acc := 0.0

func ensure() -> void:
    if not _players.is_empty():
        return
    _build_all()

func _ready() -> void:
    set_process(false)

func start_music() -> void:
    ensure()
    _step = 0
    _next_t = 0.0
    set_process(true)

func stop_music() -> void:
    set_process(false)

# ===================================================================== synth core
static func _wave(phase: float, type: int) -> float:
    match type:
        0: return sin(phase)                     # sine
        1: return 1.0 if fmod(phase, TAU) < PI else -1.0   # square
        2:                                       # saw
            var p := fmod(phase, TAU)
            return (p / PI - 1.0)
        3:                                       # triangle
            var p2 := fmod(phase, TAU)
            return (2.0 / PI) * abs(p2 - PI) - 1.0 if p2 < PI else 1.0 - (2.0 / PI) * abs(p2 - PI)
        _: return sin(phase)

func _tone_bytes(f0: float, f1: float, dur: float, type: int, vol: float, atk := 0.005) -> PackedFloat32Array:
    var n := int(dur * SR)
    var out := PackedFloat32Array()
    out.resize(n)
    var phase := 0.0
    var ratio := 1.0
    if f1 > 0.0 and f0 > 0.0 and dur > 0.0:
        ratio = pow(f1 / f0, 1.0 / dur)
    for i in n:
        var t := float(i) / SR
        var f := f0 * pow(ratio, t)
        phase += TAU * f / SR
        var env := minf(1.0, t / atk) * pow(10.0, -4.0 * t / dur)
        out[i] = _wave(phase, type) * env * vol
    return out

func _noise_bytes(dur: float, kind: int, f0: float, f1: float, vol: float) -> PackedFloat32Array:
    var n := int(dur * SR)
    var out := PackedFloat32Array()
    out.resize(n)
    var y := 0.0
    var prev := 0.0
    var ratio := 1.0
    if f1 > 0.0 and f0 > 0.0 and dur > 0.0:
        ratio = pow(f1 / f0, 1.0 / dur)
    var rng := RandomNumberGenerator.new()
    rng.seed = 12345
    for i in n:
        var t := float(i) / SR
        var f := f0 * pow(ratio, t)
        var x := rng.randf() * 2.0 - 1.0
        var a := 1.0 - exp(-TAU * f / SR)
        if kind == 0:      # lowpass
            y += a * (x - y)
        else:              # highpass
            y = x - prev
            prev += a * (x - prev)
        var env := pow(10.0, -4.0 * t / dur)
        out[i] = clampf(y, -1.0, 1.0) * env * vol
    return out

func _mix(dst: PackedFloat32Array, src: PackedFloat32Array, at: float, gain := 1.0) -> void:
    var off := int(at * SR)
    for i in src.size():
        var j := i + off
        if j >= 0 and j < dst.size():
            dst[j] += src[i] * gain

func _to_stream(data: PackedFloat32Array) -> AudioStreamWAV:
    var bytes := PackedByteArray()
    bytes.resize(data.size() * 2)
    for i in data.size():
        var v := int(clampf(data[i], -1.0, 1.0) * 32767.0)
        bytes.encode_s16(i * 2, v)
    var wav := AudioStreamWAV.new()
    wav.format = AudioStreamWAV.FORMAT_16_BITS
    wav.mix_rate = SR
    wav.stereo = false
    wav.data = bytes
    return wav

func _mk(name: String, build: Callable) -> void:
    var buf: PackedFloat32Array = build.call()
    _sfx[name] = _to_stream(buf)

# ===================================================================== library
func _build_all() -> void:
    if not _players.is_empty():
        return
    # --- one-shots ---
    _mk("move", func(): return _tone_bytes(520.0, 520.0, 0.03, 1, 0.05))
    _mk("rot", func(): return _tone_bytes(660.0, 880.0, 0.05, 1, 0.07))
    _mk("hold", func(): return _tone_bytes(440.0, 660.0, 0.08, 3, 0.10))
    _mk("lock", func():
        var d := _noise_bytes(0.08, 0, 400.0, 400.0, 0.22)
        _mix(d, _tone_bytes(120.0, 60.0, 0.09, 0, 0.25), 0.0)
        return d)
    _mk("hard", func():
        var d := _noise_bytes(0.12, 0, 900.0, 200.0, 0.30)
        _mix(d, _tone_bytes(200.0, 70.0, 0.10, 0, 0.30), 0.0)
        return d)
    for n in range(2, 5):
        var nn := n
        _mk("clear" + str(n), func():
            var base := 440.0 * pow(1.22, nn)
            var d := PackedFloat32Array()
            d.resize(int(0.4 * SR))
            for i in 4:
                _mix(d, _tone_bytes(base * pow(1.19, i), 0.0, 0.12, 1, 0.09), i * 0.03)
            _mix(d, _noise_bytes(0.25, 0, 2400.0, 300.0, 0.22), 0.0)
            return d)
    _mk("tetris", func():
        var d := PackedFloat32Array()
        d.resize(int(0.9 * SR))
        _mix(d, _noise_bytes(0.5, 0, 3000.0, 120.0, 0.40), 0.0)
        _mix(d, _tone_bytes(90.0, 34.0, 0.55, 0, 0.50), 0.0)
        for i in 6:
            _mix(d, _tone_bytes(330.0 * pow(1.26, i), 0.0, 0.16, 2, 0.08), i * 0.055)
        return d)
    _mk("boom", func():
        var d := _noise_bytes(0.45, 0, 1400.0, 60.0, 0.42)
        _mix(d, _tone_bytes(110.0, 30.0, 0.40, 0, 0.45), 0.0)
        return d)
    _mk("zap", func():
        var d := _noise_bytes(0.14, 1, 1800.0, 1800.0, 0.28)
        _mix(d, _tone_bytes(1900.0, 200.0, 0.13, 2, 0.10), 0.0)
        return d)
    _mk("warn", func():
        var d := PackedFloat32Array()
        d.resize(int(0.75 * SR))
        _mix(d, _tone_bytes(880.0, 440.0, 0.30, 2, 0.12), 0.0)
        _mix(d, _tone_bytes(880.0, 440.0, 0.30, 2, 0.12), 0.35)
        return d)
    _mk("dmg", func():
        var d := _tone_bytes(1500.0, 400.0, 0.09, 1, 0.14)
        _mix(d, _noise_bytes(0.08, 0, 2500.0, 2500.0, 0.14), 0.0)
        return d)
    _mk("bossdie", func():
        var d := PackedFloat32Array()
        d.resize(int(1.8 * SR))
        for i in 8:
            _mix(d, _noise_bytes(0.4, 0, 2000.0, 80.0, 0.30), i * 0.12)
            _mix(d, _tone_bytes(60.0, 28.0, 0.5, 0, 0.40), i * 0.12)
        return d)
    _mk("up", func():
        var d := PackedFloat32Array()
        d.resize(int(0.7 * SR))
        var semis := [0, 4, 7, 12]
        for i in 4:
            _mix(d, _tone_bytes(523.25 * pow(2.0, semis[i] / 12.0), 0.0, 0.22, 3, 0.13), i * 0.09)
        return d)
    _mk("pick", func():
        var d := PackedFloat32Array()
        d.resize(int(0.7 * SR))
        var semis := [0, 5, 9, 14]
        for i in 4:
            _mix(d, _tone_bytes(700.0 * pow(2.0, semis[i] / 12.0), 0.0, 0.3, 0, 0.12), i * 0.06)
        return d)
    _mk("over", func():
        var d := PackedFloat32Array()
        d.resize(int(2.2 * SR))
        var semis := [12, 7, 3, 0, -5]
        for i in 5:
            _mix(d, _tone_bytes(440.0 * pow(2.0, semis[i] / 12.0), 0.0, 0.5, 2, 0.10), i * 0.22)
        return d)
    _mk("swallow", func(): return _tone_bytes(70.0, 900.0, 0.5, 0, 0.20))
    _mk("chain", func(): return _tone_bytes(180.0, 120.0, 0.15, 0, 0.30))
    # --- instruments for music ---
    _inst["kick"] = _to_stream(_tone_bytes(150.0, 40.0, 0.15, 0, 0.55))
    _inst["hat"] = _to_stream(_noise_bytes(0.05, 1, 6000.0, 6000.0, 0.10))
    _inst["snare"] = _to_stream(_noise_bytes(0.10, 1, 3000.0, 3000.0, 0.16))
    for i in 17:
        _inst["bass" + str(i)] = _to_stream(_tone_bytes(55.0 * pow(2.0, i / 12.0), 0.0, 0.16, 2, 0.34))
    for i in 27:
        _inst["arp" + str(i)] = _to_stream(_tone_bytes(220.0 * pow(2.0, i / 12.0), 0.0, 0.07, 1, 0.09))
    _inst["lead"] = _to_stream(_tone_bytes(440.0, 0.0, 0.24, 3, 0.12))
    # --- player pools ---
    for i in 10:
        var p := AudioStreamPlayer.new()
        p.bus = "Master"
        add_child(p)
        _players.append(p)
    for i in 14:
        var p2 := AudioStreamPlayer.new()
        p2.volume_db = -6.0
        add_child(p2)
        _music_players.append(p2)

func sfx(name: String) -> void:
    if muted:
        return
    ensure()
    var stream: AudioStreamWAV = _sfx.get(name)
    if stream == null:
        return
    var p: AudioStreamPlayer = _players[_pi % _players.size()]
    _pi += 1
    p.stream = stream
    p.play()

func _play_inst(key: String, pitch_semi := 0.0) -> void:
    var idx: int
    if key.begins_with("bass"):
        idx = clampi(int(pitch_semi), 0, 16)
        key = "bass" + str(idx)
    elif key.begins_with("arp"):
        idx = clampi(int(pitch_semi), 0, 26)
        key = "arp" + str(idx)
    var stream: AudioStreamWAV = _inst.get(key)
    if stream == null:
        return
    var p: AudioStreamPlayer = _music_players[_mi % _music_players.size()]
    _mi += 1
    p.stream = stream
    p.play()

func set_heat(h: int) -> void:
    heat = clampi(h, 0, 3)

func set_boss(on: bool) -> void:
    boss_mode = on

func _process(delta: float) -> void:
    if muted:
        return
    bpm = 108.0
    var spb := 60.0 / bpm / 2.0
    _acc += delta
    if _next_t == 0.0:
        _next_t = _acc + 0.05
    while _next_t < _acc + 0.12:
        var st := _step % 16
        var shift := 6 if boss_mode else 0
        if st % 4 == 0:
            _play_inst("kick")
        if heat >= 1 and st % 2 == 1:
            _play_inst("hat")
        if heat >= 2 and st % 4 == 2:
            _play_inst("snare")
        var bass: Array = [0, 0, 3, 0, 5, 3, 0, -2, 0, 0, 3, 0, 7, 5, 3, 2] if boss_mode \
            else [0, 0, 5, 0, 7, 0, 5, 3, 0, 0, 5, 0, 10, 7, 5, 3]
        if st % 2 == 0:
            _play_inst("bass", bass[st] + shift)
        if heat >= 2:
            var arp: int = [0, 3, 7, 10, 12, 10, 7, 3][st % 8] + 12 + shift
            _play_inst("arp", arp)
        if heat >= 3 and st % 4 == 0:
            _play_inst("lead", shift)
        _next_t += spb
        _step += 1
