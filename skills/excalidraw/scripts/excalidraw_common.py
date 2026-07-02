"""Shared utilities for Excalidraw diagram generation.

Provides element builders, defaults, I/O helpers, and color themes.
All scripts in this directory use this module.
"""

import json
import os
import re
import sys

# ---------------------------------------------------------------------------
# Default element properties
# ---------------------------------------------------------------------------

COMMON_DEFAULTS = {
    "angle": 0,
    "strokeColor": "#1e1e1e",
    "backgroundColor": "transparent",
    "fillStyle": "solid",
    "strokeWidth": 2,
    "strokeStyle": "solid",
    "roughness": 1,
    "opacity": 100,
    "isDeleted": False,
    "groupIds": [],
    "boundElements": None,
    "link": None,
    "locked": False,
}

# ---------------------------------------------------------------------------
# Color themes
# ---------------------------------------------------------------------------

THEMES = {
    "default": {
        "start": {"backgroundColor": "#b2f2bb", "strokeColor": "#2f9e44"},
        "end": {"backgroundColor": "#ffc9c9", "strokeColor": "#e03131"},
        "process": {"backgroundColor": "#a5d8ff", "strokeColor": "#1971c2"},
        "decision": {"backgroundColor": "#fff3bf", "strokeColor": "#e67700"},
        "io": {"backgroundColor": "#d5f4e6", "strokeColor": "#087f5b"},
        "text": {"strokeColor": "#1e1e1e"},
        "arrow": {"strokeColor": "#1e1e1e"},
        "line": {"strokeColor": "#1e1e1e"},
    },
    "pastel": {
        "start": {"backgroundColor": "#d3f9d8", "strokeColor": "#51cf66"},
        "end": {"backgroundColor": "#ffd8d8", "strokeColor": "#ff6b6b"},
        "process": {"backgroundColor": "#d0ebff", "strokeColor": "#339af0"},
        "decision": {"backgroundColor": "#fff9db", "strokeColor": "#fcc419"},
        "io": {"backgroundColor": "#e6fcf5", "strokeColor": "#20c997"},
        "text": {"strokeColor": "#495057"},
        "arrow": {"strokeColor": "#495057"},
        "line": {"strokeColor": "#495057"},
    },
    "dark": {
        "start": {"backgroundColor": "#2b8a3e", "strokeColor": "#69db7c"},
        "end": {"backgroundColor": "#c92a2a", "strokeColor": "#ff8787"},
        "process": {"backgroundColor": "#1864ab", "strokeColor": "#74c0fc"},
        "decision": {"backgroundColor": "#e8590c", "strokeColor": "#ffd43b"},
        "io": {"backgroundColor": "#087f5b", "strokeColor": "#63e6be"},
        "text": {"strokeColor": "#f8f9fa"},
        "arrow": {"strokeColor": "#f8f9fa"},
        "line": {"strokeColor": "#f8f9fa"},
    },
}

NODE_TYPE_MAP = {
    "start": "ellipse",
    "end": "ellipse",
    "process": "rectangle",
    "decision": "diamond",
    "io": "diamond",
}

# ---------------------------------------------------------------------------
# Seed / version helpers
# ---------------------------------------------------------------------------

_seed_counter = [0]


def auto_seed(base=None):
    """Return a unique, deterministic seed integer.

    If *base* is given, combine it with an incrementing counter so that
    multiple calls with the same base still yield unique values.
    """
    _seed_counter[0] += 1
    if base is not None:
        h = hash(str(base)) & 0x7FFFFFFF
        return h + _seed_counter[0]
    return 1000000 + _seed_counter[0]


def reset_seed_counter():
    """Reset the internal seed counter (useful for tests)."""
    _seed_counter[0] = 0


def next_version():
    """Return a version/nonce pair starting at 1."""
    _seed_counter[0] += 1
    return _seed_counter[0], _seed_counter[0] + 10000

# ---------------------------------------------------------------------------
# Element builders
# ---------------------------------------------------------------------------


def default_element(type_, id_, x, y, w, h, **overrides):
    """Create a complete Excalidraw element dict.

    All required fields are populated with sensible defaults.  Pass any
    additional keyword arguments to override individual fields.

    Returns a dict suitable for inclusion in the ``elements`` array of a
    ``.excalidraw`` file.
    """
    seed, nonce = next_version()
    el = {
        "id": id_,
        "type": type_,
        "x": x,
        "y": y,
        "width": w,
        "height": h,
        "seed": seed,
        "version": 1,
        "versionNonce": nonce,
        **COMMON_DEFAULTS,
    }
    el.update(overrides)
    return el


