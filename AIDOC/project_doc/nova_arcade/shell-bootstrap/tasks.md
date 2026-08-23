# 任务：CR-1 shell-bootstrap

> 依据 `design.md`（已确认）。格式：三要素（task-representation.md）；Shell CR 通用 Constraints（hybrid §3.2）+ override §2 逐任务隐含生效，不重复列出。
> 状态标记：⬜ 未开始 / 🔨 进行中 / ✅ 完成；验收 📋 待验 / 🎯 通过 / 🔄 整改中

## 进度摘要

- 总任务数：10　已完成：9　进行中：0　待开始：1（T10）
- 当前阶段：批次4（T5 ✅ / T4 ✅ / T3 ✅ / T6 ✅ / T7 ✅；子代理两度中断，改主会话直接执行）

## 依赖关系（按接口调用关系重推，Review 第3轮；第4轮 A-5 简化图）

三层结构（以下 bullet 为信息源）：

1. **串行链**：T1 → T2 → T3
2. **并行组**：{T5, T6, T7}（T3 后互不阻塞）；独立于主线：{T4}（仅依赖 T2）、{T9}（仅依赖 T1，任意时点）
3. **汇聚**：T8 = T5+T6+T7 → T10（全部完成后）

- **关键路径**：T1 → T2 → T3 → {T5/T6/T7 并行} → T8 → T10
- 去多余边：T3→T4（Registry 扫 meta.json 文件，不碰 DB）、T4→T8（空态骨架不读 Registry，数据源 CR-3）、T4→T9（Mock 桩仅依赖 T1）
- 补缺失边：T3→T6（Sound 读 profile 开关）、T6→T8（🎨 按钮播 Sound.toggle）

---

### Task 1: 工程初始化 ✅ 🎯

**复杂度**: 中

**Scope:**
- 创建 `projects/nova_arcade/nova-arcade/`：project.godot、目录骨架（shell/{theme,components,pages,layers} core/ services/ games/tetra_nova/ data/ tools/）、.gitignore（.godot/、*.import）
- project.godot：name/description/main_scene、720×1560 + canvas_items + expand、renderer=mobile、autoload×6 注册
- 6 个 autoload 脚本建**最小桩**（`extends Node` + 文件头注释，空实现；Review B-3：注册路径缺失会导致编辑器报 missing script，T2~T8 各任务填实）
- 占位 `shell/main.tscn`（空场景，T8 换完整场景树）+ `tools/smoke_boot.gd` 最小启动脚本（Review B-4）
- 不触碰：现有 nova_arcade 解决方案下其他工程

**Constraints:**
- 配置值与 design.md §2 完全一致；渲染器 `mobile`（Q2=A）
- autoload 顺序 EventBus → QualitySettings → DB → Registry → Nav → Sound

**Acceptance:**
- [ ] 工程在 Godot 4.5 编辑器中打开无报错、无缺失资源提示
- [ ] headless 冒烟门槛：`Godot --headless --path projects/nova_arcade/nova-arcade -s res://tools/smoke_boot.gd`（最小启动脚本，跑完 quit(0)）退出码 0 且输出无脚本错误（Review A-3；后续任务回归复用）
- [ ] project.godot 各项配置与 design.md §2 逐项一致
- [ ] 目录骨架完整，.gitignore 生效（.godot/ 不入库）

---

### Task 2: EventBus + QualitySettings ✅ 🎯

**复杂度**: 中

**Scope:**
- 创建 `core/event_bus.gd`（12 信号，design.md §3.1）、`core/quality_settings.gd`（复制 `projects/demo/demo_game/autoload/quality_settings.gd` 裁剪——实查路径，godot-engine.md §3.1 中 `projects/demo_game/` 为旧路径；若文件不可用则按 quality-settings-spec.md 生成，design.md §3.2）
- 创建 `assets/textures/tier_desktop|tier_mobile/` 空目录占位
- 不触碰：demo_game 原工程

**Constraints:**
- autoload 脚本不声明 class_name（godot-engine §六）；类型标注完整、中文注释、文件头注释
- QualitySettings 保留档位检测/resolve_texture/get_scalar，裁剪纹理分档业务逻辑并注释理由

**Acceptance:**
- [ ] 两脚本语法通过，信号表与 design.md §3.1 一致
- [ ] 桌面默认判 desktop_high；`--aidev-tier=mobile_high` 可覆盖且 resolve_texture 返回 tier_mobile 路径（Review A-4）
- [ ] 单文件 ≤200 行

---

### Task 3: DB 单例 ✅ 🎯

**复杂度**: 高

