//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"context"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/util"
)

var (
	RecordsProducedCounter uint64 = 0

	recordsProduced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "records_produced_total",
		Namespace: "strimzi_canary",
		Help:      "The total number of records produced",
	}, []string{"clientid", "partition"})

	recordsProducedFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "records_produced_failed_total",
		Namespace: "strimzi_canary",
		Help:      "The total number of records failed to produce",
	}, []string{"clientid", "partition"})

	// it's defined when the service is created because buckets are configurable
	recordsProducedLatency *prometheus.HistogramVec

	refreshMetadataError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "producer_refresh_metadata_error_total",
		Namespace: "strimzi_canary",
		Help:      "Total number of errors while refreshing producer metadata",
	}, []string{"clientid"})
)

type ProducerService interface {
	Send(partitionsAssignments map[int32][]int32)
	Refresh()
	Close()
}

// ProducerService defines the service for producing messages
type producerService struct {
	canaryConfig *config.CanaryConfig
	client       sarama.Client
	producer     sarama.SyncProducer
	// index of the next message to send
	index int
}

// NewProducerService returns an instance of ProductService
func NewProducerService(canaryConfig *config.CanaryConfig, client sarama.Client) ProducerService {

	recordsProducedLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "records_produced_latency",
		Namespace: "strimzi_canary",
		Help:      "Records produced latency in milliseconds",
		Buckets:   canaryConfig.ProducerLatencyBuckets,
	}, []string{"clientid", "partition"})

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		glog.Fatalf("Error creating the Sarama sync producer: %v", err)
	}
	producer = otelsarama.WrapSyncProducer(client.Config(), producer)
	ps := producerService{
		canaryConfig: canaryConfig,
		client:       client,
		producer:     producer,
	}
	return &ps
}

// Send sends one message to partitions assigned to brokers
func (ps *producerService) Send(partitionsAssignments map[int32][]int32) {
	numPartitions := len(partitionsAssignments)
	msg := &sarama.ProducerMessage{
		Topic: ps.canaryConfig.Topic,
	}
	for i := 0; i < numPartitions; i++ {
		// build the message JSON payload and send to the current partition
		cm := ps.newCanaryMessage()
		msg.Value = sarama.StringEncoder(cm.Json())
		msg.Partition = int32(i)
		otel.GetTextMapPropagator().Inject(context.Background(), otelsarama.NewProducerMessageCarrier(msg))
		glog.V(1).Infof("Sending message: value=%s on partition=%d", msg.Value, msg.Partition)
		partition, offset, err := ps.producer.SendMessage(msg)
		timestamp := util.NowInMilliseconds() // timestamp in milliseconds
		labels := prometheus.Labels{
			"clientid":  ps.canaryConfig.ClientID,
			"partition": strconv.Itoa(i),
		}
		recordsProduced.With(labels).Inc()
		RecordsProducedCounter++
		if err != nil {
			glog.Warningf("Error sending message: %v", err)
			recordsProducedFailed.With(labels).Inc()
		} else {
			duration := timestamp - cm.Timestamp
			glog.V(1).Infof("Message sent: partition=%d, offset=%d, duration=%d ms", partition, offset, duration)
			recordsProducedLatency.With(labels).Observe(float64(duration))
		}
	}
}

// Refresh does a refresh metadata on the underneath Sarama client
func (ps *producerService) Refresh() {
	glog.Infof("Producer refreshing metadata")
	if err := ps.client.RefreshMetadata(ps.canaryConfig.Topic); err != nil {
		labels := prometheus.Labels{
			"clientid": ps.canaryConfig.ClientID,
		}
		refreshMetadataError.With(labels).Inc()
		glog.Errorf("Error refreshing metadata in producer: %v", err)
	}
}

// Close closes the underneath Sarama producer instance
func (ps *producerService) Close() {
	glog.Infof("Closing producer")
	err := ps.producer.Close()
	if err != nil {
		glog.Fatalf("Error closing the Sarama sync producer: %v", err)
	}
	glog.Infof("Producer closed")
}

func (ps *producerService) newCanaryMessage() CanaryMessage {
	ps.index++
	timestamp := util.NowInMilliseconds() // timestamp in milliseconds
	cm := CanaryMessage{
		ProducerID: ps.canaryConfig.ClientID,
		MessageID:  ps.index,
		Timestamp:  timestamp,
	}
	return cm
}
