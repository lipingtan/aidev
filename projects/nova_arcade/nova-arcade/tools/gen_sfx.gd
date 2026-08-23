extends SceneTree
## 生成占位短音效 WAV（CR-1；M2 换正式音效后删除本工具）
## 用法：Godot --headless --path <proj> -s res://tools/gen_sfx.gd

func _initialize() -> void:
	_gen("click", 880.0, 880.0, 0.08)
	_gen("toggle", 660.0, 660.0, 0.12)
	_gen("success", 660.0, 990.0, 0.25)
	_gen("error", 220.0, 180.0, 0.2)
	_gen("coin", 1320.0, 1760.0, 0.18)
	quit(0)

## 生成单声道 16bit PCM WAV（正弦扫频 + 淡入淡出包络）
func _gen(name: String, f0: float, f1: float, secs: float) -> void:
	var sr := 22050
	var n := int(sr * secs)
	var bytes := PackedByteArray()
	bytes.append_array(_str("RIFF"))
	bytes.append_array(_u32(36 + n * 2))
	bytes.append_array(_str("WAVE"))
	bytes.append_array(_str("fmt "))
	bytes.append_array(_u32(16))
	bytes.append_array(_u16(1))
	bytes.append_array(_u16(1))
	bytes.append_array(_u32(sr))
	bytes.append_array(_u32(sr * 2))
	bytes.append_array(_u16(2))
	bytes.append_array(_u16(16))
	bytes.append_array(_str("data"))
	bytes.append_array(_u32(n * 2))
	for i in n:
		var t := float(i) / float(n)
		var f := f0 + (f1 - f0) * t
		var env := minf(1.0, t * 8.0) * (1.0 - t)
		var s := 0.5 * env * sin(TAU * f * float(i) / float(sr))
		var v := int(clampf(s, -1.0, 1.0) * 32767.0)
		bytes.append(v & 0xFF)
		bytes.append((v >> 8) & 0xFF)
	var path := "res://assets/sfx/%s.wav" % name
	var file := FileAccess.open(path, FileAccess.WRITE)
	file.store_buffer(bytes)
	file.close()
	print("[gen_sfx] ", path, " (", bytes.size(), " bytes)")

func _str(s: String) -> PackedByteArray:
	return PackedByteArray(s.to_ascii_buffer())

func _u16(v: int) -> PackedByteArray:
	return PackedByteArray([v & 0xFF, (v >> 8) & 0xFF])

func _u32(v: int) -> PackedByteArray:
	return PackedByteArray(
		[v & 0xFF, (v >> 8) & 0xFF, (v >> 16) & 0xFF, (v >> 24) & 0xFF]
	)
