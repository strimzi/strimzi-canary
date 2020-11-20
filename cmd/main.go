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
	"github.com/strimzi/strimzi-canary/internal/workers"
)

func main() {
	// get canary configuration
	config := config.NewCanaryConfig()
	log.Printf("Starting Strimzi canary tool with config: %+v\n", config)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	topicService := services.NewTopicService(config)
	producerService := services.NewProducerService(config)

	canaryManager := workers.NewCanaryManager(config, topicService, producerService)
	canaryManager.Start()

	consumer := workers.NewConsumer(config)
	consumer.Start()

	select {
	case sig := <-signals:
		log.Printf("Got signal: %v\n", sig)
	}
	canaryManager.Stop()
	consumer.Stop()
	log.Printf("Strimzi canary stopped")
}
