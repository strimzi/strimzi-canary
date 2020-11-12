//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

type Producer struct {
	config   *config.CanaryConfig
	producer sarama.SyncProducer
	index    int
}

func NewProducer(config *config.CanaryConfig) Service {
	// TODO: add specific producer configuration
	producer, err := sarama.NewSyncProducer([]string{config.BootstrapServers}, nil)
	if err != nil {
		log.Printf("Error creating the Sarama sync producer: %v", err)
		panic(err)
	}
	p := Producer{
		config:   config,
		producer: producer,
	}
	return &p
}

func (p *Producer) Start() {
	log.Printf("Starting producer")
	go func() {
		for {
			cmJSON := p.newCanaryMessage().Json()
			msg := &sarama.ProducerMessage{
				Topic: p.config.Topic,
				Value: sarama.StringEncoder(cmJSON),
			}
			log.Printf("Sending message: value=%s\n", msg.Value)
			partition, offset, err := p.producer.SendMessage(msg)
			if err != nil {
				log.Printf("Erros sending message: %v\n", err)
			} else {
				log.Printf("Message sent: partition=%d, offset=%d\n", partition, offset)
			}
			time.Sleep(time.Duration(p.config.Delay) * time.Millisecond)
		}
	}()
}

func (p *Producer) Stop() {
	log.Printf("Stopping producer")
	err := p.producer.Close()
	if err != nil {
		log.Printf("Error closing the Sarama sync producer: %v", err)
		os.Exit(1)
	}
	log.Printf("Producer closed")
}

func (p *Producer) newCanaryMessage() CanaryMessage {
	p.index++
	timestamp := time.Now().UnixNano() / 1000000 // timestamp in milliseconds
	cm := CanaryMessage{
		ProducerID: p.config.ProducerClientID,
		MessageID:  p.index,
		Timestamp:  timestamp,
	}
	return cm
}
