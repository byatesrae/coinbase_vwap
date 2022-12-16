#!/bin/bash

# Installs/upgrades all dependencies.
# Intended to be invoked from the repository root.

set -e

echo " * Upgrading/Installing dependencies ..."
echo

go get -u -t -d -v ./...
go mod tidy
go mod vendor

echo
echo " * Done."
