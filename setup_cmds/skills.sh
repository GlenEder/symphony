#!/bin/bash
# Symphony setup — symlink skill directories into config dir
#
# Usage: setup skills [--dry-run] [--cursor] [--claude] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup skills [--dry-run] [--cursor] [--claude] [-h/--help]

  Symlink each skills/<name>/ directory into the Maki, Cursor, or Claude config directory.

  --cursor     Target Cursor config (~/.cursor/skills/) instead of Maki
  --claude     Target Claude config (~/.claude/skills/) instead of Maki
  --dry-run    Preview without making changes
  -h, --help   Show this help message"

DRY_RUN=false
CURSOR=false
CLAUDE=false

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --dry-run) DRY_RUN=true ;;
    --cursor) CURSOR=true ;;
    --claude) CLAUDE=true ;;
    -h|--help) echo "$HELP_MESSAGE"; exit 0 ;;
    *) echo "Unknown parameter passed: $1"; echo "$HELP_MESSAGE"; exit 1 ;;
  esac
  shift
done

if $CURSOR && $CLAUDE; then
  echo "--cursor and --claude are mutually exclusive"
  echo "$HELP_MESSAGE"
  exit 1
fi

SYMPHONY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if $CLAUDE; then
  CONFIG_DIR="$HOME/.claude"
elif $CURSOR; then
  CONFIG_DIR="$HOME/.cursor"
else
  CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"
fi

link() {
  local src="$1"
  local dst="$2"

  if [ ! -e "$src" ] && [ ! -L "$src" ]; then
    return
  fi

  if $DRY_RUN; then
    echo "[DRY RUN] ln -sfn $src -> $dst"
    return
  fi

  mkdir -p "$(dirname "$dst")"

  if [ -L "$dst" ] || [ -f "$dst" ]; then
    rm -f "$dst"
  fi

  ln -sfn "$src" "$dst"
  echo "  linked $(basename "$dst")"
}

if [ -d "$SYMPHONY_DIR/skills" ]; then
  for skill_dir in "$SYMPHONY_DIR/skills"/*/; do
    [ -d "$skill_dir" ] || continue
    name="$(basename "$skill_dir")"
    link "$skill_dir" "$CONFIG_DIR/skills/$name"
  done
fi
