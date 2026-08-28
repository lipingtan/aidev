# 需求计划：CR-6 M2 小游戏批量接入（GameModule CR）

> 依据 `nova-arcade-design.md` §10 里程碑 M2「+4~6 个小游戏」+ `dev-workflow-override.md` §7 路由（M2 GameModule CR）。
> 类型：GameModule CR（工程 `projects/nova_arcade/nova-arcade/`，激活层 hybrid-project-workflow §3.3）。
> 前置状态：M1 全部闭环（shell-bootstrap + CR-2~CR-5 + Fix-1）；Registry/Launcher/GameModule 协议已就位，仅 tetra_nova 一款游戏。
> 成就墙/本地推荐画像按用户决定后置，不在本 CR。

---

## 需求理解

**目标**：把盒子从「1 款游戏」扩充到「5~7 款可玩内容」，验证多游戏批量接入管线（meta.json → Registry 扫描 → Launcher 生命周期 → 结算卡），为 M2 收尾和 M3 云端化打基础。

**范围**：
1. 接入现有 Godot demo `magic-tower-godot`（魔塔，完整格子制玩法 + 触屏十字盘）→ `games/magic_tower/`
2. 新建 2~3 款轻量小游戏（候选：2048 / 贪吃蛇 / 打砖块 / 扫雷），GDScript 实现，走同一 GameModule 协议
3. 每款游戏：meta.json + module.tscn + module_adapter.gd + src/ + icon.png，headless 测试套件
4. `data/editorial.json` 补充 banner/featured 条目（分类页/首页多游戏内容）

**不在范围**：
- airwar（HTML5 纵版打飞机）——依赖 HtmlRunner（WebView），归独立 CR 或 M2 末
- 成就墙、本地推荐画像——后置
- Arcade PoC Android 真机验证——M2 独立 CR
- Shell 代码改动（Registry/Launcher/Nav 已支持多游戏，仅数据层扩展）

## 假设列表

- [假设-1] magic-tower-godot 为 Godot 4.5 工程，可被 4.7.2 主工程直接导入（4.x 向后兼容）；若导入有 API 不兼容，按 `kb/godot-4.7-api-facts.md` 逐项修复。
- [假设-2] 新游戏均为竖屏 `720×1560` 基准分辨率内运行，与 Shell stretch/expand 模式兼容（横屏游戏归 Arcade PoC 范畴）。
- [假设-3] tetra_nova 的目录结构（meta.json + module.tscn + module_adapter.gd + src/）是接入标准模板，新游戏全部照此结构。
- [假设-4] 每款新游戏规模控制在单场景 ≤300 行核心逻辑 + 少量 UI，符合「1~2 屏完整可玩循环」定位（非完整 ARPG）。

## 澄清问题

- [Question-1] 本 CR 接入哪些游戏、几款？
  - 业界最佳实践：批量接入先「1 款现有资产走通管线」再「批量新建小体量内容」，降低管线风险与返工面；每款独立可验收。
  - 推荐答案及理由：**3 款 = magic_tower（移植现有 demo）+ 2048 + 贪吃蛇（新建）**。理由：魔塔验证「移植现有工程」路径（src/ 拷贝 + adapter），2048/贪吃蛇验证「从零新建轻量游戏」路径；三款覆盖两种接入模式且工作量可控（魔塔 1 天 + 两款各 0.5~1 天）。打砖块/扫雷 顺延到下一 CR 或并入 M2 收尾。
  [Answer-1] 采纳推荐：magic_tower + 2048 + 贪吃蛇，共 3 款（2026-08-28 确认）
- [Question-2] 新游戏的 price_model 与试玩策略？
  - 业界最佳实践：内容扩充期以免费游戏堆量（提高盒子留存），付费点集中在少数精品（M4 商业化再上 trial/paid）。
  - 推荐答案及理由：**三款全部 `price_model: "free"`**。理由：M2 目标是「完整可玩、可分发的免费盒子」（design §10 注）；tetra_nova 已承担 trial 付费点示范，新游戏不再设试玩门槛。
  [Answer-2] 采纳推荐：三款全部 free（2026-08-28 确认）
- [Question-3] magic_tower 移植方式：整工程拷贝进 `games/magic_tower/src/`（tetra 模式），还是保持独立工程 + PCK 引用？
  - 业界最佳实践：内置游戏随主包分发时，源码并入主工程最简单（Registry 直接 load res:// scene）；PCK/DLC 路线留给 M5。
  - 推荐答案及理由：**整工程拷贝进 `games/magic_tower/src/`**（tetra 模式）。理由：M2 无 DLC 管线，Registry.reload() 已按 `res://games/*/meta.json` 扫描，module.tscn 指向 src/ 内场景即可；独立工程保留为参考副本不删。
  [Answer-3] 采纳推荐：整工程拷贝进 games/magic_tower/src/，独立工程保留为参考副本（2026-08-28 确认）
- [Question-4] 新游戏存档与结算的最小契约？
  - 业界最佳实践：轻量游戏也走统一 ctx 注入（save_dir/trial_mode/best）+ quit_requested({score, playtime, achievements})，保证结算卡/继续游戏/排行榜数据同构。
  - 推荐答案及理由：**完整走 GameModule 协议最小集**：boot(ctx) 读 save_dir 存 best score；quit_requested 必含 score/playtime(秒)/achievements[]；pause/resume 转发到游戏自身暂停（无暂停机制的记 push_warning）。理由：Launcher/结算卡零改动，M3 云排行榜直接复用。
  [Answer-4] 采纳推荐：完整走 GameModule 协议最小集（2026-08-28 确认）
- [Question-5] 每款游戏的验收标准怎么定？
  - 业界最佳实践：headless 测试套件（逻辑断言 + 生命周期走查）+ GUI 截图核对（双主题下无布局跳变），与 CR-3/CR-4/CR-5 验收口径一致。
  - 推荐答案及理由：**每款一个 `tools/test_{game}.tscn`**（≥10 断言：boot/pause/resume/quit/result 字段/save_dir 写入/best 读取）+ GUI 截图（neon/elegant × 游戏画面）+ Shell 集成冒烟 §5 全过 + 全量回归。理由：与既有 CR 验收口径一致，防「只跑通不验证」。
  [Answer-5] 采纳推荐：与 CR-3/CR-4/CR-5 验收口径一致（2026-08-28 确认）

## 非功能需求建议

- 性能：新游戏逻辑帧率无压力（≤300 行级），重点防 UI 层节点泄漏（on_exit queue_free）
- 一致性：主题色全走 ThemeTokens，禁硬编码；存档写入分级遵循 override §2（best → 强写）
- 目录：每款游戏 icon.png 需生成（可复用 tetra 风格霓虹图标）

## 影响范围预判

- 涉及模块：`nova-arcade/games/`（+3 目录）、`nova-arcade/data/editorial.json`、`nova-arcade/tools/`（+3 测试套件）
- 涉及文件（预估）：~15 新增文件 + editorial.json 修改；Shell core/services 零改动
- 可能的副作用：Registry 扫描耗时增加（可忽略，≤10 meta）；首页「继续游戏」区块出现多游戏条目（预期行为）
