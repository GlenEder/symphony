#!/bin/bash
# Symphony setup — install AGENTS.md into config dir
#
# Usage: setup agents [--dry-run] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup agents [--dry-run] [-h/--help]

  Copy AGENTS.md into the Maki config directory.

  --dry-run    Preview without making changes
  -h, --help   Show this help message"

DRY_RUN=false

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --dry-run) DRY_RUN=true ;;
    -h|--help) echo "$HELP_MESSAGE"; exit 0 ;;
    *) echo "Unknown parameter passed: $1"; echo "$HELP_MESSAGE"; exit 1 ;;
  esac
  shift
done

SYMPHONY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"

src="$SYMPHONY_DIR/AGENTS.md"
dst="$CONFIG_DIR/AGENTS.md"

if [ ! -f "$src" ]; then
  exit 0
fi

if $DRY_RUN; then
  echo "[DRY RUN] cp $src -> $dst"
  exit 0
fi

mkdir -p "$(dirname "$dst")"

# Break existing symlink if it points to our source (cp refuses identical files)
if [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ]; then
  rm "$dst"
fi

cp "$src" "$dst"
echo "  installed AGENTS.md"
