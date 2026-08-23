# Pixel-level analysis of game screenshots: is the board there? are blocks visible?
param([string]$Dir = "dev")

Add-Type -AssemblyName System.Drawing

function Analyze-Png([string]$path) {
    if (-not (Test-Path $path)) { "MISSING $path"; return }
    $bmp = [System.Drawing.Bitmap]::FromFile((Resolve-Path $path))
    $w = $bmp.Width; $h = $bmp.Height
    "=== $(Split-Path $path -Leaf)  ${w}x${h} ==="

    # board region (design 720x1280): x 130..590, y 161..1082
    $bx0 = [int](130/720*$w); $bx1 = [int](590/720*$w)
    $by0 = [int](161/1280*$h); $by1 = [int](1082/1280*$h)

    $bright = 0; $colored = 0; $total = 0; $sumL = 0.0
    for ($y = $by0; $y -lt $by1; $y += 3) {
        for ($x = $bx0; $x -lt $bx1; $x += 3) {
            $c = $bmp.GetPixel($x, $y)
            $max = [Math]::Max($c.R, [Math]::Max($c.G, $c.B)); $min = [Math]::Min($c.R, [Math]::Min($c.G, $c.B))
            $lum = 0.299*$c.R + 0.587*$c.G + 0.114*$c.B
            $sumL += $lum
            $total++
            if ($lum -gt 100) { $bright++ }
            if (($max - $min) -gt 60 -and $max -gt 80) { $colored++ }
        }
    }
    "board region: avg_lum={0:n1} bright(>100)={1}% colored={2}%" -f ($sumL/$total), (100*$bright/$total), (100*$colored/$total)

    # row profile: for each board grid row (20), count colored pixels on a horizontal strip
    $rows = @()
    for ($r = 0; $r -lt 20; $r++) {
        $y = [int]($by0 + ($r + 0.5) * ($by1 - $by0) / 20)
        $cnt = 0
        for ($x = $bx0 + 4; $x -lt $bx1 - 4; $x += 4) {
            $c = $bmp.GetPixel($x, $y)
            $max = [Math]::Max($c.R, [Math]::Max($c.G, $c.B)); $min = [Math]::Min($c.R, [Math]::Min($c.G, $c.B))
            if (($max - $min) -gt 50 -and $max -gt 70) { $cnt++ }
        }
        $rows += $cnt
    }
    "row colored-pixel counts (top->bottom): " + ($rows -join ",")
    $occupied = ($rows | Where-Object { $_ -gt 3 }).Count
    "rows with blocks: $occupied / 20"

    # outside-board brightness (should be dark starfield)
    $oc = 0; $on = 0
    for ($y = 0; $y -lt $h; $y += 17) {
        for ($x = 0; $x -lt $w; $x += 17) {
            if (($x -ge $bx0 -and $x -le $bx1) -or ($y -ge $by0 -and $y -le $by1)) { continue }
            $c2 = $bmp.GetPixel($x, $y)
            $lum2 = 0.299*$c2.R + 0.587*$c2.G + 0.114*$c2.B
            if ($lum2 -gt 100) { $oc++ }; $on++
        }
    }
    "outside-board bright pixels: {0}%" -f (100*$oc/$on)
    $bmp.Dispose()
    ""
}

Analyze-Png "$Dir\shot_01_just_started.png"
Analyze-Png "$Dir\shot_02_piece_falling.png"
Analyze-Png "$Dir\shot_03_after_drops.png"
