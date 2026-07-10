#!/bin/bash
# Symphony setup — build Maestro and configure PATH in shell rc file
#
# Usage: setup maestro [--dry-run] [--maestro-dir <path>] [--config-file-path <path>] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup maestro [--dry-run] [--maestro-dir <path>] [--config-file-path <path>] [-h/--help]

  Build the Maestro binary and add it to PATH in your shell rc file.

  --dry-run                Preview without making changes
  --maestro-dir <path>     Path to the maestro source directory
  --config-file-path <path>  Shell rc file to update (default: ~/.zshrc)
  -h, --help               Show this help message"

DRY_RUN=false
MAESTRO_DIR_ARG=""
CONFIG_FILE_PATH=""

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --dry-run) DRY_RUN=true ;;
    --maestro-dir)
      if [[ -z "${2:-}" ]]; then echo "Error: --maestro-dir requires a path"; exit 1; fi
      MAESTRO_DIR_ARG="$2"
      shift
      ;;
    --config-file-path)
      if [[ -z "${2:-}" ]]; then echo "Error: --config-file-path requires a path"; exit 1; fi
      CONFIG_FILE_PATH="$2"
      shift
      ;;
    -h|--help) echo "$HELP_MESSAGE"; exit 0 ;;
    *) echo "Unknown parameter passed: $1"; echo "$HELP_MESSAGE"; exit 1 ;;
  esac
  shift
done

SYMPHONY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MAESTRO_DIR="${MAESTRO_DIR_ARG:-$SYMPHONY_DIR/maestro}"
RC_FILE="${CONFIG_FILE_PATH:-$HOME/.zshrc}"

if $DRY_RUN; then
  echo "[DRY RUN] cd $MAESTRO_DIR && go build -o maestro ."
  echo "[DRY RUN] Would add to $RC_FILE:"
  echo "[DRY RUN]   # Maestro Planning Server"
  echo "[DRY RUN]   export PATH=\"\$PATH:$MAESTRO_DIR\""
  echo "[DRY RUN]   export MAESTRO_DIR=\"$MAESTRO_DIR\""
  echo "[DRY RUN]   export MAESTRO_PLANS_DIR=\"$MAESTRO_DIR/plans\""
  MAESTRO_INJECT_SRC="$SYMPHONY_DIR/maestro/AGENTS_INJECT.md"
  CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"
  AGENTS_DST="$CONFIG_DIR/AGENTS.md"
  if [ -f "$MAESTRO_INJECT_SRC" ]; then
    echo "[DRY RUN] Would append $MAESTRO_INJECT_SRC -> $AGENTS_DST"
  fi
  exit 0
fi

# Resolve to absolute path
MAESTRO_DIR="$(cd "$MAESTRO_DIR" 2>/dev/null && pwd || { echo "Error: directory not found: $MAESTRO_DIR" >&2; exit 1; })"

echo "  Building maestro..."
(cd "$MAESTRO_DIR" && go build -o maestro .)
echo "  built $MAESTRO_DIR/maestro"

# Add to shell rc file (idempotent: checks for a marker comment)
BLOCK_MARKER="# Maestro Planning Server"
if grep -qF "$BLOCK_MARKER" "$RC_FILE" 2>/dev/null; then
  echo "  already in $RC_FILE: $BLOCK_MARKER (and exports below)"
else
  {
    echo ""
    echo "$BLOCK_MARKER"
    echo "export PATH=\"\$PATH:$MAESTRO_DIR\""
    echo "export MAESTRO_DIR=\"$MAESTRO_DIR\""
    echo "export MAESTRO_PLANS_DIR=\"\$MAESTRO_DIR/plans\""
  } >> "$RC_FILE"
  echo "  added Maestro config to $RC_FILE"
fi

echo "  Run 'source $RC_FILE' or open a new terminal to use 'maestro'."

# ── Inject Maestro agent instructions into AGENTS.md ────────────────────
MAESTRO_INJECT_SRC="$SYMPHONY_DIR/maestro/AGENTS_INJECT.md"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"
AGENTS_DST="$CONFIG_DIR/AGENTS.md"
INJECT_MARKER="<!-- maestro-agent-instructions -->"

if [ -f "$MAESTRO_INJECT_SRC" ]; then
  if grep -qF "$INJECT_MARKER" "$AGENTS_DST" 2>/dev/null; then
    echo "  maestro agent instructions already in $AGENTS_DST"
  else
    mkdir -p "$CONFIG_DIR"
    # Ensure the dest file exists before appending
    [ -f "$AGENTS_DST" ] || touch "$AGENTS_DST"
    {
      echo ""
      echo "$INJECT_MARKER"
      cat "$MAESTRO_INJECT_SRC"
    } >> "$AGENTS_DST"
    echo "  added maestro agent instructions to $AGENTS_DST"
  fi
fi
