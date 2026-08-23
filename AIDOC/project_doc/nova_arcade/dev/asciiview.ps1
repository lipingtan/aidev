param([string]$Path = "dev\shot_03_after_drops.png", [int]$Step = 18)
Add-Type -AssemblyName System.Drawing
$bmp = [System.Drawing.Bitmap]::FromFile((Resolve-Path $Path))
"image {0}x{1}, step {2}  (chars: . dim, - low, + mid, # bright, @ very bright)" -f $bmp.Width, $bmp.Height, $Step
$sb = New-Object System.Text.StringBuilder
for ($y = 0; $y -lt $bmp.Height; $y += $Step) {
    $line = ""
    for ($x = 0; $x -lt $bmp.Width; $x += $Step) {
        $c = $bmp.GetPixel($x, $y)
        $max = [Math]::Max($c.R, [Math]::Max($c.G, $c.B))
        $ch = "."
        if ($max -ge 220) { $ch = "@" }
        elseif ($max -ge 140) { $ch = "#" }
        elseif ($max -ge 70) { $ch = "+" }
        elseif ($max -ge 34) { $ch = "-" }
        $line += $ch
    }
    [void]$sb.AppendLine($line)
}
$sb.ToString()
$bmp.Dispose()
