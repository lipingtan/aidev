# Fix-1 双主题切换未真正生效（Theme 资源不传播 + Bg 底色硬编码）

> 发现方式：§5.5 视觉验证实际 GUI 截图实测（2026-08-24，tools/_shot2.gd + read_image）。
> CR-1 挂账项「双主题切换无布局跳变」在视觉层面不通过——主题资源链路整体失效。

## 当前行为（缺陷）

1. WHEN 用户切换到 elegant（青瓷）主题 THEN 页面 Label/Button/Panel 颜色**仍按 neon（白色文字等）渲染**——`ThemeTokens.apply_theme/restore` 只设置 `root.theme`（Window），而 Godot 4.5 实测 **root Window 的 theme 不向子 Control 传播**（预设场景加载前赋值同样无效；Label `get_theme()` 返回 null，颜色落到内置默认白色）。
2. WHEN 切换到 elegant THEN 基础背景**仍为霓虹深色**——`bglayer.tscn` 的 `Bg` ColorRect 硬编码 neon bg 值（0.02,0.024,0.059），`bglayer.gd::_apply()` 只更新网格可见性与 nebula 颜色，从不更新 Bg 底色。

## 期望行为（正确）

1. WHEN `apply_theme(name)` / `restore()` THEN 主题 Theme 资源同时作用于 root 与包含全部主题化控件的 Control 容器（`Main/$App`），Label/Button/Panel 按对应主题 token 渲染
2. WHEN 主题切换 THEN Bg 底色经 400ms Tween 过渡到当前主题 `bg` token（neon 深空 / elegant 米白）

## 不变行为（回归防护）

- WHEN neon 主题 THEN 页面文字白色、背景深色、网格线可见（现状保持）
- WHEN 主题切换 THEN BgLayer 仅响应 theme_changed，400ms Tween 过渡、nebula 颜色规则不变
- WHEN 重启 THEN profile 恢复上次主题（RG-4）行为不变

## 根因分析

| # | 文件/行 | 根因 |
|---|---|---|
| 1 | `shell/theme/theme_tokens.gd` apply_theme/restore | 只赋 `root.theme`（Window）。4.5 实测：root Window theme 不传播给子 Control（预设赋值同样无效）；把同一资源设到 Control 祖先（Home）立即生效 → 必须落到 Control 容器 |
| 2 | `shell/bglayer.gd::_apply()` + `bglayer.tscn Bg` | Bg 底色为 tscn 硬编码 neon 占位，代码无更新路径 |

## 影响范围

- `theme_tokens.gd`：新增 `theme_target`（Main._ready 置 `$App`），apply_theme/restore 经 `_set_theme_res()` 同时赋 root + target
- `main.gd`：_ready 在 restore 前设置 `ThemeTokens.theme_target = $App`
- `bglayer.gd`：_ready 取 `Bg` 引用；_apply 更新底色（instant 直设 / tween 400ms）
- 不改信号契约、不改 API 签名（home.gd 🎨 按钮调用方式不变）

## 验证

GUI 截图矩阵（tools/_shot2.gd，$env:APPDATA 隔离）：neon/elegant × 720×1560/720×1600 四张 + RG-4 重启恢复双相；elegant 下背景米白、文字墨色、无溢出；rect 跨主题一致（无跳变）。
