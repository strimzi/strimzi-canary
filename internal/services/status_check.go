//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/util"
)

type StatusService struct {
	canaryConfig           *config.CanaryConfig
	producedRecordsSamples util.TimeWindowRing
	consumedRecordsSamples util.TimeWindowRing
	stop                   chan struct{}
	syncStop               sync.WaitGroup
}

// NewStatusService returns an instance of StatusService
func NewStatusServiceService(canaryConfig *config.CanaryConfig) *StatusService {
	ss := StatusService{
		canaryConfig:           canaryConfig,
		producedRecordsSamples: *util.NewTimeWindowRing(canaryConfig.StatusTimeWindow, canaryConfig.StatusCheckInterval),
		consumedRecordsSamples: *util.NewTimeWindowRing(canaryConfig.StatusTimeWindow, canaryConfig.StatusCheckInterval),
	}
	return &ss
}

// Open starts the status check loop
func (ss *StatusService) Open() {
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
func (ss *StatusService) Close() {
	glog.Infof("Closing status check service")

	// ask to stop the ticker reconcile loop and wait
	close(ss.stop)
	ss.syncStop.Wait()

	glog.Infof("Status check service closed")
}

// statusCheck does a check of produced and consumed records to fill the time window ring buffers
func (ss *StatusService) statusCheck() {
	ss.producedRecordsSamples.Put(RecordsProducedCounter)
	ss.consumedRecordsSamples.Put(RecordsConsumedCounter)
	glog.Infof("Status check: produced [head = %d, tail = %d, count = %d], consumed [head = %d, tail = %d, count = %d]",
		ss.producedRecordsSamples.Head(), ss.producedRecordsSamples.Tail(), ss.producedRecordsSamples.Count(),
		ss.consumedRecordsSamples.Head(), ss.consumedRecordsSamples.Tail(), ss.consumedRecordsSamples.Count())
}

func (ss *StatusService) StatusHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// sampling for produced (and consumed records) not done yet
		if ss.producedRecordsSamples.IsEmpty() {
			rw.Write([]byte("KO"))
			return
		}

		// get number of records consumed and produced since the beginning of the time window (tail of ring buffers)
		consumed := ss.consumedRecordsSamples.Head() - ss.consumedRecordsSamples.Tail()
		produced := ss.producedRecordsSamples.Head() - ss.producedRecordsSamples.Tail()

		if produced == 0 {
			rw.Write([]byte("KO"))
			return
		}

		percentage := consumed * 100 / produced
		glog.Infof("Percentage = %d", percentage)
		rw.Write([]byte("OK"))
	})
}
