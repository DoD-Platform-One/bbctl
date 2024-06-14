#! /usr/bin/env bash

if [ "$1" == "username" ]; then
    echo "username"
    exit 0
elif [ "$1" == "password" ]; then
    # no-op
    exit 0
else
    echo "Invalid credentials field"
    exit 1
fi
