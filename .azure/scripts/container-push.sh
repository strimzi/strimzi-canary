#!/usr/bin/env bash
set -e

echo "Login into container registry ..."
docker login -u $DOCKER_USER -p $DOCKER_PASS $DOCKER_REGISTRY

make docker_load docker_push