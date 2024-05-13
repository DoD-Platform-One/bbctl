#! /usr/bin/env bash

# Exit on error
set -e

# Get dirs
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
. "${DIR}/get_dirs.sh"

# ensure devpod is installed
if ! command -v devpod &> /dev/null
then
    echo "devpod could not be found, please install it" >&2
    echo "https://devpod.sh/docs/getting-started/install#optional-install-devpod-cli" >&2
    exit 1
fi

# Get arg
arg=$1

# Run devpod
if [ "$arg" == "up" ]; then
    echo "Running devpod in $PACKAGE_DIR..."
    devpod up "$PACKAGE_DIR"
elif [ "$arg" == "stop" ]; then
    echo "Stopping devpod in $PACKAGE_DIR..."
    devpod stop "$PACKAGE_DIR"
elif [ "$arg" == "build" ]; then
    echo "Building devpod in $PACKAGE_DIR..."
    devpod up "$PACKAGE_DIR" --recreate
else
    echo "Invalid argument. Usage: devpod.sh [up|stop|build]" >&2
    exit 1
fi
