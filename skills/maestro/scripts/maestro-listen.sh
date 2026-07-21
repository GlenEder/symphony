#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: maestro-listen.sh --plan-name <name> [options]

  Watch a plan file for changes using fswatch (primary) or stat mtime polling
  (fallback). On file change, fetches and outputs the plan JSON to stdout.

  Does NOT manage heartbeats — run maestro-heartbeat.sh separately.

   --plan-name <name>        Plan name to watch (required, matches .json filename without extension)
  --plan-id <name>          Alias for --plan-name
  --maestro-dir <path>      Path to maestro directory (default: MAESTRO_DIR env or .)
  --port <port>             Maestro server port (default: 8080)
  --timeout <s>             Max seconds to wait for file change (default: 7200, 0 = no limit)
  --poll-fallback-sleep <s> Seconds between stat polls on fallback (default: 2)
  -h, --help                Show this help message

Exit codes:
  0 - File change detected (plan JSON output to stdout)
  1 - Error (bad args, plan file not found, API unreachable, etc.)
  2 - Timeout reached with no file change"

plan_name=""
maestro_dir="${MAESTRO_DIR:-.}"
port=8080
timeout=7200
poll_fallback_sleep=2

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --plan-name|--plan-id)
      plan_name=$2
      shift 2
      ;;

    --maestro-dir)
      maestro_dir=$2
      shift 2
      ;;

    --port)
      port=$2
      shift 2
      ;;

    --timeout)
      timeout=$2
      shift 2
      ;;

    --poll-fallback-sleep)
      poll_fallback_sleep=$2
      shift 2
      ;;

    -h | --help)
      echo "$HELP_MESSAGE"
      exit 0
      ;;

    *)
      echo "Unknown parameter passed: $1"
      echo "$HELP_MESSAGE"
      exit 1
      ;;
  esac
done

if [[ -z "$plan_name" ]]; then
  echo "Error: --plan-name is required" >&2
  echo "$HELP_MESSAGE"
  exit 1
fi

# Determine plans directory: use MAESTRO_PLANS_DIR if set, otherwise default to maestro_dir/plans
if [[ -n "${MAESTRO_PLANS_DIR:-}" ]]; then
  plans_dir="$MAESTRO_PLANS_DIR"
else
  plans_dir="$maestro_dir/plans"
fi

plan_file="$plans_dir/$plan_name.json"
base_url="http://localhost:$port"

if ! [[ -f "$plan_file" ]]; then
  echo "Error: Plan file not found: $plan_file" >&2
  exit 1
fi

# -- File watch (fswatch primary, stat polling fallback) ------------------

watch_loop() {
  if command -v fswatch >/dev/null 2>&1; then
    fswatch -1 --latency 0.5 "$plan_file" >/dev/null 2>&1
  else
    last_mtime=$(stat -f %m "$plan_file" 2>/dev/null || echo "0")
    while true; do
      cur_mtime=$(stat -f %m "$plan_file" 2>/dev/null || echo "0")
      if [[ "$cur_mtime" != "$last_mtime" ]]; then
        break
      fi
      sleep "$poll_fallback_sleep"
    done
  fi
}

if [[ "$timeout" -gt 0 ]]; then
  watch_loop &
  watch_pid=$!

  elapsed=0
  while kill -0 "$watch_pid" 2>/dev/null && [[ "$elapsed" -lt "$timeout" ]]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done

  if kill -0 "$watch_pid" 2>/dev/null; then
    kill "$watch_pid" 2>/dev/null || true
    wait "$watch_pid" 2>/dev/null
    echo "Error: Timeout reached ($timeout seconds) with no file change" >&2
    exit 2
  fi

  wait "$watch_pid" 2>/dev/null
else
  watch_loop
fi

# -- Fetch and output plan JSON -------------------------------------------

plan_json=$(curl -s "$base_url/api/plan/$plan_name")
if [[ -z "$plan_json" ]]; then
  echo "Error: Failed to fetch plan from $base_url/api/plan/$plan_name" >&2
  exit 3
fi

echo "$plan_json"
