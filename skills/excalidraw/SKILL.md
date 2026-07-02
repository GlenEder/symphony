---
name: excalidraw
description: Generate Excalidraw diagrams by writing .excalidraw JSON files.
  Covers all element types, arrow bindings, text containers, frames, and file
  serialization. Use when the user asks for diagrams, flowcharts, architecture
  drawings, wireframes, concept sketches, mind maps, or any visual
  representation that could be drawn in Excalidraw.
---

# Excalidraw Diagram Generation

## Purpose

This skill teaches you to generate Excalidraw-compatible `.excalidraw` files by
writing raw JSON. The output files can be opened directly in
[excalidraw.com](https://excalidraw.com) or the Excalidraw desktop app.

No npm packages or external tools are needed — you write the JSON directly.

## File Format

Every `.excalidraw` file is a JSON document with this top-level structure:

```json
{
  "type": "excalidraw",
  "version": 2,
  "source": "https://excalidraw.com",
  "elements": [],
  "appState": {}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | `"excalidraw"` | Format discriminator. Must be exactly `"excalidraw"`. |
| `version` | `2` | Schema version. Currently always `2`. |
| `source` | `string` | Origin URL. Use `"https://excalidraw.com"` or `"https://excalidraw.maki"`. |
| `elements` | `array` | Array of element objects (the drawing content). |
| `appState` | `object` | Editor state (canvas config, preferences). Can be minimal. |

Save the file with a `.excalidraw` extension.

## Common Element Properties

Every element shares these base properties. Only `id`, `type`, `x`, `y`,
`width`, `height`, `seed`, and `versionNonce` are strictly required, but you
should set the common styling fields for a good result.

```json
{
  "id": "rect-1",
  "type": "rectangle",
  "x": 100,
  "y": 100,
  "width": 200,
  "height": 120,
  "angle": 0,
  "strokeColor": "#1e1e1e",
  "backgroundColor": "transparent",
  "fillStyle": "solid",
  "strokeWidth": 2,
  "strokeStyle": "solid",
  "roughness": 1,
  "opacity": 100,
  "seed": 1000001,
  "version": 1,
  "versionNonce": 1,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false
}
```

### Property Table — Common

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `id` | `string` | — | Unique ID. Use descriptive names like `"rect-header"`, `"arrow-main"`. |
| `type` | `string` | — | Element type discriminator (see types below). |
| `x`, `y` | `number` | — | Top-left corner coordinates in pixels. |
| `width`, `height` | `number` | — | Bounding-box dimensions in pixels. |
| `angle` | `number` | `0` | Rotation in radians. |
| `strokeColor` | `string` | `"#1e1e1e"` | Border/line color as hex. |
| `backgroundColor` | `string` | `"transparent"` | Fill color as hex or `"transparent"`. |
| `fillStyle` | `string` | `"solid"` | `"solid"`, `"hachure"`, `"cross-hatch"`, `"dots"`, `"zigzag"`, `"zigzag-line"`. |
| `strokeWidth` | `number` | `2` | Line thickness: `1` (thin), `2` (normal), `4` (thick). |
| `strokeStyle` | `string` | `"solid"` | `"solid"`, `"dashed"`, `"dotted"`. |
| `roughness` | `number` | `1` | Hand-drawn effect: `0` (perfect), `1` (slight), `2` (cartoonist). |
| `opacity` | `number` | `100` | 0–100. `100` = fully opaque. |
| `seed` | `number` | — | Random seed for hand-drawn rendering. Use a unique integer per element. |
| `version` | `number` | — | Start at `1` and increment for each change. |
| `versionNonce` | `number` | — | Random nonce (any unique integer). Increment when mutating. |
| `isDeleted` | `boolean` | `false` | Soft-delete flag. Always `false` for active elements. |
| `groupIds` | `array` | `[]` | Array of group IDs this element belongs to. |
| `boundElements` | `array` | `null` | Array of `{ id, type }` objects for elements bound to this one (e.g. arrows pointing to it). |
| `link` | `string` | `null` | URL attached to the element. |
| `locked` | `boolean` | `false` | Prevents editing when `true`. |

## Element Types

### 1. Rectangle

```json
{
  "id": "rect-1",
  "type": "rectangle",
  "x": 50,
  "y": 50,
  "width": 200,
  "height": 100,
  "angle": 0,
  "strokeColor": "#1e1e1e",
  "backgroundColor": "transparent",
  "fillStyle": "solid",
  "strokeWidth": 2,
  "strokeStyle": "solid",
  "roughness": 1,
  "opacity": 100,
  "seed": 1000001,
  "version": 1,
  "versionNonce": 100001,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "roundness": { "type": 3 }
}
```

**Type-specific properties:**

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `roundness` | `object` | `null` | Corner rounding. `{ "type": 3 }` for rounded, `null` for sharp. |

### 2. Diamond

```json
{
  "id": "diamond-1",
  "type": "diamond",
  "x": 300,
  "y": 50,
  "width": 150,
  "height": 150,
  "seed": 1000002,
  "version": 1,
  "versionNonce": 100002,
  "strokeColor": "#1e1e1e",
  "backgroundColor": "#fff3bf",
  "fillStyle": "solid",
  "strokeWidth": 2,
  "strokeStyle": "solid",
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false
}
```

No type-specific properties beyond the common set.

### 3. Ellipse

```json
{
  "id": "ellipse-1",
  "type": "ellipse",
  "x": 500,
  "y": 50,
  "width": 150,
  "height": 120,
  "seed": 1000003,
  "version": 1,
  "versionNonce": 100003,
  "strokeColor": "#1e1e1e",
  "backgroundColor": "#a5d8ff",
  "fillStyle": "solid",
  "strokeWidth": 2,
  "strokeStyle": "solid",
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false
}
```

No type-specific properties beyond the common set.

### 4. Arrow

```json
{
  "id": "arrow-1",
  "type": "arrow",
  "x": 250,
  "y": 100,
  "width": 100,
  "height": 0,
  "seed": 1000004,
  "version": 1,
  "versionNonce": 100004,
  "strokeColor": "#1e1e1e",
  "strokeWidth": 2,
  "strokeStyle": "solid",
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "points": [
    [0, 0],
    [100, 0]
  ],
  "startArrowhead": null,
  "endArrowhead": "arrow",
  "startBinding": null,
  "endBinding": null
}
```

**Type-specific properties:**

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `points` | `array` | — | Array of `[x, y]` pairs defining the path. First point is always `[0, 0]`. Subsequent points are relative offsets. |
| `startArrowhead` | `string` | `null` | `null`, `"arrow"`, `"circle"`, `"triangle"`, `"bar"`. |
| `endArrowhead` | `string` | `"arrow"` | Same options as `startArrowhead`. |
| `startBinding` | `object` | `null` | Binding to a start element (see Arrow Bindings). |
| `endBinding` | `object` | `null` | Binding to an end element (see Arrow Bindings). |

**Points explained:** The `x` and `y` of the arrow element define the origin.
The `points` array defines the path relative to that origin. For a simple
horizontal arrow of length 200:

```json
{
  "x": 100, "y": 200, "width": 200, "height": 0,
  "points": [[0, 0], [200, 0]]
}
```

For a bent arrow (right then down):

```json
{
  "x": 100, "y": 100, "width": 150, "height": 80,
  "points": [[0, 0], [100, 0], [100, 80], [150, 80]]
}
```

### 5. Line

Same structure as Arrow but with `"type": "line"`. Lines typically have no
arrowheads.

```json
{
  "id": "line-1",
  "type": "line",
  "x": 100,
  "y": 300,
  "width": 200,
  "height": 50,
  "seed": 1000005,
  "version": 1,
  "versionNonce": 100005,
  "strokeColor": "#2f9e44",
  "strokeWidth": 2,
  "strokeStyle": "dashed",
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "points": [
    [0, 0],
    [100, -50],
    [200, 0]
  ],
  "startArrowhead": null,
  "endArrowhead": null,
  "startBinding": null,
  "endBinding": null
}
```

### 6. Freedraw

Free-hand drawing path. Used for sketches and annotations.

```json
{
  "id": "freedraw-1",
  "type": "freedraw",
  "x": 100,
  "y": 100,
  "width": 50,
  "height": 50,
  "seed": 1000006,
  "version": 1,
  "versionNonce": 100006,
  "strokeColor": "#e03131",
  "strokeWidth": 2,
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "pressures": [0.5, 0.6, 0.7],
  "points": [
    [0, 0, 0.5],
    [10, 15, 0.6],
    [25, 30, 0.7],
    [40, 45, 0.65]
  ]
}
```

**Type-specific properties:**

| Property | Type | Description |
|----------|------|-------------|
| `pressures` | `number[]` | Pen pressure at each point (0–1). |
| `points` | `array` | Array of `[x, y, pressure]` triplets. |

### 7. Text

```json
{
  "id": "text-1",
  "type": "text",
  "x": 100,
  "y": 100,
  "width": 200,
  "height": 30,
  "seed": 1000007,
  "version": 1,
  "versionNonce": 100007,
  "strokeColor": "#1e1e1e",
  "backgroundColor": "transparent",
  "fillStyle": "solid",
  "strokeWidth": 2,
  "roughness": 1,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "text": "Hello World",
  "originalText": "Hello World",
  "rawText": "Hello World",
  "fontSize": 20,
  "fontFamily": 1,
  "textAlign": "left",
  "verticalAlign": "top",
  "containerId": null,
  "autoResize": true,
  "lineHeight": 1.25
}
```

**Type-specific properties:**

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `text` | `string` | `""` | The display text. May wrap. |
| `originalText` | `string` | — | The unwrapped original text. |
| `rawText` | `string` | — | Text with original formatting. |
| `fontSize` | `number` | `20` | Font size in pixels. |
| `fontFamily` | `number` | `1` | `1`=Virgil (hand-drawn), `2`=Helvetica, `3`=Cascadia (mono), `4`=Local, `5`=Excalifont. |
| `textAlign` | `string` | `"left"` | `"left"`, `"center"`, `"right"`. |
| `verticalAlign` | `string` | `"top"` | `"top"`, `"middle"`, `"bottom"`. |
| `containerId` | `string` | `null` | ID of the container shape if this text is inside one. |
| `autoResize` | `boolean` | `true` | Whether text box auto-resizes to fit content. |
| `lineHeight` | `number` | `1.25` | Line height multiplier. |

**Sizing hint:** When `autoResize` is `true`, set `width` and `height` to
rough estimates — Excalidraw adjusts them at render time. For single-line 20px
text, `width` ≈ 12 × character count.

### 8. Image

```json
{
  "id": "image-1",
  "type": "image",
  "x": 100,
  "y": 100,
  "width": 200,
  "height": 150,
  "seed": 1000008,
  "version": 1,
  "versionNonce": 100008,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "fileId": null,
  "status": "saved",
  "scale": [1, 1],
  "crop": null
}
```

**Note:** Image elements require a matching entry in the `files` map with
base64-encoded image data. This is impractical for agent-generated files. Use
`"fileId": null` as a placeholder, or skip image elements entirely.

### 9. Frame

A frame groups elements into a named container. It positions itself to
encompass all its children.

```json
{
  "id": "frame-1",
  "type": "frame",
  "x": 30,
  "y": 30,
  "width": 500,
  "height": 300,
  "seed": 1000009,
  "version": 1,
  "versionNonce": 100009,
  "strokeColor": "#888888",
  "backgroundColor": "transparent",
  "fillStyle": "solid",
  "strokeWidth": 1,
  "strokeStyle": "solid",
  "roughness": 0,
  "opacity": 100,
  "isDeleted": false,
  "groupIds": [],
  "boundElements": null,
  "link": null,
  "locked": false,
  "name": "My Frame",
  "children": ["rect-1", "text-1", "arrow-1"]
}
```

**Type-specific properties:**

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `name` | `string` | — | Display name shown on the frame label. |
| `children` | `array` | `[]` | Array of element IDs that belong to this frame. |

The frame's `x`, `y`, `width`, `height` should encompass all children with
padding (~20px on each side).

## Arrow Bindings

Arrow bindings connect arrows to shapes so the arrow moves with the shape.
Binding requires two sides: the **arrow** side and the **target shape** side.

### On the target shape

Add a `boundElements` entry pointing to the arrow:

```json
{
  "id": "rect-1",
  "type": "rectangle",
  "x": 100, "y": 100,
  "width": 150, "height": 80,
  "boundElements": [
    { "id": "arrow-1", "type": "arrow" }
  ]
}
```

### On the arrow

Set `startBinding` and/or `endBinding` to reference the target:

```json
{
  "id": "arrow-1",
  "type": "arrow",
  "x": 250, "y": 140,
  "width": 150, "height": 0,
  "points": [[0, 0], [150, 0]],
  "startBinding": {
    "elementId": "rect-1",
    "gap": 8,
    "focus": 0
  },
  "endBinding": {
    "elementId": "ellipse-1",
    "gap": 8,
    "focus": 0
  }
}
```

| Binding property | Type | Description |
|------------------|------|-------------|
| `elementId` | `string` | ID of the shape this arrow connects to. |
| `gap` | `number` | Pixel gap between arrow tip and shape edge. Use `8` as default. |
| `focus` | `number` | Attachment point on the shape edge: `0` = center, `-1` = far left/top, `1` = far right/bottom. |

### Complete binding example

```json
[
  {
    "id": "rect-1",
    "type": "rectangle",
    "x": 100, "y": 200,
    "width": 150, "height": 80,
    "seed": 1000011,
    "version": 1,
    "versionNonce": 100011,
    "strokeColor": "#1e1e1e",
    "backgroundColor": "#a5d8ff",
    "fillStyle": "solid",
    "strokeWidth": 2,
    "roughness": 1,
    "opacity": 100,
    "isDeleted": false,
    "groupIds": [],
    "boundElements": [
      { "id": "arrow-1", "type": "arrow" }
    ],
    "link": null,
    "locked": false
  },
  {
    "id": "ellipse-1",
    "type": "ellipse",
    "x": 400, "y": 190,
    "width": 120, "height": 100,
    "seed": 1000012,
    "version": 1,
    "versionNonce": 100012,
    "strokeColor": "#1e1e1e",
    "backgroundColor": "#b2f2bb",
    "fillStyle": "solid",
    "strokeWidth": 2,
    "roughness": 1,
    "opacity": 100,
    "isDeleted": false,
    "groupIds": [],
    "boundElements": [
      { "id": "arrow-1", "type": "arrow" }
    ],
    "link": null,
    "locked": false
  },
  {
    "id": "arrow-1",
    "type": "arrow",
    "x": 250, "y": 240,
    "width": 150, "height": 0,
    "seed": 1000013,
    "version": 1,
    "versionNonce": 100013,
    "strokeColor": "#1e1e1e",
    "strokeWidth": 2,
    "roughness": 1,
    "opacity": 100,
    "isDeleted": false,
    "groupIds": [],
    "boundElements": null,
    "link": null,
    "locked": false,
    "points": [[0, 0], [150, 0]],
    "startArrowhead": null,
    "endArrowhead": "arrow",
    "startBinding": {
      "elementId": "rect-1",
      "focus": 0,
      "gap": 8
    },
    "endBinding": {
      "elementId": "ellipse-1",
      "focus": 0,
      "gap": 8
    }
  }
]
```

## Text Containers

To embed text inside a shape, create a separate `text` element whose
`containerId` references the shape's ID.

```json
[
  {
    "id": "rect-box",
    "type": "rectangle",
    "x": 100, "y": 100,
    "width": 200, "height": 80,
    "seed": 1000014,
    "version": 1,
    "versionNonce": 100014,
    "strokeColor": "#1e1e1e",
    "backgroundColor": "#fff3bf",
    "fillStyle": "solid",
    "strokeWidth": 2,
    "roughness": 1,
    "opacity": 100,
    "isDeleted": false,
    "groupIds": [],
    "boundElements": [
      { "id": "text-label", "type": "text" }
    ],
    "link": null,
    "locked": false
  },
  {
    "id": "text-label",
    "type": "text",
    "x": 120, "y": 120,
    "width": 160, "height": 40,
    "seed": 1000015,
    "version": 1,
    "versionNonce": 100015,
    "strokeColor": "#1e1e1e",
    "backgroundColor": "transparent",
    "fillStyle": "solid",
    "strokeWidth": 2,
    "roughness": 1,
    "opacity": 100,
    "isDeleted": false,
    "groupIds": [],
    "boundElements": null,
    "link": null,
    "locked": false,
    "text": "Hello inside box",
    "originalText": "Hello inside box",
    "rawText": "Hello inside box",
    "fontSize": 20,
    "fontFamily": 1,
    "textAlign": "center",
    "verticalAlign": "middle",
    "containerId": "rect-box",
    "autoResize": true,
    "lineHeight": 1.25
  }
]
```

**Key points:**
- The shape's `boundElements` must include `{ "id": "<text-id>", "type": "text" }`.
- The text element's `containerId` must match the shape's `id`.
- Position the text element inside the shape bounds; with `autoResize: true`,
  Excalidraw adjusts it.
- Set `textAlign` and `verticalAlign` to control text placement inside the
  container.

## appState Reference

Minimal `appState` for a valid file:

```json
{
  "appState": {
    "gridSize": null,
    "viewBackgroundColor": "#ffffff"
  }
}
```

Common `appState` options:

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `gridSize` | `number` | `null` | Grid spacing in px. `null` = grid off, `20` = grid on. |
| `viewBackgroundColor` | `string` | `"#ffffff"` | Canvas background color. |
| `zenModeEnabled` | `boolean` | `false` | Hides UI chrome. |
| `currentItemStrokeColor` | `string` | — | Default stroke color for new elements. |
| `currentItemBackgroundColor` | `string` | — | Default fill color for new elements. |
| `currentItemFillStyle` | `string` | — | Default fill style. |
| `currentItemStrokeWidth` | `number` | — | Default stroke width. |
| `currentItemRoughness` | `number` | — | Default roughness. |
| `currentItemOpacity` | `number` | — | Default opacity. |
| `currentItemFontFamily` | `number` | — | Default font family. |
| `currentItemFontSize` | `number` | — | Default font size. |
| `currentItemStrokeStyle` | `string` | — | Default stroke style. |

## Workflow

Follow these steps to generate an Excalidraw diagram:

### Step 1: Plan the diagram

Ask the user what they want to diagram. Determine:
- What shapes are needed (rectangles for processes, diamonds for decisions,
  ellipses for start/end, arrows for flows)
- How they connect (which arrows go where)
- Any labels needed (text elements, text containers)
- Color scheme / styling preferences

### Step 2: List elements with coordinates

For each element, decide:
- Type
- Position (`x`, `y` relative to the canvas)
- Size (`width`, `height`)
- Style (colors, stroke, fill)

**Coordinate system:** Origin `(0, 0)` is top-left. `x` increases right, `y`
increases down.

**Spacing guideline:** Leave 40–80px between shapes. Standard node sizes:
- Process box: 160×80px
- Decision diamond: 120×120px
- Terminator (ellipse): 160×60px

### Step 3: Wire up bindings

If arrows connect to shapes, add `boundElements` on the shapes and matching
`startBinding` / `endBinding` on the arrows. If text lives inside a shape, add
`containerId` on the text and `boundElements` on the shape.

### Step 4: Compute dimensions

Set `width` and `height` on each element to its bounding box dimensions. For
arrows, `width`/`height` should match the last point's offset.

### Step 5: Serialize to JSON

Wrap all elements in the `.excalidraw` envelope:

```json
{
  "type": "excalidraw",
  "version": 2,
  "source": "https://excalidraw.com",
  "elements": [ /* all elements */ ],
  "appState": {
    "gridSize": null,
    "viewBackgroundColor": "#ffffff"
  }
}
```

### Step 6: Validate

Checklist:
- Every element has an `id`, `type`, `x`, `y`, `width`, `height`, `seed`,
  `version`, `versionNonce`
- `seed` values are unique per element
- Arrow `points` start with `[0, 0]`
- `boundElements` references use correct IDs
- `containerId` references match the target element's `id`
- JSON is valid (no trailing commas)

### Step 7: Write the file

Write the JSON to a file with a `.excalidraw` extension. Inform the user they
can open it at [excalidraw.com](https://excalidraw.com) or in the desktop app.

## Complete Examples

### Example 1: Simple Flowchart

A flowchart with a start terminator, process, decision, and end terminator.

```json
{
  "type": "excalidraw",
  "version": 2,
  "source": "https://excalidraw.com",
  "elements": [
    {
      "id": "start",
      "type": "ellipse",
      "x": 220,
      "y": 20,
      "width": 160,
      "height": 60,
      "seed": 1000101,
      "version": 1,
      "versionNonce": 1000101,
      "strokeColor": "#1e1e1e",
      "backgroundColor": "#b2f2bb",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-start-process", "type": "arrow" }
      ],
      "link": null,
      "locked": false
    },
    {
      "id": "process",
      "type": "rectangle",
      "x": 220,
      "y": 140,
      "width": 160,
      "height": 80,
      "seed": 1000102,
      "version": 1,
      "versionNonce": 1000102,
      "strokeColor": "#1e1e1e",
      "backgroundColor": "#a5d8ff",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-start-process", "type": "arrow" },
        { "id": "arrow-process-decision", "type": "arrow" },
        { "id": "text-process-label", "type": "text" }
      ],
      "link": null,
      "locked": false,
      "roundness": { "type": 3 }
    },
    {
      "id": "text-process-label",
      "type": "text",
      "x": 250,
      "y": 165,
      "width": 100,
      "height": 30,
      "seed": 1000103,
      "version": 1,
      "versionNonce": 1000103,
      "strokeColor": "#1e1e1e",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "text": "Process",
      "originalText": "Process",
      "rawText": "Process",
      "fontSize": 20,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "containerId": "process",
      "autoResize": true,
      "lineHeight": 1.25
    },
    {
      "id": "decision",
      "type": "diamond",
      "x": 240,
      "y": 280,
      "width": 120,
      "height": 120,
      "seed": 1000104,
      "version": 1,
      "versionNonce": 1000104,
      "strokeColor": "#1e1e1e",
      "backgroundColor": "#fff3bf",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-process-decision", "type": "arrow" },
        { "id": "arrow-decision-yes", "type": "arrow" },
        { "id": "arrow-decision-no", "type": "arrow" }
      ],
      "link": null,
      "locked": false
    },
    {
      "id": "end",
      "type": "ellipse",
      "x": 520,
      "y": 310,
      "width": 160,
      "height": 60,
      "seed": 1000105,
      "version": 1,
      "versionNonce": 1000105,
      "strokeColor": "#1e1e1e",
      "backgroundColor": "#ffc9c9",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-decision-yes", "type": "arrow" }
      ],
      "link": null,
      "locked": false
    },
    {
      "id": "arrow-start-process",
      "type": "arrow",
      "x": 300,
      "y": 80,
      "width": 0,
      "height": 60,
      "seed": 1000106,
      "version": 1,
      "versionNonce": 1000106,
      "strokeColor": "#1e1e1e",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [0, 60]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "start",
        "focus": 0,
        "gap": 8
      },
      "endBinding": {
        "elementId": "process",
        "focus": 0,
        "gap": 8
      }
    },
    {
      "id": "arrow-process-decision",
      "type": "arrow",
      "x": 300,
      "y": 220,
      "width": 0,
      "height": 60,
      "seed": 1000107,
      "version": 1,
      "versionNonce": 1000107,
      "strokeColor": "#1e1e1e",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [0, 60]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "process",
        "focus": 0,
        "gap": 8
      },
      "endBinding": {
        "elementId": "decision",
        "focus": 0,
        "gap": 8
      }
    },
    {
      "id": "arrow-decision-yes",
      "type": "arrow",
      "x": 360,
      "y": 340,
      "width": 160,
      "height": 0,
      "seed": 1000108,
      "version": 1,
      "versionNonce": 1000108,
      "strokeColor": "#2f9e44",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [160, 0]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "decision",
        "focus": 0,
        "gap": 8
      },
      "endBinding": {
        "elementId": "end",
        "focus": 0,
        "gap": 8
      }
    },
    {
      "id": "arrow-decision-no",
      "type": "arrow",
      "x": 240,
      "y": 340,
      "width": -120,
      "height": 0,
      "seed": 1000109,
      "version": 1,
      "versionNonce": 1000109,
      "strokeColor": "#e03131",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [-120, 0]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "decision",
        "focus": 0,
        "gap": 8
      },
      "endBinding": null
    }
  ],
  "appState": {
    "gridSize": null,
    "viewBackgroundColor": "#ffffff"
  }
}
```

**Layout diagram of the flowchart:**

```
     ┌──────────┐
     │  Start   │ ◄── ellipse, x=220 y=20
     └────┬─────┘
          │  arrow-start-process (down 60px)
     ┌────┴─────┐
     │ Process  │ ◄── rectangle, x=220 y=140 (rounded)
     └────┬─────┘
          │  arrow-process-decision (down 60px)
     ┌────┴────┐
     │  OK?    │ ◄── diamond, x=240 y=280
     └────┬────┘
     yes  │    no
     ┌────┴────┐   ┌────
     │  End    │   │(no arrow end)
     └─────────┘
