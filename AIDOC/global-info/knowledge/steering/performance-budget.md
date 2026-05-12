# 性能与画质预算（工作台默认定稿）

> **本文件为默认定稿**：仅保留 **桌面中端及以上**、**移动中端及以上** 各 **一档**（不做低配机 / 老机型专档）。新游戏复制到 `AIDOC/projects/{游戏名}/design/performance-budget.md` 后只改实测表与机型名。若只做 PC，删除 **移动档**相关内容并在澄清中声明。
>
> **ID 勿改名**：`desktop_high`、`mobile_high`，与 `QualitySettings`、表、导出一致。

## 1. QualityProfile 档位表（固定 2 档）

| 档位 ID | 平台覆盖 | Godot 渲染方法（导出预设） | 目标 FPS | 参考分辨率 |
|---------|----------|----------------------------|----------|------------|
| `desktop_high` | 台式机 / 笔记本 **中端 GPU 及以上**（不承诺集显办公本） | `forward_plus` | 60 | 1920×1080 |
| `mobile_high` | **中端 SoC 及以上** Android / iOS（不承诺入门级整机） | `mobile`（Compatibility），Forward+ 仅当全量实测通过 | 60 | 1080p |

**解析规则**：

- `OS.has_feature("mobile")` → `mobile_high`。
- 否则 → `desktop_high`。
- 开发覆盖：命令行 **`--aidev-tier=mobile_high`** 在 PC 上可模拟移动标量与 `tier_mobile` 路径（**不再提供** `desktop_low` / `mobile_low`）。

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
| `android_mobile` / `ios_mobile` | 移动正式包，`mobile`，含 `tier_mobile`；可剔除 `tier_desktop`（分目录策略时） |
| `*_dev` | 调试，与对应正式预设一致 |

**Steam Deck**：按 **桌面中端** 验收，使用 `windows_desktop` 同源资源与 `desktop_high`；若实机不稳再单开导出预设调分辨率，**不新增** `desktop_low` 档 ID。

## 7. 实测记录

| 日期 | 档位 | 设备 / 预设 | 场景 | 平均 FPS | 峰值纹理内存 | 备注 |
|------|------|-------------|------|----------|--------------|------|
| | | | | | | |

---

**关联**：`godot/quality-settings-spec.md`、`asset-pipeline.md` 第五节、`godot-engine.md` 第三节。
