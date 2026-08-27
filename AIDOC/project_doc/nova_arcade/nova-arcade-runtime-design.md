# NOVA ARCADE 运行时架构 — 三运行时统一设计

> 上游：`nova-arcade-client-design.md` §2.4。本文定义游戏运行时的统一抽象与三种实现。
> 已全部 PoC 验证的部分标注 ✅（见 `nova-arcade-poc/`）。
>
> **状态（v2 定稿）**：街机运行时以 `projects/nova_arcade/mame-godot-plugin/DESIGN.md`（v2）为准：
> **标准 Godot 4.5 Android 导出 APK + 独立 native MAME4droid Activity（原包名复用）+ AndroidRuntime 插件 `MameRuntime` 桥接（拉起/收回），MAME 源码零改动，大 .so 运行时下载。**
> §3.1–§3.6 为现行设计；**附录 A 为已废弃的 myosd 进程内直嵌方案**（仅作历史参考，不得按此实现）。
> v2 行为细节（帧率/音频/进程语义/插件打包）待 M2 真机 spike 验证后定稿（DESIGN.md §6 风险项）。

---

## 0. 总览：Runner 抽象

```
商店/列表/详情页（对运行时完全无感知）
        │ meta.json {runtime: "pck"|"html"|"arcade"}
        ▼
┌───────────────── Launcher ─────────────────┐
│ 权益门 → 下载/校验 → RunnerRegistry.dispatch │
└──────┬──────────────┬──────────────┬───────┘
       ▼              ▼              ▼
  PckRunner      HtmlRunner     ArcadeRunner
  (Godot原生)    (WebView)      (MAME4droid native Activity + MameRuntime 桥, v2)
       │              │              │
       └──────── 统一 GameModule 生命周期协议 ────────┘
   boot(ctx{save_dir,tmp_dir,...}) / pause / resume
   quit_requested({score,playtime,achievements})
```

所有 Runner 实现同一接口（`services/runner.gd`）：

```gdscript
class_name Runner extends Node
func supports(runtime: String) -> bool
func launch(ctx: Dictionary) -> void      # ctx 含 gid/save_dir/tmp_dir/entry/meta
func pause() / func resume()
func request_quit() -> void               # 盒子侧主动（来电/用户返回）
signal quit_requested(result: Dictionary) # 游戏侧主动（正常退出）
```

结算卡、成就、时长统计、评价门槛只认 result 协议——**换运行时零改动**。
Arcade 在 v2 下由 `MameRuntime.game_exited` 信号驱动 `quit_requested`，与其他运行时同构。

---

## 1. PCK 运行时 ✅（已完整 PoC）

| 项 | 方案 |
|---|---|
| 包格式 | `<gid>.pck`（`godot --export-pack`，含编译后 GDScript） |
| 前缀隔离 | 强制包内资源全部位于 `res://games/<gid>/` 下；`load_resource_pack(pck, false)` replace_files=false 防覆盖盒子资源 |
| 运行 | `load("res://games/<gid>/game.tscn").instantiate()` → GameHost 子节点 |
| 存储 | `boot(ctx)` 注入 `user://saves/<gid>/` |
| 退出 | GDScript 信号 `quit_requested(result)` → `queue_free()` 释放整棵子树 |
| 已知限制 | PCK 无法卸载，资源索引残留至盒子重启（前缀隔离使其无害） |
| PoC | `shell/main.gd`：挂载→boot→存档→退出→tmp 清理→二次启动读到 best ✅ |

---

## 2. HTML 运行时 ✅（可无头部分已 PoC）

| 项 | 方案 |
|---|---|
| 包格式 | `<gid>.novahtml` = zip（index.html + 资源 + meta.json） |
| 解压 | 引擎核心 `ZIPReader` → `user://games/<gid>/` |
| 服务 | 本地 HTTP 文件服务（127.0.0.1:随机端口）：正确 MIME、CORS `*`、**路径穿越防护**（`..`/`\`/`%00` → 403）、404 兜底 |
| 加载 | 平台 WebView：桌面 gdcef / Android 系统 WebView / iOS WKWebView（各一个插件适配层） |
| 桥协议 | `nova://` URL 拦截：`nova://quit?score=&playtime=&achievements=` 退出；`nova://save?key=&value=` 写存档（写进该游戏 save_dir 的 kv.json）；`nova://get?key=` 读 |
| 存储 | kv.json（桥协议持久化）+ WebView 自身 localStorage（进程级，卸载即清） |
| 退出 | 拦截 WebView 导航事件 → 非 `nova://` 且非本域的导航一律阻断 |
| PoC | zip 往返/解压 ✅、HTTP 服务+MIME+穿越防护 ✅、桥协议解析 ✅（WebView 本体待 M5 真机） |

---

## 3. Arcade 运行时 — MAME4droid native Activity（v2，不用 libretro）

### 3.1 决策与参照实现

