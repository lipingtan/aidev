$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'ref'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$sites = [ordered]@{
  'pixilart'   = 'https://www.pixilart.com'
  'pixeljoint' = 'http://pixeljoint.com'
  'pixelera'   = 'https://www.pixelera.art'
  'lospec'     = 'https://lospec.com'
  'floor796'   = 'https://floor796.com'
  'wplace'     = 'https://wplace.live'
}
foreach ($k in $sites.Keys) {
  $out = Join-Path $dir ($k + '.html')
  curl.exe -s -L --max-time 30 -A $ua $sites[$k] -o $out
  if (Test-Path $out) { $len = (Get-Item $out).Length } else { $len = -1 }
  Write-Host "$k -> LEN:$len"
}
