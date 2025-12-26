#!/bin/bash
# Debug script to examine tag output format in detail

SOCKET="/tmp/test-tag-debug.sock"
OUTPUT="/tmp/test-tag-debug.out"
rm -f "$SOCKET" "$OUTPUT"

echo "Debug: Testing tag output format"

# Start with tag output - command waits before outputting to give client time to connect
./build/ptyee --socket "$SOCKET" --tag-output -- bash -c 'sleep 0.3; echo "hi"; sleep 1' 2>/dev/null &
PID=$!

# Wait for socket
for i in {1..10}; do
    if [ -S "$SOCKET" ]; then
        break
    fi
    sleep 0.1
done

# Connect and read BEFORE the command outputs
nc -U "$SOCKET" > "$OUTPUT" 2>&1 &
NCPID=$!

# Wait
sleep 1.5

# Kill
kill $NCPID $PID 2>/dev/null || true

# Show what we got
echo "=== Raw output ==="
cat "$OUTPUT"
echo
echo "=== Hex dump ==="
xxd "$OUTPUT" | head -10
echo
echo "=== Checking for P (0x50) or T (0x54) tags ==="
if xxd "$OUTPUT" | grep -q " 50\| 54"; then
    echo "FOUND: P (0x50) or T (0x54) tags present!"
    echo "Tag bytes found at:"
    xxd "$OUTPUT" | grep " 50\| 54"
else
    echo "NOT FOUND: No P/T tags detected"
fi

rm -f "$SOCKET" "$OUTPUT"
