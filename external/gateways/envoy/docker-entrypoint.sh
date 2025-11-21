#!/bin/bash

echo "Starting OpenFTV Envoy AuthZEN plugin"
/usr/local/bin/openftv &

./envoy-entry.sh &

sleep 1

echo "services started"
PID1=`pgrep openftv`
PID2=`pgrep envoy`

cleanup() {
  echo "killing services"
  kill -l $PID2
  kill -l $PID1

  echo "exiting"
  exit 0
}

trap cleanup SIGINT SIGQUIT SIGHUP SIGTERM ERR

wait "$PID1"
