DOCKER_REGISTRY ?= docker.io
DOCKER_ORG ?= $(USER)
DOCKER_TAG ?= latest

BINARY ?= strimzi-canary

go_build:
	echo "Building Golang binary ..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmd/target/$(BINARY) cmd/main.go

docker_build:
	echo "Building Docker image ..."
	docker build -t ${DOCKER_REGISTRY}/${DOCKER_ORG}/canary:${DOCKER_TAG} .

docker_push:
	echo "Pushing Docker image ..."
	docker push ${DOCKER_REGISTRY}/${DOCKER_ORG}/canary:${DOCKER_TAG}

all: go_build docker_build docker_push

clean:
	echo "Cleaning ..."
	rm -rf cmd/target