**决策：不引 libretro，也不做进程内直嵌。** v2 = 标准 Godot 导出 APK 里复用现有 MAME4droid 游戏 Activity（native Java、自带 UI + GLSurfaceView），Godot 只负责「拉起/收回」；MAME 代码零改动（用预编译 .so）；大 .so 运行时下载。

> **参照实现 = phoenixui 的 `mjarch3` flavor**（「街机麻将84合1经典」，productno mj70103，已上线单 APK 街机盒子）。其内容管线（romset zip + cfg/nvram/states 预置 + 版本标记升级语义）被 v2 沿用（§3.4/§3.5）。
> ⚠️ **不参照 gamearch 系列**：gamearch 走 RetroArch + libretro core（mamearcade/mame2016）+ `.lpl` 播放列表路线，与 NOVA 方向不同；mjarch3 zip 内残留的 `retroarch.cfg`/`*.opt` 属早期迭代遗留，现行启动路径不经过 RetroArch。

#### 3.1.1 mjarch3 实现拆解（已逐文件核实，内容管线参照）

| 层 | 实现 | 位置 |
|---|---|---|
| 壳 | baseuilib：`MyStubActivity`(LAUNCHER) → `entryactivity` 字符串或默认 `GameListActivity`；Tab=游戏列表/关于(+/日志 debug)；金币余额栏（首次安装赠 5 枚） | `baseuilib/.../GameListActivity.java` |
| 内容包 | `res/raw/files.zip`：`roms/*.zip`×84 + `playlist/gamelist(.default)` JSON + 预置 `cfg/<rom>.cfg`（键位/DIP）+ 预置 nvram/states + `cfg/default.cfg` | `app/src/mjarch3/res/raw/` |
| 引擎 | **libmain.so（MAME 0.288 myosd 构建）打进 APK jniLibs**；`init()` 时 `dlopen(nativeLibraryDir)` + `dlsym` 全部 `myosd_droid_*` / `myosd_video_*` 符号 | `MAME4droid/.../jni/mame4droid-jni.c load_lib()` |
| 首启 | 解压 files.zip → app 私有目录 `HBMJ1/`（进度弹窗，禁取消）；版本标记 `saves/dont-delete-<ver>.bin`；升级时删 cfg/nvram 后重解新文件（已存在文件跳过） | `Utils.copyGameFiles` / `GameListActivity.ensureFiles` |
| 游戏列表 | 本地 `playlist/gamelist` JSON（空则从 `.default` 种子）；可选扫描 `arom*/roms` 子目录发现用户 ROM zip；远端 `checkGameConfig` 拉取列表 | `GameListFragment.java` |
| 每游戏配置 | `GameConfig{romName, romPath(相对), controllerType: virtualkey/joystick, buttonNumber=20, buttonConfigJson(按钮几何), cansl(可否存状态), title/desc/image}` | `common/GameConfig.java` |
| 启动 | 列表点击 → `MAME4droid` activity（intent: gamename+gameconfig）→ 后台线程：`init(nativeLibDir, HBMJ1/, w, h)` → `setValueStr(VERSION/ROM_NAME/ROM_PATH)` → 每游戏 ButtonConfig + `updateEmuValues()` → `runT()` | `MAME4droid.java` / `Emulator.emulateRom` |
| native 入口 | `myosd_droid_main`：拼 argv（rom_name + `-noconfirm_quit -natural -nocoin_lockout [-rompath] -samplerate [-plugin hiscore …]`）→ `myosd_main(argc,argv,callbacks)` 独立线程 | `src/osd/myosd/myosd-droid.cpp` |
| 视频/音频 | `video_draw` 回调 → Java `requestRenderFrame` → GL `onDrawFrame` → `myosd_video_onDrawFrame`（核心内 GLES3 renderer，shader 面已预留）；音频默认 OpenSL ES，可切 Java SoundThread 走 `dumpSound` 回调 | `mame4droid-jni.c` / `opensl_snd.cpp` |
| 准入/签名 | FlavorFactory 反射（strings.xml 类名）：`IAccessControl` 激活码 / `ICoinManager` 金币；APK 签名校验 `sig_hash`（security.cpp，对应 app_new.jks） | `common/FlavorFactory.java` |

#### 3.1.2 本树可复用资产盘点

本地 `C:\data\developer\projects\Mame\mame0288` 树内已有：

| 资产 | 内容 | 作用（v2 视角） |
|---|---|---|
| `MAME4droid/` 完整 Android 工程 | 游戏 Activity + `Emulator`(JNI 声明) + `GLRenderer`/`EmulatorViewGL` + ROM 加载/生命周期（原包名 `com.seleuco.mame4droid`） | **v2 直接复用的 native 侧**（Activity + Java UI 原样保留，只确保从 Intent extra 读 ROM 路径） |
| `mame4droid-jni.c` | JNI 胶水：`load_lib`（dlopen+dlsym）、SAF 回调、视频/音频 Java 回传 | 随包小 .so 的绑定参照；**JNI 按原 Java 类全限定名注册 → MAME Java 必须留原包名/类名** |
| `Retro/RetroArch/pkg/android/phoenixui/app/src/mjarch3/` + `baseuilib/` | 已上线街机盒子：内容包解压/版本标记/游戏列表 JSON/每游戏按钮配置/启动时序 | **内容管线与产品流程参照**（§3.1.1/§3.4/§3.5） |
| `src/osd/myosd/`、`renderer/`、`security.cpp` 等 MAME 源码 | myosd C ABI、GLES renderer+shader、APK 签名校验 | 构建链参考（v2 用预编译 .so，不改源码）；附录 A 历史方案素材 |
| `android-project/` | Gradle 工程（构建 myosd APK / libmain.so） | 核心 .so 构建链参考 |

