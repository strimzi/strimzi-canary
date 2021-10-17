include ./Makefile.docker
include ./Makefile.binary

PROJECT_NAME ?=  canary
RELEASE_VERSION ?= latest

.PHONY: all
all: go_build docker_build docker_push

.PHONY: clean
clean: go_clean