**Scope:**
- 创建 `core/db.gd`：JSON 持久化（user://db/）、强写/防抖写分级、design.md §3.3 全部接口
- 不触碰：其他单例的读写路径

**Constraints:**
- 强写范围（trial_used/orders/playtime/best/finish_count）立即落盘；防抖写（search_history/daily/profile）500ms 合并，WILL_EXIT 强制 flush
- `save_profile(patch, force := false)`：默认防抖，关键项（如主题切换）force=true 立即落盘（Review B-1，与 design §3.3/§5 一致）
- JSON 损坏兜底：坏文件改名 `.corrupt` + 默认值重建 + 日志告警（design.md §3.3）
- 所有持久化经 DB 出口，页面/服务不得直接 FileAccess 写 user://（高频 Bug #2）
- 单文件 ≤200 行，超出则拆 db_io.gd
- **偏差（实测）**：headless 下 `NOTIFICATION_WM_CLOSE_REQUEST` 不触发，退出 flush 补 `_exit_tree()` 钩子双保险（autoload 移出场景树时落盘）

**Acceptance:**
- [x] 强写数据落盘后杀进程重启可恢复（RG-5）
- [x] 防抖写在 500ms 内多次调用只产生一次落盘；WILL_EXIT 时缓冲全部 flush
- [x] 人为损坏某 JSON 文件后启动：不崩溃，坏文件被备份，数据以默认值重建
- [x] 接口与 design.md §3.3 签名一致，类型标注完整

---

### Task 4: Registry + GameMeta + tetra_nova meta ✅ 🎯

**复杂度**: 高

**Scope:**
- 创建 `core/game_meta.gd`（GameMeta 类，字段按 CD §3.4）、`core/registry.gd`（design.md §3.4 接口）、`games/tetra_nova/meta.json`（CD §3.4 字段，runtime="pck"、trial plays:3）、`data/editorial.json` 空模板、icon/scene 占位资源
- 不触碰：游戏本体代码（CR-2）

**Constraints:**
- meta 字段按 CD §3.4；viewport_size/orientation 省略继承 Shell
- 启动扫描 `games/*/meta.json`，加载失败仅日志告警不崩溃
- **偏差（实测）**：查询接口 `get(gid)` 因与 Node 基类 Object.get 签名冲突改名为 `lookup(gid)`；4.5 解析器不支持单行 lambda / `Type?` 可空注解 / natural_compare，已按替代写法实现并注释

**Acceptance:**
- [x] Registry.reload() 后 lookup("tetra_nova") 返回完整 meta
- [x] query({category:"puzzle"}) 命中 tetra_nova；坏 meta.json 不阻断启动
- [x] 单文件 ≤200 行

---

### Task 5: Nav + Page 基类 ✅ 🎯

**复杂度**: 高

**Scope:**
- 创建 `core/nav.gd`（push/pop/pop_to_root/switch_tab/current，design.md §3.5）、`shell/pages/page.gd`（四接口，CD §3.2）
- 不触碰：具体页面实现（Task 8）

**Constraints:**
- 页面间禁止互相引用；跳转只走 Nav、通知只走 EventBus
- push/pop 转场 220ms；switch_tab 清栈回根页并发 tab_changed；返回键拦截 Dialog → on_back() → 默认 pop
- Page 基类 `on_enter` 原设计 `@abstract` 编译期强制（Review P-2）；**偏差（实测）**：Godot 4.5 无法解析多方法类中的无函数体 @abstract 方法（报 Expected indented block），改为类级 @abstract + on_enter 默认实现运行时 push_warning 提示

**Acceptance:**
- [ ] push/pop 动画时长与栈行为正确（pop_to_root 清空）
- [ ] switch_tab 后 PageStack 仅剩目标根页，tab_changed 参数正确（T5 时点用临时占位场景验证栈行为；最终形态 T10 复验，Review P-4）
- [x] Page 基类四接口齐全；未覆写 on_enter 运行时 push_warning 提示（@abstract 引擎限制，见 Constraints 偏差）

---

### Task 6: Sound 单例 ✅ 🎯

**复杂度**: 低

**Scope:**
- 创建 `core/sound.gd`：click/toggle/success/error/coin/haptic + profile 开关读取；占位短音资源

**Acceptance:**
- [x] 各音效接口可调用无报错；profile 关闭后静默
- [x] 震动调用受平台能力保护（桌面不报错）
- **偏差（实测）**：Godot 4.5 无内置跨平台 haptic API（DisplayServer 无 haptic_pulse），haptic() 仅做 mobile 能力保护 + `_haptic_impl` 实现位，真机震动 M2 接插件；占位音效由 tools/gen_sfx.gd 生成

