# DLC Pack Builder

## 用途

此 Godot 工程用于制作 DLC 内容并导出为 `.pck` 文件。

## 目录结构

```
dlc_pack/
├── dlc/
│   └── poison_dlc/          # DLC 内容（与主游戏 res://dlc/poison_dlc/ 路径对应）
│       ├── manifest.json
│       ├── components/
│       │   └── poison_component.gd
│       └── systems/
│           └── poison_system.gd
├── exported/                # 导出的 .pck 文件（运行后生成）
├── scenes/
│   └── export_helper.tscn   # 导出辅助场景
└── scripts/
    └── export_helper.gd     # 导出逻辑
```

## 使用步骤

1. 用 Godot 4.5 打开此工程
2. 运行主场景（F5）
3. 点击「导出 DLC PCK」按钮
4. 记录输出的 SHA-256 哈希值
5. 将 `exported/poison_dlc.pck` 复制到 `../dlc_server/dlc_files/`
6. 将 SHA-256 填入 `../dlc_server/config.json`

## 注意事项

- DLC 脚本中引用的类（如 `EcsSystem`、`RuntimeStatsComponent`）
  在主游戏运行时才存在，DLC 工程本身不需要这些类
- PCK 内的文件路径必须与主游戏中的 `res://dlc/poison_dlc/` 路径一致
