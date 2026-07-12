#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

HELP_MESSAGE="Usage: maestro-heartbeat.sh --plan-name <name> [options]

  Start a background heartbeat loop for a Maestro plan so the server knows
  the agent is still listening. Runs until killed, the --timeout expires,
  or the parent process (agent/shell) exits.
  Saves its PID to a file so subsequent invocations reuse the same process.

  --plan-name <name>        Plan name (required, matches .toon filename without extension)
  --plan-id <name>          Alias for --plan-name
  --maestro-dir <path>      Path to maestro directory (default: MAESTRO_DIR env or .)
  --port <port>             Maestro server port (default: 8080)
  --interval <s>            Seconds between heartbeats (default: 15)
  --timeout <s>             Max seconds to keep heartbeating (default: 0 = run forever)
  -s, --stop                Stop a running heartbeat for this plan
  -h, --help                Show this help message

Exit codes:
  0 - Heartbeat started (or stopped with --stop)
  1 - Error (bad args, API unreachable)"

plan_name=""
maestro_dir="${MAESTRO_DIR:-.}"
port=8080
interval=15
timeout=0
stop=false

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

    --interval)
      interval=$2
      shift 2
      ;;

    --timeout)
      timeout=$2
      shift 2
      ;;

    -s | --stop)
      stop=true
      shift
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

pid_file="$maestro_dir/.maestro-hb-$plan_name.pid"
base_url="http://localhost:$port"

# Capture the PID of the process that launched us (agent/shell)
# so the heartbeat can self-destruct if the parent dies.
original_ppid=$PPID

# -- Stop mode ------------------------------------------------------------

if [[ "$stop" == true ]]; then
  if [[ ! -f "$pid_file" ]]; then
    echo "No heartbeat running for plan '$plan_name'"
    exit 0
  fi
  existing_pid=$(cat "$pid_file")
  if kill -0 "$existing_pid" 2>/dev/null; then
    kill "$existing_pid" 2>/dev/null || true
    echo "Heartbeat stopped (PID $existing_pid)"
  else
    echo "Stale PID file removed"
  fi
  rm -f "$pid_file"
  curl -s -X POST "$base_url/api/agent/$plan_name/status" \
    -H "Content-Type: application/json" \
    -d '{"status":"offline"}' >/dev/null 2>&1 || true
  exit 0
fi

# -- Start mode -----------------------------------------------------------

# Reuse existing heartbeat if still alive
if [[ -f "$pid_file" ]]; then
  existing_pid=$(cat "$pid_file")
  if kill -0 "$existing_pid" 2>/dev/null; then
    exit 0
  fi
  rm -f "$pid_file"
fi

# Wait for server readiness
while ! curl -s "$base_url/api/plans" >/dev/null 2>&1; do
  sleep 0.2
done

# Start heartbeat in a detached process group so it survives the script exit
(
  # If timeout is set, kill self after that many seconds
  if [[ "$timeout" -gt 0 ]]; then
    (sleep "$timeout" && kill $$ 2>/dev/null) &
  fi

  while true; do
    # Self-destruct if the parent process (agent/shell) is gone
    if ! kill -0 "$original_ppid" 2>/dev/null; then
      rm -f "$pid_file"
      curl -s -X POST "$base_url/api/agent/$plan_name/status" \
        -H "Content-Type: application/json" \
        -d '{"status":"offline"}' >/dev/null 2>&1 || true
      exit 0
    fi

    curl -s -X POST "$base_url/api/agent/$plan_name/heartbeat" >/dev/null 2>&1 || true
    sleep "$interval"
  done
) &

echo $! > "$pid_file"