---

### Task 7: 主题系统 ✅ 🎯

**复杂度**: 高

**Scope:**
- 创建 `shell/theme/theme_tokens.gd`（TOKENS 双套 + color()/icon_grad()）、`theme_neon.tres`、`theme_elegant.tres`（StyleBoxFlat 样式，CD §8.5 映射）
- BgLayer 主题响应（nebula/GridLines/DotGrid 切换 + 400ms Tween）
- 不触碰：页面业务逻辑

**Constraints:**
- 主题色禁止硬编码，非 Theme 控件一律 `ThemeTokens.color(key)`（override §2）
- 禁用 hdr_2d/glow；霓虹用分层半透明 ColorRect/StyleBoxFlat shadow 模拟
- apply_theme 流程：Token → root.theme → profile 强写 → settings_changed + theme_changed

**Acceptance:**
- [x] 信号流正确：apply_theme → Token 切换 → root.theme 换资源 → profile force 写 → settings_changed+theme_changed；BgLayer 仅响应 theme_changed（400ms Tween）
- [x] grep 验证页面/BgLayer 无硬编码 Color/十六进制色值（ThemeTokens 内部除外；bglayer.tscn 初始色为占位，_ready 即被 token 覆盖）
- [x] 工程未启用 hdr_2d/glow（project.godot 仅 renderer=mobile；quality_settings 的 glow_enabled 为 demo 拷贝查表项，CR-1 未应用）
- [ ] 视觉项（双主题切换无布局跳变 / 重启恢复上次选择，RG-4）移交 T10 在完整场景树后执行（Review P-5）
- **偏差（实测）**：4.5 Gradient 无 set_point_color/point_count/points 属性，用 `offsets + colors` 直接赋值构造渐变

---

### Task 8: 四 Tab 根页骨架 + Main.tscn 组装 ✅ 🎯

**复杂度**: 中

**Scope:**
- 创建 home/category/search/library 四根页（标题 + EmptyState 占位）、`shell/components/tab_bar.gd`、`shell/main.tscn`（BgLayer/App/TabBar/OverlayLayer/ToastLayer/GameHost，design.md §4）
- 首页右上角 ThemeToggleButton（🎨 占位，CD §8.4；Review 高优项：主题切换 UI 入口）
- ToastLayer 基础实现（300ms 滑入 / 2200ms 展示）
- 不触碰：详情页/列表页等 CR-3 内容

**Constraints:**
- 页面仅展示与转发，不含业务规则；on_enter 打印参数供冒烟
- 布局按 720×1560 设计，兼容两种视口比例（Risk-1）

**Acceptance:**
- [x] 启动进入 Home，四 Tab 可切换无崩溃（RG-1/RG-2）——test_main 13/13
- [ ] 两种视口比例（约 9:19.5 与 9:20）下布局无溢出/错位 → 视觉项移交 T10（Risk-1，headless 不可验）
- [x] Toast 触发/消失时序正确——test_main 13/13；根因：`tween_interval` 之后重复 `set_parallel(true)` 压平停留，已改为 interval 后不再 toggle
- [x] 点击 🎨 完成主题切换、播放音效、有按压缩放反馈（RG-4）

---

### Task 9: CoreManager 桩 ✅ 🎯

**复杂度**: 低

**Scope:**
- 创建 `services/core_manager.gd`：is_installed/check_installed/download/remove + Mock 恒"未安装"（Q3=A）

**Acceptance:**
- [ ] 四接口可调用，Mock 返回未安装并记录日志（RG-6）
- [ ] 文件头注释标注"真实实现待 mame-godot-plugin v2 验证后填"

---

### Task 10: 集成冒烟 + 回归验证 ⬜ 📋

**复杂度**: 中

**Scope:**
- 无代码变更（如有缺陷回对应任务整改）：跑 override §5 冒烟子集 + RG-1~RG-6 + 两种视口比例；承接 T5 栈行为复验与 T7 视觉验收（P-4/P-5）
- 视觉验证用 GUI 版 Godot 截图 + read_image（kb/visual-verify.md），.tres 样式 headless 无法验证
- 产出 `AIDOC/project_doc/nova_arcade/shell-bootstrap/smoke_report.md`

**Acceptance:**
- [ ] 冒烟 3 项通过：工程打开无报错 / Tab 切换无崩溃 / 双主题无跳变（附截图）
- [ ] RG-1~RG-6 逐条运行验证通过（含 RG-5 杀进程重启）
- [ ] smoke_report.md 记录每项结果与视口比例实测截图/日志
