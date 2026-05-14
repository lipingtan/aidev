# ============================================================
# 游戏管理服务 一键构建脚本（Windows PowerShell）
# 用法：在 game_server/ 目录下执行 .\build.ps1
# ============================================================

param(
    [switch]$FrontendOnly,   # 只构建前端
    [switch]$BackendOnly,    # 只构建后端
    [switch]$Run             # 构建完成后直接运行
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# 设置 Go 代理
$env:HTTPS_PROXY = "socks5://127.0.0.1:10808"
$env:HTTP_PROXY  = "socks5://127.0.0.1:10808"
$env:GOPROXY     = "https://proxy.golang.org,direct"
$env:GONOSUMDB   = "*"

function Write-Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Success($msg) { Write-Host "✅ $msg" -ForegroundColor Green }
function Write-Fail($msg) { Write-Host "❌ $msg" -ForegroundColor Red; exit 1 }

# ─────────────────────────────────────────────
# Step 1: 构建前端（pure-admin + Vite）
# ─────────────────────────────────────────────
if (-not $BackendOnly) {
    Write-Step "构建前端（pure-admin / Vite）..."

    $FrontendDir = Join-Path $ScriptDir "frontend"
    if (-not (Test-Path $FrontendDir)) { Write-Fail "前端目录不存在: $FrontendDir" }

    Push-Location $FrontendDir

    # 安装依赖（如果 node_modules 不存在）
    if (-not (Test-Path "node_modules")) {
        Write-Step "安装前端依赖（pnpm）..."
        pnpm install
        if ($LASTEXITCODE -ne 0) { Write-Fail "依赖安装失败" }
    }

    # 构建
    Write-Step "打包前端..."
    pnpm build
    if ($LASTEXITCODE -ne 0) { Write-Fail "前端构建失败" }

    Pop-Location
    Write-Success "前端构建完成"

    # ─────────────────────────────────────────────
    # Step 2: 将前端产物复制到后端 web/dist/ 目录
    # ─────────────────────────────────────────────
    Write-Step "将前端产物复制到后端 web/dist/ 目录..."

    $DistDir = Join-Path $FrontendDir "dist"
    $WebDir  = Join-Path $ScriptDir "backend" "web" "dist"

    if (-not (Test-Path $DistDir)) { Write-Fail "前端构建产物不存在: $DistDir" }

    if (Test-Path $WebDir) {
        Get-ChildItem $WebDir -Exclude ".gitkeep" | Remove-Item -Recurse -Force
    } else {
        New-Item -ItemType Directory -Path $WebDir -Force | Out-Null
    }

    Copy-Item -Recurse -Force "$DistDir\*" $WebDir
    Write-Success "前端产物已复制到 backend/web/dist/"
}

# ─────────────────────────────────────────────
# Step 3: 编译 Go 后端
# ─────────────────────────────────────────────
if (-not $FrontendOnly) {
    Write-Step "编译 Go 后端..."

    $BackendDir = Join-Path $ScriptDir "backend"
    if (-not (Test-Path $BackendDir)) { Write-Fail "后端目录不存在: $BackendDir" }

    $OutputDir  = Join-Path $ScriptDir "dist"
    $OutputFile = Join-Path $OutputDir "game-server.exe"
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

    Write-Step "整理 Go 依赖..."
    Push-Location $BackendDir
    go mod tidy
    if ($LASTEXITCODE -ne 0) { Write-Fail "go mod tidy 失败" }

    Write-Step "编译中..."
    go build -ldflags="-s -w" -o $OutputFile .
    if ($LASTEXITCODE -ne 0) { Write-Fail "Go 编译失败" }
    Pop-Location

    Write-Success "后端编译完成: dist/game-server.exe"
}

# ─────────────────────────────────────────────
# 完成
# ─────────────────────────────────────────────
Write-Host "`n============================================" -ForegroundColor Green
Write-Host "  构建完成！" -ForegroundColor Green
Write-Host "  可执行文件: projects/game_server/dist/game-server.exe" -ForegroundColor Green
Write-Host "  启动方式: cd dist && .\game-server.exe server" -ForegroundColor Green
Write-Host "============================================`n" -ForegroundColor Green

if ($Run) {
    Write-Step "启动服务..."
    $OutputFile = Join-Path $ScriptDir "dist" "game-server.exe"
    Set-Location (Join-Path $ScriptDir "dist")
    & $OutputFile server
}
