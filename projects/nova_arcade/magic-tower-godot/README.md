# Magic Tower — Godot 4.5 移动端 Demo

一个可直接运行的 **Godot 4.5** demo，演示你要的两件事：
1. **滚动镜头 + 大地图**（20×20 世界，同屏约 10×10 格，相机平滑跟随人物、限位到地图边界）
2. **清晰锐利的人物 / 怪物 / 道具**（像素画 + Nearest 最近邻过滤 + 整数倍缩放）

> 美术与逻辑来自之前的 JS 原型；这里把精灵导出成 PNG 直接放进 Godot。

## 打开 & 运行
1. 启动 Godot **4.5** → 项目管理器点 **导入(Import)** → 选本目录的 `project.godot`。
2. 点 **编辑(Edit)** 进入编辑器，按 **F6**（或右上角 ▶）运行。
3. 操作：
   - 移动：**方向键 / WASD**（逐格走、撞墙阻挡、相机跟随）；手机 = 左下 **十字盘(D-pad)** 按住连走（Phoenix 主题图、8 向高亮）。项目已开 `Emulate Touch From Mouse`，编辑器里用鼠标点也能测
   - 战斗：走到怪物旁进入**回合制战斗**，按 **空格 / 回车 / J**（或 A 键）攻击
   - 拾取：走到道具上自动生效（回血/加攻/加防/金币），顶部实时显示 HP/ATK/DEF/Gold
   - 走到 **楼梯(S)** 触发提示（本 demo 为单层）；右侧 A/B/C/D 与 Start/投币/退币/退出按钮可点（按下换 `_press` 图）

## 无头自检（逻辑验证）
设环境变量后运行，会打印伤害公式 / 道具数值 / 网格 / 战斗模拟的断言结果：
```powershell
$env:MT_SELFTEST=1
& "C:\data\developer\devtool\godot\godot4.5\Godot_v4.5-stable_win64_console.exe" --headless --path .\magic-tower-godot --quit
```
输出 `[selftest] ALL PASS ✅` 即核心逻辑正确（断言在编辑器/调试构建下生效）。

## 关键设置（在 project.godot / Project Settings）
| 项 | 值 | 作用 |
|---|---|---|
| `Viewport Width/Height` | **320 × 320** | 基础分辨率 = 同屏 10×10 格（每格 32px） |
| `Stretch → Mode` | **canvas_items** | 按内容整数缩放，像素对齐 |
| `Stretch → Aspect` | **keep** | 保持方形、留黑边不变形 |
| `Default Texture Filter` | **Nearest** | 像素画锐利（关键！） |
| `Rendering Method` | gl_compatibility | 桌面/手机/Web 通用；真机可改 `mobile` |
| `Handheld → Orientation` | portrait | 竖屏 |

**整数缩放（可选，保证绝对清晰）**：在 `scripts/main.gd` 的 `_ready()` 里加——按屏幕物理尺寸把 `get_tree().root.content_scale_factor` 设成能塞进的最大整数倍。基础分辨率越小、放大倍数越整，边缘越锐利。

## 导出到手机
- **Web（最快，免安装）**：`Project → Export` 添加 **Web** 预设 → 把该预设的 **Thread Support 关掉**（省去 COOP/COEP 头）→ 导出得到 `.html/.wasm/.js` → 丢到任意静态服务器（如 `python -m http.server`）→ 手机浏览器打开。
- **Android**：装 Android SDK + 配 debug keystore → USB 连机开 USB 调试 → 一键部署；或导出 APK `adb install`。
- **iOS**：需 macOS + Xcode + Apple 开发者账号（Windows 走不了）。

## 目录结构
```
project.godot        # 工程配置（分辨率/Nearest/竖屏等）
icon.png            # 图标
assets/             # 游戏精灵 PNG（hero、6 怪物含狼人、5 道具、map.png）
assets/dpad/        # 十字方向键 D-pad（普通档，9 个方向态）
assets/dpad_large/  # 十字方向键 D-pad（大图档，9 个方向态）
assets/buttons/     # A/B/C/D + 投币/退币/退出/开始 按钮（各含按下态，备用）
scenes/main.tscn    # 主场景（空壳，逻辑在脚本里构建）
scripts/main.gd     # 地图/怪物/道具/玩家+Camera2D/HUD/十字盘(按方向换图)
```

## 已实现 / 下一步
**已实现**：格子地图 + 墙体碰撞、回合制战斗（魔塔伤害公式 `(atk-def)²/100`）、道具属性（血/攻/防/金币）、楼梯、十字盘 + A/B/C/D 与功能键触屏、**无头自检**。
**下一步**：
- **多楼层 + 楼梯切换**（当前单层）
- **门需要对应钥匙**（红/蓝/金门 + `key_*` 道具，图已可导出）
- 怪物**巡逻 AI**（当前静态站立）
- 用 **TileMapLayer** 替换烘焙图做可编辑地图；精灵升 **48px 原生**加细节
- 十字盘切大图档：`scripts/main.gd` 里 `USE_LARGE_DPAD = true`
- 导出 Web / Android（`Project → Export`）真机试玩
