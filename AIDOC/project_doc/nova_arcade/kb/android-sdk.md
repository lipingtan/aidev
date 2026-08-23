# Android SDK / NDK / 签名配置

## SDK 路径

```
C:\data\androidsdk\
```

`local.properties` 写法（phoenixui 工程已配置，Godot 导出工程同理）：

```properties
sdk.dir=C\:\\data\\androidsdk
```

## 已安装组件

### Build Tools
| 版本 |
|---|
| 35.0.0 ← phoenixui 用这个 |
| 36.0.0 |
| 34.0.0 |
| 30.0.3 |
| 25.0.x |

### NDK
| 版本 | 备注 |
|---|---|
| `22.1.7171670` | phoenixui `build.gradle` 里指定的版本（`ndkVersion`） |
| `28.2.13676358` | 最新稳定版 |
| `x16.1.4479499` | 旧版本 |

### Platforms（Android API Level）
已安装：14 / 15 / 18 / 21 / 24–29 / 32 / 35 / 36

phoenixui 用 `compileSdk 35`，`minSdk 21`，`targetSdk 28`。

### JDK
路径：`C:\data\androidsdk\jdk\`

| 目录 | 版本 |
|---|---|
| `jdk-21` | Java 21（推荐，Android Gradle Plugin 8.x 要求） |
| `OpenJDK11U` | Java 11 |
| `jdk1.8` | Java 8（旧项目兼容） |

---

## 签名 Keystore

参考工程：`phoenixui`（mjarch3 等已上线产品均用此签名）

| 文件 | 路径 |
|---|---|
| `app_new.jks` | 本工作区 `AIDOC/project_doc/nova_arcade/kb/app_new.jks`（原件位于 `C:\data\developer\projects\Mame\mame0288\Retro\RetroArch\pkg\android\phoenixui\app_new.jks`） |
| `app.jks` | 旧版本，仍在原目录，暂不迁移 |

签名配置（来自 `config.gradle`，debug 和 release 共用同一个 keystore）：

```gradle
// phoenixui/config.gradle
ext {
    releaseSign = [
        storeFile    : project.file('app_new.jks'),
        storePassword: "111111",
        keyAlias     : "key0",
        keyPassword  : "111111"
    ]
    debugSign = [
        storeFile    : project.file('app_new.jks'),
        storePassword: "111111",
        keyAlias     : "key0",
        keyPassword  : "111111"
    ]
}
```

在 `app/build.gradle` 里通过 `rootProject.ext.releaseSign` 引用，debug 和 release 的 `signingConfig` 都指向同一套。

---

## Godot 导出配置（Android）

在 Godot 编辑器 → Editor Settings → Export → Android 里填入：

| 设置项 | 值 |
|---|---|
| Android SDK Path | `C:\data\androidsdk` |
| Debug Keystore | `AIDOC/project_doc/nova_arcade/kb/app_new.jks`（本工作区绝对路径） |
| Debug Keystore User | `key0` |
| Debug Keystore Password | `111111` |
| Release Keystore | 同 Debug Keystore（本工作区 `AIDOC/project_doc/nova_arcade/kb/app_new.jks` 绝对路径） |
| Release Keystore User | `key0` |
| Release Keystore Password | `111111` |

配置完成后命令行导出才能生效：

```powershell
$G = 'C:\data\developer\devtool\godot\godot4.5\Godot_v4.5-stable_win64_console.exe'
& $G --headless --path <工程目录> --export-release Android <输出.apk路径>
```

---

## phoenixui 工程参考

**路径**：`C:\data\developer\projects\Mame\mame0288\Retro\RetroArch\pkg\android\phoenixui`

已上线的 Android 街机盒子工程，NOVA ARCADE 开发时可用作以下方面的代码参考：

| 模块 | 路径 | 参考内容 |
|---|---|---|
| 主 UI 壳 | `baseuilib/` | `GameListActivity`（游戏列表）、金币系统、Tab 结构 |
| 业务逻辑 | `app/src/` | `GameConfig.java`（每游戏按钮/控制器配置）、`FlavorFactory`（渠道反射）|
| MAME JNI 胶水 | `MAME4droid/.../jni/mame4droid-jni.c` | `load_lib()`（dlopen+dlsym myosd）、视频/音频回调 |
| 内容包解压 | `Utils.copyGameFiles` | 版本标记、已存在文件跳过逻辑 |
| 外部 ROM 退出 | `Emulator.emulateRom` | `Process.killProcess` 语义（对应 NOVA ArcadeRunner 的进程重启路径）|

**Gradle 构建版本**（`build.gradle`）：

```
AGP          : 8.1.0
compileSdk   : 35
ndkVersion   : 22.1.7171670
minSdk       : 21
targetSdk    : 28
ABI          : armeabi-v7a, arm64-v8a
```

**Flavor 体系**：每个 flavor 对应一个独立产品包（mjguoguan1/mjarch3/gamearch 等），通过 `applicationId` 区分，共用同一套签名。
NOVA ARCADE 可参考此体系做渠道包（国内各应用商店不同包名）。
