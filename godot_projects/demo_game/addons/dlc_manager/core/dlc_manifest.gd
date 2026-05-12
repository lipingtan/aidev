class_name DlcManifest extends RefCounted
## DLC 清单数据
##
## 解析 manifest.json 后的结构化数据。

var id: String = ""
var version: String = ""
var display_name: String = ""
var description: String = ""
var min_game_version: String = ""
var max_game_version: String = ""
var dependencies: Array[String] = []
var conflicts: Array[String] = []
var load_order: int = 100
var content: Dictionary = {}
var base_path: String = ""


## 从 JSON 字典解析
static func from_dict(data: Dictionary, path: String) -> DlcManifest:
	var manifest := DlcManifest.new()
	manifest.id = data.get("id", "")
	manifest.version = data.get("version", "0.0.0")
	manifest.display_name = data.get("display_name", manifest.id)
	manifest.description = data.get("description", "")
	manifest.min_game_version = data.get("min_game_version", "0.0.0")
	manifest.max_game_version = data.get("max_game_version", "99.99.99")
	manifest.load_order = data.get("load_order", 100)
	manifest.content = data.get("content", {})
	manifest.base_path = path
	
	var deps = data.get("dependencies", [])
	for dep in deps:
		manifest.dependencies.append(str(dep))
	
	var confs = data.get("conflicts", [])
	for conf in confs:
		manifest.conflicts.append(str(conf))
	
	return manifest


## 验证清单完整性
func is_valid() -> bool:
	return id != "" and version != ""
