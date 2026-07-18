# 设计：DLC 管理功能分离为独立插件

## 技术方案

### 插件项目结构

```
projects/demo/platform_admin/plugins/dlc/
├── go.mod              # module platform-admin/plugins/dlc
├── main.go             # 插件入口，实现 PluginService 接口
├── handler.go          # HTTP 请求路由分发
├── service.go          # 业务逻辑（从 Host 迁移）
├── model.go            # 数据模型（game_dlc 表，加 dlc_ 前缀）
├── dto.go              # 请求/响应 DTO
├── plugin.json         # 插件描述文件
└── frontend/           # 前端（复用现有 DLC 管理页面，后续迁移）
    └── src/index.ts
```

### API 设计

插件内部路由（通过 HandleRequest 的 req.Path 分发）：

| 方法 | 插件内路径 | 对外路径（经代理） | 功能 |
|------|-----------|-------------------|------|
| GET | /list | /api/v1/plugin/dlc/list | DLC 分页列表 |
| GET | /:id | /api/v1/plugin/dlc/:id | DLC 详情 |
| POST | / | /api/v1/plugin/dlc | 创建 DLC |
| PUT | /:id | /api/v1/plugin/dlc/:id | 更新 DLC |
| DELETE | /:id | /api/v1/plugin/dlc/:id | 删除 DLC |
| POST | /:id/upload | /api/v1/plugin/dlc/:id/upload | 上传 PCK |
| GET | /:id/download | /api/v1/plugin/dlc/:id/download | 下载 URL |
| GET | /stats | /api/v1/plugin/dlc/stats | 统计 |

### 数据库设计

插件使用 Host 传入的 db_dsn 连接主库，表名使用 `dlc_` 前缀：

```sql
-- 插件自动迁移创建（原 game_dlc 改名为 dlc_game_dlc）
CREATE TABLE dlc_game_dlc (
  id INT AUTO_INCREMENT PRIMARY KEY,
  game_id INT NOT NULL COMMENT '所属游戏ID',
  dlc_key VARCHAR(64) NOT NULL COMMENT 'DLC标识',
  name VARCHAR(128) NOT NULL COMMENT 'DLC名称',
  version VARCHAR(32) DEFAULT '1.0.0' COMMENT '版本号',
  description VARCHAR(512) COMMENT '描述',
  file_path VARCHAR(256) COMMENT 'PCK文件路径',
  file_size BIGINT DEFAULT 0 COMMENT '文件大小字节',
  sha256 VARCHAR(64) COMMENT 'SHA256哈希',
  price INT DEFAULT 0 COMMENT '价格分',
  is_free INT DEFAULT 1 COMMENT '是否免费 1是 2否',
  status INT DEFAULT 1 COMMENT '状态 1上架 2下架',
  download_count INT DEFAULT 0 COMMENT '下载次数',
  min_game_version VARCHAR(32) DEFAULT '0.0.0' COMMENT '最低游戏版本',
  tenant_id INT DEFAULT 0 COMMENT '租户ID',
  create_by INT DEFAULT 0,
  update_by INT DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 核心逻辑

#### 插件启动流程
```
1. Host 调用 plugin.Register() → 返回 PluginInfo（name=dlc, routePrefix=dlc）
2. Host 注册路由 /api/v1/plugin/dlc/* → 代理到插件
3. 插件内部用 db_dsn 初始化 GORM 连接，执行 AutoMigrate
```

#### 请求处理流程
```
Client → Host Proxy → HandleRequest(HttpRequest) → handler.go 路由分发 → service.go 业务逻辑 → 返回 HttpResponse
```

#### 数据迁移策略
- 首次部署：插件 AutoMigrate 创建 `dlc_game_dlc` 表
- 数据迁移：手动执行 `INSERT INTO dlc_game_dlc SELECT * FROM game_dlc`（一次性）
- 迁移完成后删除 Host 中的 `game_dlc` 表（可选）

### 文件变更

**新增（插件工程）：**
```
plugins/dlc/
├── go.mod, main.go, handler.go, service.go, model.go, dto.go, plugin.json
└── frontend/src/index.ts
```

**删除（Host 清理）：**
```
backend/app/game/apis/dlc.go
backend/app/game/service/dlc.go
backend/app/game/service/dto/dlc.go  （DLC 相关部分）
backend/app/game/models/dlc.go
backend/app/game/router/dlc.go       （registerDlcRouter 函数）
```

**修改（Host）：**
```
backend/app/game/router/router.go    （从 init() 中移除 registerDlcRouter）
backend/app/setup/setup.go           （从 runMigrations 中移除 game_dlc 迁移）
```

## 不变行为清单（回归防护）

| 编号 | 不变行为 | 验证方式 |
|------|----------|----------|
| RG-1 | 游戏管理 CRUD 不受影响 | /api/v1/game 接口正常 |
| RG-2 | 玩家管理不受影响 | /api/v1/game-player 接口正常 |
| RG-3 | 订单管理不受影响 | /api/v1/game-order 接口正常 |
| RG-4 | 支付配置不受影响 | /api/v1/game-payment-config 接口正常 |
| RG-5 | H5 页面管理不受影响 | /api/v1/game-h5 接口正常 |
| RG-6 | Host 编译通过 | go build ./... 零错误 |

## 正确性属性

- 插件通过 db_dsn 获得完整数据库访问权限，可查询 game_order 表
- 租户隔离通过 HttpRequest 注入的 tenant_id 实现，插件查询必须带 tenant_id 过滤
- 文件存储路径与原逻辑一致，已上传的文件无需迁移
