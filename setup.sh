#!/bin/bash
# Symphony repo setup — symlinks contents into ~/.config/maki/
# Run once on each device after cloning. Safe to re-run.
#
# Usage: ./setup.sh [--dry-run] [--maestro-dir <path>] [--config-file-path <path>]

set -euo pipefail

DRY_RUN=false
MAESTRO_DIR_ARG=""
CONFIG_FILE_PATH=""

while [[ $# -gt 0 ]]; do
    case "$1" in
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
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
    shift
done

CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/maki"
SYMPHONY_DIR="$(cd "$(dirname "$0")" && pwd)"
RC_FILE="${CONFIG_FILE_PATH:-$HOME/.zshrc}"

if $DRY_RUN; then
    echo "[DRY RUN] Would link from: $SYMPHONY_DIR"
    echo "[DRY RUN] Would link into: $CONFIG_DIR"
    echo ""
fi

link() {
    local src="$1"
    local dst="$2"

    if [ ! -e "$src" ] && [ ! -L "$src" ]; then
        return  # source doesn't exist, skip
    fi

    if $DRY_RUN; then
        echo "[DRY RUN] ln -sfn $src -> $dst"
        return
    fi

    # Create parent dir if needed
    mkdir -p "$(dirname "$dst")"

    # Remove existing file/symlink (but not a directory we don't own)
    if [ -L "$dst" ] || [ -f "$dst" ]; then
        rm -f "$dst"
    fi

    ln -sfn "$src" "$dst"
    echo "  linked $(basename "$dst")"
}

echo "=== Symphony Setup ==="
echo "  Repo: $SYMPHONY_DIR"
echo "  Target: $CONFIG_DIR"
echo ""

install_agents_md() {
    local src="$SYMPHONY_DIR/AGENTS.md"
    local dst="$CONFIG_DIR/AGENTS.md"
    local inject="$SYMPHONY_DIR/maestro/AGENTS_INJECT.md"
    local begin_marker="# <<< MAESTRO AGENTS INJECT"
    local end_marker="# >>> MAESTRO AGENTS INJECT"

    if [ ! -f "$src" ]; then
        return
    fi

    if $DRY_RUN; then
        echo "[DRY RUN] cp $src -> $dst"
        if [ -f "$inject" ]; then
            echo "[DRY RUN] Would inject $inject into $dst between:"
            echo "[DRY RUN]   $begin_marker"
            echo "[DRY RUN]   ... content from maestro/AGENTS_INJECT.md ..."
            echo "[DRY RUN]   $end_marker"
        fi
        return
    fi

    mkdir -p "$(dirname "$dst")"

    # Break existing symlink if it points to our source (cp refuses identical files)
    if [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ]; then
        rm "$dst"
    fi

    # Copy the base AGENTS.md (not a symlink, so target is writable)
    cp "$src" "$dst"
    echo "  installed AGENTS.md"

    # Inject maestro rules if the inject file exists
    if [ -f "$inject" ]; then
        # Strip any existing sentinel block from target
        if grep -qF "$begin_marker" "$dst" 2>/dev/null; then
            # Use sed to delete from begin to end marker (inclusive)
            sed -i '' "/^$begin_marker$/,/^$end_marker$/d" "$dst"
        fi

        # Append fresh injection block
        {
            echo ""
            echo "$begin_marker"
            cat "$inject"
            echo "$end_marker"
        } >> "$dst"
        echo "  injected maestro/AGENTS_INJECT.md"
    fi
}

# --- AGENTS.md ---
install_agents_md

# --- Skills (individual subdirectories) ---
if [ -d "$SYMPHONY_DIR/skills" ]; then
    for skill_dir in "$SYMPHONY_DIR/skills"/*/; do
        [ -d "$skill_dir" ] || continue
        name="$(basename "$skill_dir")"
        link "$skill_dir" "$CONFIG_DIR/skills/$name"
    done
fi

# --- Commands (individual files) ---
if [ -d "$SYMPHONY_DIR/commands" ]; then
    for cmd_file in "$SYMPHONY_DIR/commands"/*; do
        [ -f "$cmd_file" ] || continue
        link "$cmd_file" "$CONFIG_DIR/commands/$(basename "$cmd_file")"
    done
fi

# --- Providers (individual scripts) ---
if [ -d "$SYMPHONY_DIR/providers" ]; then
    for prov_file in "$SYMPHONY_DIR/providers"/*; do
        [ -f "$prov_file" ] || continue
        link "$prov_file" "$CONFIG_DIR/providers/$(basename "$prov_file")"
    done
fi

# --- init.lua ---
link "$SYMPHONY_DIR/init.lua" "$CONFIG_DIR/init.lua"

echo ""
echo "=== Building Maestro & Updating PATH ==="

MAESTRO_DIR="${MAESTRO_DIR_ARG:-$SYMPHONY_DIR/maestro}"

if $DRY_RUN; then
    echo "[DRY RUN] cd $MAESTRO_DIR && go build -o maestro ."
    echo "[DRY RUN] Would add to $RC_FILE:"
    echo "[DRY RUN]   # Maestro Planning Server"
    echo "[DRY RUN]   export PATH=\"\$PATH:$MAESTRO_DIR\""
    echo "[DRY RUN]   export MAESTRO_DIR=\"$MAESTRO_DIR\""
    echo "[DRY RUN]   export MAESTRO_PLANS_DIR=\"$MAESTRO_DIR/plans\""
else
    # Resolve to absolute path (may fail if dir doesn't exist)
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
fi

echo "Run 'maki' to verify."
