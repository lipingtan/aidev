# 可视化验证（能看到画面）

> **关键事实：本 agent 有图像输入**，可用 `read_image <png>` 直接"看"工作区里任意图片。所以验证不止靠无头自检——可以把画面渲染成 PNG 再读回来看，形成"改 → 渲染 → 看 → 再改"的迭代闭环。

## 1. 资产级验证（本会话已反复使用）
把精灵 / 按钮 / 地图烘焙成 PNG，再 `read_image` 逐张检查清晰度、方向态、按下态等：
- 实例：用 `read_image` 看过 `hero.png`、9 张 `dpad_*.png`、`button_a.png`（普通）与 `buttons_large/button_a.png`（2×），确认锐利 / 放大不糊。
- 生成 PNG 的两条管道：
  - JS canvas 导出：`magic-tower/tools/exportall.html` → base64 JSON → PowerShell 写盘
  - PowerShell `System.Drawing` 缩放：`HighQualityBicubic` 2×（平滑渐变图放大不糊，见 `assets/buttons_large/`）

## 2. 看"正在运行的游戏"画面
1. 起 GUI：`& 'C:\data\developer\devtool\godot\godot4.7\Godot_v4.7-stable_win64.exe' --path <工程目录>`（或在编辑器里按 **F6**）。
2. 若当前有可见桌面 / 显示器会话，用 PowerShell 截屏存 PNG：
   ```powershell
   Add-Type -AssemblyName System.Drawing
   $b = New-Object System.Drawing.Bitmap 1280,720
   $g = [System.Drawing.Graphics]::FromImage($b)
   $g.CopyFromScreen(0,0,0,0,$b.Size)      # 或指定窗口区域 (left,top,x,y)
   $b.Save('AIDOC\project_doc\nova_arcade\dev\shot.png',[System.Drawing.Imaging.ImageFormat]::PNG)  # 相对本工作区根目录
   ```
3. `read_image AIDOC/project_doc/nova_arcade/dev/shot.png` → 看到实际画面，据此迭代布局 / 清晰度。

> ⚠️ 需要真实显示器 / 桌面会话；纯无头服务器可能截到黑屏或失败——此时退回**逻辑级验证**（下条），它永远可用。

## 3. 逻辑级验证（无头，永远可用）
- `& $G --headless --path <工程> --quit`：确认无脚本 / 加载错误（exit 0）。
- 自带断言自检：设 `$env:MT_SELFTEST=1` 再跑，打印伤害公式 / 道具数值 / 网格 / 战斗模拟结果（如 `[selftest] ALL PASS ✅`）。
- 详见各工程 README（如 `magic-tower-godot/README.md` 的"无头自检"节）。

## 本项目（magic-tower-godot）速查
- 资产：`assets/{dpad,dpad_large,buttons,buttons_large}/`、`map.png`、hero / 怪物 / 道具 PNG。
- 自检入口：`scripts/main.gd` 的 `_selftest()`（`MT_SELFTEST=1` 触发）。
- 触屏：左下十字盘（8 向高亮）+ 右侧 A/B/C/D 与 Start/投币/退币/退出（按下换 `_press` 图，多指独立）。
