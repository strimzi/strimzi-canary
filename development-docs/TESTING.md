# Testing

The available tests are grouped using two different build tags:

* `unit_test`: for the unit tests alongside the different Go files.
* `e2e`: for running an end-to-end test using a Kafka cluster (via Docker compose) and the canary running alongside it, in order to verify that it's able to exchange messages.

Use the `go test` command in order to run the tests.

## Unit Tests

For running the unit tests:

```shell
go test ./internal/... --tags=unit_test
```

## End-to-end Tests

For running the e2e tests:

```shell
go test ./test/... --tags=e2e
```