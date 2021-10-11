[![Build Status](https://dev.azure.com/cncf/strimzi/_apis/build/status/canary?branchName=main)](https://dev.azure.com/cncf/strimzi/_build/latest?definitionId=42&branchName=main)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Twitter Follow](https://img.shields.io/twitter/follow/strimziio.svg?style=social&label=Follow&style=for-the-badge)](https://twitter.com/strimziio)

# Strimzi canary

This repository contains the Strimzi canary tool implementation.
It acts as an indicator of whether Kafka clusters are operating correctly.
This is achieved by creating a canary topic and periodically producing and consuming events on the topic and getting metrics out of these exchanges.

## Configuration

When running the Strimzi canary tool, it is possible to configure different aspects by using the following environment variables.

* `KAFKA_BOOTSTRAP_SERVERS`: comma separated bootstrap servers of the Kafka cluster to connect to. Default `localhost:9092`.
* `KAFKA_BOOTSTRAP_BACKOFF_MAX_ATTEMPTS`: maximum numeber of attempts for connecting to the Kafka cluster if it is not ready yet. Defualt `10`.
* `KAFKA_BOOTSTRAP_BACKOFF_SCALE`: the scale used to delay between attempts to connect to the Kafka cluster (in ms). Default `5000`.
* `TOPIC`: the name of the topic used by the tool to send and receive messages. Default `__strimzi_canary`.
* `TOPIC_CONFIG`: topic configuration defined as a list of semicolon separated `key=value` pairs (i.e. `retention.ms=600000;segment.bytes=16384`). Default empty.
* `RECONCILE_INTERVAL_MS`: it defines how often the tool has to send and receive messages (in ms). Default `30000`.
* `CLIENT_ID`: the client id used for configuring producer and consumer. Default `strimzi-canary-client`.
* `CONSUMER_GROUP_ID`: group id for the consumer group joined by the canary consumer. Default `strimzi-canary-group`.
* `PRODUCER_LATENCY_BUCKETS`: buckets of the hystogram related to the producer latency metric (in ms). Default `100,200,400,800,1600`.
* `ENDTOEND_LATENCY_BUCKETS`: buckets of the hystogram related to the end to end latency metric between producer and consumer (in ms). Default `100,200,400,800,1600`.
* `EXPECTED_CLUSTER_SIZE`: expected number of brokers in the Kafka cluster where the canary connects to. This parameter avoid that the tool runs more partitions reassignment of the topic while the Kafka cluster is starting up and the brokers are coming one by one. Default `-1` means "dynamic" reassignment as described above. When greater than 0, the canary waits for the Kafka cluster having the expected number of brokers running before creating the topic and assigning the partitions.
* `KAFKA_VERSION`: version of the Kafka cluster. Default `2.8.0`.
* `SARAMA_LOG_ENABLED`: enables the Sarama client logging. Default `false`.
* `VERBOSITY_LOG_LEVEL`: verbosity of the tool logging. Default `0`. Allowed values 0 = INFO, 1 = DEBUG, 2 = TRACE.
* `TLS_ENABLED`: if the canary has to use TLS to connect to the Kafka cluster. Default `false`.
* `TLS_CA_CERT`: TLS CA certificate, in PEM format, to use to connect to the Kafka cluster. When this parameter is empty (default behaviour) and the TLS connection is enabled, the canary uses the system certificates trust store. When a TLS CA certificate is specified, it is added to the system certificates trust store. 
* `TLS_CLIENT_CERT`: TLS client certificate, in PEM format, to use for enabling TLS client authentication against the Kafka cluster. Default empty.
* `TLS_CLIENT_KEY`: TLS client private key, in PEM format, to use for enabling TLS client authentication against the Kafka cluster. Default empty.
* `TLS_INSECURE_SKIP_VERIFY`:  if the underneath Sarama client has to verify the server's certificate chain and host name. Default `false`.
* `SASL_MECHANISM`: mechanism to use for SASL authentication against the Kafka cluster. Default empty.
* `SASL_USER`: username for SASL authentication against the Kafka cluster when PLAIN or SCRAM-SHA are used. Default empty.
* `SASL_PASSWORD`: password for SASL authentication against the Kafka cluster when PLAIN or SCRAM-SHA are used. Default empty.
* `CONNECTION_CHECK_INTERVAL_MS`: it defines how often the tool has to check the connection with brokers (in ms). Default `120000`.
* `CONNECTION_CHECK_LATENCY_BUCKETS`: buckets of the hystogram related to the brokers connection latency metric (in ms). Default `100,200,400,800,1600`. 

## Build

The canary repository provides a `Makefile` with different make targets for building the project.

### Make targets

The available make targets are the following:

* `go_build`: for building the application binary in the `cmd/target` folder.
* `docker_build`: for building the Docker image.
* `docker_save`: for saving the Docker image as a tar.gz file.
* `docker_push`: for pushing the Docker image to a Docker repository.
* `docker_load`: for loading the Docker image from a tar.gz file into the local registry.
* `all`: for running all the previous ones to build the application and then build and push the Docker image.
* `clean`: to clean the folder where the application binary is built.

### Docker image customization

The make targets used for building and pushing the Docker image use the following environment variables for allowing some degree of customization:

* `DOCKER_REGISTRY`: the Docker registry where the image will be pushed (default is `docker.io`).
* `DOCKER_ORG`: the Docker organization for tagging/pushing the image (defaults to the value of t.he $USER environment variable).
* `DOCKER_TAG`: the Docker tag (default is `latest`).
* `DOCKER_REPO`: the Docker repository where the image will be pushed (default is `canary`).