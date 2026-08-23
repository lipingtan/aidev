$ErrorActionPreference = 'SilentlyContinue'
$ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
$dir = Join-Path $PSScriptRoot 'ref'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$targets = [ordered]@{
  'slynyrd_catalogue' = 'https://www.slynyrd.com/pixelblog-catalogue'
}
foreach ($k in $targets.Keys) {
  $out = Join-Path $dir ($k + '.out')
  curl.exe -s -L --max-time 30 -A $ua $targets[$k] -o $out
  if (Test-Path $out) { $len = (Get-Item $out).Length } else { $len = -1 }
  Write-Host "$k -> LEN:$len"
}
