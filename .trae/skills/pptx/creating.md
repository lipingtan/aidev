# Creating Presentations from Scratch

Use this workflow ONLY when no template `.pptx` file is provided.
If a `.pptx` template exists, use [editing.md](editing.md) instead.

---

## Pre-installed Templates (Preferred)

PREFER starting from a pre-built `.js` template file. These encode professional layouts, font choices, and color palettes.

1. **Search** the template taxonomy for the best visual match:
   ```bash
   python .kiro/skills/pptx/scripts/search_templates.py white luxury
   python .kiro/skills/pptx/scripts/search_templates.py brown aesthetic elegant
   python .kiro/skills/pptx/scripts/search_templates.py --typography serif
   python .kiro/skills/pptx/scripts/search_templates.py --color white red
   python .kiro/skills/pptx/scripts/search_templates.py minimalist --mood corporate --density data-heavy
   python .kiro/skills/pptx/scripts/search_templates.py --mood academic --limit 10
   ```
   Positional arguments are keywords searched across all fields. Field flags (`--mood`, `--color`, `--density`, `--typography`, `--background`, `--accent`) narrow to specific attributes.

2. **Choose** the template whose visual attributes best match the desired presentation style.

3. **Copy** it to your working directory:
   ```bash
   Copy-Item ".kiro\skills\pptx\templates\<chosen_template>.js" ".\presentation.js"
   ```

4. **Edit** the copied file to adapt content, titles, data, and colors. Preserve the template's overall structure, fonts, and layout style.

5. **Run** the script to generate the `.pptx`:
   ```bash
   node presentation.js
   ```

## Template Visual Attributes

**Do not pick a template only based on its name or topic.** Select by **visual design**:
- **Color scheme** — does the palette fit the tone?
- **Typography and density** — does the text style match the audience?
- **Mood** — does the overall aesthetic align with the context?

### Important Rules

- **DO NOT** use `python-pptx` or any Python-based PPTX library for creation.
- **DO NOT** write a PPTX generation script from scratch without checking templates first.
- Always use the PptxGenJS `.js` templates or the PptxGenJS API described below.

---

# PptxGenJS Tutorial

## Setup & Basic Structure

```javascript
const pptxgen = require("pptxgenjs");

let pres = new pptxgen();
pres.layout = 'LAYOUT_16x9';  // 10" × 5.625"
pres.author = 'Your Name';
pres.title = 'Presentation Title';

let slide = pres.addSlide();
slide.addText("Hello World!", { x: 0.5, y: 0.5, fontSize: 36, color: "363636" });

pres.writeFile({ fileName: "Presentation.pptx" });
```

## Layout Dimensions

- `LAYOUT_16x9`: 10" × 5.625" (default)
- `LAYOUT_16x10`: 10" × 6.25"
- `LAYOUT_4x3`: 10" × 7.5"
- `LAYOUT_WIDE`: 13.3" × 7.5"

---

## Text & Formatting

```javascript
// Basic text
slide.addText("Simple Text", {
  x: 1, y: 1, w: 8, h: 2, fontSize: 24, fontFace: "Arial",
  color: "363636", bold: true, align: "center", valign: "middle"
});

// Character spacing (use charSpacing, not letterSpacing)
slide.addText("SPACED TEXT", { x: 1, y: 1, w: 8, h: 1, charSpacing: 6 });

// Rich text arrays
slide.addText([
  { text: "Bold ", options: { bold: true } },
  { text: "Italic ", options: { italic: true } }
], { x: 1, y: 3, w: 8, h: 1 });

// Multi-line text (requires breakLine: true)
slide.addText([
  { text: "Line 1", options: { breakLine: true } },
  { text: "Line 2", options: { breakLine: true } },
  { text: "Line 3" }
], { x: 0.5, y: 0.5, w: 8, h: 2 });

// margin: 0 for precise alignment with shapes/icons
slide.addText("Title", { x: 0.5, y: 0.3, w: 9, h: 0.6, margin: 0 });
```

---

## Lists & Bullets

```javascript
// ✅ CORRECT: Multiple bullets
slide.addText([
  { text: "First item", options: { bullet: true, breakLine: true } },
  { text: "Second item", options: { bullet: true, breakLine: true } },
  { text: "Third item", options: { bullet: true } }
], { x: 0.5, y: 0.5, w: 8, h: 3 });

// ❌ WRONG: Never use unicode bullets
slide.addText("• First item", { ... });  // Creates double bullets

// Sub-items and numbered lists
{ text: "Sub-item", options: { bullet: true, indentLevel: 1 } }
{ text: "First", options: { bullet: { type: "number" }, breakLine: true } }
```

---

## Shapes

