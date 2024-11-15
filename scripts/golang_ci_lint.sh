#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# TODO: Remove this installation once we have golangci-lint pre-installed in the pipeline image
GOBIN=$(pwd)/bin/ go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

# golangci Lint
echo "linting in $PACKAGE_DIR..."

mv .git .git-hidden
trap 'mv .git-hidden .git' EXIT

GOGC=30 bin/golangci-lint run ./... --timeout=30m

echo "No linting errors detected"