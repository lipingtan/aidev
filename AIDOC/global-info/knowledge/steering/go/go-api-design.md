# Go API 设计规范

## URL 设计

### 路径规范

| 规则 | 示例 |
|------|------|
| 使用复数名词表示资源集合 | `/api/v1/games`、`/api/v1/users` |
| 使用 kebab-case | `/api/v1/game-configs`（非 `gameConfigs`） |
| 嵌套资源最多两层 | `/api/v1/games/:id/dlcs`（不再嵌套） |
| 版本号放在路径中 | `/api/v1/`、`/api/v2/` |
| 动作用动词（非 CRUD 操作） | `POST /api/v1/games/:id/publish` |

### HTTP 方法选择

| 方法 | 用途 | 幂等 | 示例 |
|------|------|:---:|------|
| GET | 查询（列表/详情） | ✅ | `GET /api/v1/games` |
| POST | 创建资源 | ❌ | `POST /api/v1/games` |
| PUT | 全量更新 | ✅ | `PUT /api/v1/games/:id` |
| PATCH | 部分更新 | ✅ | `PATCH /api/v1/games/:id` |
| DELETE | 删除资源 | ✅ | `DELETE /api/v1/games/:id` |

---

## 分页、排序、过滤

### 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `pageIndex` | int | 1 | 页码（从1开始） |
| `pageSize` | int | 10 | 每页条数（最大100） |

### 排序参数

```
GET /api/v1/games?orderBy=created_at&orderDir=desc
```

| 参数 | 说明 |
|------|------|
| `orderBy` | 排序字段（必须是允许的字段） |
| `orderDir` | `asc` / `desc` |

### 过滤参数

```
GET /api/v1/games?name=xxx&status=1
```

- 精确匹配：`status=1`
- 模糊搜索：`name=xxx`（Service 层加 `LIKE`）
- 范围查询：`createdAtStart=2026-01-01&createdAtEnd=2026-12-31`

---

## 响应格式

### 成功响应

**列表：**
```json
{
  "code": 200,
  "msg": "ok",
  "data": {
    "list": [...],
    "count": 100,
    "pageIndex": 1,
    "pageSize": 10
  }
}
```

**单条：**
```json
{
  "code": 200,
  "msg": "ok",
  "data": { ... }
}
```

**操作成功（无返回数据）：**
```json
{
  "code": 200,
  "msg": "ok"
}
```

### 错误响应

```json
{
  "code": 400,
  "msg": "参数错误: name 不能为空"
}
```

### 状态码使用

| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| 200 | 成功 | 所有成功操作 |
| 400 | 参数错误 | 请求参数校验失败 |
| 401 | 未认证 | JWT 缺失或过期 |
| 403 | 无权限 | Casbin 权限拒绝 |
| 404 | 资源不存在 | 查询的 ID 不存在 |
| 500 | 服务器错误 | 未预期的内部错误 |

---

## DTO 设计规范

### 命名规则

| 类型 | 命名格式 | 示例 |
|------|----------|------|
| 列表请求 | `{Resource}GetPageReq` | `GameGetPageReq` |
| 详情请求 | `{Resource}GetReq` | `GameGetReq` |
| 创建请求 | `{Resource}InsertReq` | `GameInsertReq` |
| 更新请求 | `{Resource}UpdateReq` | `GameUpdateReq` |
| 删除请求 | `{Resource}DeleteReq` | `GameDeleteReq` |

### 字段规则

```go
type GameInsertReq struct {
    Name      string `json:"name" binding:"required,min=1,max=128"`
    AppKey    string `json:"appKey" binding:"required,alphanum,len=32"`
    AppSecret string `json:"appSecret" binding:"omitempty"`  // 可选，不传则自动生成
    Status    int    `json:"status" binding:"oneof=1 2"`
}
```

- 必填字段加 `binding:"required"`
- 字符串限制长度 `min=X,max=Y`
- 枚举值用 `oneof=`
- 可选字段用 `omitempty`
- 不从请求接收的字段（如 tenant_id）不放在 DTO 中

---

## 接口文档

使用 Swagger 注解自动生成：

```go
// GetPage
// @Summary 获取游戏列表
// @Tags 游戏管理
// @Accept json
// @Produce json
// @Param pageIndex query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param name query string false "游戏名称（模糊搜索）"
// @Success 200 {object} response.Page{list=[]models.Game}
// @Router /api/v1/games [get]
// @Security Bearer
func (e Game) GetPage(c *gin.Context) {
```

---

## 版本策略

- 当前版本：`/api/v1/`
- 破坏性变更时新增版本：`/api/v2/`
- 旧版本保留至少 3 个月
- 非破坏性变更（新增字段、新增接口）不升版本
