$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
# PixelJoint 评分榜 · 奇幻题材（骑士/怪物/龙/城堡）
$urls = @(
  'https://pixeljoint.com/files/icons/spartan_small.gif',
  'https://pixeljoint.com/files/icons/wayuki_warrior_of_light_icon.gif',
  'https://pixeljoint.com/files/icons/protecteur_v.gif',
  'https://pixeljoint.com/files/icons/orcspreview.gif',
  'https://pixeljoint.com/files/icons/skavenpreview.gif',
  'https://pixeljoint.com/files/icons/sm_22.gif',
  'https://pixeljoint.com/files/icons/isocastle_prev.png',
  'https://pixeljoint.com/files/icons/kraken__r177826211.png'
)
$i=0
foreach ($u in $urls) {
  $fn = Join-Path $dir ("pj_$i" + [System.IO.Path]::GetExtension($u))
  curl.exe -s -L --max-time 25 -A $ua $u -o $fn
  if (Test-Path $fn) { $b=[System.IO.File]::ReadAllBytes($fn); if($b.Length -ge4){$m=($b[0..3]|ForEach-Object{$_.ToString('X2')}) -join ' '}else{$m=''}; Write-Host "[$i] $($b.Length)B magic:$m  <= $([System.IO.Path]::GetFileNameWithoutExtension($u))" } else { Write-Host "[$i] no-file  <= $u" }
  $i++
}
