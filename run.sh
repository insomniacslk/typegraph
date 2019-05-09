#!/bin/bash
set -eu

# This script automates calling typegraph, converting the DOT file to PNG, and
# opening it with xdg-open (Linux-only).

SRC=$@
if [ -z "${SRC}" ]
then
    echo "missing arguments for typegraph"
    exit 1
fi

go run github.com/insomniacslk/typegraph $SRC > types.dot
dot -Tpng types.dot > types.png
xdg-open types.png