### 3.2 Godot 桥接：MameRuntime（AndroidRuntime 插件，v2）

```
┌────────────────────── 单个 APK（标准 Godot 导出）─────────────────────┐
│                                                                      │
│  Godot 引擎 + 你的场景                 MAME 侧 (native)               │
│  ┌──────────────────────┐            ┌──────────────────────────┐    │
│  │ RomList 场景（主菜单） │ launch(rom)│ MAME 游戏 Activity        │    │
│  │  - ROM 列表 / 封面      │ ─────────▶ │ (com.seleuco.mame4droid  │    │
│  │  - 设置 / 关于         │ gameExited │  + 现成 Java UI)          │    │
│  │ （纯 Godot UI）       │ ◀───────── │  GLSurfaceView 渲染       │    │
│  └──────────┬───────────┘ finish→回   │  JNI → libMAME4droid.so  │    │
│             │                          └────────────┬───────────┘     │
│      AndroidRuntime 插件 MameRuntime（JavaClassWrapper 单例）          │
│      - ensureNative(): 下载 .so+ROM → System.load                     │
│      - launch(rom) = startActivity(Intent→游戏Activity, extra=rom)    │
│      - 信号 gameExited                                                │
└──────────────────────────────────────────────────────────────────────┘
```

**组件清单：**

| 组件 | 形态 | 说明 |
|---|---|---|
| `MameRuntime`（AndroidRuntime 插件） | Java，注册为 GDScript 单例 | `ensure_native() / launch(rom, extras) / is_ready() / download_progress()` + 信号 `game_exited`；复用现有 MAME4droid Java（Activity + `Emulator` JNI 声明 + GLRenderer/EmulatorViewGL），**原包名不动** |
| Downloader（插件内） | Java | 拉 `.so` + ROM 到 app 私有目录：ABI 选择、断点续传、完整性校验（sha256，服务端签名见 §3.4） |
| 随包小 .so | `libmame4droid-jni.so`(~0.1MB) + `libc++_shared.so` | JNI 胶水按原 Java 类全限定名注册 → **MAME 核心 Java 必须留在原包名/类名，不能重命名/换包**，否则预编译 .so 绑不上（DESIGN.md §3.3 关键约束） |
| 大 .so 运行时下载 | `libMAME4droid.so`（~74MB，按 ABI 分发） | 首启下载到 app 私有目录 → `System.load`；版本随 MAME 0.288 锁定 |

**GDScript API（单例）**：`MameRuntime.ensure_native() / .launch(rom) / .is_ready()` + 信号 `game_exited`。
Shell 侧 `CoreManager`（client §2.2）是该插件的包装层：Android 委托插件；桌面/无 Arcade 环境返回「不可用」（arcade 游戏仅真机可玩，CTA 置灰）。

**v2 待验证风险（M2 spike 项，DESIGN.md §6）：**
1. 4.5 AndroidRuntime 插件能否「零模板手改、标准导出自动带上 Java+清单」，还是需要极小打包钩子
2. Application/manifest 合并工作量（Godot Application 与 MAME 侧初始化并入同一子类；两个 Activity）
3. JNI 包名绑定约束（MAME Java 留原包）
4. `.so` 下载：ABI 选择 / 完整性校验 / 首跑体验

### 3.3 启动时序与生命周期（v2）

```
launch(ctx):
 1. 权益门（PayGate/TrialGuard，客户端 §4.6）
 2. 内容检查：user://games/<gid>/ 说明文件 + rom zip 齐全？→ 缺则 DownloadDialog（§3.4）
 3. 核心检查：MameRuntime.is_ready()？→ 未就绪 ensure_native()（Godot UI 显示下载进度，§3.5）
 4. MameRuntime.launch(rom, extras)：startActivity(Intent→MAME 游戏 Activity，extra=ROM 路径/cfg/buttonConfig)
    → MAME Activity 加载 ROM（JNI init/runT）全屏出画面；Godot 主 Activity 后台 pause
 5. 游戏中：native 侧独立渲染（GLSurfaceView + OpenSL 音频），Godot 不参与帧/音频/输入；
    虚拟按键/buttonConfig/cansl/hiscore 由 MAME 自带 UI 处理（GameConfig 经 Intent extra 传入）
 6. 退出：游戏内退出 → Activity finish() → Android 回 Godot RomList（resume）
    → MameRuntime 发 game_exited → Launcher 走结算卡流程（client §4.6 步骤11-16）
```

