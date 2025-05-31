#!/bin/bash

# Function to check if a port is in use
check_port() {
    if lsof -i :$1 > /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to wait for a service to be ready
wait_for_service() {
    local port=$1
    local service=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for $service to be ready on port $port..."
    while ! check_port $port; do
        if [ $attempt -ge $max_attempts ]; then
            echo "$service failed to start"
            exit 1
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    echo "$service is ready!"
}

# Store PIDs in a file for stopping later
PID_FILE=".service_pids"
touch $PID_FILE

# Kill any existing processes
if [ -f $PID_FILE ]; then
    while read pid; do
        if ps -p $pid > /dev/null; then
            kill $pid 2>/dev/null
        fi
    done < $PID_FILE
    rm $PID_FILE
fi

echo "Starting all services..."

# Start Rust service (port 8000)
cd rust-hw-metrics
cargo run &
RUST_PID=$!
echo $RUST_PID > $PID_FILE
cd ..
wait_for_service 8000 "Rust metrics service"

# Start Python service (port 8001)
cd py-detector
source venv/bin/activate
python app.py &
PYTHON_PID=$!
echo $PYTHON_PID >> $PID_FILE
cd ..
wait_for_service 8001 "Python anomaly detector"

# Start Go service (port 8002)
cd go-orchestrator
go run main.go &
GO_PID=$!
echo $GO_PID >> $PID_FILE
cd ..
wait_for_service 8002 "Go orchestrator"

# Start web dashboard (port 3000)
cd web-dashboard
python -m http.server 3000 &
WEB_PID=$!
echo $WEB_PID >> $PID_FILE
cd ..
wait_for_service 3000 "Web dashboard"

echo "All services are running!"
echo "Open http://localhost:3000 in your browser"
echo "To stop all services, run: ./stop-services.sh"

# Keep the script running to maintain the services
wait 