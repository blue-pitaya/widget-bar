#!/usr/bin/bash

set -e

source .env

if [ ! -n "$INSTALL_PATH" ]; then
    echo "Install path not defined."
    exit 1
fi

go build .
cp widget-bar "$INSTALL_PATH/widget-bar"
