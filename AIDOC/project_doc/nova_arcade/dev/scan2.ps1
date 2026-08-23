param([string]$Path = "dev\shot_03_after_drops.png")
Add-Type -AssemblyName System.Drawing
$bmp = [System.Drawing.Bitmap]::FromFile((Resolve-Path $Path))
"image {0}x{1}" -f $bmp.Width, $bmp.Height

function Scan([int]$y, [int]$first = 14) {
    "=== y=$y ==="
    $cur = $null; $s = 0; $out = @()
    for ($x = 0; $x -lt $bmp.Width; $x++) {
        $c = $bmp.GetPixel($x, $y)
        $k = "{0:X2}{1:X2}{2:X2}" -f $c.R, $c.G, $c.B
        if ($k -ne $cur) {
            if ($cur -and ($x - $s) -ge 4) { $out += "{0,4}-{1,4} #{2} w{3}" -f $s, $x, $cur, ($x - $s) }
            $cur = $k; $s = $x
        }
    }
    if ($out.Count -eq 0) { "  (uniform #{0})" -f $cur } else { $out | Select-Object -First $first | ForEach-Object { "  $_" } }
}

# stack rows 14-19 centers: img y = 120.75 + (r+0.5)*34.55
Scan 660   # row ~15
Scan 730   # row ~17
Scan 790   # row ~19
# mystery object row 9
Scan 449
# falling piece rows 0-1
Scan 140
$bmp.Dispose()
