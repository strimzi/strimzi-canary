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

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
)

const (
	// environment variables declaration
	BootstrapServersEnvVar              = "KAFKA_BOOTSTRAP_SERVERS"
	BootstrapBackoffMaxAttemptsEnvVar   = "KAFKA_BOOTSTRAP_BACKOFF_MAX_ATTEMPTS"
	BootstrapBackoffScaleEnvVar         = "KAFKA_BOOTSTRAP_BACKOFF_SCALE"
	TopicEnvVar                         = "TOPIC"
	TopicConfigEnvVar                   = "TOPIC_CONFIG"
	ReconcileIntervalEnvVar             = "RECONCILE_INTERVAL_MS"
	ClientIDEnvVar                      = "CLIENT_ID"
	ConsumerGroupIDEnvVar               = "CONSUMER_GROUP_ID"
	ProducerLatencyBucketsEnvVar        = "PRODUCER_LATENCY_BUCKETS"
	EndToEndLatencyBucketsEnvVar        = "ENDTOEND_LATENCY_BUCKETS"
	ExpectedClusterSizeEnvVar           = "EXPECTED_CLUSTER_SIZE"
	KafkaVersionEnvVar                  = "KAFKA_VERSION"
	SaramaLogEnabledEnvVar              = "SARAMA_LOG_ENABLED"
	VerbosityLogLevelEnvVar             = "VERBOSITY_LOG_LEVEL"
	TLSEnabledEnvVar                    = "TLS_ENABLED"
	TLSCACertEnvVar                     = "TLS_CA_CERT"
	TLSClientCertEnvVar                 = "TLS_CLIENT_CERT"
	TLSClientKeyEnvVar                  = "TLS_CLIENT_KEY"
	TLSInsecureSkipVerifyEnvVar         = "TLS_INSECURE_SKIP_VERIFY"
	SASLMechanismEnvVar                 = "SASL_MECHANISM"
	SASLUserEnvVar                      = "SASL_USER"
	SASLPasswordEnvVar                  = "SASL_PASSWORD"
	ConnectionCheckIntervalEnvVar       = "CONNECTION_CHECK_INTERVAL_MS"
	ConnectionCheckLatencyBucketsEnvVar = "CONNECTION_CHECK_LATENCY_BUCKETS"
	StatusCheckIntervalEnvVar           = "STATUS_CHECK_INTERVAL_MS"
	StatusTimeWindowEnvVar              = "STATUS_TIME_WINDOW_MS"

	// default values for environment variables
	BootstrapServersDefault              = "localhost:9092"
	BootstrapBackoffMaxAttemptsDefault   = 10
	BootstrapBackoffScaleDefault         = 5000
	TopicDefault                         = "__strimzi_canary"
	TopicConfigDefault                   = ""
	ReconcileIntervalDefault             = 30000
	ClientIDDefault                      = "strimzi-canary-client"
	ConsumerGroupIDDefault               = "strimzi-canary-group"
	ProducerLatencyBucketsDefault        = "100,200,400,800,1600"
	EndToEndLatencyBucketsDefault        = "100,200,400,800,1600"
	ExpectedClusterSizeDefault           = -1 // "dynamic" reassignment is enabled
	KafkaVersionDefault                  = "2.8.0"
	SaramaLogEnabledDefault              = false
	VerbosityLogLevelDefault             = 0 // default 0 = INFO, 1 = DEBUG, 2 = TRACE
	TLSEnabledDefault                    = false
	TLSCACertDefault                     = ""
	TLSClientCertDefault                 = ""
	TLSClientKeyDefault                  = ""
	TLSInsecureSkipVerifyDefault         = false
	SASLMechanismDefault                 = ""
	SASLUserDefault                      = ""
	SASLPasswordDefault                  = ""
	ConnectionCheckIntervalDefault       = 120000
	ConnectionCheckLatencyBucketsDefault = "100,200,400,800,1600"
	StatusCheckIntervalDefault           = 30000
	StatusTimeWindowDefault              = 300000
)

