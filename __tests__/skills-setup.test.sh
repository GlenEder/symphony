#!/bin/bash
#
# Tests for setup_cmds/skills.sh stale-link cleanup.
# Usage: bash __tests__/skills-setup.test.sh
#
# Runs the real script against a temp HOME so no real config is touched.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILLS_SH="$REPO/setup_cmds/skills.sh"
errors=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; errors=$((errors + 1)); }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
CFG="$TMP/.config/maki/skills"
mkdir -p "$CFG"

# Arrange: one stale repo link (source deleted), one foreign dangling link, one live repo link
ln -s "$REPO/skills/maestro/" "$CFG/maestro"       # stale — skill deleted from this repo
ln -s "/tmp/not-from-symphony" "$CFG/other"       # dangling but not ours — must survive
ln -s "$REPO/skills/grilling/" "$CFG/grilling"    # live — must survive

# Dry run: reports the stale link, removes nothing
out="$(XDG_CONFIG_HOME= HOME="$TMP" bash "$SKILLS_SH" --dry-run 2>&1)"
if echo "$out" | grep -q "rm stale link .*$CFG/maestro"; then
	pass "Dry run reports the stale link"
else
	fail "Dry run does not report the stale link"
fi
if [ -L "$CFG/maestro" ]; then
	pass "Dry run removed nothing"
else
	fail "Dry run removed the link — must not modify"
fi

# Real run: removes only the stale repo link
XDG_CONFIG_HOME= HOME="$TMP" bash "$SKILLS_SH" >/dev/null 2>&1
if [ ! -e "$CFG/maestro" ] && [ ! -L "$CFG/maestro" ]; then
	pass "Stale repo link removed"
else
	fail "Stale repo link still present"
fi
if [ -L "$CFG/other" ]; then
	pass "Foreign link untouched"
else
	fail "Foreign link was removed — out of scope"
fi
if [ -e "$CFG/grilling/" ]; then
	pass "Live skill link intact and target exists"
else
	fail "Live skill link broken"
fi

# Idempotent re-run: clean, exits 0
if XDG_CONFIG_HOME= HOME="$TMP" bash "$SKILLS_SH" >/dev/null 2>&1; then
	pass "Re-run exits 0"
else
	fail "Re-run exits non-zero"
fi

echo "=== skills-setup: $errors failure(s) ==="
[ "$errors" -gt 0 ] && exit 1
exit 0
