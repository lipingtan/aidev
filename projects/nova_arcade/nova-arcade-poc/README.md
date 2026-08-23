# NOVA ARCADE 运行时 PoC

真 Godot 4.5 无头验证三运行时的壳层编排。设计文档：`../nova-arcade-runtime-design.md`。

## 三个 PoC

### 1. PCK 运行时（全链路，含真实 PCK 导出）

```powershell
$G = 'C:\data\developer\devtool\godot\godot4.5\Godot_v4.5-stable_win64_console.exe'
$env:APPDATA = "<可写目录>"   # 沙箱下需重定向

# 导出迷你游戏为 PCK
& $G --headless --path game-mini-demo --export-pack PoC dlcs/mini_demo.pck

# 壳挂载运行（完整生命周期 + 二次启动存档持久化验证）
& $G --headless --path . -- <pck绝对路径>
```

验证：挂载(replace_files=false) → 前缀隔离加载 → boot(ctx) → 存档写 save_dir →
quit_requested → tmp 清理 → 退出码 0；第二次启动读到上次 best。

### 2. HTML 运行时（zip/HTTP/桥协议，WebView 本体除外）

```
& $G --headless --path . -s res://poc/html_poc.gd
```

验证：.novahtml zip 打包/解压往返 → 本地 HTTP 文件服务（MIME/CORS）→
路径穿越 403 / 缺失 404 → `nova://quit`/`nova://save` 桥协议解析。

### 3. Arcade 运行时（壳层编排，Mock=libretro API 形状）

```
& $G --headless --path . -s res://poc/arcade_poc.gd
```

验证：RunnerRegistry 三路分发 → ROM zip sha256+romset 成员校验（含篡改用例）→
Mock 核心逐帧 run/video/audio 回调 → 组合键退出 → savestate/nvram 落
save_dir → 跨游戏存储隔离 → tmp 清理 → 二次启动从 savestate 恢复状态。

> ⚠️ Mock 核心的调用面（`retro_load_game`/`retro_run`/`retro_input_state`/`retro_serialize`/`retro_unserialize`/`retro_get_hiscore`）按 libretro 形状设计；决策演进：libretro → myosd C ABI 直嵌（见 `../nova-arcade-runtime-design.md` §3，**临时保留待调整**）→ **最新定稿 = `../mame-godot-plugin/DESIGN.md` v2（独立 native MAME4droid Activity + AndroidRuntime 插件 `MameRuntime`，以 v2 为准）**。壳层编排（分发/校验/生命周期/存储隔离）结论有效；接真实核心时 Mock 调用面按 v2 替换为 `MameRuntime.ensureNative()/launch(rom)/gameExited`，壳层不动。

## 结构

```
project.godot          壳工程
shell/main.gd|tscn     PCK 运行时壳（挂载/boot/quit 协议）
game-mini-demo/        迷你游戏工程（导出为 dlcs/mini_demo.pck）
poc/html_runtime_driver.gd   HTML 运行时 PoC
poc/arcade_runtime_driver.gd Arcade 运行时 PoC（含 Registry/Validator/MockCore/Runner）
```

## 尚未验证（里程碑 M5）

- WebView 真机加载（gdcef/Android WebView/WKWebView 适配层）
- 真实 MAME 0.288 myosd 核心编译（BSD-3，同 mjarch3 libmain.so 构建链）+ MyOsdHost GDExtension 绑定（runtime-design §3.2）
- ROM 授权分发链路（服务端签名 rom zip 集 + 说明文件 → 客户端校验，runtime-design §3.4）
