//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/util"
)

const (
	// timeout on waiting the consumer to join the consumer group successfully
	waitConsumeTimeout = 30 * time.Second
	// maximum number of attempts for consumer to join the consumer group successfully
	maxConsumeAttempts = 3
	// delay to wait after a consume error (i.e. due to a broker offline, leader election, ...)
	consumeDelay = 2 * time.Second
)

var (
	RecordsConsumedCounter       uint64 = 0
	recordsConsumed              *prometheus.CounterVec
	recordsConsumerFailed        *prometheus.CounterVec
	recordsEndToEndLatency       *prometheus.HistogramVec
	timeoutJoinGroup             *prometheus.CounterVec
	refreshConsumerMetadataError *prometheus.CounterVec
)

// ConsumerService defines the service for consuming messages
type ConsumerService interface {
	Consume()
	Refresh()
	Leaders() (map[int32]int32, error)
	Close()
}

type consumerService struct {
	canaryConfig  *config.CanaryConfig
	client        sarama.Client
	consumerGroup sarama.ConsumerGroup
	// reference to the function for cancelling the Sarama consumer group context
	// in order to ending the session and allowing a rejoin with rebalancing
	cancel context.CancelFunc
	ready  chan bool
}

// NewConsumerService returns an instance of ConsumerService
func NewConsumerService(canaryConfig *config.CanaryConfig, client sarama.Client) ConsumerService {

	recordsConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "records_consumed_total",
		Namespace:   "strimzi_canary",
		Help:        "The total number of records consumed",
		ConstLabels: canaryConfig.PrometheusConstantLabels,
	}, []string{"clientid", "partition"})

	recordsConsumerFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "consumer_error_total",
		Namespace:   "strimzi_canary",
		Help:        "Total number of errors reported by the consumer",
		ConstLabels: canaryConfig.PrometheusConstantLabels,
	}, []string{"clientid"})

	recordsEndToEndLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "records_consumed_latency",
		Namespace:   "strimzi_canary",
		Help:        "Records end-to-end latency in milliseconds",
		ConstLabels: canaryConfig.PrometheusConstantLabels,
		Buckets:     canaryConfig.EndToEndLatencyBuckets,
	}, []string{"clientid", "partition"})

	timeoutJoinGroup = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "consumer_timeout_join_group_total",
		Namespace:   "strimzi_canary",
		Help:        "The total number of consumers not joining the group within the timeout",
		ConstLabels: canaryConfig.PrometheusConstantLabels,
	}, []string{"clientid"})

	refreshConsumerMetadataError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:        "consumer_refresh_metadata_error_total",
		Namespace:   "strimzi_canary",
		Help:        "Total number of errors while refreshing consumer metadata",
		ConstLabels: canaryConfig.PrometheusConstantLabels,
	}, []string{"clientid"})

	consumerGroup, err := sarama.NewConsumerGroupFromClient(canaryConfig.ConsumerGroupID, client)
	if err != nil {
		glog.Fatalf("Error creating the Sarama consumer: %v", err)
	}

	cs := consumerService{
		canaryConfig:  canaryConfig,
		client:        client,
		consumerGroup: consumerGroup,
		ready:         make(chan bool),
	}

	labels := prometheus.Labels{
		"clientid": canaryConfig.ClientID,
	}
	// initialize all error-related metrics with starting value of 0
	recordsConsumerFailed.With(labels).Add(0)
	refreshConsumerMetadataError.With(labels).Add(0)

	go func() {
		for err := range consumerGroup.Errors() {
			glog.Errorf("Error received whilst consuming from topic: %v", err)
			recordsConsumerFailed.With(labels).Inc()
		}
	}()

	return &cs
}

