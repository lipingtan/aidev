# NOVA ARCADE 开发工作流 Override

> 本文件是对通用规范 `steering/hybrid-project-workflow.md` 的项目专属 override。
> **优先级**：本文件 > `hybrid-project-workflow.md` > 上游通用规范。
> 只记录与通用规范不同或额外的内容，不重复通用内容。

---

## 1. 项目工程映射

| CR 类型 | 对应工程路径 |
|---|---|
| Go 后端 CR | `projects/nova_arcade/` 下 Go 服务工程（M3+ 启用） |
| Shell CR | `projects/nova_arcade/nova-arcade/`（待建主工程） |
| GameModule CR | `projects/nova_arcade/{游戏id}-godot/` 各游戏工程 |
| 游戏内部 CR | 同上 |

---

## 2. Shell CR 额外 Constraints

以下约束补充到所有 Shell CR 任务的 Constraints 中：

- **禁用 hdr_2d/glow**（mobile 渲染器实心矩形 bug 已踩坑；霓虹效果用分层半透明 ColorRect 模拟）
- **基准分辨率** `720×1560`，`stretch/mode=canvas_items`，`stretch/aspect=expand`
- **主题颜色禁止硬编码**，必须从 `ThemeTokens.color(key)` 读取
- **存档写入分级**：trial_used / orders / playtime / best → 强写（立即落盘）；search_history / daily / profile → 防抖写 500ms
- **游戏视口切换**：分辨率/方向切换必须在转场遮罩完全不透明后执行

---

## 3. GameModule CR 额外 Constraints

- `meta.json` 必须声明 `viewport_size` 和 `orientation`（横屏游戏不得省略）
- `ctx.trial_mode = true` 时游戏内禁止访问付费内容
- Arcade 游戏（`runtime: "arcade"`）必须额外声明 `core` 和 `roms[]` 字段

---

## 4. Go 后端 CR 额外 Constraints

- M3 前接口全部 Mock，用配置开关切换（`use_cloud_catalog` / `pay_channel` 等）
- API 路径前缀 `/v1/`，JWT Bearer 鉴权
- 评价：一人一游戏一条 active 记录（PUT 覆盖更新，非新增）
- 支付验签幂等：同 order_id 二次验证时校验 receipt_hash 一致性

---

## 5. Shell 集成冒烟（Shell CR 或跨层 CR 完成后必跑）

> 对应 `hybrid-project-workflow.md` §3.2 集成验证的项目具体实现。

| # | 检查项 |
|---|---|
| 1 | Shell 主工程编辑器中可打开无报错 |
| 2 | 底部四 Tab 可切换，无崩溃 |
| 3 | 详情页 CTA 状态机正确（free / trial / paid 三种模式） |
| 4 | Launcher 完整生命周期可走通（launch → boot → quit_requested → 结算卡） |
| 5 | 双主题切换（霓虹/青瓷）无布局跳变 |

---

## 6. 高频 Bug 排查优先级

1. **渲染**：hdr_2d/glow 导致实心矩形异常 → 检查 WorldEnvironment 配置
2. **存档路径**：游戏写 `user://` 根目录而非 `ctx.save_dir` → grep 所有 FileAccess.open
3. **GameModule 协议**：result 缺字段导致结算卡崩溃 → 检查 quit_requested 发射的 dict
4. **评价状态机**：同设备同游戏写入多条 active → 检查唯一约束和 PUT 覆盖逻辑
5. **跨层字段名**：客户端解析 cursor/status 字段名与服务端不匹配 → 检查 JSON tag

---

## 7. 里程碑 CR 路由速查

| 里程碑 | 典型 CR | CR 类型 |
|---|---|---|
| **M1** | Shell 工程初始化、Autoload 骨架 | Shell CR |
| **M1** | TETRA NOVA 接入 GameModule | GameModule CR |
| **M1** | 首页/分类/搜索/详情页 | Shell CR |
| **M1** | Launcher 完整生命周期 | Shell CR |
| **M2** | 魔塔/空战/其他游戏接入 | GameModule CR |
| **M2** | 成就 / 每日任务系统 | Shell CR |
| **M2** | Arcade PoC Android 真机验证 | 游戏内部 CR |
| **M3** | 云评价 API | Go 后端 CR |
| **M3** | 客户端接入云评价、目录增量下发 | 跨层集成 CR |
| **M4** | 支付 SDK + 订单验签 | Go 后端 CR + 跨层集成 CR |
| **M5** | PCK 打包流水线 + DLC 商店 | Shell CR + Go 后端 CR |

---

## 8. AI 交互行为规范（防失忆规则）

> 这些规则约束 AI 在本项目任何会话中的行为，违反即为错误。

### 8.1 needs_plan.md / requirements_plan.md 的 Answer 处理

- **读文件确认 Answer，不要反问用户**：用户在 requirements_plan.md 或 design_plan.md 中填写 Answer 后说"继续"，AI 必须先读取文件获取所有 Answer，直接进入下一阶段，**禁止重复向用户提问已经在文件中回答过的问题**。
- Answer 所在位置：`[Answer-N]` 行紧跟在对应 `[Question-N]` 的推荐答案之后，或在推荐文字之后另起一行。
- 若文件中 Answer 确实缺失（空白行或标记但无内容），才允许询问；但先尝试用推荐答案兜底，在询问前标注"已按推荐答案兜底"。

### 8.2 进度文件（progress.md）

- 每个 CR 完成后必须创建 `cr{N}-{name}/progress.md`，记录：完成状态、交付文件清单、测试结果、执行期关键决策、遗留事项。
- 下一个会话恢复时，先读 progress.md 了解上下文，不重复已完成工作。

### 8.3 测试验证规范

- 子代理汇报"全绿"后，主代理必须用实际 headless 命令实机验证，不得仅凭报告确认。
- 多个测试套件连续执行时，每次测试前清空 `dev/tmp_test/Godot/app_userdata/` 下项目 DB 目录，避免跨测试数据污染。
- test_rg5 双相调用方式：`--rg5-phase=write` 后 `--rg5-phase=verify`（同 `$env:APPDATA` 不清 DB 直接 verify）。

### 8.4 Godot 可执行文件路径（防版本混淆）

- headless：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64_console.exe`
- GUI：`C:\data\developer\devtool\godot\godot4.7\Godot_v4.7.2-stable_win64.exe`
- 执行前环境：`$env:APPDATA = "C:\data\developer\studio\games\aidev\dev\tmp_test"`；杀孤儿进程：`Get-Process -Name "Godot*" -ErrorAction SilentlyContinue | Stop-Process -Force`
