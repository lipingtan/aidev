# 技能：platform_admin 插件开发

## 触发条件

当用户消息包含以下词语时激活本技能：
"开发插件"、"写插件"、"新建插件"、"plugin开发"、"插件后端"、"插件前端"、"插件模块"、"插件管理"、"插件测试"、"插件部署"、"插件升级"

---

## 角色

你是 platform_admin 插件体系的首席开发工程师，熟悉插件 SDK、代理路由机制、前端加载架构，
以及从历次 review 和 e2e 测试中积累的全套工程实践。

---

## 插件体系核心架构（必须了解）

```
主服务（backend/）
  └─ 路由代理：/api/v1/admin/plugin/:name/*action
        └─ PluginProxy → 找到已注册并 Running 的插件实例
              └─ 通过 hashicorp/go-plugin（gRPC）调用插件进程
                    └─ 插件进程（独立二进制）HandleRequest()
                          └─ 自建路由分发 + 自建数据库连接

插件前端（frontend/）
  └─ Vue3 + Element Plus，打包为 ESM bundle（dist/index.js）
  └─ 部署到 backend/static/plugins/{name}/admin-pc/bundle.js
  └─ 主服务 /api/v1/admin/plugin-frontend/load 动态注入宿主页面
```

---

## 零、从零新建插件（完整流程）

### Step 1：创建目录结构

```powershell
$name = "your_plugin"  # 插件名，英文小写下划线
$pluginDir = "projects/demo/platform_admin/plugins/$name"

New-Item -ItemType Directory -Path "$pluginDir/frontend/src/views" -Force

# 创建后端骨架文件
@("main.go","handler.go","models.go","db.go","go.mod") | ForEach-Object {
    New-Item -ItemType File -Path "$pluginDir/$_" -Force
}
# 创建 plugin.json
New-Item -ItemType File -Path "$pluginDir/plugin.json" -Force
```

### Step 2：填写 plugin.json

```json
{
  "name": "{name}",
  "version": "1.0.0",
  "displayName": "插件显示名称",
  "description": "插件功能描述",
  "routePrefix": "/api/v1/plugin/{name}",
  "platforms": ["admin:pc"],
  "modules": [
    {"code": "mod1", "name": "模块一"}
  ],
  "frontends": [
    {"platform": "admin", "device": "pc", "entry": "dist/index.js"}
  ],
  "menus": [
    {
      "type": "menu", "name": "插件名称", "permissionCode": "{name}:menu",
      "path": "/plugin/{name}", "icon": "Monitor", "platform": "admin",
      "moduleCode": "mod1", "sort": 10, "component": "plugin/container",
      "children": []
    }
  ],
  "apiPermissions": [
    {
      "type": "GROUP", "name": "模块一", "moduleCode": "mod1",
      "children": [
        {
          "type": "ENDPOINT", "name": "列表", "displayName": "获取列表",
          "permissionCode": "mod1:list",
          "urlPattern": "/api/v1/plugin/{name}/mod1",
          "httpMethod": "GET", "moduleCode": "mod1"
        }
      ]
    }
  ],
  "exposedActions": [],
  "subscribedEvents": [],
  "breakingUpgrade": false
}
```

### Step 3：实现 main.go、handler.go、db.go

参照第一章规范。以下是最小可运行的 `main.go` 骨架：

```go
package main

import (
    "context"
    "platform-admin/plugin-sdk/helper"
    "platform-admin/plugin-sdk/proto"
)

type MyPlugin struct{}

func (p *MyPlugin) Register(ctx context.Context) (*proto.PluginInfo, error) {
    return &proto.PluginInfo{
        Name:        "{name}",
        Version:     "1.0.0",
        Description: "插件描述",
        RoutePrefix: "{name}",  // 短前缀，不加 /api/v1/plugin/
        Menus:       []*proto.MenuItem{},
        Perms:       []*proto.Permission{},
    }, nil
}

func (p *MyPlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    return handleRequest(ctx, req)
}

func (p *MyPlugin) Healthcheck(ctx context.Context) (*proto.HealthResponse, error) {
    return &proto.HealthResponse{Healthy: true, Message: "ok"}, nil
}

func (p *MyPlugin) OnEvent(ctx context.Context, event *proto.Event) error { return nil }

func (p *MyPlugin) CallPlugin(ctx context.Context, req *proto.CallPluginRequest) (*proto.CallPluginResponse, error) {
    return &proto.CallPluginResponse{Payload: []byte(`{}`)}, nil
}

func main() {
    helper.Serve(&MyPlugin{})
}
```

### Step 4：初始化 go.mod

