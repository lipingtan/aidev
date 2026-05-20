# Tasks: DLC 管理功能完善

## Task Dependency Graph

```
T1 → T2 → T3 → T4 → T5 → T6
```

---

- [ ] 1. 实现 FileStore 接口和本地存储
  - 新增 `common/storage/file_store.go`：FileStore 接口（Save/GetURL/Delete/Exists）
  - 新增 `common/storage/local_store.go`：LocalStore 本地磁盘实现
  - GetURL 返回 `/static/uploadfile/` + 相对路径
  - go build ./... 零错误
  - _Requirements: FR-2_

- [ ] 2. 后端 DLC 上传/下载/统计接口
  - 修改 `app/game/apis/dlc.go`：重构 UploadPck 使用 FileStore、新增 Download 和 Stats
  - 修改 `app/game/service/dlc.go`：重构上传逻辑、新增统计方法
  - 修改 `app/game/router/dlc.go`：注册 upload/download/stats 路由
  - 上传限制 500MB，通过 FileStore 接口存储
  - 下载返回 JSON `{ "url": "..." }`
  - 统计从 game_order 聚合（product_type='dlc', status='paid'）
  - go build ./... 零错误，现有 CRUD 不受影响
  - _Requirements: FR-2, FR-3, FR-8_

- [ ] 3. 后端删除保护和上架校验
  - 修改 `app/game/service/dlc.go` 的 Remove 和 Update 方法
  - 删除时查询 game_order 关联订单，有则拒绝
  - 上架时校验 filePath 非空
  - go build ./... 零错误
  - _Requirements: FR-6, FR-7_

- [ ] 4. 前端 API 新增
  - 修改 `frontend/src/api/game.ts`
  - 新增 uploadDlcPck(id, file)、getDlcDownloadUrl(id)、getDlcStats(params)
  - 现有函数签名不变
  - _Requirements: FR-2, FR-3, FR-8_

- [ ] 5. 前端 DLC 页面改造
  - 修改 `frontend/src/views/game/dlc.vue`
  - gameId 改为 el-select 下拉选择游戏
  - 新增/编辑弹窗内嵌 el-upload 上传 PCK（accept=".pck"，限制 500MB）
  - 列表新增文件大小列和下载按钮（filePath 非空时显示）
  - isFree=1 时 price 置 0 并 disabled
  - 新增统计子 Tab（el-tabs）：总下载、总收入、折线图、排行表格
  - 现有列表/新增/编辑/删除功能不受影响
  - _Requirements: FR-1, FR-2, FR-3, FR-4, FR-5, FR-8_

- [ ] 6. 后端编译验证
  - go build ./... 零错误
  - go vet ./... 无警告
  - 后端可正常启动
  - 现有所有接口正常
