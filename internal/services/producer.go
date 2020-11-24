//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

// ProducerService defines the service for producing messages
type ProducerService struct {
	canaryConfig *config.CanaryConfig
	client       sarama.Client
	producer     sarama.SyncProducer
	// index of the next message to send
	index int
}

// NewProducerService returns an instance of ProductService
func NewProducerService(canaryConfig *config.CanaryConfig, client sarama.Client) *ProducerService {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		log.Printf("Error creating the Sarama sync producer: %v", err)
		panic(err)
	}
	ps := ProducerService{
		canaryConfig: canaryConfig,
		client:       client,
		producer:     producer,
	}
	return &ps
}

// Send sends one message to each partition from 0 to numPartitions specified as parameter
func (ps *ProducerService) Send(numPartitions int) {
	msg := &sarama.ProducerMessage{
		Topic: ps.canaryConfig.Topic,
	}
	for i := 0; i < numPartitions; i++ {
		// build the message JSON payload and send to the current partition
		cmJSON := ps.newCanaryMessage().Json()
		msg.Value = sarama.StringEncoder(cmJSON)
		msg.Partition = int32(i)
		log.Printf("Sending message: value=%s on partition=%d\n", msg.Value, msg.Partition)
		partition, offset, err := ps.producer.SendMessage(msg)
		if err != nil {
			log.Printf("Erros sending message: %v\n", err)
		} else {
			log.Printf("Message sent: partition=%d, offset=%d\n", partition, offset)
		}
	}
}

// Refresh does a refresh metadata on the underneath Sarama client
func (ps *ProducerService) Refresh() {
	log.Printf("Producer refreshing metadata")
	if err := ps.client.RefreshMetadata(ps.canaryConfig.Topic); err != nil {
		log.Printf("Errors producer refreshing metadata: %v\n", err)
	}
}

// Close closes the underneath Sarama producer instance
func (ps *ProducerService) Close() {
	log.Printf("Closing producer")
	err := ps.producer.Close()
	if err != nil {
		log.Printf("Error closing the Sarama sync producer: %v", err)
		os.Exit(1)
	}
	log.Printf("Producer closed")
}

func (ps *ProducerService) newCanaryMessage() CanaryMessage {
	ps.index++
	timestamp := time.Now().UnixNano() / 1000000 // timestamp in milliseconds
	cm := CanaryMessage{
		ProducerID: ps.canaryConfig.ClientID,
		MessageID:  ps.index,
		Timestamp:  timestamp,
	}
	return cm
}
