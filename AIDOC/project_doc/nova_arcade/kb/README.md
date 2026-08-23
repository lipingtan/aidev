# KB — 知识库索引

项目开发过程中积累的工具链、环境、约定等参考文档。
与设计文档（`../nova-arcade-*.md`）互补——设计文档描述"做什么"，KB 描述"怎么做/用什么做"。

---

## 文件列表

| 文件 | 内容 | 更新时间 |
|---|---|---|
| [godot-toolchain.md](./godot-toolchain.md) | Godot 4.5 可执行路径、常用命令、工程目录、PoC 运行示例 | 2026-08-21 |
| [android-sdk.md](./android-sdk.md) | Android SDK/NDK/JDK 路径与版本、keystore 签名配置、Godot 导出设置、phoenixui 工程参考 | 2026-08-21 |
| [visual-verify.md](./visual-verify.md) | 可视化验证：`read_image` 看资产 / GUI+截屏看运行画面 / 无头自检（`MT_SELFTEST`） | 2026-08-21 |

---

## 待补充

以下内容计划后续整理进 KB：

- 构建与导出流程（Shell APK + DLC PCK 完整打包步骤）
- Git 分支与发布约定
- 代码规范（GDScript 命名、文件组织）
- 测试与 PoC 验证流程（可视化/看画面部分见 [visual-verify.md](./visual-verify.md)）
- MAME 核心（`libMAME4droid.so` 按 ABI 分发）与 ROM 下载/校验流程说明（以 `projects/nova_arcade/mame-godot-plugin/DESIGN.md` v2 为准，基于 mame0288 源码）
