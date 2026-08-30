# Review Report: CR-7 M2 盒子级能力收尾（成就墙 + 本地推荐画像）

> 日期：2026-08-28
> 范围：requirements.md (FR1~FR6) + design_plan.md + design.md
> 模式：三角色 review（业务专家/产品经理/架构师），适配 nova_arcade Godot 项目场景

---

## 汇总问题清单（共 11 项）

| # | 来源 | 严重度 | 问题描述 | 修复方案 | 状态 |
|---|------|--------|---------|---------|------|
| 1 | B-1 | **高** | result_overlay._fill_achievements() 显示 aid 字符串而非 def.name | _fill_achievements 改为查 Registry.ach_def → def.name，null 回退 aid | ✅ 已修复（design.md §2.3 补充说明） |
| 2 | P-1 | **高** | toast_layer.show_msg() 无 per-message 颜色能力 → 金色 Toast 需扩展 API | show_msg(msg, color) 增加可选 color 参数；Engine 传 ThemeTokens.color("gold") | ✅ 已修复（design.md §4.3） |
| 3 | P-2 | **高** | toast_layer 无消息排队 → Toast 与结算卡可能重叠 | 增加 _queue Array + _busy flag，tween.finished 触发 _pop_queue | ✅ 已修复（design.md §4.3） |
| 4 | A-2 | **高** | Registry.reload() 读 achievements.json 无错误处理 → 格式异常崩溃 | 增加 file_exists + parse_string 类型校验 + ga.has("id") + push_warning | ✅ 已修复（design.md §3.2） |
| 5 | B-2 | **中** | _calc_total_hours() 用 total_playtime 字段，需确认累加式写入 | 核实 launcher_util.finish_record() 写入累加式 total_playtime → 正确 | ✅ 已确认（design.md §4.1 注释） |
| 6 | B-3 | **中** | _fallback_editorial() 数据结构未对齐 for_you() 格式 | charts("all") 返回转 {"meta","record"} 格式（丢弃 rank/players）+ slice(0,4) | ✅ 已修复（design.md §5.2） |
| 7 | P-3 | **中** | 收藏家阈值 ≥5 但盒子仅 4 款 → 玩家永远无法解锁 | GLOBAL_DEFS desc 改为"拥有全部游戏记录"，_check_global 用 Registry.all().size() | ✅ 已修复（design.md §2.2） |
| 8 | P-4 | **中** | for_you() top 4 在 HBoxContainer 可能挤压过窄 | home.tscn 已有 ScrollContainer 包裹 → 布局正常；GUI验收核对 | ✅ 已确认（design.md §5.3 补充说明） |
| 9 | P-5 | **低** | library.tscn 无 RootVBox/Header/AchievementWall → @onready 引用报错 | library.tscn 场景树变更：新增 RootVBox/ScrollView + Header + AchievementWall | ✅ 已修复（design.md §4.1/§4.2） |
| 10 | A-3 | **中** | for_you() 被频繁调用无缓存 → O(N²) 遍历浪费 | 增加 _cached_result + _cache_dirty flag，EventBus.records_updated 失效 | ✅ 已修复（design.md §5.2） |
| 11 | A-1/A-5 | **高/中** | Engine._ready() 信号连接时序 + section_achievement_wall instantiate | _enter_tree() 替代 _ready()；library.tscn 预建节点挂载 class_name 脚本 | ✅ 已修复（design.md §2.2/§4.2） |

**统计：** 高 5 项，中 5 项，低 1 项 → 全部修复

---

## 代码交叉验证结果

| 文档声明 | 实际代码状态 | 结论 |
|---------|------------|------|
| event_bus.achievement_unlocked(def, points) 已就位 | ✅ event_bus.gd:28 signal 定义存在 | 一致 |
| Registry.ach_def(aid) 跨游戏查询 | ✅ registry.gd:199 遍历 all() | 一致，需扩展全局成就 |
| db.gd unlock(aid) 幂等 + 强写 | ✅ db.gd:176 has(aid) 检查 + write_json | 一致 |
| recommender.for_you() 当前字母排序 | ✅ recommender.gd:34 all_metas.sort_custom title.to_lower | 一致，需替换为画像加权 |
| home.gd 已消费 for_you() | ✅ home.gd:45 _render_for_you() 调 _recommender.call("for_you") | 一致，零改动可行 |
| library.gd CR-1 骨架 | ✅ library.gd:4 on_enter(print) | 需大幅扩展 |
| launcher_util.finish_record 写入 total_playtime | ✅ 累加式写入：int_field + result.playtime | design.md §2.2 正确 |
| toast_layer 无排队/颜色参数 | ✅ show_msg(msg) 单参数，直接 kill tween | 需扩展（P-1/P-2） |

---

## 修订记录更新

design.md 修订记录已追加：
```
| 2026-08-28 | 多角色review修复（B1/B2/B3/P1/P2/P3/P4/P5/A1/A2/A3/A5） | 三角色review汇总
```

---

## 结论

**✅ 三角色 review 完成，全部 11 项问题已修复，design.md 可进入 tasks.md 阶段。**

修复涉及文档变更：
- design.md §2.2: _enter_tree() + 动态收藏家阈值
- design.md §3.2: Registry reload 错误处理
- design.md §4.1: total_playtime 确认注释 + library.tscn 场景树补充
- design.md §4.2: section_achievement_wall class_name + tscn 预建节点
- design.md §4.3: toast_layer color 参数 + 消息队列
- design.md §5.2: _fallback_editorial 完整实现 + 结果缓存
