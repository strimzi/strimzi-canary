//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package workers defines an interface for canary workers and related implementations
package workers

import (
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
)

// CanaryManager defines the manager driving the different producer, consumer and topic services
type CanaryManager struct {
	canaryConfig      *config.CanaryConfig
	topicService      services.TopicService
	producerService   services.ProducerService
	consumerService   services.ConsumerService
	connectionService services.ConnectionService
	statusService     services.StatusService
	stop              chan struct{}
	syncStop          sync.WaitGroup
}

var (
	expectedClusterSizeError *prometheus.CounterVec
)

// NewCanaryManager returns an instance of the cananry manager worker
func NewCanaryManager(canaryConfig *config.CanaryConfig,
	topicService services.TopicService, producerService services.ProducerService,
	consumerService services.ConsumerService, connectionService services.ConnectionService,
	statusService services.StatusService) Worker {

	cm := CanaryManager{
		canaryConfig:      canaryConfig,
		topicService:      topicService,
		producerService:   producerService,
		consumerService:   consumerService,
		connectionService: connectionService,
		statusService:     statusService,
	}
	return &cm
}

func (cm *CanaryManager) RegisterMetrics() {
	expectedClusterSizeError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "expected_cluster_size_error_total",
		Namespace:   "strimzi_canary",
		Help:        "Total number of errors while waiting the Kafka cluster having the expected size",
		ConstLabels: cm.canaryConfig.PrometheusConstantLabels,
	}, nil)
}

// Start runs a first reconcile and start a timer for periodic reconciling
func (cm *CanaryManager) Start() {
	glog.Infof("Starting canary manager")

	cm.stop = make(chan struct{})
	cm.syncStop.Add(1)

	cm.connectionService.Open()
	cm.statusService.Open()

	// using the same bootstrap configuration that makes sense during the canary start up
	backoff := services.NewBackoff(cm.canaryConfig.BootstrapBackoffMaxAttempts, cm.canaryConfig.BootstrapBackoffScale*time.Millisecond, services.MaxDefault)
	for {
		// start first reconcile immediately
		if result, err := cm.topicService.Reconcile(); err == nil {
			// consumer will subscribe to the topic so all partitions (even if we have less brokers)
			cm.consumerService.Consume()
			// producer has to send to partitions assigned to brokers
			cm.producerService.Send(result.Assignments)
			break
		} else if e, ok := err.(*services.ErrExpectedClusterSize); ok {
			// if the "dynamic" reassignment is disabled, an error may occur with expected cluster size not met yet
			delay, backoffErr := backoff.Delay()
			if backoffErr != nil {
				glog.Fatalf("Max attempts waiting for the expected cluster size: %v", e)
			}
			expectedClusterSizeError.With(nil).Inc()
			glog.Warningf("Error on expected cluster size. Retrying in %d ms", delay.Milliseconds())
			time.Sleep(delay)
		} else {
			glog.Fatalf("Error starting canary manager: %v", err)
		}
	}

	ticker := time.NewTicker(cm.canaryConfig.ReconcileInterval * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				cm.reconcile()
			case <-cm.stop:
				ticker.Stop()
				defer cm.syncStop.Done()
				glog.Infof("Stopping canary manager reconcile loop")
				return
			}
		}
	}()
}

// Stop stops the services and the reconcile timer
func (cm *CanaryManager) Stop() {
	glog.Infof("Stopping canary manager")

	// ask to stop the ticker reconcile loop and wait
	close(cm.stop)
	cm.syncStop.Wait()

	cm.producerService.Close()
	cm.consumerService.Close()
	cm.topicService.Close()
	cm.connectionService.Close()
	cm.statusService.Close()

	glog.Infof("Canary manager closed")
}

func (cm *CanaryManager) reconcile() {
	glog.Infof("Canary manager reconcile ...")

	if result, err := cm.topicService.Reconcile(); err == nil {
		if result.RefreshProducerMetadata {
			cm.producerService.Refresh()
		}

		leaders, err := cm.consumerService.Leaders()
		if err != nil || !reflect.DeepEqual(result.Leaders, leaders) {
			cm.consumerService.Refresh()
		}
		// producer has to send to partitions assigned to brokers
		cm.producerService.Send(result.Assignments)
	}

	glog.Infof("... reconcile done")
}