```powershell
cd "projects/demo/platform_admin/plugins/{name}"
# ⚠️ 模块名按实际项目命名，不要照抄 game-server
go mod init platform-admin/plugins/{name}
go get gorm.io/driver/mysql@v1.5.7
go get gorm.io/gorm@v1.25.12
go get platform-admin/plugin-sdk@v0.0.0  # 通过 replace 指令指向本地
```

在 `go.mod` 末尾加：
```
replace platform-admin/plugin-sdk => ../../plugin-sdk
```

### Step 5：编译二进制到主服务路径

```powershell
# 确保主服务插件目录存在
New-Item -ItemType Directory -Path "projects/demo/platform_admin/backend/plugins/{name}" -Force

cd "projects/demo/platform_admin/plugins/{name}"
go build -o "../../backend/plugins/{name}/{name}.exe" .
```

### Step 6：将插件注册到主服务

> ⚠️ 以下命令在 `projects/demo/platform_admin/` 根目录执行。

**方式一：通过管理 API 上传安装（推荐，适合生产）**

先将插件文件汇总到临时目录后打包（避免路径层级混乱）：
```powershell
# 在 projects/demo/platform_admin/ 根目录执行
$tmp = "dist/_pkg_{name}"
New-Item -ItemType Directory -Path $tmp -Force
Copy-Item "backend/plugins/{name}/{name}.exe"  $tmp -Force
Copy-Item "plugins/{name}/plugin.json"          $tmp -Force
# 如有前端产物也一并复制
# Copy-Item "plugins/{name}/frontend/dist/*" $tmp -Recurse -Force
Compress-Archive -Path "$tmp/*" -DestinationPath "dist/{name}-1.0.0.zip" -Force
Remove-Item $tmp -Recurse -Force
```

然后通过 API 安装：
```powershell
$token = ...  # 登录获取 token（参见第四章 loginAdmin）
Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins/upload" `
    -Method POST `
    -Headers @{Authorization="Bearer $token"} `
    -Form @{file=Get-Item "dist/{name}-1.0.0.zip"} `
    -UseBasicParsing
```

**方式二：直接放目录后重启（适合本地开发）**

```powershell
# 先获取 platform_token（插件管理接口用）
$loginBody2 = (Invoke-WebRequest -Uri "http://localhost:8000/auth/login" `
    -Method POST -ContentType "application/json" `
    -Body '{"username":"admin","password":"admin123"}' `
    -UseBasicParsing).Content | ConvertFrom-Json
$token = $loginBody2.data.token
```

1. 确保 `backend/plugins/{name}/plugin.json` 和 `{name}.exe` 都在
2. 重启主服务（主服务启动时自动扫描 `plugins/` 目录并注册）
3. 或调用 API 手动触发：
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins/{name}/start" `
    -Method POST -Headers @{Authorization="Bearer $token"} -UseBasicParsing
```

### Step 7：验证插件运行状态

```powershell
$resp = Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins" `
    -Headers @{Authorization="Bearer $token"} -UseBasicParsing
$resp.Content  # 应看到 "runStatus": "running"
```

### Step 8：创建前端子工程

```powershell
cd "projects/demo/platform_admin/plugins/{name}/frontend"
pnpm init
pnpm add -D vite @vitejs/plugin-vue typescript
pnpm add vue element-plus
```

创建 `src/index.ts`（路由入口）、各 Vue 页面，参照第二章规范开发，最后 `pnpm build` 并复制产物。

### Step 9（可选）：升级已有插件

插件版本迭代时（修改了代码并更新 `version` 字段），使用升级接口而非重新安装：

```powershell
# 1. 更新 plugin.json 中的 version 字段（如 1.0.0 → 1.1.0）
# 2. 重新编译二进制到主服务路径（参见 Step 5）
# 3. 打包新版 zip（用临时目录方式，与 Step 6 保持一致）
$name = "your_plugin"
# 在 projects/demo/platform_admin/ 根目录执行
$tmp = "dist/_pkg_${name}_upgrade"
New-Item -ItemType Directory -Path $tmp -Force
Copy-Item "backend/plugins/$name/$name.exe" $tmp -Force
Copy-Item "plugins/$name/plugin.json"        $tmp -Force
Compress-Archive -Path "$tmp/*" -DestinationPath "dist/$name-1.1.0.zip" -Force
Remove-Item $tmp -Recurse -Force

# 4. 调用升级接口（主服务会自动 stop→替换二进制→start）
$token = ...
Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins/$name/upgrade" `
    -Method PUT `
    -Headers @{Authorization="Bearer $token"} `
    -Form @{file=Get-Item "dist/$name-1.1.0.zip"} `
    -UseBasicParsing
```

