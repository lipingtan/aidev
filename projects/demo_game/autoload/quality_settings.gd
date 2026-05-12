extends Node
## 全局画质档与纹理路径解析。须在 DataManager、DlcManager 之前 Autoload。
##
## 定稿：仅 desktop_high + mobile_high（中端及以上），见 performance-budget.md

const TIER_DESKTOP_HIGH := "desktop_high"
const TIER_MOBILE_HIGH := "mobile_high"

const CLI_PREFIX := "--aidev-tier="

const TEXTURES_ROOT := "res://assets/textures"

var current_tier: String = TIER_DESKTOP_HIGH

const _TIER_TO_FOLDER: Dictionary = {
	TIER_DESKTOP_HIGH: "tier_desktop",
	TIER_MOBILE_HIGH: "tier_mobile",
}


func _ready() -> void:
	current_tier = _detect_tier()
	apply_renderer_and_quality_settings()


func _detect_tier() -> String:
	for arg in OS.get_cmdline_args():
		if arg.begins_with(CLI_PREFIX):
			var v := arg.substr(CLI_PREFIX.length())
			if _is_valid_tier_id(v):
				return v
	if OS.has_feature("mobile"):
		return TIER_MOBILE_HIGH
	return TIER_DESKTOP_HIGH


func _is_valid_tier_id(id: String) -> bool:
	return id == TIER_DESKTOP_HIGH or id == TIER_MOBILE_HIGH


func get_tier_folder() -> String:
	return _TIER_TO_FOLDER.get(current_tier, "tier_desktop")


func get_texture_base_path() -> String:
	return "%s/%s" % [TEXTURES_ROOT, get_tier_folder()]


## relative_path：相对于 tier 目录，例 "characters/hero/albedo.png"
func resolve_texture(relative_path: String) -> String:
	var rel := relative_path.lstrip("/")
	var tiers_try: Array[String] = []
	tiers_try.append(get_tier_folder())
	if get_tier_folder() != "tier_mobile":
		tiers_try.append("tier_mobile")
	if get_tier_folder() != "tier_desktop":
		tiers_try.append("tier_desktop")
	for folder in tiers_try:
		var full := "%s/%s/%s" % [TEXTURES_ROOT, folder, rel]
		if ResourceLoader.exists(full):
			return full
	push_warning("QualitySettings: 纹理未找到 %s（已试档位 %s）" % [rel, str(tiers_try)])
	return "%s/%s/%s" % [TEXTURES_ROOT, get_tier_folder(), rel]


func get_scalar(key: StringName) -> Variant:
	if current_tier == TIER_MOBILE_HIGH:
		return _scalar_mobile_high(key)
	return _scalar_desktop_high(key)


func _scalar_desktop_high(key: StringName) -> Variant:
	match key:
		&"shadow_distance":
			return 80.0
		&"shadow_max_blur":
			return 5
		&"ssao_enabled":
			return true
		&"ssil_enabled":
			return false
		&"glow_enabled":
			return true
		&"msaa_3d":
			return 2
		&"max_skinned_actors":
			return 20
		&"hscene_lod_skip":
			return 0
		&"particle_multiplier":
			return 1.0
		_:
			return null


func _scalar_mobile_high(key: StringName) -> Variant:
	match key:
		&"shadow_distance":
			return 40.0
		&"shadow_max_blur":
			return 2
		&"ssao_enabled":
			return false
		&"ssil_enabled":
			return false
		&"glow_enabled":
			return false
		&"msaa_3d":
			return 0
		&"max_skinned_actors":
			return 8
		&"hscene_lod_skip":
			return 1
		&"particle_multiplier":
			return 0.5
		_:
			return null


func apply_renderer_and_quality_settings() -> void:
	var vp := get_viewport()
	if vp == null:
		return
	var msaa_i: int = int(get_scalar(&"msaa_3d"))
	match msaa_i:
		0:
			vp.msaa_3d = Viewport.MSAA_DISABLED
		1:
			vp.msaa_3d = Viewport.MSAA_2X
		_:
			vp.msaa_3d = Viewport.MSAA_4X
