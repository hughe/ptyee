#!/bin/bash
# Run all ptyee tests

set -e

echo "========================================="
echo "Running all ptyee tests"
echo "========================================="
echo

# Make sure we're in the project root
cd "$(dirname "$0")/.."

# Make sure ptyee is built
if [ ! -f "./build/ptyee" ]; then
    echo "Building ptyee..."
    make build
fi

# Run each test
TESTS=(
    "tests/01-basic-execution.sh"
    "tests/02-tag-output.sh"
    "tests/03-exit-code.sh"
    "tests/04-single-client.sh"
    "tests/05-interactive-python.sh"
    "tests/06-wait-for-connection.sh"
    "tests/07-signal-handling.sh"
)

PASSED=0
FAILED=0

for test in "${TESTS[@]}"; do
    echo
    echo "========================================="
    if bash "$test"; then
        ((PASSED++))
    else
        ((FAILED++))
        echo "FAILED: $test"
    fi
    echo "========================================="
done

echo
echo "========================================="
echo "Test Summary:"
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "========================================="

if [ $FAILED -gt 0 ]; then
    exit 1
fi

echo
echo "All tests passed!"