**升级 vs 重装的区别：**

| 操作 | 适用场景 | 数据影响 |
|------|---------|---------|
| `upgrade`（PUT） | 代码变更、功能迭代 | 保留已有数据，AutoMigrate 增量建表 |
| 重新 `upload`（POST） | 插件名变更、破坏性重构 | 会重新注册菜单/权限，可能影响已有配置 |

---

## 一、插件后端开发规范

### 1.1 目录结构（标准）

```
plugins/{name}/
├── main.go          # 插件入口，实现 plugin-sdk 接口
├── handler.go       # 路由分发（所有 case 集中在这里）
├── models.go        # 数据模型（GORM struct）
├── db.go            # 数据库初始化（RWMutex 并发安全）
├── service_{mod}.go # 按业务模块拆分 service
├── go.mod
└── frontend/        # 前端子工程
```

### 1.2 plugin.json 与 PluginInfo 的 routePrefix 区分（重要）

两个地方都有 `routePrefix`，含义不同，**不能混淆**：

| 位置 | 字段 | 格式 | 用途 |
|------|------|------|------|
| `plugin.json` | `routePrefix` | `/api/v1/plugin/{name}` | 主服务注册 App 时的完整 API 前缀，用于 RBAC 权限校验和应用目录 |
| `proto.PluginInfo.RoutePrefix`（main.go） | `RoutePrefix` | `{name}`（短前缀） | 代理路由匹配键，主服务 PluginProxy 用这个短前缀找插件 |

```json
// plugin.json（完整路径）
{
  "routePrefix": "/api/v1/plugin/game",
  ...
}
```

```go
// main.go（短前缀，代理层自动拼完整路径）
return &proto.PluginInfo{
    RoutePrefix: "game",   // ✅ 只写短名，不写完整路径
    ...
}, nil
```

**⚠️ 常见错误**：在 `proto.PluginInfo` 里写了完整路径 `/api/v1/plugin/game`，导致代理路由匹配失败（Registry 用短前缀做 key）。

### 1.3 main.go 标准实现

> 完整骨架见 Step 3。关键字段说明：

```go
return &proto.PluginInfo{
    Name:        "name",          // 与 plugin.json 的 name 字段一致
    Version:     "1.0.0",         // 与 plugin.json 的 version 字段一致
    RoutePrefix: "name",          // ✅ 只写短前缀，代理层自动拼 /api/v1/admin/plugin/
    Menus:       []*proto.MenuItem{...},
    Perms:       []*proto.Permission{...},
}, nil
```

`HandleRequest` 直接委托 handler.go：

```go
func (p *MyPlugin) HandleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    return handleRequest(ctx, req)
}
```

### 1.4 handler.go 路由分发（强制模式）

**反例（手写字符串匹配，已知问题来源）：**
```go
// ❌ 错误：容易漏路由、无法处理复合路径
case strings.HasPrefix(path, "payment"):
    sub := strings.TrimPrefix(path, "payment/")
    return routePayment(method, sub, req)
```

**正例（标准路由表模式）：**
```go
func handleRequest(ctx context.Context, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    if req.DbDsn != "" {
        if err := initDB(req.DbDsn); err != nil {
            return jsonResp(500, "数据库初始化失败: "+err.Error(), nil)
        }
    }
    if db == nil {
        return jsonResp(503, "数据库未初始化", nil)
    }

    path := strings.TrimPrefix(req.Path, "/")  // ✅ 一层即可，req.Path 已是去前缀后的 action
    method := strings.ToUpper(req.Method)

    switch {
    // 子模块路由（单层 TrimPrefix 提取子路径）
    case strings.HasPrefix(path, "game"):
        return routeGame(method, strings.TrimPrefix(path[len("game"):], "/"), req)
    case strings.HasPrefix(path, "payment"):
        return routePayment(method, strings.TrimPrefix(path[len("payment"):], "/"), req)
    default:
        return jsonResp(404, "接口不存在: "+method+" /"+path, nil)
    }
}

func routeGame(method, sub string, req *proto.HttpRequest) (*proto.HttpResponse, error) {
    switch {
    case method == "GET" && sub == "":               return handleGameList(req)
    case method == "GET" && isNumeric(sub):          return handleGameGet(req, toInt(sub))
    case method == "POST" && sub == "":              return handleGameCreate(req)
    case method == "PUT" && isNumeric(sub):          return handleGameUpdate(req, toInt(sub))
    case method == "DELETE" && isNumeric(sub):       return handleGameDelete(req, toInt(sub))
    // ✅ 复合动作路由（如 regen-secret）必须显式声明
    case method == "POST" && hasNumericPrefix(sub) && strings.HasSuffix(sub, "/regen-secret"):
        return handleGameRegenSecret(req, toInt(sub))
    default:
        return jsonResp(404, "game接口不存在: "+method+" /game/"+sub, nil)
    }
}
```

