#!/bin/bash

# Generates all generated code.
# Intended to be invoked from the repository root.

set -e

echo " * Generating code ..."
echo

go generate ./...

echo
echo " * Done."
