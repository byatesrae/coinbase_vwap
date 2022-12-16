#!/bin/bash

# Runs linting.
# Intended to be invoked from the repository root.

set -e

echo " * Linting ..."
echo

$(go env GOPATH)/bin/golangci-lint run

echo
echo " * Done."
