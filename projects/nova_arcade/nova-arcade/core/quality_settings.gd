extends Node
## 多平台画质档与纹理路径解析（Autoload 名 QualitySettings，须在 DB/Registry 之前注册）
##
## 裁剪自 projects/demo/demo_game/autoload/quality_settings.gd（实查唯一参考实现）：
## - 保留：档位检测（--aidev-tier= CLI 覆盖 + 平台判定）、resolve_texture()、get_scalar()
## - 裁剪：apply_renderer_and_quality_settings()——3D MSAA 视口调整，盒子为 mobile 2D 渲染器（project.godot 已定稿），无 3D 场景可应用
## - 裁剪：resolve_texture 内的多档位回退业务逻辑——Shell UI 为矢量/ColorRect 绘制、几乎无纹理资产；
##   纹理分档在 CR-2 迁入游戏后按需启用（design.md §3.2 / Risk-3），届时恢复回退链并补存在性校验
##
## 依赖：
## - 无（仅读 OS 命令行参数与平台特性，须先于数据类单例注册）

const TIER_DESKTOP_HIGH := "desktop_high"
const TIER_MOBILE_HIGH := "mobile_high"

const CLI_PREFIX := "--aidev-tier="

const TEXTURES_ROOT := "res://assets/textures"

const _TIER_TO_FOLDER: Dictionary = {
	TIER_DESKTOP_HIGH: "tier_desktop",
	TIER_MOBILE_HIGH: "tier_mobile",
}

var current_tier: String = TIER_DESKTOP_HIGH


func _ready() -> void:
	detect_tier()


## 档位检测：--aidev-tier= CLI 覆盖优先；否则平台判定（mobile → mobile_high，其余 → desktop_high）
## 公开供回归脚本在 -s 模式下显式调用（该模式 Autoload _ready 晚于脚本 _initialize 触发）
func detect_tier() -> void:
	current_tier = _detect_tier()


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


## 当前档位对应的纹理目录名（tier_desktop / tier_mobile）
func get_tier_folder() -> String:
	return _TIER_TO_FOLDER.get(current_tier, "tier_desktop")


## 当前档位的纹理根路径
func get_texture_base_path() -> String:
	return "%s/%s" % [TEXTURES_ROOT, get_tier_folder()]


## 解析纹理完整路径（relative_path 相对当前档位目录，例 "characters/hero/albedo.png"）
## 注：Shell UI 为矢量绘制、无纹理资产，本接口为 CR-2 迁入游戏预留；
## demo 工程的多档位回退逻辑已裁剪，CR-2 后按需恢复。
func resolve_texture(relative_path: String) -> String:
	var rel := relative_path.lstrip("/")
	return "%s/%s/%s" % [TEXTURES_ROOT, get_tier_folder(), rel]


## 画质标量参数（数值沿用 demo_game 定稿；盒子 2D UI 暂不消费，CR-2 迁入游戏后生效）
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
