#!/usr/bin/env bash
set -e

echo "Build reason: ${BUILD_REASON}"
echo "Source branch: ${BRANCH}"

RELEASE_VERSION := $(shell cat ../../release.version)
echo "Build binary for ${RELEASE_VERSION}"

make go_build