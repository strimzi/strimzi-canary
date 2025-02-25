[![Build Status](https://dev.azure.com/cncf/strimzi/_apis/build/status/canary?branchName=main)](https://dev.azure.com/cncf/strimzi/_build/latest?definitionId=42&branchName=main)
[![GitHub release](https://img.shields.io/github/release/strimzi/strimzi-canary.svg)](https://github.com/strimzi/strimzi-canary/releases/latest)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
[![Twitter Follow](https://img.shields.io/twitter/follow/strimziio?style=social)](https://twitter.com/strimziio)

# Strimzi canary

Based on the proposal [86 - Archive Canary project](https://github.com/strimzi/proposals/blob/main/086-archive-canary.md) this repository is archived.

## Summary
This repository contains the Strimzi canary tool implementation.
The canary tool acts as an indicator of whether [Apache Kafka®](https://kafka.apache.org) clusters are operating correctly.
This is achieved by creating a canary topic and periodically producing and consuming events on the topic and getting metrics out of these exchanges.

## Deployment

Deploy the Strimzi canary tool to the Kubernetes cluster where your Apache Kafka cluster is running. Download and unzip the installation files from [Releases](https://github.com/strimzi/strimzi-canary/releases), then edit the `KAFKA_BOOTSTRAP_SERVERS` environment variable in the `Deployment` file to specify the bootstrap servers for connecting to the Apache Kafka cluster.

The `Deployment` file also has the following configuration by default:

- `RECONCILE_INTERVAL_MS` is set at`"10000"` millseconds, which means that the canary tool produces and consumes messages every 10 seconds
- `TLS_ENABLED` is set as `false` so that TLS is not enabled

To deploy the canary tool to your Kubernetes cluster, run the following command:

```shell
kubectl apply -f ./install
```

Other than creating the corresponding `Deployment`, the canary will run with a specific `ServiceAccount`. A `Service` is created to make the Prometheus metrics accessible through HTTP on port `8080`.

### Encryption and TLS

If your Apache Kafka cluster has TLS enabled to encrypt traffic on the listener canary will use to connect, enable TLS for the canary tool as well.

Set `TLS_ENABLED` to `true` in the `Deployment` file.
You'll also need a cluster CA certificate in PEM format that validates the identity of the Kafka brokers. 
You can reference a cluster CA certificate using the  `TLS_CA_CERT` environment variable. 
If you use the cluster CA certificate generated by the Cluster Operator, extract it from the `<cluster_name>-cluster-ca-cert` `Secret`.
If you leave `TLS_CA_CERT` empty, canary will use the system certificates already installed (i.e. Verisign, Let's Encrypt, ...).

### Authentication and authorization

If the Apache Kafka cluster has TLS mutual (client) authentication enabled, the canary has to be configured with a client certificate and private key in PEM format. Use the corresponding environment variables `TLS_CLIENT_CERT` and `TLS_CLIENT_KEY`.
If you're using the Strimzi User Operator, the values for these environment variables are provided by the `Secret` for the `KafkaUser` configured with TLS authentication.

If the Apache Kafka cluster has authentication enabled with the `PLAIN`, `SCRAM-SHA-256`, or `SCRAM-SHA-512` SASL mechanism, the canary must be configured to use it as well.
The SASL mechanism is specified using the `SASL_MECHANISM` environment variable. The username and password are specified using the `SASL_USER` and `SASL_PASSWORD` environment variables.
If you're using the Strimzi User Operator, the values for these environment variables are provided by the corresponding `Secret` for the `KafkaUser` configured to use one of the SASL authentication mechanisms.

## Configuration

When running the Strimzi canary tool, it is possible to configure different aspects by using the environment variables listed in the following table.
In addition, certain aspects can be overridden dynamically at runtime from a JSON configuration file.
Where this is possible, a field name is provided in the table.
The configuration file described in more detail the next section.

| Environment variable | Description | Default | Dynamic Configuration field name |
|---|---|---|---|
| `KAFKA_BOOTSTRAP_SERVERS` | Comma separated bootstrap servers of the Kafka cluster to connect to. | `localhost:9092` |  |
| `KAFKA_BOOTSTRAP_BACKOFF_MAX_ATTEMPTS` | Maximum number of attempts for connecting to the Kafka cluster if it is not ready yet. | `10` |  |
| `KAFKA_BOOTSTRAP_BACKOFF_SCALE` | The scale used to delay between attempts to connect to the Kafka cluster (in ms) | `5000` |  |
| `TOPIC` | The name of the topic used by the tool to send and receive messages. | `__strimzi_canary` |  |
| `TOPIC_CONFIG` | Topic configuration defined as a list of semicolon separated `key=value` pairs (i.e. `retention.ms=600000;segment.bytes=16384`). | empty |  |
| `RECONCILE_INTERVAL_MS` | It defines how often the tool has to send and receive messages (in ms). | `30000` |  |
| `CLIENT_ID` | The client id used for configuring producer and consumer. | `strimzi-canary-client` |  |
| `CONSUMER_GROUP_ID` | Group id for the consumer group joined by the canary consumer. | `strimzi-canary-group` |  |
| `PRODUCER_LATENCY_BUCKETS` | Buckets of the histogram related to the producer latency metric (in ms). | `2,5,10,20,50,100,200,400` |  |
| `ENDTOEND_LATENCY_BUCKETS` | Buckets of the histogram related to the end to end latency metric between producer and consumer (in ms). | `5,10,20,50,100,200,400,800` |  |
| `EXPECTED_CLUSTER_SIZE` | Expected number of brokers in the Kafka cluster where the canary connects to. This parameter avoids that the tool runs more partitions reassignment of the topic while the Kafka cluster is starting up and the brokers are coming one by one. `-1` means "dynamic" reassignment as described above. When greater than 0, the canary waits for the Kafka cluster having the expected number of brokers running before creating the topic and assigning the partitions | `-1` |  |
| `KAFKA_VERSION` | Version of the Kafka cluster | `3.2.0` |  |
| `SARAMA_LOG_ENABLED` | Enables the Sarama client logging. | `false` | `saramaLogEnabled` |
| `VERBOSITY_LOG_LEVEL` | Verbosity of the tool logging. Allowed values 0 = INFO, 1 = DEBUG, 2 = TRACE | `0` | `verbosityLogLevel` |
| `TLS_ENABLED` | If the canary has to use TLS to connect to the Kafka cluster. | `false` |  |
| `TLS_CA_CERT` | TLS CA certificate, in PEM format, to use to connect to the Kafka cluster. When this parameter is empty (default behaviour) and the TLS connection is enabled, the canary uses the system certificates trust store. When a TLS CA certificate is specified, it is added to the system certificates trust store | empty |  |
| `TLS_CLIENT_CERT` | TLS client certificate, in PEM format, to use for enabling TLS client authentication against the Kafka cluster. | empty |  |
| `TLS_CLIENT_KEY` | TLS client private key, in PEM format, to use for enabling TLS client authentication against the Kafka cluster. | empty |  |
| `TLS_INSECURE_SKIP_VERIFY` | if the underneath Sarama client has to verify the server's certificate chain and host name. | `false` |  |
| `SASL_MECHANISM` | Mechanism to use for SASL authentication against the Kafka cluster. Supported are `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-512`. | empty |  |
| `SASL_USER` | Username for SASL authentication against the Kafka cluster when one of `PLAIN`, `SCRAM-SHA-256` or `SCRAM-SHA-512` is used. | empty |  |
| `SASL_PASSWORD` | Password for SASL authentication against the Kafka cluster when one of `PLAIN`, `SCRAM-SHA-256` or `SCRAM-SHA-512` is used. | empty |  |
| `CONNECTION_CHECK_INTERVAL_MS` | It defines how often the tool has to check the connection with brokers (in ms). | `120000` |  |
| `CONNECTION_CHECK_LATENCY_BUCKETS` | Buckets of the histogram related to the broker's connection latency metric (in ms). | `100,200,400,800,1600` |  |
| `STATUS_CHECK_INTERVAL_MS` | It defines how often (in ms) the tool updates internal status information (i.e. percentage of consumed messages) to expose outside on the corresponding HTTP endpoint. | `30000` |  |
| `STATUS_TIME_WINDOW_MS` | It defines the sliding time window size (in ms) in which status information are sampled. | `300000` |  |
| `DYNAMIC_CONFIG_FILE` | Location of an optional external config file that provides configuration at runtime. | empty |  |
| `DYNAMIC_CONFIG_WATCHER_INTERVAL` | Interval that dynamic config file is examined for changes in content (in ms)  | `30000` |  |
| `EXPORTER_TYPE_TRACING` | Tracing Exporter use. Empty value disable tracing, other possible values are `jaeger` or `otlp`  | `` |  |
| `PROMETHEUS_CONSTANT_LABELS` | A list of semicolon separated `key=value` pairs (i.e. `my-label-1=one;my-label-2=two`) that will be set as constant labels on all prometheus metrics. | empty |  |

## Dynamic Configuration file

As mentioned above certain aspects of behaviour can be overridden dynamically at runtime from a JSON configuration file.
If a config file reference is provided by `DYNAMIC_CONFIG_FILE`, that file will be monitored for changes in content and creation/deletion, with any changes being applied dynamically to the Canary's runtime state.
Configuration values by the config file take precedence over configuration values provided by equivalent environment variable.

```json
{
  "saramaLogEnabled": true,
  "verbosityLogLevel": 1
}
```

In a kubernetes environment this file could be provided by a projected configmap.

## Endpoints

The canary exposes some HTTP endpoints, on port 8080, to provide information about status, health and metrics.

### Liveness and readiness

The `/liveness` and `/readiness` endpoints report back if the canary is live and ready by proving just an `OK` HTTP body.

### Metrics

The `/metrics` endpoint provides useful metrics in Prometheus format.

### Status

The `/status` endpoint provides status information through a JSON object structured with different sections.

The `Consuming` field provides information about the `Percentage` of messages correctly consumed in a sliding `TimeWindow` (in ms), whose maximum size is configured via the `STATUS_TIME_WINDOW_MS` environment variable; until that size is reached, the `TimeWindow` field reports the current covered time window with gathered samples.

```json
{
  "Consuming": {
    "TimeWindow": 150000,
    "Percentage": 100
  }
}
```

If the time window has not ended, the `/status` endpoint cannot report a percentage of correctly consumed messages. Instead, it returns `Percentage: -1`. The canary also logs `Error processing consumed records percentage: No data samples available in the time window ring`.  In this case, you wait until the time window has ended for the sampling to complete. 

## Metrics

In order to check how your Apache Kafka cluster is behaving, the Canary provides the following metrics on the corresponding HTTP endpoint.

| Name                                    | Description                                                                       |
|-----------------------------------------|-----------------------------------------------------------------------------------|
| `client_creation_error_total`           | Total number of errors while creating Sarama client                               |
| `expected_cluster_size_error_total`     | Total number of errors while waiting the Kafka cluster having the expected size   |
| `topic_creation_failed_total`           | Total number of errors while creating the canary topic                            |
| `topic_describe_cluster_error_total`    | Total number of errors while describing cluster                                   |
| `topic_describe_error_total`            | Total number of errors while getting canary topic metadata                        |
| `topic_alter_assignments_error_total`   | Total number of errors while altering partitions assignments for the canary topic |
| `topic_alter_configuration_error_total` | Total number of errors while altering configuration for the canary topic          |
| `records_produced_total`                | The total number of records produced                                              |
| `records_produced_failed_total`         | The total number of records failed to produce                                     |
| `producer_refresh_metadata_error_total` | Total number of errors while refreshing producer metadata                         |
| `records_produced_latency`              | Records produced latency in milliseconds                                          |
| `records_consumed_total`                | The total number of records consumed                                              |
| `consumer_error_total`                  | Total number of errors reported by the consumer                                   |
| `consumer_timeout_join_group_total`     | The total number of consumers not joining the group within the timeout            |
| `consumer_refresh_metadata_error_total` | Total number of errors while refreshing consumer metadata                         |
| `records_consumed_latency`              | Records end-to-end latency in milliseconds                                        |
| `connection_error_total`                | (Deprecated - use `connection_total`) Total number of errors while checking the connection to Kafka brokers|
| `connection_total`                      | Total number of (failed or successful) connections to Kafka brokers               |
| `connection_latency`                    | Latency in milliseconds for established or failed connections                     |

Following an example of metrics output.

```shell
# HELP strimzi_canary_records_produced_total The total number of records produced
# TYPE strimzi_canary_records_produced_total counter
strimzi_canary_records_produced_total{clientid="strimzi-canary-client",partition="0"} 1
strimzi_canary_records_produced_total{clientid="strimzi-canary-client",partition="1"} 1
strimzi_canary_records_produced_total{clientid="strimzi-canary-client",partition="2"} 1

# HELP strimzi_canary_records_consumed_total The total number of records consumed
# TYPE strimzi_canary_records_consumed_total counter
strimzi_canary_records_consumed_total{clientid="strimzi-canary-client",partition="0"} 1
strimzi_canary_records_consumed_total{clientid="strimzi-canary-client",partition="1"} 1
strimzi_canary_records_consumed_total{clientid="strimzi-canary-client",partition="2"} 1

# HELP strimzi_canary_records_produced_latency Records produced latency in milliseconds
# TYPE strimzi_canary_records_produced_latency histogram
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="0",le="50"} 0
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="0",le="100"} 0
...
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="0",le="+Inf"} 1
strimzi_canary_records_produced_latency_sum{clientid="strimzi-canary-client",partition="0"} 151
strimzi_canary_records_produced_latency_count{clientid="strimzi-canary-client",partition="0"} 1
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="1",le="50"} 0
...
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="1",le="+Inf"} 1
strimzi_canary_records_produced_latency_sum{clientid="strimzi-canary-client",partition="1"} 125
strimzi_canary_records_produced_latency_count{clientid="strimzi-canary-client",partition="1"} 1
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="2",le="50"} 0
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="2",le="100"} 0
...
strimzi_canary_records_produced_latency_bucket{clientid="strimzi-canary-client",partition="2",le="+Inf"} 1
strimzi_canary_records_produced_latency_sum{clientid="strimzi-canary-client",partition="2"} 263
strimzi_canary_records_produced_latency_count{clientid="strimzi-canary-client",partition="2"} 1

# HELP strimzi_canary_records_consumed_latency Records end-to-end latency in milliseconds
# TYPE strimzi_canary_records_consumed_latency histogram
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="0",le="100"} 0
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="0",le="200"} 1
...
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="0",le="+Inf"} 1
strimzi_canary_records_consumed_latency_sum{clientid="strimzi-canary-client",partition="0"} 161
strimzi_canary_records_consumed_latency_count{clientid="strimzi-canary-client",partition="0"} 1
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="1",le="100"} 0
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="1",le="200"} 1
...
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="1",le="+Inf"} 1
strimzi_canary_records_consumed_latency_sum{clientid="strimzi-canary-client",partition="1"} 133
strimzi_canary_records_consumed_latency_count{clientid="strimzi-canary-client",partition="1"} 1
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="2",le="100"} 0
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="2",le="200"} 0
...
strimzi_canary_records_consumed_latency_bucket{clientid="strimzi-canary-client",partition="2",le="+Inf"} 1
strimzi_canary_records_consumed_latency_sum{clientid="strimzi-canary-client",partition="2"} 266
strimzi_canary_records_consumed_latency_count{clientid="strimzi-canary-client",partition="2"} 1

# HELP strimzi_canary_connection_latency Latency in milliseconds for established or failed connections
# TYPE strimzi_canary_connection_latency histogram
strimzi_canary_connection_latency_bucket{brokerid="0",connected="true",le="100"} 1
strimzi_canary_connection_latency_bucket{brokerid="0",connected="true",le="200"} 1
...
strimzi_canary_connection_latency_bucket{brokerid="0",connected="true",le="+Inf"} 1
strimzi_canary_connection_latency_sum{brokerid="0",connected="true"} 23
strimzi_canary_connection_latency_count{brokerid="0",connected="true"} 1
strimzi_canary_connection_latency_bucket{brokerid="1",connected="true",le="100"} 1
strimzi_canary_connection_latency_bucket{brokerid="1",connected="true",le="200"} 1
...
strimzi_canary_connection_latency_bucket{brokerid="1",connected="true",le="+Inf"} 1
strimzi_canary_connection_latency_sum{brokerid="1",connected="true"} 8
strimzi_canary_connection_latency_count{brokerid="1",connected="true"} 1
strimzi_canary_connection_latency_bucket{brokerid="2",connected="true",le="100"} 1
strimzi_canary_connection_latency_bucket{brokerid="2",connected="true",le="200"} 1
...
strimzi_canary_connection_latency_bucket{brokerid="2",connected="true",le="+Inf"} 1
strimzi_canary_connection_latency_sum{brokerid="2",connected="true"} 6
strimzi_canary_connection_latency_count{brokerid="2",connected="true"} 1

# HELP strimzi_canary_client_creation_error_total Total number of errors while creating Sarama client
# TYPE strimzi_canary_client_creation_error_total counter
strimzi_canary_client_creation_error_total 4
# HELP strimzi_canary_connection_error_total Total number of errors while checking the connection to Kafka brokers
# TYPE strimzi_canary_connection_error_total counter
strimzi_canary_connection_error_total{brokerid="1",connected="false"} 1
strimzi_canary_connection_error_total{brokerid="2",connected="false"} 1
HELP strimzi_canary_connection_total Total number of connections to Kafka brokers (connection may have succeeded or failed)
# TYPE strimzi_canary_connection_total counter
strimzi_canary_connection_total{brokerid="0",connected="true"} 2
strimzi_canary_connection_total{brokerid="0",connected="false"} 3
```

### Using Prometheus and Grafana

You can use Prometheus to visualize the above metrics on the example Grafana dashboard. The PodMonitor resource file and the example Grafana dashboard file are available in the [metrics example directory](https://github.com/strimzi/strimzi-canary/tree/main/packaging/examples/metrics).

If you have not enabled Prometheus and Grafana for your Apache Kafka cluster, follow the instruction given in Strimzi documentation sections
[Using Prometheus with Strimzi](https://strimzi.io/docs/operators/latest/deploying.html#assembly-metrics-setup-str) and
[Enabling the example Grafana dashboards](https://strimzi.io/docs/operators/latest/deploying.html#proc-metrics-grafana-dashboard-str) to deploy and setup Prometheus and Grafana, respectively.

To deploy the PodMonitor resource for Strimzi canary and to use the example Grafana dashboard,
edit the `PodMonitor` resource in `prometheus-install/canary-monitor.yaml` to
set the `namespaceSelector.matchNames` property to the namespace where Strimzi canary is deployed, and
run the following command:

```shell
kubectl apply -f prometheus-install/canary-monitor.yaml
```

Finally, import the dashboard file `grafana-dashboards/strimzi-kafka-canary.json` into Grafana.


## Getting help

If you encounter any issues while using Strimzi, you can get help using:

- [#strimzi channel on CNCF Slack](https://slack.cncf.io/)
- [Strimzi Users mailing list](https://lists.cncf.io/g/cncf-strimzi-users/topics)
- [GitHub Discussions](https://github.com/strimzi/strimzi-canary/discussions)

## Contributing

You can contribute by raising any issues you find and/or fixing issues by opening Pull Requests.
All bugs, tasks or enhancements are tracked as [GitHub issues](https://github.com/strimzi/strimzi-canary/issues).

The [development documentation](./development-docs) describe how to build, test and release Strimzi Canary.

## License

Strimzi is licensed under the [Apache License](./LICENSE), Version 2.0
