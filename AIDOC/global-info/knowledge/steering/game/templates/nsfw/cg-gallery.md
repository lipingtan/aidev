# CG Gallery / 回想系统模板

> 几乎所有 NSFW 游戏的标配系统，实现 CG 解锁、分类浏览、场景回放功能。

---

## 一、系统功能范围

| 功能 | 说明 | 优先级 |
|------|------|--------|
| CG 解锁记录 | 游玩中触发 CG/场景后自动解锁 | 高 |
| 分类浏览 | 按角色/场景类型/系列分类展示 | 高 |
| CG 查看 | 全屏查看 + 缩放 + 翻页 | 高 |
| 场景回放 | 重新播放已解锁的 H-Scene（从选定阶段开始） | 中 |
| 音频回放 | 重播特定语音/BGM | 低 |
| 解锁提示 | 显示未解锁 CG 的模糊缩略图 + 解锁条件提示 | 中 |
| 存档独立存储 | Gallery 解锁状态独立存档，不被主线存档覆盖 | 高 |

## 二、数据结构

```gdscript
# cg_gallery_data.gd
class_name CGGalleryData
extends Resource

## CG 记录条目
@export var cg_id: String                    # 唯一 ID
@export var title: String                    # CG 标题（用于 Gallery 显示）
@export var thumbnail_path: String           # 缩略图路径
@export var full_image_paths: Array[String]  # 全图路径（可多张，支持翻页）
@export var character_tags: Array[String]    # 关联角色（用于过滤）
@export var content_tags: Array[String]      # 内容标签（用于玩家过滤开关）
@export var related_scene_id: String = ""    # 关联的 H-Scene ID（用于场景回放）
@export var unlock_hint: String = ""         # 未解锁时显示的提示

# Gallery 解锁存档
class_name GalleryUnlockRecord
extends Resource

var unlocked_cg_ids: Array[String] = []
var unlocked_scene_ids: Array[String] = []
var total_cg_count: int = 0
```

## 三、解锁流程

```gdscript
# cg_gallery_manager.gd（全局单例）
signal cg_unlocked(cg_id: String, cg_data: CGGalleryData)
signal scene_unlocked(scene_id: String)

## 触发解锁（H-Scene 结束时调用）
func unlock_from_scene(scene_id: String) -> void:
    var cg_list = _get_cg_list_for_scene(scene_id)
    for cg_id in cg_list:
        if not _unlock_record.unlocked_cg_ids.has(cg_id):
            _unlock_record.unlocked_cg_ids.append(cg_id)
            cg_unlocked.emit(cg_id, _get_cg_data(cg_id))
            _show_unlock_toast(cg_id)
    
    _unlock_record.unlocked_scene_ids.append(scene_id)
    _save_unlock_record()  # 独立保存，不依赖主线存档

## 存档独立保存（关键：Gallery 解锁不受主线存档影响）
func _save_unlock_record() -> void:
    ResourceSaver.save(_unlock_record, "user://gallery_unlock.tres")
```

## 四、Gallery UI 布局规范

```
┌─────────────────────────────────────┐
│ [全部] [角色A] [角色B] [场景类型▼]   │  ← 过滤标签栏
├─────────────────────────────────────┤
│ □CG  □CG  ■CG  □CG  ■CG  □CG      │  ← CG 缩略图网格
│ □CG  ■CG  □CG  ■CG  □CG  □CG      │    ■=未解锁（模糊显示）
│ ■CG  □CG  □CG  □CG  ■CG  □CG      │    □=已解锁
├─────────────────────────────────────┤
│ 已解锁: 12/30                        │  ← 进度条
└─────────────────────────────────────┘
```

**UI 规范**：
- 未解锁 CG 显示模糊缩略图（不完全隐藏，给玩家方向感）
- 鼠标悬停在未解锁 CG 上显示解锁条件提示
- 支持内容标签过滤（对应玩家偏好开关）

## 五、场景回放

Gallery 中点击已解锁场景，可重新播放：

```gdscript
## 场景回放（绕过触发条件检查，直接播放）
func replay_scene(scene_id: String, start_phase: HScenePhase = HScenePhase.APPROACH) -> void:
    # 回放模式跳过触发条件评估
    HSceneManager.start_scene_replay(scene_id, start_phase)
```

## 六、存档跨版本兼容

- Gallery 解锁记录存放在独立文件 `user://gallery_unlock.tres`
- 主线存档重置/新游戏不影响 Gallery 解锁状态
- 版本升级时，旧解锁记录自动保留，新增 CG 的 `unlocked` 默认为 `false`
- 提供"重置 Gallery"选项（需用户二次确认）
