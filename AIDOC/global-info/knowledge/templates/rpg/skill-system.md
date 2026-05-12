# RPG 技能系统

## 目标

技能 = **数据（SkillData）** + **施放管线**（前摇 / 生效帧 / 后摇），与输入缓冲、冷却、资源消耗一致。

## 要点

| 项 | 说明 |
|----|------|
| SkillData | id、消耗（MP/体力）、冷却、动画名、射程、碰撞层或投射物 |
| 施放状态机 | Idle → Windup → Active → Recovery；Active 内生成 hitbox 或弹道 |
| 表现 | 动画与 Hitbox **时间轴配置**，避免魔法数字散落脚本 |

## 与动作模板

- 输入：`input_buffer` + `combo-system.md` 决定「当前接入哪段 skill」
- 朝向：`lock-on.md` / `character-controller.md` 提供面向向量

## 验收

- [ ] 冷却与消耗在存档侧一致
- [ ] 打断规则明确（受击、闪避、是否取消 Recovery）

## 常见陷阱

- 用动画时长代替冷却配置，调动画后平衡全崩
