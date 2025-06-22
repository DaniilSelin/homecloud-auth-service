#!/bin/bash

echo "Starting auth service..."
go run cmd/server/main.go &
SERVER_PID=$!

echo "Waiting for server to start..."
sleep 3

echo "Testing user registration..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test4@example.com","username":"testuser4","password":"password123"}'

echo ""
echo "Testing another user registration..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test5@example.com","username":"testuser5","password":"password123"}'

echo ""
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed!" 