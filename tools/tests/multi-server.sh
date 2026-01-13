#!/bin/bash
set -euo pipefail

LOGDIR="logs"
BINDIR="bin"
BINARY_NAME="$BINDIR/dord"
mkdir -p "$LOGDIR" "$BINDIR"

USE_CONSOLE=false
NUM_SERVERS=5
START_PORT=62503
GO_ARGS=""
SERVER_ARGS=""

function help {
  echo "Usage: $0 [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  --console, -c           Display logs in console (colored)"
  echo "  --num-servers, -n N     Number of servers to start (default: 5)"
  echo "  --start-port, -p PORT   Starting port number (default: 62503)"
  echo "  --go-args ARGS          Additional Go build arguments"
  echo "  --server-args ARGS      Additional server runtime arguments (e.g., '--log-level debug')"
  echo "  --help, -h              Show this help message"
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --console|-c)
      USE_CONSOLE=true
      shift
      ;;
    --num-servers|-n)
      NUM_SERVERS="$2"
      shift 2
      ;;
    --start-port|-p)
      START_PORT="$2"
      shift 2
      ;;
    --go-args)
      GO_ARGS="$2"
      shift 2
      ;;
    --server-args)
      SERVER_ARGS="$2"
      shift 2
      ;;
    --help|-h)
      help
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

function cleanup {
  echo ""
  echo "[CLEANUP] Stopping DOR servers..."
  kill $(jobs -p) 2>/dev/null || true
  echo "[CLEANUP] Done."
  exit 0
}

trap cleanup EXIT INT TERM

echo "[BUILD] Compiling dord to $BINARY_NAME..."
go build $GO_ARGS -o "$BINARY_NAME" cmd/dord/main.go
echo "[BUILD] Success."

COLORS=(
  "\033[31m"
  "\033[32m"
  "\033[33m"
  "\033[34m"
  "\033[35m"
  "\033[36m"
)
RESET="\033[0m"

for i in $(seq 1 $NUM_SERVERS); do
  DIR="/tmp/dor_id/s_$i"
  PORT=$((START_PORT + i - 1))
  mkdir -p "$DIR"
  COLOR="${COLORS[$(( (i-1) % ${#COLORS[@]} ))]}"

  CMD="$BINARY_NAME --id-dir=$DIR --port=$PORT $SERVER_ARGS"

  if [ "$USE_CONSOLE" = true ]; then
    echo -e "${COLOR}[START] Server on port $PORT${RESET}"
    $CMD 2>&1 | while IFS= read -r line; do
        echo -e "${COLOR}[${PORT}]${RESET} $line"
      done &
  else
    $CMD > "$LOGDIR/server_$PORT.log" 2>&1 &
  fi
done

if [ "$USE_CONSOLE" = true ]; then
  echo "[INFO] $NUM_SERVERS DOR servers running (Console Mode). Ctrl+C to stop."
else
  echo "[INFO] $NUM_SERVERS DOR servers running. Logs in '$LOGDIR'. Ctrl+C to stop."
fi

wait
