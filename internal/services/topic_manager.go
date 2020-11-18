//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines an interface for canary services and related implementations
package services

import (
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

type TopicManager struct {
	config   *config.CanaryConfig
	admin    sarama.ClusterAdmin
	stop     chan struct{}
	syncStop sync.WaitGroup
}

func NewTopicManager(config *config.CanaryConfig) Service {
	admin, err := sarama.NewClusterAdmin([]string{config.BootstrapServers}, nil)
	if err != nil {
		log.Printf("Error creating the Sarama cluster admin: %v", err)
		panic(err)
	}
	tm := TopicManager{
		config: config,
		admin:  admin,
	}
	return &tm
}

func (tm *TopicManager) Start() {
	log.Printf("Starting topic manager")

	tm.stop = make(chan struct{})
	tm.syncStop.Add(1)

	// start first reconcile immediately
	tm.reconcile()
	ticker := time.NewTicker(tm.config.TopicReconcile * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Topic manager reconcile")
				tm.reconcile()
			case <-tm.stop:
				ticker.Stop()
				defer tm.syncStop.Done()
				log.Println("Stopping topic manager reconcile loop")
				return
			}
		}
	}()
}

func (tm *TopicManager) Stop() {
	log.Printf("Stopping topic manager")

	// ask to stop the ticker reconcile loop and wait
	close(tm.stop)
	tm.syncStop.Wait()

	err := tm.admin.Close()
	if err != nil {
		log.Printf("Error closing the Sarama cluster admin: %v", err)
		os.Exit(1)
	}
	log.Printf("Topic manager closed")
}

func (tm *TopicManager) reconcile() {
	// getting brokers for assigning canary topic replicas accordingly
	// on creation or cluster scale up/down when topic already exists
	brokers, _, err := tm.admin.DescribeCluster()
	if err != nil {
		log.Printf("Error describing cluster: %v", err)
		panic(err)
	}

	metadata, err := tm.admin.DescribeTopics([]string{tm.config.Topic})
	if err != nil {
		log.Printf("Error retrieving metadata for topic %s: %v", tm.config.Topic, err)
		panic(err)
	}

	if metadata[0].Err == sarama.ErrUnknownTopicOrPartition {
		// canary topic doesn't exist, going to create it
		log.Printf("The canary topic %s doesn't exist\n", metadata[0].Name)
		if err := tm.createTopic(len(brokers)); err != nil {
			log.Printf("Error creating topic %s: %v", metadata[0].Name, err)
			panic(err)
		}
		log.Printf("The canary topic %s was created\n", metadata[0].Name)
	} else {
		// canary topic already exists, check replicas assignments
		log.Printf("The canary topic %s already exists\n", metadata[0].Name)

		if len(metadata[0].Partitions) < len(brokers) {
			log.Printf("Cluster scaled up, need to add partitions")
			// TODO: logic for adding partitions and assigning replicas across more brokers
		} else if len(metadata[0].Partitions) > len(brokers) {
			log.Printf("Cluster scaled down, need to declare 'orphan' partitions")
			// TODO: need to track "orphan" partitions to avoid producer sending to them but anyway assigning replicas across less brokers
		}
	}
}

func (tm *TopicManager) createTopic(brokersNumber int) error {
	partitions := brokersNumber
	replicationFactor := math.Min(float64(brokersNumber), 3)
	minISR := math.Max(1, replicationFactor-1)

	assignments := make(map[int32][]int32, partitions)
	for p := 0; p < partitions; p++ {
		assignments[int32(p)] = make([]int32, int(replicationFactor))
		k := p
		for r := 0; r < int(replicationFactor); r++ {
			assignments[int32(p)][r] = int32(k % int(replicationFactor))
			k++
		}
	}

	v := strconv.Itoa(int(minISR))
	topicConfig := map[string]*string{
		"min.insync.replicas": &v,
	}

	topicDetail := sarama.TopicDetail{
		NumPartitions:     -1,
		ReplicationFactor: -1,
		ReplicaAssignment: assignments,
		ConfigEntries:     topicConfig,
	}
	return tm.admin.CreateTopic(tm.config.Topic, &topicDetail, false)
}
