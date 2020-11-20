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

type ProducerService interface {
	Send(numPartitions int)
	Close()
}

type producerService struct {
	config   *config.CanaryConfig
	producer sarama.SyncProducer
	index    int
}

func NewProducerService(config *config.CanaryConfig) ProducerService {
	producerConfig := sarama.NewConfig()
	// set manual partitioner in order to specify the destination partition on sending
	producerConfig.Producer.Partitioner = sarama.NewManualPartitioner
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{config.BootstrapServers}, producerConfig)
	if err != nil {
		log.Printf("Error creating the Sarama sync producer: %v", err)
		panic(err)
	}
	ps := producerService{
		config:   config,
		producer: producer,
	}
	return &ps
}

func (ps *producerService) Send(numPartitions int) {
	msg := &sarama.ProducerMessage{
		Topic: ps.config.Topic,
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

func (ps *producerService) Close() {
	log.Printf("Closing producer")
	err := ps.producer.Close()
	if err != nil {
		log.Printf("Error closing the Sarama sync producer: %v", err)
		os.Exit(1)
	}
	log.Printf("Producer closed")
}

func (ps *producerService) newCanaryMessage() CanaryMessage {
	ps.index++
	timestamp := time.Now().UnixNano() / 1000000 // timestamp in milliseconds
	cm := CanaryMessage{
		ProducerID: ps.config.ProducerClientID,
		MessageID:  ps.index,
		Timestamp:  timestamp,
	}
	return cm
}
