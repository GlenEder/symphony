#!/bin/bash
# Symphony setup — install AGENTS.md into config dir
#
# Usage: setup agents [--dry-run] [--claude] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup agents [--dry-run] [--claude] [-h/--help]

  Copy AGENTS.md into the Maki or Claude config directory.

  --claude     Target Claude config (~/.claude/CLAUDE.md) instead of Maki
  --dry-run    Preview without making changes
  -h, --help   Show this help message

  Installing replaces any existing destination file."

DRY_RUN=false
CLAUDE=false

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --dry-run) DRY_RUN=true ;;
    --claude) CLAUDE=true ;;
    -h|--help) echo "$HELP_MESSAGE"; exit 0 ;;
    *) echo "Unknown parameter passed: $1"; echo "$HELP_MESSAGE"; exit 1 ;;
  esac
  shift
done

SYMPHONY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if $CLAUDE; then
  CONFIG_DIR="$HOME/.claude"
  filename="CLAUDE.md"
else
  CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"
  filename="AGENTS.md"
fi

src="$SYMPHONY_DIR/AGENTS.md"
dst="$CONFIG_DIR/$filename"

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
echo "  installed $filename"
