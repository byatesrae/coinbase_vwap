#!/bin/bash

# Run all tests.
# Intended to be invoked from the repository root.

set -e

echo " * Running tests ..."
echo

go test -race -count=1 ./...

echo
echo " * Done."
