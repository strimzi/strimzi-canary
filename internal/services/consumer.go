//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
)

const (
	// timeout on waiting the consumer to join the consumer group successfully
	waitConsumeTimeout = 30 * time.Second
	// delay betweet retries about the consumer to join the consumer group
	retryConsumeDelay = 5 * time.Second
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
	cancel      context.CancelFunc
	syncConsume sync.WaitGroup
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
		log.Printf("Error creating the Sarama consumer: %v", err)
		panic(err)
	}
	cs := ConsumerService{
		canaryConfig:  canaryConfig,
		client:        client,
		consumerGroup: consumerGroup,
	}
	return &cs
}

// Consume starts a Sarama consumer group instance consuming messages
//
// This function starts a goroutine calling in an endless loop the consume on the Sarama consumer group
// It can be exited cancelling the corresponding context through the cancel function provided by the ConsumerService instance
// Before returning, it waits for the consumer to join the group for all the partitions provided with numPartitions parameter
func (cs *ConsumerService) Consume(numPartitions int) {
	for {
		cgh := &consumerGroupHandler{
			consumerService: cs,
		}
		// creating new context with cancellation, for exiting Consume when metadata refresh is needed
		ctx, cancel := context.WithCancel(context.Background())
		cs.cancel = cancel
		cs.syncConsume.Add(numPartitions)
		go func() {
			// the Consume has to be in a loop, because each time a metadata refresh happens, this method exits
			// and needs to be called again for a new session and rejoining group
			for {
				// this method calls the methods handler on each stage: setup, consume and cleanup
				cs.consumerGroup.Consume(ctx, []string{cs.canaryConfig.Topic}, cgh)

				// check if context was cancelled, because of forcing a refresh metadata or exiting the consumer
				if ctx.Err() != nil {
					return
				}
			}
		}()

		// wait that the consumer is now subscribed to all partitions
		if isTimeout := cs.wait(waitConsumeTimeout); isTimeout {
			cs.cancel()
			labels := prometheus.Labels{
				"clientid": cs.canaryConfig.ClientID,
			}
			timeoutJoinGroup.With(labels).Inc()
			log.Printf("Consumer joining group timed out!")
			// TODO: improving with a backoff algorithm and max retries before giving up?
			time.Sleep(retryConsumeDelay)
		} else {
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
		cs.syncConsume.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

// Refresh does a refresh metadata
//
// Because of the way how Sarama consumer group works, the refresh is done in the following way:
// 1. calling the cancel context function for allowing the consumer group exiting the Consume function
// 2. calling again the Consume for refreshing metadata internally in Sarama and rejoining the consumer group
//
// NOTE: because of calling the Consume function, it waits for the consumer to join the group
// for all the partitions provided with numPartitions parameter and then returns
func (cs *ConsumerService) Refresh(numPartitions int) {
	if cs.cancel != nil {
		// cancel the consumer context to allow exiting the Consume loop
		// Consume will be called again to sync metadata and rejoining the group
		log.Printf("Consumer refreshing metadata")
		cs.cancel()
		cs.Consume(numPartitions)
	}
}

// Close closes the underneath Sarama consumer group instance
func (cs *ConsumerService) Close() {
	log.Printf("Closing consumer")
	err := cs.consumerGroup.Close()
	if err != nil {
		log.Printf("Error closing the Sarama consumer: %v", err)
		os.Exit(1)
	}
	log.Printf("Consumer closed")
}

// consumerGroupHandler defines the handler for the consuming Sarama functions
type consumerGroupHandler struct {
	consumerService *ConsumerService
}

func (cgh *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group handler setup\n")
	return nil
}

func (cgh *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group handler cleanup\n")
	return nil
}

func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Consumer group handler consumeclaim\n")
	// signal that consumer is now ready for the current partition
	cgh.consumerService.syncConsume.Done()

	for message := range claim.Messages() {
		timestamp := time.Now().UnixNano() / 1000000 // timestamp in milliseconds
		cm := NewCanaryMessage(message.Value)
		duration := timestamp - cm.Timestamp
		log.Printf("Message received: value=%+v, partition=%d, offset=%d, duration=%d ms", cm, message.Partition, message.Offset, duration)
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