**生命周期语义（v2，替代 myosd 直嵌的进程语义表）：**

| ROM 来源 | [再来一局] | [返回] |
|---|---|---|
| 内置/授权 ROM（走内容包管线，已签名+校验） | `launch(rom)` 重新拉起 Activity（每局一次 Activity 生命周期） | `game_exited` → 回盒子 |
| 用户自导入 ROM（路径经 Intent extra，§3.4 自导入路线） | 同上 | 同上 |

- Godot 进程全程不杀、无进程内 native 状态——**没有 mjarch3 killProcess 分支**；「再来一局」= 重新 launch。
- **分数提取**：hiscore 插件 / nvram 路线（MAME 侧写 game_dir，Shell resume 后读取）；提取不到报 `score:-1`，结算卡隐藏分数栏。成就用 playtime/帧数规则兜底。

### 3.4 内容下发格式（服务端 → 游戏专属文件夹）

**实际交付模型（与 mjarch3 全量 files.zip 的差异点）**：mjarch3 把 84 个 ROM + 列表 + cfg 打进一个 `res/raw/files.zip`；NOVA 改为**服务端按游戏直接下发 MAME 原生 romset zip（可能多个）+ 一份说明文件**，客户端下载后放进该游戏的专属文件夹——不再二次打包：

```
user://games/<gid>/            # 专属文件夹 = 盒子里的一个"游戏"
  meta.json                   # 说明文件（schema 见下）
  roms/
    pacman.zip                # MAME 原生 romset zip，原样落盘
    pac-man2.zip              # 允许多个：一个 gid 可含多台机器/克隆体
  cfg/                        # 可选：每机键位/DIP 覆盖（MAME <rom>.cfg 格式，同 mjarch3 预置思路）
  states/                     # 可选：预置 savestate 种子数据（同 mjarch3 预置 states 思路）
  nvram/                      # 可选：预置 hiscore.dat/NVRAM 种子数据（高分榜冷启动）
```

> Android 上 `user://` 对应 app 私有目录（filesDir），MAME Activity 经 Intent extra 拿到的绝对路径可直接读写这些文件。

**为什么一个 gid 会需要多个 rom zip？** MAME 有三种常见场景：

| 场景 | 说明 | 举例 |
|---|---|---|
| **多机合一**（mjarch3"84合1"） | 一个 gid 包含 84 个独立游戏，每个游戏一个驱动 + 一个 rom zip | "街机麻将合集" |
| **Parent/Clone** | MAME 中每个驱动有 `parent` 字段（`gamedrv.h`）：clone 的 romset 只包含和 parent 不同的 ROM，加载时 MAME 自动沿 parent 链查找所有 rom 文件 | `pac-man2` 的 parent 是 `pacman`；启动 `pac-man2` 时需要 `pacman.zip` + `pac-man2.zip` 都在 rompath |
| **BIOS 依赖** | 一些驱动是 BIOS 系统（`MACHINE_IS_BIOS_ROOT`），其他驱动引用该 BIOS 的 `parent` 字段。CPS1/CPS2/CPS3/NeoGeo 等系统级 BIOS 就是典型 | `cps1` BIOS 被所有 CPS1 游戏引用；启动时 CPS1 游戏 zip + `cps1.zip` 都需存在 |

> **关键约束**：MAME 的 romset 校验以 zip 为单位——每个 zip 包含一个完整 romset（或其独有部分）；启动时 MAME 在 `rompath` 下搜索每个 rom 文件，沿 parent/BIOS 链查找。因此服务端下发时**每个 zip 文件独立签名**，客户端逐个校验，MAME 原生 romset 校验做最终兜底。

说明文件 `meta.json`（= GameMeta 街机字段 + GameConfig 等价物，每个 ROM 一条 entry）：

