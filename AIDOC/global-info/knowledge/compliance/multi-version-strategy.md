# 多版本/补丁发布策略

> 描述 All-Ages vs R-18 双版本架构，以及与 DLC 动态挂接系统的对接方式。

---

## 一、版本架构设计原则

利用现有 **DLC 动态挂接系统**（Phase 4 已预留挂接点），将成人内容封装为可选 DLC/补丁：

```
主体工程（Base Game）
  ├── 核心玩法（All-Ages 安全）
  ├── 角色系统（无露骨内容）
  ├── 关系数值系统（骨架，内容由 DLC 填充）
  └── NSFW DLC 挂接点预留

成人内容 DLC（Adult Patch）
  ├── H-Scene 内容包
  ├── 无码/有码版本切换
  ├── 额外 CG Gallery 内容
  └── 成人语音包
```

**优点**：
- 主体游戏可在 Steam 正常上架（非 Adult Only），扩大曝光
- 补丁可在 itch.io/自建网站分发，绕过部分平台限制
- 满足"内容本体可独立游玩"的平台审核要求

## 二、DLC 挂接点规范

在代码中预留统一的成人内容挂接点，通过 `DLCManager` 单例动态加载：

```gdscript
# 示例：H-Scene 触发点
func try_trigger_h_scene(scene_id: String, characters: Array) -> void:
    if DLCManager.is_adult_dlc_active():
        HSceneManager.trigger(scene_id, characters)
    else:
        # 全年龄版：播放替代动画（拥抱、淡出等）
        FallbackAnimationPlayer.play_fade_alternative(scene_id)
```

**规范要求**：
- 所有 NSFW 内容触发点必须有全年龄版 fallback
- fallback 内容必须在 NSFW DLC 未加载时正常运行
- DLC 状态检查统一走 `DLCManager.is_adult_dlc_active()`，不散落在各系统

## 三、资产分离规范

| 资产类型 | 存放位置 | 说明 |
|----------|----------|------|
| 全年龄角色立绘 | `assets/characters/{name}/base/` | Base Game 内置 |
| 成人角色立绘 | `assets/characters/{name}/adult/` | DLC 包内 |
| H-Scene CG | `assets/cg/adult/{scene_id}/` | DLC 包内 |
| 成人语音 | `assets/audio/voice/adult/` | DLC 包内 |
| 马赛克版贴图 | `assets/characters/{name}/censored/` | 日本版 DLC 内 |

## 四、打包配置矩阵

**与多平台画质档**：移动端基座包须使用 `tier_mobile`（或项目定义的移动档）纹理与资源集；桌面高清 / Adult DLC 高清包须与 `asset-pipeline.md` 第五节、**`QualitySettings`** 的解析规则一致，避免移动端安装包误打入 `tier_desktop_only` 资源。

| 版本 | 包含内容 | 目标平台 | 导出配置 |
|------|----------|----------|----------|
| Steam 全年龄版 | Base Game | Steam（普通标签） | `export_steam_all_ages.cfg` |
| Steam 成人版 | Base + Adult DLC | Steam（Adult Only 标签） | `export_steam_adult.cfg` |
| itch.io 无码版 | Base + Adult DLC | itch.io | `export_itch_uncensored.cfg` |
| DLsite/FANZA 版 | Base + Adult DLC（有码） | DLsite/FANZA | `export_dlsite_censored.cfg` |
| Patreon 早期版 | Base + Adult DLC | 自建/Dropbox | `export_patreon_early.cfg` |

## 五、存档兼容性规范（跨版本）

NSFW 独立游戏更新频繁，存档跨版本兼容是核心问题：

**规范要求**：
1. 存档格式版本号化（`save_version: 1.2.0`）
2. 存档加载时执行迁移脚本（`SaveMigrator.migrate(save_data)`）
3. 新增字段使用默认值填充，不强制重置
4. CG Gallery 解锁状态独立存储，不因版本升级丢失
5. 月度更新前必须测试当前存档在新版本的兼容性

## 六、隐私与反盗版

### 隐私功能（老板键）

| 功能 | 实现方式 | 优先级 |
|------|----------|--------|
| 一键最小化 | 快捷键监听 → `OS.set_window_minimized(true)` + 静音 | 高 |
| 窗口标题伪装 | 启动参数 / 设置选项，显示为"工具软件" | 中 |
| 默认禁用截图 API | `DisplayServer.screen_set_keep_on(false)` + 禁用 Steam 截图 | 高 |
| 存档加密 | 使用 XOR/AES 对存档文件内容加密 | 中 |
| 进程名伪装 | 修改 `.exe` 图标和文件名（发布包配置） | 低 |

### 反盗版（轻度）

| 方案 | 说明 | 适用场景 |
|------|------|----------|
| Steam DRM | Steamworks 内置，零成本 | Steam 版 |
| 关键资源动态解密 | 运行时解密 CG/音频，防静态提取 | DLsite/itch 版 |
| 激活码 | 购买后生成激活码，首次运行验证 | 独立发行版 |
| Patreon 版本水印 | 每个下载包嵌入用户标识 | Patreon 版 |
