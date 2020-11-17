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
	"time"
)

const (
	// environment variables declaration
	BootstrapServersEnvVar = "KAFKA_BOOTSTRAP_SERVERS"
	TopicEnvVar            = "TOPIC"
	TopicReconcileEnvVar   = "TOPIC_RECONCILE_MS"
	DelayEnvVar            = "DELAY_MS"
	ProducerClientIDEnvVar = "PRODUCER_CLIENT_ID"
	ConsumerGroupIDEnvVar  = "CONSUMER_GROUP_ID"
	TLSEnabledEnvVar       = "TLS_ENABLED"

	// default values for environment variables
	BootstrapServersDefault = "localhost:9092"
	TopicDefault            = "strimzi-canary"
	TopicReconcileDefault   = 30000
	DelayDefault            = 1000
	ProducerClientIDDefault = "strimzi-canary-producer"
	ConsumerGroupIDDefault  = "strimzi-canary-consumer"
	TLSEnabledDefault       = false
)

// CanaryConfig defines the canary tool configuration
type CanaryConfig struct {
	BootstrapServers string
	Topic            string
	TopicReconcile   time.Duration
	Delay            time.Duration
	ProducerClientID string
	ConsumerGroupID  string
	TLSEnabled       bool
}

func NewCanaryConfig() *CanaryConfig {
	var config CanaryConfig = CanaryConfig{
		BootstrapServers: lookupStringEnv(BootstrapServersEnvVar, BootstrapServersDefault),
		Topic:            lookupStringEnv(TopicEnvVar, TopicDefault),
		TopicReconcile:   time.Duration(lookupIntEnv(TopicReconcileEnvVar, TopicReconcileDefault)),
		Delay:            time.Duration(lookupIntEnv(DelayEnvVar, DelayDefault)),
		ProducerClientID: lookupStringEnv(ProducerClientIDEnvVar, ProducerClientIDDefault),
		ConsumerGroupID:  lookupStringEnv(ConsumerGroupIDEnvVar, ConsumerGroupIDDefault),
		TLSEnabled:       lookupBoolEnv(TLSEnabledEnvVar, TLSEnabledDefault),
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

func (c CanaryConfig) String() string {
	return fmt.Sprintf("{BootstrapServers:%s, Topic:%s, TopicReconcile:%d ms, Delay:%d ms, ProducerClientID:%s, ConsumerGroupID:%s, TLSEnabled:%t}",
		c.BootstrapServers, c.Topic, c.TopicReconcile, c.Delay, c.ProducerClientID, c.ConsumerGroupID, c.TLSEnabled)
}
