---
name: pptx
description: >-
  Use this skill any time a .pptx file is involved — creating, editing, reading, or analyzing presentations.
  Trigger on: "做PPT" "生成幻灯片" "做个演示" "pitch deck" "pptx" "slides" "presentation"
  "演示文稿" "做一页" "deck" "编辑PPT" "修改PPT" or any deck/slide request.
---

# PPTX Skill

## Workflow Selection

| Situation | Action |
|-----------|--------|
| User attaches or references a `.pptx` file | Read [editing.md](editing.md) — edit the template directly |
| No `.pptx` provided — create from scratch | Read [creating.md](creating.md) — use PptxGenJS templates |
| Read/analyze an existing presentation | See Reading Content below |

**When in doubt:** If ANY `.pptx` file is attached or mentioned, use the editing workflow. The creating workflow is ONLY for when no existing presentation exists at all.

## Path Convention

All scripts are located at `.kiro/skills/pptx/scripts/`. Use relative paths from workspace root:

```bash
python .kiro/skills/pptx/scripts/<script_name>.py [args]
```

Templates are at `.kiro/skills/pptx/templates/`.
Taxonomy file is at `.kiro/skills/pptx/template_taxonomy.json`.

---

## Reading Content

```bash
# Text extraction
python -m markitdown presentation.pptx

# Visual overview
python .kiro/skills/pptx/scripts/thumbnail.py presentation.pptx

# Raw XML
python .kiro/skills/pptx/scripts/office/unpack.py presentation.pptx unpacked/

# Structured inspection (placeholders, text, images, media)
python .kiro/skills/pptx/scripts/inspect_slide.py unpacked/ --theme --media
```

---

## QA (Required)

**Assume there are problems. Your job is to find them.**

### Overlap Check & Auto-Fix (Before Packing)

```bash
# Check and auto-fix
python .kiro/skills/pptx/scripts/check_overlaps.py unpacked/ --fix

# Check only
python .kiro/skills/pptx/scripts/check_overlaps.py unpacked/
```

### Content QA

```bash
python -m markitdown output.pptx
```

Check for missing content, typos, wrong order.

**When using templates, check for leftover placeholder text:**

```bash
python -m markitdown output.pptx | Select-String -Pattern "xxxx|lorem|ipsum"
```

### Automated Checks

```bash
python .kiro/skills/pptx/scripts/detect_fonts.py output.pptx
python .kiro/skills/pptx/scripts/office/unpack.py output.pptx unpacked/
python .kiro/skills/pptx/scripts/check_overlaps.py unpacked/ --fix
python .kiro/skills/pptx/scripts/clean.py unpacked/
python .kiro/skills/pptx/scripts/office/pack.py unpacked/ final_output.pptx --original output.pptx
```

### Visual QA

Convert slides to images, then look for:

- Overlapping elements
- Text overflow or cut off
- Elements too close (< 0.3" gaps)
- Insufficient margin from slide edges (< 0.5")
- Low-contrast text/icons
- Leftover placeholder content
- Style consistency

```bash
# Convert to images for inspection
python .kiro/skills/pptx/scripts/render_slides.py output.pptx
```

### Verification Loop

1. Generate slides → Convert to images → Inspect
2. Run `detect_fonts.py` and `check_overlaps.py`
3. List issues found
4. Fix issues
5. Re-verify affected slides
6. Repeat until clean

---

## Images

### Image Aspect Ratio Rule

**Always preserve the original aspect ratio of images.**

### Image Sources (PptxGenJS)

```javascript
slide.addImage({ path: "images/chart.png", x: 1, y: 1, w: 5, h: 3 });
slide.addImage({ path: "https://example.com/image.jpg", x: 1, y: 1, w: 5, h: 3 });
slide.addImage({ data: "image/png;base64,iVBORw0KGgo...", x: 1, y: 1, w: 5, h: 3 });
```

### Image Sizing Modes

```javascript
{ sizing: { type: 'contain', w: 4, h: 3 } }  // fit inside
{ sizing: { type: 'cover', w: 4, h: 3 } }    // fill area (may crop)
{ sizing: { type: 'crop', x: 0.5, y: 0.5, w: 2, h: 2 } }
```

---

## Dependencies

```bash
pip install "markitdown[pptx]" Pillow pdf2image python-pptx numpy defusedxml
npm install -g pptxgenjs
```

**System packages** (needed for rendering):
- LibreOffice (`soffice`) — PDF conversion
- Poppler (`pdftoppm`) — PDF to images
- fontconfig (`fc-list`) — font detection

---

## Detailed Workflows

- **Creating from scratch**: See [creating.md](creating.md)
- **Editing existing .pptx**: See [editing.md](editing.md)
