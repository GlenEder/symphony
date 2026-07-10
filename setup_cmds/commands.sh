#!/bin/bash
# Symphony setup — symlink command files into config dir
#
# Usage: setup commands [--dry-run] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup commands [--dry-run] [-h/--help]

  Symlink each file in commands/ into the Maki config directory.

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

if [ -d "$SYMPHONY_DIR/commands" ]; then
  for cmd_file in "$SYMPHONY_DIR/commands"/*; do
    [ -f "$cmd_file" ] || continue
    link "$cmd_file" "$CONFIG_DIR/commands/$(basename "$cmd_file")"
  done
fi
