# This script is intended to be sourced by other scripts to get the directories

# Get current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Get root directory
ROOT_DIR="$(dirname "$DIR")"

# Get the package name
PACKAGE_NAME="$(basename "$ROOT_DIR")"

# Get package directory
PACKAGE_DIR="$ROOT_DIR"
mkdir -p "$PACKAGE_DIR/bin"
cd "$PACKAGE_DIR"
