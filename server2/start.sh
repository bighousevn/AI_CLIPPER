#!/bin/sh


/app/cmd/worker/worker &


/app/cmd/worker-email/worker-email &

echo "Switching to server directory..."
cd /app/cmd/server

echo "Starting API Server..."

./server