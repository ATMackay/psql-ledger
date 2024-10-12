#!/usr/bin/env bash

set -e

# Version control
version=$(git describe --tags)
build_date=$(date -u +'%Y-%m-%d %H:%M:%S')
commit_hash=$(git rev-parse HEAD)
commit_hash_short=$(git rev-parse --short HEAD)
commit_date=$(git show -s --format=%ci "$commit_hash")


# Set ARCH variable
architecture=$(uname -m)
if [ -z "$architecture" ]; then
    export ARCH="x86_64" # Linux, Windows (default)
else
    export ARCH=${architecture} # Mac
fi

echo "Executing docker build"
docker build \
       --build-arg ARCH="$ARCH" \
       --build-arg BUILD_DATE="$build_date"\
       --build-arg SERVICE=psqlledger \
       --build-arg VERSION="$version" \
       --build-arg COMMIT_HASH="$commit_hash" \
       --build-arg COMMIT_DATE="$commit_date" \
       -t "${ECR}"psqlledger:latest  \
       -t "${ECR}"psqlledger:"$commit_hash_short"  \
       -f Dockerfile ..