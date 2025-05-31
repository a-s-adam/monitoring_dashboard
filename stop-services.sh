#!/bin/bash

PID_FILE=".service_pids"

if [ ! -f $PID_FILE ]; then
    echo "No services are running"
    exit 0
fi

echo "Stopping all services..."

while read pid; do
    if ps -p $pid > /dev/null; then
        echo "Stopping process $pid"
        kill $pid 2>/dev/null
    fi
done < $PID_FILE

rm $PID_FILE
echo "All services stopped" 