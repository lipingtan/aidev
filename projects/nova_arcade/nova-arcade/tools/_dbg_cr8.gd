extends Node
func _ready() -> void:
	var scene := load("res://games/slot_machine/module.tscn") as PackedScene
	var module := scene.instantiate()
	add_child(module)
	await get_tree().process_frame
	print("[dbg] A=", module.find_children("Cell*", "*", true, false).size())
	print("[dbg] B=", module.find_children("Cell*", "", true, false).size())
	var ring := module.get_node("Src/RingLayer")
	print("[dbg] C=", ring.get_children().size())
	get_tree().quit(0)
