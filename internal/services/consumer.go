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

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

type ConsumerService interface {
	Consume()
	Refresh()
	Close()
}

type consumerService struct {
	canaryConfig  *config.CanaryConfig
	client        sarama.Client
	consumerGroup sarama.ConsumerGroup
	cancel        context.CancelFunc
}

func NewConsumerService(canaryConfig *config.CanaryConfig, client sarama.Client) ConsumerService {
	consumerGroup, err := sarama.NewConsumerGroupFromClient(canaryConfig.ConsumerGroupID, client)
	if err != nil {
		log.Printf("Error creating the Sarama consumer: %v", err)
		panic(err)
	}
	cs := consumerService{
		canaryConfig:  canaryConfig,
		client:        client,
		consumerGroup: consumerGroup,
	}
	return &cs
}

func (cs *consumerService) Consume() {
	cgh := &consumerGroupHandler{}
	// creating new context with cancellation, for exiting Consume when metadata refresh is needed
	ctx, cancel := context.WithCancel(context.Background())
	cs.cancel = cancel
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
}

func (cs *consumerService) Refresh() {
	if cs.cancel != nil {
		// cancel the consumer context to allow exiting the Consume loop
		// Consume will be called again to sync metadata and rejoining the group
		log.Printf("Consumer refreshing metadata")
		cs.cancel()
		cs.Consume()
	}
}

func (cs *consumerService) Close() {
	log.Printf("Closing consumer")
	err := cs.consumerGroup.Close()
	if err != nil {
		log.Printf("Error closing the Sarama consumer: %v", err)
		os.Exit(1)
	}
	log.Printf("Consumer closed")
}

// struct defining the handler for the consuming Sarama method
type consumerGroupHandler struct {
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
	for message := range claim.Messages() {
		cm := NewCanaryMessage(message.Value)
		log.Printf("Message received: value=%+v, partition=%d, offset=%d", cm, message.Partition, message.Offset)
		session.MarkMessage(message, "")
	}
	return nil
}
