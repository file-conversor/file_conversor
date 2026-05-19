#!/bin/bash
# scripts/build.sh
#  - A script to build the app using Go.

# Exit immediately if a command exits with a non-zero status, treat unset variables as an error, and prevent errors in a pipeline from being masked.
set -Eeuo pipefail

# Trap errors and print the command that failed, its exit code, and the line number before exiting with the same code.
trap 'rc=$?;
echo "ERROR: command \"${BASH_COMMAND}\" failed with exit code $rc at line ${LINENO}";
exit $rc' ERR

. "$(dirname "$0")/_env.sh"

if [ -f "before.log" ] && [ -f "after.log" ]; then
    echo "Comparing benchmark results in before.log and after.log ..."
    benchstat before.log after.log
    exit 0
fi

echo "Benchmarking ..."

echo "   1. Running go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ..."
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...

echo "   2. Installing benchstat if not already installed ..."
if !which benchstat > /dev/null 2>&1; then
    go install golang.org/x/perf/cmd/benchstat@latest
fi

if ! [ -f "before.log" ]; then
    echo "   3. Saving benchmark results to before.log ..."
    benchstat before.log
elif ! [ -f "after.log" ]; then
    echo "   3. Saving benchmark results to after.log ..."
    benchstat after.log
fi

echo "Benchmarking ... Done!"
echo