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

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
)

const (
	// timeout on waiting the consumer to join the consumer group successfully
	waitConsumeTimeout = 30 * time.Second
	// maximum number of attempts for consumer to join the consumer group successfully
	maxConsumeAttempts = 3
)

var (
	recordsConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "records_consumed_total",
		Namespace: "strimzi_canary",
		Help:      "The total number of records consumed",
	}, []string{"clientid", "partition"})

	// it's defined when the service is created because buckets are configurable
	recordsEndToEndLatency *prometheus.HistogramVec

	timeoutJoinGroup = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "consumer_timeout_join_group_total",
		Namespace: "strimzi_canary",
		Help:      "The total number of consumer not joining the group within the timeout",
	}, []string{"clientid"})
)

// ConsumerService defines the service for consuming messages
type ConsumerService struct {
	canaryConfig  *config.CanaryConfig
	client        sarama.Client
	consumerGroup sarama.ConsumerGroup
	// reference to the function for cancelling the Sarama consumer group context
	// in order to ending the session and allowing a rejoin with rebalancing
	cancel context.CancelFunc
	ready  chan bool
}

// NewConsumerService returns an instance of ConsumerService
func NewConsumerService(canaryConfig *config.CanaryConfig, client sarama.Client) *ConsumerService {
	recordsEndToEndLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "records_consumed_latency",
		Namespace: "strimzi_canary",
		Help:      "Records end-to-end latency in milliseconds",
		Buckets:   canaryConfig.EndToEndLatencyBuckets,
	}, []string{"clientid", "partition"})

	consumerGroup, err := sarama.NewConsumerGroupFromClient(canaryConfig.ConsumerGroupID, client)
	if err != nil {
		glog.Errorf("Error creating the Sarama consumer: %v", err)
		panic(err)
	}
	cs := ConsumerService{
		canaryConfig:  canaryConfig,
		client:        client,
		consumerGroup: consumerGroup,
		ready:         make(chan bool),
	}
	return &cs
}

// Consume starts a Sarama consumer group instance consuming messages
//
// This function starts a goroutine calling in an endless loop the consume on the Sarama consumer group
// It can be exited cancelling the corresponding context through the cancel function provided by the ConsumerService instance
// Before returning, it waits for the consumer to join the group for all the topic partitions
func (cs *ConsumerService) Consume() {
	backoff := NewBackoff(maxConsumeAttempts, 5000*time.Millisecond, MaxDefault)
	for {
		cgh := &consumerGroupHandler{
			consumerService: cs,
		}
		// creating new context with cancellation, for exiting Consume when metadata refresh is needed
		ctx, cancel := context.WithCancel(context.Background())
		cs.cancel = cancel
		go func() {
			// the Consume has to be in a loop, because each time a metadata refresh happens, this method exits
			// and needs to be called again for a new session and rejoining group
			for {

				glog.Infof("Consumer group consume starting...")
				// this method calls the methods handler on each stage: setup, consume and cleanup
				cs.consumerGroup.Consume(ctx, []string{cs.canaryConfig.Topic}, cgh)

				// check if context was cancelled, because of forcing a refresh metadata or exiting the consumer
				if ctx.Err() != nil {
					glog.Infof("Consumer group context cancelled")
					return
				}
				cs.ready = make(chan bool)
			}
		}()

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
func (cs *ConsumerService) wait(timeout time.Duration) bool {
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

// Close closes the underneath Sarama consumer group instance
func (cs *ConsumerService) Close() {
	glog.Infof("Closing consumer")
	err := cs.consumerGroup.Close()
	if err != nil {
		glog.Fatalf("Error closing the Sarama consumer: %v", err)
	}
	glog.Infof("Consumer closed")
}

// consumerGroupHandler defines the handler for the consuming Sarama functions
type consumerGroupHandler struct {
	consumerService *ConsumerService
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

	for message := range claim.Messages() {
		timestamp := time.Now().UnixNano() / 1000000 // timestamp in milliseconds
		cm := NewCanaryMessage(message.Value)
		duration := timestamp - cm.Timestamp
		glog.V(1).Infof("Message received: value=%+v, partition=%d, offset=%d, duration=%d ms", cm, message.Partition, message.Offset, duration)
		session.MarkMessage(message, "")
		labels := prometheus.Labels{
			"clientid":  cgh.consumerService.canaryConfig.ClientID,
			"partition": strconv.Itoa(int(message.Partition)),
		}
		recordsEndToEndLatency.With(labels).Observe(float64(duration))
		recordsConsumed.With(labels).Inc()
	}
	return nil
}
