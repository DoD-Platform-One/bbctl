#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Run tests
echo "Running tests in $PACKAGE_DIR..."
go test -v -coverprofile=cover.txt ./...
go tool cover -html=cover.txt
