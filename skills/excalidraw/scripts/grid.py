#!/usr/bin/env python3
"""Grid layout generator — produces .excalidraw files from a row/col item spec.

Usage:
  python3 grid.py input.yaml -o grid.excalidraw
  cat input.json | python3 grid.py - > grid.excalidraw
"""

import json
import os
import sys

_THIS_DIR = os.path.dirname(os.path.abspath(__file__))
if _THIS_DIR not in sys.path:
    sys.path.insert(0, _THIS_DIR)

from excalidraw_common import (
    auto_seed,
    default_arrow,
    default_shape,
    parse_output_flag,
    read_input,
    serialize,
    usage,
    wire_binding,
)

# ---------------------------------------------------------------------------
# Layout constants
# ---------------------------------------------------------------------------

DEFAULT_COLS = 3
DEFAULT_CELL_W = 160
DEFAULT_CELL_H = 80
DEFAULT_SPACING = 30
LEFT_MARGIN = 50
TOP_MARGIN = 50


def layout_grid(spec):
    """Compute positions for grid items.

    Parameters
    ----------
    spec : dict
        Parsed input with keys ``items``, ``cols``, ``cell_width``,
        ``cell_height``, ``spacing``, ``style``, ``connect``.

    Returns
    -------
    elements : list[dict]
        Excalidraw element dicts ready for serialisation.
    """
    items = spec.get("items", [])
    if not items:
        print("Error: at least one item is required.", file=sys.stderr)
        sys.exit(1)

    cols = spec.get("cols", DEFAULT_COLS)
    cw = spec.get("cell_width", DEFAULT_CELL_W)
    ch = spec.get("cell_height", DEFAULT_CELL_H)
    spacing = spec.get("spacing", DEFAULT_SPACING)
    global_style = spec.get("style", {})
    connect = spec.get("connect", False)

    elements = []

    # Place items in row-major order
    for idx, item in enumerate(items):
        item_id = item.get("id", f"item-{idx}")
        label = item.get("label", item_id)
        row = idx // cols
        col = idx % cols

        x = LEFT_MARGIN + col * (cw + spacing)
        y = TOP_MARGIN + row * (ch + spacing)

        shape_type = item.get("shape", "rectangle")
        item_style = item.get("style", {})

        # Merge global style with per-item overrides
        merged_style = {**global_style, **item_style}

        comp_els = default_shape(
            shape_type,
            item_id,
            x,
            y,
            cw,
            ch,
            label=label,
            **merged_style,
        )
        elements.extend(comp_els)

    # Optionally connect adjacent cells horizontally and vertically
    if connect:
        n_rows = (len(items) + cols - 1) // cols

        for r in range(n_rows):
            for c in range(cols):
                idx = r * cols + c
                if idx >= len(items):
                    break
                item_id = items[idx].get("id", f"item-{idx}")
                x = LEFT_MARGIN + c * (cw + spacing)
                y = TOP_MARGIN + r * (ch + spacing)

                # Horizontal connection to right neighbor
                if c + 1 < cols:
                    right_idx = r * cols + c + 1
                    if right_idx < len(items):
                        right_id = items[right_idx].get("id", f"item-{right_idx}")
                        ax = x + cw
                        ay = y + ch // 2
                        bx = x + cw + spacing
                        by = y + ch // 2
                        points = [[0, 0], [spacing, 0]]
                        arrow_id = f"arrow-{item_id}-{right_id}"
                        arrow = default_arrow(
                            arrow_id,
                            ax,
                            ay,
                            points,
                            start_id=item_id,
                            end_id=right_id,
                        )
                        elements.append(arrow)
                        for el in elements:
                            if el["type"] != "text" and el["id"] in (item_id, right_id):
                                el.setdefault("boundElements", None)
                                if el["boundElements"] is None:
                                    el["boundElements"] = []
                                el["boundElements"].append(
                                    wire_binding(el["id"], arrow_id, "arrow")
                                )

                # Vertical connection to bottom neighbor
                if r + 1 < n_rows:
                    down_idx = (r + 1) * cols + c
                    if down_idx < len(items):
                        down_id = items[down_idx].get("id", f"item-{down_idx}")
                        ax = x + cw // 2
                        ay = y + ch
                        bx = x + cw // 2
                        by = y + ch + spacing
                        points = [[0, 0], [0, spacing]]
                        arrow_id = f"arrow-{item_id}-{down_id}"
                        arrow = default_arrow(
                            arrow_id,
                            ax,
                            ay,
                            points,
                            start_id=item_id,
                            end_id=down_id,
                        )
                        elements.append(arrow)
                        for el in elements:
                            if el["type"] != "text" and el["id"] in (item_id, down_id):
                                el.setdefault("boundElements", None)
                                if el["boundElements"] is None:
                                    el["boundElements"] = []
                                el["boundElements"].append(
                                    wire_binding(el["id"], arrow_id, "arrow")
                                )

    return elements


def main(argv=None):
    if argv is None:
        argv = sys.argv[1:]

    if "-h" in argv or "--help" in argv:
        print(
            usage(
                "grid.py",
                "Generate a grid layout .excalidraw file from an item spec.",
                extra_flags="",
            )
        )
        return

    rest, out_path = parse_output_flag(argv)

    if not rest:
        print("Error: missing input file (use '-' for stdin).", file=sys.stderr)
        sys.exit(1)

    spec = read_input(rest[0])
    elements = layout_grid(spec)
    json_str = serialize(elements)

    if out_path and out_path != "-":
        with open(out_path, "w") as f:
            f.write(json_str)
        print(f"Wrote {len(elements)} elements to {out_path}", file=sys.stderr)
    else:
        sys.stdout.write(json_str)


if __name__ == "__main__":
    main()
