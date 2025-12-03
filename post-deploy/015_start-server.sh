#!/usr/bin/env bash

MATH_HOME="/apps/math"
cd $MATH_HOME

echo "$MATH_HOME"
echo "Starting new server..."

# Store output in a logfile and save the PID to a file so we can kill the process later
./dist/server >> /apps/math/gosvr.log 2>&1 & echo $! > /apps/math/gosvr.pid

echo "Verifying the server is running..."
if ! ps -p $(cat /apps/math/gosvr.pid) > /dev/null 2>&1; then
    echo "ERROR: Process is not running!"
    exit 1
fi
echo "OK!"
