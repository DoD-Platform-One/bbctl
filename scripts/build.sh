#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Set custom date
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo
echo "Build Time: \"$BUILD_DATE\""

# Build
echo "building..."
REPO_DIR="repo1.dso.mil/big-bang/product/packages/bbctl"
go build -o "$PACKAGE_DIR/bin/$PACKAGE_NAME" -ldflags "-X $REPO_DIR/static.buildDate=$BUILD_DATE"
