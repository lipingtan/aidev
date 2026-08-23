# 设计计划：CR-1 shell-bootstrap（Shell 工程初始化 + Autoload 骨架）

> CR 类型：**Shell CR**（override §7 M1 路由）。上游设计文档：`nova-arcade-client-design.md`（明确设计，按 hybrid §2 简化规则跳过需求计划+需求文档）。
> 叠加约束：`dev-workflow-override.md` §2 Shell Constraints + `hybrid-project-workflow.md` §3.2/§4.1。

## 设计方向

新建 Godot 4.5 工程 `projects/nova_arcade/nova-arcade/`，按 client-design §1/§10 搭建盒子骨架：

- **project.godot**：720×1560 竖屏、`stretch/mode=canvas_items` + `aspect=expand`（override §2）；渲染器按 [Question-2] 定
- **Autoload**：client-design 五单例（EventBus / DB / Registry / Nav / Sound）+ 规范强制的 QualitySettings（pre-development-defaults），注册顺序 EventBus → QualitySettings → DB → Registry → Nav → Sound；遵守 godot-engine §六（Autoload 注册名 ≠ class_name）
- **主场景树**：`Main.tscn` = BgLayer + App(PageStack/TabBar/OverlayLayer/ToastLayer) + GameHost（client-design §1.1）
- **主题系统**：ThemeTokens 双主题（霓虹·星穹 / 青瓷·素笺），颜色一律 `ThemeTokens.color(key)`，禁硬编码（override §2）
- **页面骨架**：四 Tab 根页（Home/Category/Search/Library）空态占位 + Nav push/pop 转场 + Tab 切换清栈
- 目录结构按 client-design §10（shell/ services/ core/ data/），偏离 godot-bootstrap 通用目录模板，以项目设计文档为准（design.md 中记录该偏离及理由）

## 技术选型

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| 渲染器 `mobile`（Forward Mobile） | 与 Android 目标一致，移动端 PBR/后效可用 | 桌面调试观感与真机略有差异 | ✓（若 [Question-2] 选 Android 先行） |
| 渲染器 `forward_plus` | 桌面开发观感好 | 与移动目标不一致，后期需切 | ✗ |
| DB 持久化：JSON + DB 单例（强写/防抖写分级） | client-design §2.1 已定稿，崩溃恢复可靠 | 无并发写保护需求（单线程主循环） | ✓ |
| 主题切换：双 .tres Theme 资源运行时切换 | Godot 原生支持，记住选择存 profile | — | ✓ |
| CoreManager（Arcade 核心包）实现口径 | 见 [Question-3] | — | 待定 |

## 澄清问题

- [Question-1] 工程位置确认：`projects/nova_arcade/nova-arcade/`（KB 标注的计划位置，解决方案 nova_arcade 下新建）？
  [Answer-1]
  确认
- [Question-2] M1 平台策略：
  - A. **Android 先行**（推荐）：渲染器 `mobile`，桌面仅开发调试用，M1 不做双平台 smoke；M2 真机验证时补 Android 出包
  - B. 双平台：桌面 `forward_plus` + 移动 `mobile` 双预设，M1 即做桌面 smoke
  [Answer-2]
  A
- [Question-3] CoreManager（Arcade 核心包管理）实现口径：
  - A. **留接口桩**（推荐）：`is_installed/check_installed/download/remove` 接口 + Mock 实现返回"未装"，真实下载/校验等 mame-godot-plugin v2 验证后填
  - B. 按旧 myosd 口径完整实现（与 v2 冲突，后期要返工）
  [Answer-3]
  A
- [Question-4] CR-1 任务范围：
  - A. **含 DB/Registry 骨架**（推荐）：工程初始化 + Autoload×6 + 主题 Token + 主场景树 + 四 Tab 空页骨架 + DB/Registry/GameMeta 定义与内置 tetra_nova meta（页面数据源依赖它，工作量小，一次做完避免 CR-3 返工）
  - B. 仅工程初始化 + Autoload + 主题 + 场景树 + 空页；DB/Registry 拆入 CR-3
  [Answer-4]
A
## 风险点

- [Risk-1] 720×1560 + `aspect=expand` 在极端宽高比设备（19.5:9 / 20:9）可能布局溢出 → 冒烟时至少用两种视口比例验证
- [Risk-2] Autoload 注册名与 class_name 冲突导致实例方法调用报错（godot-engine §六）→ POST-CHECK 专项检查
- [Risk-3] QualitySettings 依赖 demo_game 参考实现与 `tier_*` 纹理目录；Shell UI 以矢量/ColorRect 为主、几乎无纹理 → Autoload 必须注册（规范强制），纹理分档目录建空占位即可，design.md 记录裁剪理由
- [Risk-4] 集成冒烟 5 项中 CTA 状态机 / Launcher 生命周期属 CR-3/CR-4 范围 → CR-1 只跑第 1（工程打开无报错）、2（Tab 切换）、5（双主题无跳变）三项，其余在对应 CR 完成
