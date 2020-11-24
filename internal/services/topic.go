//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"log"
	"os"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

type TopicReconcileResult struct {
	BrokersNumber   int
	Assignments     map[int32][]int32
	RefreshMetadata bool
}

type TopicService struct {
	canaryConfig *config.CanaryConfig
	client       sarama.Client
	admin        sarama.ClusterAdmin
}

func NewTopicService(canaryConfig *config.CanaryConfig, client sarama.Client) *TopicService {
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		log.Printf("Error creating the Sarama cluster admin: %v", err)
		panic(err)
	}
	ts := TopicService{
		canaryConfig: canaryConfig,
		client:       client,
		admin:        admin,
	}
	return &ts
}

func (ts *TopicService) Reconcile() (TopicReconcileResult, error) {
	result := TopicReconcileResult{0, nil, false}
	// getting brokers for assigning canary topic replicas accordingly
	// on creation or cluster scale up/down when topic already exists
	brokers, _, err := ts.admin.DescribeCluster()
	if err != nil {
		log.Printf("Error describing cluster: %v", err)
		return result, err
	}
	result.BrokersNumber = len(brokers)

	metadata, err := ts.admin.DescribeTopics([]string{ts.canaryConfig.Topic})
	if err != nil {
		log.Printf("Error retrieving metadata for topic %s: %v", ts.canaryConfig.Topic, err)
		return result, err
	}
	topicMetadata := metadata[0]

	if topicMetadata.Err == sarama.ErrUnknownTopicOrPartition {
		// canary topic doesn't exist, going to create it
		log.Printf("The canary topic %s doesn't exist\n", topicMetadata.Name)
		if result.Assignments, err = ts.createTopic(len(brokers)); err != nil {
			log.Printf("Error creating topic %s: %v", topicMetadata.Name, err)
			return result, err
		}
		log.Printf("The canary topic %s was created\n", topicMetadata.Name)
	} else {
		// canary topic already exists, check replicas assignments
		log.Printf("The canary topic %s already exists\n", topicMetadata.Name)

		// if we scale up then scale down and then scale up again, the preferred leader are not elected immediately
		// we should check current assignments, leaders and maybe forcing a leader election (not supported by Sarama right now)
		// TODO

		result.RefreshMetadata = len(brokers) != len(topicMetadata.Partitions)
		if result.Assignments, err = ts.alterTopic(len(topicMetadata.Partitions), len(brokers)); err != nil {
			log.Printf("Error altering topic %s: %v", topicMetadata.Name, err)
			return result, err
		}
		ts.checkTopic(len(brokers), topicMetadata)
	}
	return result, err
}

func (ts *TopicService) Close() {
	log.Printf("Closing topic service")

	err := ts.admin.Close()
	if err != nil {
		log.Printf("Error closing the Sarama cluster admin: %v", err)
		os.Exit(1)
	}
	log.Printf("Topic service closed")
}

func (ts *TopicService) createTopic(brokersNumber int) (map[int32][]int32, error) {
	assignments, minISR := ts.assignments(0, brokersNumber)

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
	err := ts.admin.CreateTopic(ts.canaryConfig.Topic, &topicDetail, false)
	return assignments, err
}

func (ts *TopicService) alterTopic(currentPartitions int, brokersNumber int) (map[int32][]int32, error) {
	assignments, _ := ts.assignments(currentPartitions, brokersNumber)

	ass := make([][]int32, len(assignments))
	for i := 0; i < len(ass); i++ {
		ass[i] = make([]int32, len(assignments[int32(i)]))
		copy(ass[i], assignments[int32(i)])
	}
	log.Printf("%v", ass)

	var err error
	// less partitions than brokers (scale up)
	if currentPartitions < brokersNumber {
		// passing the assigments just for the partitions that needs to be created
		err = ts.admin.CreatePartitions(ts.canaryConfig.Topic, int32(brokersNumber), ass[currentPartitions:], false)
	} else {
		// more or equals partitions than brokers, just need reassignment
		err = ts.admin.AlterPartitionReassignments(ts.canaryConfig.Topic, ass)
	}
	return assignments, err
}

func (ts *TopicService) checkTopic(brokersNumber int, metadata *sarama.TopicMetadata) {
	electLeader := false
	if len(metadata.Partitions) == brokersNumber {
		for _, p := range metadata.Partitions {
			if p.ID != p.Leader {
				electLeader = true
			}
		}
	}
	log.Printf("Elect leader = %t\n", electLeader)
}

func (ts *TopicService) assignments(currentPartitions int, brokersNumber int) (map[int32][]int32, int) {
	partitions := max(currentPartitions, brokersNumber)
	replicationFactor := min(brokersNumber, 3)
	minISR := max(1, replicationFactor-1)

	assignments := make(map[int32][]int32, int(partitions))
	for p := 0; p < int(partitions); p++ {
		assignments[int32(p)] = make([]int32, int(replicationFactor))
		k := p
		for r := 0; r < int(replicationFactor); r++ {
			assignments[int32(p)][r] = int32(k % int(brokersNumber))
			k++
		}
	}
	log.Printf("assignments = %v, minISR = %d", assignments, int(minISR))
	return assignments, int(minISR)
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
