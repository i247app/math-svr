#!/bin/sh
# cd /apps/math

# echo "Killing old server..."
# Only attempt to kill the process if the process same as PID file is running
# kill $(cat /apps/math/gosvr.pid) 2>/dev/null

ps -ef | grep mathsvr | grep -v grep

PID_TO_KILL=$(ps -ef | grep 'mathsvr' | grep -v 'grep' | awk '{print $2; exit}')

if [ -n "${PID_TO_KILL}" ]
then
  echo "found pid $PID_TO_KILL"
  echo "kill -HUP $PID_TO_KILL"

  #sudo kill -HUP $PID_TO_KILL
  kill -HUP $PID_TO_KILL

  ps -ef | grep mathsvr | grep -v grep
else
  echo "no pid found. server might not be running"
fi


echo "OK!"
