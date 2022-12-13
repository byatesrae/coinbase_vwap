#!/bin/bash

# Runs bash commands in the dockerised build/ci environment.
# Intended to be invoked from the repository root. For example:
#
# ./build/docker/run.sh "make env lint; echo Done!"

set -e

if [ -z "$1" ]; then 
    echo "ERR: First argument must be a bash command(s) to run. For example:
    ./build/docker/run.sh \"make env lint; echo Done!\"";
    
    exit 1; 
fi

# The user inside the docker container is root. Any files created need to have 
# permissions updated after the container exits.
reset_permissions() {
    echo " * Resetting permissions ..."
    docker run --rm -v $(pwd):/src busybox:stable chown -R $(id -u):$(id -u) src
}
trap reset_permissions ERR

echo " * Building image ..."
docker build \
    -t coinbasevwap_build \
    $( [ -n "$BUILD_IMAGE" ] && printf %s "--build-arg BUILD_IMAGE=$BUILD_IMAGE" ) \
    ./build/docker/

# Run the bash command(s).
echo " * Running \"$1\" ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --workdir="/src" \
    --entrypoint /bin/bash \
    coinbasevwap_build \
    "-c" "./build/docker/config.sh; $1" 

reset_permissions