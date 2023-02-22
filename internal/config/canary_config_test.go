//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

//go:build unit_test

// Package config defining the canary configuration parameters
package config

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfigDefault(t *testing.T) {
	c := NewCanaryConfig()
	bootstrapServersDefault := strings.Split(BootstrapServersDefault, ",")
	assertStringSlicesConfigParameter(c.BootstrapServers, bootstrapServersDefault, t)
	assertIntConfigParameter(c.BootstrapBackoffMaxAttempts, BootstrapBackoffMaxAttemptsDefault, t)
	assertDurationConfigParameter(c.BootstrapBackoffScale, BootstrapBackoffScaleDefault, t)
	assertStringConfigParameter(c.Topic, TopicDefault, t)
	topicConfigDefault := convertKVPairsToMap(TopicConfigDefault)
	assertMapConfigParameter(c.TopicConfig, topicConfigDefault, t)
	assertDurationConfigParameter(c.ReconcileInterval, ReconcileIntervalDefault, t)
	assertStringConfigParameter(c.ClientID, ClientIDDefault, t)
	assertStringConfigParameter(c.ConsumerGroupID, ConsumerGroupIDDefault, t)
	producerLatencyBucketsDefault := latencyBuckets(ProducerLatencyBucketsDefault)
	assertBucketsConfigParameter(c.ProducerLatencyBuckets, producerLatencyBucketsDefault, t)
	endToEndLatencyBucketsDefault := latencyBuckets(EndToEndLatencyBucketsDefault)
	assertBucketsConfigParameter(c.EndToEndLatencyBuckets, endToEndLatencyBucketsDefault, t)
	assertIntConfigParameter(c.ExpectedClusterSize, ExpectedClusterSizeDefault, t)
	assertStringConfigParameter(c.KafkaVersion, KafkaVersionDefault, t)
	assertBoolConfigParameter(*c.SaramaLogEnabled, SaramaLogEnabledDefault, t)
	assertIntConfigParameter(*c.VerbosityLogLevel, VerbosityLogLevelDefault, t)
	assertBoolConfigParameter(c.TLSEnabled, TLSEnabledDefault, t)
	assertStringConfigParameter(c.TLSCACert, TLSCACertDefault, t)
	assertStringConfigParameter(c.TLSClientCert, TLSClientCertDefault, t)
	assertStringConfigParameter(c.TLSClientKey, TLSClientKeyDefault, t)
	assertBoolConfigParameter(c.TLSInsecureSkipVerify, TLSInsecureSkipVerifyDefault, t)
	assertStringConfigParameter(c.SASLMechanism, SASLMechanismDefault, t)
	assertStringConfigParameter(c.SASLUser, SASLUserDefault, t)
	assertStringConfigParameter(c.SASLPassword, SASLPasswordDefault, t)
	assertDurationConfigParameter(c.ConnectionCheckInterval, ConnectionCheckIntervalDefault, t)
	connectionCheckLatencyBucketsDefault := latencyBuckets(ConnectionCheckLatencyBucketsDefault)
	assertBucketsConfigParameter(c.ConnectionCheckLatencyBuckets, connectionCheckLatencyBucketsDefault, t)
	assertDurationConfigParameter(c.StatusCheckInterval, StatusCheckIntervalDefault, t)
	assertDurationConfigParameter(c.StatusTimeWindow, StatusTimeWindowDefault, t)
	prometheusConsantLabelsDefault := convertKVPairsToPrometheusLabels(PrometheusConsantLabelsDefault)
	assertMapConfigParameter(c.PrometheusConstantLabels, prometheusConsantLabelsDefault, t)
}

