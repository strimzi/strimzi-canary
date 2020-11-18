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

	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
)

func main() {
	// get canary configuration
	config := config.NewCanaryConfig()
	log.Printf("Starting Strimzi canary tool with config: %+v\n", config)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	topicManager := services.NewTopicManager(config)
	topicManager.Start()

	producer := services.NewProducer(config)
	producer.Start()

	consumer := services.NewConsumer(config)
	consumer.Start()

	select {
	case sig := <-signals:
		log.Printf("Got signal: %v\n", sig)
	}
	topicManager.Stop()
	producer.Stop()
	consumer.Stop()
	log.Printf("Strimzi canary stopped")
}
