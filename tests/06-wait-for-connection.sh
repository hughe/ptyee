#!/bin/bash
# Test 6: Wait for connection feature
# Tests that ptyee waits for a socket connection before starting the command

SOCKET="/tmp/ptyee-test-wait.sock"
OUTPUT="/tmp/ptyee-test-wait.out"
rm -f "$SOCKET" "$OUTPUT"

echo "Test 6: Wait for connection before starting command"

# Start ptyee - it should wait for connection
./build/ptyee --socket "$SOCKET" -- echo "hello" 2>/dev/null &
PID=$!

# Wait for socket to be created
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Check if socket exists
if [ ! -S "$SOCKET" ]; then
    echo "FAIL: Socket not created after 1 second"
    kill $PID 2>/dev/null || true
    exit 1
fi

# Check if process is still waiting (command shouldn't have run yet)
if ! ps -p $PID > /dev/null 2>&1; then
    echo "FAIL: Process exited before connection (should be waiting)"
    exit 1
fi

echo "PASS: Process waiting for connection"

# Now connect
nc -U "$SOCKET" > "$OUTPUT" 2>&1 &
NC_PID=$!

# Wait for command to complete
sleep 1

# Check output
if grep -q "hello" "$OUTPUT"; then
    echo "PASS: Received output after connection"
else
    echo "FAIL: No output received"
    cat "$OUTPUT"
    kill $NC_PID $PID 2>/dev/null || true
    rm -f "$SOCKET" "$OUTPUT"
    exit 1
fi

# Cleanup
kill $NC_PID $PID 2>/dev/null || true
rm -f "$SOCKET" "$OUTPUT"

echo "Test 6: PASSED"
