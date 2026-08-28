# 需求：CR-6 M2 小游戏批量接入（魔塔 + 2048 + 贪吃蛇）

## 背景

M1 已闭环（Shell 骨架 + TETRA NOVA 接入 + 本地评价/继续游戏），盒子仅 1 款游戏。M2 目标「内容扩充」要求 +4~6 款小游戏，验证多游戏批量接入管线（meta.json → Registry 扫描 → Launcher 生命周期 → 结算卡），达成「完整可玩、可分发的免费盒子」。

本 CR 接入 3 款：magic_tower（移植现有 Godot demo）、2048、贪吃蛇（新建）。成就墙/本地推荐画像/airwar(HTML)/Arcade PoC 不在本 CR。

## 用户故事

- 作为玩家，我希望在盒子内直接玩魔塔/2048/贪吃蛇，以便 不用装多个 App 就能玩多种小游戏
- 作为玩家，我希望每款游戏都有统一结算卡（分数/时长）和「继续游戏」入口，以便 体验与 TETRA NOVA 一致
- 作为玩家，我希望我的最佳成绩按游戏分别保存，以便 下次打开能看到并挑战

## 功能需求

### FR-1: magic_tower 接入（移植现有 demo）
**描述：** 将 `magic-tower-godot` 整工程拷贝至 `games/magic_tower/src/`，新建 meta.json + module.tscn + module_adapter.gd（tetra 模式），原独立工程保留为参考副本。
**验收标准：**
- WHEN Registry.reload() 扫描 THEN 系统 SHALL 加载 magic_tower meta.json 且 lookup("magic_tower") 返回完整 GameMeta
- WHEN Launcher launch("magic_tower") THEN 系统 SHALL 走通 boot → 游戏画面 → quit_requested 完整生命周期
- WHEN 游戏内存档写入 THEN 系统 SHALL 只写 ctx.save_dir 下路径（grep 无硬编码 user:// 绝对路径）
- WHEN 魔塔逻辑自检（MT_SELFTEST）THEN 系统 SHALL 全部断言通过（移植后逻辑零改动）

### FR-2: 2048 新建游戏
**描述：** GDScript 实现 2048（4×4 网格、滑动合并、分数累计、游戏结束判定），竖屏触屏滑动 + 键盘方向键双输入，走 GameModule 协议。
**验收标准：**
- WHEN 玩家滑动/按键移动网格 THEN 系统 SHALL 正确执行合并规则（同值合并翻倍、每格至多合并一次）
- WHEN 分数累计 THEN 系统 SHALL 在 quit_requested.result.score 中返回本局最高分
- WHEN 网格无合法移动 THEN 系统 SHALL 触发游戏结束并显示结算入口
- WHEN 玩家退出回盒 THEN 系统 SHALL 发射 quit_requested({score, playtime, achievements})，best score 写入 ctx.save_dir

### FR-3: 贪吃蛇新建游戏
**描述：** GDScript 实现贪吃蛇（网格移动、食物生长、撞墙/自撞死亡、速度递增），竖屏触屏方向按钮 + 键盘双输入，走 GameModule 协议。
**验收标准：**
- WHEN 蛇吃到食物 THEN 系统 SHALL 增长 1 节并提升移动速度（按设计曲线）
- WHEN 蛇撞墙或自身 THEN 系统 SHALL 立即结束本局并显示结算入口
- WHEN 玩家退出回盒 THEN 系统 SHALL 发射 quit_requested({score, playtime, achievements})，best score 写入 ctx.save_dir

### FR-4: 统一结算与继续游戏
**描述：** 三款新游戏复用现有结算卡/继续游戏/推荐区块，Shell core/services 零改动。
**验收标准：**
- WHEN 任一新游戏退出回盒 THEN 系统 SHALL 弹出结算卡（score/playtime），与 tetra_nova 同构
- WHEN 玩家从首页「继续游戏」进入任一已玩新游戏 THEN 系统 SHALL 恢复上次进度（best/存档）
- WHEN 分类页/搜索页查询 THEN 系统 SHALL 能命中三款新游戏（meta.json title/aliases/pinyin 完整）

### FR-5: editorial 数据补充
**描述：** `data/editorial.json` 补充 banner[]/featured 条目覆盖三款新游戏，首页/分类页内容不再只有 tetra_nova。
**验收标准：**
- WHEN 首页加载 THEN 系统 SHALL 在 Banner/推荐区块中出现新游戏条目（至少 1 条 banner）
- WHEN Registry.reload() THEN 系统 SHALL 解析 editorial.json 无报错

## 非功能需求

- 性能：新游戏逻辑 ≤300 行级，UI 层 on_exit 无节点泄漏；全量回归 14+3 场景独立 APPDATA 全绿
- 一致性：主题色全走 ThemeTokens（禁硬编码）；best score 强写立即落盘（override §2 分级）
- 兼容性：magic_tower 从 Godot 4.5 导入 4.7.2，API 不兼容项按 kb/godot-4.7-api-facts.md 逐项修复
- 目录规范：每款游戏 `games/{id}/` = meta.json + module.tscn + module_adapter.gd + src/ + icon.png
