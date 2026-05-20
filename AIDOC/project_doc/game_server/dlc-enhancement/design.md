# 设计：DLC 管理功能完善

## 技术方案

### API 设计

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/v1/game-dlc/:id/upload | 上传 PCK 文件 | JWT + 租户 |
| GET | /api/v1/game-dlc/:id/download | 获取下载 URL | JWT + 租户 |
| GET | /api/v1/game-dlc/stats | DLC 统计数据 | JWT + 租户 |

### 数据库设计

无新增表，无 DDL 变更。统计数据从 `game_order` 表按日聚合查询。

### 核心逻辑

#### FileStore 接口（`common/storage/`）

```go
// common/storage/file_store.go
type FileStore interface {
    Save(category string, filename string, reader io.Reader) (path string, err error)
    GetURL(path string) (url string, err error)  // 本地返回相对路径，OSS返回签名URL
    Delete(path string) error
    Exists(path string) bool
}

// common/storage/local_store.go
type LocalStore struct {
    BaseDir string // 默认 "static/uploadfile"
}
func (s *LocalStore) GetURL(path string) (string, error) {
    return "/static/uploadfile/" + path, nil  // 本地直接返回静态路径
}
```

#### 上传流程
1. 前端 el-upload 选择 .pck 文件 → 校验 ≤ 500MB
2. POST multipart 到 `/api/v1/game-dlc/:id/upload`
3. 后端调用 `FileStore.Save("dlc", filename, file)` 存储
4. 计算 SHA256，更新 DLC 记录的 filePath/fileSize/sha256

#### 下载流程
1. GET `/api/v1/game-dlc/:id/download`
2. 后端调用 `FileStore.GetURL(dlc.FilePath)` 获取 URL
3. 返回 `{ "url": "..." }`，前端用 `window.open(url)` 下载
4. 后续切换 OSS 时，GetURL 返回签名临时 URL，前端逻辑不变

#### 删除保护
```
service.Remove():
  1. 查询 game_order WHERE product_type='dlc' AND product_id=dlcId
  2. count > 0 → 返回错误"该 DLC 已有玩家购买，无法删除"
  3. count == 0 → 执行软删除
```

#### 上架校验
```
service.Update():
  IF req.Status == 1 AND model.FilePath == "" THEN
    返回错误"请先上传 PCK 文件后再上架"
```

#### 统计接口
```sql
-- 总下载/总收入
SELECT COUNT(*) as total_downloads, SUM(amount) as total_revenue
FROM game_order WHERE product_type='dlc' AND status='paid'

-- 近30天趋势
SELECT DATE(paid_at) as date, COUNT(*) as count, SUM(amount) as revenue
FROM game_order WHERE product_type='dlc' AND status='paid'
  AND paid_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(paid_at) ORDER BY date

-- DLC排行
SELECT product_id, product_name, COUNT(*) as downloads, SUM(amount) as revenue
FROM game_order WHERE product_type='dlc' AND status='paid'
GROUP BY product_id, product_name ORDER BY revenue DESC LIMIT 10
```

### 文件结构变更

```
backend/
├── common/storage/
│   ├── file_store.go         # 新增：FileStore 接口
│   └── local_store.go        # 新增：本地磁盘实现
├── app/game/apis/dlc.go      # 修改：新增 UploadPck/Download/Stats 方法
├── app/game/service/dlc.go   # 修改：重构上传逻辑、新增删除保护/上架校验/统计
├── app/game/router/dlc.go    # 修改：注册新路由

frontend/
├── src/api/game.ts           # 修改：新增 uploadDlcPck/getDlcDownloadUrl/getDlcStats
└── src/views/game/dlc.vue    # 修改：游戏下拉、上传组件、下载按钮、免费联动、统计Tab
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | DLC 列表/新增/编辑基础功能不受影响 | CRUD 操作正常 |
| RG-2 | 游戏管理页功能不受影响 | 游戏列表正常 |
| RG-3 | 现有 API 签名不变 | getDlcList/createDlc/updateDlc 调用无报错 |
| RG-4 | 订单数据不受影响 | game_order 表只读查询，不做写操作 |

## 正确性属性

- FileStore 接口与实现分离，替换 OSS 只需新增实现类，不改业务代码
- 上传必须计算 SHA256 并持久化到 DLC 记录
- 删除保护查询 game_order 表 product_type='dlc' 且 product_id 匹配
- 上架校验在后端 service 层强制执行，前端校验仅为辅助
- 下载通过 GetURL 抽象，本地返回静态路径，OSS 返回签名 URL
- 统计金额以"分"存储，前端负责 ÷100 显示
