# Godot 4.7 API 已核实事实（执行期验证，勿凭记忆改回）

> 用途：写代码前先查本文件；未覆盖的 API 先 grep 本地官方文档，仍不确定则用 `tools/api_probe.tscn` 实机探针确认后再写。
> 所有条目均在 Godot_v4.7.2-stable 实机验证过（headless 探针输出）。

## 本地官方文档（GFW 下以此为准，勿在线查）

- 完整副本：`C:\data\developer\devtool\godot\godot-docs-html-stable\`（Godot 4.7 英文）
- 类参考：`...\classes\`（1079 个类；文件名 = `class_` + 全小写去下划线类名，如 `class_texturerect.html`、`class_classdb.html`）
- 官方最佳实践：`...\tutorials\best_practices\`（13 篇：scene/project organization、scenes vs scripts、data/logic preferences 等）
- 查 API：`grep -oE "方法名\([^)]{0,60}" classes/class_<类>.html`；设计节点结构/数据流前先读 best_practices

## 已核实 API 事实（2026-08-27 实机验证）

| API / 语义 | 正确用法 | 踩坑记录 |
|---|---|---|
| `ClassDB` 反射（4.7） | `class_exists` / `class_get_integer_constant_list` / `class_has_integer_constant` / `class_get_integer_constant` / `class_has_method` / `class_get_method_argument_count` / `class_get_property_list`；**不存在** `get_class_integer_constants`、`class_has_property`（`class_get_property` 第一参是 Object 不是类名） | api_probe.gd 首版两个 API 名写错致解析失败 |
| headless 场景挂起 | 脚本解析失败 → 无 `_ready` → `quit()` 不执行 → **进程永久挂起**；跑测试场景必加 `--quit-after <frames>` 兜底 | api_probe/_probe_classdb 首跑 exit 124 |
| 项目 warning-as-error | GDScript 推断 Variant 类型等警告按错误处理，探针/测试脚本须显式标注类型（`var x: Array = ...`） | _probe_classdb 首版解析失败 |
| `StyleBoxFlat` 圆角 | `set_corner_radius_all(r)`；`set_corner_radius(side, r)` 需 2 参（side 枚举） | CR-5：单参调用致脚本编译失败、节点不挂脚本 |
| `TextureRect.stretch_mode` | `0=STRETCH_SCALE` `1=TILE` `2=KEEP`；铺满容器用 0 + `expand_mode=1(IGNORE_SIZE)` | CR-5：tscn 里写 stretch_mode=2 → 渐变只占 64×64 左上角"窄条" |
| `Image.get_pixel` | 两个 int 参数 `get_pixel(x, y) -> Color`，不接受 Vector2i | CR-5 视觉脚本 |
| `.instantiate()` | 返回 Variant，后续调方法须 `as <class>` 显式转换，否则解析失败致挂起 | test_home 首跑 exit 124 |
| GDScript lambda 捕获 | 按值捕获局部变量；信号断言等需跨帧读的值用成员变量 | test_home RG-24 |
| `Panel` / PanelContainer | 无 `texture` 属性；承载 GradientTexture2D 必须用 TextureRect | CR-5 BannerRect |
| autoload `_ready` 时序 | 页面组件 `_ready` 早于部分 autoload（如 Registry.reload）完成 → 数据读取用 `call_deferred` | CR-5 _load_banners 同步读 editorial 必空 |
| `SceneTreeTimer` | 一次性；循环需重建；reload 时置空旧引用 | banner carousel |
| `DB.get_record` | 返回 live reference，比较前 `.duplicate()` | 测试断言 |
| 卡片/组件 setup 时序 | `add_child(node)` 之后再调 `node.setup()`（@onready 在入树时同步解析） | home.gd + category.gd:106 |

## 探针用法（未覆盖 API 先探后写）

```
$G --headless --path . --quit-after 60 res://tools/api_probe.tscn -- TextureRect stretch_mode   # 成员查询
$G --headless --path . --quit-after 60 res://tools/api_probe.tscn -- StyleBoxFlat              # 列出全部整型常量 + 方法/属性计数
```

输出：class 存在性、成员是枚举常量（含值）/方法（含参数数）/属性，还是不存在。