**强制检查清单（写完 handler.go 必须逐条验证）：**
- [ ] 每个在 `plugin.json` `apiPermissions` 中声明的接口，`handler.go` 里都有对应 case
- [ ] 编辑（PUT）、重置（POST action）等操作路由与 plugin.json 声明一致
- [ ] 不存在的路径返回 404，不要 panic

### 1.5 db.go 数据库单例（强制模式）

```go
var (
    db    *gorm.DB
    dbDSN string
    dbMu  sync.RWMutex  // 保护 db 和 dbDSN 的并发读写
)

// initDB 初始化数据库连接，支持 DSN 变更重连（多租户场景）
func initDB(dsn string) error {
    // 快路径：同 DSN 且已初始化，直接复用（读锁）
    dbMu.RLock()
    if db != nil && dsn == dbDSN {
        dbMu.RUnlock()
        return nil
    }
    dbMu.RUnlock()

    // 慢路径：需要初始化或重连（写锁）
    dbMu.Lock()
    defer dbMu.Unlock()

    // double-check：防止锁等待期间其他 goroutine 已初始化
    if db != nil && dsn == dbDSN {
        return nil
    }

    newDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("数据库连接失败: %w", err)
    }
    // ⚠️ 将 &Model1{}, &Model2{} 替换为插件实际定义的所有 GORM 模型
    // 例如：db.AutoMigrate(&Game{}, &Dlc{}, &Player{}, &Order{})
    if err = newDB.AutoMigrate(&Model1{}, &Model2{}); err != nil {
        return fmt.Errorf("表迁移失败: %w", err)
    }
    db    = newDB
    dbDSN = dsn
    return nil
}
```

**与旧的 `sync.Once` 方案对比：**

| 方案 | 优点 | 缺点 |
|------|------|------|
| `sync.Once` | 简单，永远只执行一次 | DSN 变更时无法重连；首次失败后无法重试 |
| `RWMutex + double-check` | 支持 DSN 变更重连；并发安全；支持重试 | 代码稍多 |

**⚠️ 旧 `sync.Once` 的已知陷阱**：`sync.Once` 只执行一次，如果首次 DSN 有误，后续即使传入正确 DSN 也不会重连。多租户场景下 DSN 可能随请求变化，必须用 RWMutex 方案。

### 1.6 service 层已知 Bug 与修正规范

#### 分页参数：兼容 pageIndex 和 page（强制）

```go
// ✅ 正确：前端统一用 pageIndex，兼容历史 page 参数
func parsePage(params map[string]string) (page, pageSize int) {
    if v := params["pageIndex"]; v != "" {
        page, _ = strconv.Atoi(v)
    } else {
        page, _ = strconv.Atoi(params["page"])
    }
    pageSize, _ = strconv.Atoi(params["pageSize"])
    if page <= 0    { page = 1 }
    if pageSize <= 0 { pageSize = 10 }
    return
}
```

#### 分页响应字段：统一用 total（强制）

```go
// ✅ 正确：前端读 res.data?.total
type PageResponse struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total"`    // ✅ 用 total，不用 count
    Page     int         `json:"page"`
    PageSize int         `json:"pageSize"`
}
```

#### 价格/金额联动字段（数字 0 可更新问题）

```go
// ❌ 错误：price > 0 导致无法将付费 DLC 改回免费（price=0）
if body.Price > 0 { existing.Price = body.Price }

// ✅ 正确：isFree 优先，免费时强制 price=0，否则 >= 0 均可更新
if body.IsFree > 0 { existing.IsFree = body.IsFree }
if existing.IsFree == 1 {
    existing.Price = 0
} else if body.Price >= 0 {
    existing.Price = body.Price
}
```

#### 字符串非空校验（防空格绕过）

```go
import "strings"

// ❌ 错误：空格字符串可绕过
if body.Reason == "" { return jsonResp(400, "原因不能为空", nil) }

// ✅ 正确：TrimSpace 后校验
body.Reason = strings.TrimSpace(body.Reason)
if body.Reason == "" { return jsonResp(400, "原因不能为空", nil) }
```

#### 软删除：AutoMigrate 兼容（强制）

```go
// ✅ 使用 gorm.DeletedAt 而非 *time.Time，获得自动软删除过滤
import "gorm.io/gorm"

type ModelTime struct {
    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}
