# 动作系统模板

| 文件 | 内容 |
|------|------|
| `character-controller.md` | 移动、地面检测、与相机相对方向 |
| `animation-fsm.md` | 状态驱动动画、AnimationTree 与游戏状态同步 |
| `combo-system.md` | 输入缓冲、窗口、连段数据 |
| `dodge-block.md` | 无敌帧、精力、与伤害系统契约 |
| `lock-on.md` | 目标选择、硬锁定与软瞄准 |
| `camera-control.md` | 跟随、遮挡回收、锁定混合 |

与 **`performance-budget.md`**：移动端同屏角色、粒子倍率受 `QualitySettings.get_scalar()` 约束；动画与 LOD 在 `asset-pipeline.md`。
