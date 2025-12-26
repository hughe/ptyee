#!/bin/bash
# Test 2: Tag output format
# Tests that --tag-output prefixes bytes with 'T' or 'P'

set -e

SOCKET="/tmp/ptyee-test-tag.sock"
OUTPUT_FILE="/tmp/ptyee-test-tag.out"

echo "Test 2: Tag output format"

# Cleanup from previous runs
rm -f "$SOCKET" "$OUTPUT_FILE"

# Start ptyee with --tag-output (waits for connection by default)
./build/ptyee --socket "$SOCKET" --tag-output -- echo "hi" 2>/dev/null &
PTYEE_PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Connect and read tagged output - command starts when we connect
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

# Check for 'P' and 'T' tags in output using hex dump
if [ -s "$OUTPUT_FILE" ]; then
    echo "PASS: Received data from socket"
    echo "Output (hex):"
    xxd "$OUTPUT_FILE" | head -3

    # Check for P (0x50) or T (0x54) tags
    if xxd "$OUTPUT_FILE" | grep -q " 50\| 54"; then
        echo "PASS: Found P or T tags in output"
    else
        echo "FAIL: No P/T tags found"
        rm -f "$SOCKET" "$OUTPUT_FILE"
        exit 1
    fi
else
    echo "FAIL: No data received from socket"
    rm -f "$SOCKET" "$OUTPUT_FILE"
    exit 1
fi

# Cleanup
rm -f "$SOCKET" "$OUTPUT_FILE"

echo "Test 2: PASSED"
