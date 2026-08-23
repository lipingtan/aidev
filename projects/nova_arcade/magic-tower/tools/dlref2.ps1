$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'img'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
# 抓 OGA 搜索页，提取真实 /sites/default/files/ 图片 URL
$terms = @{ 'knight'='knight'; 'monster'='monster sprite'; 'fantasy'='fantasy character'; 'rpg'='rpg character sprite' }
foreach ($t in $terms.Keys) {
  $out = Join-Path $dir ("search_" + $t + ".html")
  curl.exe -s -L --max-time 25 -A $ua ("https://opengameart.org/content-search?keys=" + [uri]::EscapeDataString($terms[$t])) -o $out
  if (Test-Path $out) {
    $c = Get-Content $out -Raw
    $urls = [regex]::Matches($c, 'href="(/sites/default/files/[^"]+\.(?:png|jpg|jpeg))"') | ForEach-Object { $_.Groups[1].Value } | Where-Object { $_ -match '\.(png|jpg|jpeg)$' } | Select-Object -Unique
    Write-Host "### $t : $($urls.Count) file urls"
    $i = 0
    foreach ($u in $urls) {
      if ($i -ge 3) { break }
      $fn = Join-Path $dir ("ref_" + $t + "_" + $i + ([System.IO.Path]::GetExtension($u)))
      curl.exe -s -L --max-time 25 -A $ua ("https://opengameart.org" + $u) -o $fn
      if (Test-Path $fn) { $b=[System.IO.File]::ReadAllBytes($fn); $m=($b[0..3]|ForEach-Object{$_.ToString('X2')}) -join ' '; Write-Host "   $i $u -> $($b.Length)B magic:$m" }
      $i++
    }
  } else { Write-Host "### $t : search page not fetched" }
}
