#!/usr/bin/env bash
# bin/verify-graceful-shutdown.sh — verify session serialization on shutdown.
#
# Positive test: SIGTERM → gex onShutdown hooks fire → session file written.
# Regression test: SIGHUP → process dies via Go default → session file NOT
# written (until/unless a SIGHUP handler is added to the app).
#
# Usage:
#   ./bin/verify-graceful-shutdown.sh [session-file-path]
#
# Prereqs:
#   - A valid .env in the project root with SERIALIZED_SESSION_FILE set
#   - `make build` must succeed
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

SESSION_FILE="${1:-}"
BIN="./dist/mathsvr"
LOG_FILE="/tmp/mathsvr-verify.log"
STARTUP_WAIT_SECS=5
SHUTDOWN_WAIT_SECS=12

if [ -z "$SESSION_FILE" ]; then
  # Try to read SERIALIZED_SESSION_FILE from .env if the caller didn't supply one.
  if [ -f .env ]; then
    SESSION_FILE=$(grep -E '^SERIALIZED_SESSION_FILE=' .env | tail -n1 | cut -d= -f2- | tr -d '"' | tr -d "'")
  fi
fi

if [ -z "$SESSION_FILE" ]; then
  echo "FAIL: session file path not provided and SERIALIZED_SESSION_FILE not set in .env"
  exit 2
fi

echo "Session file: $SESSION_FILE"
echo "Binary:       $BIN"
echo "Log:          $LOG_FILE"
echo ""

echo "Building..."
make build >/dev/null

cleanup() {
  if [ -n "${PID:-}" ] && kill -0 "$PID" 2>/dev/null; then
    kill -9 "$PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

start_server() {
  : > "$LOG_FILE"
  "$BIN" >>"$LOG_FILE" 2>&1 &
  PID=$!

  # Wait for the server to finish Init() and reload sessions. Poll for a known
  # log line rather than sleeping blindly, so slower machines don't flake.
  for i in $(seq 1 "$STARTUP_WAIT_SECS"); do
    if grep -q "Starting math" "$LOG_FILE" 2>/dev/null; then
      sleep 1  # grace for middleware registration
      return 0
    fi
    sleep 1
  done
  echo "FAIL: server did not start within ${STARTUP_WAIT_SECS}s — see $LOG_FILE"
  tail -40 "$LOG_FILE" || true
  return 1
}

wait_exit() {
  local pid=$1
  for i in $(seq 1 "$SHUTDOWN_WAIT_SECS"); do
    if ! kill -0 "$pid" 2>/dev/null; then
      return 0
    fi
    sleep 1
  done
  return 1
}

# ── Positive test: SIGTERM ────────────────────────────────────────────────
echo "── Positive test: SIGTERM should write session file ──"
rm -f "$SESSION_FILE"
start_server || exit 1

if ! kill -0 "$PID" 2>/dev/null; then
  echo "FAIL: server died during startup"
  tail -40 "$LOG_FILE" || true
  exit 1
fi

echo "Sending SIGTERM to PID $PID ..."
kill -TERM "$PID"

if ! wait_exit "$PID"; then
  echo "FAIL: server did not exit within ${SHUTDOWN_WAIT_SECS}s after SIGTERM"
  kill -9 "$PID" 2>/dev/null || true
  exit 1
fi

if [ ! -s "$SESSION_FILE" ]; then
  echo "FAIL: session file $SESSION_FILE not written after SIGTERM"
  echo "---- server log ----"
  tail -60 "$LOG_FILE" || true
  exit 1
fi

echo "PASS: session file written after SIGTERM ($(wc -c <"$SESSION_FILE") bytes)"
echo ""

# ── Regression test: SIGHUP ───────────────────────────────────────────────
echo "── Regression test: SIGHUP (the old deploy-script bug) ──"
rm -f "$SESSION_FILE"
start_server || exit 1

echo "Sending SIGHUP to PID $PID ..."
kill -HUP "$PID" 2>/dev/null || true

if ! wait_exit "$PID"; then
  # If the process is still alive after SIGHUP, someone added a SIGHUP handler
  # that doesn't exit — that's fine but unexpected.
  echo "NOTE: process still alive after SIGHUP — forcing kill"
  kill -9 "$PID" 2>/dev/null || true
  sleep 1
fi

if [ -s "$SESSION_FILE" ]; then
  echo "NOTE: SIGHUP also produced a session file — a SIGHUP handler has been added to the app. Good."
else
  echo "NOTE: SIGHUP did NOT write a session file — expected under current gex behavior."
  echo "      This is why the old 'kill -HUP' in pre-deploy/0_stop-server.sh was losing sessions."
fi

echo ""
echo "Done."
