#!/usr/bin/env python3
"""Flowchart generator — produces .excalidraw files from a node/edge spec.

Usage:
  python3 flowchart.py input.yaml -o flowchart.excalidraw
  cat input.json | python3 flowchart.py - > flowchart.excalidraw
"""

import json
import sys
import os

# Add parent dir so that excalidraw_common is importable when running
# directly from the scripts/ directory or via ``python3 -m``.
_THIS_DIR = os.path.dirname(os.path.abspath(__file__))
if _THIS_DIR not in sys.path:
    sys.path.insert(0, _THIS_DIR)

# ---------------------------------------------------------------------------
# Local imports
# ---------------------------------------------------------------------------

from excalidraw_common import (                       # noqa: E402
    auto_seed,
    default_arrow,
    default_shape,
    detect_cycles,
    longest_path_layers,
    parse_output_flag,
    read_input,
    serialize,
    usage,
    wire_binding,
)

# ---------------------------------------------------------------------------
# Layout constants
# ---------------------------------------------------------------------------

DEFAULT_NODE_W = 160
DEFAULT_NODE_H = 60
DEFAULT_SPACING = 80
DECISION_W = 120
DECISION_H = 120

# ---------------------------------------------------------------------------
# Layout engine
# ---------------------------------------------------------------------------


def layout_flowchart(spec):
    """Compute positions for all nodes and arrow paths.

    Parameters
    ----------
    spec : dict
        Parsed input with keys ``nodes``, ``edges``, ``flow``, ``spacing``,
        ``node_width``, ``node_height``, ``theme``.

    Returns
    -------
    elements : list[dict]
        Excalidraw element dicts ready for serialisation.
    """
    nodes = spec.get("nodes", [])
    edges = spec.get("edges", [])
    flow = spec.get("flow", "TB").upper()  # TB or LR
    spacing = spec.get("spacing", DEFAULT_SPACING)
    nw = spec.get("node_width", DEFAULT_NODE_W)
    nh = spec.get("node_height", DEFAULT_NODE_H)
    theme = spec.get("theme", "default")

    # Index nodes by id
    node_map = {n["id"]: n for n in nodes}
    node_ids = list(node_map.keys())
    edge_tuples = [(e["from"], e["to"]) for e in edges]

    # Detect cycles and warn
    if detect_cycles(node_ids, edge_tuples):
        print(
            "Warning: Graph contains cycles. Layout may look odd.",
            file=sys.stderr,
        )

    # Assign layers
    layers = longest_path_layers(node_ids, edge_tuples)
    max_layer = max(layers.values()) if layers else 0

    # Group nodes by layer
    layer_nodes = {}
    for nid, layer in layers.items():
        layer_nodes.setdefault(layer, []).append(nid)

    elements = []

    # Compute positions
    positions = {}  # node_id -> (x, y, w, h)

    for layer_idx in range(max_layer + 1):
        nids = layer_nodes.get(layer_idx, [])
        count = len(nids)
        total_w = count * nw + (count - 1) * spacing
        start_x = 50  # left margin

        if flow == "LR":
            # Layers as columns (left-to-right)
            for i, nid in enumerate(nids):
                x = 50 + layer_idx * (nw + spacing)
                y = 50 + i * (nh + spacing)
                positions[nid] = (x, y, nw, nh)
        else:
            # TB — layers as rows (top-to-bottom)
            for i, nid in enumerate(nids):
                x = start_x + i * (nw + spacing)
                y = 50 + layer_idx * (nh + spacing)
                positions[nid] = (x, y, nw, nh)

    # Determine node shape dimensions based on type
    def shape_for(node):
        kind = node.get("type", "process")
        shape_type = _node_shape(kind)
        if shape_type == "diamond":
            return shape_type, DECISION_W, DECISION_H
        return shape_type, nw, nh

    # Create shape elements
    for nid, node in node_map.items():
        x, y, w, h = positions[nid]
        kind = node.get("type", "process")
        shape_type, sw, sh = shape_for(node)

        # Centre the diamond in its cell (different dimensions)
        if shape_type == "diamond":
            dx = x + (w - sw) // 2
            dy = y + (h - sh) // 2
        else:
            dx, dy = x, y

        label = node.get("label", "")
        elements.extend(
            default_shape(
                shape_type,
                nid,
                dx,
                dy,
                sw,
                sh,
                label=label,
                theme_name=theme,
                node_kind=kind,
            )
        )

    # Create arrow elements
    for edge in edges:
        fid = edge["from"]
        tid = edge["to"]
        if fid not in positions or tid not in positions:
            continue

        fx, fy, fw, fh = positions[fid]
        tx, ty, tw, th = positions[tid]

        if flow == "LR":
            # Left-to-right: arrow exits right edge of source, enters left edge of target
            ax = fx + fw
            ay = fy + fh // 2
            bx = tx
            by = ty + th // 2
            points = [[0, 0], [bx - ax, 0]]
        else:
            # Top-to-bottom: arrow exits bottom edge of source, enters top edge of target
            ax = fx + fw // 2
            ay = fy + fh
            bx = tx + tw // 2
            by = ty
            points = [[0, 0], [0, by - ay]]

        arrow_id = f"arrow-{fid}-{tid}"
        label = edge.get("label")
        edge_style = edge.get("style", {})

        arrow = default_arrow(
            arrow_id,
            ax,
            ay,
            points,
            start_id=fid,
            end_id=tid,
            **edge_style,
        )

        # Wire boundElements on the two shapes
        for shape in elements:
            if shape["type"] != "text" and shape["id"] in (fid, tid):
                shape.setdefault("boundElements", None)
                if shape["boundElements"] is None:
                    shape["boundElements"] = []
                shape["boundElements"].append(
                    wire_binding(shape["id"], arrow_id, "arrow")
                )

        elements.append(arrow)

        if label:
            # Place label at midpoint of the arrow
            mx = ax + points[-1][0] // 2 - 30
            my = ay + points[-1][1] // 2 - 20
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
                "strokeColor": edge_style.get("strokeColor", "#1e1e1e"),
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
                "text": label,
                "originalText": label,
                "rawText": label,
                "fontSize": 14,
                "fontFamily": 1,
                "textAlign": "center",
                "verticalAlign": "middle",
                "containerId": None,
                "autoResize": True,
                "lineHeight": 1.25,
            }
            elements.append(label_el)

    # Re-order for outline rendering (alternative ideas later)
    return elements


def _node_shape(kind):
    """Map semantic node kind to Excalidraw element type."""
    return {
        "start": "ellipse",
        "end": "ellipse",
        "process": "rectangle",
        "decision": "diamond",
        "io": "diamond",
    }.get(kind, "rectangle")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def main(argv=None):
    if argv is None:
        argv = sys.argv[1:]

    if "-h" in argv or "--help" in argv:
        print(
            usage(
                "flowchart.py",
                "Generate a flowchart .excalidraw file from a node/edge spec.",
                extra_flags="",
            )
        )
        return

    rest, out_path = parse_output_flag(argv)

    if not rest:
        print("Error: missing input file (use '-' for stdin).", file=sys.stderr)
        sys.exit(1)

    spec = read_input(rest[0])

    elements = layout_flowchart(spec)
    json_str = serialize(elements)
    if out_path and out_path != "-":
        with open(out_path, "w") as f:
            f.write(json_str)
        print(f"Wrote {len(elements)} elements to {out_path}", file=sys.stderr)
    else:
        sys.stdout.write(json_str)


if __name__ == "__main__":
    main()