```javascript
slide.addShape(pres.shapes.RECTANGLE, {
  x: 0.5, y: 0.8, w: 1.5, h: 3.0,
  fill: { color: "FF0000" }, line: { color: "000000", width: 2 }
});

// With transparency
slide.addShape(pres.shapes.RECTANGLE, {
  x: 1, y: 1, w: 3, h: 2,
  fill: { color: "0088CC", transparency: 50 }
});

// With shadow (offset MUST be non-negative)
slide.addShape(pres.shapes.RECTANGLE, {
  x: 1, y: 1, w: 3, h: 2,
  fill: { color: "FFFFFF" },
  shadow: { type: "outer", color: "000000", blur: 6, offset: 2, angle: 135, opacity: 0.15 }
});
```

**Note**: Gradient fills are not natively supported. Use a gradient image as background instead.

---

## Charts

```javascript
// Bar chart
slide.addChart(pres.charts.BAR, [{
  name: "Sales", labels: ["Q1", "Q2", "Q3", "Q4"], values: [4500, 5500, 6200, 7100]
}], { x: 0.5, y: 0.6, w: 6, h: 3, barDir: 'col', showTitle: true, title: 'Quarterly Sales' });

// Line chart
slide.addChart(pres.charts.LINE, [{
  name: "Temp", labels: ["Jan", "Feb", "Mar"], values: [32, 35, 42]
}], { x: 0.5, y: 4, w: 6, h: 3, lineSize: 3, lineSmooth: true });

// Pie chart
slide.addChart(pres.charts.PIE, [{
  name: "Share", labels: ["A", "B", "Other"], values: [35, 45, 20]
}], { x: 7, y: 1, w: 5, h: 4, showPercent: true });
```

### Better-Looking Charts

```javascript
slide.addChart(pres.charts.BAR, chartData, {
  x: 0.5, y: 1, w: 9, h: 4, barDir: "col",
  chartColors: ["0D9488", "14B8A6", "5EEAD4"],
  chartArea: { fill: { color: "FFFFFF" }, roundedCorners: true },
  catAxisLabelColor: "64748B", valAxisLabelColor: "64748B",
  valGridLine: { color: "E2E8F0", size: 0.5 },
  catGridLine: { style: "none" },
  showValue: true, dataLabelPosition: "outEnd", dataLabelColor: "1E293B",
  showLegend: false,
});
```

---

## Tables

```javascript
slide.addTable([
  ["Header 1", "Header 2"],
  ["Cell 1", "Cell 2"]
], { x: 1, y: 1, w: 8, h: 2, border: { pt: 1, color: "999999" }, fill: { color: "F1F1F1" } });
```

---

## Slide Backgrounds

```javascript
slide.background = { color: "F1F1F1" };
slide.background = { path: "https://example.com/bg.jpg" };
slide.background = { data: "image/png;base64,iVBORw0KGgo..." };
```

---

## Icons (react-icons → PNG)

```javascript
const React = require("react");
const ReactDOMServer = require("react-dom/server");
const sharp = require("sharp");
const { FaCheckCircle } = require("react-icons/fa");

function renderIconSvg(IconComponent, color = "#000000", size = 256) {
  return ReactDOMServer.renderToStaticMarkup(
    React.createElement(IconComponent, { color, size: String(size) })
  );
}

async function iconToBase64Png(IconComponent, color, size = 256) {
  const svg = renderIconSvg(IconComponent, color, size);
  const pngBuffer = await sharp(Buffer.from(svg)).png().toBuffer();
  return "image/png;base64," + pngBuffer.toString("base64");
}

const iconData = await iconToBase64Png(FaCheckCircle, "#4472C4", 256);
slide.addImage({ data: iconData, x: 1, y: 1, w: 0.5, h: 0.5 });
```

Install: `npm install -g react-icons react react-dom sharp`

---

## Common Pitfalls

1. **NEVER use "#" with hex colors** — `color: "FF0000"` ✅ / `color: "#FF0000"` ❌
2. **NEVER encode opacity in hex color strings** — use `opacity` property instead
3. **Use `bullet: true`** — NEVER unicode "•"
4. **Use `breakLine: true`** between array items
5. **Avoid `lineSpacing` with bullets** — use `paraSpaceAfter` instead
6. **Each presentation needs fresh instance** — don't reuse `pptxgen()` objects
7. **NEVER reuse option objects across calls** — PptxGenJS mutates objects in-place. Use factory functions.
8. **Don't use `ROUNDED_RECTANGLE` with accent borders** — use `RECTANGLE` instead

---

## Design Principles

**Only use these when building from raw PptxGenJS code.** If using a template, follow the template's style.

### For Each Slide

- Every slide needs a visual element — image, chart, icon, or shape
- Vary layouts across slides (two-column, icon rows, grids, half-bleed image)
- Large stat callouts (60–72pt) for key numbers

### Typography

| Element | Size |
|---------|------|
| Slide title | 36-44pt bold |
| Section header | 20-24pt bold |
| Body text | 14-16pt |
| Captions | 10-12pt muted |

### Spacing

- 0.5" minimum margins
- 0.3–0.5" between content blocks

### Avoid

- Don't repeat the same layout
- Don't center body text (left-align paragraphs)
- Don't default to blue — pick topic-appropriate colors
- Don't create text-only slides
- **NEVER use accent lines under titles** — hallmark of AI-generated slides
