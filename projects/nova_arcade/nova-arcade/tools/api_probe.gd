extends Node
## 通用 Godot 4.7 API 探针（API 名已实机核实）：
##   --headless --quit-after 60 res://tools/api_probe.tscn -- <ClassName> [member|doc] [filter|docs_dir]
## Modes:
##   no "doc" arg      → runtime reflection (ClassDB): enum constants / method / property existence
##   "doc" as member   → parse local doc HTML for the class, print extracted signature table:
##                        methods (with defaults) + properties (type+default) + signals + enum constants
##   "doc" + 3rd arg   → filter output to members containing that string
##   optional 4th arg  → docs dir (default C:\data\developer\devtool\godot\godot-docs-html-stable\classes)
const DEFAULT_DOCS_DIR := 'C:\\data\\developer\\devtool\\godot\\godot-docs-html-stable\\classes'

func _ready() -> void:
	var args: PackedStringArray = OS.get_cmdline_user_args()
	if args.is_empty():
		print('usage: api_probe.tscn -- <ClassName> [member|doc] [filter|docs_dir]')
		get_tree().quit(1)
		return
	var cname: String = args[0]
	var member: String = args[1] if args.size() > 1 else ''
	if not ClassDB.class_exists(cname):
		print('[probe] class NOT EXISTS: %s' % cname)
		get_tree().quit(1)
		return
	if member == 'doc':
		var filter := args[2] if args.size() > 2 else ''
		var docs_dir: String = args[3] if args.size() > 3 else DEFAULT_DOCS_DIR
		do_doc(cname, filter, docs_dir)
	elif member.is_empty():
		var consts: Array = ClassDB.class_get_integer_constant_list(cname)
		for c in consts:
			var cd: Dictionary = c
			print('[const] %s.%s = %d' % [cname, String(cd.get('name', '')), int(cd.get('value', -1))])
		print('[probe] methods=%d props=%d (constants listed above)' % [
			ClassDB.class_get_method_list(cname).size(),
			ClassDB.class_get_property_list(cname).size()])
	else:
		var found := false
		if ClassDB.class_has_integer_constant(cname, member):
			print('[const] %s.%s = %d' % [cname, member, int(ClassDB.class_get_integer_constant(cname, member))])
			found = true
		if ClassDB.class_has_method(cname, member):
			print('[method] %s.%s exists (args=%d)' % [cname, member, ClassDB.class_get_method_argument_count(cname, member)])
			found = true
		for p in ClassDB.class_get_property_list(cname):
			var pd: Dictionary = p
			if String(pd.get('name', '')) == member:
				print('[property] %s.%s exists (type=%d hint=%d)' % [cname, member, int(pd.get('type', -1)), int(pd.get('hint', -1))])
				found = true
				break
		if not found:
			print('[probe] MEMBER NOT FOUND: %s.%s' % [cname, member])
	get_tree().quit(0)

func _strip_tags(s: String) -> String:
	var out := s.replace('\n', ' ').replace('\t', ' ')
	while out.contains('<'):
		var i := out.find('<')
		var j := out.find('>', i + 1)
		if j < 0:
			break
		out = out.substr(0, i) + out.substr(j + 1)
	return out.strip_edges()

func _unesc(s: String) -> String:
	return s.replace('&lt;', '<').replace('&gt;', '>').replace('&amp;', '&').replace('&quot;', '"')

func _regex(pattern: String, multiline := false) -> RegEx:
	var re := RegEx.new()
	if multiline:
		pattern = '(?s)' + pattern
	re.compile(pattern)
	return re

func do_doc(cname: String, filter: String, docs_dir: String) -> void:
	var fname := 'class_' + cname.to_lower().replace('_', '') + '.html'
	var path := docs_dir.path_join(fname)
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		print('[doc] FILE NOT FOUND: %s (err=%d)' % [path, FileAccess.get_open_error()])
		get_tree().quit(1)
		return
	var html := f.get_as_text()
	f.close()
	# Methods: signature lines in the Methods table (with default values)
	var methods: Array[String] = []
	var re_m := _regex('[a-z_][a-z0-9_]*</span></a>\\([^)]*\\)')
	for m in re_m.search_all(html):
		var sig := _unesc(_strip_tags(m.get_string()))
		if not methods.has(sig):
			methods.append(sig)
	# Properties: whole paragraphs inside the property-descriptions section only
	var props: Array[String] = []
	var pi := html.find('id="property-descriptions"')
	if pi >= 0:
		var re_p := _regex('<p class="classref-property"[^>]*>(.*?)</p>', true)
		for m in re_p.search_all(html.substr(pi)):
			var sig := _unesc(_strip_tags(m.get_string(1)))
			sig = sig.replace('🔗', '').strip_edges()
			if not props.has(sig):
				props.append(sig)
	# Signals: whole paragraphs (name + signature)
	var signals: Array[String] = []
	var re_s := _regex('<p class="classref-signal"[^>]*>(.*?)</p>', true)
	for m in re_s.search_all(html):
		var sig := _unesc(_strip_tags(m.get_string(1)))
		sig = sig.replace('🔗', '').strip_edges()
		if not signals.has(sig):
			signals.append(sig)
	# Enum constants (with values)
	var consts: Array[String] = []
	var re_c := _regex('<p class="classref-enumeration-constant"[^>]*>(.*?)</p>', true)
	for m in re_c.search_all(html):
		var c := _unesc(_strip_tags(m.get_string(1)))
		if c.contains('=') and not consts.has(c):
			consts.append(c)
	print('[doc] %s: methods=%d props=%d signals=%d enum_consts=%d (source=%s)' % [cname, methods.size(), props.size(), signals.size(), consts.size(), path])
	if filter.is_empty():
		for s in methods:
			print('[doc:m] ' + s)
		for s in props:
			print('[doc:p] ' + s)
		for s in signals:
			print('[doc:s] ' + s)
		for s in consts:
			print('[doc:c] ' + s)
	else:
		var n := 0
		for arr in [methods, props, signals, consts]:
			for s in arr:
				if s.to_lower().contains(filter.to_lower()):
					print('[doc] ' + s)
					n += 1
		print('[doc] matched=%d (filter=%s)' % [n, filter])