def default_shape(
    type_,
    id_,
    x,
    y,
    w,
    h,
    label=None,
    theme_name="default",
    node_kind=None,
    **style_overrides,
):
    """Create a shape element, optionally with a bound text label.

    Parameters
    ----------
    type_ : str
        Excalidraw element type (``rectangle``, ``ellipse``, ``diamond``, …).
    id_ : str
        Unique element id.
    x, y, w, h : float
        Position and size.
    label : str or None
        If provided, a bound text element is created inside the shape.
    theme_name : str
        One of ``"default"``, ``"pastel"``, ``"dark"``.
    node_kind : str or None
        Semantic node kind (``"start"``, ``"end"``, ``"process"``, …) for
        automatic colour mapping.  Falls back to *type_* if not given.
    **style_overrides
        Additional element properties (e.g. ``strokeWidth=4``).

    Returns
    -------
    list[dict]
        A list with the shape element (and the text element if *label* was
        given).  Use ``extend()`` to add the result to your elements array.
    """
    if node_kind and node_kind in THEMES.get(theme_name, {}):
        theme = THEMES[theme_name][node_kind]
    else:
        theme = {}

    shape = default_element(
        type_,
        id_,
        x,
        y,
        w,
        h,
        **theme,
        **style_overrides,
    )

    elements = [shape]

    if label:
        text_id = f"text-{id_}"
        # Position text roughly in the centre of the shape
        tx = x + 10
        ty = y + 10
        tw = w - 20
        th = h - 20

        text_theme = THEMES.get(theme_name, {}).get("text", {})
        text_el = default_element(
            "text",
            text_id,
            tx,
            ty,
            tw,
            th,
            text=label,
            originalText=label,
            rawText=label,
            fontSize=18,
            fontFamily=1,
            textAlign="center",
            verticalAlign="middle",
            containerId=id_,
            autoResize=True,
            lineHeight=1.25,
            **text_theme,
        )
        # Wire up the container binding
        shape["boundElements"] = shape.get("boundElements") or []
        if not isinstance(shape["boundElements"], list):
            shape["boundElements"] = []
        shape["boundElements"].append({"id": text_id, "type": "text"})

        elements.append(text_el)

    return elements


def default_arrow(id_, x, y, points, start_id=None, end_id=None, **style_overrides):
    """Create an arrow element, optionally bound to two shapes.

    Parameters
    ----------
    id_ : str
        Unique element id.
    x, y : float
        Arrow origin (matches the first point).
    points : list[list[float]]
        Path points relative to (x, y), e.g. ``[[0, 0], [200, 0]]``.
    start_id : str or None
        Id of the element this arrow starts from.
    end_id : str or None
        Id of the element this arrow ends at.
    **style_overrides
        Additional element properties.

    Returns
    -------
    dict
        The arrow element, with bindings wired if *start_id* / *end_id* were
        given.  The caller is responsible for adding the matching
        ``boundElements`` entries on the target shapes.
    """
    last = points[-1] if points else [0, 0]
    w = abs(last[0])
    h = abs(last[1])

    arrow = default_element(
        "arrow",
        id_,
        x,
        y,
        w,
        h,
        points=points,
        startArrowhead=None,
        endArrowhead="arrow",
        startBinding=None,
        endBinding=None,
        **style_overrides,
    )

    if start_id:
        arrow["startBinding"] = {
            "elementId": start_id,
            "focus": 0,
            "gap": 8,
        }
    if end_id:
        arrow["endBinding"] = {
            "elementId": end_id,
            "focus": 0,
            "gap": 8,
        }

    return arrow


def wire_binding(target_id, source_id, binding_type="arrow"):
    """Return a ``boundElements`` entry dict."""
    return {"id": source_id, "type": binding_type}

# ---------------------------------------------------------------------------
# Serialization
# ---------------------------------------------------------------------------

DEFAULT_APPSTATE = {
    "gridSize": None,
    "viewBackgroundColor": "#ffffff",
}


def serialize(elements, app_state=None):
    """Wrap *elements* in the ``.excalidraw`` envelope.

    Returns a JSON string.
    """
    if app_state is None:
        app_state = DEFAULT_APPSTATE
    data = {
        "type": "excalidraw",
        "version": 2,
        "source": "https://excalidraw.com",
        "elements": elements,
        "appState": app_state,
    }
    return json.dumps(data, indent=2, ensure_ascii=False) + "\n"

