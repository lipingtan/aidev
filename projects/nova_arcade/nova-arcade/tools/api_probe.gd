extends Node
## 通用 Godot 4.7 API 探针（API 名已实机核实）：
##   --headless --quit-after 60 res://tools/api_probe.tscn -- <ClassName> [member]
## 无 member arg → lists all int-constant (enum) fields of the class + methods/props count.
## With member arg → reports whether it exists as enum constant / method / property, plus value if constant.
func _ready() -> void:
	var args: PackedStringArray = OS.get_cmdline_user_args()
	if args.is_empty():
		print("usage: api_probe.tscn -- <ClassName> [member]")
		get_tree().quit(1)
		return
	var cname: String = args[0]
	var member: String = args[1] if args.size() > 1 else ""
	if not ClassDB.class_exists(cname):
		print("[probe] class NOT EXISTS: %s" % cname)
		get_tree().quit(1)
		return
	if member.is_empty():
		var consts: Array = ClassDB.class_get_integer_constant_list(cname)
		for c in consts:
			var cd: Dictionary = c
			print("[const] %s.%s = %d" % [cname, String(cd.get("name", "")), int(cd.get("value", -1))])
		print("[probe] methods=%d props=%d (constants listed above)" % [
			ClassDB.class_get_method_list(cname).size(),
			ClassDB.class_get_property_list(cname).size()])
	else:
		var found := false
		if ClassDB.class_has_integer_constant(cname, member):
			print("[const] %s.%s = %d" % [cname, member, int(ClassDB.class_get_integer_constant(cname, member))])
			found = true
		if ClassDB.class_has_method(cname, member):
			print("[method] %s.%s exists (args=%d)" % [cname, member, ClassDB.class_get_method_argument_count(cname, member)])
			found = true
		for p in ClassDB.class_get_property_list(cname):
			var pd: Dictionary = p
			if String(pd.get("name", "")) == member:
				print("[property] %s.%s exists (type=%d hint=%d)" % [cname, member, int(pd.get("type", -1)), int(pd.get("hint", -1))])
				found = true
				break
		if not found:
			print("[probe] MEMBER NOT FOUND: %s.%s" % [cname, member])
	get_tree().quit(0)
