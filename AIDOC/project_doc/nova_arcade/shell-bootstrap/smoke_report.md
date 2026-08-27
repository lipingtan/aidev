# CR-1 shell-bootstrap 集成冒烟报告（T10）

> 依据 `dev-workflow-override.md` §5 冒烟清单 + RG-1~6。
> 环境：Godot 4.5-stable（headless 用 `godot4.5\Godot_v4.5-stable_win64_console.exe` 驱动 `res://tools/test_*.tscn`；§5.5 视觉用 GUI 版 `Godot_v4.5-stable_win64.exe`），`$env:APPDATA` 隔离。
> 结论：**CR-1 骨架冒烟通过**——§5.2 + RG-1~6 全过；§5.1 编辑器打开已于 2026-08-24 GUI `--editor --quit` 实测关闭（顺删废弃 `_shot.gd`）；§5.5 视觉验证 GUI 实测关闭（含 Fix-1 修复）；仅剩 §5.3/§5.4 因 Services 层未实现挂账至 CR-2/CR-3。

---

## 1. §5 冒烟清单

| # | 检查项 | 结果 | 说明 |
|---|---|---|---|
| 1 | Shell 主工程编辑器中可打开无报错 | ✅ **通过（2026-08-24 GUI 实测）** | `--editor --quit` 三次：首跑发现 `_shot.gd:22` 解析错误（废弃工具，已被 _shot2 取代）→ 删除后复跑零脚本/资源错误；剩余 stderr（TLS 证书/resthumb 预览缓存）为沙箱环境产物 |
| 2 | 底部四 Tab 可切换，无崩溃 | ✅ 通过 | test_smoke：四 Tab 切换无崩溃（见 §2） |
| 3 | 详情页 CTA 状态机（free/trial/paid） | ❌ **不可验（挂账）** | 详情页未建、PayService/TrialGuard 未实现（Services 层缺失）。→ **CR-2/CR-3** |
| 4 | Launcher 完整生命周期（launch→boot→quit→结算卡） | ❌ **不可验（挂账）** | LauncherService 未实现（仅单例+页面骨架）。→ **CR-2** |
| 5 | 双主题切换无布局跳变 | ✅ **通过（2026-08-24 GUI 实测，含 Fix-1 修复）** | 见 §4.1：neon/elegant × 720×1560/720×1600 四张截图 + 关键节点 rect 跨主题一致 + 溢出检查全过；期间发现并修复 Fix-1（Theme 资源不传播 + Bg 底色硬编码） |

---

## 2. 功能验证（test_smoke，12/12 全过，exit 0）

驱动真实 `shell/main.tscn`（含 PageStack），覆盖 §5.2 + RG-1/2/3/4/6：

| 断言 | 对应 |
|---|---|
| RG-1 启动进入 Home（§5.2） | Nav.current()==home |
| RG-2 切 Tab(category/search/library/home) 无崩溃 | 四 Tab 切换 |
| RG-3 push 入栈 / pop 回 Home | Nav.push/pop（占位 page_a） |
| RG-4 首页 🎨 按钮存在 | find_child("ThemeButton") |
| RG-4 主题切换生效 | ThemeTokens.current 变化 |
| RG-4 profile force 写（可持久化） | DB.get_profile().theme |
| RG-5 upsert_record 强写可落盘 | DB.upsert_record |
| RG-6 CoreManager Mock 返回未安装 | CoreManager.is_installed()=false |

```
[smoke] PASS RG-1 启动进入 Home（§5.2）
[smoke] PASS RG-2 切 Tab(category/search/library/home) ×4
[smoke] PASS RG-3 push 入栈 / pop 回 Home
[smoke] PASS RG-4 首页 🎨 按钮存在 / 主题切换生效 / profile force 写
[smoke] PASS RG-5 upsert_record 强写可落盘
[smoke] PASS RG-6 CoreManager Mock 返回未安装
[smoke] 合计 PASS=12 FAIL=0   [exit=0]
```