```json
{
  "id": "pacman",
  "title": "PAC-MAN",
  "runtime": "arcade",
  "core": "mame4droid",
  "core_version": "0.288",
  "roms": [
    {
      "name": "pacman",
      "title": "PAC-MAN",
      "parent": "",
      "bios": "",
      "zips": ["pacman.zip"],
      "controllerType": "virtualkey",
      "buttonNumber": 4,
      "cansl": true,
      "buttonConfig": {"buttonwidth": 140, "buttonheight": 260, "intervalwidth": 5, "offsetleft": 58},
      "hiscore": "plugin"
    },
    {
      "name": "pac-man2",
      "title": "PAC-MAN (alternate)",
      "parent": "pacman",
      "bios": "",
      "zips": ["pac-man2.zip"],
      "controllerType": "virtualkey",
      "buttonNumber": 4,
      "cansl": true,
      "buttonConfig": {"buttonwidth": 140, "buttonheight": 260, "intervalwidth": 5, "offsetleft": 58},
      "hiscore": "plugin"
    }
  ],
  "price_model": "trial"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `roms[].name` | `string` | MAME 驱动名（romset 名），launch 时经 Intent extra 传入 |
| `roms[].parent` | `string` | 父驱动名（空=无父；clone 时填 parent 名）；与 `gamedrv.h` / `myosd_game_info.parent` 同源 |
| `roms[].bios` | `string` | 依赖的 BIOS 驱动名（空=无 BIOS 依赖）；与 `MACHINE_IS_BIOS_ROOT` 标志关联 |
| `roms[].zips` | `string[]` | **该驱动需要的 rom zip 文件列表**（每个 zip 对应 rompath 下的一个文件）。单 romset 一个元素；多 zip 时按 MAME 原生要求列出 |
| `roms[].title` | `string` | 标题 |
| `roms[].controllerType` | `enum` | `virtualkey` / `joystick` |
| `roms[].buttonNumber` | `int` | 按键数量 |
| `roms[].cansl` | `bool` | 是否允许存状态 |
| `roms[].buttonConfig` | `object` | 按钮几何：`buttonwidth`/`buttonheight`/`intervalwidth`/`offsetleft`（MAME 侧 UI 消费，经 Intent extra 传入） |
| `roms[].hiscore` | `enum` | `plugin` / `nvram` / `none` |

**服务端 /catalog 对 arcade 游戏的 manifest 结构：**

```json
{
  "gid": "pacman",
  "version": 5,
  "resources": {
    "core": { "type": "core", "version": "0.288", "required": true },
    "arcade": {
      "type": "arcade",
      "files": [
        { "path": "pacman.zip", "sha256": "abc123...", "cdn": "https://cdn/.../pacman.zip" },
        { "path": "pac-man2.zip", "sha256": "def456...", "cdn": "https://cdn/.../pac-man2.zip" }
      ],
      "meta_json": { "path": "meta.json", "sha256": "ghi789...", "cdn": "https://cdn/.../meta.json" },
      "signatures": {
        "pacman.zip": "sig1",
        "pac-man2.zip": "sig2",
        "meta.json": "sig3"
      }
    }
  }
}
```

- **每个 gid 的 manifest 包含该 gid 下所有 rom zip 文件的 CDN 地址和 sha256**（不管单 zip 还是多 zip）
- `roms[].zips[]` 每个元素对应 `files[]` 中的一个条目，通过文件名匹配
- 服务端按每个 zip 单独签名（`signatures` 字段），客户端先验签名再写入磁盘

**下载与校验流程（单 gid，含 parent/clone/BIOS 场景）：**

```
1. /catalog 获取 manifest → 解析所有文件清单 {path, sha256, cdn, sig}
2. 对每个文件：
   a. 如果 user://games/<gid>/roms/<path> 已存在 → sha256 比对
      - 一致 → 跳过
      - 不一致 → 重新下载
   b. 如果不存在 → 从 cdn 下载 → sha256 校验 → 签名校验 → 写入
3. MAME 原生校验：
   a. MAME Activity 启动后，MAME 在 rompath 下扫描每个驱动
   b. 对每个 romset：
      - 检查 zip 内所有 rom 文件的 CRC/SHA1 是否在 game_list 中匹配
      - 沿 parent 链：clone 的每个 rom → 先在 clone.zip 中找 → 找不到则到 parent.zip 中找
      - BIOS 链：BIOS 依赖的每个 rom → 在 bios.zip 中查找
   c. 全部通过 → 启动；任一失败 → 透传 MAME 错误文本到 UI
```

- **一个 gid = 盒子一个条目**；gid 内多个 ROM 在列表/详情页展开为可玩子条目（mjarch3「84合1」即此形态：一个产品、84 个 ROM、各自 buttonConfig）
- **不重打包**：romset zip 保持 MAME 原生格式 → 加载时走 **MAME 原生 romset 校验**（逐 CRC/SHA1，比自建清单更权威，报错文本直接透传 UI）
- 校验链：每文件 sha256（服务端签名）→ MAME romset 校验 → 驱动存在性（`game_list` 缓存查询）
- **核心分发（v2）**：mjarch3 把 libmain.so 打进 APK jniLibs；NOVA 主包体积策略不允许 → 小 JNI 胶水 .so（~0.1MB）+ `libc++_shared.so` **随包**，`libMAME4droid.so`（~74MB，按 ABI 分发）**运行时下载**到 app 私有目录（首启一次），版本随 MAME 0.288 锁定
- **用户自导入 ROM 不走服务端**：扫描 `user://roms/arom*/` 子目录（mjarch3 `getRomsaNames` 路线）→ 驱动名=文件名自动生成 meta → 标注「自导入」，路径经 Intent extra 传入

**Parent/Clone 对 UI 的影响：**

| 场景 | 列表/详情页行为 | 启动方式 |
|---|---|---|
| 单 ROM（无 parent） | 直接显示标题 + CTA | `MameRuntime.launch(rom_name)` |
| Clone（有 parent） | 子条目标题标注 "(clone of <parent_title>)"；列表显示时与 parent 区分 | `launch(clone_name)`（ROM 路径经 Intent extra；MAME 自动在 rompath 下找 parent.zip） |
| 多机合一（84合1） | 展开为 84 个可玩子条目（每个有独立 buttonConfig）；每个子条目的 parent/bios 各自声明 | 用户选择哪个子条目 → `launch(roms[].name)` |
| BIOS 游戏 | 如果所需 BIOS zip 未下载，详情页显示 "需要 <bios_title> BIOS" 并附带 [下载] CTA（如果 BIOS 也属于本 gid） | MAME 自动在 rompath 下找 bios.zip |

