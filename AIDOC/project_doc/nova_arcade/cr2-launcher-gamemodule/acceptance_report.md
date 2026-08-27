# CR-2 Launcher + GameModule 验收报告

日期：2026-08-25 ｜ 引擎：Godot 4.5-stable（mobile renderer，720×1560 竖屏）｜ 状态：**通过**

## 1. GUI 双轨视觉检查（neon / elegant × home / run / result）

截图（720×1560，GUI exe + 隔离 APPDATA；驱动 `tools/_shot3.tscn`，脚本 `dev/tmp_vis/vis_run3.ps1`）：

| 画面 | neon | elegant |
|------|------|---------|
| home | dev/tmp_vis/shots/nova_neon_home.png | dev/tmp_vis/shots/nova_elegant_home.png |
| run | dev/tmp_vis/shots/nova_neon_run.png | dev/tmp_vis/shots/nova_elegant_run.png |
| result | dev/tmp_vis/shots/nova_neon_result.png | dev/tmp_vis/shots/nova_elegant_result.png |

read_image 逐张断言结论：

- **home**：首页标题/🎨主题按钮/空列表占位（「暂无内容（CR-3 接入列表）」）/PlayButton「▶ TETRA NOVA」/四 Tab 完整；App 子树溢出检查 none。
- **run**：Shell App 完全隐藏，仅游戏 HUD（HOLD/0/NEXT/WAVE 1 LINES 0）+ 棋盘面板 + BgLayer 装饰层；无 Shell 叠加。
- **result**：结算卡可见——标题 TETRA NOVA / 得分 0 / 用时 0:02 / 历史最佳 0 / [再来一局][还需 598 分钟（灰显）][返回]；「换一个」区隐藏（Q2=A，Registry.similar 空）；Shell 经 Dim 压暗在卡后；无 HUD 泄漏；Overflow 检查 none。
- 双主题配色均走 ThemeTokens.color()（neon 深底青色描边 / elegant 白卡灰绿按钮），无硬编码色。

**视觉检查发现并已修复的缺陷（3 项）**：

1. RUNNING 期间 Shell App 未隐藏 → home UI 与游戏 HUD 叠加。修复：`LauncherUtil.set_game_ui(on)`（GameHost 显隐与 App 相反）+ launcher.gd 5 处替换。
2. 结算卡父层 OverlayLayer 恒 visible=false，卡片永不渲染。修复：`ResultOverlay.show_card` 打开父层、_on_again/_on_back 关闭。
3. tetra_nova `ui` 为 CanvasLayer，不随祖先 Control.visible 隐藏 → HUD 泄漏到结算画面。修复：`module_adapter._set_ui_visible()`（quit 隐藏 / boot、reset_run 显示）。

## 2. 回归防护 RG-1~8

| # | WHEN | THEN | 验证手段 | 结果 |
|---|------|------|---------|------|
| RG-1~6 | （沿用 CR-1 §8：启动/Tab/push-pop/主题强写/CoreManager Mock） | 行为不破坏 | test_smoke / test_main / test_nav / test_theme / test_viewport / test_core_manager(-s) | ✅ 全绿 |
| RG-7 | launch 后强杀进程重启 | trial_used/playtime/best/finish_count 不丢；save.cfg 完整 | test_rg5 双相（write/verify）+ test_launch DB 强写断言 | ✅ 全绿 |
| RG-8 | 未 launch 时打开应用 | 首页/四 Tab/主题切换与 CR-1 一致（临时按钮除外） | test_main / test_nav + GUI home 截图复核 | ✅ 全绿 |

## 3. 自测矩阵

| 测试 | 范围 | 结果 |
|------|------|------|
| tools/test_launch.tscn（T8） | launch→boot→quit 全链路 25 断言：ctx 完整/遮满后切视口/App 显隐/sessions+trial/强写三字段/play_again 复用不耗 trial/耗尽弹 Mock 中止/坏场景回退 | ✅ ALL PASS，exit 0 |
| tools/test_smoke / test_main / test_nav / test_registry / test_theme / test_viewport / test_sound | CR-1 Shell 回归 | ✅ 全绿 |
| tools/test_db（×3）+ test_rg5 write/verify | DB 拆分后接口不变 + 双相持久化 | ✅ 全绿 |
| test_core_manager -s / test_tier -s | Mock 桩位 / 存储分级 | ✅ exit 0 |
| games/tetra_nova/src/scripts：test_runner / probe3 / probe4（-s） | 源工程基线（adapter 改动后复跑） | ✅ 全绿 exit 0 |

## 4. 设计解读注记（执行期裁定，留档）

1. BgLayer 青/紫色块为 CR-1 装饰星云（NebulaA/B），非渲染缺陷。
2. 评价按钮灰显规则：finish_count≥2 且本局 playtime≥600s 才可点亮；时长不足显示「还需 X 分钟」、局数不足显示「还需 N 局」（本次截图为 fresh APPDATA → best=0、「还需 598 分钟」，符合预期）。
3. 「换一个」区：Registry.similar(tetra_nova) 空数组 → 整区隐藏（Q2=A）；非空时迷你卡点击仅 Toast「详情页即将上线」（CR-3）。

## 5. 遗留项（显式登记）

| 项 | 说明 | 归属 |
|----|------|------|
| D3 挂账 | QualitySettings resolve_texture/get_scalar 不恢复，首个有纹理的 CR 恢复 | 后续 CR |
| M4 PurchaseDialog | trial 耗尽弹 Mock 购买弹窗（无支付链路） | M4 |
| M5 DLC 检查 / M2 成就判定 | Mock 桩位；本 CR achievements 恒空数组 | M2/M5 |
| 评价提交链路 | 「评价」点亮后仅 Toast「评价功能即将上线」 | CR-3/M1 |
| 首页 PlayButton | 临时启动按钮，CR-3 接入列表后移除 | CR-3 |
| Analytics.track | 不上报；本地 EventBus.game_launched/game_finished 为数据源 | M3 |

## 6. 结论

T1~T9 全部完成：代码级 POST-CHECK、GameModule 协议合规、headless 全链路自测、迁入基线、RG-1~8、双主题视觉检查均通过。CR-2 验收**通过**，具备进入 CR-3（requirements_plan）条件。
