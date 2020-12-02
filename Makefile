IMAGE ?= strimzi-canary
VERSION ?= latest
TAG ?= latest


.PHONY: help dobuild build tag push all

help:
	@echo "Makefile arguments:"
	@echo "REGISTRY: Image registry"
	@echo "ORG: Image organisation"
	@echo "Makefile commands:"
	@echo "build"
	@echo "tag"
	@echo "push"
	@echo "all"

.DEFAULT_GOAL := all

gobuild:
	go build -o target/main ./cmd

build:
	@docker	build --pull --build-arg version=$(VERSION) -t ${IMAGE}:${VERSION} .

tag:
	@echo "Tagging  image ${IMAGE}:${VERSION} ${REGISTRY}/${ORG}/${IMAGE}:${TAG}"
	@docker tag ${IMAGE}:${VERSION} ${REGISTRY}/${ORG}/${IMAGE}:${TAG}

push:
	@docker push ${REGISTRY}/${ORG}/${IMAGE}:${TAG}

all: build tag push