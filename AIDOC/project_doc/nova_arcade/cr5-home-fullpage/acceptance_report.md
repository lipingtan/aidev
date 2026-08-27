# CR-5 三角色验收报告（执行后）

> 验收对象：CR-5 T1~T9 全部实现代码 + test_home 27 断言 + 全量回归 14 场景 + GUI 双主题截图。
> 验收方式：三角色（产品/架构/Skill专家）逐项核对 requirements.md FR、design.md §、tasks.md AC，以实机 headless/GUI 输出为准。
> 日期：2026-08-27

## 产品经理视角（FR / 验收标准可测性）

| 项 | 结论 | 证据 |
|----|------|------|
| FR-1 banner[] 预置 tetra_nova | ✅ | editorial.json banner[1]；GUI 截图标题 "TETRA NOVA" 可见 |
| FR-2 Banner 渐变色块（token 驱动） | ✅ | `ThemeTokens.grad("banner")`+GradientTexture2D；双主题截图渐变随主题变化（neon 靛紫/elegant 薄荷粉） |
| FR-3 热门榜三 Tab + 排名 + 人数格式 | ✅ | test_home RG-23 rank=1 players≥10000；GUI "综合/新游/好评" + "1.0万+" |
| FR-4 每日任务 3 条 + 完成态 + 奖励 Toast | ✅ | RG-24/25/26 全绿；GUI "每日任务 0/3" 三行 |
| 验收标准可测性 | ✅ | 每条 FR 对应 test_home 断言（RG-21~30），headless 可复现 exit 0 |

## 架构视角（design § / 约束合规）

| 项 | 结论 | 证据 |
|----|------|------|
| DailyTaskService 非 Autoload，挂 Main/Services | ✅ | main.tscn Services/DailyTaskService；home.gd get_node_or_null 注入（T8 AC 通过） |
| touch_daily 跨天保留 tasks_data（§3 共存） | ✅ | test_home "touch_daily 跨天保留 tasks_data" PASS |
| _load_banners 幂等（§12，A-2） | ✅ | _items.clear()+_timer=null；call_deferred 规避 autoload 时序 |
| set_recommender 末尾 _render()（A-3） | ✅ | section_charts.gd；RG-30 无 Recommender → EmptyState PASS |
| 行数约束 | ✅ | daily_task_service 99(≤150) / section_banner 99(≤100) / section_charts 110(≤150) / section_daily 56(≤80) / home.gd 70(≤80) |
| 禁单行 lambda / 具名比较器 | ✅ | charts 三比较器 + _calc_players 均具名方法 |
| 主题色走 ThemeTokens，禁硬编码 | ✅ | banner/dots/rank 全 token；GUI 双主题对比度 neon 12.38/elegant 9.80（>4.5） |
| 不破坏 CR-3/CR-4（回归） | ✅ | test_main/test_category/test_detail/test_smoke 独立 APPDATA 全绿；RG-28 on_resume 刷新 PASS |

## Skill 专家视角（工作流 / 防失忆规则）

| 项 | 结论 | 证据 |
|----|------|------|
| 每 CR 必写 progress.md | ✅ | cr5-home-fullpage/progress.md 已落盘（含执行期决策 + 遗留项） |
| 子代理报"全绿"主代理实机复验 | ✅ | 主代理独立跑 test_home(27)+14 场景，非采信子代理结论 |
| headless 隔离 APPDATA + 杀孤儿 godot | ✅ | 每场景独立目录；taskkill 前置；test_db 三相同目录连续 |
| GUI 截图核对 Banner 对比度（P-3） | ✅ | _shot_cr5.gd 采样 BannerRect 背景 vs ink，双主题 ratio 输出 + vision 目检 |
| 发现并修复执行期缺陷（非仅"跑通"） | ✅ | set_corner_radius_all / TextureRect 改 Panel / call_deferred / stretch_mode=0 / setup 时序 / lambda 捕获 / 测试挂 root，共 7 处 |

## 三角色结论

**通过。** 无 High 级阻塞项。Medium/Low 遗留（banner 真实图片、EventBus.tasks_updated 广播、rated 模式数据）均为 M2 范围，已记入 progress.md 遗留项。

### 执行期发现并修复的缺陷汇总
1. `set_corner_radius(3)` → 4.7 需 2 参，改 `set_corner_radius_all(3)`（section_banner.gd 编译失败致脚本未挂载）
2. BannerRect Panel→TextureRect（Panel 无 texture 属性，GradientTexture2D 赋值崩）
3. `_load_banners` 同步读 editorial 为空 → `call_deferred`（Registry.reload 在 autoload _ready）
4. stretch_mode=2(KEEP)→0(SCALE)，否则渐变只占 64×64 左上角（GUI 目检发现"窄条"）
5. card.setup() 须 add_child 之后（@onready 时序），home.gd + category.gd:106 统一修复
6. test_home `.instantiate()` 缺 `as` 类型转换 → 后续方法调用解析失败挂起
7. lambda 按值捕获局部变量 → tasks_updated 信号改用成员变量 `_signal_fired`
