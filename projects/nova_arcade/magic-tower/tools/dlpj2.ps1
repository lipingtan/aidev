$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$html = curl.exe -s -L --max-time 25 -A $ua "https://pixeljoint.com/"
Set-Content -Path (Join-Path $dir 'pj_home.html') -Value $html -Encoding UTF8
Write-Host "HTML len: $($html.Length)"

Write-Host "=== all image-ish refs (src/href/data-src, relative+absolute) ==="
$refs = [regex]::Matches($html, '(?:src|data-src|data-original|href)\s*=\s*"([^"]+\.(?:png|jpe?g|gif|webp))"', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
$refs | ForEach-Object { $_.Groups[1].Value } | Where-Object {$_} | Select-Object -Unique | Select-Object -First 40

Write-Host "=== /pixels/ paths (any) ==="
[regex]::Matches($html, '/pixels/[^\s"''<>\\]{2,60}') | ForEach-Object { $_.Value } | Select-Object -Unique | Select-Object -First 30

Write-Host "=== api/json endpoints ==="
[regex]::Matches($html, '(?:/api/[^\s"''<>\\]{2,50}|\.json[^\s"''<>\\]{0,20})') | ForEach-Object { $_.Value } | Select-Object -Unique | Select-Object -First 20
