# TETRA NOVA — Godot 4.5 移动端版

网页版《TETRA NOVA》（变异 Roguelike 俄罗斯方块）的 Godot 4.5 完整移植。
零素材：所有图形程序绘制、所有音效/BGM 运行时 PCM 合成。

## 打开

用 Godot 4.5（stable）打开本目录的 `project.godot`，F5 运行。

## 操作

| 动作 | 键盘 | 触摸 |
|---|---|---|
| 移动 | ← → | 水平滑动（按格吸附） |
| 软降 | ↓ 按住 | 下滑按住 |
| 硬降 | 空格 | 快速下滑甩出 |
| 旋转 | ↑ / X（反旋 Z） | 点按（左 1/6 反旋） |
| 暂存 | C / Shift | 长按 0.45s |
| 选卡 | 1 / 2 / 3，R 重抽 | 点卡牌 |
| 暂停 / 静音 / 按钮排 | P / M / B | — |

## 玩法（与网页版一致）

- 经典 SRS（踢墙旋转）+ 7-bag + Hold + 幽灵块 + DAS + 锁定延迟
- 每波清行 → 3 选 1 变异（16 种，可叠加，带 SYNERGY 协同）
- 每 5 波 Boss：垃圾行攻击、核心块弱点、狂暴二连发、战利品稀有保底
- 变异：炸弹、闪电链、余烬、黑洞、坍缩连锁、奇点爆发、双子 AI、第二次呼吸等

## 导出 Android

1. 编辑器 → 编辑 → 管理导出模板（已含 `Godot_v4.5-stable_export_templates.tpz` 则跳过）
2. 项目 → 导出 → 选 `Android` preset（arm64 + armeabi-v7a）
3. 需要本机 Android SDK/Keystore 配置（Godot 编辑器一次配置，全局生效）
4. 导出为 `builds/TetraNova.apk`

竖屏锁定、触控 D-pad 默认显示（B 键或双指关闭后用手势）。

## 无头测试（无需打开编辑器）

```
# 核心逻辑 20 项断言（消行/变异/Boss/双子/二段呼吸/重开）
godot --headless --path . -s res://scripts/test_runner.gd

# 全脚本解析检查
godot --headless --path . -s res://scripts/probe3.gd

# UI 集成（真实场景装配 + 选卡 + Boss 击杀 + Game Over 信号链）
godot --headless --path . -s res://scripts/probe4.gd
```

全部输出 `ALL ... DONE` 即通过（本仓库交付前已在 Godot 4.5 stable 全绿）。

### 视觉回归（窗口模式截图）

```
godot --path . -s res://scripts/shot.gd
```

自动开局→落块→硬降，存 3 张 PNG 到 `user://`（约 5 秒，会闪一个窗口）。
配套像素分析脚本在仓库外层 `dev/analyze_shots.ps1`、`dev/asciiview.ps1`。

## 渲染注意事项（踩坑记录）

- **不要用 WorldEnvironment glow + hdr_2d**：mobile 渲染器上 2D HDR glow 后处理
  会把实心矩形吃成窄亮斑（已用最小复现验证）。霓虹效果全部用
  `board_view.gd` 里分层半透明矩形模拟，确定性渲染且手机更省带宽。
- **装配双向引用**：`board_view.game = game` 别漏（漏了的现象就是棋盘完全不画、
  无任何报错——_draw 里 game==null 直接 return）。

## 结构

```
project.godot / icon.svg / export_presets.cfg
scenes/Main.tscn
scripts/
  main.gd            装配：环境辉光/布局/输入路由/音乐驱动/存档
  game.gd            核心逻辑（纯状态机，含 dev_test）
  board_view.gd      棋盘+方块+Boss 绘制
  fx_layer.gd        粒子/冲击波/闪电/弹字/震屏/顿帧/闪屏
  audio_manager.gd   PCM 合成音效 + 16 步 BGM 序列器
  ui.gd              HUD/选卡/菜单/暂停/结算
  mini_piece.gd      HOLD/NEXT 预览
  touch_controls.gd  手势 + 虚拟按键
  star_bg.gd         星空/网格背景
  fx_stub.gd au_stub.gd test_runner.gd     逻辑测试
  probe3.gd probe4.gd integ_driver.gd      解析/集成测试
```
