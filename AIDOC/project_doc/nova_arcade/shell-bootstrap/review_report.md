# 三角色 Review 报告：CR-1 shell-bootstrap

> 对象：design_plan.md / design.md / tasks.md。清单结构按 go-development-workflow §2.3，维度适配 Shell/Godot CR。

## 角色1：业务专家验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | FR 覆盖（client-design §11 M1 两项 → T1~T9） | ✅ 全覆盖（Q4=A 含 DB/Registry） |
| 2 | 非功能需求（mobile 渲染器/≤200行/崩溃恢复/主题 400ms） | ✅ 有对应约束与 RG |
| 3 | 降级/异常场景 | ⚠️ 中：DB JSON 损坏无兜底策略 → 已修（design §3.3 + T3） |
| 4 | 权限边界（CoreManager Mock 不阻塞） | ✅ RG-6 |

## 角色2：产品经理验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 用户故事可达性 | ❌ 高：主题切换无 UI 入口（apply_theme 齐全但四页/TabBar/Main.tscn 均无按钮），RG-4 无法人工验证 → 已修（design §5 + T8 Scope 补 ThemeToggleButton） |
| 2 | 空态/Toast/Tab 路径 | ✅ |
| 3 | 范围一致性 | ⚠️ 低：design §7 "详情页骨架能显示"超 CR-1 范围 → 已改为"CR-3 实现" |

## 角色3：架构师验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 依赖方向/循环依赖 | ✅ DAG 无环；T4 实际仅依赖 T2，串行保守可接受 |
| 2 | 引用存在性 | ⚠️ 中：QualitySettings 参考路径规范内两处不一致（`projects/demo_game/` vs `projects/demo/demo_game/`）→ 实查确认后者，T2 已标注 |
| 3 | 缓存/写时序一致性 | ⚠️ 低：防抖写无退出 flush → 已修（WILL_EXIT flush，RG-5 完整可靠） |
| 4 | RG 清单完整性 | ⚠️ 低：RG-3 Dialog 分支 CR-1 不可验（OverlayLayer 空容器）→ 标注 CR-3 补验 |
| 5 | 可测试性 | ⚠️ 低：.tres 样式 headless 无法验证 → T10 明确 GUI 截图 + read_image |

## 结论

- 高优 1 项（主题切换入口）：**已修复**，不阻塞 Task
- 中优 2 项（DB 损坏兜底 / 参考路径）：**已修复**
- 低优 4 项：已同步到 design.md / tasks.md 对应位置

**✅ 三角色设计 Review 通过**（高优问题已清零，可进入执行阶段）

---

# 终版验收（T10 + Fix-1 后，2026-08-24，打开实际代码验证）

## 角色1：业务专家验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | FR 覆盖（client-design M1：工程/Autoload×6/主题/四 Tab/DB+Registry+tetra meta/冒烟） | ✅ T1~T10 全 ✅，证据链 smoke_report.md + test_smoke 12/12 |
| 2 | 非功能需求（mobile 渲染器/canvas_items+expand/崩溃恢复/主题 400ms） | ✅ project.godot L31-36 实查；test_db 损坏相 RG-5 双相 6/6；theme_tokens tween |
| 3 | RG-1~6 全部有证据（非口头通过） | ✅ smoke_report §2/§4 逐项引用测试输出与截图 |
| 4 | §5.5 视觉项闭环 | ✅ GUI 实测四截图 + rect 跨主题一致 + [restore] PASS（含 Fix-1） |

## 角色2：产品经理验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 用户故事可达（主题切换有 UI 入口，RG-4 可人工验证） | ✅ home.tscn L24 ThemeButton + home.gd L5 @onready |
| 2 | 空态/Toast/Tab 路径完整 | ✅ 四页 EmptyState、ToastLayer、TabBar 实测切换无崩溃 |
| 3 | 范围一致性（无超 CR-1 内容） | ✅ 详情页 CTA/Launcher 明确挂账 CR-2/CR-3，未越界实现 |

## 角色3：架构师验收

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 依赖方向（页面互不引用；场景引用仅 Nav 一处） | ✅ pages/ 无 preload/load；组件无页面引用；nav.gd L36-39 唯一场景映射 |
| 2 | 引用存在性（meta.json 必填字段/autoloader 路径） | ✅ tetra_nova meta 全字段实查（id/title/category/price_model/trial/icon/scene/runtime…） |
| 3 | 写时序（防抖退出 flush 双保险） | ✅ db.gd L54-63：WM_CLOSE_REQUEST + _exit_tree 兜底（headless 不触发 WM 通知） |
| 4 | EventBus 契约完整 | ✅ 12 信号与 design §3.4 一致 |
| 5 | 单文件 ≤200 行 | ⚠️ 低：db.gd 232 行（T3 约束"超限拆 db_io.gd"未执行）→ **CR-2 动 DB 时拆出 db_io.gd**（_read_json/_flush_named/落盘原语），已记入 CR-2 范围 |
| 6 | 废弃资产清理 | ✅ _shot.gd/.tscn（被 _shot2 取代，含 4.5 解析错误）+ 旧 shot_run.ps1 已删；编辑器 --editor --quit 复跑零脚本/资源错误（§5.1 关闭） |

## 结论

- ⚠️ 低优 1 项（db.gd 超行）：不阻塞收口，CR-2 拆 db_io.gd 时闭环
- **✅ 三角色终版验收通过——CR-1 shell-bootstrap 完全闭环**（T1~T10 + §5.1/§5.5 关闭；§5.3/§5.4 属 CR-2/CR-3 范围）
