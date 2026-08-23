# Editing Presentations

## When Given a Template — EDIT IT, NEVER Recreate

**If a template .pptx is attached, you MUST edit it directly using unpack/edit/pack. NEVER use pptxgenjs to create from scratch.**

### Critical Rules

1. **ALWAYS edit the template file, even when the content doesn't match.** The visual style IS the template.
2. **NEVER create from scratch with pptxgenjs when a template exists.**
3. **NEVER use `sed`, `awk`, `perl`, or raw regex shell commands on XML files.** Use the provided Python scripts.
4. **Do NOT adapt the template's colors or fonts to "match" the topic.** Keep the template's visual style exactly.
5. **Reuse, don't recreate.** Use `inspect_slide.py --media` to see existing assets.
6. **Content mismatch is expected.** Replace ALL text while preserving every visual element.
7. **Replace EVERY piece of template text.** Use `replace_nth_text.py --list` to find leftovers.

### Style Matching Rule

Always match the existing style — colors, fonts, layout patterns, background treatment.

### Font Preservation Rule

Always preserve original font declarations in the XML. Never substitute fonts just because they aren't installed locally.

---

## Workflow Summary

**unpack → overview → pick slides → delete extras → replace ALL text → verify no leftovers → fix overlaps → clean → pack**

```bash
python .kiro/skills/pptx/scripts/office/unpack.py template.pptx unpacked/
python .kiro/skills/pptx/scripts/inspect_slide.py unpacked/ --summary --theme --media
python .kiro/skills/pptx/scripts/inspect_slide.py unpacked/ppt/slides/slide1.xml
python .kiro/skills/pptx/scripts/delete_slide.py unpacked/ --keep 1 4 8
python .kiro/skills/pptx/scripts/replace_text.py unpacked/ppt/slides/slide1.xml --ph ctrTitle --text "New Title"

# Verify no template text remains
python .kiro/skills/pptx/scripts/replace_nth_text.py unpacked/ppt/slides/slide1.xml --list

# Fix overlaps and pack
python .kiro/skills/pptx/scripts/check_overlaps.py unpacked/ --fix
python .kiro/skills/pptx/scripts/clean.py unpacked/
python .kiro/skills/pptx/scripts/office/pack.py unpacked/ output.pptx --original template.pptx
```

---

## Template-Based Workflow (Detailed)

1. **Unpack**: `python .kiro/skills/pptx/scripts/office/unpack.py template.pptx unpacked/`

2. **Get overview**:
   ```bash
   python .kiro/skills/pptx/scripts/inspect_slide.py unpacked/ --summary --theme --media
   python .kiro/skills/pptx/scripts/thumbnail.py template.pptx
   python .kiro/skills/pptx/scripts/media_grid.py unpacked/
   python .kiro/skills/pptx/scripts/inspect_slide.py unpacked/ppt/slides/slide1.xml
   ```

3. **Plan slide mapping** — decide which template slides to KEEP and what content goes on each.

4. **Build presentation**:
   ```bash
   python .kiro/skills/pptx/scripts/delete_slide.py unpacked/ --keep 1 2 4 8
   python .kiro/skills/pptx/scripts/add_slide.py unpacked/ slide2.xml --after slide4.xml
   ```

5. **Replace placeholder text**:
   ```bash
   python .kiro/skills/pptx/scripts/replace_text.py unpacked/ppt/slides/slide1.xml --ph ctrTitle --text "New Title"
   python .kiro/skills/pptx/scripts/replace_text.py unpacked/ppt/slides/slide1.xml --ph subTitle --text "Line 1\nLine 2"

   # Custom formatting per run
   python .kiro/skills/pptx/scripts/replace_text.py unpacked/ppt/slides/slide1.xml --ph ctrTitle --runs "[{\"text\": \"Title\", \"size\": 4800, \"color\": \"accent1\"}]"

   # Match by text content
   python .kiro/skills/pptx/scripts/replace_text.py unpacked/ppt/slides/slide1.xml --match "Christmas" --text "Holiday Special"
   ```

6. **Clean**: `python .kiro/skills/pptx/scripts/clean.py unpacked/`

7. **Pack**: `python .kiro/skills/pptx/scripts/office/pack.py unpacked/ output.pptx --original template.pptx`

---

## Scripts Reference

