$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null

function Get-ImgUrls($html) {
  # 抓所有 http(s) 图片 URL（png/jpg/gif/webp），排除 logo/icon/avatar 小图
  $m = [regex]::Matches($html, 'https?://[^\s"''<>\\]+?\.(?:png|jpe?g|gif|webp)', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)
  $seen = @{}; $out = @()
  foreach ($x in $m) {
    $u = $x.Value
    if ($u -match 'logo|icon|favicon|avatar|sprite-sheet|\.gif') { continue }
    if ($seen[$u]) { continue }; $seen[$u] = $true
    $out += $u
  }
  return $out
}

function Try-Dl($url, $fn) {
  curl.exe -s -L --max-time 25 -A $ua $url -o $fn
  if (Test-Path $fn) { $b=[System.IO.File]::ReadAllBytes($fn); if($b.Length -ge4){$m=($b[0..3]|ForEach-Object{$_.ToString('X2')}) -join ' '}else{$m=''}; return "$($b.Length)B magic:$m" }
  return 'no-file'
}

# ---- PixelJoint ----
$html = curl.exe -s -L --max-time 25 -A $ua "https://pixeljoint.com/"
$pj = Get-ImgUrls $html
Write-Host "PixelJoint: $($pj.Count) img urls"
$i=0; foreach($u in $pj){ if($i -ge 6){break}; $fn=Join-Path $dir ("pj_$i.png"); $r=Try-Dl $u $fn; Write-Host "  [$i] $r  <= $u"; $i++ }

# ---- Pixilart ----
$html2 = curl.exe -s -L --max-time 25 -A $ua "https://www.pixilart.com/"
Write-Host "Pixilart HTML len: $($html2.Length)"
if ($html2 -match 'Just a moment|cf-challenge|challenge-platform') { Write-Host "  -> Cloudflare challenge (blocked)" }
$px = Get-ImgUrls $html2
Write-Host "Pixilart: $($px.Count) img urls"
$i=0; foreach($u in $px){ if($i -ge 6){break}; $fn=Join-Path $dir ("px_$i.png"); $r=Try-Dl $u $fn; Write-Host "  [$i] $r  <= $u"; $i++ }
