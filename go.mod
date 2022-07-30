module github.com/strimzi/strimzi-canary

go 1.13

require (
	github.com/Shopify/sarama v1.35.0
	github.com/golang/glog v1.0.0
	github.com/prometheus/client_golang v1.12.1
	github.com/xdg-go/scram v1.1.1
	go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama v0.32.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/jaeger v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)
