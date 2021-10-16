DOCKER_REGISTRY ?= docker.io
DOCKER_ORG ?= $(USER)
DOCKER_TAG ?= latest
DOCKER_REPO ?= canary

BINARY ?= strimzi-canary

RELEASE_VERSION ?= latest

go_build:
	echo "Building Golang binary ..."
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.version=${RELEASE_VERSION}'" -a -installsuffix cgo -o cmd/target/$(BINARY) cmd/main.go

docker_build:
	echo "Building Docker image ..."
	docker build -t ${DOCKER_REGISTRY}/${DOCKER_ORG}/${DOCKER_REPO}:${DOCKER_TAG} .

docker_save:
	echo "Saving Docker image as tar.gz file ..."
	docker save ${DOCKER_REGISTRY}/${DOCKER_ORG}/${DOCKER_REPO}:${DOCKER_TAG} | gzip > canary-container.tar.gz

docker_push:
	echo "Pushing Docker image ..."
	docker push ${DOCKER_REGISTRY}/${DOCKER_ORG}/${DOCKER_REPO}:${DOCKER_TAG}

docker_load:
	echo "Loading Docker image from tar.gz file ..."
	docker load < canary-container.tar.gz 

all: go_build docker_build docker_push

clean:
	echo "Cleaning ..."
	rm -rf cmd/target