#!/bin/bash

# Installs dependencies.
# Intended to be invoked from the repository root.

set -e

echo " * Installing dependencies ..."
echo

go mod tidy
go mod vendor

echo
echo " * Done."
