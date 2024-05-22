#!/bin/bash

# Help message
help_message() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -s, --shell SHELL   Specify the shell environment to use: 'fish', 'zsh', or 'bash'"
    echo "  -h, --help          Display this help message"
    exit 0
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -s|--shell)
            SHELL=$2
            shift 2
            ;;
        -h|--help)
            help_message
            ;;
        *)
            echo "Error: Unknown option $1"
            help_message
            ;;
    esac
done

# Check if shell is specified
if [ -z "$SHELL" ]; then
    echo "Error: Please specify the shell environment using the '-s' option."
    help_message
fi

# Check if shell is valid
case "$SHELL" in
    bash|zsh|fish)
        ;;
    *)
        echo "Error: Unsupported shell environment. Please choose 'bash', 'zsh', or 'fish'."
        help_message
        ;;
esac

# Cross-compile Go application for Linux
GOOS=linux GOARCH=amd64 go build -o bbctl_linux ../..

# Build Docker image based on selected shell
DOCKERFILE="${SHELL}.Dockerfile"
docker build -t my_${SHELL}_container -f $DOCKERFILE .

# Clean up build artifact
rm bbctl_linux

KUBECONFIG_PATH=$(bbctl config kubeconfig 2>/dev/null || echo $KUBECONFIG)

# KUBECONFIG_PATH=$(bbctl config kubeconfig)

# Run container
docker run -it --rm \
    -v $KUBECONFIG_PATH:/root/.kube/config \
     my_${SHELL}_container 

