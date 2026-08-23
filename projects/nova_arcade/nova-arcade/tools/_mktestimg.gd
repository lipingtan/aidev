extends Node

func _ready() -> void:
	var img := Image.create(4, 4, false, Image.FORMAT_RGBA8)
	img.fill(Color(1.0, 0.0, 0.0, 1.0))
	var res_path := "res://tools/_testimg_test.png"
	var abs := ProjectSettings.globalize_path(res_path)
	var ok := img.save_png(abs)
	print("[mk] abs=", abs, " save_png_err=", ok)
	get_tree().quit(0)
