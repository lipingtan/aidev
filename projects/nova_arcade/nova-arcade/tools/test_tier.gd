extends SceneTree
## 档位判定回归脚本（T2 验收资产，保留复用）
##
## 用法：Godot --headless --path <工程目录> -s res://tools/test_tier.gd [--aidev-tier=mobile_high]
## 打印 current_tier 与 resolve_texture 结果：
## 桌面默认应判 desktop_high；加 --aidev-tier=mobile_high 后应返回 tier_mobile 路径。

func _initialize() -> void:
	var qs := root.get_node_or_null(NodePath("QualitySettings"))
	if qs == null:
		push_error("[test_tier] 缺少 Autoload: QualitySettings")
		quit(1)
		return
	# -s 脚本模式下 Autoload _ready 晚于本脚本 _initialize 触发，须显式触发一次档位检测
	qs.detect_tier()
	print("[test_tier] tier=%s" % str(qs.current_tier))
	var path := str(qs.resolve_texture("probe.png"))
	print("[test_tier] resolve_texture(probe.png)=%s" % path)
	if qs.current_tier == "mobile_high" and not path.ends_with("/tier_mobile/probe.png"):
		push_error("[test_tier] mobile_high 应返回 tier_mobile 路径: %s" % path)
		quit(1)
		return
	print("[test_tier] OK")
	quit(0)