### 3.5 存储布局与升级语义

```
<filesDir>/mame/               # 核心包（v2）：libMAME4droid.so（按 ABI，~74MB）+ 版本标记文件 .version
                               # （小 jni .so + libc++_shared.so 随 APK 打包，不在此目录）
user://games/<gid>/            # 见 §3.4
  states/                      # MAME savestate（自动+快捷槽）
  nvram/                       # NVRAM / hiscore.dat 高分
  cfg/                         # 键位、DIP（-cfg_directory）
user://roms/arom*/            # 用户自导入 ROM（可选功能）
user://tmp/<gid>/             # ROM 工作副本（如需解包），退出清理
```

MAME 各目录参数全部指向 game_dir 子目录——**天然按游戏隔离**，文件布局与 MAME4droid/phoenixui 生态工具兼容。

**升级语义（借鉴 mjarch3，规避其已知坑）**：
- 内容包版本变化 → 重下变更文件（已存在文件跳过，同 `copyGameFiles`）+ 重置 `cfg/`（键位/DIP 可能随 ROM 更新失效）
- **保留 `nvram/` 与 `states/`**——mjarch3 的 `removeGameCfgFiles` 升级时连 nvram 一起删（游戏进度被清，其知识库已记录该问题），我们只重置 cfg
- 核心 .so 版本变化 → 重下 `libMAME4droid.so` 至 `<filesDir>/mame/`（删除旧文件），更新所有游戏 cfg（每游戏读 .version 标记确认一致性）

### 3.6 合规要点（更新）

- MAME 0.274+ 为 BSD-3-Clause，**可商用分发核心**；myosd 同为 BSD-3（文件头注明）；发行标注 MAME 许可与版本
- ROM 只走两条路：① 版权方授权（服务端下发 rom zip 集 + 说明文件，签名）② 用户自导入（扫描路线，标注「自导入」，盒子不提供 ROM）
- **ROM 合规加固（review B-1）**：MAME ROM 多为无正式授权渠道的灰区，「授权下发」需限定为**授权/公有域 ROM 白名单**并对每个 rom zip 保留授权凭证；「自导入」ROM 独立目录存放+明确免责标注，盒子侧不预置/不分发任何 ROM；商用分发前完成 ROM 版权尽调
- mjarch3 的激活码体系（`IAccessControl`）**不采用**——NOVA 走应用商店分发 + PayGate 权益门；其金币系统（`ICoinManager`，首次安装赠 5 枚）M4 可作为可选「街机币」变现 flavor 再引入
- APK 签名校验（sig_hash/security.cpp）→ 替换为核心包/内容包 sha256 + 服务端签名

### 3.7 性能与热路径（v2）

- 游戏帧率/音频/输入全部由 native 侧承担（GLSurfaceView + OpenSL ES），**Godot 主循环不参与热路径**；游戏中 Godot 进程后台 pause
- DRC JIT：mjarch3 默认 `-nodrc`（稳定性优先）；NOVA 建议默认 `-drc -drc_use_c`（移动端性能），经设置页可切回（MAME 侧参数，经 Intent extra 传入）
- **真机待验证（M2 spike）**：低端机帧率 / 音频延迟 / Godot pause 期间后台行为与功耗 / `.so` 首跑下载体验

---

## 4. 三运行时对照总表

| | PCK | HTML | Arcade |
|---|---|---|---|
| 包 | .pck | .novahtml (zip) | **rom zip 集（可多个）+ 说明文件**（不重打包，落专属文件夹） |
| 下载校验 | sha256 | sha256 | 每文件 sha256 + **MAME 原生 romset 校验** |
| 运行体 | Godot 场景树 | WebView | **MAME4droid native Activity + MameRuntime 插件桥（v2，不用 libretro）** |
| 首次附加下载 | 无 | 无 | `libMAME4droid.so` ~74MB（按 ABI，一次性；小 jni .so 0.1MB 随包） |
| 视频路径 | 原生渲染 | WebView 合成 | native GLSurfaceView 直渲（Godot 不参与） |
| 音频 | 原生 | WebView | native OpenSL ES（Godot 不参与） |
| 输入 | 原生事件 | 触摸/桥 | MAME 自带 UI/虚拟按键（buttonConfig 经 Intent extra 传入） |
| 存档 | save_dir 直写 | 桥协议 kv | MAME 目录参数 → game_dir 子目录；升级只重置 cfg，保留 nvram/states |
| 退出信号 | GDScript signal | nova://quit | Activity `finish()` → `game_exited` 信号（v2 无进程 kill 分支） |
| 分数来源 | result.score | URL 参数 | hiscore 插件 / nvram 提取 |
| 离线可玩 | ✅ | ✅ | ✅ |
| 合规要点 | 无特殊 | 内容审核 | 核心 BSD-3 + ROM 授权/自导入 |
| PoC 状态 | ✅ 全通 | ✅ 壳层全通 | ⏳ v2 spike 未开始（M2：拉起/收回/帧率/音频 + DESIGN.md §6 风险项） |
| 云存档 | 登录后后台同步 blob（M3+） | 登录后后台同步 blob（M3+） | **不走云存档**；savestate/nvram/cfg 仅存客户端本地；best_score/playtime 随 leaderboard 接口上报 |