| Script | Purpose |
|--------|---------|
| `inspect_slide.py` | Inspect placeholders, text, images, media |
| `check_overlaps.py` | Lint for overlapping shapes |
| `replace_text.py` | Replace placeholder text preserving formatting |
| `replace_nth_text.py` | Replace substrings or duplicate text |
| `resize_shape.py` | Resize or reposition shapes |
| `delete_slide.py` | Delete slides from presentation |
| `add_slide.py` | Duplicate slide or create from layout |
| `office/unpack.py` | Extract and pretty-print PPTX |
| `clean.py` | Remove orphaned files |
| `media_grid.py` | Visual thumbnail grid of all media |
| `office/pack.py` | Repack with validation |
| `thumbnail.py` | Create visual grid of slides |
| `render_slides.py` | Convert slides to images for QA |

All scripts are at `.kiro/skills/pptx/scripts/`.

---

## replace_text.py

```bash
# Simple replacement
python .kiro/skills/pptx/scripts/replace_text.py slide.xml --ph ctrTitle --text "New Title"

# Multi-line
python .kiro/skills/pptx/scripts/replace_text.py slide.xml --ph ctrTitle --text "Line 1\nLine 2"

# Custom formatting (JSON runs)
python .kiro/skills/pptx/scripts/replace_text.py slide.xml --ph ctrTitle --runs "[{\"text\": \"Title\", \"size\": 4800, \"color\": \"accent1\", \"font\": \"Satisfy\"}]"

# Match by text content
python .kiro/skills/pptx/scripts/replace_text.py slide.xml --match "Christmas" --text "Holiday"

# Dry run
python .kiro/skills/pptx/scripts/replace_text.py slide.xml --ph ctrTitle --text "New" --dry-run
```

**JSON run properties** (all optional):

| Property | Type | Example | Notes |
|----------|------|---------|-------|
| `text` | string | `"Title"` | The text content |
| `br` | boolean | `true` | Line break |
| `size` | integer | `4800` | Hundredths of a point (4800 = 48pt) |
| `color` | string | `"accent1"` or `"#FF0000"` | Scheme name or hex |
| `font` | string | `"Satisfy"` | Typeface name |
| `bold` | boolean | `true` | Bold weight |
| `italic` | boolean | `true` | Italic style |

---

## replace_nth_text.py

```bash
# List all text elements
python .kiro/skills/pptx/scripts/replace_nth_text.py slide.xml --list

# Substring replacement (preserves surrounding text)
python .kiro/skills/pptx/scripts/replace_nth_text.py slide.xml --find "YYYY" --all --text "2015"

# Replace nth occurrence (1-based)
python .kiro/skills/pptx/scripts/replace_nth_text.py slide.xml --find "Chapter Title" --nth 1 --text "NEW TITLE" --full

# Target by absolute index (0-based)
python .kiro/skills/pptx/scripts/replace_nth_text.py slide.xml --index 3 --text "NEW TEXT"
```

### When to use which tool

| Scenario | Tool |
|----------|------|
| Replace text in a placeholder shape | `replace_text.py --ph` |
| Replace text matched by content (whole shape) | `replace_text.py --match` |
| Swap a substring preserving context | `replace_nth_text.py --find --text` |
| Duplicate text in same shape | `replace_nth_text.py --find --nth --full` |
| Target a specific `<a:t>` element | `replace_nth_text.py --index` |
| Bulk replace all occurrences | `replace_nth_text.py --find --all` |

---

## resize_shape.py

```bash
python .kiro/skills/pptx/scripts/resize_shape.py slide.xml --list
python .kiro/skills/pptx/scripts/resize_shape.py slide.xml --ph body --height 3.5
python .kiro/skills/pptx/scripts/resize_shape.py slide.xml --ph body --dh 0.5
python .kiro/skills/pptx/scripts/resize_shape.py slide.xml --ph body --dy 0.3
```

---

## Preventing Text Overflow

1. **Use `--autofit`** when replacing text
2. **Resize the text box** with `resize_shape.py`
3. **Reduce font size** in `--runs`
4. **Adapt content to fit** — shorten text
5. **Reduce line spacing** in XML
6. **Set autofit manually** — add `<a:normAutofit/>` inside `<a:bodyPr>`

---

## Common Pitfalls

- **Template slots ≠ Source items**: If template has 4 items but source has 3, delete the extra group entirely.
- **Multi-item content**: Create separate `<a:p>` elements, never concatenate into one string.
- **Smart quotes**: Use XML entities (`&#x201C;` `&#x201D;` `&#x2018;` `&#x2019;`).
- **Shell quoting**: If text contains `$`, use single quotes to prevent expansion.
- **Whitespace**: Use `xml:space="preserve"` on `<a:t>` with leading/trailing spaces.
- **XML parsing**: Use `defusedxml.minidom`, not `xml.etree.ElementTree` (corrupts namespaces).
