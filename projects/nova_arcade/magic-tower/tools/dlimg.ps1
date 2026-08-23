$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
# 候选像素画图片源（Kenney CC0 / OpenGameArt / 通用 sprite sheet）
$cands = [ordered]@{
  'kenney_rpg'      = 'https://kenney.nl/media/pages/assets/1226/rpg-urban-pack.png'
  'og_a1'           = 'https://opengameart.org/sites/default/files/spritesheet.png'
  'lospec_logo'     = 'https://lospec.com/static/img/logo.png'
  'github_sheet'    = 'https://raw.githubusercontent.com/lospec/palette-list/master/palettes/lospec.png'
}
foreach ($k in $cands.Keys) {
  $out = Join-Path $dir ($k + '.png')
  curl.exe -s -L --max-time 25 -A $ua $cands[$k] -o $out
  if (Test-Path $out) {
    $bytes = [System.IO.File]::ReadAllBytes($out)
    $magic = ''
    if ($bytes.Length -ge 4) { $magic = ($bytes[0..3] | ForEach-Object { $_.ToString('X2') }) -join ' ' }
    Write-Host "$k -> size:$($bytes.Length) magic:$magic"
  } else { Write-Host "$k -> no file" }
}