---

## 5. 对客户端/服务端设计的增量修订

- `GameMeta.runtime: "pck"|"html"|"arcade"`（客户端 §3.4 已有字段，语义升级为分发依据）
- 下载器增加"核心依赖"步骤：arcade 游戏先查 `MameRuntime.is_ready()`（未就绪 → `ensure_native()` 下载，§3.2/§3.5）
- 我的→设置 增加：核心管理（版本/删除重下）、省电模式、组合键自定义；**ROM 管理**（[扫描 user://roms/ 目录] 手动触发按钮，发现的 ROM zip 列表标注「自导入」，点击可直接启动；入口位于设置页而非 Library Tab，避免与商店内容混合）
- 服务端 `/catalog` 增加 `core` 资源类型，以及**街机内容包（rom zip 集 + 说明文件）**的签名清单下发；用户自导入 ROM 不经过服务端
- 商店合规位：每个 arcade 游戏标注「授权分发」或「请自行导入 ROM」
- **驱动清单离线化**：首次核心安装后把 `game_list` 缓存为本地 JSON，商店的 arcade 目录/校验/搜索别名可离线工作

---

## 附录 A. 已废弃方案：myosd 进程内直嵌（历史参考）

> 以下为 v2 之前的进程内直嵌方案（MyOsdHost GDExtension + ImageTexture 帧回传 + Godot 侧虚拟按键）。
> **已被 §3.2/§3.3 v2 完全替代，仅作历史参考，不得按此实现。** 废弃原因：GDExtension 绑定工作量大、帧回传 60fps 稳定性未验证、与 mjarch3 已上线的 native Activity 路线重复造轮子。

### A.1 Godot 绑定：MyOsdHost（GDExtension）

新增一个 `myosd-godot.cpp`（与 `myosd-droid.cpp` 平级，互不影响），实现 `myosd_callbacks` 并注册到 ClassDB：

```
MAME 0.288 (src/emu, src/frontend)          ← 不改
   ▲ myosd_core.h  C ABI
   ▼
myosd-godot.cpp  (GDExtension, C++)
   - dlopen 核心包内 libmain.so（同 mjarch3 load_lib 模式：.gdextension 注册入口加载 myosd-godot.cpp，其内部 dlopen libmain.so）
   - myosd_main 独立线程（runT 同构）；回调经 call_deferred 回传 Godot 主线程
   - video_draw → 帧缓冲 → ImageTexture.update（60fps，节流见 A.3）
   - sound_play(OpenSL 同构) → AudioStreamGenerator 推流（Godot 音频管线接管混音/静音）
   - input_poll ← Godot 每帧写 myosd_input_state（虚拟按键位标志/触摸/键盘/光枪）
   - game_list → 驱动清单缓存（ROM 目录/校验/搜索别名的离线数据源）
   ▲ ClassDB (Godot API)
   ▼
ArcadeRunner (GDScript) — 实现 Runner 接口
```

**接口面与 mjarch3 JNI 同构**（`Emulator.java` native 面 → ClassDB，键表同源 `com_seleuco_mame4droid_Emulator.h`）：

| mjarch3 JNI | MyOsdHost (ClassDB) | ArcadeRunner 用途 |
|---|---|---|
| `init(libDir, resPath, w, h)`（dlopen+回调注册+chdir） | `init(core_dir, game_dir, w, h)` | core_dir=核心包目录；game_dir=`user://games/<gid>/` |
| `setValueStr(VERSION/ROM_NAME/ROM_PATH)` | `set_version(str)` / `set_rom(name, path="")` | ROM_PATH 空=内置 roms/ 目录；非空=用户导入路径（触发 A.2 进程语义） |
| `setValueStr(CLI_PARAMS)` | `set_cli_params(str)` | 追加 MAME 参数（DRC/滤镜/采样率覆盖，见 A.3） |
| `setValue(PAUSE/F6存/F7读/HISCORE/分辨率/DRC…)` | `set(key, value)` | 整型键表与 droid 层一致：暂停、savestate/loadstate、hiscore 插件开关等 |
| `getValue(IN_GAME/IN_MENU/NUMBTNS/NUMWAYS/IS_MOUSE/IS_LIGHTGUN)` | `get(key) -> int` | 驱动盒子 UI 状态机（菜单中/游戏中/玩家数/按键数） |
| `setDigitalData/setAnalogData` | `set_digital(bits, player)` / `set_analog(type, player, x, y)` | 虚拟摇杆 → MYOSD_* 位标志（UP=0x1 LEFT=0x4 DOWN=0x10 RIGHT=0x40，START/SELECT/A-D/L1-R2） |
| `setKeyData/setMouseData/setTouchData` | `push_key(code, action, ch)` / `push_mouse(ev)` / `push_touch(ev)` | 键盘通道（服务菜单 F2）、鼠标、光枪 |
| `onDrawFrame(renderer, hdr)` → `myosd_video_onDrawFrame` | signal `frame_ready` → `ImageTexture.update` | mjarch3 走 GL 直渲；Godot 侧走纹理（shader 面 getShaders/setShader 保留，M5+ 评估）。`ImageTexture.update` 在 Godot 4.5 Mobile 渲染器上的 60fps 稳定性**待 M2 Android 真机验证**（估算：640×480 纹理约 2~4ms/帧，但实际受设备 GPU/驱动影响）；如真机帧率不达标可切 M5+ 的 GL 直渲路径（参考 mjarch3 onDrawFrame→myosd_video_onDrawFrame） |
| `runT()`（线程：android_main→myosd_droid_main→myosd_main） | `run_threaded(argv: PackedStringArray)` | argv 模板见 A.2；myosd_main 起独立线程 |
| （退出）`keyboard[MYOSD_KEY_EXIT]` / `pushEvent(MYOSD_EXIT)` | `request_quit()` → signal `exited(code)` | 组合键/盒子主动退出 → 退出协议 |

### A.2 启动时序与进程语义（镜像 mjarch3）

```
launch(ctx):
 1. 权益门（PayGate/TrialGuard，客户端 §4.6）
 2. 内容检查：user://games/<gid>/ 说明文件 + rom zip 齐全？→ 缺则 DownloadDialog（§3.4）
 3. 核心检查：myosd-0.288 核心包已装？→ 未装先下（系统级 DLC，§3.4）
 4. MyOsdHost.init(core_dir, game_dir, w, h)          # mjarch3: init(nativeLibDir, HBMJ1/, w, h)；core_dir=user://cores/myosd-{version}/
 5. set_rom(rom_name[, rom_path]) + set_version + 每游戏按钮配置/cansl/hiscore 标志
 6. run_threaded(argv)：                              # mjarch3: runT()
      argv = [rom_name, -noconfirm_quit, -natural, -nocoin_lockout,
              -rompath <game_dir/roms>,
              -state_directory <game_dir/states>, -nvram_directory <game_dir/nvram>,
              -cfg_directory <game_dir/cfg>,
              -samplerate 48000, -plugin hiscore(若启用), ...]
      → myosd_main(argc, argv, &callbacks)            # 独立线程，阻塞至游戏退出
 7. 游戏中：input_poll ← Godot 触摸/虚拟按键；frame_ready → TextureRect；sound_play → StreamGenerator
 8. 退出（见下）
```

**进程语义（照抄 mjarch3 的分野）**：

| ROM 来源 | 退出行为 | 依据 |
|---|---|---|
| 内置/授权 ROM（走内容包管线，已签名+校验） | 自动 savestate → `push_event(EXIT)` → `myosd_main` 返回 → `quit_requested(result)`；**同会话可再进**（"再来一局"不重建 native 层） | MAME4droid 验证过的多局模式（isEmulating 复位即可再 emulateRom） |
| 用户自导入 ROM（`rom_path` 非空） | 退出后**强制进程重启**：Godot `OS::quit()` + Android activity 拉起；盒子 UI 状态经 DB 完整恢复 | mjarch3 对外部 ROM 的 `Process.killProcess`——外部 ROM 加载后的 native 状态不保证可干净重初始化，进程级 teardown 最可靠 |

**演进选项（M5+ 评估）**：myosd_main 移入 helper 子进程（`nova-mame-helper`，dlopen 同一 libmain.so），IPC = UDS + 共享内存帧环——盒子进程永不死、崩溃隔离、"再来一局"=重启子进程。M1 基线用进程内方案（部件最少），真机验证后按稳定性数据决定是否切子进程。

### A.3 性能与热路径

- `myosd_main` 独立线程 + `call_deferred` 传帧（不阻塞 Godot 主循环）
- 音频缓冲：AudioStreamGenerator 为 push 模式，`sound_play` 回调每次推一批样本到生成器队列；建议缓冲约 3 帧（@60fps ≈ 50ms，3×16.7ms）防止 myosd 线程与 Godot 音频线程帧率轻微不同步造成的欠压漂移。具体毫秒数待 M2 Android 真机实测后按设备音频延迟调整。
- 帧率以 MAME av info 为准（60 帧机器 vs 120Hz 屏：纹理更新节流至 60fps）；省电模式半帧率
- DRC JIT：mjarch3 默认 `-nodrc`（稳定性优先）；NOVA 建议默认 `-drc -drc_use_c`（移动端性能），经 `set_cli_params`/设置页可切回
- 渲染走 myosd 自带 renderer（可后续接 shader：`getShaders/setShader` 面已在 ABI 里预留）
