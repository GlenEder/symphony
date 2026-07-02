#!/usr/bin/env python3
"""Architecture diagram generator — produces .excalidraw files from a tiered spec.

Usage:
  python3 architecture.py input.yaml -o arch.excalidraw
  cat input.json | python3 architecture.py - > arch.excalidraw
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
    default_element,
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

DEFAULT_COMPONENT_H = 60
DEFAULT_SPACING = 60
TOP_MARGIN = 50
LEFT_MARGIN = 50
TIER_LABEL_H = 30
TIER_PADDING = 40
TIER_GAP = 80  # vertical gap between tier bands
MIN_TIER_W = 200


def layout_tiers(spec):
    """Compute positions for all tier components and arrows.

    Parameters
    ----------
    spec : dict
        Parsed input with keys ``tiers``, ``connections``, ``spacing``.

    Returns
    -------
    elements : list[dict]
        Excalidraw element dicts ready for serialisation.
    """
    tiers = spec.get("tiers", [])
    connections = spec.get("connections", [])
    spacing = spec.get("spacing", DEFAULT_SPACING)

    if not tiers:
        print("Error: at least one tier is required.", file=sys.stderr)
        sys.exit(1)

    elements = []
    tier_top = TOP_MARGIN

    # Map component_id -> (x, y, w, h, tier_index)
    comp_positions = {}

    # Build each tier band
    for ti, tier in enumerate(tiers):
        tier_name = tier.get("name", f"Tier {ti + 1}")
        tier_color = tier.get("color", "#e9ecef")
        components = tier.get("components", [])

        # Compute component dimensions within this tier
        n = len(components)
        comp_heights = []
        comp_widths = []
        for comp in components:
            cw = comp.get("width", 140)
            ch = comp.get("height", DEFAULT_COMPONENT_H)
            comp_widths.append(cw)
            comp_heights.append(ch)

        max_ch = max(comp_heights) if comp_heights else DEFAULT_COMPONENT_H
        total_w = sum(comp_widths) + (n - 1) * spacing
        band_left = LEFT_MARGIN
        band_w = max(total_w, MIN_TIER_W)

        # Create tier background rectangle (subtle)
        bg_id = f"bg-tier-{ti}"
        bg = default_element(
            "rectangle",
            bg_id,
            band_left - 10,
            tier_top,
            band_w + 20,
            max_ch + TIER_LABEL_H + TIER_PADDING,
            fillStyle="solid",
            strokeWidth=0,
            backgroundColor=tier_color,
            opacity=20,
            roughness=0,
        )
        elements.append(bg)

        # Create tier label
        label_id = f"label-tier-{ti}"
        label_x = band_left
        label_y = tier_top + 10
        label_el = default_element(
            "text",
            label_id,
            label_x,
            label_y,
            200,
            TIER_LABEL_H,
            text=tier_name,
            originalText=tier_name,
            rawText=tier_name,
            fontSize=20,
            fontFamily=1,
            textAlign="left",
            verticalAlign="middle",
            containerId=None,
            autoResize=True,
            lineHeight=1.25,
            strokeColor="#495057",
            backgroundColor="transparent",
            fillStyle="solid",
        )
        elements.append(label_el)

        # Layout components horizontally within tier
        comp_y = tier_top + TIER_LABEL_H + TIER_PADDING // 2
        cx = band_left
        for ci, comp in enumerate(components):
            cid = comp["id"]
            cw = comp_widths[ci]
            ch = comp_heights[ci]
            label = comp.get("label", cid)
            shape_type = comp.get("shape", "rectangle")

            # Center component vertically within its cell
            cy = comp_y + (max_ch - ch) // 2

            # Use theme-like color from tier color (or inherit)
            style = {}
            if tier_color:
                style["backgroundColor"] = tier_color

            comp_els = default_shape(
                shape_type,
                cid,
                cx,
                cy,
                cw,
                ch,
                label=label,
                **style,
            )
            elements.extend(comp_els)

            comp_positions[cid] = (cx, cy, cw, ch, ti)
            cx += cw + spacing

        # Update tier_top for next tier
        tier_top += max_ch + TIER_LABEL_H + TIER_PADDING + TIER_GAP

    # Create connection arrows
    for conn in connections:
        fid = conn.get("from")
        tid = conn.get("to")
        if fid not in comp_positions or tid not in comp_positions:
            print(
                f"Warning: connection references unknown component '{fid}' or '{tid}'.",
                file=sys.stderr,
            )
            continue

        fx, fy, fw, fh, fti = comp_positions[fid]
        tx, ty, tw, th, tti = comp_positions[tid]

        # Arrow exits bottom-center of source, enters top-center of target
        ax = fx + fw // 2
        ay = fy + fh
        bx = tx + tw // 2
        by = ty
        points = [[0, 0], [bx - ax, by - ay]]

        arrow_id = f"arrow-{fid}-{tid}"
        conn_label = conn.get("label")
        conn_style = conn.get("style", {})

        arrow = default_arrow(
            arrow_id,
            ax,
            ay,
            points,
            start_id=fid,
            end_id=tid,
            **conn_style,
        )
        elements.append(arrow)

        # Wire boundElements on both shapes
        for el in elements:
            if el["type"] != "text" and el["id"] in (fid, tid):
                el.setdefault("boundElements", None)
                if el["boundElements"] is None:
                    el["boundElements"] = []
                el["boundElements"].append(
                    wire_binding(el["id"], arrow_id, "arrow")
                )

        # Connection label at midpoint
        if conn_label:
            mx = ax + (bx - ax) // 2 - 30
            my = ay + (by - ay) // 2 - 20
            label_el = {
                "id": f"label-{fid}-{tid}",
                "type": "text",
                "x": mx,
                "y": my,
                "width": 60,
                "height": 20,
                "seed": auto_seed(),
                "version": 1,
                "versionNonce": auto_seed(),
                "strokeColor": conn_style.get("strokeColor", "#1e1e1e"),
                "backgroundColor": "transparent",
                "fillStyle": "solid",
                "strokeWidth": 1,
                "roughness": 1,
                "opacity": 100,
                "isDeleted": False,
                "groupIds": [],
                "boundElements": None,
                "link": None,
                "locked": False,
                "text": conn_label,
                "originalText": conn_label,
                "rawText": conn_label,
                "fontSize": 14,
                "fontFamily": 1,
                "textAlign": "center",
                "verticalAlign": "middle",
                "containerId": None,
                "autoResize": True,
                "lineHeight": 1.25,
            }
            elements.append(label_el)

    return elements


def main(argv=None):
    if argv is None:
        argv = sys.argv[1:]

    if "-h" in argv or "--help" in argv:
        print(
            usage(
                "architecture.py",
                "Generate an architecture diagram .excalidraw file from a tiered spec.",
            )
        )
        return

    rest, out_path = parse_output_flag(argv)

    if not rest:
        print("Error: missing input file (use '-' for stdin).", file=sys.stderr)
        sys.exit(1)

    spec = read_input(rest[0])
    elements = layout_tiers(spec)
    json_str = serialize(elements)

    if out_path and out_path != "-":
        with open(out_path, "w") as f:
            f.write(json_str)
        print(f"Wrote {len(elements)} elements to {out_path}", file=sys.stderr)
    else:
        sys.stdout.write(json_str)


if __name__ == "__main__":
    main()
