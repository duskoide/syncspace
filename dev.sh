#!/bin/bash
set -e

cleanup() {
  echo ""
  echo "Shutting down..."
  kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
  wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
  echo "Done."
}

trap cleanup EXIT INT TERM

mkdir -p dev-data/uploads

echo "=== SyncSpace Dev Mode ==="
echo "Backend:  http://localhost:8081"
echo "Frontend: http://localhost:5173"
echo "Database: ./dev-data/syncspace.db"
echo "Press Ctrl+C to stop"
echo ""

cd backend
SYNCSPACE_ADDR=:8081 \
SYNCSPACE_DB_PATH=../dev-data/syncspace.db \
SYNCSPACE_UPLOAD_DIR=../dev-data/uploads \
go run ./cmd/syncspace &
BACKEND_PID=$!
cd ..

cd frontend
VITE_API_URL=http://localhost:8081 npm run dev &
FRONTEND_PID=$!
cd ..

wait
