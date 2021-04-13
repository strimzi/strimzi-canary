//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package workers defines an interface for canary workers and related implementations
package workers

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
)

// CanaryManager defines the manager driving the different producer, consumer and topic services
type CanaryManager struct {
	canaryConfig    *config.CanaryConfig
	topicService    *services.TopicService
	producerService *services.ProducerService
	consumerService *services.ConsumerService
	stop            chan struct{}
	syncStop        sync.WaitGroup
}

// NewCanaryManager returns an instance of the cananry manager worker
func NewCanaryManager(canaryConfig *config.CanaryConfig, topicService *services.TopicService, producerService *services.ProducerService, consumerService *services.ConsumerService) Worker {
	cm := CanaryManager{
		canaryConfig:    canaryConfig,
		topicService:    topicService,
		producerService: producerService,
		consumerService: consumerService,
	}
	return &cm
}

// Start runs a first reconcile and start a timer for periodic reconciling
func (cm *CanaryManager) Start() {
	log.Printf("Starting canary manager")

	cm.stop = make(chan struct{})
	cm.syncStop.Add(1)

	// using the same bootstrap configuration that makes sense during the canary start up
	backoff := services.NewBackoff(cm.canaryConfig.BootstrapBackoffMaxAttempts, cm.canaryConfig.BootstrapBackoffScale*time.Millisecond, services.MaxDefault)
	for {
		// start first reconcile immediately
		if result, err := cm.topicService.Reconcile(); err == nil {
			// consumer will subscribe to the topic so all partitions (even if we have less brokers)
			cm.consumerService.Consume(len(result.Assignments))
			// producer just needs to send from partition 0 to brokersNumber - 1
			cm.producerService.Send(result.BrokersNumber)
			break
		} else if e, ok := err.(*services.ErrExpectedClusterSize); ok {
			// if the "dynamic" reassignment is disabled, an error may occur with expected cluster size not met yet
			delay, backoffErr := backoff.Delay()
			if backoffErr != nil {
				log.Printf("Max attempts waiting for the expected cluster size: %v", e)
				os.Exit(1)
			}
			log.Printf("Error on expected cluster size. Retrying in %d ms", delay.Milliseconds())
			time.Sleep(delay)
		} else {
			log.Printf("Error starting manager: %v", err)
			panic(err)
		}
	}

	ticker := time.NewTicker(cm.canaryConfig.ReconcileInterval * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Canary manager reconcile")
				cm.reconcile()
			case <-cm.stop:
				ticker.Stop()
				defer cm.syncStop.Done()
				log.Println("Stopping canary manager reconcile loop")
				return
			}
		}
	}()
}

// Stop stops the services and the reconcile timer
func (cm *CanaryManager) Stop() {
	log.Printf("Stopping canary manager")

	// ask to stop the ticker reconcile loop and wait
	close(cm.stop)
	cm.syncStop.Wait()

	cm.producerService.Close()
	cm.consumerService.Close()
	cm.topicService.Close()

	log.Printf("Canary manager closed")
}

func (cm *CanaryManager) reconcile() {
	log.Printf("Canary manager reconcile ...")

	if result, err := cm.topicService.Reconcile(); err == nil {
		if result.RefreshMetadata {
			// consumer will subscribe to the topic so all partitions (even if we have less brokers)
			cm.consumerService.Refresh(len(result.Assignments))
			cm.producerService.Refresh()
		}
		// producer just needs to send from partition 0 to brokersNumber - 1
		cm.producerService.Send(result.BrokersNumber)
	}

	log.Printf("... reconcile done")
}
