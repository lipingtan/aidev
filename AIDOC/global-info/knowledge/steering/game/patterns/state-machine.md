# 状态机模式（游戏实现向）

## 概念说明

对象行为拆成 **有限状态 + 转移条件 + 进入/退出钩子**。逻辑状态机与动画表现解耦：状态机输出 **`state_id` + 参数**，驱动 AnimationTree。

## 适用场景

- 角色：Idle / Locomotion / Attack / Hit / Dead
- AI：Patrol / Chase / Attack / Flee
- UI：页签、弹窗栈（可有单独栈式 FSM）

## Godot 要点

- **优先** 脚本内 `enum State + match` 或小型 `StateMachine` 资源，避免过深继承。
- 转移条件集中在一处（如 `can_transition_to`），便于断点与日志。
- `physics_process` 里做移动，`process` 里做纯表现；状态切换若在物理帧，注意与输入缓冲对齐。

## 常见陷阱

- 状态里嵌套 `await` 导致退出状态后协程仍改写变量。
- 忘记 `exit` 清理：Timer、信号连接、临时 Modifier 未卸下。

## 关联

- `templates/action/animation-fsm.md`
- `addons/gd_ecs/state_machine/`（demo 工程）
