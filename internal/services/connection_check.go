//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/util"
)

var (
	connectionError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "connection_error_total",
		Namespace: "strimzi_canary",
		Help:      "Total number of errors while checking the connection to Kafka brokers",
	}, []string{"brokerid", "connected"})

	// it's defined when the service is created because buckets are configurable
	connectionLatency *prometheus.HistogramVec
)

type ConnectionService struct {
	canaryConfig *config.CanaryConfig
	client       sarama.Client
	admin        sarama.ClusterAdmin
	brokers      []*sarama.Broker
	stop         chan struct{}
	syncStop     sync.WaitGroup
}

// NewConnectionService returns an instance of ConnectionService
func NewConnectionService(canaryConfig *config.CanaryConfig, client sarama.Client) *ConnectionService {
	connectionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "connection_success_latency",
		Namespace: "strimzi_canary",
		Help:      "Latency in milliseconds for established or failed connections",
		Buckets:   canaryConfig.ConnectionCheckLatencyBuckets,
	}, []string{"brokerid", "connected"})

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		glog.Fatalf("Error creating the Sarama cluster admin: %v", err)
	}
	cs := ConnectionService{
		canaryConfig: canaryConfig,
		client:       client,
		admin:        admin,
	}
	return &cs
}

// Open starts the connection check loop
func (cs *ConnectionService) Open() {
	cs.stop = make(chan struct{})
	cs.syncStop.Add(1)

	cs.connectionCheck()

	ticker := time.NewTicker(cs.canaryConfig.ConnectionCheckInterval * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				cs.connectionCheck()
			case <-cs.stop:
				ticker.Stop()
				defer cs.syncStop.Done()
				glog.Infof("Stopping connection check loop")
				return
			}
		}
	}()
}

// Close stops the connection check loop and closes the underneath Sarama admin instance
func (cs *ConnectionService) Close() {
	glog.Infof("Closing connection check service")

	// ask to stop the ticker reconcile loop and wait
	close(cs.stop)
	cs.syncStop.Wait()

	if err := cs.admin.Close(); err != nil {
		glog.Fatalf("Error closing the Sarama cluster admin: %v", err)
	}
	glog.Infof("Connection check service closed")
}

// connectionCheck does a connection check to the Kafka brokers
//
// It uses the Sarama admin client to get brokers metadata and then the low level Broker "object" in order to:
//
// 1. open a connection
// 2. check if the connection was ok
// 3. close the connection
//
// When the expected cluster size is set, the connection check service gets the brokers metadata just on the first check
// in order to report connection errors when one of them is down. When the expected cluster size is not set (canary
// handles the topic partitions reassignment process when Kafka cluster scales), the connection check service gets the brokers
// metadata on each check so it doesn't try to connect to not running brokers (the user could have scaled down the cluster).
//
// It also reports the time needed to open a connection successfully or connection errors as metrics.
func (cs *ConnectionService) connectionCheck() {
	var err error

	if cs.isDynamicScalingEnabled() || cs.brokers == nil {
		cs.brokers, _, err = cs.admin.DescribeCluster()
		if err != nil {
			glog.Errorf("Error describing cluster: %v", err)
			return
		}
	}

	for _, b := range cs.brokers {

		start := util.NowInMilliseconds() // timestamp in milliseconds
		// ignore error because it will be reported by Connected() call if "not connected"
		b.Open(cs.client.Config())
		connected, err := b.Connected()
		duration := util.NowInMilliseconds() - start

		labels := prometheus.Labels{
			"brokerid":  strconv.Itoa(int(b.ID())),
			"connected": strconv.FormatBool(connected),
		}

		if connected {
			b.Close()
		} else {
			connectionError.With(labels).Inc()
		}
		connectionLatency.With(labels).Observe(float64(duration))
		glog.V(1).Infof("Connection to broker %d [%t] in [%d] ms (error [%v])", b.ID(), connected, duration, err)
	}
}

// If the "dynamic" scaling is enabled
func (cs *ConnectionService) isDynamicScalingEnabled() bool {
	return cs.canaryConfig.ExpectedClusterSize == config.ExpectedClusterSizeDefault
}
