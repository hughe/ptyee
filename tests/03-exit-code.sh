#!/bin/bash
# Test 3: Exit code propagation
# Tests that ptyee exits with the same code as the child process

echo "Test 3: Exit code propagation"

# Test with exit code 0
./build/ptyee -- bash -c 'exit 0'
if [ $? -eq 0 ]; then
    echo "PASS: Exit code 0 propagated correctly"
else
    echo "FAIL: Expected exit code 0, got $?"
    exit 1
fi

# Test with exit code 42
./build/ptyee -- bash -c 'exit 42'
if [ $? -eq 42 ]; then
    echo "PASS: Exit code 42 propagated correctly"
else
    echo "FAIL: Expected exit code 42, got $?"
    exit 1
fi

# Test with exit code 1
./build/ptyee -- bash -c 'exit 1'
if [ $? -eq 1 ]; then
    echo "PASS: Exit code 1 propagated correctly"
else
    echo "FAIL: Expected exit code 1, got $?"
    exit 1
fi

echo "Test 3: PASSED"
