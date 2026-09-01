#!/usr/bin/env bash
echo "Starting mathsvr via systemd..."
sudo systemctl start mathsvr.service
sleep 2
if ! sudo systemctl is-active --quiet mathsvr.service; then
    echo "ERROR: mathsvr.service failed to start. Last 50 log lines:"
    sudo journalctl -u mathsvr.service -n 50 --no-pager
    exit 1
fi
echo "OK!"

# MATH_HOME="/apps/math"
# cd $MATH_HOME

# echo "$MATH_HOME"
# echo "Starting new server..."

# # Store output in a logfile and save the PID to a file so we can kill the process later
# ./dist/mathsvr >> /apps/math/mathsvr.log 2>&1 & echo $! > /apps/math/mathsvr.pid

# echo "Verifying the server is running..."
# if ! ps -p $(cat /apps/math/mathsvr.pid) > /dev/null 2>&1; then
#     echo "ERROR: Process is not running!"
#     exit 1
# fi
# echo "OK!"
