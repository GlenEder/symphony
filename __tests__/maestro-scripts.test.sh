#!/bin/bash
#
# Tests for maestro shell scripts.
# Usage: bash __tests__/maestro-scripts.test.sh
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HEARTBEAT="$SCRIPT_DIR/../skills/maestro/scripts/maestro-heartbeat.sh"
LISTEN="$SCRIPT_DIR/../skills/maestro/scripts/maestro-listen.sh"
errors=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; errors=$((errors + 1)); }

echo "=== maestro-heartbeat.sh ==="

# Test --help mentions --plan-id
if grep -q -- '--plan-id' "$HEARTBEAT"; then
	pass "Help message mentions --plan-id"
else
	fail "Help message does NOT mention --plan-id"
fi

# Test flag parsing accepts --plan-id (verify by running with --help to see it's listed)
output=$(bash "$HEARTBEAT" --help 2>&1 || true)
if echo "$output" | grep -q -- "--plan-id"; then
	pass "--plan-id listed in help output"
else
	fail "--plan-id not in help output"
fi

# Test that --plan-id does not produce "Unknown parameter" error by parsing args
# (the script will fail later with a real error since no server is running)
output=$(bash "$HEARTBEAT" --plan-id testplan --timeout 1 2>&1 || true)
if ! echo "$output" | grep -q "Unknown parameter"; then
	pass "--plan-id does not produce Unknown parameter error"
else
	fail "--plan-id produced Unknown parameter error"
fi

echo "=== maestro-listen.sh ==="

# Test --help mentions --plan-id
if grep -q -- '--plan-id' "$LISTEN"; then
	pass "Help message mentions --plan-id"
else
	fail "Help message does NOT mention --plan-id"
fi

# Test flag parsing accepts --plan-id (will fail on missing plan file, not unknown flag)
output=$(bash "$LISTEN" --plan-id testplan 2>&1 || true)
if echo "$output" | grep -q "Error: Plan file not found\|Error:.*--plan-name"; then
	pass "--plan-id accepted as flag (script parsed it)"
else
	fail "--plan-id flag not parsed correctly (output: $output)"
fi

echo ""
if [ $errors -eq 0 ]; then
	echo "All tests passed."
else
	echo "$errors test(s) failed."
	exit 1
fi
