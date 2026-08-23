# Scanline dump: identify what's actually drawn at key rows.
param([string]$Path = "dev\shot_03_after_drops.png")
Add-Type -AssemblyName System.Drawing
$bmp = [System.Drawing.Bitmap]::FromFile((Resolve-Path $Path))
$w = $bmp.Width; $h = $bmp.Height
"image ${w}x${h}"

function Scan([int]$y) {
    "=== scanline y=$y ==="
    $runs = @()
    $cur = $null; $start = 0
    for ($x = 0; $x -lt $w; $x++) {
        $c = $bmp.GetPixel($x, $y)
        $key = "{0:X2}{1:X2}{2:X2}" -f $c.R, $c.G, $c.B
        if ($key -ne $cur) {
            if ($cur -ne $null -and ($x - $start) -ge 3) { $runs += ("{0,4}-{1,4} #{2} w{3}" -f $start, $x, $cur, ($x - $start)) }
            $cur = $key; $start = $x
        }
    }
    $runs | Select-Object -First 22 | ForEach-Object { $_ }
}

# design 720x1280 -> image 540x960 (scale .75)
# board top y=161d -> 121i ; each design row 46.08d -> 34.56i
# blob at design rows 7-11 -> image y ~ 362..535 ; scan y=450
Scan 450
# where stack SHOULD be: design row 17 -> 120.75 + 17.5*34.55 = 725
Scan 725
# falling piece row 0-1: image y ~ 140
Scan 140
# top HUD strip: y=60
Scan 60
$bmp.Dispose()
