# CHANGELOG

## 0.5.0

* Added support for OpenTelemetry (#197)
  * Only "jaeger" and "otlp" are supported as exporter protocols for tracing. See documentation for more details.

## 0.4.0

* Fixed replicas assignment taking rack configuration into account (#190)
* Use separate sarama clients for producer/consumer in order to reliably measure message latency (#188)
* Adjust default histogram bucket sizes (#194)

## 0.3.0

* Forced cleanup.policy to "delete" for the canary topic (#173)
* Allow to change canary logging verbosity and enable/disable Sarama logger at runtime (#166)
* Added Prometheus support via `PodMonitor` and sample Grafana dashboard for exposed metrics
* Treat ETIMEDOUT (TCP keep-alive failure) as a disconnection condition too (#159)
* Updated dependency to Sarama 1.34.0 (#180)
* Updated default Kafka version used by Sarama library to 3.1.0
* Use 250ms as consumer fetch max wait timeout with Sarama 1.34.0 (#184)

## 0.2.0

* Added support for SCRAM-SHA-256 and SCRAM-SHA-512 authentication (#130)
* Added Python script tool to calculate latencies and quantiles from canary log for producer, consumer and connection buckets
* Turned off producer retry and added a consumer error metric (#133)
* Updated dependency to Sarama 1.30.0 with Kafka 3.0.0 support (#134)
* Fixed connection check service missing latency metrics for one broker (#141)
* Fixed wrong content type "text/plain" returned by status endpoint instead of "application/json" (#144)
* Fixed canary not recovering after a "connection reset by peer" from brokers (#148)

## 0.1.0

First implementation.