```

### Example 2: Architecture Diagram

A simple system architecture with labeled boxes and connecting arrows.

```json
{
  "type": "excalidraw",
  "version": 2,
  "source": "https://excalidraw.com",
  "elements": [
    {
      "id": "frontend",
      "type": "rectangle",
      "x": 60,
      "y": 100,
      "width": 180,
      "height": 100,
      "seed": 1000201,
      "version": 1,
      "versionNonce": 1000201,
      "strokeColor": "#1971c2",
      "backgroundColor": "#a5d8ff",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-fe-be", "type": "arrow" },
        { "id": "text-frontend", "type": "text" }
      ],
      "link": null,
      "locked": false,
      "roundness": { "type": 3 }
    },
    {
      "id": "text-frontend",
      "type": "text",
      "x": 90,
      "y": 130,
      "width": 120,
      "height": 40,
      "seed": 1000202,
      "version": 1,
      "versionNonce": 1000202,
      "strokeColor": "#1971c2",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "text": "Frontend",
      "originalText": "Frontend",
      "rawText": "Frontend",
      "fontSize": 22,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "containerId": "frontend",
      "autoResize": true,
      "lineHeight": 1.25
    },
    {
      "id": "backend",
      "type": "rectangle",
      "x": 380,
      "y": 100,
      "width": 180,
      "height": 100,
      "seed": 1000203,
      "version": 1,
      "versionNonce": 1000203,
      "strokeColor": "#2f9e44",
      "backgroundColor": "#b2f2bb",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-fe-be", "type": "arrow" },
        { "id": "arrow-be-db", "type": "arrow" },
        { "id": "text-backend", "type": "text" }
      ],
      "link": null,
      "locked": false,
      "roundness": { "type": 3 }
    },
    {
      "id": "text-backend",
      "type": "text",
      "x": 410,
      "y": 130,
      "width": 120,
      "height": 40,
      "seed": 1000204,
      "version": 1,
      "versionNonce": 1000204,
      "strokeColor": "#2f9e44",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "text": "Backend",
      "originalText": "Backend",
      "rawText": "Backend",
      "fontSize": 22,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "containerId": "backend",
      "autoResize": true,
      "lineHeight": 1.25
    },
    {
      "id": "database",
      "type": "ellipse",
      "x": 420,
      "y": 280,
      "width": 100,
      "height": 100,
      "seed": 1000205,
      "version": 1,
      "versionNonce": 1000205,
      "strokeColor": "#e67700",
      "backgroundColor": "#ffec99",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [
        { "id": "arrow-be-db", "type": "arrow" },
        { "id": "text-db", "type": "text" }
      ],
      "link": null,
      "locked": false
    },
    {
      "id": "text-db",
      "type": "text",
      "x": 445,
      "y": 315,
      "width": 50,
      "height": 30,
      "seed": 1000206,
      "version": 1,
      "versionNonce": 1000206,
      "strokeColor": "#e67700",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "text": "DB",
      "originalText": "DB",
      "rawText": "DB",
      "fontSize": 20,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "containerId": "database",
      "autoResize": true,
      "lineHeight": 1.25
    },
    {
      "id": "arrow-fe-be",
      "type": "arrow",
      "x": 240,
      "y": 150,
      "width": 140,
      "height": 0,
      "seed": 1000207,
      "version": 1,
      "versionNonce": 1000207,
      "strokeColor": "#1e1e1e",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [140, 0]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "frontend",
        "focus": 0,
        "gap": 8
      },
      "endBinding": {
        "elementId": "backend",
        "focus": 0,
        "gap": 8
      }
    },
    {
      "id": "arrow-be-db",
      "type": "arrow",
      "x": 470,
      "y": 200,
      "width": 0,
      "height": 80,
      "seed": 1000208,
      "version": 1,
      "versionNonce": 1000208,
      "strokeColor": "#1e1e1e",
      "strokeWidth": 2,
      "roughness": 1,
      "opacity": 100,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "link": null,
      "locked": false,
      "points": [[0, 0], [0, 80]],
      "startArrowhead": null,
      "endArrowhead": "arrow",
      "startBinding": {
        "elementId": "backend",
        "focus": 0,
        "gap": 8
      },
      "endBinding": {
        "elementId": "database",
        "focus": 0,
        "gap": 8
      }
    }
  ],
  "appState": {
    "gridSize": null,
    "viewBackgroundColor": "#ffffff"
  }
}
```

## Tips and Gotchas

### seed values
Each element must have a unique `seed` (integer). Use different number ranges
for different diagrams to avoid collisions. A simple approach: start at
`1000001` and increment.

### version and versionNonce
- `version`: Start at `1`. Increment by 1 whenever the element is modified.
- `versionNonce`: Any unique integer. Change it when the element changes.

### Arrow points
The first point in `points` is **always** `[0, 0]` — it's relative to the
element's `x`, `y` origin. All subsequent points are offsets from the first.

### Arrow width/height
For arrows, `width` and `height` should reflect the total span of the line.
For a pure horizontal arrow, `height` = 0. For a pure vertical, `width` = 0.

### Text sizing
With `autoResize: true`, Excalidraw recalculates text dimensions at load time.
Set approximate `width`/`height` — they'll be corrected.

### JSON validity
Excalidraw is strict about JSON. No trailing commas. All strings in double
quotes. Use a JSON validator if unsure.

### Canvas boundaries
Keep elements within reasonable coordinates (0–2000px range). Very large or
negative coordinates may cause viewport issues.

### Element overlap
Elements can overlap. Z-order is determined by array position — later elements
render on top. If something is hidden, check the order in the `elements` array.

### No trailing commas
Standard JSON rules apply. No trailing commas after the last element in an
array or the last property in an object.