# ---------------------------------------------------------------------------
# I/O
# ---------------------------------------------------------------------------


def read_input(path):
    """Read and parse a JSON or YAML file.

    If *path* is ``"-"`` read from stdin.  Returns a Python dict.
    """
    if path == "-":
        text = sys.stdin.read()
    else:
        with open(path) as f:
            text = f.read()

    # Try JSON first
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    # Fall back to YAML (if PyYAML is available)
    try:
        import yaml
    except ImportError:
        raise SystemExit(
            "Input is not valid JSON, and PyYAML is not installed.\n"
            "  pip install pyyaml"
        )
    try:
        return yaml.safe_load(text)
    except yaml.YAMLError as exc:
        raise SystemExit(f"Invalid YAML: {exc}")


def write_output(data, path, pretty=True):
    """Write *data* (a dict) to *path* as JSON.

    If *path* is ``"-"`` (or ``None``) write to stdout.
    """
    text = serialize(data) if pretty else json.dumps(data, ensure_ascii=False)
    if path and path != "-":
        with open(path, "w") as f:
            f.write(text)
    else:
        sys.stdout.write(text)


def pretty_path(path):
    """Return a human-friendly path for messages."""
    if not path or path == "-":
        return "<stdout>"
    return os.path.abspath(path)

# ---------------------------------------------------------------------------
# CLI argument helpers
# ---------------------------------------------------------------------------


def parse_output_flag(args):
    """Extract ``-o`` / ``--output`` value from *args* and return (cleaned_args, out_path)."""
    out = None
    rest = []
    i = 0
    while i < len(args):
        if args[i] in ("-o", "--output") and i + 1 < len(args):
            out = args[i + 1]
            i += 2
        elif args[i].startswith("--output="):
            out = args[i].split("=", 1)[1]
            i += 1
        else:
            rest.append(args[i])
            i += 1
    return rest, out


def usage(script_name, description, extra_flags=""):
    """Return a formatted usage string."""
    return f"""\
Usage: python3 {script_name} <input> [options]

{description}

Arguments:
  input                  Input file (.json / .yaml) or "-" for stdin.

Options{extra_flags}
  -o, --output <path>    Output .excalidraw file (default: stdout).
  -h, --help             Show this help message.
"""


# ---------------------------------------------------------------------------
# Simple dag layout utilities
# ---------------------------------------------------------------------------

def longest_path_layers(nodes, edges):
    """Assign each node to a layer using longest-path / topological order.

    *nodes* is a list of node-id strings.
    *edges* is a list of ``(from_id, to_id)`` tuples.

    Returns a dict ``{node_id: layer_index}`` where layer 0 is the top-most /
    left-most layer.
    """
    graph = {n: [] for n in nodes}
    in_degree = {n: 0 for n in nodes}
    for f, t in edges:
        if f in graph and t in graph:
            graph[f].append(t)
            in_degree[t] = in_degree.get(t, 0) + 1

    # Kahn's algorithm to get topological order
    queue = [n for n in nodes if in_degree.get(n, 0) == 0]
    topo = []
    while queue:
        node = queue.pop(0)
        topo.append(node)
        for neighbor in graph.get(node, []):
            in_degree[neighbor] -= 1
            if in_degree[neighbor] == 0:
                queue.append(neighbor)

    # Assign layers using longest-path
    layers = {n: 0 for n in nodes}
    for node in topo:
        for neighbor in graph.get(node, []):
            layers[neighbor] = max(layers[neighbor], layers[node] + 1)

    return layers


def detect_cycles(nodes, edges):
    """Return True if the graph contains a cycle (conservative)."""
    graph = {n: [] for n in nodes}
    for f, t in edges:
        if f in graph and t in graph:
            graph[f].append(t)

    WHITE, GRAY, BLACK = 0, 1, 2
    color = {n: WHITE for n in nodes}

    def dfs(n):
        color[n] = GRAY
        for nb in graph.get(n, []):
            if color[nb] == GRAY:
                return True
            if color[nb] == WHITE and dfs(nb):
                return True
        color[n] = BLACK
        return False

    for n in nodes:
        if color[n] == WHITE:
            if dfs(n):
                return True
    return False
