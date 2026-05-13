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

function Write-Step($msg) {
    Write-Host "`n==> $msg" -ForegroundColor Cyan
}

function Write-Success($msg) {
    Write-Host "✅ $msg" -ForegroundColor Green
}

function Write-Fail($msg) {
    Write-Host "❌ $msg" -ForegroundColor Red
    exit 1
}

# ─────────────────────────────────────────────
# Step 1: 构建前端
# ─────────────────────────────────────────────
if (-not $BackendOnly) {
    Write-Step "构建前端（Vue 3 + Vite）..."

    $FrontendDir = Join-Path $ScriptDir "frontend"
    if (-not (Test-Path $FrontendDir)) {
        Write-Fail "前端目录不存在: $FrontendDir"
    }

    Push-Location $FrontendDir

    # 安装依赖（如果 node_modules 不存在）
    if (-not (Test-Path "node_modules")) {
        Write-Step "安装前端依赖..."
        npm install
        if ($LASTEXITCODE -ne 0) { Write-Fail "npm install 失败" }
    }

    # 构建
    npm run build:prod
    if ($LASTEXITCODE -ne 0) { Write-Fail "前端构建失败" }

    Pop-Location
    Write-Success "前端构建完成"

    # ─────────────────────────────────────────────
    # Step 2: 将前端产物复制到后端 web/ 目录
    # ─────────────────────────────────────────────
    Write-Step "将前端产物复制到后端 web/ 目录..."

    $DistDir = Join-Path $FrontendDir "dist"
    $WebDir  = Join-Path $ScriptDir "backend" "web"

    if (-not (Test-Path $DistDir)) {
        Write-Fail "前端构建产物不存在: $DistDir"
    }

    # 清空旧的 web/ 目录
    if (Test-Path $WebDir) {
        Remove-Item -Recurse -Force $WebDir
    }
    New-Item -ItemType Directory -Path $WebDir -Force | Out-Null

    # 复制
    Copy-Item -Recurse -Force "$DistDir\*" $WebDir
    Write-Success "前端产物已复制到 backend/web/"
}

# ─────────────────────────────────────────────
# Step 3: 编译 Go 后端
# ─────────────────────────────────────────────
if (-not $FrontendOnly) {
    Write-Step "编译 Go 后端..."

    $BackendDir = Join-Path $ScriptDir "backend"
    if (-not (Test-Path $BackendDir)) {
        Write-Fail "后端目录不存在: $BackendDir"
    }

    Push-Location $BackendDir

    # 下载依赖
    Write-Step "下载 Go 依赖..."
    go mod tidy
    if ($LASTEXITCODE -ne 0) { Write-Fail "go mod tidy 失败" }

    # 编译（输出到 game_server/dist/ 目录）
    $OutputDir  = Join-Path $ScriptDir "dist"
    $OutputFile = Join-Path $OutputDir "game-server.exe"

    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

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
Write-Host "============================================`n" -ForegroundColor Green

# 可选：直接运行
if ($Run) {
    Write-Step "启动服务..."
    $OutputFile = Join-Path $ScriptDir "dist" "game-server.exe"
    & $OutputFile
}
