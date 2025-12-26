#!/bin/bash
# Test 5: Interactive program (Python REPL)
# Tests that ptyee works with interactive programs

set -e

SOCKET="/tmp/ptyee-test-python.sock"
OUTPUT_FILE="/tmp/ptyee-test-python.out"

echo "Test 5: Interactive Python REPL"

# Cleanup
rm -f "$SOCKET" "$OUTPUT_FILE"

# Start ptyee with Python REPL
(sleep 1; echo "print('hello from python')"; sleep 0.5; echo "exit()") | \
    timeout 5 ./build/ptyee --socket "$SOCKET" -- python3 &
PTYEE_PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Read from socket
timeout 4 nc -U "$SOCKET" > "$OUTPUT_FILE" 2>&1 &
NC_PID=$!

wait $NC_PID 2>/dev/null || true
wait $PTYEE_PID 2>/dev/null || true

# Check for Python output
if grep -q "hello from python" "$OUTPUT_FILE"; then
    echo "PASS: Python REPL output received via socket"
    grep "hello from python" "$OUTPUT_FILE"
else
    echo "FAIL: Python output not found in socket data"
    echo "Output received:"
    cat "$OUTPUT_FILE"
    exit 1
fi

# Cleanup
rm -f "$SOCKET" "$OUTPUT_FILE"

echo "Test 5: PASSED"
