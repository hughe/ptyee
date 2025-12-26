#!/bin/bash
# Test 3: Exit code propagation
# Tests that ptyee exits with the same code as the child process

SOCKET="/tmp/ptyee-test-exitcode.sock"

echo "Test 3: Exit code propagation"

# Test with exit code 0
rm -f "$SOCKET"
./build/ptyee --socket "$SOCKET" -- bash -c 'exit 0' 2>/dev/null &
PTYEE_PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Connect to socket to start command
nc -U "$SOCKET" > /dev/null 2>&1 &
NC_PID=$!

# Wait for ptyee to finish
wait $PTYEE_PID
EXIT_CODE=$?

kill $NC_PID 2>/dev/null || true
rm -f "$SOCKET"

if [ $EXIT_CODE -eq 0 ]; then
    echo "PASS: Exit code 0 propagated correctly"
else
    echo "FAIL: Expected exit code 0, got $EXIT_CODE"
    exit 1
fi

# Test with exit code 42
rm -f "$SOCKET"
./build/ptyee --socket "$SOCKET" -- bash -c 'exit 42' 2>/dev/null &
PTYEE_PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Connect to socket
nc -U "$SOCKET" > /dev/null 2>&1 &
NC_PID=$!

# Wait for ptyee
wait $PTYEE_PID
EXIT_CODE=$?

kill $NC_PID 2>/dev/null || true
rm -f "$SOCKET"

if [ $EXIT_CODE -eq 42 ]; then
    echo "PASS: Exit code 42 propagated correctly"
else
    echo "FAIL: Expected exit code 42, got $EXIT_CODE"
    exit 1
fi

# Test with exit code 1
rm -f "$SOCKET"
./build/ptyee --socket "$SOCKET" -- bash -c 'exit 1' 2>/dev/null &
PTYEE_PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Connect to socket
nc -U "$SOCKET" > /dev/null 2>&1 &
NC_PID=$!

# Wait for ptyee
wait $PTYEE_PID
EXIT_CODE=$?

kill $NC_PID 2>/dev/null || true
rm -f "$SOCKET"

if [ $EXIT_CODE -eq 1 ]; then
    echo "PASS: Exit code 1 propagated correctly"
else
    echo "FAIL: Expected exit code 1, got $EXIT_CODE"
    exit 1
fi

echo "Test 3: PASSED"
