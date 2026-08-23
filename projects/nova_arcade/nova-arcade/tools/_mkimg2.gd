extends Node

func _ready() -> void:
	var img := Image.create(8, 8, false, Image.FORMAT_RGBA8)
	img.fill(Color(0.2, 0.6, 0.9, 1.0))
	var jpg := ProjectSettings.globalize_path("res://tools/_testimg_fmt.jpg")
	var webp := ProjectSettings.globalize_path("res://tools/_testimg_fmt.webp")
	print("[mk] jpg_err=", img.save_jpg(jpg, 0.9), " path=", jpg)
	print("[mk] webp_err=", img.save_webp(webp, 0.9), " path=", webp)
	get_tree().quit(0)
