# 开发前定稿（单一入口）

> 新开工程 **写代码前** 对齐本文 + 复制 `performance-budget.md` 到项目 `design/`，并在 Godot 工程中注册 `QualitySettings`。不要求你再「猜」档位数和路径规则。

## 必读顺序

1. **`performance-budget.md`** — 档位 ID、纹理目录名、贴图分辨率、显存与 `get_scalar` 键 **已定死**。
2. **`godot/quality-settings-spec.md`** — Autoload 顺序、API、`project.godot` 渲染方法切换规则。
3. **`asset-pipeline.md`** 第五节 — 分目录 / PCK / 导入变体如何接 `tier_*`。
4. **`compliance/content-guidelines.md`** — 含 NSFW 或 AI 生成资产时必读。
5. **`AIDOC/global-info/knowledge/GAME-DEV-INDEX.md`** — 知识库**收什么**（架构约束 / best practice，不抄官方手册）。

## 项目仓库最小产出（Phase 0 结束前）

| 路径 | 说明 |
|------|------|
| `AIDOC/game_doc/{游戏名}/README.md` | 入口：链到 `projects/{解决方案名}/{游戏名}/`、`design/`、`tracker.md`（见 `project-bootstrap.md`） |
| `AIDOC/game_doc/{游戏名}/design/performance-budget.md` | 从知识库 `performance-budget.md` 复制 |
| `AIDOC/game_doc/{游戏名}/design/architecture.md` | 增加一节「多平台画质」：引用 **固定档位 ID**，写明是否只做 PC |
| `projects/{解决方案名}/{游戏名}/autoload/quality_settings.gd` | 复制 `demo_game` 同名文件或按 spec 生成 |
| `projects/{解决方案名}/{游戏名}/assets/textures/tier_desktop/…` | 至少铺一条英雄测试路径，移动可先 symlink 或重复一份到低档 |
| `projects/{解决方案名}/{游戏名}/project.godot` | `[rendering] renderer/rendering_method` 与首档一致；Autoload 含 `QualitySettings` |

## 与 demo 工程对齐

工作台参考实现：**`projects/demo/demo_game/autoload/quality_settings.gd`** + `project.godot` 中已注册顺序。新游戏按 spec 复制即可。

## 粘贴到 `design/architecture.md` 的固定小节（原文可拷贝）

```markdown
## 多平台画质（工作台默认定稿）

- **唯一档位 ID**：`desktop_high`、`mobile_high` 两档（中端及以上；见 `performance-budget.md`）。
- **纹理物理路径**：`res://assets/textures/{tier_desktop|tier_mobile}/**`；逻辑路径仅写相对于 tier 的相对路径，经 `QualitySettings.resolve_texture()` 解析。
- **运行时单例**：`QualitySettings`（`autoload/quality_settings.gd`），Autoload 顺序早于 `DataManager` / `DlcManager`。
- **检测**：`OS.has_feature("mobile")` → `mobile_high`；PC 上模拟移动档：`--aidev-tier=mobile_high`。
- **本作裁剪**：若不做移动端，在澄清中声明并工程内可只保留 `tier_desktop`；`current_tier` 恒为 `desktop_high`。
```

---

## 仍属「项目专有」、无法全局写死的部分

- 各平台 **商店 ID、签名、版号**。
- **法律审查**结论（合规文档不可替代律师）。
- 本作 **独有** 玩法导致的额外 CPU 预算（在 `performance-budget.md` 第七节实测后改表）。
