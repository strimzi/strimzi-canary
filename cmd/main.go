//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/servers"
	"github.com/strimzi/strimzi-canary/internal/services"
	"github.com/strimzi/strimzi-canary/internal/workers"
)

func main() {
	// get canary configuration
	canaryConfig := config.NewCanaryConfig()
	log.Printf("Starting Strimzi canary tool with config: %+v\n", canaryConfig)

	httpServer := servers.NewHttpServer()
	httpServer.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	client, err := newClient(canaryConfig)
	if err != nil {
		log.Printf("Error creating new Sarama client: %v", err)
		os.Exit(1)
	}

	topicService := services.NewTopicService(canaryConfig, client)
	producerService := services.NewProducerService(canaryConfig, client)
	consumerService := services.NewConsumerService(canaryConfig, client)

	canaryManager := workers.NewCanaryManager(canaryConfig, topicService, producerService, consumerService)
	canaryManager.Start()

	select {
	case sig := <-signals:
		log.Printf("Got signal: %v\n", sig)
	}
	canaryManager.Stop()
	httpServer.Stop()

	log.Printf("Strimzi canary stopped")
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

	if canaryConfig.SaramaLogEnabled {
		sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	}

	backoff := services.NewBackoff(canaryConfig.BootstrapBackoffMaxAttempts, canaryConfig.BootstrapBackoffScale*time.Millisecond, services.MaxDefault)
	for {
		client, clientErr := sarama.NewClient([]string{canaryConfig.BootstrapServers}, config)
		if clientErr == nil {
			return client, nil
		}
		delay, backoffErr := backoff.Delay()
		if backoffErr != nil {
			log.Printf("Error connecting to the Kafka cluster after %d retries: %v", canaryConfig.BootstrapBackoffMaxAttempts, backoffErr)
			return nil, backoffErr
		}
		log.Printf("Error creating new Sarama client, retrying in %d ms: %v", delay.Milliseconds(), clientErr)
		time.Sleep(delay)
	}
}
