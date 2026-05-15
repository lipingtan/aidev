# NSFW 执行协议补充

> **本文件是 `execution-protocol.md` 的 NSFW 扩展**。
> 仅当 `clarification.md` 声明项目含 NSFW/R-18 内容时生效。
> 通用执行协议请参见 `../execution-protocol.md`。

---

## 一、NSFW 规范触发时机速查

| 规范 | 触发时机 | 触发条件 |
|------|----------|----------|
| `content-guidelines.md` | Phase 0 澄清 + 每次创建角色 + 每次写 NSFW 内容 | 项目含 NSFW 标记 |
| `platform-policies.md` | Phase 5 发布前 | 项目含 NSFW 标记 |
| `multi-version-strategy.md` | Phase 0 确认 + Phase 5 打包 | 项目含 NSFW 标记 |

---

## 二、Phase 0 NSFW POST-CHECK

- [ ] `clarification.md` 中已填写目标平台合规声明（Steam AO / itch.io 等）
- [ ] `clarification.md` 中已填写 All-Ages vs R-18 版本发布策略

---

## 三、Phase 5 NSFW POST-CHECK

- [ ] All-Ages 版本可独立运行（无 DLC 时无报错）
- [ ] Adult DLC 加载后所有 H-Scene 可正常触发
- [ ] 马赛克开关在日本地区测试可用
- [ ] 各目标平台导出包已按 `multi-version-strategy.md` 打包矩阵验证
- [ ] 所有角色档案的 `declared_adult` 字段已最终核查

---

## 四、NSFW 内容合规 POST-CHECK

> 涉及 H-Scene、角色设定、NSFW 资产描述、提示词生成时执行。

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | **角色年龄声明** | 场景中涉及的所有角色档案有 `declared_adult: true` 字段 |
| 2 | **无未成年化描述** | 代码/注释/文档中无 "childlike"、"young-looking" 等未成年暗示词（在 NSFW 语境中） |
| 3 | **触发条件含合规检查** | H-Scene 触发器调用了 `character.compliance.declared_adult` 检查 |
| 4 | **DLC 门控完整** | NSFW 内容入口有 `DLCManager.is_adult_dlc_active()` 判断 |
| 5 | **全年龄 fallback 存在** | 每个 NSFW 触发点有对应的 All-Ages 替代动画/分支 |
| 6 | **马赛克开关已接入** | 明确性描绘的视觉元素已接入 `CensorshipManager` |
