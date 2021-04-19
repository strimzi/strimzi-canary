IMAGE_REGISTRY ?= quay.io
IMAGE_ORG ?= strimzi
IMAGE_REPO ?= strimzi-canary
IMAGE_TAG ?= 0.0.1

BINARY ?= strimzi-canary

go_build:
	echo "Building Golang binary ..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmd/target/$(BINARY) cmd/main.go

docker_build:
	echo "Building Docker image ..."
	docker build -t ${IMAGE_REGISTRY}/${IMAGE_ORG}/${IMAGE_REPO}:${IMAGE_TAG} .

docker_push:
	echo "Pushing Docker image ..."
	docker push ${IMAGE_REGISTRY}/${IMAGE_ORG}/${IMAGE_REPO}:${IMAGE_TAG}

all: go_build docker_build docker_push

clean:
	echo "Cleaning ..."
	rm -rf cmd/target