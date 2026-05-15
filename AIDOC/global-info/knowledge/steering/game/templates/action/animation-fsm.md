# 动作游戏 · 动画状态机

## 目标

游戏逻辑状态（Idle/Move/Attack/Hit/Stun）与 **AnimationPlayer / AnimationTree** 同步；过渡时间与 **可操作窗口** 统一配置。

## 要点

| 项 | 说明 |
|----|------|
| FSM | 优先 **代码状态机**（与 `state_machine` 知识库一致），动画树只做 Blend |
| 参数 | 速度、方向角、武器类型 → `AnimationTree` set；禁止场景里硬编码字符串散落 |
| 事件帧 | 攻击出框、脚步音效用 **Animation Method Track** 或时间配置表 |

## 验收

- [ ] 任意状态可被全局打断（死亡、播片）且不锁死树
- [ ] 切场景时 AnimationTree 状态可重置

## 常见陷阱

- 逻辑已进入下一状态，动画过渡未完成导致输入「能吃不能放」
