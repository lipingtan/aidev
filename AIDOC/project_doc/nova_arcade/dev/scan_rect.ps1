param([string]$Path = "dev\rect_test.png")
Add-Type -AssemblyName System.Drawing
$bmp = [System.Drawing.Bitmap]::FromFile((Resolve-Path $Path))
"image {0}x{1}" -f $bmp.Width, $bmp.Height
foreach ($y in @(200, 500)) {
    "=== y=$y ==="
    $cur = $null; $s = 0; $out = @()
    for ($x = 0; $x -lt $bmp.Width; $x++) {
        $c = $bmp.GetPixel($x, $y)
        $k = "{0:X2}{1:X2}{2:X2}" -f $c.R, $c.G, $c.B
        if ($k -ne $cur) {
            if ($cur -and ($x - $s) -ge 6) { $out += "{0,4}-{1,4} #{2} w{3}" -f $s, $x, $cur, ($x - $s) }
            $cur = $k; $s = $x
        }
    }
    $out | Select-Object -First 10
}
$bmp.Dispose()
