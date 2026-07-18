# Tasks: 游戏管理全功能插件化

## 目标

将 Host 中 `app/game/` 的所有功能迁移到 `plugins/game/` 插件中，删除现有 `plugins/dlc/`（合并进 game 插件）。

Host 只保留底座能力：用户/角色/菜单/字典/租户/定时任务 + 插件框架。

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5 → T6
```

---

- [ ] 1. 创建 game 插件工程骨架
  - 新增 `plugins/game/`：go.mod、main.go、plugin.json
  - Register() 返回完整菜单树（游戏列表/DLC/玩家/订单/支付/H5）
  - go build 通过

- [ ] 2. 迁移后端业务逻辑到 game 插件
  - 从 Host `app/game/` 迁移：models、service、dto
  - 实现 handler.go 路由分发（按 path 前缀分发到各子模块）
  - 表名加 `game_` 前缀（保持原有表名不变，因为已经是 game_ 开头）
  - go build 通过

- [ ] 3. 从 Host 中删除 game 模块
  - 删除 `app/game/` 整个目录
  - 删除 `cmd/api/game.go`
  - 修改 setup.go 移除 game models 迁移
  - 修改迁移脚本移除 game 相关
  - Host go build ./... 通过

- [ ] 4. 创建 game 插件前端
  - 从 Host `src/views/game/` 迁移页面到 `plugins/game/frontend/`
  - 构建为 ES module bundle
  - 导出 routes 和 menus

- [ ] 5. 清理 Host 前端
  - 删除 `src/views/game/` 目录
  - 删除旧的 DLC 插件相关文件
  - menu_init.sql 只保留扩展功能 + 插件管理

- [ ] 6. 删除旧 DLC 插件 + 集成验证
  - 删除 `plugins/dlc/` 目录
  - 重新打包 game 插件（后端 + 前端）
  - 全量编译验证
  - 端到端测试