// Consume starts a Sarama consumer group instance consuming messages
//
// This function starts a goroutine calling in an endless loop the consume on the Sarama consumer group
// It can be exited cancelling the corresponding context through the cancel function provided by the ConsumerService instance
// Before returning, it waits for the consumer to join the group for all the topic partitions
func (cs *consumerService) Consume() {
	backoff := NewBackoff(maxConsumeAttempts, 5000*time.Millisecond, MaxDefault)
	for {
		cgh := &consumerGroupHandler{
			consumerService: cs,
		}
		h := otelsarama.WrapConsumerGroupHandler(cgh)

		// creating new context with cancellation, for exiting Consume when metadata refresh is needed
		ctx, cancel := context.WithCancel(context.Background())
		cs.cancel = cancel
		go func() {
			// the Consume has to be in a loop, because each time a metadata refresh happens, this method exits
			// and needs to be called again for a new session and rejoining group
			for {

				glog.Infof("Consumer group consume starting...")
				// this method calls the methods handler on each stage: setup, consume and cleanup
				if err := cs.consumerGroup.Consume(ctx, []string{cs.canaryConfig.Topic}, h); err != nil {
					glog.Errorf("Error consuming topic: %s", err.Error())
					time.Sleep(consumeDelay)
					continue
				}

				// check if context was cancelled, because of forcing a refresh metadata or exiting the consumer
				if ctx.Err() != nil {
					glog.Infof("Consumer group context cancelled")
					return
				}
				cs.ready = make(chan bool)
			}
		}()

		glog.Infof("Waiting consumer group to be up and running")
		// wait that the consumer is now subscribed to all partitions
		if isTimeout := cs.wait(waitConsumeTimeout); isTimeout {
			cs.cancel()
			labels := prometheus.Labels{
				"clientid": cs.canaryConfig.ClientID,
			}
			timeoutJoinGroup.With(labels).Inc()
			glog.Warningf("Consumer joining group timed out!")
			delay, err := backoff.Delay()
			if err != nil {
				glog.Fatalf("Error joining the consumer group: %v", err)
			}
			time.Sleep(delay)
		} else {
			glog.Infof("Sarama consumer group up and running")
			break
		}
	}
}

// Waits on the wait group about the consumer joining the group and starting
//
// It is possible to specify a timeout on waiting
// Returns true if waiting timed out, otherwise false
func (cs *consumerService) wait(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		<-cs.ready
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

// Refresh does a refresh metadata on the underneath Sarama client
func (cs *consumerService) Refresh() {
	glog.Infof("Consumer refreshing metadata")
	if err := cs.client.RefreshMetadata(cs.canaryConfig.Topic); err != nil {
		labels := prometheus.Labels{
			"clientid": cs.canaryConfig.ClientID,
		}
		refreshConsumerMetadataError.With(labels).Inc()
		glog.Errorf("Error refreshing metadata in consumer: %v", err)
	}
}

func (cs *consumerService) Leaders() (map[int32]int32, error) {
	partitions, err := cs.client.Partitions(cs.canaryConfig.Topic)
	if err != nil {
		return nil, err
	}
	leaders := make(map[int32]int32, len(partitions))
	for _, p := range partitions {
		leader, err := cs.client.Leader(cs.canaryConfig.Topic, p)
		if err != nil {
			return nil, err
		}
		leaders[p] = leader.ID()
	}
	return leaders, err
}

// Close closes the underneath Sarama consumer group instance
func (cs *consumerService) Close() {
	glog.Infof("Closing consumer")
	cs.cancel()
	err := cs.consumerGroup.Close()
	if err != nil {
		glog.Fatalf("Error closing the Sarama consumer: %v", err)
	}
	glog.Infof("Consumer closed")
}

// consumerGroupHandler defines the handler for the consuming Sarama functions
type consumerGroupHandler struct {
	consumerService *consumerService
}

func (cgh *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	glog.Infof("Consumer group setup")
	// signaling the consumer group is ready
	close(cgh.consumerService.ready)
	return nil
}

func (cgh *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	glog.Infof("Consumer group cleanup")
	return nil
}

func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	glog.Infof("Consumer group consumeclaim on %s [%d]", claim.Topic(), claim.Partition())
	tr := otel.Tracer("consumer")
	for message := range claim.Messages() {
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), otelsarama.NewConsumerMessageCarrier(message))
		_, span := tr.Start(ctx, "consume message", trace.WithAttributes(
			semconv.MessagingOperationProcess,
		))
		timestamp := util.NowInMilliseconds() // timestamp in milliseconds
		cm := NewCanaryMessage(message.Value)
		duration := timestamp - cm.Timestamp
		glog.V(1).Infof("Message received: value=%+v, partition=%d, offset=%d, duration=%d ms", cm, message.Partition, message.Offset, duration)
		span.End()
		session.MarkMessage(message, "")
		labels := prometheus.Labels{
			"clientid":  cgh.consumerService.canaryConfig.ClientID,
			"partition": strconv.Itoa(int(message.Partition)),
		}
		recordsEndToEndLatency.With(labels).Observe(float64(duration))
		recordsConsumed.With(labels).Inc()
		RecordsConsumedCounter++
	}
	return nil
}