> 注：`Nav: 未找到 PageStack` 为占位/测试场景预期警告（test_smoke 用真实 main.tscn 时 PageStack 存在，警告来自 autoload _ready 时序）；`ObjectDB instances leaked` 为 headless 退出已知现象，不影响退出码。

---

## 3. RG-5 强写持久化（test_rg5，双相 6/6 全过，exit 0）

双相验证「重启不丢」：`write` 相写盘+flush 后退出 → `verify` 相重启读回断言。

| 断言 | 结果 |
|---|---|
| trial_used=5 重启不丢（强写） | ✅ |
| total_playtime=1000 重启不丢（强写） | ✅ |
| best=999 重启不丢（强写） | ✅ |
| 订单 ord_test/status=paid 重启不丢（强写） | ✅ |
| 防抖写搜索历史重启不丢（正常退出已 flush） | ✅ |
| profile 重启不丢（force 写） | ✅ |

> 结论：强写（trial_used/orders/playtime 等）与正常退出防抖 flush 均在重启后完整恢复，**RG-5 通过**。

---

## 4. 挂账项（非文档缺陷，依赖 Services/真机）

| 项 | 原因 | 承接 |
|---|---|---|
| §5.1 编辑器打开无报错 | headless 无编辑器 | 人工补（或 CI 装 Godot editor） |
| §5.3 详情页 CTA 状态机 | 详情页 + PayService/TrialGuard 未实现 | **CR-2/CR-3** |
| §5.4 Launcher 生命周期 | LauncherService 未实现 | **CR-2** |

**§5.5 视觉验证已完成（2026-08-24，GUI 实测 + read_image）**：Godot GUI 版（`Godot_v4.5-stable_win64.exe`，Vulkan/RTX 5090）跑 `tools/_shot2.{gd,tscn}` 矩阵——① neon/elegant × 720×1560/720×1600 四张截图（`dev/tmp_vis/vis_*.png`）read_image 核验：青瓷主题米白底+墨字+豆绿星云+点阵，霓虹主题正常；② 关键节点 rect（App/TabBar/Home/Title/ThemeButton）跨主题同尺寸完全一致 → **无布局跳变**；③ 溢出检查（App 子树遍历，1px 容差）四种组合全过；④ RG-4 重启恢复：C1 写 elegant → C2 `--restore` 读回 `[restore] PASS current=elegant==profile=elegant`。期间发现并修复 **Fix-1**（Theme 资源不传播 + Bg 底色硬编码，见 `Fix-1-dual-theme-not-applied/bugfix.md`），修复后复测全绿。

---

## 5. 测试脚本（新增于 `projects/nova_arcade/nova-arcade/tools/`）

| 文件 | 用途 | 运行 |
|---|---|---|
| `test_smoke.tscn` / `.gd` | §5.2 + RG-1/2/3/4/6 综合冒烟 | `--headless --path <proj> --scene res://tools/test_smoke.tscn` |
| `test_rg5.tscn` / `.gd` | RG-5 强写持久化双相 | `--scene res://tools/test_rg5.tscn --rg5-phase=write` 然后 `--rg5-phase=verify` |
| （已有）`smoke_boot.gd` | T1 启动门槛：6 autoload 加载 | `-s res://tools/smoke_boot.gd` |
| （已有）`test_main.gd` | T8 集成：启动/四Tab/Toast/主题 | `--scene res://tools/test_main.tscn` |
| `_shot2.tscn` / `.gd`（GUI 专用） | §5.5 视觉矩阵：`--theme=` `--scale=WxH` `--restore` `--shot_out=`，打印 [vis]/[diag]/[rect]/[overflow]/[restore]；25s 兜底 quit | GUI 版 Godot：`Godot_v4.5-stable_win64.exe --path <proj> res://tools/_shot2.tscn --theme=elegant ...`（批量见 `dev/tmp_vis/vis_run.ps1`） |

**结论：CR-1 shell-bootstrap 冒烟收口完成（§5.1 编辑器 + §5.2 + RG-1~6 + §5.5 视觉全过，三角色终版验收通过）**；遗留挂账：§5.3 详情页 CTA、§5.4 Launcher 生命周期（Services 层 → CR-2/CR-3）。
