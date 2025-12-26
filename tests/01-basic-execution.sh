#!/bin/bash
# Test 1: Basic execution
# Tests that ptyee runs a command and outputs are sent to the socket

set -e

SOCKET="/tmp/ptyee-test-basic.sock"
OUTPUT_FILE="/tmp/ptyee-test-basic.out"

echo "Test 1: Basic execution"

# Cleanup from previous runs
rm -f "$SOCKET" "$OUTPUT_FILE"

# Start ptyee in background (waits for connection by default)
./build/ptyee --socket "$SOCKET" -- echo "Hello from test 1" 2>/dev/null &
PTYEE_PID=$!

# Wait for socket to be created
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

if [ ! -S "$SOCKET" ]; then
    echo "FAIL: Socket was not created"
    kill $PTYEE_PID 2>/dev/null || true
    exit 1
fi

# Connect to socket - command will start once we connect
nc -U "$SOCKET" > "$OUTPUT_FILE" 2>&1 &
NC_PID=$!

# Wait for command to complete (with timeout)
for i in {1..20}; do
    if ! ps -p $PTYEE_PID > /dev/null 2>&1; then
        break
    fi
    sleep 0.1
done

# Kill processes if still running
kill $PTYEE_PID $NC_PID 2>/dev/null || true
sleep 0.2

# Check output
if [ -s "$OUTPUT_FILE" ] && grep -q "Hello from test 1" "$OUTPUT_FILE"; then
    echo "PASS: Received expected output from socket"
else
    echo "FAIL: Expected output not found"
    echo "Got:"
    cat "$OUTPUT_FILE"
    rm -f "$SOCKET" "$OUTPUT_FILE"
    exit 1
fi

# Cleanup
rm -f "$SOCKET" "$OUTPUT_FILE"

echo "Test 1: PASSED"
