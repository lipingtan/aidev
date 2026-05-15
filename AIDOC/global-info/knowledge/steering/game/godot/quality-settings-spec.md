# QualitySettings 实现定稿（Godot 4.x）

> 与 `performance-budget.md` 数值一致。Autoload 名建议 **`QualitySettings`**，路径 **`res://autoload/quality_settings.gd`**。

## 1. Autoload 顺序（强制）

在 **`DataManager`、`DlcManager`、一切会 `preload` 带贴图路径的资源之前** 注册 `QualitySettings`。

**SaveManager 与 DlcManager**：默认 **SaveManager 在前**（存档可记录 DLC / 内容版本）；若加载管线要求先挂载 DLC 再读档，须在项目 `architecture.md` 写明并全局调整顺序。

推荐完整顺序（与 `demo_game` 一致）：

1. `EventBus`（若工程有）
2. **`QualitySettings`**
3. `EcsWorld`
4. `ObjectPool`
5. `DataManager`
6. `SaveManager`
7. `DlcManager`

若工程无 `ObjectPool`，`DataManager` 仍须排在 **`QualitySettings` 之后**。

## 2. `project.godot` 与运行时画质

- **渲染方法**：以 **导出预设** 为准——Android 预设默认 `mobile`（**Forward Mobile**，非 Compatibility），Windows 用 `forward_plus`。`gl_compatibility` 仅作为移动 Forward Mobile 实测不达标时的兜底导出预设。运行时改 `rendering_method` 往往不可靠，**不作为本工作台默认要求**。
- **`QualitySettings.apply_renderer_and_quality_settings()`**（参考实现）：根据 `get_scalar(&"msaa_3d")` 设置**主视口 MSAA**；其余键（`shadow_distance`、`ssao_enabled` 等）由**各系统在读取后自行应用**（灯光、Environment、后处理节点），避免单例强耦合场景树。

## 3. 公开 API（须实现）

```gdscript
# 当前档位 ID，与 performance-budget 表一致
var current_tier: String

func get_tier_folder() -> String:
    # desktop_high -> tier_desktop … 映射见 performance-budget 第二节

func get_texture_base_path() -> String:
    return "res://assets/textures/%s" % get_tier_folder()

## relative 例："characters/hero/albedo.png"，不含 tier 前缀
func resolve_texture(relative_path: String) -> String:
    pass  # 带 fallback 链，见 quality_settings.gd

func get_scalar(key: StringName) -> Variant:
    pass  # 键表见 performance-budget 第五节

func apply_renderer_and_quality_settings() -> void:
    pass  # 参考实现：仅 MSAA；其余标量由灯光/Environment/系统消费
```

## 4. 档位检测优先级（固定）

1. 命令行：扫描 `OS.get_cmdline_args()`，匹配 `--aidev-tier=desktop_high` 或 `--aidev-tier=mobile_high`（**在 PC 上模拟移动标量与 tier_mobile**）。
2. `OS.has_feature("mobile")` → `mobile_high`。
3. 否则 `desktop_high`。

（工作台不再维护 `desktop_low` / `mobile_low`。）

## 5. 纹理 fallback 链（固定）

解析 `resolve_texture` 时：

1. `{tier_current}/relative_path`
2. `tier_mobile/relative_path`
3. `tier_desktop/relative_path`

任一步 `ResourceLoader.exists` 为真即用；全否在 DEV 下 `push_warning`。

## 7. 验证清单（Phase 2）

- [ ] PC 编辑器 F5：`current_tier` 为 `desktop_high`（无参数时）。
- [ ] Android 导出：无参启动为 `mobile_high`；画面无整片洋红缺失贴图。
- [ ] 命令行 `--aidev-tier=mobile_high` 在桌面跑：`get_scalar(&"shadow_distance")` 等于 `performance-budget.md` 中 **mobile_high** 列。

参考脚本：`projects/demo_game/autoload/quality_settings.gd`。