func TestConfigCustom(t *testing.T) {
	os.Setenv(BootstrapServersEnvVar, "kafka-broker-1:9092,kafka-broker-2:9092")
	os.Setenv(BootstrapBackoffMaxAttemptsEnvVar, "3")
	os.Setenv(BootstrapBackoffScaleEnvVar, "1000")
	os.Setenv(TopicEnvVar, "my-strimzi-canary-topic")
	os.Setenv(TopicConfigEnvVar, "retention.ms=600000;segment.bytes=16384;cleanup.policy=compact,delete")
	os.Setenv(ReconcileIntervalEnvVar, "10000")
	os.Setenv(ClientIDEnvVar, "my-client-id")
	os.Setenv(ConsumerGroupIDEnvVar, "my-consumer-group-id")
	os.Setenv(ProducerLatencyBucketsEnvVar, "400,800,1600,3200,6400")
	os.Setenv(EndToEndLatencyBucketsEnvVar, "800,1600,3200,6400,12800")
	os.Setenv(ExpectedClusterSizeEnvVar, "3")
	os.Setenv(KafkaVersionEnvVar, "2.6.0")
	os.Setenv(SaramaLogEnabledEnvVar, "true")
	os.Setenv(VerbosityLogLevelEnvVar, "1")
	os.Setenv(TLSEnabledEnvVar, "true")
	os.Setenv(TLSCACertEnvVar, "CA cert")
	os.Setenv(TLSClientCertEnvVar, "Client cert")
	os.Setenv(TLSClientKeyEnvVar, "Client key")
	os.Setenv(TLSInsecureSkipVerifyEnvVar, "true")
	os.Setenv(SASLMechanismEnvVar, "PLAIN")
	os.Setenv(SASLUserEnvVar, "user")
	os.Setenv(SASLPasswordEnvVar, "password")
	os.Setenv(ConnectionCheckIntervalEnvVar, "20000")
	os.Setenv(ConnectionCheckLatencyBucketsEnvVar, "200,400,800")
	os.Setenv(StatusCheckIntervalEnvVar, "30000")
	os.Setenv(StatusTimeWindowEnvVar, "200000")
	os.Setenv(PrometheusConsantLabelsEnvVar, "retry.backoff.ms=10000;request.timeout.ms=10000")
	c := NewCanaryConfig()
	bootstrapServers := strings.Split("kafka-broker-1:9092,kafka-broker-2:9092", ",")
	assertStringSlicesConfigParameter(c.BootstrapServers, bootstrapServers, t)
	assertIntConfigParameter(c.BootstrapBackoffMaxAttempts, 3, t)
	assertDurationConfigParameter(c.BootstrapBackoffScale, 1000, t)
	assertStringConfigParameter(c.Topic, "my-strimzi-canary-topic", t)
	topicConfig := convertKVPairsToMap("retention.ms=600000;segment.bytes=16384;cleanup.policy=compact,delete")
	assertMapConfigParameter(c.TopicConfig, topicConfig, t)
	assertDurationConfigParameter(c.ReconcileInterval, 10000, t)
	assertStringConfigParameter(c.ClientID, "my-client-id", t)
	assertStringConfigParameter(c.ConsumerGroupID, "my-consumer-group-id", t)
	producerLatencyBuckets := latencyBuckets("400,800,1600,3200,6400")
	assertBucketsConfigParameter(c.ProducerLatencyBuckets, producerLatencyBuckets, t)
	endToEndLatencyBuckets := latencyBuckets("800,1600,3200,6400,12800")
	assertBucketsConfigParameter(c.EndToEndLatencyBuckets, endToEndLatencyBuckets, t)
	assertIntConfigParameter(c.ExpectedClusterSize, 3, t)
	assertStringConfigParameter(c.KafkaVersion, "2.6.0", t)
	assertBoolConfigParameter(*c.SaramaLogEnabled, true, t)
	assertIntConfigParameter(*c.VerbosityLogLevel, 1, t)
	assertBoolConfigParameter(c.TLSEnabled, true, t)
	assertStringConfigParameter(c.TLSCACert, "CA cert", t)
	assertStringConfigParameter(c.TLSClientCert, "Client cert", t)
	assertStringConfigParameter(c.TLSClientKey, "Client key", t)
	assertBoolConfigParameter(c.TLSInsecureSkipVerify, true, t)
	assertStringConfigParameter(c.SASLMechanism, "PLAIN", t)
	assertStringConfigParameter(c.SASLUser, "user", t)
	assertStringConfigParameter(c.SASLPassword, "password", t)
	assertDurationConfigParameter(c.ConnectionCheckInterval, 20000, t)
	connectionCheckLatencyBuckets := latencyBuckets("200,400,800")
	assertBucketsConfigParameter(c.ConnectionCheckLatencyBuckets, connectionCheckLatencyBuckets, t)
	assertDurationConfigParameter(c.StatusCheckInterval, 30000, t)
	assertDurationConfigParameter(c.StatusTimeWindow, 200000, t)
	prometheusConstantLabels := convertKVPairsToPrometheusLabels("retry.backoff.ms=10000;request.timeout.ms=10000")
	assertPrometheusLabelsParameter(c.PrometheusConstantLabels, prometheusConstantLabels, t)
}

