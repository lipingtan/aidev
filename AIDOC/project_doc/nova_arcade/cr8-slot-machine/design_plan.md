# 设计计划：CR-8 幸运轮盘押注机

> 输入：requirements.md（已确认）。本文档说明 design.md 将如何组织，先行圈定技术决策面，交用户确认后再展开 design.md 全文。

## 计划产出的 design.md 章节

1. **架构与文件布局**
   - `games/slot_machine/`：meta.json / module.tscn / module_adapter.gd / src/{slot_config.gd, slot_main.gd} + icon.png
   - module.tscn 结构：Module(GameModule, 普通 Node) → Src(Node2D 面板绘制) + UILayer(CanvasLayer：押注区/中央提示/开始遮罩/回菜单)
   - 设计决策：面板符号用 Node2D 自绘（绕圈摆位用三角函数算坐标），UI 层用 Control——与 magic_tower（Node2D 地图）+ snake/2048（纯 Control）的既有两派模式一致，轮盘的圆形摆位天然适合 Node2D

2. **核心数据结构**（design.md 给出完整定义）
   - SlotConfig：SYMBOLS[6]（字符/颜色 key/赔率）、POSITIONS[12]（符号索引排列）、BET_STEPS[1,5,10]
   - 状态机：IDLE_BET → SPINNING → SETTLE →（救济分支 IDLE_BROKE）
   - 余额账本接口：load/save（save.cfg, ConfigFile）

3. **关键算法**
   - 滚动动画：Tween 旋转 Src 内 wheel 节点（或逐符号位相位推进），ease out，终点对齐预抽结果；符号预建池不逐帧 new
   - 开奖抽签：RandomNumberGenerator（seed 可注入，供统计测试复现）→ 0..11 均匀 → POSITIONS 映射符号
   - 结算：遍历押注表 → 押中份额 × 赔率（连击再 ×2）→ 余额更新 → 强写

4. **UI/UX 细节**
   - 双主题适配：全部颜色经 ThemeTokens.color(key)；押注位/按钮焦点 FOCUS_NONE；滚动期锁交互
   - 连击提示动效：中央 Label 缩放+渐隐 Tween
   - 开始遮罩：标题/说明/开始按钮（对齐 c484ded 约定），boot 不自动开局
   - 回菜单：adapter 侧 CanvasLayer(layer=20) 常驻（对齐 7259589 约定）

5. **测试设计**
   - `tools/test_slot_machine.tscn` headless：生命周期 ≥12 断言 / 赔付逐例 ≥10 例 / 万局统计（固定 seed，断言区间来自 requirements AC-3）/ 存档往返 / 救济 / 成就构造局
   - GUI 探针 `_gui_cr8`：双主题五态截图（复用 CR-6/CR-7 harness 模式，save_png 到 dev/）
   - 已知坑位清单入 design（heads-up）：CanvasLayer 遮罩/anchors size=0/Container 忽略手动 position/visible 断链——均为本项目 GUI 实测教训，design 直接规避

6. **T 任务分解草案**（design 确认后展开为 tasks.md）
   - T1 脚手架（meta/icon/adapter/module.tscn 空 src）→ T2 SlotConfig+账本 → T3 面板渲染+押注交互 → T4 滚动开奖+结算+连击 → T5 遮罩/回菜单/音效 → T6 headless 测试套件 → T7 统计断言 → T8 GUI 验证+双主题 → T9 editorial banner+Registry 集成验证 → T10 全量回归+acceptance_report

## 需要用户拍板的技术选项（含推荐）

- [DQ-1] 面板渲染层：**Node2D 自绘（推荐）** vs 纯 Control+手动坐标。理由：圆形摆位与旋转动画在 Node2D 下更自然，magic_tower 已验证 GameHost 下 Node2D 渲染路径（含 visible 断链的框架级修复已就位）。
- [DQ-2] 滚动实现：**整轮节点旋转（推荐，一个 Tween 转一个 Node2D）** vs 逐符号位移。理由：整轮旋转只动一个节点，停位=角度取模，实现和测试都最简。
- [DQ-3] 统计测试 seed：**固定 seed 注入（推荐）** vs 真随机跑万局。理由：CI/回归可复现，无偶发挂。

确认后进入 design.md 全文编写。
