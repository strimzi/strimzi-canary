//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package workers defines an interface for canary workers and related implementations
package workers

import (
	"context"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/services"
)

type Consumer struct {
	config        *config.CanaryConfig
	consumerGroup sarama.ConsumerGroup
}

func NewConsumer(config *config.CanaryConfig) Worker {
	// TODO: add specific consumer configuration
	consumerGroup, err := sarama.NewConsumerGroup([]string{config.BootstrapServers}, config.ConsumerGroupID, nil)
	if err != nil {
		log.Printf("Error creating the Sarama consumer: %v", err)
		panic(err)
	}
	c := Consumer{
		config:        config,
		consumerGroup: consumerGroup,
	}
	return &c
}

func (c *Consumer) Start() {
	cgh := &consumerGroupHandler{}
	ctx := context.Background()
	go func() {
		for {
			// this method calls the methods handler on each stage: setup, consume and cleanup
			c.consumerGroup.Consume(ctx, []string{c.config.Topic}, cgh)
		}
	}()
}

func (c *Consumer) Stop() {
	log.Printf("Stopping consumer")
	err := c.consumerGroup.Close()
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
	for message := range claim.Messages() {
		cm := services.NewCanaryMessage(message.Value)
		log.Printf("Message received: value=%+v, partition=%d, offset=%d", cm, message.Partition, message.Offset)
		session.MarkMessage(message, "")
	}
	return nil
}
