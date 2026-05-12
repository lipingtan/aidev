# 动作游戏 · 角色控制器

## 目标

在 **CharacterBody3D**（或项目选定方案）上实现：地面/空中移动、朝向、与 **相机相对** 的输入方向，并与动画根运动策略一致（根运动 vs 程序位移）。

## 要点

| 项 | 说明 |
|----|------|
| 输入 | 原始输入 → 相机基向量投影 → `velocity` 水平分量 |
| 地面 | `floor_snap`、最大坡度、`is_on_floor()` 防抖（Coyote 可选） |
| 重力与跳 | 统一在 `_physics_process`；跳跃可消耗 `coyote_time` / 缓冲跳 |

## 与性能

- `QualitySettings.get_scalar(&"max_skinned_actors")` 不直接改控制器，但同屏角色多时须控 AI 更新开销。

## 验收

- [ ] 斜坡/台阶不穿模、不抽风抖动
- [ ] 锁定开启时「前」指向目标与相机协调（见 `lock-on.md`）

## 常见陷阱

- `move_and_slide` 与动画根运动重复叠加导致滑步加倍
