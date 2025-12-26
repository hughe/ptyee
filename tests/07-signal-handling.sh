#!/bin/bash
# Test 7: Signal handling (Ctrl-C)
# Tests that ptyee cleanly terminates when receiving SIGINT

set -e

SOCKET="/tmp/ptyee-test-signal.sock"

echo "Test 7: Signal handling (Ctrl-C)"

# Cleanup from previous runs
rm -f "$SOCKET"

# Start ptyee with a long-running command (sleep 100)
# Use --no-wait so we don't need to connect to the socket first
# Redirect stdin from /dev/null to avoid blocking
./build/ptyee --socket "$SOCKET" --no-wait -- sleep 100 < /dev/null 2>/dev/null &
PTYEE_PID=$!

# Give it a moment to start
sleep 0.5

# Verify process is running
if ! ps -p $PTYEE_PID > /dev/null 2>&1; then
    echo "FAIL: Process not running after startup"
    exit 1
fi

echo "PASS: Process started successfully"

# Send SIGINT (simulating Ctrl-C)
kill -INT $PTYEE_PID

# Wait for process to terminate (with timeout)
TERMINATED=0
for i in {1..20}; do
    if ! ps -p $PTYEE_PID > /dev/null 2>&1; then
        TERMINATED=1
        break
    fi
    sleep 0.1
done

if [ $TERMINATED -eq 0 ]; then
    echo "FAIL: Process did not terminate after SIGINT"
    kill -9 $PTYEE_PID 2>/dev/null || true
    rm -f "$SOCKET"
    exit 1
fi

echo "PASS: Process terminated cleanly after SIGINT"

# Verify child process (sleep) is also terminated
# We can't easily get the PID of the sleep command, but if ptyee terminated,
# it should have killed its child. We'll verify by checking if any sleep 100
# processes are still running from our user.
sleep 0.2
SLEEP_COUNT=$(ps -u $(whoami) | grep -c "sleep 100" || true)
if [ $SLEEP_COUNT -gt 0 ]; then
    echo "WARNING: Child process (sleep 100) may still be running"
    # This is not a hard fail as it could be from another process
fi

# Cleanup
rm -f "$SOCKET"

echo "Test 7: PASSED"
