#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: maestro-listen.sh --plan-name <name> [options]

  Start the Maestro plan listener: background heartbeat loop + file watch with
  fswatch (primary) and stat mtime polling (fallback). On file change, fetches
  the plan JSON from the API and outputs it to stdout.

  --plan-name <name>        Plan name to watch (required, matches .toon filename without extension)
  --maestro-dir <path>      Path to maestro directory (default: current dir)
  --port <port>             Maestro server port (default: 8080)
  --heartbeat-interval <s>  Seconds between heartbeats (default: 15)
  --timeout <s>             Max seconds to wait for file change (default: 7200, 0 = no limit)
  --poll-fallback-sleep <s> Seconds between stat polls on fallback (default: 2)
  -h, --help                Show this help message

Exit codes:
  0 - File change detected (plan JSON output to stdout)
  1 - Error (bad args, plan file not found, API unreachable, etc.)
  2 - Timeout reached with no file change"

plan_name=""
maestro_dir="maestro"
port=8080
heartbeat_interval=15
timeout=7200
poll_fallback_sleep=2

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --plan-name)
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

    --heartbeat-interval)
      heartbeat_interval=$2
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

    -h|--help)
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

plan_file="$maestro_dir/plans/$plan_name.toon"
base_url="http://localhost:$port"
heartbeat_pid=""

# -- Cleanup on exit -------------------------------------------------------

cleanup() {
  if [[ -n "$heartbeat_pid" ]] && kill -0 "$heartbeat_pid" 2>/dev/null; then
    kill "$heartbeat_pid" 2>/dev/null || true
  fi
  curl -s -X POST "$base_url/api/agent/$plan_name/status" \
    -H "Content-Type: application/json" \
    -d '{"status":"offline"}' > /dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

# -- Heartbeat loop (background) ------------------------------------------

while true; do
  curl -s -X POST "$base_url/api/agent/$plan_name/heartbeat" > /dev/null 2>&1 || true
  sleep "$heartbeat_interval"
done &
heartbeat_pid=$!

# -- Wait for server readiness --------------------------------------------

while ! curl -s "$base_url/api/plans" > /dev/null 2>&1; do
  sleep 0.2
done

# -- File watch (fswatch primary, stat polling fallback) ------------------

watch_loop() {
  if command -v fswatch > /dev/null 2>&1; then
    fswatch -1 --latency 0.5 "$plan_file" > /dev/null 2>&1
  else
    if ! [[ -f "$plan_file" ]]; then
      echo "Error: Plan file not found: $plan_file" >&2
      exit 1
    fi
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
  # Run watch in background so we can apply a timeout
  watch_loop &
  watch_pid=$!

  elapsed=0
  while kill -0 "$watch_pid" 2>/dev/null && [[ "$elapsed" -lt "$timeout" ]]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done

  if kill -0 "$watch_pid" 2>/dev/null; then
    # watch still running → timeout
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
