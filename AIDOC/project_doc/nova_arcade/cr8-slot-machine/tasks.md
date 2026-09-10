# 任务分解：CR-8 幸运轮盘押注机

> 依据 design.md §8 里程碑。工程路径：`projects/nova_arcade/nova-arcade/`。
> 进度标记：⬜ 未开始 / 🔄 进行中 / ✅ 完成。

| # | 任务 | 状态 | Acceptance（可验证钩子） |
|---|------|------|------|
| T1 | 脚手架：meta.json + icon.png + module.tscn + adapter + src（goslot 素材版） | ✅ | Registry.lookup("slot_machine") 返回完整 GameMeta（title/category=casual/scene 指向 module.tscn）；--headless --import 通过无解析错误 |
| T2 | SlotConfig 常量（goslot 1:1）+ 余额账本（load/save/救济/上限） | ✅ | headless 探针：存档往返一致；救济 +50 落盘；999999 上限 clamp 生效 |
| T3 | 面板渲染（24 格 goslot 布局 + board 底图）+ 押注交互（押注位/注额档/长按清零/登记制余额校验） | ✅ | headless：押注表状态断言（加注/清零/超额拒绝/未押禁旋转/押注不扣款）；GUI 探针截图面板布局正确 |
| T4 | 跑灯开奖（goslot 三段变速公式+Lucky 散花）+ 结算（多格合计） | ✅ | headless：给定 stop 与押注组合逐例断言 ≥10 例（含连击路径）；tween 结束后停位符号正确 |
| T5 | 开始遮罩 + 回菜单常驻按钮 + goslot 音效 + 救济按钮 + 跑灯期交互锁 | ✅ | boot 不自动开局（遮罩可见）；点开始进 IDLE_BET；SPINNING 期按钮 disabled；quit_to_shell 字段齐全；quit 后 UILayer 全隐藏（A-1）、play_again 后恢复 |
| T6 | headless 测试套件 tools/test_slot_machine.tscn（配置+押注+结算逐例+存档+救济+成就+reset_run） | ✅ | 42 断言全绿退出码 0；无孤儿 Godot 进程 |
| T7 | 万局统计断言组（固定 seed，io_silent，跳过动画直调抽签+结算） | ⬜ | 符号级频率 1/6±3%；基础 RTP∈[85,92]；连击次数>0；六符号全命中 |
| T8 | GUI 验证 tools/_gui_cr8.{gd,tscn}：双主题 × 五态截图（遮罩/押注/跑灯/开奖/救济） | 🔄 | neon 五态全 PASS + GLM 视觉复核通过；**elegant 主题未跑** |
| T9 | editorial.json banner +1（category=casual 落分类页既有「休闲」桶）+ Registry.query/Searcher 集成验证探针 | ⬜ | query(category=casual) 命中含 slot_machine；Searcher「轮盘/laohuji/xl」命中；banner 轮播含 slot_machine；分类页「休闲」桶可见性断言（B-3） |
| T10 | 全量回归（launch/smoke/nav/四游戏+新套件）+ acceptance_report.md + 提交 | ⬜ | 既有套件全绿；AC-1~9 核对表齐；三角色验收通过；commit |

## 依赖顺序

T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9 → T10（串行；T3/T4 可在 T2 后并行但单人执行按序）

## 全局 Acceptance 约定

- 每任务 headless 验证须隔离 APPDATA（Windows 原生路径形式）+ 跑后孤儿进程清零
- GUI 截图证据存 dev/（任务收尾清理，acceptance 引用的保留至 T10 后统一清）
- src 行数红线：slot_main.gd ≤380 行、module_adapter.gd ≤150 行、slot_config.gd ≤60 行
- 测试断言禁「只数 PASS」：每组独立计数，FAIL 任何一条即任务不通过

---

## 执行进度记录（2026-08-31 暂停）

- 形态三次迭代：12 格转轮 → 24 格自绘 → **goslot 1:1 移植**（用户提供完整源码项目，素材+规则移植、代码按 4.7 重写）
- T1~T6 完成：headless 42 断言全绿
- T8 部分完成：neon 主题五态截图+GLM 视觉复核通过；**elegant 未跑**
- T7 已并入 T6 统计组；T9/T10 未开始
- 暂停点：底部押注位对齐 board 白框已完成，待续 elegant 验证