// 使用 gorm.DeletedAt 后，db.Find/First 自动加 deleted_at IS NULL，无需手写
```

#### 敏感字段遮蔽（密钥/密码类字段）

```go
// 列表和详情接口返回前必须遮蔽
func maskSensitive(cfg *PaymentConfig) {
    cfg.StripeSecretKey  = "***"
    cfg.AlipayPrivateKey = "***"
    cfg.WechatApiKey     = "***"
}

// 编辑接口更新时：传入 "***" 表示不修改，传入非空非"***"值才更新
func updateIfNotMasked(incoming, existing *string) {
    if *incoming != "" && *incoming != "***" {
        *existing = *incoming
    }
}
```

#### 删除保护（关联数据检查）

```go
// 删除前必须检查是否有关联子数据
var count int64
db.Model(&ChildModel{}).Where("parent_id = ? AND deleted_at IS NULL", id).Count(&count)
if count > 0 {
    return jsonResp(400, "存在关联数据，无法删除", nil)
}
```

---

## 二、插件前端开发规范

### 2.1 公共工具提取（强制）

**必须创建 `src/utils/request.ts`，不得在每个 Vue 文件中复制 getToken/request 函数。**

```typescript
// src/utils/request.ts（标准模板）
export function getToken(): string {
  try {
    const cookieMatch = document.cookie.match(/authorized-token=([^;]+)/);
    if (cookieMatch) {
      const data = JSON.parse(decodeURIComponent(cookieMatch[1]));
      return data?.accessToken || "";
    }
    const stored = localStorage.getItem("responsive-user-info");
    if (stored) return JSON.parse(stored)?.accessToken || "";
  } catch {}
  return "";
}

export async function request(method: string, url: string, body?: any) {
  const headers: Record<string, string> = { Authorization: `Bearer ${getToken()}` };
  const opts: RequestInit = { method, headers };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  return (await fetch(url, opts)).json();
}

export async function authFetch(url: string, opts: RequestInit = {}): Promise<Response> {
  return fetch(url, {
    ...opts,
    headers: { Authorization: `Bearer ${getToken()}`, ...(opts.headers as Record<string,string>) }
  });
}
```

每个 Vue 页面只需一行导入：
```typescript
import { request, getToken } from "../utils/request";
```

### 2.2 分页参数规范（强制）

```typescript
// ✅ 统一用 pageIndex，与后端 parsePage() 对应
const query = reactive({ pageIndex: 1, pageSize: 10, name: "" });

async function loadData() {
  const params = new URLSearchParams();
  params.set("pageIndex", String(query.pageIndex));
  params.set("pageSize", String(query.pageSize));
  if (query.name) params.set("name", query.name);
  const res = await request("GET", `${API_BASE}?${params.toString()}`);
  if (res.code === 200) {
    list.value  = res.data?.list  || [];
    total.value = res.data?.total || 0;  // ✅ 读 total，不读 count
  }
}
```

### 2.3 编辑路由规范（PUT vs POST）

```typescript
// ✅ 新增用 POST，编辑用 PUT /:id（后端必须实现对应路由）
const res = form.id
  ? await request("PUT", `${API_BASE}/${form.id}`, form)
  : await request("POST", API_BASE, form);
```

**⚠️ 高频错误**：`PaymentConfig.vue` 等页面编辑时发 `POST`，但后端 `routePayment` 只实现了 `POST`（新增），导致编辑走了 upsert 而非精确更新。必须对应 PUT 路由。

### 2.4 二次确认文案规范（强制）

```vue
<!-- ✅ 正确：文案包含操作对象的标识符 -->
<el-popconfirm :title="`确认删除「${row.name}」？`" @confirm="onDelete(row)">

<!-- ❌ 错误：通用文案，无法告知用户操作的具体对象 -->
<el-popconfirm title="确认删除？" @confirm="onDelete(row)">
```

### 2.5 敏感字段编辑体验（密钥类字段）

```typescript
// 打开编辑对话框时：服务端返回 "***" 的密钥字段，清空让用户按需重填
function openDialog(row?: any) {
  Object.assign(form, defaultForm());
  if (row) {
    Object.assign(form, row);
    // 清空占位符，placeholder 提示"不修改请留空"
    if (form.stripeSecretKey === "***") form.stripeSecretKey = "";
    if (form.alipayPrivateKey === "***") form.alipayPrivateKey = "";
    if (form.wechatApiKey     === "***") form.wechatApiKey = "";
  }
  dialogVisible.value = true;
}
```

```vue
<!-- 编辑时密钥字段加 placeholder 说明 -->
<el-input v-model="form.stripeSecretKey" show-password
  :placeholder="form.id ? '不修改请留空' : ''" />
