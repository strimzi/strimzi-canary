//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/security"
	"github.com/strimzi/strimzi-canary/internal/servers"
	"github.com/strimzi/strimzi-canary/internal/services"
	"github.com/strimzi/strimzi-canary/internal/workers"
)

var (
	version = "development"

	clientCreationFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "client_creation_error_total",
		Namespace: "strimzi_canary",
		Help:      "Total number of errors while creating Sarama client",
	}, nil)
)

var saramaLogger = log.New(io.Discard, "[Sarama] ", log.LstdFlags)

func main() {
	// get canary configuration
	canaryConfig := config.NewCanaryConfig()

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Errorf("Error on setting logtostderr to true")
	}
	sarama.Logger = saramaLogger

	applyDynamicConfig(&canaryConfig.DynamicCanaryConfig)

	glog.Infof("Starting Strimzi canary tool [%s] with config: %+v", version, canaryConfig)

	dynamicConfigWatcher, err := config.NewDynamicConfigWatcher(canaryConfig, applyDynamicConfig, config.NewDynamicCanaryConfig)
	if err != nil {
		glog.Fatalf("Failed to create dynamic config watcher: %v", err)
	}

	statusService := services.NewStatusServiceService(canaryConfig)
	httpServer := servers.NewHttpServer(statusService)
	httpServer.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	client, err := newClient(canaryConfig)
	if err != nil {
		glog.Fatalf("Error creating new Sarama client: %v", err)
	}

	topicService := services.NewTopicService(canaryConfig, client.Config())
	producerService := services.NewProducerService(canaryConfig, client)
	consumerService := services.NewConsumerService(canaryConfig, client)
	connectionService := services.NewConnectionService(canaryConfig, client.Config())

	canaryManager := workers.NewCanaryManager(canaryConfig, topicService, producerService, consumerService, connectionService, statusService)
	canaryManager.Start()

	sig := <-signals
	glog.Infof("Got signal: %v", sig)
	canaryManager.Stop()
	httpServer.Stop()
	dynamicConfigWatcher.Close()

	glog.Infof("Strimzi canary stopped")
}

func newClient(canaryConfig *config.CanaryConfig) (sarama.Client, error) {
	config := sarama.NewConfig()
	kafkaVersion, err := sarama.ParseKafkaVersion(canaryConfig.KafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = kafkaVersion
	config.ClientID = canaryConfig.ClientID
	// set manual partitioner in order to specify the destination partition on sending
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 0
	config.Consumer.Return.Errors = true
	// this Sarama fix https://github.com/Shopify/sarama/pull/2227 increases the canary e2e latency
	// it shows a potential bug in Sarama. We revert the value back here while waiting for a Sarama fix
	config.Consumer.MaxWaitTime = 250 * time.Millisecond

	if canaryConfig.TLSEnabled {
		config.Net.TLS.Enable = true
		if config.Net.TLS.Config, err = security.NewTLSConfig(canaryConfig); err != nil {
			glog.Fatalf("Error configuring TLS: %v", err)
		}
	}

	if canaryConfig.SASLMechanism != "" {
		if err = security.SetAuthConfig(canaryConfig, config); err != nil {
			glog.Fatalf("Error configuring SASL authentication: %v", err)
		}
	}

	backoff := services.NewBackoff(canaryConfig.BootstrapBackoffMaxAttempts, canaryConfig.BootstrapBackoffScale*time.Millisecond, services.MaxDefault)
	for {
		client, clientErr := sarama.NewClient(canaryConfig.BootstrapServers, config)
		if clientErr == nil {
			return client, nil
		}
		delay, backoffErr := backoff.Delay()
		if backoffErr != nil {
			glog.Errorf("Error connecting to the Kafka cluster after %d retries: %v", canaryConfig.BootstrapBackoffMaxAttempts, backoffErr)
			return nil, backoffErr
		}
		clientCreationFailed.With(nil).Inc()
		glog.Warningf("Error creating new Sarama client, retrying in %d ms: %v", delay.Milliseconds(), clientErr)
		time.Sleep(delay)
	}
}

func applyDynamicConfig(dynamicCanaryConfig *config.DynamicCanaryConfig) {
	if dynamicCanaryConfig.VerbosityLogLevel != nil {
		flag.Set("v", strconv.Itoa(*dynamicCanaryConfig.VerbosityLogLevel))
		flag.Parse()
	}

	if dynamicCanaryConfig.SaramaLogEnabled != nil && *dynamicCanaryConfig.SaramaLogEnabled {
		saramaLogger.SetOutput(os.Stdout)
	} else {
		saramaLogger.SetOutput(io.Discard)
	}
	glog.Warningf("Applied dynamic config %s", dynamicCanaryConfig)
}
