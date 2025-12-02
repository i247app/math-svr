#!/usr/bin/env bash

DEST_DIR="/apps/math"
PID_FILE="$DEST_DIR/server.pid"
PORT=8080

cd "$DEST_DIR"

# Stop by PID file
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p $OLD_PID > /dev/null 2>&1; then
        echo "Stopping old server (PID: $OLD_PID)..."
        kill $OLD_PID
        sleep 2
        if ps -p $OLD_PID > /dev/null 2>&1; then
            echo "Force killing old server..."
            kill -9 $OLD_PID
        fi
    fi
    rm -f "$PID_FILE"
fi

# Also kill any process on port 8080 (fallback)
echo "Checking port $PORT..."
PORT_PID=$(sudo lsof -ti:$PORT)
if [ ! -z "$PORT_PID" ]; then
    echo "Found process on port $PORT (PID: $PORT_PID), killing..."
    sudo kill -9 $PORT_PID
    sleep 1
fi

# Start new server
echo "Starting new server..."
nohup ./dist/server > server.log 2>&1 &
NEW_PID=$!
echo $NEW_PID > "$PID_FILE"

# Wait a moment and verify it started
sleep 2
if ps -p $NEW_PID > /dev/null 2>&1; then
    echo "✅ Server started successfully with PID: $NEW_PID"
    echo "📋 Check logs: tail -f $DEST_DIR/server.log"
else
    echo "❌ Server failed to start, check logs"
    exit 1
fi