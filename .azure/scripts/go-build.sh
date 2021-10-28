#!/usr/bin/env bash
set -e

echo "Build reason: ${BUILD_REASON}"
echo "Source branch: ${BRANCH}"

export RELEASE_VERSION=$(cat ./release.version)
echo "Building for ${RELEASE_VERSION}"

make go_build