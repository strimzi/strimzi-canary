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
	config          *config.CanaryConfig
	topicService    services.TopicService
	producerService services.ProducerService
	stop            chan struct{}
	syncStop        sync.WaitGroup
}

func NewCanaryManager(config *config.CanaryConfig, topicService services.TopicService, producerService services.ProducerService) Worker {
	cm := CanaryManager{
		config:          config,
		topicService:    topicService,
		producerService: producerService,
	}
	return &cm
}

func (cm *CanaryManager) Start() {
	log.Printf("Starting canary manager")

	cm.stop = make(chan struct{})
	cm.syncStop.Add(1)

	// start first reconcile immediately
	cm.reconcile()
	ticker := time.NewTicker(cm.config.TopicReconcile * time.Millisecond)
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

func (cm *CanaryManager) Stop() {
	log.Printf("Stopping canary manager")

	// ask to stop the ticker reconcile loop and wait
	close(cm.stop)
	cm.syncStop.Wait()

	cm.producerService.Close()
	cm.topicService.Close()

	log.Printf("Canary manager closed")
}

func (cm *CanaryManager) reconcile() {
	log.Printf("Canary manager reconcile ...")

	// don't care about assignments, for now.
	// producer just needs to send from partition 0 to brokersNumber - 1
	if brokersNumber, _, err := cm.topicService.Reconcile(); err == nil {
		cm.producerService.Send(brokersNumber)
	}

	log.Printf("... reconcile done")
}
