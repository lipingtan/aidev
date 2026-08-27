# CR-5 三角色 Review 报告（2026-08-27）

> 对象：requirements.md / design.md / tasks.md；对照工程实查（db.gd / event_bus.gd / theme_tokens.gd / editorial.json / home.tscn）。
> 结果：11 项问题（高:0 中:6 低:5），用户确认全部修复，✅ 已落盘。

## 业务专家

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| B-1 | 中 | FR-6 daily 结构（tasks[]）与 design §3（tasks_data dict）不一致 | ✅ FR-6 改为 tasks_data 表述，与 design §3 对齐 |
| B-2 | 低 | FR-5 任务进度字段 {task_id,label,done} 与实际 get_tasks() 返回不符 | ✅ FR-5 改为 {id,label,done,progress,target} |
| B-3 | 中 | FR-1 banner[] 预置条目无任务承接（editorial.json 当前为空数组） | ✅ T4 Scope 补 editorial.json banner[] 预置 1 条 tetra_nova |

## 产品经理

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| P-1 | 中 | FR-4 列表行含图标，design §7 _make_row() 无图标节点 | ✅ design §7 补 icon TextureRect（meta.icon 空/不存在时隐藏）；T5 Scope 同步 |
| P-2 | 低 | FR-2 引用不存在的 banner_grad_start token；实查发现 ThemeTokens.grad("banner") 已存在（返回 Gradient） | ✅ FR-2/T4/design §6 统一为 `ThemeTokens.grad("banner")` + GradientTexture2D 真渐变色块 |
| P-3 | 低 | Banner 标题 ink 色叠加霓虹深紫渐变，对比度未验证 | ✅ T9 AC 增加 GUI 双主题截图核对项 |

## 架构师

| # | 严重度 | 问题 | 处置 |
|---|--------|------|------|
| A-1 | 中 | touch_daily() 跨天重置 `_daily = {date,plays}` 会抹掉 tasks_data（两路共用同一 dict/daily.json，当前无调用方属潜伏雷） | ✅ design §3 补共存规则 + touch_daily() 修订片段（跨天保留 tasks_data）；T1 Scope/AC 同步 |
| A-2 | 中 | _load_banners() 非幂等（未 clear），旧 SceneTreeTimer 不回收可致双轮播 | ✅ design §6 _load_banners() 开头 _items.clear() + _timer=null；T4 Constraints 同步 |
| A-3 | 中 | set_recommender() 只赋值不刷新，_ready 先跑时列表恒为空态，与 T5 AC 矛盾 | ✅ design §7 set_recommender 末尾补 is_node_ready 时 _render()；T5 Constraints 同步 |
| A-4 | 低 | "new" 模式空 version 排最前，M2 多游戏后排序语义可能误 | ✅ design §4 注释标注 M2 接入真实 version 后复核 |
| A-5 | 低 | RG 清单缺 target_gid 空点击、charts 空态两边界 | ✅ design §11 补 RG-29/RG-30；T9 AC 同步（RG-1~30） |

## 结论

✅ 三角色 review 完成，11/11 已修复落盘（requirements.md / design.md / tasks.md 同步修订，design §13 记录修订）。Q1~Q6/D1~D7 决策未受影响。