// CanaryConfig defines the canary tool configuration
type CanaryConfig struct {
	BootstrapServers              []string
	BootstrapBackoffMaxAttempts   int
	BootstrapBackoffScale         time.Duration
	Topic                         string
	TopicConfig                   map[string]string
	ReconcileInterval             time.Duration
	ClientID                      string
	ConsumerGroupID               string
	ProducerLatencyBuckets        []float64
	EndToEndLatencyBuckets        []float64
	ExpectedClusterSize           int
	KafkaVersion                  string
	SaramaLogEnabled              bool
	VerbosityLogLevel             int
	TLSEnabled                    bool
	TLSCACert                     string
	TLSClientCert                 string
	TLSClientKey                  string
	TLSInsecureSkipVerify         bool
	SASLMechanism                 string
	SASLUser                      string
	SASLPassword                  string
	ConnectionCheckInterval       time.Duration
	ConnectionCheckLatencyBuckets []float64
	StatusCheckInterval           time.Duration
	StatusTimeWindow              time.Duration
}

// NewCanaryConfig returns an configuration instance from environment variables
func NewCanaryConfig() *CanaryConfig {
	var config CanaryConfig = CanaryConfig{
		BootstrapServers:              strings.Split(lookupStringEnv(BootstrapServersEnvVar, BootstrapServersDefault), ","),
		BootstrapBackoffMaxAttempts:   lookupIntEnv(BootstrapBackoffMaxAttemptsEnvVar, BootstrapBackoffMaxAttemptsDefault),
		BootstrapBackoffScale:         time.Duration(lookupIntEnv(BootstrapBackoffScaleEnvVar, BootstrapBackoffScaleDefault)),
		Topic:                         lookupStringEnv(TopicEnvVar, TopicDefault),
		TopicConfig:                   topicConfig(lookupStringEnv(TopicConfigEnvVar, TopicConfigDefault)),
		ReconcileInterval:             time.Duration(lookupIntEnv(ReconcileIntervalEnvVar, ReconcileIntervalDefault)),
		ClientID:                      lookupStringEnv(ClientIDEnvVar, ClientIDDefault),
		ConsumerGroupID:               lookupStringEnv(ConsumerGroupIDEnvVar, ConsumerGroupIDDefault),
		ProducerLatencyBuckets:        latencyBuckets(lookupStringEnv(ProducerLatencyBucketsEnvVar, ProducerLatencyBucketsDefault)),
		EndToEndLatencyBuckets:        latencyBuckets(lookupStringEnv(EndToEndLatencyBucketsEnvVar, EndToEndLatencyBucketsDefault)),
		ExpectedClusterSize:           lookupIntEnv(ExpectedClusterSizeEnvVar, ExpectedClusterSizeDefault),
		KafkaVersion:                  lookupStringEnv(KafkaVersionEnvVar, KafkaVersionDefault),
		SaramaLogEnabled:              lookupBoolEnv(SaramaLogEnabledEnvVar, SaramaLogEnabledDefault),
		VerbosityLogLevel:             lookupIntEnv(VerbosityLogLevelEnvVar, VerbosityLogLevelDefault),
		TLSEnabled:                    lookupBoolEnv(TLSEnabledEnvVar, TLSEnabledDefault),
		TLSCACert:                     lookupStringEnv(TLSCACertEnvVar, TLSCACertDefault),
		TLSClientCert:                 lookupStringEnv(TLSClientCertEnvVar, TLSClientCertDefault),
		TLSClientKey:                  lookupStringEnv(TLSClientKeyEnvVar, TLSClientKeyDefault),
		TLSInsecureSkipVerify:         lookupBoolEnv(TLSInsecureSkipVerifyEnvVar, TLSInsecureSkipVerifyDefault),
		SASLMechanism:                 lookupStringEnv(SASLMechanismEnvVar, SASLMechanismDefault),
		SASLUser:                      lookupStringEnv(SASLUserEnvVar, SASLUserDefault),
		SASLPassword:                  lookupStringEnv(SASLPasswordEnvVar, SASLPasswordDefault),
		ConnectionCheckInterval:       time.Duration(lookupIntEnv(ConnectionCheckIntervalEnvVar, ConnectionCheckIntervalDefault)),
		ConnectionCheckLatencyBuckets: latencyBuckets(lookupStringEnv(ConnectionCheckLatencyBucketsEnvVar, ConnectionCheckLatencyBucketsDefault)),
		StatusCheckInterval:           time.Duration(lookupIntEnv(StatusCheckIntervalEnvVar, StatusCheckIntervalDefault)),
		StatusTimeWindow:              time.Duration(lookupIntEnv(StatusTimeWindowEnvVar, StatusTimeWindowDefault)),
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
			glog.Fatalf("Error parsing buckets configuration for %s: %v", bucketsConfig, err)
		}
		fBuckets[i] = f
	}
	return fBuckets
}

