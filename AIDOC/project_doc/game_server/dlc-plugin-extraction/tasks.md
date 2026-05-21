# Tasks: DLC 插件分离

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5
```

---

- [ ] 1. 创建 DLC 插件工程骨架
  - **复杂度**: 中
  - **Scope:** 新增 `plugins/dlc/` 目录：go.mod、main.go、plugin.json；不触碰 Host 代码
  - **Acceptance:**
  - AC: go.mod 定义 module game-server/plugins/dlc，require plugin-sdk
  - AC: main.go 实现 PluginService 接口（Register 返回正确的 PluginInfo）
  - AC: plugin.json 包含 name/version/description
  - AC: go build . 零错误
  - _Requirements: FR-1_

- [ ] 2. 实现 DLC 插件业务逻辑
  - **复杂度**: 高
  - **Scope:** 新增 `plugins/dlc/` 下的 model.go、dto.go、service.go、handler.go；不触碰 Host 代码
  - **Constraints:** 表名使用 `dlc_game_dlc`；查询必须带 tenant_id 过滤；使用 HttpRequest 中的 db_dsn 初始化 GORM；文件存储使用本地目录
  - **Acceptance:**
  - AC: handler.go 根据 method+path 正确路由到对应处理函数
  - AC: service.go 实现 GetPage/Get/Insert/Update/Delete/UploadPck/Download/Stats
  - AC: model.go 定义 DlcGameDlc 结构体，TableName 返回 "dlc_game_dlc"
  - AC: 删除保护：有订单时拒绝删除
  - AC: 上架校验：无文件时拒绝上架
  - AC: go build . 零错误
  - _Requirements: FR-1, FR-2, FR-3, FR-4_

- [ ] 3. 实现插件数据库自动迁移
  - **复杂度**: 中
  - **Scope:** 修改 `plugins/dlc/main.go`，在 Register 或首次 HandleRequest 时执行 AutoMigrate
  - **Constraints:** 使用 HttpRequest.DbDsn 连接数据库；连接池复用（只初始化一次）
  - **Acceptance:**
  - AC: 插件首次处理请求时自动创建 dlc_game_dlc 表
  - AC: 数据库连接只初始化一次（sync.Once）
  - AC: go build . 零错误
  - _Requirements: FR-4_

- [ ] 4. 从 Host 中删除 DLC 相关代码
  - **复杂度**: 中
  - **Scope:** 删除 `backend/app/game/` 下的 DLC 文件；修改 router.go 和 setup.go
  - **Constraints:** 不触碰非 DLC 相关代码；确保其他路由注册不受影响
  - **Acceptance:**
  - AC: 删除 apis/dlc.go、service/dlc.go、models/dlc.go
  - AC: 从 service/dto/dlc.go 中删除 DLC 相关 DTO（如果是独立文件则整文件删除）
  - AC: router.go init() 中移除 registerDlcRouter
  - AC: dlc.go 中的 registerDlcRouter 函数删除（注意 withTenant 辅助函数被其他路由使用则保留）
  - AC: setup.go runMigrations 中移除 game_dlc 相关迁移（如果有）
  - AC: go build ./... 零错误
  - AC: 【回归】RG-1~RG-6 不受影响
  - _Requirements: FR-5_

- [ ] 5. 集成验证
  - **复杂度**: 中
  - **Scope:** 编写集成测试验证 DLC 插件通过 PluginManager 注册/启动/处理请求的完整流程
  - **Acceptance:**
  - AC: 测试覆盖：注册 → 启动 → CRUD 请求代理 → 停止
  - AC: Host go build ./... 零错误
  - AC: 插件 go build . 零错误
  - AC: 集成测试通过
  - _Requirements: FR-1, FR-2, FR-3_
