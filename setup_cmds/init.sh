#!/bin/bash
# Symphony setup — symlink init.lua into config dir
#
# Usage: setup init [--dry-run] [-h/--help]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: setup init [--dry-run] [-h/--help]

  Symlink init.lua into the Maki config directory.

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

link "$SYMPHONY_DIR/init.lua" "$CONFIG_DIR/init.lua"
