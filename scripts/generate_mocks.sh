#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# Check if mockery is installed
# If not, install it
if ! command -v mockery &> /dev/null
then
    echo "mockery not found. Please install it and make sure it is in your PATH."
    echo "Installation intructions can be found here: https://vektra.github.io/mockery/latest/installation"
    exit 1
fi

# Generate mocks
echo "Generating mocks in $PACKAGE_DIR..."


mockery --config ./.mockery.yaml
