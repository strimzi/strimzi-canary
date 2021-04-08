//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package config defining the canary configuration parameters
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// environment variables declaration
	BootstrapServersEnvVar       = "KAFKA_BOOTSTRAP_SERVERS"
	TopicEnvVar                  = "TOPIC"
	ReconcileIntervalEnvVar      = "RECONCILE_INTERVAL_MS"
	ClientIDEnvVar               = "CLIENT_ID"
	ConsumerGroupIDEnvVar        = "CONSUMER_GROUP_ID"
	TLSEnabledEnvVar             = "TLS_ENABLED"
	ProducerLatencyBucketsEnvVar = "PRODUCER_LATENCY_BUCKETS"
	EndToEndLatencyBucketsEnvVar = "ENDTOEND_LATENCY_BUCKETS"
	ExpectedClusterSizeEnvVar    = "EXPECTED_CLUSTER_SIZE"

	// default values for environment variables
	BootstrapServersDefault       = "localhost:9092"
	TopicDefault                  = "__strimzi_canary"
	ReconcileIntervalDefault      = 30000
	ClientIDDefault               = "strimzi-canary-client"
	ConsumerGroupIDDefault        = "strimzi-canary-group"
	TLSEnabledDefault             = false
	ProducerLatencyBucketsDefault = "100,200,400,800,1600"
	EndToEndLatencyBucketsDefault = "100,200,400,800,1600"
	ExpectedClusterSizeDefault    = -1 // "dynamic" reassignment is enabled
)

// CanaryConfig defines the canary tool configuration
type CanaryConfig struct {
	BootstrapServers       string
	Topic                  string
	ReconcileInterval      time.Duration
	ClientID               string
	ConsumerGroupID        string
	TLSEnabled             bool
	ProducerLatencyBuckets []float64
	EndToEndLatencyBuckets []float64
	ExpectedClusterSize    int
}

// NewCanaryConfig returns an configuration instance from environment variables
func NewCanaryConfig() *CanaryConfig {
	var config CanaryConfig = CanaryConfig{
		BootstrapServers:       lookupStringEnv(BootstrapServersEnvVar, BootstrapServersDefault),
		Topic:                  lookupStringEnv(TopicEnvVar, TopicDefault),
		ReconcileInterval:      time.Duration(lookupIntEnv(ReconcileIntervalEnvVar, ReconcileIntervalDefault)),
		ClientID:               lookupStringEnv(ClientIDEnvVar, ClientIDDefault),
		ConsumerGroupID:        lookupStringEnv(ConsumerGroupIDEnvVar, ConsumerGroupIDDefault),
		TLSEnabled:             lookupBoolEnv(TLSEnabledEnvVar, TLSEnabledDefault),
		ProducerLatencyBuckets: latencyBuckets(lookupStringEnv(ProducerLatencyBucketsEnvVar, ProducerLatencyBucketsDefault)),
		EndToEndLatencyBuckets: latencyBuckets(lookupStringEnv(EndToEndLatencyBucketsEnvVar, EndToEndLatencyBucketsDefault)),
		ExpectedClusterSize:    lookupIntEnv(ExpectedClusterSizeEnvVar, ExpectedClusterSizeDefault),
	}
	return &config
}

func lookupStringEnv(envVar string, defaultValue string) string {
	envVarValue, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}
	return envVarValue
}

func lookupIntEnv(envVar string, defaultValue int) int {
	envVarValue, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}
	intVal, _ := strconv.Atoi(envVarValue)
	return intVal
}

func lookupBoolEnv(envVar string, defaultValue bool) bool {
	envVarValue, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}
	boolVal, _ := strconv.ParseBool(envVarValue)
	return boolVal
}

func latencyBuckets(bucketsConfig string) []float64 {
	sBuckets := strings.Split(bucketsConfig, ",")
	fBuckets := make([]float64, len(sBuckets))
	for i := 0; i < len(sBuckets); i++ {
		f, err := strconv.ParseFloat(sBuckets[i], 64)
		if err != nil {
			panic(err)
		}
		fBuckets[i] = f
	}
	return fBuckets
}

func (c CanaryConfig) String() string {
	return fmt.Sprintf("{BootstrapServers:%s, Topic:%s, ReconcileInterval:%d ms, ClientID:%s, "+
		"ConsumerGroupID:%s, TLSEnabled:%t, ProducerLatencyBuckets:%v, EndToEndLatencyBuckets:%v, ExpectedClusterSize:%d}",
		c.BootstrapServers, c.Topic, c.ReconcileInterval, c.ClientID, c.ConsumerGroupID,
		c.TLSEnabled, c.ProducerLatencyBuckets, c.EndToEndLatencyBuckets, c.ExpectedClusterSize)
}
