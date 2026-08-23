$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null

function Try-Dl($url, $fn) {
  curl.exe -s -L --max-time 25 -A $ua $url -o $fn
  if (Test-Path $fn) { $b=[System.IO.File]::ReadAllBytes($fn); if($b.Length -ge4){$m=($b[0..3]|ForEach-Object{$_.ToString('X2')}) -join ' '}else{$m=''}; return "$($b.Length)B magic:$m" }
  return 'no-file'
}

# 抓评分榜 + 最新，提取全尺寸作品图 /files/icons/full/...
$all = @()
foreach ($page in @('new_icons.asp?ob=rating','new_icons.asp?ob=date')) {
  $h = curl.exe -s -L --max-time 25 -A $ua ("https://pixeljoint.com/pixels/" + $page)
  Write-Host "### $page  html=$($h.Length)"
  $u = [regex]::Matches($h, 'https?://pixeljoint\.com/files/icons/full/[^\s"''<>\\]+?\.(?:png|jpe?g|gif)', [System.Text.RegularExpressions.RegexOptions]::IgnoreCase) | ForEach-Object { $_.Value } | Where-Object {$_} | Select-Object -Unique
  $all += $u
}
$all = $all | Select-Object -Unique
Write-Host "TOTAL full artwork urls: $($all.Count)"
$i=0
foreach ($u in $all) {
  if ($i -ge 8) { break }
  $fn = Join-Path $dir ("pjart_$i" + [System.IO.Path]::GetExtension($u))
  $r = Try-Dl $u $fn
  Write-Host "[$i] $r  <= $u"
  $i++
}
