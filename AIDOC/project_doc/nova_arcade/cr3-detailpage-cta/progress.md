# CR-3 进度记录（完成：2026-08-25）

## 当前状态

- **CR-2 已关闭**：T1~T9 全绿、三角色验收通过、`cr2-launcher-gamemodule/acceptance_report.md` 落盘。
- **CR-3 已关闭**：requirements_plan → requirements.md → design_plan.md → design.md（两轮三角色 Review）→ tasks.md（T1~T9）→ 执行完成，三角色验收通过（2026-08-25）。

## 交付物

| Task | 文件 | 状态 |
|------|------|------|
| T1 | `services/recommender.gd` | ✅ |
| T2 | `shell/components/game_card.tscn/.gd`、`confirm_bubble.tscn/.gd` | ✅ |
| T3 | `shell/components/section_continue.tscn/.gd` | ✅ |
| T4 | `shell/pages/home.gd/.tscn`（继续游戏+为你推荐区块） | ✅ |
| T5 | `shell/pages/detail.tscn/.gd`、`cta_bar.tscn/.gd` | ✅ |
| T6 | `shell/main.tscn`（Recommender+ConfirmBubble 接线）、`core/nav.gd`（模态拦截） | ✅ |
| T7 | `tools/test_detail.tscn/.gd`（23/23 全过） | ✅ |
| T8 | `tools/_shot_cr3.tscn/.gd`（6 张截图无溢出，全套件回归全绿） | ✅ |
| T9 | PlayButton 移除，test_main/test_nav 回归全绿 | ✅ |

## 验收结论

- test_detail 23/23 ✅
- test_launch 25/25 ✅
- test_db × 3 ✅
- test_rg5 双相 6/6 ✅
- test_main 14/14 ✅
- test_nav 13/13 ✅
- test_smoke 12/12 ✅
- GUI 双轨截图（neon/elegant × 首页/详情页）无溢出 ✅
- RG-1~12 全绿 ✅

## 关键文件索引

- CR-3 文档：`cr3-detailpage-cta/requirements.md`、`design.md`、`tasks.md`
- 协议基线：`nova-arcade-client-design.md` §4.0/§4.2/§4.5/§4.6/§6；`dev-workflow-override.md`
- CR-3 代码入口：`services/recommender.gd`、`shell/pages/detail.gd`、`shell/pages/cta_bar.gd`、`shell/components/section_continue.gd`、`shell/pages/home.gd`
- 信号现状：EventBus 已有 records_updated / trial_consumed / payment_succeeded / dlc_installed；CR-3 接前两个，后两个占位 M3/M4

## 执行期环境备忘（累计踩坑）

- GUI exe 必须隔离 `$env:APPDATA = dev\tmp_gui`（默认 AppData 不可写 → signal 11 崩溃）；截图前清 `tmp_gui\Godot`
- headless：console exe + 隔离临时 APPDATA + 先杀孤儿 Godot 进程
- Godot 路径：见 `AIDOC/global-info/knowledge/steering/game/godot/godot-engine.md` §一-A
- tetra_nova `ui` 是 CanvasLayer 不随祖先 visible 隐藏（adapter._set_ui_visible 已处理）
- DB.get_record 返回活引用，测试比较前必须 duplicate()
- Recommender 是非单例，测试场景中无 Services 节点时会 push_warning（预期行为）
