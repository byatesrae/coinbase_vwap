#!/bin/bash

# Configures the build/ci environment.
# This should only be invoked from run.sh.

set -e

echo " * Configuring env ..."
echo

go env -w GOPRIVATE=github.com/byatesrae/*

echo
echo " * Done."