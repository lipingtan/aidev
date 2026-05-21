# 需求计划：DLC 管理功能分离为独立插件

## 目标

将现有 Host 中的 DLC 管理功能（CRUD + 文件上传/下载 + 统计）从 `app/game/` 模块中抽离，改造为基于插件架构的独立插件。

## 当前功能清单

| 接口 | 方法 | 路径 | 功能 |
|------|------|------|------|
| GetPage | GET | /api/v1/game-dlc | DLC 分页列表 |
| Get | GET | /api/v1/game-dlc/:id | DLC 详情 |
| Insert | POST | /api/v1/game-dlc | 创建 DLC |
| Update | PUT | /api/v1/game-dlc/:id | 更新 DLC |
| Delete | DELETE | /api/v1/game-dlc/:id | 删除 DLC |
| UploadPck | POST | /api/v1/game-dlc/:id/upload | 上传 PCK 文件 |
| Download | GET | /api/v1/game-dlc/:id/download | 获取下载 URL |
| Stats | GET | /api/v1/game-dlc/stats | DLC 统计 |

## 澄清问题

[Question] 分离后，DLC 的 API 路径是否变更？
- 方案 A：保持 `/api/v1/game-dlc/*`（通过插件代理 `/api/v1/plugin/dlc/*` 内部映射）
- 方案 B：迁移到 `/api/v1/plugin/dlc/*`（前端同步修改）

[Answer]
方案B
[Question] DLC 表 `game_dlc` 是否保留在主库中，还是由插件自行管理（插件启动时 AutoMigrate）？

[Answer]
插件启动时 AutoMigrate
[Question] DLC 功能依赖 `game_order` 表做统计和删除保护，分离后如何处理跨表查询？
- 方案 A：插件直接连主库，通过 db_dsn 访问所有表
- 方案 B：通过 Host 事件/RPC 获取订单数据

[Answer]
方案A
[Question] 分离后是否从 Host 的 `app/game/` 中删除 DLC 相关代码，还是保留但标记废弃？

[Answer]
删除