```

### 2.6 导出 CSV 功能规范

```typescript
// ✅ 按钮文案必须标明"当前页"，避免用户误以为是全量导出
// ✅ 空数据时提示，不要静默生成空文件
function onExport() {
  if (list.value.length === 0) { ElMessage.warning("当前页无数据"); return; }
  // ... 生成 CSV ...
  ElMessage.info(`已导出第 ${query.pageIndex} 页共 ${list.value.length} 条数据`);
}
```

---

## 三、构建与部署规范

### 3.1 插件二进制构建（必须显式指定输出路径）

**⚠️ 已知陷阱**：在 `plugins/game/` 下编译的 `game.exe` 和主服务加载的 `backend/plugins/game/game.exe` 是两个不同路径。修改代码重新编译后必须同步到主服务读取的路径，否则测试打的是旧进程。

```powershell
# ✅ 正确：直接编译到主服务加载路径
cd projects/demo/platform_admin/plugins/{name}
go build -o ../../backend/plugins/{name}/{name}.exe .

# ❌ 错误：只编译本地，忘记复制
go build -o {name}.exe .
# 然后忘记 Copy-Item
```

**build.ps1 模板（推荐）：**
```powershell
param([string]$PluginName)
$srcDir  = "plugins/$PluginName"
$dstExe  = "backend/plugins/$PluginName/$PluginName.exe"
Set-Location $srcDir
go build -o "../../$dstExe" .
if ($LASTEXITCODE -ne 0) { Write-Error "编译失败"; exit 1 }
Write-Host "✅ 编译成功：$dstExe"
```

### 3.2 插件热重载（修改代码后）

修改插件代码并重新编译后，必须通过主服务 API 重启插件进程。
stop/start 是插件管理接口，使用平台级 token（`platform_token`）即可：

```powershell
$loginBody = (Invoke-WebRequest -Uri "http://localhost:8000/auth/login" `
    -Method POST -ContentType "application/json" `
    -Body '{"username":"admin","password":"admin123"}' `
    -UseBasicParsing).Content | ConvertFrom-Json

# ✅ 插件管理接口（stop/start）用 platform_token
$platformToken = $loginBody.data.token

# 先停后启，确保加载新二进制
Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins/{name}/stop" `
    -Method POST -Headers @{Authorization="Bearer $platformToken"} -UseBasicParsing
Start-Sleep 1
Invoke-WebRequest -Uri "http://localhost:8000/api/v1/admin/plugins/{name}/start" `
    -Method POST -Headers @{Authorization="Bearer $platformToken"} -UseBasicParsing
```

> 注：调用插件业务接口（如游戏列表）需要 `access_token`（租户级），参见第四章 4.2。

### 3.3 前端构建与部署

```powershell
# 构建前端
cd plugins/{name}/frontend
pnpm build  # 或 npm run build

# 复制到主服务静态目录（路径需与 plugin.json frontends[].entry 匹配）
Copy-Item dist/index.js   ../../backend/static/plugins/{name}/admin-pc/bundle.js -Force
Copy-Item dist/index.css  ../../backend/static/plugins/{name}/admin-pc/bundle.css -Force
```

**⚠️ vite.config.ts 关键配置（external 必须声明）：**
```typescript
export default defineConfig({
  build: {
    lib: { entry: "src/index.ts", name: "PluginXxx", fileName: () => "index.js", formats: ["es"] },
    rollupOptions: {
      // ✅ vue 和 element-plus 必须 external，否则打包体积过大且与宿主冲突
      external: ["vue", "element-plus"],
      output: {
        paths: {
          vue: "/plugin-shims/vue.js",
          "element-plus": "/plugin-shims/element-plus.js"
        }
      }
    }
  }
});
```

---

## 四、e2e 测试规范（插件专用）

### 4.1 API 路径（插件代理路由）

插件接口**不是** `/api/v1/plugin/game/...`，而是：
```
/api/v1/admin/plugin/{pluginName}/{子路由}
```

例如 game 插件的游戏列表：
```
GET /api/v1/admin/plugin/game/game?pageIndex=1&pageSize=10
```

### 4.2 认证：必须走租户选择流程

```typescript
async function loginAdmin(api: APIRequestContext): Promise<string> {
  const resp = await api.post(`${API_BASE}/auth/login`, {
    data: { username: "admin", password: "admin123" }
  });
  const body = await resp.json();
  // ✅ 必须走租户选择，否则拿到的是平台级 token，插件接口会返回 503
  if (body.data?.tenants?.length > 0) {
    const tenantResp = await api.post(`${API_BASE}/auth/tenant/select`, {
      data: { tenant_id: String(body.data.tenants[0].id) },
      headers: { Authorization: `Bearer ${body.data.token}` }
    });
    return (await tenantResp.json()).data?.access_token;
  }
  return body.data?.access_token;
}
```

### 4.3 插件业务错误码规范

插件内部业务错误通过 `resp.body.code` 返回，HTTP status 固定为 200：

| 场景 | HTTP Status | body.code |
|------|------------|-----------|
| 成功 | 200 | 200 |
| 参数错误 | 200 | 400 |
| 未找到资源 | 200 | 404 |
| 服务不可用（DB未初始化） | 503 | 503 |
| 插件未运行 | 503 | 50301 |
| 插件不存在 | 404（主服务层） | 40400 |

测试中检查方式：
```typescript
// ✅ 正确：先检查 HTTP 200，再检查业务码
expect(resp.status()).toBe(200);
const body = await resp.json();
expect(body.code).toBe(200);

// ✅ 正确：处理主服务层 404（插件不存在或路由不存在）
if (resp.status() === 200) {
  expect(body.code).toBe(404);
} else {
  expect(resp.status()).toBe(404);
}
```

### 4.4 test.skip 正确用法（Playwright 规范）

```typescript
// ✅ 正确：在测试体第一行调用，Playwright 会真正跳过后续代码
test("TC-G06 重置密钥", async () => {
  test.skip(!createdGameId, "依赖 TC-G01，跳过");
  // 以下代码在 createdGameId=0 时不会执行
  const resp = await api.post(...)
});

// ❌ 错误：if + return 模式，test.skip() 内部调用后 return 并不能真正终止测试
test("TC-G06 重置密钥", async () => {
  if (!createdGameId) { test.skip(true, "跳过"); return; }
  // ⚠️ 如果 test.skip 抛出异常被某处捕获，后续代码仍然执行
});
```

---

## 五、安全规范

### 5.1 敏感数据存储

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
    "os"
)