func topicConfig(topicConfig string) map[string]string {
	if len(topicConfig) == 0 {
		return nil
	}

	mapTopicConfig := make(map[string]string)
	kvPairs := strings.Split(topicConfig, ";")
	for i, kvPair := range kvPairs {
		// just exits if latest kvPair is empty (allows trailing ";")
		if i == len(kvPairs)-1 && len(kvPair) == 0 {
			break
		}
		kv := strings.Split(kvPair, "=")
		// key-value pair split has to have two fields (key has to be not empty)
		if len(kv) != 2 || len(kv[0]) == 0 {
			panic(fmt.Errorf("error parsing topic configuration [%s]: [%s] is not a valid key-value pair", topicConfig, kvPair))
		}
		mapTopicConfig[kv[0]] = kv[1]
	}
	return mapTopicConfig
}

func (c CanaryConfig) String() string {

	// just using placeholders for certs/keys (content or paths)
	TLSCACert := ""
	if c.TLSCACert != "" {
		TLSCACert = "[CA cert]"
	}
	TLSClientCert := ""
	if c.TLSClientCert != "" {
		TLSClientCert = "[Client cert]"
	}
	TLSClientKey := ""
	if c.TLSClientKey != "" {
		TLSClientKey = "[Client key]"
	}

	// is one of SASL mechanisms needing user/password is enabled, using placeholders for them
	SASLUser := ""
	SASLPassword := ""
	if c.SASLMechanism == sarama.SASLTypePlaintext {
		if c.SASLUser != "" {
			SASLUser = "[SASL user]"
		}

		if c.SASLPassword != "" {
			SASLPassword = "[SASL password]"
		}
	}

	return fmt.Sprintf("{BootstrapServers:%s, BootstrapBackoffMaxAttempts:%d, BootstrapBackoffScale:%d, Topic:%s, TopicConfig:%v, ReconcileInterval:%d ms, "+
		"ClientID:%s, ConsumerGroupID:%s, ProducerLatencyBuckets:%v, EndToEndLatencyBuckets:%v, ExpectedClusterSize:%d, KafkaVersion:%s,"+
		"SaramaLogEnabled:%t, VerbosityLogLevel:%d, TLSEnabled:%t, TLSCACert:%s, TLSClientCert:%s, TLSClientKey:%s, TLSInsecureSkipVerify:%t,"+
		"SASLMechanism:%s, SASLUser:%s, SASLPassword:%s, ConnectionCheckInterval:%d ms, ConnectionCheckLatencyBuckets:%v, StatusCheckInterval:%d ms, StatusTimeWindow:%d ms}",
		c.BootstrapServers, c.BootstrapBackoffMaxAttempts, c.BootstrapBackoffScale, c.Topic, c.TopicConfig, c.ReconcileInterval, c.ClientID, c.ConsumerGroupID,
		c.ProducerLatencyBuckets, c.EndToEndLatencyBuckets, c.ExpectedClusterSize, c.KafkaVersion, c.SaramaLogEnabled, c.VerbosityLogLevel,
		c.TLSEnabled, TLSCACert, TLSClientCert, TLSClientKey, c.TLSInsecureSkipVerify, c.SASLMechanism, SASLUser, SASLPassword,
		c.ConnectionCheckInterval, c.ConnectionCheckLatencyBuckets, c.StatusCheckInterval, c.StatusTimeWindow)
}
