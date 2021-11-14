# Developing Canary

## Build

This project is developed by using [Golang](https://golang.org/).
The repository provides a `Makefile` with different make targets for building the project.

### Building the binary and running the application

In order to build the application binary, use the `go_build` make target.

```shell
make go_build
```

The binary will be available in the `cmd/target` folder named as `strimzi-canary` that you can just run locally from the root project folder.

```shell
./cmd/target/strimzi-canary
```

### Building a container image

To build the container image, use the `docker_build` make target together with the `docker_push` one if you want to push it to a registry as well.

```shell
make docker_build docker_push
```

It is possible to customize the container image build by using the following environment variables:

* `DOCKER_REGISTRY`: the Docker registry where the image will be pushed (default is `docker.io`).
* `DOCKER_ORG`: the Docker organization for tagging/pushing the image (defaults to the value of t.he $USER environment variable).
* `DOCKER_TAG`: the Docker tag (default is `latest`).
* `DOCKER_REPO`: the Docker repository where the image will be pushed (default is `canary`).