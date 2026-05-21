# Go 插件开发规范

## 插件工程结构

```
projects/{solution-name}/{project-name}/plugins/{plugin-name}/
├── go.mod              # module game-server/plugins/{plugin-name}
├── main.go             # 插件入口，实现 PluginService 接口
├── handler.go          # HTTP 请求路由分发
├── service.go          # 业务逻辑
├── model.go            # 数据模型（表名前缀 {plugin-name}_）
├── dto.go              # 请求/响应 DTO
├── plugin.json         # 插件描述文件（name/version/description）
└── frontend/           # 插件前端工程（独立 Vite 项目）
    ├── package.json
    ├── vite.config.ts  # library 模式，vue/element-plus 作为 external
    ├── tsconfig.json
    └── src/
        ├── index.ts    # 导出 { routes, menus }
        ├── env.d.ts
        └── views/      # 页面组件
```

## 插件打包规范

### 构建流程

```powershell
# 1. 编译后端二进制
cd plugins/{plugin-name}
go build -o {plugin-name}.exe .

# 2. 构建前端 bundle
cd frontend
pnpm install
pnpm build
# 产出 frontend/dist/index.js

# 3. 打包为 zip
cd ..
# zip 包内容：
#   {plugin-name}.exe
#   plugin.json
#   frontend/dist/index.js
```

### 打包产物路径

插件 zip 包统一放到项目的 `dist/plugins/` 目录下：

```
projects/{solution-name}/{project-name}/
├── dist/
│   ├── {project-name}.exe          # Host 主程序
│   └── plugins/                    # 插件包目录
│       ├── dlc-plugin.zip
│       ├── shop-plugin.zip
│       └── ...
```

### zip 包内部结构

```
{plugin-name}-plugin.zip
├── {plugin-name}.exe       # 插件二进制（Windows）或 {plugin-name}（Linux）
├── plugin.json             # 插件描述文件
└── frontend/
    └── dist/
        └── index.js        # 前端 bundle
```

### plugin.json 格式

```json
{
  "name": "dlc",
  "version": "1.0.0",
  "description": "DLC 管理插件"
}
```

## 安装后目录结构

通过管理界面上传 zip 后，Host 自动解压到：

```
backend/
├── plugins/{plugin-name}/          # 插件二进制 + plugin.json
│   ├── {plugin-name}.exe
│   └── plugin.json
└── static/plugins/{plugin-name}/   # 前端 bundle
    └── index.js
```

## 数据库规范

- 插件表名必须使用 `{plugin-name}_` 前缀（如 `dlc_game_dlc`）
- 插件启动时通过 `db_dsn` 连接主库，使用 `sync.Once` 初始化连接
- 插件负责自己的 AutoMigrate

## 前端 bundle 规范

- 使用 Vite library 模式构建
- `vue`、`vue-router`、`element-plus` 作为 external（Host 已加载）
- 入口文件 `src/index.ts` 必须导出 `routes` 和 `menus`
- 使用原生 `fetch` 发请求（不依赖 Host 的 axios 封装）

## 构建脚本示例

```powershell
# build-plugin.ps1 — 插件一键构建打包脚本
param([string]$PluginName)

$PluginDir = "plugins/$PluginName"
$DistDir = "dist/plugins"

# 编译后端
Push-Location $PluginDir
go build -o "$PluginName.exe" .
Pop-Location

# 构建前端
Push-Location "$PluginDir/frontend"
pnpm install --frozen-lockfile
pnpm build
Pop-Location

# 打包 zip
New-Item -ItemType Directory -Path $DistDir -Force
Compress-Archive -Path @(
    "$PluginDir/$PluginName.exe",
    "$PluginDir/plugin.json",
    "$PluginDir/frontend/dist"
) -DestinationPath "$DistDir/$PluginName-plugin.zip" -Force

Write-Host "插件打包完成: $DistDir/$PluginName-plugin.zip"
```
