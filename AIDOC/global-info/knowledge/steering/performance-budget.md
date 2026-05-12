# 性能与画质预算（工作台默认定稿）

> **本文件为默认定稿**：仅保留 **桌面中端及以上**、**移动中端及以上** 各 **一档**（不做低配机 / 老机型专档）。新游戏复制到 `AIDOC/projects/{游戏名}/design/performance-budget.md` 后只改实测表与机型名。若只做 PC，删除 **移动档**相关内容并在澄清中声明。
>
> **ID 勿改名**：`desktop_high`、`mobile_high`，与 `QualitySettings`、表、导出一致。

## 1. QualityProfile 档位表（固定 2 档）

| 档位 ID | 平台覆盖 | Godot 渲染方法（导出预设） | 目标 FPS | 参考分辨率 |
|---------|----------|----------------------------|----------|------------|
| `desktop_high` | 台式机 / 笔记本 **中端 GPU 及以上**（代表性配置见 **下文 1.1 节**，不承诺集显办公本） | `forward_plus`（Forward+） | 60 | 1920×1080 |
| `mobile_high` | **中端 SoC 及以上** Android / iOS（代表性样机见 **下文 1.1 节**，不承诺入门级整机） | `mobile`（**Forward Mobile**，默认首选）；老旧机型可单独导出预设切 `gl_compatibility`（Compatibility） | 60 | 1080p |

> **注意**：`mobile` 指 **Forward Mobile** 渲染器（保留 PBR、阴影、后效，仅简化阴影与屏幕空间效果），与 `gl_compatibility`（Compatibility）是 **不同的两个后端**。本工作台默认对移动档位用 Forward Mobile，仅在实测不达标或需要 Web/极老机型兜底时切 Compatibility，并在导出预设中显式标注。

**解析规则**：

- `OS.has_feature("mobile")` → `mobile_high`。
- 否则 → `desktop_high`。
- 开发覆盖：命令行 **`--aidev-tier=mobile_high`** 在 PC 上可模拟移动标量与 `tier_mobile` 路径（**不再提供** `desktop_low` / `mobile_low`）。

### 1.1 参考中端约定用机（2025–2026）

> 用于 **团队对齐、测试机选型、第七节实测表填设备名**；各项目可替换为代表性上市机型，**不**等同对外「最低配置」文案。

#### 桌面（`desktop_high`）

| 维度 | 约定 |
|------|------|
| GPU | 与 **GeForce RTX 3060 / RTX 4060** 或 **Radeon RX 6600 / RX 7600** 同级或更强的 **独显**；默认 **不** 以核显办公本为验收基线 |
| CPU | 近四年 **6 核及以上** 主流桌面 CPU，或同代游戏本 **H 系**（笔电可测，预算仍以 GPU 为先） |
| 内存 | **16 GB** |

#### 移动（`mobile_high`）

| 维度 | 约定 |
|------|------|
| SoC（Android） | **骁龙 7+ Gen 2 / 7 Gen 3 / 8s Gen 3**；**天玑 8200 / 8300（含 Ultra）** 等 **2023 年起主流中端芯**；**不含** 骁龙 4 系、天玑 700 级入门整机构成的默认锚点 |
| SoC（iOS） | **A15（iPhone 13 系）及以上** |
| 内存 | Android 以 **8 GB** 为默认锚点（**6 GB** 可进第七节单测，不作为全团队默认承诺） |
| 分辨率 | **1080p** 级画面负载；高刷 **非** 本档目标项 |

## 2. 纹理目录档（固定 2 文件夹）

| 档位 ID | 目录后缀（`assets/textures/` 下） |
|---------|-------------------------------------|
| `desktop_high` | `tier_desktop` |
| `mobile_high` | `tier_mobile` |

**逻辑路径**：相对路径如 `characters/hero/albedo.png` → `QualitySettings.resolve_texture()`。  
**Fallback**：当前档 → `tier_mobile` → `tier_desktop`（开发期可只铺一套）。

## 3. 贴图 Max Size（导入侧，3D PBR）

| 档位 ID | 角色主贴图 | 角色法线 | 环境大型物 | 道具 / 杂物 | UI |
|---------|-----------|----------|------------|-------------|-----|
| `desktop_high` | 2048 | 2048 | 4096 | 1024 | 512 |
| `mobile_high` | 1024 | 1024 | 2048 | 512 | 512 |

## 4. 场景显存与网格（软预算）

| 档位 ID | 场景纹理 VRAM 软上限 | 单场景三角面（万三角） | 同屏骨骼角色 | Draw Call 软上限 | 实时光源 |
|---------|----------------------|------------------------|--------------|------------------|----------|
| `desktop_high` | 900 MB | ≤ 120 | ≤ 20 | ≤ 200 | ≤ 8 |
| `mobile_high` | 350 MB | ≤ 60 | ≤ 8 | ≤ 100 | ≤ 4 |

## 5. get_scalar 键表（与代码一致，仅两列）

| 键（StringName） | 类型 | desktop_high | mobile_high |
|------------------|------|--------------|-------------|
| `shadow_distance` | float | 80 | 40 |
| `shadow_max_blur` | int | 5 | 2 |
| `ssao_enabled` | bool | true | false |
| `ssil_enabled` | bool | false | false |
| `glow_enabled` | bool | true | false |
| `msaa_3d` | int | 2（2×） | 0 |
| `max_skinned_actors` | int | 20 | 8 |
| `hscene_lod_skip` | int | 0 | 1 |
| `particle_multiplier` | float | 1.0 | 0.5 |

`hscene_lod_skip`：NSFW 场景简化等级（与 H-Scene 系统约定）。

## 6. 导出预设命名（建议）

| 预设 ID | 用途 |
|---------|------|
| `windows_desktop` | PC 正式包，`forward_plus`，含 `tier_desktop` |
| `android_mobile` / `ios_mobile` | 移动正式包，`mobile`（Forward Mobile），含 `tier_mobile`；可剔除 `tier_desktop`（分目录策略时） |
| `*_compat`（可选） | 仅当移动 Forward Mobile 实测不达标时单独建：渲染方法 `gl_compatibility` | 
| `*_dev` | 调试，与对应正式预设一致 |

**Steam Deck**：按 **桌面中端** 验收，使用 `windows_desktop` 同源资源与 `desktop_high`；若实机不稳再单开导出预设调分辨率，**不新增** `desktop_low` 档 ID。

## 7. 实测记录

| 日期 | 档位 | 设备 / 预设 | 场景 | 平均 FPS | 峰值纹理内存 | 备注 |
|------|------|-------------|------|----------|--------------|------|
| | | | | | | |

---

**关联**：`godot/quality-settings-spec.md`、`asset-pipeline.md` 第五节、`godot-engine.md` 第三节。
