# Godot 4.5 + MAME 街机 ROM 运行 — Android 插件设计方案（v2）

**决策基线**：同一个 APK；方案②（游戏 = 独立全屏 native Activity，Godot 只负责"拉起/收回"，不在 Godot 场景里叠 HUD）；MAME 代码零改动；大 `.so` 运行时下载。

## 1. 目标
一个**标准 Godot 4.5 Android 导出** APK，同时包含：
- **Godot App**：ROM 列表 + 全部 UX/导航（GDScript 场景）。
- **现有 MAME4droid 游戏 Activity**：native Java、全屏、自带 UI + `GLSurfaceView` 渲染。
- **一个 AndroidRuntime 插件 `MameRuntime`**：作为 Godot ↔ native 的桥，并负责 `.so`/ROM 动态下载。

## 2. 架构
```
┌────────────────────── 单个 APK（标准 Godot 导出）─────────────────────┐
│                                                                      │
│  Godot 引擎 + 你的场景                 MAME 侧 (native)               │
│  ┌────────────────────────┐            ┌──────────────────────────┐    │
│  │ RomList 场景（主菜单）  │ launch(rom)│ MAME 游戏 Activity        │    │
│  │  - ROM 列表 / 封面      │ ─────────▶ │ (com.seleuco.mame4droid  │    │
│  │  - 设置 / 关于         │ gameExited │  + 现成 Java UI)          │    │
│  │  （纯 Godot UI）       │ ◀───────── │  GLSurfaceView 渲染       │    │
│  └──────────┬───────────┘ finish→回   │  JNI → libMAME4droid.so  │    │
│             │                          └────────────┬───────────┘     │
│      AndroidRuntime 插件 MameRuntime（JavaClassWrapper 单例）          │
│      - ensureNative(): 下载 .so+ROM → System.load                     │
│      - launch(rom) = startActivity(Intent→游戏Activity, extra=rom)    │
│      - 信号 gameExited                                                │
└──────────────────────────────────────────────────────────────────────┘
  运行时下载到 app 私有目录:  libMAME4droid.so (74MB, 按 ABI) + ROMs
```

## 3. 组件

### 3.1 AndroidRuntime 插件 `MameRuntime`（可复用单元）
- **Java**：
  - `MameRuntime extends <Godot AndroidRuntime / JavaClassWrapper 基类>`，注册为 GDScript 单例。
    - 方法：`ensureNative()` / `launch(rom, extras)` / `isReady()` / `downloadProgress()`
    - 信号：`gameExited`
  - **复用现有 MAME4droid Java（原包名不动）**：游戏 Activity + `Emulator`(JNI 声明) + `GLRenderer`/`EmulatorViewGL` + 最小 ROM 加载/生命周期。UI 保持它自己那套（方案②，不重做）。
  - **Downloader**：拉 `.so` + ROM 到私有目录；选 ABI、断点续传、完整性校验。
- **Native**：
  - `libmame4droid-jni.so`(0.1MB) + `libc++_shared.so` → **随包带**（小，立即可用）。
  - `libMAME4droid.so`(74MB) → **运行时下载**。
- **GDScript API**（单例）：`MameRuntime.ensure_native()` / `.launch(rom)` / `.is_ready()` / signal `game_exited`。

### 3.2 Godot App（你的工程）
- `RomList` 场景：主菜单，列 ROM（来自 ship/下载的清单），封面/描述，纯 Godot UI。
- 点 ROM → `ensure_native()`（没下完就显示进度）→ `MameRuntime.launch(rom)`。
- `game_exited` 信号 → 回/刷新 RomList（可记"最近玩过"）。
- 设置 / 关于等全部 Godot。

### 3.3 Manifest / 打包集成（同 APK 的真实工作量在这）
- **一个 Application**：把 Godot 的 Application 与 MAME 侧所需初始化并到同一个 `Application` 子类。
- **两个 Activity**：Godot 主 Activity + MAME 游戏 Activity。
- **包名对齐 + JNI 绑定约束（关键）**：`libmame4droid-jni.so` 的 native 方法按**原 Java 类全限定名**注册（如 `com/seleuco/mame4droid/Emulator`）。→ **MAME 核心 Java 必须留在原包名/类名，不能重命名/换包**，否则预编译 .so 绑不上。
- **权限**：存储（下载 .so+ROM）、网络。
- **依赖**：现有工程若引了其它 AAR/库，一并 reconcile。

## 4. 运行时流程
1. 安装 → Godot App 启动 → `RomList`。
2. 首启 / 点 ROM 时 `ensure_native()`：私有目录缺 `.so`/ROM 就下载（Godot UI 显示进度）。
3. 点 ROM → `launch(rom)` → `startActivity`（extra=rom 路径）→ MAME Activity 加载 ROM（JNI `init/runT`）全屏出画面；Godot Activity 在后台 pause。
4. 游戏内退出 → `finish()` → Android 回到 Godot `RomList`（resume），刷新列表。

## 5. 保持不动的
- **MAME 源码**：0 改动（用预编译 `.so`）。
- **现有游戏 Activity + Java UI**：原样复用（只确保它从 Intent extra 读 ROM 路径、包名与 JNI 绑定一致）。
- **Godot 模板**：标准导出（待确认插件打包钩子①）。

## 6. 风险 / 待确认
- **① 插件打包**：4.5 AndroidRuntime plugin 是否"零模板手改、正常导出自动带上 Java+清单"，还是要一个极小的一次性 hook。→ 决定"100% 标准" vs "标准 + 极小声明"。（本机抓不到文档，需核原文或 spike）
- **Application / manifest 合并**：具体工作量取决于现有工程结构。
- **JNI 包名绑定约束**：MAME Java 必须留原包。
- **`.so` 下载**：ABI 选择、完整性校验、首跑体验。

## 7. 下一步
1. （可选）核 4.5 文档确认①的注册方式。
2. 翻现有 `mjguoguan1` 工程：找出真正的游戏 Activity + Application + manifest + 依赖 → 给出精确的 manifest 合并 / Application 并法 / 插件打包清单。
3. 搭最小 spike：一个最小 Godot 4.5 工程 + `MameRuntime` 插件 → 手机上"点 ROM → 进现有 MAME Activity 出画面 → 退出回列表"跑通。
