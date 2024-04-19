#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Build
echo "Building in $PACKAGE_DIR..."
go build -o "$PACKAGE_DIR/bin/$PACKAGE_NAME"
