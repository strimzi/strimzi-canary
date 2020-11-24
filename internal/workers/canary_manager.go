//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package workers defines an interface for canary workers and related implementations
package workers

import (
	"log"
	"sync"
	"time"

	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
)

type CanaryManager struct {
	canaryConfig    *config.CanaryConfig
	topicService    *services.TopicService
	producerService *services.ProducerService
	consumerService *services.ConsumerService
	stop            chan struct{}
	syncStop        sync.WaitGroup
}

func NewCanaryManager(canaryConfig *config.CanaryConfig, topicService *services.TopicService, producerService *services.ProducerService, consumerService *services.ConsumerService) Worker {
	cm := CanaryManager{
		canaryConfig:    canaryConfig,
		topicService:    topicService,
		producerService: producerService,
		consumerService: consumerService,
	}
	return &cm
}

func (cm *CanaryManager) Start() {
	log.Printf("Starting canary manager")

	cm.stop = make(chan struct{})
	cm.syncStop.Add(1)

	// start first reconcile immediately
	cm.reconcile()
	ticker := time.NewTicker(cm.canaryConfig.TopicReconcile * time.Millisecond)
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
	// start consumer service to get messages
	cm.consumerService.Consume()
}

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

	// don't care about assignments, for now.
	// producer just needs to send from partition 0 to brokersNumber - 1
	if result, err := cm.topicService.Reconcile(); err == nil {
		if result.RefreshMetadata {
			cm.consumerService.Refresh()
			cm.producerService.Refresh()
		}
		cm.producerService.Send(result.BrokersNumber)
	}

	log.Printf("... reconcile done")
}
