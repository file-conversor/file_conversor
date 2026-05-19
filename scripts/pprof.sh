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

if [ "$#" -eq 0 ]; then
    echo "Usage: $0 <go run arguments>"
    echo "Example: $0 cmd/file_conversor/main.go"
    exit 1
fi

rm -f *.prof
rm -f *.trace
export FC_PROFILE=1

echo "Profiling the app ..."

echo "   1. Running the app with profiling enabled..."
go run $@ || echo "[WARNING] App failed, continuing with profiling data collection..."
echo

echo "   2. Opening pprof visualization in the browser..."
go tool pprof -http=:8080 *.prof || echo "[WARNING] Failed to open pprof visualization, please check the .prof files manually."
echo 

echo "   3. Opening trace visualization in the browser..."
go tool trace trace.trace || echo "[WARNING] Failed to open trace visualization, please check the .trace file manually."
echo 

echo "Profiling the app ... Done!"
echo 