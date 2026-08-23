extends Node
## No-op audio stub for headless logic tests.

var boss_mode := false
var heat := 0

func ensure() -> void: pass
func sfx(_name: String) -> void: pass
func set_heat(h: int) -> void: heat = h
