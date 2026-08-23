$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'ref'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$targets = [ordered]@{
  'sly_knights'    = 'https://www.slynyrd.com/blog/2025/7/28/pixelblog-57-knights-monsters-amp-castles'
  'sly_topdown3'   = 'https://www.slynyrd.com/blog/2025/10/2/pixelblog-58-top-down-character-animation-part-3'
}
foreach ($k in $targets.Keys) {
  $out = Join-Path $dir ($k + '.out')
  curl.exe -s -L --max-time 30 -A $ua $targets[$k] -o $out
  if (Test-Path $out) { $len = (Get-Item $out).Length } else { $len = -1 }
  Write-Host "$k -> LEN:$len"
}
