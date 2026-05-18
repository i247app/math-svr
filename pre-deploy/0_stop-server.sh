#!/bin/sh
# cd /apps/math

echo "Stopping old server..."

PID_FILE=/apps/math/mathsvr.pid

# Prefer the PID file written by post-deploy/015_start-server.sh; fall back to
# pgrep -x (exact binary name match, so it won't catch mathsvr-amd64).
PID_TO_KILL=""
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  PID_TO_KILL=$(cat "$PID_FILE")
else
  PID_TO_KILL=$(pgrep -x mathsvr | head -n1)
fi

ps -ef | grep mathsvr | grep -v grep

if [ -n "${PID_TO_KILL}" ]
then
  echo "found pid $PID_TO_KILL"
  # IMPORTANT: use SIGTERM, not SIGHUP. The gex server subscribes only to
  # SIGINT and SIGTERM via signal.NotifyContext; SIGHUP uses Go's default
  # disposition and kills the process immediately, bypassing the onShutdown
  # hooks that serialize sessions to disk.
  echo "kill -TERM $PID_TO_KILL"
  kill -TERM "$PID_TO_KILL"

  ps -ef | grep mathsvr | grep -v grep
else
  echo "no pid found. server might not be running"
fi


echo "OK!"
