$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
# 首页已暴露的月度获奖作品全尺寸图
$urls = @(
  'https://pixeljoint.com/files/icons/full/butterfly_6_anim_18_fixed.gif',
  'https://pixeljoint.com/files/icons/full/planche01b.png',
  'https://pixeljoint.com/files/icons/full/04262026_greens_6.png',
  'https://pixeljoint.com/files/icons/full/forest_angel2.png',
  'https://pixeljoint.com/files/icons/full/candystudypj.png',
  'https://pixeljoint.com/files/icons/full/desertsnakepj4.gif'
)
$i=0
foreach ($u in $urls) {
  $fn = Join-Path $dir ("pjart_$i" + [System.IO.Path]::GetExtension($u))
  curl.exe -s -L --max-time 25 -A $ua $u -o $fn
  if (Test-Path $fn) { $b=[System.IO.File]::ReadAllBytes($fn); if($b.Length -ge4){$m=($b[0..3]|ForEach-Object{$_.ToString('X2')}) -join ' '}else{$m=''}; Write-Host "[$i] $($b.Length)B magic:$m  <= $([System.IO.Path]::GetFileName($u))" } else { Write-Host "[$i] no-file  <= $u" }
  $i++
}
Write-Host "--- debug: new_icons.asp?ob=rating content ---"
$c = curl.exe -s -L --max-time 20 -A $ua "https://pixeljoint.com/pixels/new_icons.asp?ob=rating"
Write-Host $c
