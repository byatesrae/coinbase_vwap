#!/bin/bash

# Run all tests.
# Intended to be invoked from the repository root.

set -e

echo " * Running tests ..."
echo

go test -race ./...

echo
echo " * Done."