// encryptAESGCM 用 AES-256-GCM 加密（密钥从环境变量读取）
// 环境变量：PLUGIN_ENCRYPT_KEY，32字节 base64 编码
func encryptAESGCM(plaintext string) (string, error) {
    keyB64 := os.Getenv("PLUGIN_ENCRYPT_KEY")
    if keyB64 == "" {
        return "", fmt.Errorf("PLUGIN_ENCRYPT_KEY 未配置")
    }
    key, err := base64.StdEncoding.DecodeString(keyB64)
    if err != nil || len(key) != 32 {
        return "", fmt.Errorf("PLUGIN_ENCRYPT_KEY 必须是 32 字节的 base64 编码")
    }
    block, err := aes.NewCipher(key)
    if err != nil { return "", err }
    gcm, err := cipher.NewGCM(block)
    if err != nil { return "", err }
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil { return "", err }
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAESGCM 解密（调用三方 API 前使用）
func decryptAESGCM(cipherB64 string) (string, error) {
    keyB64 := os.Getenv("PLUGIN_ENCRYPT_KEY")
    if keyB64 == "" {
        return "", fmt.Errorf("PLUGIN_ENCRYPT_KEY 未配置")
    }
    key, err := base64.StdEncoding.DecodeString(keyB64)
    if err != nil || len(key) != 32 {
        return "", fmt.Errorf("PLUGIN_ENCRYPT_KEY 必须是 32 字节的 base64 编码")
    }
    data, err := base64.StdEncoding.DecodeString(cipherB64)
    if err != nil { return "", err }
    block, err := aes.NewCipher(key)
    if err != nil { return "", err }
    gcm, err := cipher.NewGCM(block)
    if err != nil { return "", err }
    nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil { return "", err }
    return string(plaintext), nil
}

// 使用示例：保存支付配置时加密，返回时遮蔽
model.StripeSecretKey, _ = encryptAESGCM(body.StripeSecretKey)
// 返回时遮蔽：
model.StripeSecretKey = "***"
```

**密钥管理最低要求**：
- 32 字节随机密钥，`openssl rand -base64 32` 生成
- 通过环境变量 `PLUGIN_ENCRYPT_KEY` 注入，不写死代码，不提交 git
- 最低要求（无加密条件时）：返回时遮蔽 `"***"`，绝不明文出现在接口响应中

### 5.2 文件上传安全

```go
// ❌ 错误：只检查 Body 非空，不验证文件类型
if len(req.Body) == 0 { return jsonResp(400, "文件为空", nil) }

// ✅ 正确：检查文件 magic bytes（以 .pck 为例）
func validatePckFile(data []byte) bool {
    // Godot PCK 文件头: "GDPC"
    return len(data) >= 4 && string(data[:4]) == "GDPC"
}
if !validatePckFile(req.Body) {
    return jsonResp(400, "无效的 PCK 文件格式", nil)
}
```

### 5.3 下载 URL 安全

```go
// ❌ 问题：静态路径无鉴权，任何人知道 URL 都能下载
return jsonResp(200, "ok", map[string]string{"url": "/static/dlc/" + filename})

// ✅ 推荐：生成带时效的签名 URL（或通过鉴权接口代理下载）
signedURL := generateSignedURL(filename, time.Now().Add(15*time.Minute))
return jsonResp(200, "ok", map[string]string{"url": signedURL})
```

---

## 六、开发前置检查清单（每次开始插件开发时执行）

```
后端 Go 代码
─────────────────────────────────────────────────────
□ plugin.json 中所有 apiPermissions 端点在 handler.go 都有对应 case
□ 编辑（PUT /:id）路由已实现，不缺失
□ 特殊动作路由（如 /regen-secret）已在路由表中显式声明
□ parsePage() 兼容 pageIndex 和 page 两种参数
□ PageResponse 使用 total 字段（不用 count）
□ 数字为 0 时的字段更新逻辑正确（price=0 可以更新）
□ 敏感字段（密钥/密码）在响应中已遮蔽为 "***"
□ 编辑接口收到 "***" 时不覆盖原值
□ 字符串校验已加 strings.TrimSpace()
□ 软删除使用 gorm.DeletedAt 或手动加 deleted_at IS NULL
□ 删除前检查关联子数据
□ go build ./... 零错误，go vet ./... 无警告
□ 编译输出到主服务路径（backend/plugins/{name}/{name}.exe）

前端 Vue 代码
─────────────────────────────────────────────────────
□ src/utils/request.ts 已创建，所有 Vue 文件 import 公共函数
□ 分页参数使用 pageIndex（不用 page）
□ 从响应读 res.data?.total（不读 count）
□ 编辑操作发 PUT /:id，新增操作发 POST
□ 危险操作有 el-popconfirm，文案含操作对象名称
□ 密钥类字段编辑时清空占位符，placeholder 提示"不修改请留空"
□ 导出功能标明"当前页"，空数据时给出提示
□ 表单校验规则完整（必填字段、格式要求）

构建与部署
─────────────────────────────────────────────────────
□ 首次新建插件已完成 Step 1-8（目录→plugin.json→go.mod→编译→注册→验证状态）
□ 后端重新编译后执行插件 stop → start 热重载
□ 前端重新构建后复制到 backend/static/plugins/{name}/admin-pc/
□ 编译路径与主服务加载路径一致（通过 api/v1/admin/plugins 验证状态）
□ 插件 runStatus 为 "running" 后再执行 e2e 测试

e2e 测试
─────────────────────────────────────────────────────
□ API 测试路径使用 /api/v1/admin/plugin/{name}/... 
□ loginAdmin() 包含租户选择步骤
□ test.skip 在测试体第一行调用（不用 if + return 模式）
□ HTTP 200 + 业务码 与 主服务层 HTTP 4xx 的断言分支都已处理
□ 跨测试共享的可变 ID（createdId 等）通过 beforeAll 正确初始化
```

---

## 七、速查：常见问题与修复

| 现象 | 根因 | 修复 |
|------|------|------|
| 接口返回 404（`接口不存在`） | handler.go 缺少对应 case | 在 route 函数中添加 case |
| 接口返回 503（`插件未初始化`） | 编译的 exe 未复制到主服务路径，或插件进程未启动 | 检查 `backend/plugins/{name}/` 目录，执行 stop→start 重载 |
| 编辑返回 404 | routeXxx 只有 POST 没有 PUT | 添加 `case method == "PUT" && isNumeric(sub)` |
| 分页始终返回第一页 | 前端传 pageIndex 但后端只读 page | 修复 parsePage() 兼容两种参数 |
| total 显示 0 | 前端读 `res.data?.count` 但后端返回 `total` | 统一改为读 `total` |
| price=0 无法保存 | `if body.Price > 0` 跳过了 0 值 | 改为 `if body.Price >= 0` 或联动 isFree |
| 密钥更新后变 "***" | 编辑时传了 "***" 占位值 | 前端清空密钥字段，后端检测 "***" 跳过 |
| 删除成功但关联数据残留 | 未检查子数据就删除父记录 | 删除前 Count 子记录，非零拒绝删除 |
| e2e 全部 503 | loginAdmin 未走租户选择，用了平台级 token | 修复 loginAdmin 加入 tenant/select 步骤 |
| e2e test.skip 无效，仍然执行 | 用了 `if (!id) { test.skip(); return; }` 写法 | 改为第一行直接 `test.skip(!id, "...")` |