func TestTopicConfigurationNoKey(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(TopicConfigEnvVar, "=600000;segment.bytes=16384;cleanup.policy=compact,delete")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func TestTopicConfigurationInvalidKeyValuePair(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(TopicConfigEnvVar, "aaaaa;segment.bytes=16384;cleanup.policy=compact,delete")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func TestTopicConfigurationEmptyKeyValuePair(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(TopicConfigEnvVar, ";;;;segment.bytes=16384;cleanup.policy=compact,delete")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func TestPrometheusConstantLabelsNoKey(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(PrometheusConsantLabelsEnvVar, "=10000;request.timeout.ms=10000")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func TestPrometheusConstantLabelsInvalidKeyValuePair(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(PrometheusConsantLabelsEnvVar, "aaaaaaaaaaaa;retry.backoff.ms=10000;request.timeout.ms=10000")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func TestPrometheusConstantLabelsEmptyKeyValuePair(t *testing.T) {
	defer func() { recover() }()
	os.Setenv(PrometheusConsantLabelsEnvVar, ";;;;retry.backoff.ms=10000;request.timeout.ms=10000")
	NewCanaryConfig()
	t.Errorf("Should have been panicked!")
}

func assertStringSlicesConfigParameter(value []string, defaultValue []string, t *testing.T) {
	if len(value) != len(defaultValue) {
		t.Errorf("Different lengths got = %d, want = %d", len(value), len(defaultValue))
	}
	for i, v := range value {
		if v != defaultValue[i] {
			t.Errorf("got = %s, want = %s", value, defaultValue)
		}
	}
}

func assertStringConfigParameter(value string, defaultValue string, t *testing.T) {
	if value != defaultValue {
		t.Errorf("got = %s, want = %s", value, defaultValue)
	}
}

func assertIntConfigParameter(value int, defaultValue int, t *testing.T) {
	if value != defaultValue {
		t.Errorf("got = %d, want = %d", value, defaultValue)
	}
}

func assertDurationConfigParameter(value time.Duration, defaultValue time.Duration, t *testing.T) {
	if value != defaultValue {
		t.Errorf("got = %d, want = %d", value, defaultValue)
	}
}

func assertBucketsConfigParameter(value []float64, defaultValue []float64, t *testing.T) {
	for i, v := range value {
		if v != defaultValue[i] {
			t.Errorf("got = %v, want = %v", v, defaultValue[i])
		}
	}
}

func assertBoolConfigParameter(value bool, defaultValue bool, t *testing.T) {
	if value != defaultValue {
		t.Errorf("got = %t, want = %t", value, defaultValue)
	}
}

func assertMapConfigParameter(value map[string]string, defaultValue map[string]string, t *testing.T) {
	for i, v := range value {
		if v != defaultValue[i] {
			t.Errorf("got = %v, want = %v", v, defaultValue[i])
		}
	}
}

func assertPrometheusLabelsParameter(value prometheus.Labels, defaultValue prometheus.Labels, t *testing.T) {
	for i, v := range value {
		if v != defaultValue[i] {
			t.Errorf("got = %v, want = %v", v, defaultValue[i])
		}
	}
}
