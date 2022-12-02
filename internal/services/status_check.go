//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/util"
)

// Status defines useful status related information
type Status struct {
	Consuming ConsumingStatus
}

// ConsumingStatus defines consuming related status information
type ConsumingStatus struct {
	TimeWindow time.Duration
	Percentage float64
}

type StatusService interface {
	Open()
	Close()
	StatusHandler() http.Handler
}

type statusService struct {
	canaryConfig           *config.CanaryConfig
	producedRecordsSamples util.TimeWindowRing
	consumedRecordsSamples util.TimeWindowRing
	stop                   chan struct{}
	syncStop               sync.WaitGroup
}

// NewStatusService returns an instance of StatusService
func NewStatusServiceService(canaryConfig *config.CanaryConfig) StatusService {
	ss := statusService{
		canaryConfig:           canaryConfig,
		producedRecordsSamples: *util.NewTimeWindowRing(canaryConfig.StatusTimeWindow, canaryConfig.StatusCheckInterval),
		consumedRecordsSamples: *util.NewTimeWindowRing(canaryConfig.StatusTimeWindow, canaryConfig.StatusCheckInterval),
	}
	return &ss
}

// Open starts the status check loop
func (ss *statusService) Open() {
	ss.stop = make(chan struct{})
	ss.syncStop.Add(1)

	ticker := time.NewTicker(ss.canaryConfig.StatusCheckInterval * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				ss.statusCheck()
			case <-ss.stop:
				ticker.Stop()
				defer ss.syncStop.Done()
				glog.Infof("Stopping status check loop")
				return
			}
		}
	}()
}

// Close stops the status check loop
func (ss *statusService) Close() {
	glog.Infof("Closing status check service")

	// ask to stop the ticker reconcile loop and wait
	close(ss.stop)
	ss.syncStop.Wait()

	glog.Infof("Status check service closed")
}

// statusCheck does a check of produced and consumed records to fill the time window ring buffers
func (ss *statusService) statusCheck() {
	ss.producedRecordsSamples.Put(RecordsProducedCounter)
	ss.consumedRecordsSamples.Put(RecordsConsumedCounter)
	glog.V(1).Infof("Status check: produced [head = %d, tail = %d, count = %d], consumed [head = %d, tail = %d, count = %d]",
		ss.producedRecordsSamples.Head(), ss.producedRecordsSamples.Tail(), ss.producedRecordsSamples.Count(),
		ss.consumedRecordsSamples.Head(), ss.consumedRecordsSamples.Tail(), ss.consumedRecordsSamples.Count())
}

func (ss *statusService) StatusHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		status := Status{}

		// update consuming related status section
		status.Consuming = ConsumingStatus{
			TimeWindow: ss.canaryConfig.StatusCheckInterval * time.Duration(ss.consumedRecordsSamples.Count()),
		}
		consumedPercentage, err := ss.consumedPercentage()
		if e, ok := err.(*util.ErrNoDataSamples); ok {
			status.Consuming.Percentage = -1
			glog.Errorf("Error processing consumed records percentage: %v", e)
		} else {
			status.Consuming.Percentage = consumedPercentage
		}

		json, _ := json.Marshal(status)
		rw.Header().Add("Content-Type", "application/json")
		rw.Write(json)
	})
}

// consumedPercentage function processes the percentage of consumed messages in the specified time window
func (ss *statusService) consumedPercentage() (float64, error) {
	// sampling for produced (and consumed records) not done yet
	if ss.producedRecordsSamples.IsEmpty() {
		return 0, &util.ErrNoDataSamples{}
	}

	// get number of records consumed and produced since the beginning of the time window (tail of ring buffers)
	consumed := ss.consumedRecordsSamples.Head() - ss.consumedRecordsSamples.Tail()
	produced := ss.producedRecordsSamples.Head() - ss.producedRecordsSamples.Tail()

	if produced == 0 {
		return 0, &util.ErrNoDataSamples{}
	}

	percentage := float64(consumed*100) / float64(produced)
	// rounding to two decimal digits
	percentage = math.Round(percentage*100) / 100
	glog.V(1).Infof("Status consumed percentage = %f", percentage)
	return percentage, nil
}
