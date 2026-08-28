# 设计计划：CR-6 M2 小游戏批量接入（魔塔 + 2048 + 贪吃蛇）

## 设计方向

三款游戏全部走 tetra_nova 既定模板：`games/{id}/` = meta.json + module.tscn + module_adapter.gd + src/ + icon.png。module.tscn 根节点挂 module_adapter.gd（extends GameModule），src/ 放游戏本体场景与脚本；adapter 只做协议桥接（boot 注入 save_dir、quit_requested 装配 result、pause/resume 转发），不写玩法逻辑。

核心策略：
1. **逻辑与视图分离**——每款游戏的核心规则写成纯函数/独立 RefCounted 类（可 headless 单测），视图层只负责渲染与输入转发。魔塔 demo 已按此模式（MT_SELFTEST 纯逻辑自检），2048/贪吃蛇照此实现。
2. **Shell 零改动**——Registry/Launcher/Nav/结算卡/继续游戏全部复用；本 CR 仅新增 games/ 目录 + editorial.json 数据 + tools/ 测试套件。
3. **存档同构**——每款游戏写自己的 `save.cfg`（JSON：best_score 等）到 ctx.save_dir，adapter 在 quit_requested 上报 score/playtime，Shell DB 记录 total_playtime/best 供继续游戏与结算卡使用（tetra 模式）。

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| A: 每款游戏独立 src/ 场景 + adapter 桥接（tetra 模式） | 与 M1 管线一致，Launcher 零改动；src/ 可独立 headless 测试 | 每款多一个 adapter 文件（~50 行） | ✓ |
| B: 游戏直接 extends GameModule，无 adapter | 少一层文件 | 玩法脚本混入协议代码，违反 override §1「适配器只依赖基类 + src」；tetra 先例已定 | ✗ |
| C: 2048/贪吃蛇 用 HTML5 runtime（airwar 模式） | 复用 airwar 管线 | HtmlRunner 未实现（M2 末独立 CR），阻塞本 CR | ✗（顺延） |

## 澄清问题

- [Question-1] 2048/贪吃蛇的输入方案？
  - 业界最佳实践：竖屏休闲游戏主流是「键盘方向键 + 触屏虚拟按键」双通道；2048 另有滑动手势（原生 2048 网页版即滑动）。滑动识别需自实现手势判定（位移 阈值 + 方向 dominance），约 30 行。
  - 推荐答案及理由：**2048 = 键盘方向键 + 触屏滑动**（贴原生体验）；**贪吃蛇 = 键盘方向键 + 右下角虚拟 D-pad**（四向高频操作，D-pad 比滑动更稳）。理由：两款游戏输入模式不同，各取最优；D-pad 复用魔塔 demo 的十字pad 视觉风格（Phoenix 主题图），保持盒子内一致性。
  [Answer-1]
- [Question-2] magic_tower 移植时 demo 自带 UI（右侧 A/B/C/D + Start/投币/退出按钮）如何处理？
  - 业界最佳实践：接入盒子的游戏统一由 Shell 控制生命周期（暂停键/回盒），游戏内只保留玩法必需控件；tetra D5=A 先例「adapter 追加『回菜单』按钮，src 零 diff」。
  - 推荐答案及理由：**保留十字盘 + A/B/C/D（玩法必需）；移除 demo 的 Start/投币/退出 三个街机按钮**（盒内无投币语义），adapter 追加「回菜单」按钮（tetra D5=A 模式，src 最小 diff）。理由：与 tetra 接入先例一致，Launcher 生命周期统一。
  [Answer-2]
- [Question-3] 三款游戏的视口规格？
  - 业界最佳实践：盒子基准 720×1560 竖屏（override §2）；网格类小游戏内容区居中 + 上下留白放 HUD，不强行铺满。
  - 推荐答案及理由：**三款 meta.json 均声明 viewport_size=[720,1560]、orientation=portrait**；魔塔 640×640 棋盘按 canvas_items 等比缩放居中，上下留白放 HP/ATK/DEF/Gold HUD 与十字盘。理由：与 Shell stretch/expand 模式兼容，横屏需求归 Arcade PoC（M2 末）。
  [Answer-3]
- [Question-4] icon.png 如何生成？
  - 业界最佳实践：盒子内图标统一霓虹风格（tetra icon 先例）；手工 PS 不可持续，程序化生成为准。
  - 推荐答案及理由：**headless GDScript Image 脚本生成**（512×512：渐变底 + 主题 glyph 文字/简单图形，neon 配色走 ThemeTokens 色值），一次跑三张落盘 icon.png。理由：程序化可复现、风格统一，不依赖外部工具链。
  [Answer-4]
- [Question-5] 每款游戏 headless 测试套件结构？
  - 业界最佳实践：逻辑层纯函数断言（合并/生长/碰撞规则）+ 协议层生命周期走查（boot/pause/resume/quit/result 字段/save_dir 写入），与 CR-3/4/5 的 test_{game}.tscn 模式一致。
  - 推荐答案及理由：**每款 `tools/test_{game}.tscn`：≥10 断言**（逻辑规则 ≥6 + 协议走查 ≥4）；魔塔额外跑 MT_SELFTEST 全量；最后全量回归（既有 14 场景 + 新 3 场景，独立 APPDATA）。理由：与既有验收口径一致，防「只跑通不验证」。
  [Answer-5]

## 风险点

- [Risk-1] magic-tower-godot 为 Godot 4.5 工程，导入 4.7.2 可能遇 API 变更（如 ScreenOrientation 枚举改名已踩坑先例）——按 kb/godot-4.7-api-facts.md 逐项修复，逻辑层零改动为原则
- [Risk-2] 触屏滑动识别阈值不当导致 2048 误触/漏触——参数走 @export 常量，GUI 验收时实测调整
- [Risk-3] Registry.reload() 扫描 games/ 后首页「继续游戏」区块出现多条目，需确认 SectionContinue 渲染逻辑对多游戏无布局跳变（GUI 截图核对）
