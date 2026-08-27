# Godot 工具链

## 可执行文件

| 文件 | 用途 |
|---|---|
| `C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64.exe` | GUI 编辑器，日常开发、场景编辑、导出 |
| `C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64_console.exe` | 无头命令行，CI/脚本/PoC 验证（`--headless`），输出直接打到终端 |

导出模板已安装：`Godot_v4.7-stable_export_templates.tpz`（同目录）。

> `win64_console.exe` 体积只有 197KB，是 stub 启动器，依赖主 exe 旁的运行时库——**两个 exe 必须在同一目录下**。

---

## 常用命令

```powershell
$G  = 'C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64_console.exe'
$GE = 'C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64.exe'

# 打开工程（GUI）
& $GE --path <工程目录>

# 无头运行脚本（PoC / 单元测试）
& $G --headless --path <工程目录> -s res://path/to/script.gd

# 无头运行场景
& $G --headless --path <工程目录> res://path/to/scene.tscn

# 导出 PCK（DLC 打包）
& $G --headless --path <游戏工程目录> --export-pack PoC <输出.pck绝对路径>

# 导出 Android APK
& $G --headless --path <工程目录> --export-release Android <输出.apk绝对路径>
```

---

## 工程目录

> 路径基准：本工作区（`C:\data\developer\studio\games\aidev`）。原先 `C:\data\run\test\*` 下的副本已废弃，一律以本工作区为准。

| 工程 | 路径（相对本工作区） |
|---|---|
| NOVA ARCADE PoC（运行时验证） | `projects/nova_arcade/nova-arcade-poc` |
| TETRA NOVA | `projects/nova_arcade/tetra-nova-godot` |
| NOVA ARCADE Shell | `projects/nova_arcade/nova-arcade` |

---

## PoC 运行示例

```powershell
$G = 'C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64_console.exe'

# HTML 运行时 PoC
& $G --headless --path projects\nova_arcade\nova-arcade-poc -s res://poc/html_runtime_driver.gd

# Arcade 运行时 PoC
& $G --headless --path projects\nova_arcade\nova-arcade-poc -s res://poc/arcade_runtime_driver.gd

# PCK 运行时 PoC（先导出 mini_demo.pck，再挂载运行）
& $G --headless --path projects\nova_arcade\nova-arcade-poc\game-mini-demo --export-pack PoC projects\nova_arcade\nova-arcade-poc\dlcs\mini_demo.pck
& $G --headless --path projects\nova_arcade\nova-arcade-poc -- projects\nova_arcade\nova-arcade-poc\dlcs\mini_demo.pck
```

---

## 官方文档

- Godot 4.7 官方文档：https://docs.godotengine.org/en/4.7/
- **API 已核实事实 + 未覆盖 API 的实机探针**：见 `kb/godot-4.7-api-facts.md`（写代码前先查；GFW 下无法在线查 docs，需本地 class reference 副本）

---

## 注意事项

- `--headless` 模式无窗口，`print()` 输出到 stdout，适合脚本化验证
- Android 导出前需在 GUI 编辑器的 Editor Settings 里配置 Android SDK 路径和 keystore，之后命令行导出才生效
- `user://` 在 Windows 下默认映射到 `%APPDATA%\Godot\app_userdata\<工程名>`；脚本验证时可临时重定向：
  ```powershell
  $env:APPDATA = 'C:\tmp\godot_test'
  ```
